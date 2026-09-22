package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"server/internal/crypto"
	"server/internal/db"
	"server/internal/messages"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	identificatorForInitAudioDialog           = "a\\"
	identificatorForStopAudioDialog           = "as\\"
	identificatorForSecondStepInitAudioDialog = "ai\\"
)

type Dialog struct {
	Users        map[string][]MessageAudioDialog
	LastUsedTime int64
	Mu           sync.RWMutex
}

type RequestCheckMessages struct {
	Gzip     bool
	Messages []byte
}

var audioDialogsMu sync.RWMutex

func (h *Handler) SendMessage(c echo.Context) error {
	var response string
	senderLogin := c.FormValue("senderlogin")
	senderPassword := c.FormValue("senderpassword")
	receiverLogin := c.FormValue("receiverlogin")
	message := c.FormValue("message")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if senderLogin == "" || senderPassword == "" || receiverLogin == "" || message == "" || device == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := h.UsersService.GetKey(c.Request().Context(), device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	senderLogin, err = crypto.StringDecrypt(senderLogin, key)
	if err != nil || senderLogin == "" {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	senderPassword, err = crypto.StringDecrypt(senderPassword, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	receiverLogin, err = crypto.StringDecrypt(receiverLogin, key)
	if err != nil || receiverLogin == "" {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	message, err = crypto.StringDecrypt(message, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, senderLogin), senderPassword) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", senderLogin, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	if message == identificatorForInitAudioDialog {
		var id string
		for {
			id = uuid.New().String()

			audioDialogsMu.RLock()
			_, exists := audioDialogs[id]
			audioDialogsMu.RUnlock()

			if !exists {
				break
			}
		}

		senderUser := rand.Text()
		receiverUser := rand.Text()

		message = identificatorForSecondStepInitAudioDialog + id + "\\" + receiverUser
		response = id + "\\" + senderUser

		dialog := &Dialog{
			Users: map[string][]MessageAudioDialog{
				senderUser:   make([]MessageAudioDialog, 0, 4),
				receiverUser: make([]MessageAudioDialog, 0, 4),
			},
			LastUsedTime: time.Now().UnixNano(),
		}

		audioDialogsMu.Lock()
		audioDialogs[id] = dialog
		audioDialogsMu.Unlock()
	} else if strings.HasPrefix(message, identificatorForStopAudioDialog) {
		dialogUUID := strings.Split(message, "\\")[1]
		delete(audioDialogs, dialogUUID)
		response = "Audio dialog stop"
	}

	usersRequestsLog.Printf("Запрос sendmessage. SenderLogin %s, ReceiverLogin %s. DeviceID: %s", senderLogin, receiverLogin, device)
	if response != "" {
		responseEncrypted, _ := crypto.StringEncrypt([]byte(response), key)
		return c.String(http.StatusOK, responseEncrypted)
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func (h *Handler) CheckMessage(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := h.UsersService.GetKey(c.Request().Context(), device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	login, err = crypto.StringDecrypt(login, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, login), password) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", login, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	statusGzip, blockMessages, err := messages.Check(login)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	messagesEncryptedByte, _ := crypto.StringEncryptByte([]byte(blockMessages), key)
	request := RequestCheckMessages{
		Gzip:     statusGzip,
		Messages: messagesEncryptedByte,
	}

	usersRequestsLog.Printf("Запрос checkmessages. Login: %s. UserID: %s", login, device)
	return c.JSON(http.StatusOK, request)
}

func (h *Handler) DelMessages(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := h.UsersService.GetKey(c.Request().Context(), device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	login, err = crypto.StringDecrypt(login, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, login), password) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", login, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	if err := messages.Del(login); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	return c.NoContent(http.StatusOK)
}
