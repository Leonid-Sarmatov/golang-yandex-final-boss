package pkg

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	//"github.com/Knetic/govaluate"
)

/*
AddArithmeticExpression принимает запрос с арифметическим
выражением, и возвращает ошибку если не удается распарсить
*/
type AddArithmeticExpression struct {
	Manager *MessageManager
}

func NewAddArithmeticExpression(manager *MessageManager) *AddArithmeticExpression {
	return &AddArithmeticExpression{
		Manager: manager,
	}
}

func (e *AddArithmeticExpression) getExecutorRoute() string {
	return "/addArithmeticExpression"
}

func (e *AddArithmeticExpression) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message ExpressionRequestJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: AddArithmeticExpression Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: AddArithmeticExpression Decoding JSON was failed: " + err.Error())
			return
		}

		task := TaskJSON{
			ID:         0,
			Expression: message.Expression,
			HashID:     "hash",
			Status:     1,
			Result:     "",
			BeginTime:  message.TimeToSend,
			EndTime:    message.TimeToSend.Add(1 * time.Second),
		}

		e.Manager.DbConnection.AddTask(task)
		w.WriteHeader(http.StatusOK)
		log.Println("[OK]: Write task to database was successful")
	}
}

/*
GetListExpressionsWithStatuses принимает запрос
и возвращает список со всеми задачами
*/
type GetListExpressionsWithStatuses struct {
	Manager *MessageManager
}

func NewGetListExpressionsWithStatuses(manager *MessageManager) *GetListExpressionsWithStatuses {
	return &GetListExpressionsWithStatuses{
		Manager: manager,
	}
}

func (e *GetListExpressionsWithStatuses) getExecutorRoute() string {
	return "/getListOfTasks"
}

func (e *GetListExpressionsWithStatuses) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получем от базы данных список со всеми задачами
		tasks, err := e.Manager.DbConnection.GetAllTasks()
		if err != nil {
			http.Error(w, "[ERROR]: GetListExpressionsWithStatuses Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: GetListExpressionsWithStatuses Can not encoding to JSON: " + err.Error())
			return
		}

		// Конвертируем отклик в json-отклик
		jsonResponse, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, "[ERROR]: GetListExpressionsWithStatuses Can not encoding to JSON"+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: GetListExpressionsWithStatuses Can not encoding to JSON" + err.Error())
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		log.Println("[OK]: Send task list was successful")
	}
}

/*
SetExecutionTimeOfOperations принимает запрос со списком
времени выполнения для каждой операции
*/
type SetTimeOfOperations struct {
	Manager *MessageManager
}

func NewSetTimeOfOperations(manager *MessageManager) *SetTimeOfOperations {
	return &SetTimeOfOperations{
		Manager: manager,
	}
}

func (e *SetTimeOfOperations) getExecutorRoute() string {
	return "/setExecutionTimeOfOperations"
}

func (e *SetTimeOfOperations) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message TimeOfOperationJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: SetTimeOfOperations Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: SetTimeOfOperations Decoding JSON was failed: " + err.Error())
			return
		}

		// Записываем время операций в словарь менеджера
		for key, val := range message.Times {
			e.Manager.OperationTimeMap[key] = val
		}

		// Записываем время операций в базу данных
		err = e.Manager.DbConnection.SetTimesToOperation(message)
		if err != nil {
			http.Error(w, "[ERROR]: Database error: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Database error: " + err.Error())
			return
		}

		log.Println("[OK]: Set operation time was successful")
	}
}

/*
GetReadyTaskToSolving принимает запрос с информацией
о вычислителе и возвращает задачу готовую к выполнению
вместе с информацией о времени выполнения арифметических операций
*/
type GetReadyTaskToSolving struct {
	Manager *MessageManager
}

func NewGetReadyTaskToSolving(manager *MessageManager) *GetReadyTaskToSolving {
	return &GetReadyTaskToSolving{
		Manager: manager,
	}
}

func (e *GetReadyTaskToSolving) getExecutorRoute() string {
	return "/getTaskToSolving"
}

