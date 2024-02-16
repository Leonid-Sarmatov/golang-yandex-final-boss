package pkg

import (
	"log"
	"net/http"
	"sync"
	"time"
	//"orchestrator_server/pkg"
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
MessageManager определяет структуру содержащю общие
данные для исполнителей.
1. Структура содежит коннект к базе данных, который могут
все использовать для подключения к таблицам.
2. Структура содержит словарь, со временем исполнения
каждой операции, для того что бы каждый раз не запрашивать из из базы.
3. Структура содержит словарь, с информацией о зарегистрированных
в системе вычислителях.
4. Структура содержит кеш таблицы из базы, где зранятся все задачи,
что бы каждый раз не делать запрос в базу данных.
*/
type MessageManager struct {
	DbConnection     *DatabaseConnection
	OperationTimeMap map[string]int
	SolverInfoMap    map[string]*Solver
	Mutex            sync.Mutex
}

func NewMessageManager() (*MessageManager, error) {
	// Создаем менеджер
	var manager MessageManager
	manager.OperationTimeMap = make(map[string]int)
	manager.SolverInfoMap = make(map[string]*Solver)


	// Создаем коннект к базе данных
	dbConn, err := NewDatabaseConnection("host=postgres port=5432 user=leonid password=password dbname=main_database sslmode=disable")
	if err != nil {
		log.Fatalln(err)
		//log.Println(err)
		return nil, err
		//time.Sleep(1 * time.Second)
	}
	log.Println("i m here!")
	

	// Кладем ссылку на соединение в менеджер
	manager.DbConnection = dbConn

	// Пробуем получить настройки времени выполнения операций из базы данных
	timesOfOperation, err := manager.DbConnection.GetAllTimesOfOperation()
	isCorrect := true
	if err != nil {
		log.Println("[ERROR]: Can not get settings from operation_table")
		// Если не удалось, то заполняем значениями по умолчанию
		manager.SetDefaultTimesOfOperation()
	} else {
		// Проверяем корректность полученных данных из базы данных
		if len(timesOfOperation) > 4 {
			isCorrect = false
		}
		for _, val := range timesOfOperation {
			if val.Operation != "+" && val.Operation != "-" && val.Operation != "/" && val.Operation != "*" {
				isCorrect = false
				break
			}
		}
	}

	// Если корректно, то заполняем словарь со временем выпонения операций
	if isCorrect {
		for _, val := range timesOfOperation {
			manager.OperationTimeMap[val.Operation] = val.TimeOfOperation
		}
	}

	// Запускаем демон с проверкой разницы во времени рукопожатий сервером
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				manager.Mutex.Lock()
				for _, val := range manager.SolverInfoMap {
					if time.Now().Sub(val.LastPing) >= 60*time.Second {
						val.InfoString = "The server is not working"
					}

					if time.Now().Sub(val.LastPing) >= 2*time.Second {
						val.InfoString = "Solver is temporarily unavailable due to reboot"
					}
				}
				manager.Mutex.Unlock()
			}
		}
	}()

	// Возвращаем ссылку на менеджер сообщений
	return &manager, nil
}

func (manager *MessageManager) SetDefaultTimesOfOperation() {
	manager.OperationTimeMap["+"] = 1
	manager.OperationTimeMap["-"] = 1
	manager.OperationTimeMap["/"] = 1
	manager.OperationTimeMap["*"] = 1
}

/*
API определяет апи.
Хранит имя апи, порт на котором будут запущены исполнители,
а так же массив с исполнителями
*/
type API struct {
	APIName      string
	APIPort      string
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
