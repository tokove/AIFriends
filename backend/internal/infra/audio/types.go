package audio

type wsHeader struct {
	Action    string `json:"action,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	Streaming string `json:"streaming,omitempty"`
	Event     string `json:"event,omitempty"`
}

type wsEnvelope struct {
	Header  wsHeader `json:"header"`
	Payload any      `json:"payload"`
}

type asrRunPayload struct {
	Model      string `json:"model"`
	Parameters struct {
		SampleRate           int    `json:"sample_rate"`
		Format               string `json:"format"`
		TranscriptionEnabled bool   `json:"transcription_enabled"`
	} `json:"parameters"`
	Input     map[string]any `json:"input"`
	Task      string         `json:"task"`
	TaskGroup string         `json:"task_group"`
	Function  string         `json:"function"`
}

type asrEventPayload struct {
	Output struct {
		Transcription *struct {
			SentenceEnd bool   `json:"sentence_end"`
			Text        string `json:"text"`
		} `json:"transcription"`
	} `json:"output"`
}

type asrEventMessage struct {
	Header  wsHeader        `json:"header"`
	Payload asrEventPayload `json:"payload"`
}

type ttsRunPayload struct {
	TaskGroup  string `json:"task_group"`
	Task       string `json:"task"`
	Function   string `json:"function"`
	Model      string `json:"model"`
	Parameters struct {
		TextType   string  `json:"text_type"`
		Voice      string  `json:"voice"`
		Format     string  `json:"format"`
		SampleRate int     `json:"sample_rate"`
		Volume     int     `json:"volume"`
		Rate       float64 `json:"rate"`
		Pitch      float64 `json:"pitch"`
	} `json:"parameters"`
	Input map[string]any `json:"input"`
}
