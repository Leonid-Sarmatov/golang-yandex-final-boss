package main

import (
	"frontend_server/pkg"
	"log"
	"os"
	"os/signal"
)



func main() {
	app := pkg.App{
		AppName: "Front server",
		AppPort: "8081",
		Executors: []pkg.Executor{
			pkg.NewSiteUpExecutor(),
			pkg.NewGetExpressionFromFirstPage(),
			pkg.NewGetListOfTasksFromSecondPage(),
			pkg.NewSendMessageWithTimeOfOperations(),
			pkg.NewGetListOfSolversFromFourthPage(),
		},
	}

	app.Run()
	
	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
	log.Println("[INFO]: Frontend service was stoped!")
}
