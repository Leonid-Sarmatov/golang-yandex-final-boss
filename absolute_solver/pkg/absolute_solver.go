package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Knetic/govaluate"
)

/*
TaskToSendToSolver описывает структуру задачи,
которая будет отправлена вычислителю, если
он задачу запросит. Включает в себя само выражение
и словарь со временем выполнения для операций
*/
type TaskToSendToSolver struct {
	Expression string         `json:"expression"`
	Times      map[string]int `json:"times"`
}

/*
ResultFromSolver описывает JSON с результатом
вычислений. Такую структуру запроса должен
иметь вычислитель, желающий отправить ответ.
Включает в себя выражение, ответ и сообщение с
ошибками, комментарием от вычислителя и т. п.
используется в исполнителе SetResultOfSolving
*/
type ResultFromSolver struct {
	SolverName string `json:"solverName"`
	Expression string `json:"expression"`
	Result     string `json:"result"`
	Status     int    `json:"status"`
}

/*
SolverRequestJSON описывает JSON запроса вычислителя
на сервер. Такую структуру должен содержать запрос,
для регулярного рукопожатия с сервером или для получения
задачи. Содержит имя вычислителя, который совершает запрос.
*/
type SolverRequestJSON struct {
	SolverName string `json:"solverName"`
}

/*
AbsoluleSolver описывает сверх-вычислитель.
Так как по заданию в "нашей вселенной" все арифметические
операции выполняются очень долго, и очень трудозатратные,
был разработан сверх-вычислитель, способный посчитать
математическое выражение ровно за пять секунд.
Сверх-вычислитель предназначен для проверки и тестов системы:

	  фронтенд сервер ----веб-страница
			|
			|
		оркестратор ----- [вычислитель]
			|
			|
		база данных

Если система дает сбои, и подозревается ошибка на стороне
вычислителей, то их можно временно заменить на сверх-вычислитель
и посмотреть как система будет работать с ним.
*/
type AbsoluleSolver struct {
	TaskURL    string
	SolverName string
	Expression string
}

func NewAbsoluleSolver() *AbsoluleSolver {
	return &AbsoluleSolver{
		TaskURL:    "http://orchestrator_server:8082/getTaskToSolving",
		SolverName: "Absolule Solver",
		Expression: "",
	}
}

func (as *AbsoluleSolver) RunHandShakeStream() {
	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Формируем JSON
	request := SolverRequestJSON{
		SolverName: as.SolverName,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
	}

	go func(jsonRequest []byte) {
		for {
			select {
			case <-ticker.C:
				jsonBytes := bytes.NewBuffer(jsonRequest)
				req, err := http.Post("http://orchestrator_server:8082/solverHandShake", "application/json", jsonBytes)
				if err != nil {
					log.Println("[ERROR]: Can not connect to orkestrator: " + err.Error())
				}
				log.Println("[OK]: Hand shake!" + req.Status)
			}
		}
	}(jsonRequest)
}

func (as *AbsoluleSolver) RunSolverStream() {
	go func() {
		for {
			// Создаем JSON запроса
			request := SolverRequestJSON{
				SolverName: as.SolverName,
			}

			// Формируем JSON
			jsonRequest, err := json.Marshal(request)
			if err != nil {
				log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			}

			// Переменная для отклика
			var resp *http.Response
			for {
				// Пробуем отправить запрос на получение задачи
				resp, err = http.Post("http://orchestrator_server:8082/getTaskToSolving", "application/json", bytes.NewBuffer(jsonRequest))
				if err != nil || resp.StatusCode != http.StatusOK {
					// Если не удалочь отправить успешный запрос,
					// то ждем две секунды, и пытаемся отправить запрос повторно
					log.Println("[ERROR]: Can not connect to orkestrator")
					time.Sleep(2 * time.Second)

				} else {
					// Если запрос успешен и получили код 200, то выходим из цикла
					log.Println("[INFO]: Successful request")
					break
				}
			}
			defer resp.Body.Close()

			// Декодируем тело запроса в JSON нужной нам структуры
			var message TaskToSendToSolver
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&message)
			if err != nil {
				log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
				panic(err)
			}

			// Запоминаем выражение которое нужно вычислить
			as.Expression = message.Expression

			// Парсим выражение
			expr, err := parseMathExpression(message.Expression)
			if err != nil {
				log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
				panic(err)
			}

			// Вчисляем выражение
			res, err := evaluateMathExpression(expr, make(map[string]interface{}))
			if err != nil {
				log.Println("[ERROR]: Decoding JSON was failed: " + err.Error())
				panic(err)
			}
			log.Println("[INFO]: Successful evaluate")

			// Ждем пять секунд
			time.Sleep(10 * time.Second)

			// Создаем JSON запроса
			result := ResultFromSolver{
				SolverName: as.SolverName,
				Expression: message.Expression,
				Result:     fmt.Sprintf("%v", res),
				Status:     0,
			}

			// Формируем JSON
			jsonResult, err := json.Marshal(result)
			if err != nil {
				log.Println("[ERROR]: Can not encoding to JSON: " + err.Error())
			}

			for {
				// Пробуем отправить запрос с ответом на задачу
				resp, err = http.Post("http://orchestrator_server:8082/setResultOfExpression", "application/json", bytes.NewBuffer(jsonResult))
				if err != nil || resp.StatusCode != http.StatusOK {
					// Если не удалочь отправить успешный запрос,
					// то ждем две секунды, и пытаемся отправить запрос повторно
					log.Println("[ERROR]: Can not send result to orkestrator")
					time.Sleep(2 * time.Second)

				} else {
					// Если запрос успешен и получили код 200, то выходим из цикла
					log.Println("[OK]: Successful sending result")
					break
				}
			}
		}
	}()
}

/*
parseMathExpression Функция для парсинга выражения из строки

Parameters:

	string: Входное выражение в строке

Returns:

	*govaluate.EvaluableExpression: Выражение, готовое к вычислению
	error: Ошибки
*/
func parseMathExpression(expr string) (*govaluate.EvaluableExpression, error) {
	parsed, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

/*
evaluateMathExpression Функция вычисления выражения

Parameters:

	*govaluate.EvaluableExpression: Входное выражение
	map[string]interface{}: Словарь со значениями переменных

Returns:

	interface{}: Результат вычисления
	error: Ошибки
*/
func evaluateMathExpression(expr *govaluate.EvaluableExpression,
	vars map[string]interface{}) (interface{}, error) {
	result, err := expr.Evaluate(vars)
	if err != nil {
		return nil, err
	}

	return result, nil
}
