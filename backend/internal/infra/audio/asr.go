package audio

import (
	"backend/pkg/constants"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"
)

func (s *service) ASR(ctx context.Context, pcmData []byte) (string, error) {
	if len(pcmData) == 0 {
		return "", errors.New(constants.ErrAudioNotFound)
	}

	conn, err := s.dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	taskID := uuid.NewString()
	if err := s.sendASRStart(conn, taskID); err != nil {
		return "", err
	}
	if err := waitForTaskStarted(ctx, conn); err != nil {
		return "", err
	}
	if err := sendASRAudio(ctx, conn, pcmData); err != nil {
		return "", err
	}
	if err := s.sendASRFinish(conn, taskID); err != nil {
		return "", err
	}

	return receiveASRText(ctx, conn)
}

// 以下是对接阿里云的流程
func (s *service) sendASRStart(conn *gorillawebsocket.Conn, taskID string) error {
	payload := asrRunPayload{
		Model:     s.cfg.ASR.Model,
		Input:     map[string]any{},
		Task:      constants.AudioASRTask,
		TaskGroup: constants.AudioTaskGroup,
		Function:  constants.AudioASRFunction,
	}
	payload.Parameters.SampleRate = s.cfg.ASR.SampleRate
	payload.Parameters.Format = s.cfg.ASR.Format
	payload.Parameters.TranscriptionEnabled = true

	msg := wsEnvelope{
		Header: wsHeader{
			Streaming: constants.AudioStreamingDuplex,
			TaskID:    taskID,
			Action:    constants.AudioRunTaskAction,
		},
		Payload: payload,
	}
	return writeWSJSON(conn, msg)
}

func waitForTaskStarted(ctx context.Context, conn *gorillawebsocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var msg asrEventMessage
		if err := readWSJSON(conn, &msg); err != nil {
			return err
		}
		switch msg.Header.Event {
		case constants.AudioTaskStartedEvent:
			return nil
		case constants.AudioTaskFailedEvent:
			return errors.New("asr task failed")
		}
	}
}

func sendASRAudio(ctx context.Context, conn *gorillawebsocket.Conn, pcmData []byte) error {
	for i := 0; i < len(pcmData); i += constants.ASRChunkSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := min(i+constants.ASRChunkSize, len(pcmData))

		if err := writeWSBinary(conn, pcmData[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func receiveASRText(ctx context.Context, conn *gorillawebsocket.Conn) (string, error) {
	var builder strings.Builder

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		raw, err := readWSMessage(conn)
		if err != nil {
			return "", err
		}

		var msg asrEventMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Header.Event {
		case constants.AudioResultEvent:
			transcription := msg.Payload.Output.Transcription
			if transcription != nil && transcription.SentenceEnd {
				builder.WriteString(transcription.Text)
			}
		case constants.AudioTaskFinishedEvent:
			return builder.String(), nil
		case constants.AudioTaskFailedEvent:
			return "", errors.New("asr task failed")
		}
	}
}

func (s *service) sendASRFinish(conn *gorillawebsocket.Conn, taskID string) error {
	msg := wsEnvelope{
		Header: wsHeader{
			Action:    constants.AudioFinishTaskAction,
			TaskID:    taskID,
			Streaming: constants.AudioStreamingDuplex,
		},
		Payload: map[string]any{
			"input": map[string]any{},
		},
	}
	return writeWSJSON(conn, msg)
}
