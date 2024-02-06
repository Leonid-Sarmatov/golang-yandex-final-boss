package main

import (
	"log"
	"orchestrator_server/pkg"
	//"time"
)

func main() {
	//c, err := pkg.NewDatabaseConnection("postgres://postgres:password@postgres:5432/main_database?sslmode=disable")
	c, err := pkg.NewDatabaseConnection("host=localhost port=5432 user=leonid password=password dbname=main_database sslmode=disable")
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return
	}

	/*
	beginTime := time.Now()
	//endTime := beginTime.Add(1 * time.Second)
	err = c.AddTask(pkg.Task {
		Expression: "2+2*2",
		HashID: "fy7Yu8kR",
		Result: "",
		BeginTime: beginTime,
	})*/
				
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
}
