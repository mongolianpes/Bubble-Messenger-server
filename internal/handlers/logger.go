package handlers

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"server/internal/db"
)

var errServerRoutineLog *log.Logger
var errRequestsLog *log.Logger
var usersRequestsLog *log.Logger
var ServerRoutineLog *log.Logger
var CountInvalidRequests int = 0
var ErrEchoLog *log.Logger

func OpenLogFiles() {
	serverRoutineLogFile, err := os.OpenFile(db.PathToServerRountineLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	errServerRoutineLog = log.New(serverRoutineLogFile, "ERROR SERVER ROUTINE: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)
	ServerRoutineLog = log.New(serverRoutineLogFile, "Server routine: ", log.Ldate|log.Ltime|log.Lmicroseconds)

	requestsLogFile, err := os.OpenFile(db.PathToRequestsLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		errServerRoutineLog.Print("Не удалось открыть файл для логирования запросов")
	}
	errRequestsLog = log.New(requestsLogFile, "ERORR: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)
	usersRequestsLog = log.New(requestsLogFile, "Request: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)

	errEchoLogFile, err := os.OpenFile(db.PathToEchoErrLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		errServerRoutineLog.Print("Не удалось открыть файл для логирования ошибок Echo")
	}
	ErrEchoLog = log.New(errEchoLogFile, "", 0)
}

var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidStr(str string, isKey bool) bool {
	if !isKey && len(str) > 20 && len(str) < 6 {
		return false
	}

	return regexpSymbols.MatchString(str)
}