func (e *GetReadyTaskToSolving) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message SolverRequestJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: GetReadyTaskToSolving Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: GetReadyTaskToSolving Decoding JSON was failed: " + err.Error())
			return
		}

		// Проверяем зарегистрирован ли вычислитель в системе
		if _, ok := e.Manager.SolverInfoMap[message.SolverName]; !ok {
			// Если нет, то регистрируем
			e.Manager.Mutex.Lock()
			e.Manager.SolverInfoMap[message.SolverName] = &Solver{
				SolverName:           message.SolverName,
				SolvingNowExpression: "None",
				LastPing:             time.Now(),
				InfoString:           "Registered",
			}
			e.Manager.Mutex.Unlock()
		}

		// Запрашиваем у базы данных список задач,
		// принятых, но не отданных вычистилелям (статус 1)
		tasks, err := e.Manager.DbConnection.GetTasksFromStatus(1)
		if err != nil {
			http.Error(w, "[ERROR]:GetReadyTaskToSolving Database error: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: GetReadyTaskToSolving Database error: " + err.Error())
			return
		}

		// Если список пуст, значит задач нет,
		// значит отказываем вычислителю в выдаче задачи
		if len(tasks) == 0 {
			http.Error(w, "[INFO]: GetReadyTaskToSolving Receipt of task denied", http.StatusInternalServerError)
			log.Println("[INFO]: GetReadyTaskToSolving Receipt of task denied")
			return
		}

		// Перем первую задачу в списке(почемы бы и нет), пробуем
		// изменить ее статус с 1 (принята в обработку) на 2 (отдана вычислителю)
		err = e.Manager.DbConnection.UpdateStatusFromExpression(2, tasks[0].Expression)
		if err != nil {
			http.Error(w, "[ERROR]: Database error: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Database error: " + err.Error())
			return
		}

		// Отдаем вычислителю задачу. Формируем JSON
		tastToSend := &TaskToSendToSolver{
			Expression: tasks[0].Expression,
			Times:      e.Manager.OperationTimeMap,
		}

		// Конвертируем отклик в json-отклик
		jsonResponse, err := json.Marshal(tastToSend)
		if err != nil {
			http.Error(w, "[ERROR]: GetReadyTaskToSolving Can not encoding to JSON"+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: GetReadyTaskToSolving Can not encoding to JSON" + err.Error())
			// Если что то пошло не так, то отнимаем ее у вычислителя
			e.Manager.DbConnection.UpdateStatusFromExpression(1, tasks[0].Expression)
			return
		}

		// Записываем в словарь о том какой вычислитель какую задачу выполняет
		e.Manager.Mutex.Lock()
		solver := e.Manager.SolverInfoMap[message.SolverName]
		solver.InfoString = "Working"
		solver.SolvingNowExpression = tasks[0].Expression
		e.Manager.SolverInfoMap[message.SolverName] = solver
		e.Manager.Mutex.Unlock()

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		log.Println("[OK]: Send expression to solver")
	}
}

/*
SetResultOfSolving принимает запрос с результатом, информацией
о вычислителе и ошибках, возникших при выполнении
*/
type SetResultOfSolving struct {
	Manager *MessageManager
}

func NewGetResultOfSolving(manager *MessageManager) *SetResultOfSolving {
	return &SetResultOfSolving{
		Manager: manager,
	}
}

func (e *SetResultOfSolving) getExecutorRoute() string {
	return "/setResultOfExpression"
}

