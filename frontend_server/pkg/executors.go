package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

type FirstPageResponseJSON struct {
	Response   string    `json:"response"`
	TimeToSend time.Time `json:"timeToSend"`
}

type GetExpressionFromFirstPage struct{}

func NewGetExpressionFromFirstPage() *GetExpressionFromFirstPage {
	return &GetExpressionFromFirstPage{}
}

func (e *GetExpressionFromFirstPage) getExecutorRoute() string {
	return "/sendExpression"
}

func (e *GetExpressionFromFirstPage) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[POST]: Was resived expression")

		var message ExpressionJSON

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			http.Error(w, "[ERROR]: Decoding JSON was failed: "+err.Error(), http.StatusBadRequest)
			log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
			return
		}

		// Создаем отклик
		response := FirstPageResponseJSON{
			Response:   message.Expression,
			TimeToSend: time.Now(),
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
	return "/getListOfExpression"
}

func (e *GetListOfTasksFromSecondPage) getExecutorHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[GET]: Was resived list")

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
		// Пишем лог
		log.Println("[PUT]: Was resived times")

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

		log.Printf("[OK]: Recive messsage was successfull: %v", message)
	}
}
// **************************************************
