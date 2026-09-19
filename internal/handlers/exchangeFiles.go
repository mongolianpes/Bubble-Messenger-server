package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"server/internal/crypto"
	"time"

	"server/internal/db"

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

func SendFile(c echo.Context) error {
	var req SendFileRequest
	if err := c.Bind(&req); err != nil {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, "Invalid JSON")
	}

	if req.Sender == "" || req.Receiver == "" || req.SenderPassword == "" || req.FileName == "" || req.Device == "" || req.KeyForServerDataBase == "" || req.File == nil {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, req.Device[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства Login %s, Device %s", req.Sender, req.Device)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), req.KeyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		errRequestsLog.Printf("Сообщение клиента не удалось расшифровать. Device %s", req.Device)
		return c.String(http.StatusBadRequest, "Decrypted error")
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

	if _, err = os.Stat(fmt.Sprintf(db.UsersDir, req.Sender)); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("There is no sender with this login"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	if _, err = os.Stat(fmt.Sprintf(db.UsersDir, req.Receiver)); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("There is no receiver with this login"), key)
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

	now := time.Now()
	fileNameInDatabase := fmt.Sprintf(
		db.PathToUserUnreceivedFilesDir+"%s.%d%d%d%d%d%d",
		req.Receiver,
		req.FileName,
		now.Year(),
		now.YearDay(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
	)

	if err := os.WriteFile(fileNameInDatabase, req.File, db.ValuesAccessFile); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Failed to upload your file"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	pathToFileWithMessages := fmt.Sprintf(db.PathToUserMessagesDir, req.Receiver)

	messageData := Message{
		req.Sender,
		"p\\" + req.FileName + "\\" + fileNameInDatabase,
		now,
	}

	jsonMessageData, err := json.Marshal(messageData)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	saveMessagesManager.Write(pathToFileWithMessages, jsonMessageData)

	usersRequestsLog.Printf("Запрос sendfile. Sender: %s, Receiver: %s, FileSize: %v. DeviceID: %s", req.Sender, req.Receiver, len(req.File), req.Device)
	return c.NoContent(http.StatusOK)
}

func GetFileRequest(c echo.Context) error {
	device := c.FormValue("device")
	login := c.FormValue("login")
	password := c.FormValue("password")
	fileName := c.FormValue("filename")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || fileName == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, device[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства Login %s, Device %s", login, device)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		errRequestsLog.Printf("Сообщение клиента не удалось расшифровать. Device %s", device)
		return c.String(http.StatusBadRequest, "Decrypted error")
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

	file, err := os.ReadFile(fmt.Sprintf(db.PathToUserUnreceivedFilesDir+fileName, login))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Can not get this is file"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	file, err = crypto.StringEncryptByte(file, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Encrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос: getfile. Login %s, FileName: %s, DeviceID: %s", login, fileName, device)
	return c.JSON(http.StatusOK, file)
}

func DelFileRequest(c echo.Context) error {
	device := c.FormValue("device")
	login := c.FormValue("login")
	password := c.FormValue("password")
	fileName := c.FormValue("filename")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || fileName == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, device[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства Login %s, Device %s", login, device)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		errRequestsLog.Printf("Сообщение клиента не удалось расшифровать. Device %s", device)
		return c.String(http.StatusBadRequest, "Unknow error")
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

	if err := os.Remove(fmt.Sprintf(db.PathToUserUnreceivedFilesDir+fileName, login)); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Deleted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос delfile. Login: %s, FileName: %s, DeviceID: %s", login, fileName, device)
	return c.NoContent(http.StatusOK)
}

func SetAvatarRequest(c echo.Context) error {
	var req Avatar
	if err := c.Bind(&req); err != nil {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	if req.DeviceInfo == "" || req.Login == "" || req.Password == "" || req.KeyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, req.DeviceInfo[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства Login %s, Device %s", req.Login, req.DeviceInfo)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), req.KeyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		errRequestsLog.Printf("Сообщение клиента не удалось расшифровать. Device %s", req.DeviceInfo)
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusBadRequest, encryptResp)
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

	if err = os.WriteFile(fmt.Sprintf(db.PathToUserAvatar, req.Login), req.Avatar, db.ValuesAccessFile); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Failed to upload avatar"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	usersRequestsLog.Printf("Запрос setavatar. Login: %s, Размер картинки: %v, DeviceID: %s", req.Login, len(req.Avatar), req.DeviceInfo)
	return c.NoContent(http.StatusOK)
}

func GetAvatarRequest(c echo.Context) error {
	device := c.FormValue("device")
	loginForSearch := c.FormValue("loginforsearch")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || loginForSearch == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, device[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства LoginForSearch %s, Device %s", loginForSearch, device)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		errRequestsLog.Printf("Сообщение клиента не удалось расшифровать. Device %s", device)
		return c.String(http.StatusBadRequest, "Unknow error")
	}
	loginForSearch, err = crypto.StringDecrypt(loginForSearch, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	avatar, err := os.ReadFile(fmt.Sprintf(db.PathToUserAvatar, loginForSearch))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("The user does not have an avatar"), key)
		return c.String(http.StatusNoContent, encryptResp)
	}

	avatar, err = crypto.StringEncryptByte(avatar, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос getavatar. Login поисковой картинки: %s, DeviceID: %s", loginForSearch, device)
	return c.JSON(http.StatusOK, avatar)
}
