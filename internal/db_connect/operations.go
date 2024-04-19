package db_connect

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
)

func GetOperationByID(ctx context.Context, db SQLQueryExec, id int) (entities.Operation, error) {
	var op entities.Operation
	row := db.QueryRowContext(ctx, "SELECT id, znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, expression_id FROM operations WHERE id = $1", id)
	err := row.Scan(&op.Id, &op.Znak, &op.LeftIsReady, &op.LeftData, &op.RightIsReady, &op.RightData, &op.FatherId, &op.SonSide, &op.ExpressionId)
	if err != nil {
		return entities.Operation{}, err
	}
	return op, nil
}

func AddOperation(ctx context.Context, db SQLQueryExec, value *entities.Operation) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, "INSERT INTO operations (znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, expression_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		value.Znak, value.LeftIsReady, value.LeftData, value.RightIsReady, value.RightData, value.FatherId, value.SonSide, value.ExpressionId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetAllOperations(ctx context.Context, db SQLQueryExec) ([]entities.Operation, error) {
	var ops []entities.Operation
	rows, err := db.QueryContext(ctx, "SELECT id, znak, left_is_ready, left_data, right_is_ready, right_data, father_id, son_side, expression_id FROM operations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var op entities.Operation
		err = rows.Scan(&op.Id, &op.Znak, &op.LeftIsReady, &op.LeftData, &op.RightIsReady, &op.RightData, &op.FatherId, &op.SonSide, &op.ExpressionId)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func UpdateLeftOperation(ctx context.Context, db SQLQueryExec, id int, leftData float64) error {
	_, err := db.ExecContext(ctx, "UPDATE operations SET left_data = $1, left_is_ready = 1 WHERE id = $2", leftData, id)
	return err
}

func UpdateRightOperation(ctx context.Context, db SQLQueryExec, id int, rightData float64) error {
	_, err := db.ExecContext(ctx, "UPDATE operations SET right_data = $1, right_is_ready = 1 WHERE id = $2", rightData, id)
	return err
}

func UpdateFatherOperation(ctx context.Context, db SQLQueryExec, id, fatherId, side int) error {
	_, err := db.ExecContext(ctx, "UPDATE operations SET father_id = $1, son_side = $2 WHERE id = $3", fatherId, side, id)
	return err
}

func DeleteOperation(ctx context.Context, db SQLQueryExec, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM operations WHERE id = $1", id)
	return err
}

func IsReadyToExecuteOperation(ctx context.Context, db SQLQueryExec, id int) (bool, error) {
	var leftIsReady, rightIsReady int
	err := db.QueryRowContext(ctx, "SELECT left_is_ready, right_is_ready FROM operations WHERE id = $1", id).Scan(&leftIsReady, &rightIsReady)
	if err != nil {
		return false, err
	}
	return leftIsReady == 1 && rightIsReady == 1, nil
}
