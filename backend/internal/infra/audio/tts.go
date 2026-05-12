package audio

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *service) TTS(ctx context.Context, text, voiceID string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("语音合成失败")
	}

	textCh := make(chan string, 1)
	textCh <- text
	close(textCh)

	audioCh, errCh := s.runTTS(ctx, textCh, voiceID, s.cfg.TTS.Format, s.cfg.TTS.SampleRate)
	audioData := make([]byte, 0, 32*1024)

	for audioCh != nil || errCh != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case audio, ok := <-audioCh:
			if !ok {
				audioCh = nil
				continue
			}
			audioData = append(audioData, audio...)
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return audioData, nil
}

func (s *service) ttsSender(ctx context.Context, textCh <-chan string, ws *websocket.Conn, taskID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-textCh:
			if !ok {
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
			if text == "" {
				continue
			}
			if err := ws.WriteJSON(map[string]any{
				"header": map[string]any{
					"action":    "continue-task",
					"task_id":   taskID,
					"streaming": "duplex",
				},
				"payload": map[string]any{
					"input": map[string]any{
						"text": text,
					},
				},
			}); err != nil {
				return err
			}
		}
	}
}

func (s *service) ttsReceiver(ctx context.Context, audioCh chan<- []byte, ws *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			return err
		}

		if messageType == websocket.BinaryMessage {
			audio := append([]byte(nil), msg...)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case audioCh <- audio:
			}
			continue
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			return err
		}
		header, _ := data["header"].(map[string]any)
		event, _ := header["event"].(string)
		if event == "task-finished" || event == "task-failed" {
			break
		}
	}

	return nil
}

func (s *service) runTTSTasks(ctx context.Context, textCh <-chan string, audioCh chan<- []byte, voiceID, format string, sampleRate int) error {
	taskID := strings.ReplaceAll(uuid.NewString(), "-", "")
	if strings.TrimSpace(voiceID) == "" {
		voiceID = s.cfg.TTS.Voice
	}

	ws, err := s.dial()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.WriteJSON(map[string]any{
		"header": map[string]any{
			"action":    "run-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]any{
			"task_group": "audio",
			"task":       "tts",
			"function":   "SpeechSynthesizer",
			"model":      s.cfg.TTS.Model,
			"parameters": map[string]any{
				"text_type":   "PlainText",
				"voice":       voiceID,
				"format":      format,
				"sample_rate": sampleRate,
				"volume":      s.cfg.TTS.Volume,
				"rate":        s.cfg.TTS.Rate,
				"pitch":       s.cfg.TTS.Pitch,
			},
			"input": map[string]any{},
		},
	}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			return err
		}
		header, _ := data["header"].(map[string]any)
		event, _ := header["event"].(string)
		if event == "task-started" {
			break
		}
	}

	var senderErr error
	var receiverErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		senderErr = s.ttsSender(ctx, textCh, ws, taskID)
	}()
	go func() {
		defer wg.Done()
		receiverErr = s.ttsReceiver(ctx, audioCh, ws)
	}()
	wg.Wait()

	if senderErr != nil {
		return senderErr
	}
	return receiverErr
}

func (s *service) StreamTTS(ctx context.Context, textCh <-chan string, voiceID string) (<-chan []byte, <-chan error) {
	return s.runTTS(ctx, textCh, voiceID, s.cfg.TTS.StreamFormat, s.cfg.TTS.StreamSampleRate)
}

func (s *service) runTTS(ctx context.Context, textCh <-chan string, voiceID, format string, sampleRate int) (<-chan []byte, <-chan error) {
	audioCh := make(chan []byte, 16)
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		defer close(audioCh)

		if err := s.runTTSTasks(ctx, textCh, audioCh, voiceID, format, sampleRate); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	return audioCh, errCh
}
