package pkg

import (
	"fmt"
	"time"
	//"regexp"
	"strconv"
	"strings"
	"sync"
	"encoding/json"
	"log"
	"bytes"
	"net/http"
)

var (
	timesMap map[string]int
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
*/
type Solver struct {
	HandShakeURL string
	GetTaskURL    string
	SendResultURL string
	SolverName    string
	Expression    string
}

/*
*/
func NewSolver(name string) *Solver {
	return &Solver{
		HandShakeURL: "http://orchestrator_server:8082/solverHandShake",
		GetTaskURL:    "http://orchestrator_server:8082/getTaskToSolving",
		SendResultURL: "http://orchestrator_server:8082/setResultOfExpression",
		SolverName:    name,
		Expression:    "",
	}
}

func (s *Solver) RunHandShakeStream() {
	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Формируем JSON
	request := SolverRequestJSON{
		SolverName: s.SolverName,
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
				req, err := http.Post(s.HandShakeURL, "application/json", jsonBytes)
				if err != nil {
					log.Println("[ERROR]: Can not connect to orkestrator: " + err.Error())
				}
				log.Println("[OK]: Hand shake!" + req.Status)
			}
		}
	}(jsonRequest)
}

func (s *Solver) RunSolverStream() {
	go func() {
		for {
			// Создаем JSON запроса
			request := SolverRequestJSON{
				SolverName: s.SolverName,
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
				resp, err = http.Post(s.GetTaskURL, "application/json", bytes.NewBuffer(jsonRequest))
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

			// Запоминаем выражение которое нужно вычислить и время выполнения операций
			s.Expression = message.Expression
			timesMap = message.Times

			// Парсим и вычисляем выражение
			res, err := Solving(message.Expression)
			if err != nil {
				log.Println("[ERROR]: Can not solving expression: " + err.Error())
				panic(err)
			}

			log.Println("[INFO]: Successful evaluate")

			// Создаем JSON запроса
			result := ResultFromSolver{
				SolverName: s.SolverName,
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
				resp, err = http.Post(s.SendResultURL, "application/json", bytes.NewBuffer(jsonResult))
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

func Solving(expression string) (float64, error) {
	// Создаем синхронизатор
	wg := sync.WaitGroup{}

	// Создаем массив чисел в виде строк
	stringArrayOfNumber := strings.FieldsFunc(expression, func(r rune) bool {
		return r == '+' || r == '-' || r == '/' || r == '*'
	})

	// Преобразуем числа из строкового представления в числовое
	arrayOfNumber := make([]float64, 0)
	for _, val := range stringArrayOfNumber {
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0.0, err
		}
		arrayOfNumber = append(arrayOfNumber, f)
	}

	// Создаем массив операций
	arrayOfOperation := make([]string, 0)
	for _, ch := range strings.Split(expression, "") {
		if ch == "+" || ch == "-" || ch == "/" || ch == "*" {
			arrayOfOperation = append(arrayOfOperation, ch)
		}
	}

	// Группа операций первого приоритета, пример:
	// В выражении 0+1*2+3+4-5*6/7 -> 1*2 и 5*6/7 являются группами
	// в этом выражении 1*2 и 5*6/7 будут сначала выполнены в отдельных потоках
	// и их результаты будут положены в: 0+_+3+4-_
	// после чего будет найдено финальное значение выражения

	// Счетчик количества групп состоящийх из операций второго приоритета
	groupCounter := 1
	for i := 0; i < len(arrayOfOperation); i += 1 {
		// Ищем операции второго приоритета
		if arrayOfOperation[i] == "+" || arrayOfOperation[i] == "-" {
			groupCounter += 1
		}
	}

	counter := 0
	operations := make([]string, 0) // Список операций группы первого приоритета
	numbers := make([]float64, 0)   // Список чисел группы первого приоритета
	begin := false                  // Нашли начало группы первого приоритета

	// Массив в который будут отправлены числа,
	// над которыми будут выполняться операции второго приоритета
	groupResultArray := make([]float64, groupCounter)
	// Массив с операциями второго приоритета
	groupOperatinArray := make([]string, groupCounter-1)
	for i := 0; i < len(arrayOfOperation); i += 1 {
		fmt.Printf(" --- Итерация: %v --- \n", i)
		// Если нашли операцию второго приоритета, значит
		// записываем число с операцией в массив, либо, если мы до этого нашли
		// группу первого приоритета, ее надо вычислить и отправить ее
		// результат в массив вместо числа
		if arrayOfOperation[i] == "+" || arrayOfOperation[i] == "-" {
			if begin {
				// Если мы нашли операцию второго приоритета,
				// а до этого была операция первого приоритета,
				// то записываем в число в список группы первого приоритета
				numbers = append(numbers, arrayOfNumber[i])
				// Записываем операцию второго приоритета в массив
				groupOperatinArray[counter] = arrayOfOperation[i]
				// запускаем вычисление группы в отдельной горутине,
				// передав ей массивы со значениями и операциями,
				// а так же индекс, куда надо положить результат вычисления группы
				wg.Add(1)
				go func(numbers []float64, operations []string, counter int) {
					defer wg.Done()
					groupResultArray[counter] = FirstPriority(numbers, operations)
				}(numbers, operations, counter)

				// Очищаем массивы для поиска следующей группы операций первого приоритета
				operations = make([]string, 0)
				numbers = make([]float64, 0)
				begin = false
				counter += 1
			} else {
				// Если найдена операция второго приоритета,
				// а до этого не быт открыт набор в группу первого приоритета,
				// то просто записываем текущий знак и текущее число
				groupResultArray[counter] = arrayOfNumber[i]
				groupOperatinArray[counter] = arrayOfOperation[i]
				counter += 1
			}

			// Если мы на конце массива с операциями, надо доподнительно записать
			// крайнее в выражении число
			if i == len(arrayOfOperation)-1 {
				groupResultArray[counter+1] = arrayOfNumber[i+1]
			}
		}

		// Если нашли операцию первого приоритета, создаем
		// группу первого приоритета, содержащую только операции первого порядка
		// Такая группа выполняется в отдельной горутине, однако внутри себя
		// группа может распраралелиться еще
		if arrayOfOperation[i] == "/" || arrayOfOperation[i] == "*" {
			if !begin {
				begin = true
			}
			// Добавляем число, после которого идет оператор первого приоритета
			numbers = append(numbers, arrayOfNumber[i])
			// Добавляем оператор после этого числа
			operations = append(operations, arrayOfOperation[i])

			// Если мы на конце, то есть выражение заканчивается произведением/делением
			// то добавляем крайнее число и запускаем подсчет
			if i == len(arrayOfOperation)-1 {
				numbers = append(numbers, arrayOfNumber[i+1])
				wg.Add(1)
				go func(numbers []float64, operations []string, counter int) {
					defer wg.Done()
					groupResultArray[counter] = FirstPriority(numbers, operations)
				}(numbers, operations, counter)
			}
		}
	}

	// Ждем пока посчитаются все операции первого приоритета
	// после чего получается массив чисел, над которыми остается
	// совершать только сложения и вычитания, то есть операции второго приоритета
	wg.Wait()
	return SecondPriority(groupResultArray, groupOperatinArray), nil
}

/*
FirstPriority
*/
func FirstPriority(arrayOfNumber []float64, arrayOfOperation []string) float64 {
	res := arrayOfNumber[0]
	for i := 0; i < len(arrayOfOperation); i += 1 {
		switch arrayOfOperation[i] {
		case "*":
			res *= arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["*"]) * time.Second)
		case "/":
			if arrayOfNumber[i+1] != 0.0 {
				res /= arrayOfNumber[i+1]
				time.Sleep(time.Duration(timesMap["/"]) * time.Second)
			} else {
				panic("fatal")
			}
		}
	}
	return res
}

/*
SecondPriority
*/
func SecondPriority(arrayOfNumber []float64, arrayOfOperation []string) float64 {
	res := arrayOfNumber[0]
	for i := 0; i < len(arrayOfOperation); i += 1 {
		switch arrayOfOperation[i] {
		case "+":
			res += arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["+"]) * time.Second)
		case "-":
			res -= arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["-"]) * time.Second)
		}
	}
	return res
}
