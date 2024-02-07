package pkg

import (
	"log"
	"net/http"
)

type Executor interface {
	getExecutorRoute() string
	getExecutorHandler() func(http.ResponseWriter, *http.Request)
}

type App struct {
	AppName   string
	AppPort   string
	Executors []Executor
}

func (app *App) Run() {
	mux := http.NewServeMux()

	for _, executor := range app.Executors {
		mux.HandleFunc(executor.getExecutorRoute(), executor.getExecutorHandler())
		log.Println("[OK]: executor was init")
	}

	// Запускаем сервер
	go func() {
		log.Printf("[RUN]: Frontend service was successfully run. Name: %v, Port: %v\n", app.AppName, app.AppPort)
		if err := http.ListenAndServe(":"+app.AppPort, AddCorsHeaders(mux)); err != nil {
			log.Fatalln(err)
			return
		}
	}()
}

/*
addCorsHeaders подключает заголовки, без них
браузер может ругаться на веб страницу
*/
func AddCorsHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
