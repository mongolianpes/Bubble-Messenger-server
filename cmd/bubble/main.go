package main

import (
	"fmt"
	"regexp"
	"server/internal/writer"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	secretServerSalt = "al2esad;famfopa14-,410-qu82304dfoaspddasdsdo934idsadasd342141das"
	valuesAccessFile = 0600
	valueAccessDir   = 0700
)

var audioDialogs = map[string]*Dialog{}
var saveMessagesManager = writer.NewFileWriterManager(2 * time.Minute)
var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type Message struct {
	Sender   string
	Message  string
	SendTime time.Time
}

type MessageAudioDialog struct {
	Time  int64
	Bytes []byte
}

func main() {
	fmt.Printf(`
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v%s
Start on ports: 23099, 23098, 23097
`, echo.Version)

	openLogFiles()
	loadPopularUsers()
	go checkStartRegAuthUsersTime()
	go checkLastUsedTimeInAudioDialog()

	go func() {
		mainService := echo.New()
		mainService.Logger.SetOutput(errEchoLog.Writer())
		mainService.Use(middleware.Recover())
		mainService.HideBanner = true
		mainService.POST("/exchangekey", exchangeKeyReqest)
		mainService.POST("/reg", regReqest)
		mainService.POST("/auth", authReqest)
		mainService.POST("/searchuser", searchUserRequest)
		mainService.POST("/sendmessage", sendMessageRequest)
		mainService.POST("/checkmessage", checkMessageRequest)

		if err := mainService.Start(":23099"); err != nil {
			errEchoLog.Printf("Ошибка основного сервиса: %s", err)
		}
	}()

	go func() {
		mediumSizeDataService := echo.New()
		mediumSizeDataService.Logger.SetOutput(errEchoLog.Writer())
		mediumSizeDataService.Use(middleware.Recover())
		mediumSizeDataService.HideBanner = true
		mediumSizeDataService.POST("/delmessages", delMessagesRequest)
		mediumSizeDataService.POST("/setavatar", setAvatarRequest)
		mediumSizeDataService.POST("/getavatar", getAvatarRequest)
		mediumSizeDataService.POST("/sendfile", sendFileRequest)
		mediumSizeDataService.POST("/getfile", getFileRequest)
		mediumSizeDataService.POST("/delfile", delFileRequest)

		if err := mediumSizeDataService.Start(":23098"); err != nil {
			errEchoLog.Printf("Ошибка сервиса принятия файлов: %s", err)
		}
	}()

	go func() {
		audioDialogService := echo.New()
		audioDialogService.Logger.SetOutput(errEchoLog.Writer())
		audioDialogService.Use(middleware.Recover())
		audioDialogService.HideBanner = true
		audioDialogService.POST("/audiodialog", audioDialogRequest)

		if err := audioDialogService.Start(":23097"); err != nil {
			errEchoLog.Printf("Ошибка сервиса аудио диалога: %s", err)
		}
	}()

	var countHours = 0
	for {
		time.Sleep(time.Hour)

		countHours += 1
		if countHours >= 22 {
			countHours = 0
			savePopularUsers()
			serverRoutineLog.Print("Сохранены популярные пользователи")
		}

		if countInvalidRequests > 0 {
			serverRoutineLog.Printf("За последний час невалидных запросов: %v", countInvalidRequests)
		}
	}
}

func isValidStr(str string, isKey bool) bool {
	if !isKey && len(str) > 20 && len(str) < 6 {
		return false
	}

	return regexpSymbols.MatchString(str)
}
