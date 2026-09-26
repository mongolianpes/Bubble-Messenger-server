package messenger

import (
	"bubble/internal/db"
	"bubble/internal/writer"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

func SendFile(sender, receiver, fileName string, file []byte) error {
	if _, err := os.Stat(fmt.Sprintf(db.UsersDir, sender)); err != nil {
		return errors.New("There is no sender with this login")
	}
	if _, err := os.Stat(fmt.Sprintf(db.UsersDir, receiver)); err != nil {
		return errors.New("There is no receiver with this login")
	}

	now := time.Now()
	fileNameInDatabase := fmt.Sprintf(
		db.PathToUserUnreceivedFilesDir+"%s.%d%d%d%d%d%d",
		receiver,
		fileName,
		now.Year(),
		now.YearDay(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
	)

	if err := os.WriteFile(fileNameInDatabase, file, db.ValuesAccessFile); err != nil {
		return errors.New("Failed to upload your file")
	}

	pathToFileWithMessages := fmt.Sprintf(db.PathToUserMessagesDir, receiver)

	messageData := UserMessage{
		Sender:   sender,
		Message:  "p\\" + fileName + "\\" + fileNameInDatabase,
		SendTime: now,
	}

	jsonMessageData, err := json.Marshal(messageData)
	if err != nil {
		return errors.New("Unable to process data")
	}

	writer.SaveMessagesManager.Write(pathToFileWithMessages, jsonMessageData)

	return nil
}

func GetFile(login, fileName string) ([]byte, error) {
	file, err := os.ReadFile(fmt.Sprintf(db.PathToUserUnreceivedFilesDir+fileName, login))
	if err != nil {
		return file, errors.New("Can not get this is file")
	}

	return file, nil
}

func DelFile(login, fileName string) error {
	if err := os.Remove(fmt.Sprintf(db.PathToUserUnreceivedFilesDir+fileName, login)); err != nil {
		return errors.New("Deleted error")
	}
	return nil
}
