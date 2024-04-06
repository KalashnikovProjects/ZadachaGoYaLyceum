package agent

import (
	"Zadacha/internal/db_connect"
	"Zadacha/internal/entities"
	"context"
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
		go Agent()
	}
}

func Agent() {
	var db db_connect.DBConnection
	var err error
	log.Println("Загрузка базы данных агента")
	for {
		db, err = db_connect.OpenDb(os.Getenv("POSTGRES_STRING"))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Println("Базы данных агента загружена")
	defer db.Close()

	id := db.AgentPing(1, "create", "Агент запускается")
	defer db.DeleteAgent(id)
	closer.Bind(func() {
		db.DeleteAgent(id)
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
		db.AgentPing(id, "my_errors", "Ошибка создания канала RabbitMQ")
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
		db.AgentPing(id, "my_errors", "Ошибка создания очереди RabbitMQ")
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
		db.AgentPing(id, "my_errors", "Ошибка чтения очереди RabbitMQ")
		return
	}
	db.AgentPing(id, "waiting", "")
	var task entities.Operation
	var FinalOperId int
	closer.Bind(func() {
		// Чистое отключение агента
		db.Delete(task.Id)
		if a, _ := db.GetFinalOperationByID(FinalOperId); a.Status == "process" {
			db.OhNoFinalOperationError(FinalOperId)
		}

	})
	for message := range msgs {
		err := message.Ack(false)
		if err != nil {
			db.AgentPing(id, "waiting", "")
			continue
		}
		taskId, err := strconv.Atoi(string(message.Body))
		if err != nil {
			db.AgentPing(id, "waiting", "")
			continue
		}
		task, err = db.GetOperation(taskId)
		if err != nil {
			db.AgentPing(id, "waiting", "")
			db.Delete(taskId)
			continue
		}
		FinalOperId = task.FinalOperationId
		db.AgentPing(id, "process", fmt.Sprintf("%f %s %f", task.LeftData, task.Znak, task.RightData))

		var res float64
		switch task.Znak {
		case "+":
			t, _ := db.GetOperationTimeByID(1)
			time.Sleep(time.Duration(t.Plus) * time.Second)
			res = task.LeftData + task.RightData
		case "-":
			t, _ := db.GetOperationTimeByID(2)
			time.Sleep(time.Duration(t.Minus) * time.Second)
			res = task.LeftData - task.RightData
		case "*":
			t, _ := db.GetOperationTimeByID(3)
			time.Sleep(time.Duration(t.Multiplication) * time.Second)
			res = task.LeftData * task.RightData
		case "/":
			if task.RightData == 0 {
				db.OhNoFinalOperationError(task.FinalOperationId)
				db.Delete(task.Id)
				db.AgentPing(id, "waiting", "")
				continue
			}
			t, _ := db.GetOperationTimeByID(4)
			time.Sleep(time.Duration(t.Division) * time.Second)
			res = task.LeftData / task.RightData
		}
		finOper, err := db.GetFinalOperationByID(task.FinalOperationId)
		if finOper.Status != "process" { // Закончено вычисление финального выражения из за ошибки или чего то ещё
			db.Delete(task.Id)
			db.AgentPing(id, "waiting", "")
			continue
		}
		if task.FatherId == -1 {
			// Это была последняя операция в выражении
			db.UpdateFinalOperation(task.FinalOperationId, res, "done")
			db.Delete(task.Id)
			db.AgentPing(id, "waiting", "")
			continue
		}
		if task.SonSide == 0 {
			err := db.UpdateLeft(task.FatherId, res)
			if err != nil {
				db.AgentPing(id, "waiting", "")
				db.OhNoFinalOperationError(task.FinalOperationId)
				db.Delete(taskId)
				continue
			}
		} else {
			err := db.UpdateRight(task.FatherId, res)
			if err != nil {
				db.AgentPing(id, "waiting", "")
				db.OhNoFinalOperationError(task.FinalOperationId)
				db.Delete(taskId)
				continue
			}
		}
		if ready, _ := db.IsReadyToExecute(task.FatherId); ready {
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
					db.AgentPing(id, "waiting", "")
					db.OhNoFinalOperationError(task.FinalOperationId)
				}
			}(task)
		}
		db.Delete(task.Id)
		db.AgentPing(id, "waiting", "")
	}
}
