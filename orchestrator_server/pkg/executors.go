package pkg

/*
AddArithmeticExpression описывпет исполнтеля
добавляения арифметического выражения на выполнение
*/
type AddArithmeticExpression struct{}

/*
GetListExpressionsWithStatuses описывает исполнителя
позволяющего получить в ответ на запрос список выражений со статусами
*/
type GetListExpressionsWithStatuses struct{}

/*
	описывает исполнителя

позволяющего получить в ответ на запрос значение выражения
по его идентификатору
*/
type GetValueOfExpressionByIdentifier struct{}

/*
	описывает исполнителя

позволяющего получить в ответ на запрос список операций и
времени их выполнения
*/
type GetListOfAvailableOperations struct{}

/*
	описывает исполнителя

принимающего запрос со временем выполнения для каждой операции
*/
type SetExecutionTimeOfOperations struct{}

/*
	описывает исполнителя

возвращающего в ответ на запрос готовую к решению задачу
*/
type GetReadyTaskToSolving struct{}

/*
	описывает исполнителя

принимающего запрос с результатом вычислений
*/
type GetResultOfSolving struct{}
