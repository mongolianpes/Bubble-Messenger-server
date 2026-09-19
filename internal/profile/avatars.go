package profile

import (
	"errors"
	"fmt"
	"os"
	"server/internal/db"
)

func SetAvatar(login string, avatar []byte) error {
	if err := os.WriteFile(fmt.Sprintf(db.PathToUserAvatar, login), avatar, db.ValuesAccessFile); err != nil {
		return errors.New("Failed to upload avatar")
	}

	return nil
}

func GetAvatar(loginForSearch string) ([]byte, error) {
	avatar, err := os.ReadFile(fmt.Sprintf(db.PathToUserAvatar, loginForSearch))
	if err != nil {
		return avatar, errors.New("The user does not have an avatar")
	}

	return avatar, nil
}
