package orchestrator

import (
	"Zadacha/internal/db_connect"
	"Zadacha/internal/entities"
	"Zadacha/internal/my_errors"
	"Zadacha/pkg/expressions"
	"context"
	"database/sql"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
	"strings"
	"time"
)

type StackPosition struct {
	id  int     // может быть -1 или id
	num float64 // Есть если id -1
}

// StartExpression полностью инициализирует expression из строки
func StartExpression(ctx context.Context, ch *amqp.Channel, db *sql.DB, expression string) (int, error) {
	var res []int
	var numsStack []StackPosition
	op := entities.Expression{Status: "process", NeedToDo: expression, StartTime: int(time.Now().Unix())}
	finalId, err := db_connect.CreateExpression(ctx, db, op)
	if err != nil {
		return 0, err
	}
	if err := expressions.Validate(expression); err != nil {
		db_connect.OhNoExpressionError(ctx, db, finalId)
		return 0, err
	}
	postfixExpression := expressions.InfixToPostfix(expression)
	for _, s := range strings.Split(postfixExpression, " ") {
		switch []rune(s)[0] {
		case '+', '-', '*', '/':
			dataLeft, dataRight := numsStack[len(numsStack)-2], numsStack[len(numsStack)-1]
			numsStack = numsStack[:len(numsStack)-2]
			operation := entities.Operation{Znak: s, FatherId: -1, ExpressionId: finalId}
			if dataLeft.id == -1 {
				operation.LeftData = dataLeft.num
				operation.LeftIsReady = 1
			}
			if dataRight.id == -1 {
				operation.RightData = dataRight.num
				operation.RightIsReady = 1
			}
			id, err := db_connect.AddOperation(ctx, db, &operation)
			if operation.LeftIsReady == 1 && operation.RightIsReady == 1 {
				res = append(res, operation.Id)
			}
			if err != nil {
				return 0, err
			}
			if dataLeft.id != -1 {
				err = db_connect.UpdateFatherOperation(ctx, db, dataLeft.id, id, 0)
				if err != nil {
					return 0, err
				}
			}
			if dataRight.id != -1 {
				err = db_connect.UpdateFatherOperation(ctx, db, dataRight.id, id, 1)
				if err != nil {
					return 0, err
				}
			}
			numsStack = append(numsStack, StackPosition{id: operation.Id, num: -1})
		default:
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, my_errors.StrangeSymbolsError
			}
			numsStack = append(numsStack, StackPosition{id: -1, num: n})
		}
	}

	for _, i := range res {
		err := ch.PublishWithContext(
			context.Background(),
			"",           // exchange
			"task_queue", // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(fmt.Sprintf("%d", i)),
			})
		if err != nil {
			return 0, err
		}
	}
	if len(res) == 0 {
		exInt, _ := strconv.ParseFloat(expression, 64)
		err := db_connect.UpdateExpression(ctx, db, finalId, exInt, "done")
		if err != nil {
			return 0, err
		}
	}
	return finalId, nil
}
