package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	//"github.com/Knetic/govaluate"
)

/*
SiteUpExecutor поднимает веб страницу
*/
type SiteUpExecutor struct{}

func NewSiteUpExecutor() *SiteUpExecutor {
	return &SiteUpExecutor{}
}

func (e *SiteUpExecutor) getExecutorRoute() string {
	return "/frontendSite"
}

func (e *SiteUpExecutor) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := "pkg/index.html"

		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("[ERROR]: %v", err)
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}

		fileSize := fileInfo.Size()
		buffer := make([]byte, fileSize)

		_, err = file.Read(buffer)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(buffer))
	}
}

/*
SendExpressionFromFirstPage принимает запрос с выражением
и отправляет запрос(с выражением) на сервер-оркестратор
*/
type ExpressionJSON struct {
	Expression string `json:"expression"`
}

type ExpressionRequestJSON struct {
	Expression string    `json:"expression"`
	TimeToSend time.Time `json:"timeToSend"`
}

type SendExpressionFromFirstPage struct{}

func NewGetExpressionFromFirstPage() *SendExpressionFromFirstPage {
	return &SendExpressionFromFirstPage{}
}

func (e *SendExpressionFromFirstPage) getExecutorRoute() string {
	return "/sendExpression"
}

func (e *SendExpressionFromFirstPage) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем тело запроса в JSON нужной нам структуры
		var message ExpressionJSON
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
			return
		}

		// Проверяем валидность выражения
		isValid := isValidExpression(message.Expression)
		if !isValid {
			http.Error(w, "[ERROR]: Can not parse expression", http.StatusBadRequest)
			log.Println("[ERROR]: Can not parse expression")
			return
		}

		// Если удалось успешно то пробуем отправить запрос на бэкенд
		requestToBack := ExpressionRequestJSON{
			Expression: message.Expression,
			TimeToSend: time.Now(),
		}

		// Формируем JSON
		jsonRequest, err := json.Marshal(requestToBack)
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			return
		}

		// Пробует отправить запрос на бэк
		resp, err := http.Post("http://orchestrator_server:8082/addArithmeticExpression", "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			http.Error(w, "[ERROR]: Can not send JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not send JSON: " + err.Error())
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(http.StatusOK)
		log.Println("[OK]: Resive expression was successful")
	}
}

func isValidExpression(expr string) bool {
	pattern1 := regexp.MustCompile(`[\d\+\-\*/]`)
	pattern2 := regexp.MustCompile(`[\+\-\*/]`)
	arr := strings.Split(expr, "")
	for i, ch := range arr {
		if !pattern1.MatchString(ch) {
			return false
		}

		if (i != len(expr)-1) &&
			pattern2.MatchString(arr[i]) &&
			pattern2.MatchString(arr[i+1]) {
			return false
		}

		if i == 0 && (ch == "*" || ch == "/") {
			return false
		}

		if (i == len(expr)-1) && pattern2.MatchString(ch) {
			return false
		}
	}

	return true
}

/*
GetListOfTasksFromSecondPage принимает запрос
и возвращает список со всеми задачами
*/
type TaskJSON struct {
	ID         int       `json:"id"`
	Expression string    `json:"expression"`
	HashID     string    `json:"hashId"`
	Status     int       `json:"status"`
	Result     string    `json:"result"`
	BeginTime  time.Time `json:"beginTime"`
	EndTime    time.Time `json:"endTime"`
}

type GetListOfTasksFromSecondPage struct{}

func NewGetListOfTasksFromSecondPage() *GetListOfTasksFromSecondPage {
	return &GetListOfTasksFromSecondPage{}
}

func (e *GetListOfTasksFromSecondPage) getExecutorRoute() string {
	return "/getListOfTask"
}

