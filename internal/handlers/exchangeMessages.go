package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"server/internal/auth"
	"server/internal/crypto"
	"server/internal/db"
	"server/internal/writer"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	identificatorForInitAudioDialog           = "a\\"
	identificatorForStopAudioDialog           = "as\\"
	identificatorForSecondStepInitAudioDialog = "ai\\"
)

type RequestCheckMessages struct {
	Gzip     bool
	Messages []byte
}

type Dialog struct {
	Users        map[string][]MessageAudioDialog
	LastUsedTime int64
	Mu           sync.RWMutex
}

type Message struct {
	Sender   string
	Message  string
	SendTime time.Time
}

var audioDialogsMu sync.RWMutex
var saveMessagesManager = writer.NewFileWriterManager(2 * time.Minute)

func SendMessageRequest(c echo.Context) error {
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

	key, err := auth.GetKey(device, keyForServerDataBase)
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

	if _, err = os.Stat(fmt.Sprintf(db.UsersDir, senderLogin)); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("There is no sender with this login"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}
	if _, err = os.Stat(fmt.Sprintf(db.UsersDir, receiverLogin)); err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("There is no receiver with this login"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, senderLogin), senderPassword) {
		errRequestsLog.Printf("Неверный пароль Login %s, Device %s", senderLogin, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Incorrect login or password"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	// далее все что закоментировано относитьяс к инициализации зашифрованного диалога
	// senderLoginHashed := c.FormValue("senderLoginHashed")
	var messageData Message
	var pathToFileWithMessages string
	// if senderLoginHashed == "" {
	// senderLoginEncrypt, err := crypto.StringEncrypt([]byte(senderLogin), secretKeyForChiperLoginsOnServerDataBase)
	// if err != nil {
	// 	encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error"), key)
	// 	return c.String(http.StatusBadRequest, encryptResp)
	// }

	messageData = Message{
		senderLogin,
		message,
		time.Now(),
	}

	pathToFileWithMessages = fmt.Sprintf(db.PathToUserMessagesDir, receiverLogin)

	// подумать о размере, потому что можно сжимать даже если отправляется из файла newMessages (сделать размер вместо 800 намного больше, так чтобы в сжатом состоянии 1000)
	// if statsNewMessages, err := os.Stat(pathToFileWithMessages); err == nil {
	// 	if statsNewMessages.Size() > 800 {
	// 		timeNow := time.Now()
	// 		numBlockMessages := fmt.Sprintf("users/%s/messages/%v", receiverLogin, fmt.Sprintf("%v.%v.%v.%v.%v.%v.%v", timeNow.Year(), int(timeNow.Month()), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), timeNow.Second(), timeNow.Nanosecond()))
	//
	// 		for {
	// 			if _, err := os.Stat(numBlockMessages); !os.IsNotExist(err) {
	// 				pathToFileWithMessages += "1"
	// 			} else {
	// 				break
	// 			}
	// 		}
	//
	// 		if err = os.Rename(pathToFileWithMessages, numBlockMessages); err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 2"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}
	//
	// 		blockMessages, err := os.ReadFile(numBlockMessages)
	// 		if err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 3"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}
	//
	// 		blockMessages, err = CompressGzip(blockMessages)
	// 		if err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error (gzip)"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}
	//
	// 		if err = os.WriteFile(numBlockMessages, blockMessages, valuesAccess); err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 5"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}
	// 	}
	// }
	//
	// } else {
	// 	messageData = Message{
	// 		senderLoginHashed,
	// 		message,
	// 		time.Now(),
	// 	}
	//
	// 	pathToFileWithMessages = fmt.Sprintf("users/%s/messages/newMessages", receiverLogin)
	// }

	jsonMessageData, err := json.Marshal(messageData)
	if err != nil {
		errRequestsLog.Printf("Запрос sendmessage. Не удалось сформировать объект для сохранения сообщения: %s Login %s, Device %s", err.Error(), senderLogin, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	saveMessagesManager.Write(pathToFileWithMessages, jsonMessageData)

	// if _, err = os.ReadFile(pathToFileWithMessages); err != nil {
	// 	if err = os.WriteFile(pathToFileWithMessages, jsonMessageData, valuesAccess); err != nil {
	// 		encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
	// 		return c.String(http.StatusInternalServerError, encryptResp)
	// 	}
	// 	if response != "" {
	//
	// 		responseEncrypted, _ := crypto.StringEncrypt([]byte(response), key)
	// 		return c.String(http.StatusOK, responseEncrypted)
	// 	} else {
	// 		return c.NoContent(http.StatusOK)
	// 	}
	// }
	//
	// file, err := os.OpenFile(pathToFileWithMessages, os.O_APPEND|os.O_WRONLY, valuesAccess)
	// if err != nil {
	// 	encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
	// 	return c.String(http.StatusInternalServerError, encryptResp)
	// }
	// defer file.Close()
	//
	// _, err = file.WriteString("\n" + string(jsonMessageData))
	// if err != nil {
	// 	encryptResp, _ := crypto.StringEncrypt([]byte("Unable to process data"), key)
	// 	return c.String(http.StatusInternalServerError, encryptResp)
	// }

	usersRequestsLog.Printf("Запрос sendmessage. SenderLogin %s, ReceiverLogin %s. DeviceID: %s", senderLogin, receiverLogin, device)
	if response != "" {
		responseEncrypted, _ := crypto.StringEncrypt([]byte(response), key)
		return c.String(http.StatusOK, responseEncrypted)
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func CheckMessageRequest(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := auth.GetKey(device, keyForServerDataBase)
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

	// далее все что закоментировано относиться к инициализации зашифрованного диалога
	// dataMessages, _ := os.ReadFile(fmt.Sprintf("users/%s/messages/newMessages", login))

	// var strInitDialogs string
	// if string(dataInitDialogs) != "" {
	// 	initDialogsSplit := strings.Split(string(dataInitDialogs), "\n")

	// 	for _, initDialog := range initDialogsSplit {
	// 		var jsonInitDialog Message
	// 		err = json.Unmarshal([]byte(initDialog), &jsonInitDialog)
	// 		if err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error code 1"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}

	// 		// jsonInitDialog.Sender, err = crypto.StringDecrypt(jsonInitDialog.Sender, secretKeyForChiperLoginsOnServerDataBase)
	// 		// if err != nil {
	// 		// 	encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error code 2"), key)
	// 		// 	return c.String(http.StatusInternalServerError, encryptResp)
	// 		// }

	// 		initDialogDecrypted, err := json.Marshal(jsonInitDialog)
	// 		if err != nil {
	// 			encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error code 3"), key)
	// 			return c.String(http.StatusInternalServerError, encryptResp)
	// 		}

	// 		strInitDialogs += "\n" + string(initDialogDecrypted)
	// 	}
	// }

	// resultMessages := string(dataMessages) //+ strInitDialogs
	// resultMessages := strInitDialogs

	blocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 1"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	var blockMessages []byte
	var statusGzip = false
	if len(blocksMessages) > 1 {
		pathToBlockMessages := fmt.Sprintf(db.PathToUserBlockMessagesArchive, login, blocksMessages[0].Name())
		blockMessages, err = os.ReadFile(pathToBlockMessages)
		if err != nil {
			usersRequestsLog.Printf("Запрос checkmessages. Номер блока сообщений к отпрвке: %s, Login: %s, DeviceID: %s", blocksMessages[0].Name(), login, device)
			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 2"), key)
			return c.String(http.StatusInternalServerError, encryptResp)
		}
		statusGzip = strings.HasSuffix(blocksMessages[0].Name(), ".gz")
	} else if len(blockMessages) == 1 {
		blockMessages, err = os.ReadFile(fmt.Sprintf(db.PathToUserNewMessagesFile, login))
		if err != nil {
			usersRequestsLog.Printf("Запрос checkmessages. Номер блока сообщений к отпрвке: newMessages, Login: %s, DeviceID: %s", login, device)
			encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 3"), key)
			return c.String(http.StatusInternalServerError, encryptResp)
		}
	}

	messagesEncryptedByte, _ := crypto.StringEncryptByte([]byte(blockMessages), key)

	request := RequestCheckMessages{
		Gzip:     statusGzip,
		Messages: messagesEncryptedByte,
	}

	usersRequestsLog.Printf("Запрос checkmessages. Login: %s. UserID: %s", login, device)
	return c.JSON(http.StatusOK, request)
}

func DelMessagesRequest(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	device := c.FormValue("device")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || login == "" || password == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := auth.GetKey(device, keyForServerDataBase)
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

	countBlocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 1"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	blocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Database callback error 2"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if len(countBlocksMessages) > 1 {
		pathToBlockMessages := fmt.Sprintf(db.PathToUserBlockMessagesArchive, login, blocksMessages[0].Name())
		if err := os.Remove(pathToBlockMessages); err != nil {
			errRequestsLog.Printf("Запрос delmessages. Ошибка в блоках сообщений. Номер блока сообщений к удалению: %s Login %s, Device %s", blocksMessages[0].Name(), login, device)
			encryptResp, _ := crypto.StringEncrypt([]byte("database error: "+err.Error()), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}
	} else {
		if err := os.Remove(fmt.Sprintf(db.PathToUserNewMessagesFile, login)); err != nil {
			encryptResp, _ := crypto.StringEncrypt([]byte("No new messages"), key)
			return c.String(http.StatusBadRequest, encryptResp)
		}
	}

	usersRequestsLog.Printf("Запрос delmessages. Номер блока сообщений к удалению: %s, Login: %s, DeviceID: %s", blocksMessages[0].Name(), login, device)
	return c.NoContent(http.StatusOK)
}
