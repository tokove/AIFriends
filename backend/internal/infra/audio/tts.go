package audio

import (
	"backend/pkg/constants"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *service) TTS(ctx context.Context, textCh <-chan string, voiceID string) (<-chan []byte, <-chan error) {
	audioCh := make(chan []byte, 16)
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		defer close(audioCh)

		conn, err := s.dial()
		if err != nil {
			trySendTTSError(errCh, err)
			return
		}
		defer conn.Close()

		taskID := uuid.NewString()
		if err := s.sendTTSStart(conn, taskID, voiceID); err != nil {
			trySendTTSError(errCh, err)
			return
		}
		if err := waitForTaskStarted(ctx, conn); err != nil {
			trySendTTSError(errCh, err)
			return
		}

		senderErrCh := make(chan error, 1)
		go func() {
			senderErrCh <- s.streamTTSSender(ctx, conn, taskID, textCh)
		}()

		for {
			raw, err := readWSMessage(conn)
			if err != nil {
				select {
				case senderErr := <-senderErrCh:
					if senderErr != nil && !errors.Is(senderErr, context.Canceled) {
						trySendTTSError(errCh, senderErr)
					}
				default:
					if ctx.Err() == nil {
						trySendTTSError(errCh, err)
					}
				}
				return
			}

			var msg asrEventMessage
			if err := json.Unmarshal(raw, &msg); err == nil {
				switch msg.Header.Event {
				case constants.AudioTaskFinishedEvent:
					select {
					case senderErr := <-senderErrCh:
						if senderErr != nil && !errors.Is(senderErr, context.Canceled) {
							trySendTTSError(errCh, senderErr)
						}
					default:
					}
					return
				case constants.AudioTaskFailedEvent:
					trySendTTSError(errCh, errors.New(constants.ErrTTSFailed))
					return
				}
				continue
			}

			audioChunk := append([]byte(nil), raw...)
			select {
			case <-ctx.Done():
				return
			case audioCh <- audioChunk:
			}
		}
	}()

	return audioCh, errCh
}

func (s *service) sendTTSStart(conn *websocket.Conn, taskID, voiceID string) error {
	if strings.TrimSpace(voiceID) == "" {
		voiceID = s.cfg.TTS.Voice
	}

	payload := ttsRunPayload{
		TaskGroup: constants.AudioTaskGroup,
		Task:      constants.AudioTTSTask,
		Function:  constants.AudioTTSFunction,
		Model:     s.cfg.TTS.Model,
		Input:     map[string]any{},
	}
	payload.Parameters.TextType = "PlainText"
	payload.Parameters.Voice = voiceID
	payload.Parameters.Format = s.cfg.TTS.Format
	payload.Parameters.SampleRate = s.cfg.TTS.SampleRate
	payload.Parameters.Volume = s.cfg.TTS.Volume
	payload.Parameters.Rate = s.cfg.TTS.Rate
	payload.Parameters.Pitch = s.cfg.TTS.Pitch

	msg := wsEnvelope{
		Header: wsHeader{
			Action:    constants.AudioRunTaskAction,
			TaskID:    taskID,
			Streaming: constants.AudioStreamingDuplex,
		},
		Payload: payload,
	}
	return writeWSJSON(conn, msg)
}

func (s *service) sendTTSContent(conn *websocket.Conn, taskID, text string) error {
	msg := wsEnvelope{
		Header: wsHeader{
			Action:    constants.AudioContinueTaskAction,
			TaskID:    taskID,
			Streaming: constants.AudioStreamingDuplex,
		},
		Payload: map[string]any{
			"input": map[string]any{
				"text": text,
			},
		},
	}
	return writeWSJSON(conn, msg)
}

func (s *service) streamTTSSender(ctx context.Context, conn *websocket.Conn, taskID string, textCh <-chan string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-textCh:
			if !ok {
				return s.sendASRFinish(conn, taskID)
			}
			if strings.TrimSpace(text) == "" {
				continue
			}
			if err := s.sendTTSContent(conn, taskID, text); err != nil {
				_ = conn.Close()
				return err
			}
		}
	}
}

func trySendTTSError(errCh chan<- error, err error) {
	if err == nil {
		return
	}
	select {
	case errCh <- err:
	default:
	}
}
