package db_connect

import (
	"Zadacha/internal/entities"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
	"time"
)

type DBConnection struct {
	db *sql.DB
}

func OpenDb(connectionString string) (DBConnection, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println(err)
		return DBConnection{}, err
	}

	// Создание таблиц, если они не существуют
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS operations (
			id SERIAL PRIMARY KEY,
			znak TEXT,
			left_is_ready INTEGER,
			left_data FLOAT8,
			right_is_ready INTEGER,
			right_data FLOAT8,
			father_id INTEGER,
			son_side INTEGER,
			final_operation_id INTEGER
		);

		CREATE TABLE IF NOT EXISTS expressions (
			id SERIAL PRIMARY KEY,
			need_to_do TEXT,
			status TEXT,
			result FLOAT8,
			start_time INTEGER,
			end_time INTEGER,
			user_id INTEGER
		);

		CREATE TABLE IF NOT EXISTS agents (
			id SERIAL PRIMARY KEY,
			status TEXT,
			status_text TEXT,
			ping_time INTEGER
		);

		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			operations_time_id INTEGER
		);

		CREATE TABLE IF NOT EXISTS operations_times (
			id SERIAL PRIMARY KEY,
			plus INTEGER,
			minus INTEGER,
			division INTEGER,
			multiplication INTEGER
		);
	`)
	if err != nil {
		return DBConnection{}, err
	}

	res := DBConnection{db: db}
	err = res.InitOperationTime()
	if err != nil {
		return DBConnection{}, err
	}
	return res, nil
}

func (conn *DBConnection) GetOperation(id int) (entities.Operation, error) {
	var op entities.Operation
	row := conn.db.QueryRow("SELECT id, znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, final_operation_id FROM operations WHERE id = $1", id)
	err := row.Scan(&op.Id, &op.Znak, &op.LeftIsReady, &op.LeftData, &op.RightIsReady, &op.RightData, &op.FatherId, &op.SonSide, &op.FinalOperationId)
	if err != nil {
		return entities.Operation{}, err
	}
	return op, nil
}

func (conn *DBConnection) AddOperation(value *entities.Operation) (int, error) {
	var id int
	err := conn.db.QueryRow("INSERT INTO operations (znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, final_operation_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		value.Znak, value.LeftIsReady, value.LeftData, value.RightIsReady, value.RightData, value.FatherId, value.SonSide, value.FinalOperationId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (conn *DBConnection) GetAll() ([]entities.Operation, error) {
	var ops []entities.Operation
	rows, err := conn.db.Query("SELECT id, znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, final_operation_id FROM operations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var op entities.Operation
		err = rows.Scan(&op.Id, &op.Znak, &op.LeftIsReady, &op.LeftData, &op.RightIsReady, &op.RightData, &op.FatherId, &op.SonSide, &op.FinalOperationId)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func (conn *DBConnection) UpdateLeft(id int, leftData float64) error {
	_, err := conn.db.Exec("UPDATE operations SET left_data = $1, left_is_ready = 1 WHERE id = $2", leftData, id)
	return err
}

func (conn *DBConnection) UpdateRight(id int, rightData float64) error {
	_, err := conn.db.Exec("UPDATE operations SET right_data = $1, right_is_ready = 1 WHERE id = $2", rightData, id)
	return err
}

func (conn *DBConnection) UpdateFather(id, fatherId, side int) error {
	_, err := conn.db.Exec("UPDATE operations SET father_id = $1, son_side = $2 WHERE id = $3", fatherId, side, id)
	return err
}

func (conn *DBConnection) Delete(id int) error {
	_, err := conn.db.Exec("DELETE FROM operations WHERE id = $1", id)
	return err
}

func (conn *DBConnection) IsReadyToExecute(id int) (bool, error) {
	var leftIsReady, rightIsReady int
	err := conn.db.QueryRow("SELECT left_is_ready, right_is_ready FROM operations WHERE id = $1", id).Scan(&leftIsReady, &rightIsReady)
	if err != nil {
		return false, err
	}
	return leftIsReady == 1 && rightIsReady == 1, nil
}

func (conn *DBConnection) AgentPing(id int, status, statusText string) int {
	if status == "create" {
		var newId int
		err := conn.db.QueryRow("INSERT INTO agents (status, status_text, ping_time) VALUES ($1, $2, $3) RETURNING id",
			status, statusText, time.Now().Unix()).Scan(&newId)
		if err != nil {
			return 0
		}
		return newId
	}

	_, err := conn.db.Exec("UPDATE agents SET status = $1, status_text = $2, ping_time = $3 WHERE id = $4",
		status, statusText, time.Now().Unix(), id)
	if err != nil {
		return 0
	}
	return id
}

func (conn *DBConnection) DeleteAgent(id int) error {
	_, err := conn.db.Exec("DELETE FROM agents WHERE id = $1", id)
	return err
}

func (conn *DBConnection) GetAllAgents() ([]entities.Agent, error) {
	var agents []entities.Agent
	rows, err := conn.db.Query("SELECT id, status, status_text, ping_time FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agent entities.Agent
		err = rows.Scan(&agent.Id, &agent.Status, &agent.StatusText, &agent.PingTime)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (conn *DBConnection) CreateFinalOperation(expression entities.Expression) (int, error) {
	var id int
	err := conn.db.QueryRow("INSERT INTO expressions (need_to_do, status, result, start_time, end_time, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		expression.NeedToDo, expression.Status, expression.Result, expression.StartTime, expression.EndTime, expression.UserId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (conn *DBConnection) GetFinalOperationByID(id int) (entities.Expression, error) {
	var ex entities.Expression
	row := conn.db.QueryRow("SELECT id, need_to_do, status, result, start_time, end_time, user_id FROM expressions WHERE id = $1", id)
	err := row.Scan(&ex.Id, &ex.NeedToDo, &ex.Status, &ex.Result, &ex.StartTime, &ex.EndTime, &ex.UserId)
	if err != nil {
		return entities.Expression{}, err
	}
	return ex, nil
}

func (conn *DBConnection) GetAllFinalOperations() ([]entities.Expression, error) {
	var ops []entities.Expression
	rows, err := conn.db.Query("SELECT id, need_to_do, status, result, start_time, end_time, user_id FROM expressions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var op entities.Expression
		err = rows.Scan(&op.Id, &op.NeedToDo, &op.Status, &op.Result, &op.StartTime, &op.EndTime, &op.UserId)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func (conn *DBConnection) UpdateFinalOperation(id int, newResult float64, status string) error {
	_, err := conn.db.Exec("UPDATE expressions SET status = $1, end_time = $2, result = $3 WHERE id = $4",
		status, time.Now().Unix(), newResult, id)
	return err
}

func (conn *DBConnection) OhNoFinalOperationError(id int) {
	// Обновление статуса финальной операции на "my_errors"
	_, _ = conn.db.Exec("UPDATE expressions SET status = 'my_errors', end_time = $1 WHERE id = $2", time.Now().Unix(), id)

	// Удаление связанных операций
	_, _ = conn.db.Exec("DELETE FROM operations WHERE final_operation_id = $1", id)
}

func (conn *DBConnection) Close() error {
	return conn.db.Close()
}

func (conn *DBConnection) InitOperationTime() error {
	for id := 1; id < 5; id++ {
		_, err := conn.db.Exec("INSERT INTO operations_times (id, plus, minus, division, multiplication) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING",
			id, 10, 10, 10, 10)
		if err != nil {
			return err
		}
	}
	return nil
}

func (conn *DBConnection) GetOperationTimeByID(id int) (entities.OperationsTime, error) {
	var times entities.OperationsTime
	row := conn.db.QueryRow("SELECT id, plus, minus, division, multiplication FROM operations_times WHERE id = $1", id)
	err := row.Scan(&times.Id, &times.Plus, &times.Minus, &times.Division, &times.Multiplication)
	if err != nil {
		return entities.OperationsTime{}, err
	}
	return times, nil
}

func (conn *DBConnection) UpdateOperationTime(id, time int) error {
	_, err := conn.db.Exec("UPDATE operations_times SET plus = $1, minus = $1, division = $1, multiplication = $1 WHERE id = $2",
		time, id)
	return err
}
