package pkg

import (
	"net/http"
	"log"
)

/*
Executor определяет необходимые и достаточные методы
для структуры, которая будет представлять полезный функционал
данного апи
*/
type Executor interface {
	/*
		Возвращает полный маршрут исполнителя, по которому ему следует
		отправлять запрос

		Returns:
			string: Маршрут
	*/
	getExecutorRoute() string
	/*
		Возвращает функцию-обработчик, которая должна вызываться
		при обращении на маршрут исполнителя. При запуске апи
		функции-обработчики исполнителей будут записаны в http.ServeMux

		Returns:
			func(http.ResponseWriter, *http.Request): Функция-обработчик запроса
	*/
	getExecutorHandler() func(http.ResponseWriter, *http.Request)
}

/*
API определяет апи.
Хранит имя апи, порт на котором будут запущены исполнители,
а так же массив с исполнителями
*/
type API struct {
	APIName string
	APIPort string
	APIExecutors []Executor
}

/*
ApiRun Запускает приложение
*/
func (api *API) APIRun() {
	// Создаем мукс
	mux := http.NewServeMux()
	for _, executor := range api.APIExecutors {
		mux.HandleFunc(executor.getExecutorRoute(), executor.getExecutorHandler())
	}

	// Запускаем сервер
	go func() {
		log.Printf("[RUN] Server begin run. Name: %v, Port: %v\n", api.APIName, api.APIPort)
		if err := http.ListenAndServe(":"+api.APIPort, mux); err != nil {
			log.Fatalln(err)
			return
		}
	}()
}
