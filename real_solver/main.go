package main

import (
	"log"
	"os"
	"os/signal"
	"real_solver/pkg"
	"time"

	//"fmt"
	//"strconv"
	//"strings"
	//"sync"
)

func main() {
	time.Sleep(5 * time.Second)

	app := pkg.NewApp("Solver", 3)
	app.AppRun()

	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
	log.Println("[INFO]: Frontend service was stoped!")
}