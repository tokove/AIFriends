package audio

import (
	"backend/internal/config"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const wsIOTimeout = 60 * time.Second

type Service interface {
	ASR(ctx context.Context, pcmData []byte) (string, error)
	TTS(ctx context.Context, textCh <-chan string, voiceID string) (<-chan []byte, <-chan error)
}

type service struct {
	cfg *config.AudioConfig
}

func NewService(cfg *config.AudioConfig) Service {
	return &service{cfg: cfg}
}

// 连接 ws
func (s *service) dial() (*websocket.Conn, error) {
	wsURL, err := url.Parse(s.cfg.WSURL)
	if err != nil {
		return nil, err
	}

	header := http.Header{
		"Authorization": []string{"Bearer " + s.cfg.APIKey},
	}

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  wsIOTimeout,
		EnableCompression: false,
	}

	conn, _, err := dialer.Dial(wsURL.String(), header)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func writeWSJSON(conn *websocket.Conn, msg any) error {
	if err := conn.SetWriteDeadline(time.Now().Add(wsIOTimeout)); err != nil {
		return err
	}
	return conn.WriteJSON(msg)
}

func writeWSBinary(conn *websocket.Conn, data []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(wsIOTimeout)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, data)
}

func readWSJSON(conn *websocket.Conn, v any) error {
	if err := conn.SetReadDeadline(time.Now().Add(wsIOTimeout)); err != nil {
		return err
	}
	return conn.ReadJSON(v)
}

func readWSMessage(conn *websocket.Conn) ([]byte, error) {
	if err := conn.SetReadDeadline(time.Now().Add(wsIOTimeout)); err != nil {
		return nil, err
	}
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return raw, nil
}
