package handlers

import (
	"fmt"
	"log"
	"os"
)

var errServerRoutineLog *log.Logger
var errRequestsLog *log.Logger
var usersRequestsLog *log.Logger
var ServerRoutineLog *log.Logger
var CountInvalidRequests int = 0
var ErrEchoLog *log.Logger

func OpenLogFiles() {
	serverRoutineLogFile, err := os.OpenFile(pathToServerRountineLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	errServerRoutineLog = log.New(serverRoutineLogFile, "ERROR SERVER ROUTINE: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)
	ServerRoutineLog = log.New(serverRoutineLogFile, "Server routine: ", log.Ldate|log.Ltime|log.Lmicroseconds)

	requestsLogFile, err := os.OpenFile(pathToRequestsLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		errServerRoutineLog.Print("Не удалось открыть файл для логирования запросов")
	}
	errRequestsLog = log.New(requestsLogFile, "ERORR: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)
	usersRequestsLog = log.New(requestsLogFile, "Request: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)

	errEchoLogFile, err := os.OpenFile(pathToEchoErrLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		errServerRoutineLog.Print("Не удалось открыть файл для логирования ошибок Echo")
	}
	ErrEchoLog = log.New(errEchoLogFile, "", 0)
}
