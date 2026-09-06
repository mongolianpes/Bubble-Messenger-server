package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type RequestAudio struct {
	UUID    string
	UserID  string
	Message []byte
}

type MessageAudioDialog struct {
	Time  int64
	Bytes []byte
}

var audioDialogs = map[string]*Dialog{}

func AudioDialogRequest(c echo.Context) error {
	var req RequestAudio
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	dialog, ok := audioDialogs[req.UUID]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dialog not found"})
	}

	now := time.Now().UnixNano()
	expire := int64(time.Second)

	dialog.Mu.Lock()
	dialog.LastUsedTime = now

	for user, msgs := range dialog.Users {
		filtered := msgs[:0]
		for _, m := range msgs {
			if now-m.Time <= expire {
				filtered = append(filtered, m)
			}
		}
		dialog.Users[user] = filtered
	}

	dialog.Users[req.UserID] = append(dialog.Users[req.UserID], MessageAudioDialog{
		Time:  now,
		Bytes: req.Message,
	})

	resp := make(map[string]map[int64][]byte)
	for user, msgs := range dialog.Users {
		if user == req.UserID {
			continue
		}
		if len(msgs) > 0 {
			m := make(map[int64][]byte, len(msgs))
			for _, msg := range msgs {
				m[msg.Time] = msg.Bytes
			}
			resp[user] = m
		}
	}

	for user := range resp {
		dialog.Users[user] = dialog.Users[user][:0]
	}

	dialog.Mu.Unlock()

	return c.JSON(http.StatusOK, resp)
}

func CheckLastUsedTimeInAudioDialog() {
	maxInactive := int64(15 * time.Second)
	sleep := time.Second * 20

	for {
		now := time.Now().UnixNano()
		for id, d := range audioDialogs {
			if now-d.LastUsedTime > maxInactive {
				delete(audioDialogs, id)
			}
		}
		time.Sleep(sleep)
	}
}
