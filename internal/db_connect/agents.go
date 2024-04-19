package db_connect

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
	"time"
)

func AgentPing(ctx context.Context, db SQLQueryExec, id int, status, statusText string) int {
	if status == "create" {
		var newId int
		err := db.QueryRowContext(ctx, "INSERT INTO agents (status, status_text, ping_time) VALUES ($1, $2, $3) RETURNING id",
			status, statusText, time.Now().Unix()).Scan(&newId)
		if err != nil {
			return 0
		}
		return newId
	}

	_, err := db.ExecContext(ctx, "UPDATE agents SET status = $1, status_text = $2, ping_time = $3 WHERE id = $4",
		status, statusText, time.Now().Unix(), id)
	if err != nil {
		return 0
	}
	return id
}

func DeleteAgent(ctx context.Context, db SQLQueryExec, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", id)
	return err
}

func GetAllAgents(ctx context.Context, db SQLQueryExec) ([]entities.Agent, error) {
	var agents []entities.Agent
	rows, err := db.QueryContext(ctx, "SELECT id, status, status_text, ping_time FROM agents")
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
