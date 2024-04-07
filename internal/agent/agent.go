package agent

import (
	"Zadacha/internal/db_connect"
	"Zadacha/internal/entities"
	"context"
	"database/sql"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xlab/closer"
	"log"
	"os"
	"strconv"
	"time"
)

func CreateAgents() {
	count, _ := strconv.Atoi(os.Getenv("AGENT_COUNT"))
	for i := 0; i < count; i++ {
		go Agent(context.Background())
	}
}

func Agent(ctx context.Context) {
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
	log.Println("Базы данных агента загружена")
	defer db.Close()

	id := db_connect.AgentPing(ctx, db, 1, "create", "Агент запускается")
	defer db_connect.DeleteAgent(ctx, db, id)
	closer.Bind(func() {
		db_connect.DeleteAgent(ctx, db, id)
	})

	var conn *amqp.Connection
	for {
		conn, err = amqp.Dial("amqp://rmuser:rmpassword@rabbitmq:5672/")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		// Ждём пока запустится кролик
	}
	defer conn.Close()

	// Создаем канал
	ch, err := conn.Channel()
	if err != nil {
		db_connect.AgentPing(ctx, db, id, "error", "Ошибка создания канала RabbitMQ")
		return
	}
	defer ch.Close()
	// Объявляем очередь
	_, err = ch.QueueDeclare(
		"task_queue", // Имя очереди
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		db_connect.AgentPing(ctx, db, id, "error", "Ошибка создания очереди RabbitMQ")
		return
	}

	msgs, err := ch.Consume(
		"task_queue", // Очередь
		"",           // Consumer
		false,        // AutoAck: автоматически подтверждать сообщения
		false,        // Exclusive
		false,        // NoLocal
		false,        // NoWait
		nil,          // Args
	)
	if err != nil {
		db_connect.AgentPing(ctx, db, id, "error", "Ошибка чтения очереди RabbitMQ")
		return
	}
	db_connect.AgentPing(ctx, db, id, "waiting", "")
	var task entities.Operation
	var expressionId int
	closer.Bind(func() {
		// Чистое отключение агента
		db_connect.DeleteOperation(ctx, db, task.Id)
		if a, _ := db_connect.GetExpressionByID(ctx, db, expressionId); a.Status == "process" {
			db_connect.OhNoExpressionError(ctx, db, expressionId)
		}

	})
	for message := range msgs {
		err := message.Ack(false)
		if err != nil {
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			continue
		}
		taskId, err := strconv.Atoi(string(message.Body))
		if err != nil {
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			continue
		}
		task, err = db_connect.GetOperation(ctx, db, taskId)
		if err != nil {
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			db_connect.DeleteOperation(ctx, db, taskId)
			continue
		}
		db_connect.AgentPing(ctx, db, id, "process", fmt.Sprintf("%f %s %f", task.LeftData, task.Znak, task.RightData))

		var res float64
		expression, err := db_connect.GetExpressionByID(ctx, db, task.ExpressionId)
		if err != nil {
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			db_connect.DeleteOperation(ctx, db, taskId)
			continue
		}
		operationsTime, err := db_connect.GetOperationsTimeByUserID(ctx, db, expression.UserId)
		if err != nil {
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			db_connect.DeleteOperation(ctx, db, taskId)
			continue
		}
		switch task.Znak {
		case "+":
			time.Sleep(time.Duration(operationsTime.Plus) * time.Second)
			res = task.LeftData + task.RightData
		case "-":
			time.Sleep(time.Duration(operationsTime.Minus) * time.Second)
			res = task.LeftData - task.RightData
		case "*":
			time.Sleep(time.Duration(operationsTime.Multiplication) * time.Second)
			res = task.LeftData * task.RightData
		case "/":
			if task.RightData == 0 {
				db_connect.OhNoExpressionError(ctx, db, task.ExpressionId)
				db_connect.DeleteOperation(ctx, db, task.Id)
				db_connect.AgentPing(ctx, db, id, "waiting", "")
				continue
			}
			time.Sleep(time.Duration(operationsTime.Division) * time.Second)
			res = task.LeftData / task.RightData
		}
		if expression.Status != "process" { // Закончено вычисление финального выражения из за ошибки или чего то ещё
			db_connect.DeleteOperation(ctx, db, task.Id)
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			continue
		}
		if task.FatherId == -1 {
			// Это была последняя операция в выражении
			db_connect.UpdateExpression(ctx, db, task.ExpressionId, res, "done")
			db_connect.DeleteOperation(ctx, db, task.Id)
			db_connect.AgentPing(ctx, db, id, "waiting", "")
			continue
		}
		if task.SonSide == 0 {
			err := db_connect.UpdateLeftOperation(ctx, db, task.FatherId, res)
			if err != nil {
				db_connect.AgentPing(ctx, db, id, "waiting", "")
				db_connect.OhNoExpressionError(ctx, db, task.ExpressionId)
				db_connect.DeleteOperation(ctx, db, taskId)
				continue
			}
		} else {
			err := db_connect.UpdateRightOperation(ctx, db, task.FatherId, res)
			if err != nil {
				db_connect.AgentPing(ctx, db, id, "waiting", "")
				db_connect.OhNoExpressionError(ctx, db, task.ExpressionId)
				db_connect.DeleteOperation(ctx, db, taskId)
				continue
			}
		}
		if ready, _ := db_connect.IsReadyToExecuteOperation(ctx, db, task.FatherId); ready {
			// Отправляем операцию выше в очередь на выполнение
			go func(task entities.Operation) {
				err = ch.PublishWithContext(
					context.Background(),
					"",           // exchange
					"task_queue", // routing key
					false,        // mandatory
					false,        // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(fmt.Sprintf("%d", task.FatherId)),
					})
				if err != nil {
					db_connect.AgentPing(ctx, db, id, "waiting", "")
					db_connect.OhNoExpressionError(ctx, db, task.ExpressionId)
				}
			}(task)
		}
		db_connect.DeleteOperation(ctx, db, task.Id)
		db_connect.AgentPing(ctx, db, id, "waiting", "")
	}
}
