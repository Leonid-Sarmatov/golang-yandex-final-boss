package main

import (
	"absolute_solver/pkg"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	as := pkg.NewAbsoluleSolver()
	as.RunHandShakeStream()
	time.Sleep(5 * time.Second)
	log.Println("[INFO]: Solver was run!")
	as.RunSolverStream()

	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
	log.Println("[INFO]: Frontend service was stoped!")
}