package main

import (
	"fmt"
	"time"

	"server/internal/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	fmt.Printf(`
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v%s
Start on ports: 23099, 23098, 23097
`, echo.Version)

	handlers.OpenLogFiles()
	handlers.LoadPopularUsers()
	go handlers.CheckStartRegAuthUsersTime()
	go handlers.CheckLastUsedTimeInAudioDialog()

	go func() {
		mainService := echo.New()
		mainService.Logger.SetOutput(handlers.ErrEchoLog.Writer())
		mainService.Use(middleware.Recover())
		mainService.HideBanner = true
		mainService.POST("/exchangekey", handlers.ExchangeKeyReqest)
		mainService.POST("/reg", handlers.RegReqest)
		mainService.POST("/auth", handlers.AuthReqest)
		mainService.POST("/searchuser", handlers.SearchUserRequest)
		mainService.POST("/sendmessage", handlers.SendMessageRequest)
		mainService.POST("/checkmessage", handlers.CheckMessageRequest)

		if err := mainService.Start(":23099"); err != nil {
			handlers.ErrEchoLog.Printf("Ошибка основного сервиса: %s", err)
		}
	}()

	go func() {
		mediumSizeDataService := echo.New()
		mediumSizeDataService.Logger.SetOutput(handlers.ErrEchoLog.Writer())
		mediumSizeDataService.Use(middleware.Recover())
		mediumSizeDataService.HideBanner = true
		mediumSizeDataService.POST("/delmessages", handlers.DelMessagesRequest)
		mediumSizeDataService.POST("/setavatar", handlers.SetAvatarRequest)
		mediumSizeDataService.POST("/getavatar", handlers.GetAvatarRequest)
		mediumSizeDataService.POST("/sendfile", handlers.SendFile)
		mediumSizeDataService.POST("/getfile", handlers.GetFileRequest)
		mediumSizeDataService.POST("/delfile", handlers.DelFileRequest)

		if err := mediumSizeDataService.Start(":23098"); err != nil {
			handlers.ErrEchoLog.Printf("Ошибка сервиса принятия файлов: %s", err)
		}
	}()

	go func() {
		audioDialogService := echo.New()
		audioDialogService.Logger.SetOutput(handlers.ErrEchoLog.Writer())
		audioDialogService.Use(middleware.Recover())
		audioDialogService.HideBanner = true
		audioDialogService.POST("/audiodialog", handlers.AudioDialogRequest)

		if err := audioDialogService.Start(":23097"); err != nil {
			handlers.ErrEchoLog.Printf("Ошибка сервиса аудио диалога: %s", err)
		}
	}()

	var countHours = 0
	for {
		time.Sleep(time.Hour)

		countHours += 1
		if countHours >= 22 {
			countHours = 0
			handlers.SavePopularUsers()
			handlers.ServerRoutineLog.Print("Сохранены популярные пользователи")
		}

		if handlers.CountInvalidRequests > 0 {
			handlers.ServerRoutineLog.Printf("За последний час невалидных запросов: %v", handlers.CountInvalidRequests)
		}
	}
}
