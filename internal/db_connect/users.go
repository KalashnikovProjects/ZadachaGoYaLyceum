package db_connect

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
)

func CreateUser(ctx context.Context, db SQLQueryExec, user entities.User) (int, error) {
	var id int
	operationsTimeId, err := CreateOperationsTime(ctx, db)
	if err != nil {
		return 0, err
	}
	err = db.QueryRowContext(ctx, "INSERT INTO users (name, password_hash, operations_time_id) VALUES ($1, $2, $3) RETURNING id",
		user.Name, user.PasswordHash, operationsTimeId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetUserByID(ctx context.Context, db SQLQueryExec, id int) (entities.User, error) {
	var user entities.User
	row := db.QueryRowContext(ctx, "SELECT id, name, password_hash, operations_time_id FROM users WHERE id = $1", id)
	err := row.Scan(&user.Id, &user.Name, &user.PasswordHash, &user.OperationsTimeId)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func GetUserByName(ctx context.Context, db SQLQueryExec, name string) (entities.User, error) {
	var user entities.User
	row := db.QueryRowContext(ctx, "SELECT id, name, password_hash, operations_time_id FROM users WHERE name = $1", name)
	err := row.Scan(&user.Id, &user.Name, &user.PasswordHash, &user.OperationsTimeId)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}
