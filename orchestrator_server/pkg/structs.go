package pkg

import (
	"time"
)

/*
TaskJSON описывает структуру задачи,
хранящейся в таблице базы данных
*/
type TaskJSON struct {
	ID         int       `json:"id"`
	Expression string    `json:"expression"`
	HashID     string    `json:"hashID"`
	Status     int       `json:"status"`
	Result     string    `json:"result"`
	BeginTime  time.Time `json:"beginTime"`
	EndTime    time.Time `json:"endTime"`
}

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
TimeOfOperationJSON описывает JSON со временем выполнения
для каждой операции. Такую структуру должен содержать
запрос клиента, желающего задать время выполнения
операциям при помощи исполнителя SetTimeOfOperations
*/
type TimeOfOperationJSON struct {
	Times map[string]int `json:"times"`
}

/*
ExpressionRequestJSON описывает JSON с выражением
для вычисления. Такую структуру должен содержать запрос
клиента, желающего добавить выражений в обработку.
Запрос содержит само выражение и время его отправки
на сервер
*/
type ExpressionRequestJSON struct {
	Expression string    `json:"expression"`
	TimeToSend time.Time `json:"timeToSend"`
}

/*
Solver описывает вычислителя и информацию о нем:
Имя вычислителя, вычисляемое выражение в данный момент,
последний раз, когда вычислитель давал о себе знать и
информационную строку от вычислителя. Эта же структура
используется для создания ответа клиенту, на запрос
об информации о вычислителях в исполнителе GetListOfSolvers
*/
type Solver struct {
	SolverName           string    `json:"solverName"`
	SolvingNowExpression string    `json:"solvingExpression"`
	LastPing             time.Time `json:"lastPing"`
	InfoString           string    `json:"infoString"`
}

/*
SolverRequestJSON описывает JSON запроса вычислителя
на сервер. Такую структуру должен содержать запрос,
для регулярного рукопожатия с сервером или для получения
задачи. Содержит имя вычислителя, который совершает запрос.
Используется в исполнителе GetReadyTaskToSolving
*/
type SolverRequestJSON struct {
	SolverName string `json:"solverName"`
}
