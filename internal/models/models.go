package models

import "time"

type FindUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type Message struct {
	SenderLogin string
	Text        string
	SendTime    time.Time
}
