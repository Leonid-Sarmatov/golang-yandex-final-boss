package pkg

import (
	"database/sql"
	"log"
	//"time"

	_ "github.com/lib/pq"
)

type SettingsTimeOfOperation struct {
	ID              int
	Operation       string
	TimeOfOperation int
}

type DatabaseConnection struct {
	DB *sql.DB
}

/*
NewDatabaseConnection создает соединение с базой данных
Parameters:

	string: Готовая строка параметров для соединения

Returns:

	*DatabaseConnection: Структура с соединением
	error: Ошибки
*/
func NewDatabaseConnection(connectString string) (*DatabaseConnection, error) {
	// Пробуем создать соединение с базой данных
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		//panic(err)
		log.Printf("[ERROR]: Spawn connection to database was failed: %v", err)
		return nil, err
	}

	// Если удалось, то добавляем соединение в возвращаемую структуру
	databaseConnection := &DatabaseConnection{
		DB: db,
	}

	// Если по какой то причине в базе нет таблицы с запросами
	// на вычисленя, то создаем таблицу
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS task_table (
        id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, 
        expression VARCHAR(255), 
        hash VARCHAR(255), 
        status BIGINT,
		result VARCHAR(255),
		time_begin TIMESTAMP,
		time_end TIMESTAMP
    );`)

	// Если таблицу создать не удалось, то возвращаем соединение
	// и ошибку создания таблицы
	if err != nil {
		return databaseConnection, err
	}

	// Если по какой то причине в базе нет таблицы с настройками
	// времени вычисленя, то создаем таблицу
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS operation_table (
        id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, 
        operation VARCHAR(255), 
        timeInSecond INT
    );`)

	// Если таблицу создать не удалось, то возвращаем соединение
	// и ошибку создания таблицы
	if err != nil {
		return databaseConnection, err
	}

	return databaseConnection, nil
}

/*
SendRequestWithoutWaitingRequest передает строку запроса в базу данных
*/
func (db *DatabaseConnection) SendRequestWithoutWaitingRequest(request string) error {
	_, err := db.DB.Exec(request)
	return err
}

/*
SendRequestAwaitingRequest передает строку запроса в базу данных и возвращает ответ
*/
func (db *DatabaseConnection) SendRequestAwaitingRequest(request string) (*sql.Rows, error) {
	rows, err := db.DB.Query(request)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db *DatabaseConnection) CloseConnecton() error {
	err := db.DB.Close()
	return err
}

/*
AddTask записывает задачу в базу данных
*/
func (db *DatabaseConnection) AddTask(task TaskJSON) error {
	_, err := db.DB.Exec(`INSERT INTO task_table (
		expression, 
        hash, 
        status,
		result,
		time_begin,
		time_end
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		task.Expression,
		task.HashID,
		task.Status,
		task.Result,
		task.BeginTime.Format("2006-01-02 15:04:05"),
		task.EndTime.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return err
	}

	return nil
}

/*
GetAllTasks возвращает все задачи из базы данных
*/
func (db *DatabaseConnection) GetAllTasks() ([]TaskJSON, error) {
	rows, err := db.DB.Query("SELECT * FROM task_table")
	if err != nil {
		return nil, err
	}

	tasks := make([]TaskJSON, 0)
	for rows.Next() {
		var t TaskJSON
		err = rows.Scan(&t.ID, &t.Expression, &t.HashID, &t.Status, &t.Result, &t.BeginTime, &t.EndTime)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

/*
GetTasksFromStatus возвращает список с задач с определенным статусом
*/
func (db *DatabaseConnection) GetTasksFromStatus(status int) ([]TaskJSON, error) {
	rows, err := db.DB.Query("SELECT * FROM task_table WHERE status=$1", status)
	if err != nil {
		return nil, err
	}

	tasks := make([]TaskJSON, 0)
	for rows.Next() {
		var t TaskJSON
		err = rows.Scan(&t.ID, &t.Expression, &t.HashID, &t.Status, &t.Result, &t.BeginTime, &t.EndTime)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

/*
UpdateStatusFromExpression обновляет статус у записи с определенным выражением
*/
func (db *DatabaseConnection) UpdateStatusFromExpression(status int, expression string) error {
	_, err := db.DB.Exec("UPDATE task_table SET status = $1 WHERE expression = $2", status, expression)
	return err
}

/*
UpdateStatusFromExpression обновляет статус и результат у записи с определенным выражением
*/
func (db *DatabaseConnection) UpdateStatusAndResultFromExpression(status int, expression string, result string) error {
	_, err := db.DB.Exec("UPDATE task_table SET status = $1, result = $3 WHERE expression = $2", 
						status, expression, result)
	return err
}

/*
DeleteTasksFromStatus удаляет задачи с определеными статусами
*/
func (db *DatabaseConnection) DeleteTasksFromStatus(status int) error {
	_, err := db.DB.Exec("DELETE FROM task_table WERE status=$1", status)
	return err
}

/*
GetTasksFromExpession возвращает задачи с определенным математическим выражением
*/
func (db *DatabaseConnection) GetTasksFromExpession(expression string) ([]TaskJSON, error) {
	rows, err := db.DB.Query("SELECT * FROM task_table WHERE expression=$1", expression)
	if err != nil {
		return nil, err
	}

	tasks := make([]TaskJSON, 0)
	for rows.Next() {
		var t TaskJSON
		err = rows.Scan(&t.ID, &t.Expression, &t.HashID, &t.Status, &t.Result, &t.BeginTime, &t.EndTime)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

/*
DeleteTasksFromExpession удаляет задачу с определенным условием
*/
func (db *DatabaseConnection) DeleteTasksFromExpession(expression string) error {
	_, err := db.DB.Exec("DELETE FROM task_table WERE expression=$1", expression)
	return err
}

/*
GetAllTimesOfOperation возвращает список со всеми операциями и временем их выполнения
*/
func (db *DatabaseConnection) GetAllTimesOfOperation() ([]SettingsTimeOfOperation, error) {
	rows, err := db.DB.Query("SELECT * FROM operation_table")
	if err != nil {
		return nil, err
	}

	times := make([]SettingsTimeOfOperation, 0)
	for rows.Next() {
		var t SettingsTimeOfOperation
		err = rows.Scan(&t.ID, &t.Operation, &t.TimeOfOperation)
		if err != nil {
			return nil, err
		}

		times = append(times, t)
	}

	return times, nil
}

/*
SetTimesToOperation записывает время выполения операции
*/
func (db *DatabaseConnection) SetTimesToOperation(times TimeOfOperationJSON) error {
	isIntable := false
	var query string

	for key, val := range times.Times {
		err := db.DB.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM operation_table WHERE operation = $1)", key).Scan(&isIntable)
		if err != nil {
			return err
		}
		if isIntable {
			query = `UPDATE operation_table SET timeInSecond = $2 WHERE operation = $1;`
		} else {
			query = `INSERT INTO operation_table (operation, timeInSecond) VALUES ($1, $2);`
		}
		_, err = db.DB.Exec(query, key, val)
		if err != nil {
			return err
		}
	}
	
	return nil
}
