package db_connect

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
)

func CreateOperationsTime(ctx context.Context, db SQLQueryExec) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, "INSERT INTO operations_times (plus, minus, division, multiplication) VALUES ($1, $2, $3, $4) RETURNING id",
		10, 10, 10, 10).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetOperationsTimeByID(ctx context.Context, db SQLQueryExec, id int) (entities.OperationsTime, error) {
	var times entities.OperationsTime
	row := db.QueryRowContext(ctx, "SELECT id, plus, minus, division, multiplication FROM operations_times WHERE id = $1", id)
	err := row.Scan(&times.Id, &times.Plus, &times.Minus, &times.Division, &times.Multiplication)
	if err != nil {
		return entities.OperationsTime{}, err
	}
	return times, nil
}

func GetOperationsTimeByUserID(ctx context.Context, db SQLQueryExec, id int) (entities.OperationsTime, error) {
	var times entities.OperationsTime
	row := db.QueryRowContext(ctx, "SELECT id, plus, minus, division, multiplication FROM operations_times WHERE id = (SELECT operations_time_id FROM users WHERE id = $1)", id)
	err := row.Scan(&times.Id, &times.Plus, &times.Minus, &times.Division, &times.Multiplication)
	if err != nil {
		return entities.OperationsTime{}, err
	}
	return times, nil
}

func UpdateOperationsTimeByID(ctx context.Context, db SQLQueryExec, operationsTimes entities.OperationsTime, id int) error {
	if userId, ok := ctx.Value("userId").(int); ok {
		_, err := db.ExecContext(ctx, "UPDATE operations_times SET plus = $1, minus = $2, division = $3, multiplication = $4 WHERE id = $5 AND user_id = $6",
			operationsTimes.Plus, operationsTimes.Minus, operationsTimes.Division, operationsTimes.Multiplication, id, userId)
		return err
	}
	_, err := db.ExecContext(ctx, "UPDATE operations_times SET plus = $1, minus = $2, division = $3, multiplication = $4 WHERE id = $5",
		operationsTimes.Plus, operationsTimes.Minus, operationsTimes.Division, operationsTimes.Multiplication, id)
	return err
}

func UpdateOperationsTimeByUserID(ctx context.Context, db SQLQueryExec, operationsTimes entities.OperationsTime, userId int) error {
	_, err := db.ExecContext(ctx, "UPDATE operations_times SET plus = $1, minus = $2, division = $3, multiplication = $4 WHERE id = (SELECT operations_time_id FROM users WHERE id = $5)",
		operationsTimes.Plus, operationsTimes.Minus, operationsTimes.Division, operationsTimes.Multiplication, userId)
	return err
}
