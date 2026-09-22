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

	hand, err := handlers.NewHand()
	if err != nil {
		panic(err)
	}

	handlers.OpenLogFiles()
	go handlers.CheckLastUsedTimeInAudioDialog()

	go func() {
		mainService := echo.New()
		mainService.Logger.SetOutput(handlers.ErrEchoLog.Writer())
		mainService.Use(middleware.Recover())
		mainService.HideBanner = true
		mainService.POST("/exchangekey", hand.TLS)
		mainService.POST("/reg", hand.Reg)
		mainService.POST("/auth", hand.Auth)
		mainService.POST("/searchuser", hand.SearchUser)
		mainService.POST("/sendmessage", hand.SendMessage)
		mainService.POST("/checkmessage", hand.CheckMessage)

		if err := mainService.Start(":23099"); err != nil {
			handlers.ErrEchoLog.Printf("Ошибка основного сервиса: %s", err)
		}
	}()

	go func() {
		mediumSizeDataService := echo.New()
		mediumSizeDataService.Logger.SetOutput(handlers.ErrEchoLog.Writer())
		mediumSizeDataService.Use(middleware.Recover())
		mediumSizeDataService.HideBanner = true
		mediumSizeDataService.POST("/delmessages", hand.DelMessages)
		mediumSizeDataService.POST("/setavatar", hand.SetAvatar)
		mediumSizeDataService.POST("/getavatar", hand.GetAvatar)
		mediumSizeDataService.POST("/sendfile", hand.SendFile)
		mediumSizeDataService.POST("/getfile", hand.GetFile)
		mediumSizeDataService.POST("/delfile", hand.DelFile)

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

	for {
		time.Sleep(time.Hour)

		if handlers.CountInvalidRequests > 0 {
			handlers.ServerRoutineLog.Printf("За последний час невалидных запросов: %v", handlers.CountInvalidRequests)
		}
	}
}
