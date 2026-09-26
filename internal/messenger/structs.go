package messenger

import "time"

type UserMessage struct {
	Sender   string
	Message  string
	SendTime time.Time
}
