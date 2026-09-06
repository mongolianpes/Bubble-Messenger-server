package handlers

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"server/internal/crypto"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var registringUsers = map[string]activeUserInfo{}
var authUsers = map[string]activeUserInfo{}

type activeUserInfo struct {
	Key  string
	Time time.Time
}

type KeyExchangeResponse struct {
	ServerPublicKey string `json:"public_key"`
}

type KeyExchangeRequest struct {
	ID              string `json:"id"`
	ClientPublicKey string `json:"public_key"`
	IsRegistring    bool   `json:"is_registring"`
}

func ExchangeKeyReqest(c echo.Context) error {
	var err error
	var req KeyExchangeRequest
	if err := c.Bind(&req); err != nil {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	if !isValidStr(req.ID, true) {
		CountInvalidRequests += 1
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid error"})
	}

	clientPubKeyBytesWithGarbage := req.ClientPublicKey
	clientPubKeyBytes, err := base64.RawURLEncoding.DecodeString(clientPubKeyBytesWithGarbage)
	if err != nil {
		errRequestsLog.Printf("exchangekey: Клиент отправил публичный ключ, который не удалось декодировать из base64. DeviceID: %s", req.ID)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid public key format"})
	}

	clientPublicKey, err := ecdh.P256().NewPublicKey(clientPubKeyBytes)
	if err != nil {
		errRequestsLog.Printf("exchangekey: Клиент отправил публичный ключ %s, который не удалось разобрать. DeviceID: %s", clientPublicKey.Bytes(), req.ID)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid public key"})
	}

	serverPrivateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		errRequestsLog.Printf("exchangekey: не удалось сгенерировать приватный ключ для Device %s", req.ID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate server key"})
	}

	sharedSecret, err := serverPrivateKey.ECDH(clientPublicKey)
	if err != nil {
		errRequestsLog.Printf("exchangekey: Клиент отправил публичный ключ %s, на основе которого не удалось вычислись общий секретный ключ. DeviceID: %s", req.ClientPublicKey, req.ID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute secret"})
	}

	if req.IsRegistring {
		if _, ok := registringUsers[req.ID]; ok {
			errRequestsLog.Printf("exchangekey: Клиент попытался зарегистрироваться, когда клиент с таким же ID уже регистрируется. DeviceID: %s", req.ID)
			return c.JSON(http.StatusConflict, map[string]string{"error": "User with this login registrating now, retry 10 minutes"})
		}
	} else {
		if _, ok := authUsers[req.ID]; ok {
			errRequestsLog.Printf("exchangekey: Клиент попытался авторизоваться, когда клиент с таким же ID уже авторизуется. DeviceID: %s", req.ID)
			return c.JSON(http.StatusConflict, map[string]string{"error": "User with this login auth now, retry 10 minutes"})
		}
	}

	if req.IsRegistring {
		if len(registringUsers) > 15 {
			//ТУТ надо придумать что может лучше такие темки под каким то другим тегом писать
			return c.JSON(http.StatusConflict, map[string]string{"error": "Try registering in 3 minutes"})
		}
		registringUsers[req.ID] = activeUserInfo{
			Key:  fmt.Sprintf("%x", sharedSecret),
			Time: time.Now(),
		}
	} else {
		if len(authUsers) > 15 {
			//ТУТ
			return c.JSON(http.StatusConflict, map[string]string{"error": "Please try logging in in 3 minutes."})
		}
		authUsers[req.ID] = activeUserInfo{
			Key:  fmt.Sprintf("%x", sharedSecret),
			Time: time.Now(),
		}
	}

	response := KeyExchangeResponse{
		ServerPublicKey: base64.RawURLEncoding.EncodeToString(serverPrivateKey.PublicKey().Bytes()),
	}

	usersRequestsLog.Printf("Получен запрос на обмен ключами. ClientPublicKey: %s, IsRegistring: %v. DeviceID: %s", req.ClientPublicKey, req.IsRegistring, req.ID)
	return c.JSON(http.StatusOK, response)
}

func RegReqest(c echo.Context) error {
	login := c.FormValue("login")
	name := c.FormValue("name")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if login == "" || name == "" || password == "" || device == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	_, ok := registringUsers[device]
	if !ok {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "This request is not expected for you")
	}

	key := registringUsers[device].Key
	delete(registringUsers, device)

	login, err := crypto.StringDecrypt(login, key)
	if err != nil {
		errRequestsLog.Printf("reg: Сообщение клиента не удалось расшифровать. Device %s", device)
		return c.String(http.StatusBadRequest, "Decryption error")
	}
	if !isValidStr(login, false) {
		errRequestsLog.Printf("reg: Клиент отправил невалидный логин. Login %s, Device: %s", login, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Use only: a - z, A - Z or 0 - 9. And < 20"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		return c.String(http.StatusBadRequest, "Decryption error")
	}
	if len(password) <= 9 {
		encryptResp, _ := crypto.StringEncrypt([]byte("The password must be more than 9 characters"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}
	name, err = crypto.StringDecrypt(name, key)
	if err != nil {
		return c.String(http.StatusBadRequest, "Decryption error")
	}
	keyForServerDataBase, err = crypto.StringDecrypt(keyForServerDataBase, key)
	if err != nil {
		return c.String(http.StatusBadRequest, "Dectyption error")
	}

	_, err = os.Stat(fmt.Sprintf(devicesDir, device[:172]+device[220:]))
	if err == nil {
		decryptedLogins, err := crypto.FileStringDecrypt(fmt.Sprintf(devicesDir, device[:172]+device[220:]), password, keyForServerDataBase)
		if err != nil {
			encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error1"), key)
			return c.String(http.StatusForbidden, encryptResp)
		}

		logins := strings.Split(decryptedLogins, "\n")
		if len(logins) >= 2 {
			errRequestsLog.Printf("reg: Клиент пытается зарегистрировать больше двух аккаунтов Deivce %s", device)
			encryptResp, _ := crypto.StringEncrypt([]byte("You have registered many accounts"), key)
			return c.String(http.StatusForbidden, encryptResp)
		}

		if err = os.Mkdir(fmt.Sprintf(usersDir, login), valueAccessDir); err != nil {
			encryptResp, _ := crypto.StringEncrypt([]byte("A user with this login is already registered"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}

		if err = crypto.FileStringEncrypt([]byte(logins[0]+"\n"+login), fmt.Sprintf(devicesDir, device[:172]+device[220:]), password, keyForServerDataBase); err != nil {
			os.RemoveAll(fmt.Sprintf(usersDir, login))
			return c.String(http.StatusBadRequest, "Unknown error2")
		}
	} else {
		if err = os.Mkdir(fmt.Sprintf(usersDir, login), valueAccessDir); err != nil {
			encryptResp, _ := crypto.StringEncrypt([]byte("A user with this login is already registered"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}

		if err = crypto.FileStringEncrypt([]byte(login), fmt.Sprintf(devicesDir, device[:172]+device[220:]), password, keyForServerDataBase); err != nil {
			os.RemoveAll(fmt.Sprintf(usersDir, login))
			encryptResp, _ := crypto.StringEncrypt([]byte(err.Error()), key)
			return c.String(http.StatusInternalServerError, encryptResp)
		}
	}

	if err = os.WriteFile(fmt.Sprintf(pathToUserName, login), []byte(name), valuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не записать имя пользователя в БД Name: %s, Device %s", name, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	hashedPassword, err := crypto.HashString(password, true)
	if err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось создать хэш пароля пользователя Login %s", login)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	if err = os.WriteFile(fmt.Sprintf(pathToUserPassword, login), []byte(hashedPassword), valuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось записать хэш пароля пользователя Login %s", login)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if err := os.Mkdir(fmt.Sprintf(pathToUserMessagesDir, login), valueAccessDir); err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось создать директорию для передачи сообщений Login %s", login)
		encryptResp, _ := crypto.StringEncrypt([]byte("Failed to create messaging feature"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}
	if err := os.Mkdir(fmt.Sprintf(pathToUserUnreceivedFilesDir, login), valueAccessDir); err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось создать директорию для передачи файлов Login %s", login)
		encryptResp, _ := crypto.StringEncrypt([]byte("Failed to create files feature"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	encryptedKey, err := crypto.StringEncrypt([]byte(key), keyForServerDataBase+secretServerSalt)
	if err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось создать ключ для шифрования в базе данных для директории ids Device %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	if err = os.WriteFile(fmt.Sprintf(idsDir, device[:220]), []byte(encryptedKey), valuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(usersDir, login))
		errRequestsLog.Printf("reg: Не удалось записать информацию в директорию ids Device %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	usersRequestsLog.Printf("Запрос: reg. Login: %s, UserName: %s. DeviceID: %s", login, name, device)
	return c.NoContent(http.StatusOK)
}

func AuthReqest(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if login == "" || password == "" || device == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	_, ok := authUsers[device]
	if !ok {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "This request is not expected for you")
	}

	key := authUsers[device].Key
	login, err := crypto.StringDecrypt(login, key)
	if err != nil {
		errRequestsLog.Printf("reg: Сообщение клиента не удалось расшифровать. Device %s", device)
		return c.String(http.StatusBadRequest, "Decryption error")
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		return c.String(http.StatusBadRequest, "Decryption error")
	}
	keyForServerDataBase, err = crypto.StringDecrypt(keyForServerDataBase, key)
	if err != nil {
		return c.String(http.StatusBadRequest, "Decryption error")
	}

	delete(authUsers, device)

	if !crypto.VerifyPassword(fmt.Sprintf(pathToUserPassword, login), password) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", login, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	encryptKey, err := crypto.StringEncrypt([]byte(key), keyForServerDataBase+secretServerSalt)
	if err != nil {
		errRequestsLog.Printf("reg: Не удалось зашифровать ключ для дальнейшего общения. Device %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unknow error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	if err = os.WriteFile(fmt.Sprintf(idsDir, device[:220]), []byte(encryptKey), valuesAccessFile); err != nil {
		errRequestsLog.Printf("reg: Не удалось записать ключ в БД для дальнейшего общения. Device %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	userNameByte, err := os.ReadFile(fmt.Sprintf(pathToUserName, login))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}
	encrypResp, _ := crypto.StringEncrypt(userNameByte, key)

	usersRequestsLog.Printf("Запрос: auth. Login %s, DeviceID: %s", login, device)
	return c.String(http.StatusOK, encrypResp)
}

func CheckStartRegAuthUsersTime() {
	durationDelete := time.Second * 12
	sleepTime := time.Second * 20
	for {
		for user := range registringUsers {
			durationStartToNow := time.Since(registringUsers[user].Time)
			if durationStartToNow > durationDelete {
				delete(registringUsers, user)
			}
		}
		for user := range authUsers {
			durationStartToNow := time.Since(authUsers[user].Time)
			if durationStartToNow > durationDelete {
				delete(authUsers, user)
			}
		}
		time.Sleep(sleepTime)
	}
}
