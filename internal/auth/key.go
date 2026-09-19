package auth

import (
	"errors"
	"fmt"
	"os"
	"server/internal/crypto"
	"server/internal/db"
)

func GetKey(device, keyForServerDataBase string) (string, error) {
	fileData, err := os.ReadFile(fmt.Sprintf(db.IdsDir, device[:220]))
	if err != nil {
		return "", errors.New("This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		return "", errors.New("Unknown error")
	}
	return key, nil
}
