package api

import (
	"Zadacha/internal/db_connect"
	"Zadacha/internal/orchestrator"
	"Zadacha/pkg/my_errors"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
		http.Error(w, fmt.Sprintf("%s not an integer", params["id"]), http.StatusBadRequest)
		return
	}
	opera, err := server.db.GetFinalOperationByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("not found expression with id %d", id), http.StatusNotFound)
		return
	}
	res, _ := json.Marshal(opera)
	w.Write(res)
}

// newExpression обработчик POST запроса на /expressions, создаёт выражение
func (server *Server) newExpression(w http.ResponseWriter, r *http.Request) {
	expressionBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad input", http.StatusBadRequest)
		return
	}
	expression := string(expressionBytes)
	finalId, err := orchestrator.StartExpression(server.ch, server.db, expression)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, my_errors.StrangeSymbolsError) || errors.Is(err, my_errors.StrangeSymbolsError) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d", finalId)))
}

// getOperations обработчик GET запроса на /operations
// Выводит время выполнения все операций (сколько длится +, сколько - и *, /)
func (server *Server) getOperations(w http.ResponseWriter, r *http.Request) {
	operations, _ := server.db.GetOperationTimeByID()
	res, _ := json.Marshal(operations)
	w.Write(res)
}

// putOperation обработчик PUT запроса на /operations/{id}
func (server *Server) putOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("%s not an input", params["id"]), http.StatusBadRequest)
		return
	}
	timeBytes, err := io.ReadAll(r.Body)
	timeInt, err := strconv.Atoi(string(timeBytes))
	if err != nil || timeInt < 0 {
		http.Error(w, "wrong time format", http.StatusBadRequest)
		return
	}
	err = server.db.UpdateOperationTime(id, timeInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("not found expression with id %d", id), http.StatusNotFound)
		return
	}
}

// getAgents обработчик POST запроса на /agents, возвращает всех агентов
func (server *Server) getAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := server.db.GetAllAgents()
	if err != nil {
		http.Error(w, "Произошла ошибка при получении агентов", http.StatusNotFound)
		return
	}
	res, _ := json.Marshal(agents)
	w.Write(res)
}

// Run запускает сервер API
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
	router.HandleFunc("/expressions", server.newExpression).Methods("POST", "OPTIONS")

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
	log.Fatal(http.ListenAndServe(":8080", corsHandler(AuthenticationMiddleware(router))))
}
