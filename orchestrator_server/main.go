package main

import (
	"log"
	"orchestrator_server/pkg"
	"os"
	"os/signal"
	//"time"
)

func main() {
	log.Println("i m here!")
	// Создаем структуру общения между исполнителями
	menager, err := pkg.NewMessageManager()
	if err != nil {
		return
	}

	// Создаем апи
	api := pkg.API{
		APIName: "Orcestrator",
		APIPort: "8082",
		APIExecutors: []pkg.Executor{
			pkg.NewAddArithmeticExpression(menager),
			pkg.NewGetListExpressionsWithStatuses(menager),
			pkg.NewSetTimeOfOperations(menager),
			pkg.NewGetReadyTaskToSolving(menager),
			pkg.NewGetResultOfSolving(menager),
			pkg.NewGetListOfSolvers(menager),
			pkg.NewGetHandShake(menager),
		},
	}

	// Запускаем апи
	api.APIRun()

	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
	log.Println("[INFO]: Frontend service was stoped!")

	/*
		//c, err := pkg.NewDatabaseConnection("postgres://postgres:password@postgres:5432/main_database?sslmode=disable")
		c, err := pkg.NewDatabaseConnection("host=localhost port=5432 user=leonid password=password dbname=main_database sslmode=disable")
		if err != nil {
			log.Printf("[ERROR]: %v", err)
			return
		}


		beginTime := time.Now()
		//endTime := beginTime.Add(1 * time.Second)
		err = c.AddTask(pkg.Task {
			Expression: "2+2*2",
			HashID: "fy7Yu8kR",
			Result: "",
			BeginTime: beginTime,
		})

		if err != nil {
			log.Printf("[ERROR]: %v", err)
			return
		}

		t, err := c.GetAllTasks()
		if err != nil {
			log.Printf("[ERROR]: %v", err)
			return
		}

		log.Fatalln(t)

		log.Println("[OK]: Successful")
	*/
}
