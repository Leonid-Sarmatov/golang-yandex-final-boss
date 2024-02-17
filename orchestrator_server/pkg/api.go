package pkg

import (
	"log"
	"net/http"
	"sync"
	"time"
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

2. Структура содержит канал, предназначенный для блокировки 
выдачи задачи, так как выдача задачи состоит из двух
запросов к таблицам, и при параллельных запросах к оркестратору
сожет возникнуть выдача одной задачи двум вычислителям и т. п.

3. Структура содержит словарь, со временем исполнения
каждой операции, для того что бы каждый раз не запрашивать из из базы.

4. Структура содержит словарь, с информацией о зарегистрированных
в системе вычислителях.

5. Структура содержит мутекс, для безопасного доступа к словарям
менеджера во время параллельных запросов
*/
type MessageManager struct {
	DbConnection     *DatabaseConnection
	DbLockChan       chan int
	OperationTimeMap map[string]int
	SolverInfoMap    map[string]*Solver
	Mutex            sync.Mutex
}

/*
NewMessageManager возвращает ссылку на новый менеджер сообщений
*/
func NewMessageManager() (*MessageManager, error) {
	// Создаем менеджер
	var manager MessageManager
	manager.OperationTimeMap = make(map[string]int)
	manager.SolverInfoMap = make(map[string]*Solver)
	manager.DbLockChan = make(chan int, 1)

	// Создаем коннект к базе данных
	dbConn, err := NewDatabaseConnection("host=postgres port=5432 user=leonid password=password dbname=main_database sslmode=disable")
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	log.Printf("[INFO]: Connect to database was successful")

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
					// Если рукопожатие нет очень долго
					if time.Now().Sub(val.LastPing) >= 10*time.Second {
						val.InfoString = "Solver is died"
						continue
					}

					// Если рукопожатие пропало
					if time.Now().Sub(val.LastPing) >= 2*time.Second {
						// Пишем что сервер недоступен
						val.InfoString = "The server is not working"
						// Задачу которую сервер решал, переводим в статус 1 (в обработке)
						// что бы сделает ее доступной для других вычислителей
						err = manager.DbConnection.UpdateStatusAndResultFromExpression(
							1, val.SolvingNowExpression, "")
						if err != nil {
							log.Println("[ERROR]: Database error: " + err.Error())
						}
					}
				}
				manager.Mutex.Unlock()
			}
		}
	}()

	// Возвращаем ссылку на менеджер сообщений
	return &manager, nil
}

/*
SetDefaultTimesOfOperation заполняет словарь со временем выполнения
операций настройками по умолчанию
*/
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
