package db_connect

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/my_errors"
	"time"
)

func CreateExpression(ctx context.Context, db SQLQueryExec, expression entities.Expression) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, "INSERT INTO expressions (need_to_do, status, result, start_time, end_time, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		expression.NeedToDo, expression.Status, expression.Result, expression.StartTime, expression.EndTime, expression.UserId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetExpressionByID(ctx context.Context, db SQLQueryExec, id int) (entities.Expression, error) {
	var ex entities.Expression
	row := db.QueryRowContext(ctx, "SELECT id, need_to_do, status, result, start_time, end_time, user_id FROM expressions WHERE id = $1", id)
	err := row.Scan(&ex.Id, &ex.NeedToDo, &ex.Status, &ex.Result, &ex.StartTime, &ex.EndTime, &ex.UserId)
	if err != nil {
		return entities.Expression{}, err
	}
	if userId, ok := ctx.Value("userId").(int); ok && userId != ex.UserId {
		return entities.Expression{}, my_errors.PermissionDeniedError
	}
	return ex, nil
}

func GetAllExpressions(ctx context.Context, db SQLQueryExec) ([]entities.Expression, error) {
	var expressions []entities.Expression
	rows, err := db.QueryContext(ctx, "SELECT id, need_to_do, status, result, start_time, end_time, user_id FROM expressions")
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
		if userId, ok := ctx.Value("userId").(int); ok && userId != op.UserId {
			continue
		}
		expressions = append(expressions, op)
	}

	return expressions, nil
}

func UpdateExpression(ctx context.Context, db SQLQueryExec, id int, newResult float64, status string) error {
	if userId, ok := ctx.Value("userId").(int); ok {
		_, err := db.ExecContext(ctx, "UPDATE expressions SET status = $1, end_time = $2, result = $3 WHERE id = $4 AND user_id = $5",
			status, time.Now().Unix(), newResult, id, userId)
		return err
	}
	_, err := db.ExecContext(ctx, "UPDATE expressions SET status = $1, end_time = $2, result = $3 WHERE id = $4",
		status, time.Now().Unix(), newResult, id)
	return err

}

func OhNoExpressionError(ctx context.Context, db SQLQueryExec, id int) {
	// Обновление статуса финальной операции на "error"
	_, _ = db.ExecContext(ctx, "UPDATE expressions SET status = 'error', end_time = $1 WHERE id = $2", time.Now().Unix(), id)

	// Удаление связанных операций
	_, _ = db.ExecContext(ctx, "DELETE FROM operations WHERE expression_id = $1", id)
}
