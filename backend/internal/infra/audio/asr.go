package audio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *service) ASR(ctx context.Context, pcmData []byte) (string, error) {
	if len(pcmData) == 0 {
		return "", errors.New("音频不存在")
	}
	return s.runASRTasks(ctx, pcmData)
}

func (s *service) asrSender(ctx context.Context, pcmData []byte, ws *websocket.Conn, taskID string) error {
	chunk := 3200
	for i := 0; i < len(pcmData); i += chunk {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := min(i+chunk, len(pcmData))
		if err := ws.WriteMessage(websocket.BinaryMessage, pcmData[i:end]); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}

	return ws.WriteJSON(map[string]any{
		"header": map[string]any{
			"action":    "finish-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]any{
			"input": map[string]any{},
		},
	})
}

func (s *service) asrReceiver(ctx context.Context, ws *websocket.Conn) (string, error) {
	text := ""

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			return "", err
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			return "", err
		}

		header, _ := data["header"].(map[string]any)
		event, _ := header["event"].(string)
		if event == "result-generated" {
			payload, _ := data["payload"].(map[string]any)
			output, _ := payload["output"].(map[string]any)
			transcription, _ := output["transcription"].(map[string]any)
			sentenceEnd, _ := transcription["sentence_end"].(bool)
			if transcription != nil && sentenceEnd {
				transcriptionText, _ := transcription["text"].(string)
				text += transcriptionText
			}
		} else if event == "task-finished" || event == "task-failed" {
			break
		}
	}

	return text, nil
}

func (s *service) runASRTasks(ctx context.Context, pcmData []byte) (string, error) {
	taskID := strings.ReplaceAll(uuid.NewString(), "-", "")
	apiKey := os.Getenv("API_KEY")
	wssURL := os.Getenv("WSS_URL")
	headers := http.Header{
		"Authorization": []string{"Bearer " + apiKey},
	}

	ws, _, err := websocket.DefaultDialer.Dial(wssURL, headers)
	if err != nil {
		return "", err
	}
	defer ws.Close()

	if err := ws.WriteJSON(map[string]any{
		"header": map[string]any{
			"streaming": "duplex",
			"task_id":   taskID,
			"action":    "run-task",
		},
		"payload": map[string]any{
			"model": "gummy-realtime-v1",
			"parameters": map[string]any{
				"sample_rate":           16000,
				"format":                "pcm",
				"transcription_enabled": true,
			},
			"input":      map[string]any{},
			"task":       "asr",
			"task_group": "audio",
			"function":   "recognition",
		},
	}); err != nil {
		return "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			return "", err
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			return "", err
		}
		header, _ := data["header"].(map[string]any)
		event, _ := header["event"].(string)
		if event == "task-started" {
			break
		}
	}

	var text string
	var senderErr error
	var receiverErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		senderErr = s.asrSender(ctx, pcmData, ws, taskID)
	}()
	go func() {
		defer wg.Done()
		text, receiverErr = s.asrReceiver(ctx, ws)
	}()
	wg.Wait()

	if senderErr != nil {
		return "", senderErr
	}
	if receiverErr != nil {
		return "", receiverErr
	}
	return text, nil
}
