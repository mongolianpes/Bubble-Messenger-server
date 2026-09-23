package handlers

import (
	"bubble/internal/crypto"
	"bubble/internal/profile"
	"fmt"
	"net/http"

	"bubble/internal/db"
	"bubble/internal/messages"

	"github.com/labstack/echo/v4"
)

type SendFileRequest struct {
	Sender               string
	Receiver             string
	SenderPassword       string
	FileName             string
	Device               string
	KeyForServerDataBase string
	File                 []byte
}

type Avatar struct {
	Avatar               []byte `json:"avatar"`
	Login                string `json:"login"`
	Password             string `json:"password"`
	DeviceInfo           string `json:"deviceinfo"`
	KeyForServerDataBase string `json:"forserver"`
}

func (h *Handler) SendFile(c echo.Context) error {
	var req SendFileRequest
	if err := c.Bind(&req); err != nil {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Sender == "" || req.Receiver == "" || req.SenderPassword == "" || req.FileName == "" || req.Device == "" || req.KeyForServerDataBase == "" || req.File == nil {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.Device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.Sender, err = crypto.StringDecrypt(req.Sender, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.SenderPassword, err = crypto.StringDecrypt(req.SenderPassword, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, req.Sender), req.SenderPassword) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", req.Sender, req.Device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	req.Receiver, err = crypto.StringDecrypt(req.Receiver, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	req.FileName, err = crypto.StringDecrypt(req.FileName, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.File, err = crypto.StringDecryptByte(req.File, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if err := messages.SendFile(req.Sender, req.Receiver, req.FileName, req.File); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос sendfile. Sender: %s, Receiver: %s, FileSize: %v. DeviceID: %s", req.Sender, req.Receiver, len(req.File), req.Device)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetFile(c echo.Context) error {
	device := c.FormValue("device")
	login := c.FormValue("login")
	password := c.FormValue("password")
	fileName := c.FormValue("filename")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || fileName == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), device)
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

	fileName, err = crypto.StringDecrypt(fileName, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, login), password) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", login, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	file, err := messages.GetFile(login, fileName)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	file, err = crypto.StringEncryptByte(file, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	usersRequestsLog.Printf("Запрос: getfile. Login %s, FileName: %s, DeviceID: %s", login, fileName, device)
	return c.JSON(http.StatusOK, file)
}

func (h *Handler) DelFile(c echo.Context) error {
	device := c.FormValue("device")
	login := c.FormValue("login")
	password := c.FormValue("password")
	fileName := c.FormValue("filename")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || fileName == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), device)
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

	fileName, err = crypto.StringDecrypt(fileName, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос delfile. Login: %s, FileName: %s, DeviceID: %s", login, fileName, device)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) SetAvatar(c echo.Context) error {
	var req Avatar
	if err := c.Bind(&req); err != nil {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	if req.DeviceInfo == "" || req.Login == "" || req.Password == "" || req.KeyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), req.DeviceInfo)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	req.Login, err = crypto.StringDecrypt(req.Login, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.Password, err = crypto.StringDecrypt(req.Password, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	req.Avatar, err = crypto.StringDecryptByte(req.Avatar, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, req.Login), req.Password) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", req.Login, req.DeviceInfo)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	if err := profile.SetAvatar(req.Login, req.Avatar); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	usersRequestsLog.Printf("Запрос setavatar. Login: %s, Размер картинки: %v, DeviceID: %s", req.Login, len(req.Avatar), req.DeviceInfo)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetAvatar(c echo.Context) error {
	device := c.FormValue("device")
	loginForSearch := c.FormValue("loginforsearch")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || loginForSearch == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, _, err := h.UsersService.GetAuthInfo(c.Request().Context(), device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	loginForSearch, err = crypto.StringDecrypt(loginForSearch, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	avatar, err := profile.GetAvatar(loginForSearch)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	avatar, err = crypto.StringEncryptByte(avatar, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос getavatar. Login поисковой картинки: %s, DeviceID: %s", loginForSearch, device)
	return c.JSON(http.StatusOK, avatar)
}
