package messages

import (
	"bubble/internal/db"
	"bubble/internal/writer"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func Send(senderLogin, receiverLogin, message string) error {
	if _, err := os.Stat(fmt.Sprintf(db.UsersDir, senderLogin)); err != nil {
		return errors.New("There is no sender with this login")
	}
	if _, err := os.Stat(fmt.Sprintf(db.UsersDir, receiverLogin)); err != nil {
		return errors.New("There is no receiver with this login")
	}

	// далее все что закоментировано относитьяс к инициализации зашифрованного диалога
	// senderLoginHashed := c.FormValue("senderLoginHashed")
	var messageData UserMessage
	var pathToFileWithMessages string
	// if senderLoginHashed == "" {
	// senderLoginEncrypt, err := crypto.StringEncrypt([]byte(senderLogin), secretKeyForChiperLoginsOnServerDataBase)
	// if err != nil {
	// 	encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error"), key)
	// 	return c.String(http.StatusBadRequest, encryptResp)
	// }

	messageData = UserMessage{
		Sender:   senderLogin,
		Message:  message,
		SendTime: time.Now(),
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
		return errors.New("Unable to process data")
	}

	writer.SaveMessagesManager.Write(pathToFileWithMessages, jsonMessageData)

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

	return nil
}

func Check(login string) (bool, []byte, error) {
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

	blockMessages := []byte{}
	statusGzip := false

	blocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		return statusGzip, blockMessages, errors.New("Database callback error 1")
	}

	if len(blocksMessages) > 1 {
		pathToBlockMessages := fmt.Sprintf(db.PathToUserBlockMessagesArchive, login, blocksMessages[0].Name())
		blockMessages, err = os.ReadFile(pathToBlockMessages)
		if err != nil {
			return statusGzip, blockMessages, errors.New("Database callback error 2")
		}
		statusGzip = strings.HasSuffix(blocksMessages[0].Name(), ".gz")
	} else if len(blockMessages) == 1 {
		blockMessages, err = os.ReadFile(fmt.Sprintf(db.PathToUserNewMessagesFile, login))
		if err != nil {
			return statusGzip, blockMessages, errors.New("Database callback error 3")
		}
	}

	return statusGzip, blockMessages, nil
}

func Del(login string) error {
	countBlocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		return errors.New("Database callback error 1")
	}

	blocksMessages, err := os.ReadDir(fmt.Sprintf(db.PathToUserMessagesDir, login))
	if err != nil {
		return errors.New("Database callback error 2")
	}

	if len(countBlocksMessages) > 1 {
		pathToBlockMessages := fmt.Sprintf(db.PathToUserBlockMessagesArchive, login, blocksMessages[0].Name())
		if err := os.Remove(pathToBlockMessages); err != nil {
			return err
		}
	} else {
		if err := os.Remove(fmt.Sprintf(db.PathToUserNewMessagesFile, login)); err != nil {
			return errors.New("No new messages")
		}
	}

	return nil
}