func (e *SetResultOfSolving) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message ResultFromSolver
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: SetResultOfSolving Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: SetResultOfSolving Decoding JSON was failed: " + err.Error())
			return
		}

		// Проверяем на всякий случай есть ли посчитаное выражение в
		// базе данных, если нет, значит ответ не записываем
		tasks, err := e.Manager.DbConnection.GetTasksFromExpession(message.Expression)
		if err != nil || len(tasks) == 0 {
			http.Error(w, "[ERROR]: Database error: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Database error: " + err.Error())
			return
		}

		// Проверяем ответ на корректность. Если ответ
		// это пустая строка или статус код не 0 (ошибка на стороне вычислителя),
		// значит вычислитель оподливился, меняем статут задачи с 2 (отдана
		// вычислителю) на 4 (ошибка выполнения, передана обратно в обработку).
		// Отправляем вычислителю код 200, что бы он не пытался снова отправить ответ
		if message.Result == "" || message.Status != 0 {
			w.WriteHeader(http.StatusOK)
			log.Println("[ERROR]: Result in invalid: " + err.Error())
			e.Manager.DbConnection.UpdateStatusFromExpression(4, message.Expression)
			return
		}

		// Если все верно, то пробуем изменить статус задачи в
		// базе данных с 2 (отдана вычислителю) на 3 (успешно посчитано)
		// и записывать в базу данных результат
		err = e.Manager.DbConnection.UpdateStatusAndResultFromExpression(
			3, message.Expression, message.Result)
		if err != nil {
			http.Error(w, "[ERROR]: Database error: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Database error: " + err.Error())
			return
		}

		// Записываем в словарь о том какой вычислитель какую задачу выполняет
		e.Manager.Mutex.Lock()
		solver := e.Manager.SolverInfoMap[message.SolverName]
		solver.InfoString = "Free"
		solver.SolvingNowExpression = "None"
		e.Manager.SolverInfoMap[message.SolverName] = solver
		e.Manager.Mutex.Unlock()

		w.WriteHeader(http.StatusOK)
		log.Println("[OK]: Get result from solver successful")
	}
}

/*
GetListOfSolvers принимает запрос
и возвращает список с вычислителями и информацией о них
*/
type GetListOfSolvers struct {
	Manager *MessageManager
}

func NewGetListOfSolvers(manager *MessageManager) *GetListOfSolvers {
	return &GetListOfSolvers{
		Manager: manager,
	}
}

func (e *GetListOfSolvers) getExecutorRoute() string {
	return "/getListOfSolvers"
}

func (e *GetListOfSolvers) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем список вычислителей, и заполняем его
		// из словаря с информацией о вычислителях из менеджера
		solvers := make([]Solver, 0)
		for _, val := range e.Manager.SolverInfoMap {
			solvers = append(solvers, *val)
		}

		// Конвертируем отклик в json-отклик
		jsonResponse, err := json.Marshal(solvers)
		if err != nil {
			http.Error(w, "[ERROR]: GetListOfSolvers Can not encoding to JSON"+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: GetListOfSolvers Can not encoding to JSON" + err.Error())
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		log.Println("[OK]: Send solvers list was successful")
	}
}

/*
GetHandShake принимает запрос от
вычислителя с его именем для регулярного рукопожатия.
Если с последнего рукопожатия прошло более двух секунд,
считается что вычислитель на перезагрузке. Если с последнего
рукопожатия прошло более минуты, вычислитель считается мертвым
*/
type GetHandShake struct {
	Manager *MessageManager
}

func NewGetHandShake(manager *MessageManager) *GetHandShake {
	return &GetHandShake{
		Manager: manager,
	}
}

func (e *GetHandShake) getExecutorRoute() string {
	return "/solverHandShake"
}

func (e *GetHandShake) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message SolverRequestJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: GetHandShake Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: GetHandShake Decoding JSON was failed: " + err.Error())
			return
		}

		// Проверяем зарегистрирован ли вычислитель в системе
		if _, ok := e.Manager.SolverInfoMap[message.SolverName]; !ok {
			// Если нет, то регистрируем
			e.Manager.Mutex.Lock()
			e.Manager.SolverInfoMap[message.SolverName] = &Solver{
				SolverName:           message.SolverName,
				SolvingNowExpression: "None",
				LastPing:             time.Now(),
				InfoString:           "Registered",
			}
			e.Manager.Mutex.Unlock()
		}

		// Записываем в словарь время рукопожатия
		e.Manager.Mutex.Lock()
		solver := e.Manager.SolverInfoMap[message.SolverName]
		solver.LastPing = time.Now()
		e.Manager.SolverInfoMap[message.SolverName] = solver
		e.Manager.Mutex.Unlock()
	}
}
