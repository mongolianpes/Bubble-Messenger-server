package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"server/internal/crypto"
	"server/internal/crypto/ecdh"
	"server/internal/db"
	"strings"
	"time"
)

var registringUsers = map[string]activeUserInfo{}
var authUsers = map[string]activeUserInfo{}

type activeUserInfo struct {
	Key  string
	Time time.Time
}

var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidStr(str string, isKey bool) bool {
	if !isKey && len(str) > 20 && len(str) < 6 {
		return false
	}

	return regexpSymbols.MatchString(str)
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

func ExchangeKey(isRegistring bool, clientPublicKey, id string) (string, error) {
	sharedSecret, serverPublicKey, err := ecdh.GenerateKeys(clientPublicKey)
	if err != nil {
		return "", err
	}

	if isRegistring {
		if _, ok := registringUsers[id]; ok {
			return "", errors.New("User with this login registrating now, retry 10 minutes")
		}
	} else {
		if _, ok := authUsers[id]; ok {
			return "", errors.New("User with this login auth now, retry 10 minutes")
		}
	}

	if isRegistring {
		if len(registringUsers) > 15 {
			//ТУТ надо придумать что может лучше такие темки под каким то другим тегом писать
			return "", errors.New("Try registering in 3 minutes")
		}
		registringUsers[id] = activeUserInfo{
			Key:  sharedSecret,
			Time: time.Now(),
		}
	} else {
		if len(authUsers) > 15 {
			//ТУТ
			return "", errors.New("Please try logging in in 3 minutes")
		}
		authUsers[id] = activeUserInfo{
			Key:  sharedSecret,
			Time: time.Now(),
		}
	}

	return serverPublicKey, nil
}

func Register(login, name, password, device, keyForServerDataBase string) (string, error) {
	_, ok := registringUsers[device]
	if !ok {
		return "", errors.New("This request is not expected for you")
	}

	key := registringUsers[device].Key
	delete(registringUsers, device)

	login, err := crypto.StringDecrypt(login, key)
	if err != nil {
		return "", errors.New("Decryption error")
	}
	if !isValidStr(login, false) {
		return key, errors.New("Use only: a - z, A - Z or 0 - 9. And < 20")
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		return "", errors.New("Decryption error")
	}
	if len(password) <= 9 {
		return key, errors.New("The password must be more than 9 characters")
	}
	name, err = crypto.StringDecrypt(name, key)
	if err != nil {
		return "", errors.New("Decryption error")
	}
	keyForServerDataBase, err = crypto.StringDecrypt(keyForServerDataBase, key)
	if err != nil {
		return "", errors.New("Decryption error")
	}

	_, err = os.Stat(fmt.Sprintf(db.DevicesDir, device[:172]+device[220:]))
	if err == nil {
		decryptedLogins, err := crypto.FileStringDecrypt(fmt.Sprintf(db.DevicesDir, device[:172]+device[220:]), password, keyForServerDataBase)
		if err != nil {
			return key, errors.New("Unknown error1")
		}

		logins := strings.Split(decryptedLogins, "\n")
		if len(logins) >= 2 {
			return key, errors.New("You have registered many accounts")
		}

		if err = os.Mkdir(fmt.Sprintf(db.UsersDir, login), db.ValueAccessDir); err != nil {
			return key, errors.New("A user with this login is already registered")
		}

		if err = crypto.FileStringEncrypt([]byte(logins[0]+"\n"+login), fmt.Sprintf(db.DevicesDir, device[:172]+device[220:]), password, keyForServerDataBase); err != nil {
			os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
			return key, errors.New("Unknown error2")
		}
	} else {
		if err = os.Mkdir(fmt.Sprintf(db.UsersDir, login), db.ValueAccessDir); err != nil {
			return key, errors.New("A user with this login is already registered")
		}

		if err = crypto.FileStringEncrypt([]byte(login), fmt.Sprintf(db.DevicesDir, device[:172]+device[220:]), password, keyForServerDataBase); err != nil {
			os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
			return key, err
		}
	}

	if err = os.WriteFile(fmt.Sprintf(db.PathToUserName, login), []byte(name), db.ValuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Unable to process data")
	}

	hashedPassword, err := crypto.HashString(password, true)
	if err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Unable to process data")
	}
	if err = os.WriteFile(fmt.Sprintf(db.PathToUserPassword, login), []byte(hashedPassword), db.ValuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Unable to process data")
	}

	if err := os.Mkdir(fmt.Sprintf(db.PathToUserMessagesDir, login), db.ValueAccessDir); err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Failed to create messaging feature")
	}
	if err := os.Mkdir(fmt.Sprintf(db.PathToUserUnreceivedFilesDir, login), db.ValueAccessDir); err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Failed to create messaging feature")
	}

	encryptedKey, err := crypto.StringEncrypt([]byte(key), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Unknown error")
	}
	if err = os.WriteFile(fmt.Sprintf(db.IdsDir, device[:220]), []byte(encryptedKey), db.ValuesAccessFile); err != nil {
		os.RemoveAll(fmt.Sprintf(db.UsersDir, login))
		return key, errors.New("Unknown error")
	}

	return key, nil
}

func Auth(login, password, device, keyForServerDataBase string) (string, string, error) {
	_, ok := authUsers[device]
	if !ok {
		return "", "", errors.New("This request is not expected for you")
	}

	key := authUsers[device].Key
	login, err := crypto.StringDecrypt(login, key)
	if err != nil {
		return "", "", errors.New("Decryption error")
	}
	password, err = crypto.StringDecrypt(password, key)
	if err != nil {
		return "", "", errors.New("Decryption error")
	}
	keyForServerDataBase, err = crypto.StringDecrypt(keyForServerDataBase, key)
	if err != nil {
		return "", "", errors.New("Decryption error")
	}

	delete(authUsers, device)

	if !crypto.VerifyPassword(fmt.Sprintf(db.PathToUserPassword, login), password) {
		return "", key, errors.New("Incorrect login or password")
	}

	encryptKey, err := crypto.StringEncrypt([]byte(key), keyForServerDataBase+db.SecretServerSalt)
	if err != nil {
		return "", key, errors.New("Unknow error")
	}
	if err = os.WriteFile(fmt.Sprintf(db.IdsDir, device[:220]), []byte(encryptKey), db.ValuesAccessFile); err != nil {
		return "", key, errors.New("Unable to process data")
	}

	userNameByte, err := os.ReadFile(fmt.Sprintf(db.PathToUserName, login))
	if err != nil {
		return "", key, errors.New("Unknow error")
	}

	return string(userNameByte), key, nil
}
