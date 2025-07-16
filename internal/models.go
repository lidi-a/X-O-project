package internal

import "time"

type Button struct {
	Text   string `json:"text"`
	Action string `json:"action"`
}

type IncomingMessage struct {
	UserID string  `json:"userId"`
	Text   *string `json:"text,omitempty"`
	Action *string `json:"action,omitempty"`
}

type OutgoingMessage struct {
	UserID  string   `json:"userId"`
	Text    string   `json:"text"`
	Buttons []Button `json:"buttons"`
}

type Game struct {
	ID        string
	PlayerX   string
	PlayerO   string
	Turn      string
	Winner    string
	Finished  bool
	Board     [3][3]string
	UpdatedAt time.Time
}
