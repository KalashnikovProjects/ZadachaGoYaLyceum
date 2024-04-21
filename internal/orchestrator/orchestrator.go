package orchestrator

import (
	"context"
	"database/sql"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/db_connect"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/my_errors"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/pkg/expressions"
	pb "github.com/KalashnikovProjects/ZadachaGoYaLyceum/proto"
	"log"
	"strconv"
	"strings"
	"time"
)

type StackPosition struct {
	id  int     // может быть -1 или id
	num float64 // Есть если id -1
}

// InitOrchestrator восстанавливает вычисления после перезапуска оркестратора
func InitOrchestrator(ctx context.Context, db db_connect.SQLTXQueryExec, gRPCClient pb.AgentsServiceClient) {
	ids, err := db_connect.GetReadyToExecuteOperations(ctx, db)
	if err != nil {
		return
	}
	for _, i := range ids {
		i := i
		go ProcessOperation(db, gRPCClient, i.OperationId, i.ExpressionId, i.OperationsTimeId)
	}
}

// StartExpression инициализирует вычисление expression из строки
func StartExpression(ctx context.Context, db db_connect.SQLTXQueryExec, gRPCClient pb.AgentsServiceClient, expression string) (int, error) {
	var res []int
	var numsStack []StackPosition
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	userId := ctx.Value("userId").(int)
	op := entities.Expression{Status: "process", NeedToDo: expression, StartTime: int(time.Now().Unix()), UserId: userId}
	expressionId, err := db_connect.CreateExpression(ctx, tx, op)
	if err != nil {
		return 0, err
	}
	if err := expressions.Validate(expression); err != nil {
		db_connect.OhNoExpressionError(ctx, tx, expressionId)
		return 0, err
	}
	postfixExpression := expressions.InfixToPostfix(expression)
	for _, s := range strings.Split(postfixExpression, " ") {
		switch []rune(s)[0] {
		case '+', '-', '*', '/':
			dataLeft, dataRight := numsStack[len(numsStack)-2], numsStack[len(numsStack)-1]
			numsStack = numsStack[:len(numsStack)-2]
			operation := entities.Operation{Znak: s, FatherId: -1, ExpressionId: expressionId}
			if dataLeft.id == -1 {
				operation.LeftData = dataLeft.num
				operation.LeftIsReady = 1
			}
			if dataRight.id == -1 {
				operation.RightData = dataRight.num
				operation.RightIsReady = 1
			}
			id, err := db_connect.AddOperation(ctx, tx, &operation)
			operation.Id = id
			if operation.LeftIsReady == 1 && operation.RightIsReady == 1 {
				res = append(res, id)
			}
			if err != nil {
				return 0, err
			}
			if dataLeft.id != -1 {
				err = db_connect.UpdateFatherOperation(ctx, tx, dataLeft.id, id, 0)
				if err != nil {
					return 0, err
				}
			}
			if dataRight.id != -1 {
				err = db_connect.UpdateFatherOperation(ctx, tx, dataRight.id, id, 1)
				if err != nil {
					return 0, err
				}
			}
			numsStack = append(numsStack, StackPosition{id: id, num: -1})
		default:
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, my_errors.StrangeSymbolsError
			}
			numsStack = append(numsStack, StackPosition{id: -1, num: n})
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	timesId, err := db_connect.GetOperationsTimeByUserID(ctx, db, userId)
	if err != nil {
		return 0, err
	}
	for _, i := range res {
		i := i
		go ProcessOperation(db, gRPCClient, i, expressionId, timesId.Id)
	}
	if len(res) == 0 {
		exInt, _ := strconv.ParseFloat(expression, 64)
		err := db_connect.UpdateExpression(ctx, db, expressionId, exInt, "done")
		if err != nil {
			log.Println(err)
			return 0, err
		}
	}
	return expressionId, nil
}

func ProcessOperation(db db_connect.SQLQueryExec, gRPCClient pb.AgentsServiceClient, operationId, expressionId, timesId int) {
	ctx := context.Background()
	defer db_connect.DeleteOperation(ctx, db, operationId)
	opera, err := db_connect.GetOperationByID(ctx, db, operationId)
	if err != nil {
		db_connect.OhNoExpressionError(ctx, db, expressionId)
		return
	}
	times, err := db_connect.GetOperationsTimeByID(ctx, db, timesId)
	if err != nil {
		db_connect.OhNoExpressionError(ctx, db, expressionId)
		return
	}
	timesReq := pb.OperationTimes{Plus: int32(times.Plus), Minus: int32(times.Minus), Division: int32(times.Division), Multiplication: int32(times.Multiplication)}
	operationRequest := &pb.OperationRequest{
		Znak:  opera.Znak,
		Left:  float32(opera.LeftData),
		Right: float32(opera.RightData),
		Times: &timesReq,
	}

	// TODO: таймаут для grpc, после чего перепопытка,
	operationResponse, err := gRPCClient.ExecuteOperation(ctx, operationRequest)
	if err != nil || operationResponse.Status == "error" {
		log.Println(err)
		db_connect.OhNoExpressionError(ctx, db, expressionId)
		return
	}
	res := float64(operationResponse.Result)
	expression, err := db_connect.GetExpressionByID(ctx, db, expressionId)
	if err != nil || expression.Status != "process" {
		// Выражение завершилось до этого (ошибка или как-то по другому)
		db_connect.OhNoExpressionError(ctx, db, expressionId)
		return
	}
	if opera.FatherId == -1 {
		// Это была последняя операция в выражении
		err := db_connect.UpdateExpression(ctx, db, expressionId, res, "done")
		if err != nil {
			db_connect.OhNoExpressionError(ctx, db, expressionId)
			return
		}
		return
	}
	if opera.SonSide == 0 {
		err := db_connect.UpdateLeftOperation(ctx, db, opera.FatherId, res)
		if err != nil {
			db_connect.OhNoExpressionError(ctx, db, expressionId)
			return
		}
	} else {
		err := db_connect.UpdateRightOperation(ctx, db, opera.FatherId, res)
		if err != nil {
			db_connect.OhNoExpressionError(ctx, db, expressionId)
			return
		}
	}
	if ready, _ := db_connect.IsReadyToExecuteOperation(ctx, db, opera.FatherId); ready {
		// Отправляем операцию уровнем выше на выполнение
		go ProcessOperation(db, gRPCClient, opera.FatherId, expressionId, timesId)
	}
}
