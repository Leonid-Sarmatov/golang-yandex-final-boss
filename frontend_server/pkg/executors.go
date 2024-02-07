package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/Knetic/govaluate"
	"bytes"
)

// ***************** SiteUpExecutor *****************
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

// **************************************************

// *********** GetExpressionFromFirstPage ***********
type ExpressionJSON struct {
	Expression string `json:"expression"`
}

type ExpressionRequestJSON struct {
	Expression string `json:"expression"`
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

		// Пробуем распарсить строку
		_, err = govaluate.NewEvaluableExpression(message.Expression)
		if err != nil {
			http.Error(w, "[ERROR]: Can not parse expression: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Can not parse expression: " + err.Error())
			return 
		}

		// Если удалось успешно то пробуем отправить запрос на бэкенд
		requestToBack := ExpressionRequestJSON {
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
		resp, err := http.Post("http://localhost:8082/api/endpoint", "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			return
		}
		defer resp.Body.Close()

		log.Println("[OK]: Resive expression was successful")
	}
}

// **************************************************

// ********** GetListOfTasksFromSecondPage **********
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
		// Создаем отклик
		response := []TaskJSON{
			{
				ID:         0,
				Expression: "2+2",
				HashID:     "g7Yg56Ty",
				Status:     1,
				Result:     "",
				BeginTime:  time.Now(),
				EndTime:    time.Now(),
			},
			{
				ID:         1,
				Expression: "9*4+2",
				HashID:     "7kT63o",
				Status:     1,
				Result:     "",
				BeginTime:  time.Now(),
				EndTime:    time.Now(),
			},
		}

		// Конвертируем отклик в json-отклик
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON"+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON" + err.Error())
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		log.Println("[OK]: Send list of tasks was successful")
	}
}

// **************************************************

// ********* SendMessageWithTimeOfOperations ********
type CalculateTimesJSON struct {
	AdditionTime       string `json:"additionTime"`
	SubtractionTime    string `json:"subtractionTime"`
	DivisionTime       string `json:"divisionTime"`
	MultiplicationTime string `json:"multiplicationTime"`
}

type CalculateTimesToSendJSON struct {
	AdditionTime       time.Time `json:"additionTime"`
	SubtractionTime    time.Time `json:"subtractionTime"`
	DivisionTime       time.Time `json:"divisionTime"`
	MultiplicationTime time.Time `json:"multiplicationTime"`
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

		log.Printf("[OK]: Recive messsage with times was successfull: %v", message)
	}
}

// **************************************************

// ********* GetListOfSolversFromFourthPage ********
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
		// Создаем отклик
		response := []SolverJSON{
			{
				SolverName:           "Main Solver",
				SolvingNowExpression: "2+2",
				LastPing:             time.Now().Format("2006-01-02 15:04:05"),
				InfoString:           "Working 5 gorutines",
			},
			{
				SolverName:           "Reserv Solver",
				SolvingNowExpression: "9*4+2",
				LastPing:             time.Now().Format("2006-01-02 15:04:05"),
				InfoString:           "Was stoped",
			},
			{
				SolverName:           "Power Solver",
				SolvingNowExpression: "1-3/7+12*9",
				LastPing:             time.Now().Format("2006-01-02 15:04:05"),
				InfoString:           "Ready to launch",
			},
		}

		// Конвертируем отклик в json-отклик
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "[ERROR]: Can not encoding to JSON"+err.Error(), http.StatusInternalServerError)
			log.Println("[ERROR]: Can not encoding to JSON" + err.Error())
			return
		}

		// Заполняем тело запроса и заголовки
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		log.Println("[OK]: Send solvers list was successful")
	}
}

// **************************************************