func (e *GetListOfTasksFromSecondPage) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Пробует отправить запрос на бэк для получения списка задач
		resp, err := http.Get("http://orchestrator_server:8082/getListOfTasks")
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			return
		}
		defer resp.Body.Close()

		// Вытаскиваем тело из ответа, в котором зашифрован JSON
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response from Server 2", http.StatusInternalServerError)
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)

		log.Println("[OK]: Send list of tasks was successful")
	}
}

/*
SendMessageWithTimeOfOperations принимает запрос от
веб страницы со временем выполнения для операций,
и отправляет запрос(со временем выполнения для
операций) на сервер-оркестратор
*/
type CalculateTimesJSON struct {
	AdditionTime       string `json:"additionTime"`
	SubtractionTime    string `json:"subtractionTime"`
	DivisionTime       string `json:"divisionTime"`
	MultiplicationTime string `json:"multiplicationTime"`
}

type TimeOfOperationJSON struct {
	Times map[string]int `json:"times"`
}

type SendMessageWithTimeOfOperations struct{}

func NewSendMessageWithTimeOfOperations() *SendMessageWithTimeOfOperations {
	return &SendMessageWithTimeOfOperations{}
}

func (e *SendMessageWithTimeOfOperations) getExecutorRoute() string {
	return "/sendTimeOfOperations"
}

func (e *SendMessageWithTimeOfOperations) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Сообщение от фронта, содержащее задержки для каждой операции
		var message CalculateTimesJSON

		// Декодируем тело запроса в сообщение
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
			return
		}

		// Переводим строки в числовые значения времени в секундах
		add, err := strconv.Atoi(message.AdditionTime)
		if err != nil {
			http.Error(w, "[ERROR]: Convert string to int was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Convert string to int was failed: " + err.Error())
			return
		}

		sub, err := strconv.Atoi(message.SubtractionTime)
		if err != nil {
			http.Error(w, "[ERROR]: Convert string to int was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Convert string to int was failed: " + err.Error())
			return
		}

		div, err := strconv.Atoi(message.DivisionTime)
		if err != nil {
			http.Error(w, "[ERROR]: Convert string to int was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Convert string to int was failed: " + err.Error())
			return
		}

		mul, err := strconv.Atoi(message.MultiplicationTime)
		if err != nil {
			http.Error(w, "[ERROR]: Convert string to int was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Convert string to int was failed: " + err.Error())
			return
		}

		request := TimeOfOperationJSON{
			map[string]int{
				"+": add,
				"-": sub,
				"/": div,
				"*": mul,
			},
		}

		// Формируем JSON
		jsonRequest, err := json.Marshal(request)
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			return
		}

		// Пробует отправить запрос на бэк
		resp, err := http.Post("http://orchestrator_server:8082/setExecutionTimeOfOperations", "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			http.Error(w, "[ERROR]: Can not send JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not send JSON: " + err.Error())
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(http.StatusOK)
		log.Println("[OK]: Resive expression was successful")

		log.Printf("[OK]: Recive messsage with times was successfull: %v", message)
	}
}

/*
GetListOfSolversFromFourthPage принимает запрос
и возвращает список с информацией о вычислителях
*/
type SolverJSON struct {
	SolverName           string `json:"solverName"`
	SolvingNowExpression string `json:"solvingExpression"`
	LastPing             string `json:"lastPing"`
	InfoString           string `json:"infoString"`
}

type GetListOfSolversFromFourthPage struct{}

func NewGetListOfSolversFromFourthPage() *GetListOfSolversFromFourthPage {
	return &GetListOfSolversFromFourthPage{}
}

func (e *GetListOfSolversFromFourthPage) getExecutorRoute() string {
	return "/getListOfSolvers"
}

func (e *GetListOfSolversFromFourthPage) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Пробует отправить запрос на бэк для получения списка вычислителей
		resp, err := http.Get("http://orchestrator_server:8082/getListOfSolvers")
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			return
		}
		defer resp.Body.Close()

		// Вытаскиваем тело из ответа, в котором зашифрован JSON
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response from server", http.StatusInternalServerError)
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)

		log.Println("[OK]: Send list of solvers was successful")
	}
}
