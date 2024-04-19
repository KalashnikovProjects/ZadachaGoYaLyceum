package agent

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/db_connect"
	pb "github.com/KalashnikovProjects/ZadachaGoYaLyceum/proto"
	"github.com/xlab/closer"
	"log"
	"os"
	"time"
)

func Agent(ctx context.Context, tasks chan *TaskAgent) {
	var db *sql.DB
	var err error
	log.Println("Загрузка базы данных агента")
	for {
		db, err = db_connect.OpenDb(ctx, os.Getenv("POSTGRES_STRING"))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Println("База данных агента загружена")
	defer db.Close()

	id := db_connect.AgentPing(ctx, db, 1, "create", "Агент запускается")
	defer db_connect.DeleteAgent(ctx, db, id)
	closer.Bind(func() {
		// Чистое отключение агента при смерти сервера
		db_connect.DeleteAgent(ctx, db, id)
	})

	db_connect.AgentPing(ctx, db, id, "waiting", "")
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-tasks:
			operation := task.task
			db_connect.AgentPing(ctx, db, id, "process", fmt.Sprintf("%f %s %f", operation.Left, operation.Znak, operation.Right))

			var res float32
			switch operation.Znak {
			case "+":
				time.Sleep(time.Duration(operation.Times.Plus) * time.Second)
				res = operation.Left + operation.Right
			case "-":
				time.Sleep(time.Duration(operation.Times.Minus) * time.Second)
				res = operation.Left - operation.Right
			case "*":
				time.Sleep(time.Duration(operation.Times.Multiplication) * time.Second)
				res = operation.Left * operation.Right
			case "/":
				if operation.Right == 0 {
					task.result <- &pb.OperationResponse{Status: "error", Result: 0}
					db_connect.AgentPing(ctx, db, id, "waiting", "")
					continue
				}
				time.Sleep(time.Duration(operation.Times.Division) * time.Second)
				res = operation.Left / operation.Right
			}

			task.result <- &pb.OperationResponse{Status: "ok", Result: res}
			db_connect.AgentPing(ctx, db, id, "waiting", "")
		}
	}
}
