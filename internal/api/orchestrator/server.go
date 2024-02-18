package orchestrator

import (
	"Zadacha/config"
	"Zadacha/internal/api/db_connect"
	"Zadacha/pkg/expressions"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type StackPosition struct {
	id  int     // может быть -1 или id
	num float64 // Есть если id -1
}

type Server struct {
	db db_connect.DBConnection
	ch *amqp.Channel
}

// getExpressions обработчик GET запроса на /expressions
func (server *Server) getExpressions(w http.ResponseWriter, r *http.Request) {
	res, _ := json.Marshal(server.db.GetAllFinalOperations())
	w.Write(res)
}

// getExpressionById обработчик для GET запроса на /expressions/{id}
func (server *Server) getExpressionById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("%s не является числом", params["id"]), http.StatusBadRequest)
		return
	}
	opera, err := server.db.GetFinalOperationByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Не найдено выражение под id %d", id), http.StatusNotFound)
		return
	}
	res, _ := json.Marshal(opera)
	w.Write(res)
}

// startExpression обработчик POST запроса на /expressions, создаёт выражение
func (server *Server) startExpression(w http.ResponseWriter, r *http.Request) {
	expressionBytes, err := io.ReadAll(r.Body)
	expression := string(expressionBytes)
	if err != nil {
		http.Error(w, "Неверный ввод", http.StatusBadRequest)
		return
	}
	var res []int
	var numsStack []StackPosition
	op := config.FinalOperation{Status: "process", NeedToDo: expression, StartTime: int(time.Now().Unix())}
	finalId, err := server.db.CreateFinalOperation(op)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if !expressions.Validate(expression) {
		server.db.OhNoFinalOperationError(finalId)
		http.Error(w, "Неверное математическое выражение", http.StatusBadRequest)
		return
	}
	postfixExpression := expressions.InfixToPostfix(expression)
	for _, s := range strings.Split(postfixExpression, " ") {
		switch []rune(s)[0] {
		case '+', '-', '*', '/':
			dataLeft, dataRight := numsStack[len(numsStack)-2], numsStack[len(numsStack)-1]
			numsStack = numsStack[:len(numsStack)-2]
			operation := config.Operation{Znak: s, FatherId: -1, FinalOperationId: finalId}
			if dataLeft.id == -1 {
				operation.LeftData = dataLeft.num
				operation.LeftIsReady = 1
			}
			if dataRight.id == -1 {
				operation.RightData = dataRight.num
				operation.RightIsReady = 1
			}
			id, err := server.db.Add(&operation)
			if operation.LeftIsReady == 1 && operation.RightIsReady == 1 {
				res = append(res, operation.Id)
			}
			if err != nil {
				http.Error(w, "Ошибка при записи в базу данных.", http.StatusInternalServerError)
				return
			}
			if dataLeft.id != -1 {
				err = server.db.UpdateFather(dataLeft.id, id, 0)
				if err != nil {
					http.Error(w, "Ошибка при записи в базу данных.", http.StatusInternalServerError)
					return
				}
			}
			if dataRight.id != -1 {
				err = server.db.UpdateFather(dataRight.id, id, 1)
				if err != nil {
					http.Error(w, "Ошибка при записи в базу данных.", http.StatusInternalServerError)
					return
				}
			}
			numsStack = append(numsStack, StackPosition{id: operation.Id, num: -1})
		default:
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				http.Error(w, "Какие то странные знаки у вас в выражении привели к ошибке.", http.StatusBadRequest)
				return
			}
			numsStack = append(numsStack, StackPosition{id: -1, num: n})
		}
	}

	for _, i := range res {
		err := server.ch.PublishWithContext(
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
			http.Error(w, "Ошибка с брокером сообщений", http.StatusInternalServerError)
		}
	}
	w.Write([]byte(fmt.Sprintf("%d", finalId)))
	if len(res) == 0 {
		exInt, _ := strconv.ParseFloat(expression, 64)
		server.db.UpdateFinalOperation(finalId, exInt, "done")
	}
}

// getOperations обработчик GET запроса на /operations
// Выводит время выполнения все операций (сколько длится +, сколько - и *, /)
func (server *Server) getOperations(w http.ResponseWriter, r *http.Request) {
	operations := server.db.GetAllOperationTime()
	slices.SortFunc(operations, func(a, b config.OperationTime) int {
		if a.Id < b.Id {
			return -1
		}
		if a.Id > b.Id {
			return 1
		}
		return 0
	})
	res, _ := json.Marshal(operations)
	w.Write(res)
}

// putOperation обработчик PUT запроса на /operations/{id}
func (server *Server) putOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("%s не является числом", params["id"]), http.StatusBadRequest)
		return
	}
	timeBytes, err := io.ReadAll(r.Body)
	timeInt, err := strconv.Atoi(string(timeBytes))
	if err != nil || timeInt < 0 {
		http.Error(w, "Неверный формат времени (нужно число секунд)", http.StatusBadRequest)
		return
	}
	err = server.db.UpdateOperationTime(id, timeInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Не найдена операция под id %d", id), http.StatusNotFound)
		return
	}
}

// getAgents обработчик POST запроса на /agents, возвращает всех агентов
func (server *Server) getAgents(w http.ResponseWriter, r *http.Request) {
	res, _ := json.Marshal(server.db.GetAllAgents())
	w.Write(res)
}

// Run запускает сервер оркестратора
func Run() {
	var conn *amqp.Connection
	var err error
	for {
		conn, err = amqp.Dial("amqp://rmuser:rmpassword@rabbitmq:5672/")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		// Ждём пока запустится кролик
	}
	defer conn.Close()
	fmt.Println("Кролик загрузился")
	// Создаем канал
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Не удалось открыть канал: %v", err)
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
		log.Fatalf("Не удалось объявить очередь: %v", err)
	}
	var db db_connect.DBConnection
	log.Println("Загрузка базы данных оркестратора")
	for {
		db, err = db_connect.OpenDb(os.Getenv("POSTGRES_STRING"))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Println("Базы данных оркестратора загружена")
	defer db.Close()
	server := Server{db: db, ch: ch}
	router := mux.NewRouter()

	router.HandleFunc("/expressions", server.getExpressions).Methods("GET", "OPTIONS")
	router.HandleFunc("/expressions/{id}", server.getExpressionById).Methods("GET", "OPTIONS")
	router.HandleFunc("/expressions", server.startExpression).Methods("POST", "OPTIONS")

	router.HandleFunc("/operations", server.getOperations).Methods("GET", "OPTIONS")
	router.HandleFunc("/operations/{id}", server.putOperation).Methods("PUT", "OPTIONS")
	router.HandleFunc("/agents", server.getAgents).Methods("GET", "OPTIONS")

	fmt.Println("API запущено на http://localhost:8080 (порт 8080)")
	corsHandler := handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	log.Fatal(http.ListenAndServe(":8080", corsHandler(router)))
}
