package pkg

import (
	"net/http"
	//"github.com/Knetic/govaluate"
)

/*
AddArithmeticExpression принимает запрос с арифметическим 
выражением, и возвращает ошибку если не удается распарсить
*/
type AddArithmeticExpression struct{}

func NewAddArithmeticExpression() *AddArithmeticExpression {
	return &AddArithmeticExpression{}
}

func (e *AddArithmeticExpression) getExecutorRoute() string {
	return "/addArithmeticExpression"
}

func (e *AddArithmeticExpression) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
GetListExpressionsWithStatuses принимает запрос
и возвращает список со всеми задачами 
*/
type GetListExpressionsWithStatuses struct{}

func NewGetListExpressionsWithStatuses() *GetListExpressionsWithStatuses {
	return &GetListExpressionsWithStatuses{}
}

func (e *GetListExpressionsWithStatuses) getExecutorRoute() string {
	return "/getListOfTasks"
}

func (e *GetListExpressionsWithStatuses) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
GetValueOfExpressionByIdentifier принимает запрос с 
арифметическим выражением и возвращает результат вычислений 
*/
type GetValueOfExpression struct{}

func NewGetValueOfExpression() *GetValueOfExpression {
	return &GetValueOfExpression{}
}

func (e *GetValueOfExpression) getExecutorRoute() string {
	return "/getResultOfExpression"
}

func (e *GetValueOfExpression) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
GetListOfAvailableOperations принимает запрос и
возвращает список времени выполнения для каждой операции
*/
type GetListOfAvailableOperations struct{}

func NewGetListOfAvailableOperations() *GetListOfAvailableOperations {
	return &GetListOfAvailableOperations{}
}

func (e *GetListOfAvailableOperations) getExecutorRoute() string {
	return "/uuu"
}

func (e *GetListOfAvailableOperations) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
SetExecutionTimeOfOperations принимает запрос со списком
времени выполнения для каждой операции
*/
type SetExecutionTimeOfOperations struct{}

func NewSetExecutionTimeOfOperations() *SetExecutionTimeOfOperations {
	return &SetExecutionTimeOfOperations{}
}

func (e *SetExecutionTimeOfOperations) getExecutorRoute() string {
	return "/setExecutionTimeOfOperations"
}

func (e *SetExecutionTimeOfOperations) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
GetReadyTaskToSolving принимает запрос с информацией
о вычислителе и возвращает задачу готовую к выполнению
вместе с информацией о времени выполнения арифметических операций
*/
type GetReadyTaskToSolving struct{}

func NewGetReadyTaskToSolving() *GetReadyTaskToSolving {
	return &GetReadyTaskToSolving{}
}

func (e *GetReadyTaskToSolving) getExecutorRoute() string {
	return "/getTaskToSolving"
}

func (e *GetReadyTaskToSolving) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

/*
SetResultOfSolving принимает запрос с результатом, информацией 
о вычислителе и ошибках, возникших при выполнении
*/
type SetResultOfSolving struct{}

func NewGetResultOfSolving() *SetResultOfSolving {
	return &SetResultOfSolving{}
}

func (e *SetResultOfSolving) getExecutorRoute() string {
	return "/setResultOfExpression"
}

func (e *SetResultOfSolving) getExecutorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
