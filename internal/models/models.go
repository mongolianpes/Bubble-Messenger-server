package models

import "time"

type FindUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type UserMessage struct {
	Sender   string
	Message  string
	SendTime time.Time
}
