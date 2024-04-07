package api

import (
	"Zadacha/internal/auth"
	"Zadacha/internal/db_connect"
	"Zadacha/internal/entities"
	"Zadacha/internal/my_errors"
	"Zadacha/internal/orchestrator"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	db *sql.DB
	ch *amqp.Channel
}

// getExpressions обработчик GET запроса на /expressions
func (server *Server) getExpressions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exp, err := db_connect.GetAllExpressions(ctx, server.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, _ := json.Marshal(exp)
	w.Write(res)
}

// getExpressionById обработчик для GET запроса на /expressions/{id}
func (server *Server) getExpressionById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("%s not an integer", params["id"]), http.StatusBadRequest)
		return
	}
	opera, err := db_connect.GetExpressionByID(ctx, server.db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("not found expression with id %d", id), http.StatusNotFound)
		return
	}
	res, _ := json.Marshal(opera)
	w.Write(res)
}

// newExpression обработчик POST запроса на /expressions, создаёт выражение
func (server *Server) newExpression(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	expressionBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad input", http.StatusBadRequest)
		return
	}
	expression := string(expressionBytes)
	expressionId, err := orchestrator.StartExpression(ctx, server.ch, server.db, expression)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, my_errors.StrangeSymbolsError) || errors.Is(err, my_errors.StrangeSymbolsError) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d", expressionId)))
}

// getOperations обработчик GET запроса на /operations
// Выводит время выполнения все операций (сколько длится +, сколько - и *, /)
func (server *Server) getOperations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := ctx.Value("userId").(int)
	operations, _ := db_connect.GetOperationsTimeByUserID(ctx, server.db, userId)
	res, _ := json.Marshal(operations)
	w.Write(res)
}

// putOperations обработчик PUT запроса на /operations
func (server *Server) putOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userId := ctx.Value("userId").(int)
	timeBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}
	var operationsTime entities.OperationsTime
	err = json.Unmarshal(timeBytes, &operationsTime)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}

	err = db_connect.UpdateOperationsTimeByUserID(ctx, server.db, operationsTime, userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("server db error"), http.StatusInternalServerError)
		return
	}
}

// getAgents обработчик GET запроса на /agents, возвращает всех агентов
func (server *Server) getAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, err := db_connect.GetAllAgents(ctx, server.db)
	if err != nil {
		http.Error(w, "server db error", http.StatusInternalServerError)
		return
	}
	res, _ := json.Marshal(agents)
	w.Write(res)
}

// login обработчик POST запроса на /login, возвращает всех агентов
func (server *Server) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}

	var user entities.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}

	dbUser, err := db_connect.GetUserByName(ctx, server.db, user.Name)
	if err != nil {
		http.Error(w, "name not found", http.StatusUnauthorized)
		return
	}
	err = auth.ComparePasswordWithHashed(user.Password, dbUser.PasswordHash)
	if err != nil {
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateTokenFromId(dbUser.Id)
	if err != nil {
		http.Error(w, "token generation error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(token))
}

// register обработчик POST запроса на /register, возвращает всех агентов
func (server *Server) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}
	var user entities.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
	}
	user.PasswordHash, err = auth.GenerateHashedPassword(user.Password)
	if err != nil {
		http.Error(w, "hashing error", http.StatusInternalServerError)
	}
	user.Password = ""
	_, err = db_connect.CreateUser(ctx, server.db, user)

	if err != nil {
		if err.(*pq.Error).Code.Name() != "unique_violation" {
			http.Error(w, "server db error", http.StatusInternalServerError)
			return
		}
		http.Error(w, "name is already claimed", http.StatusBadRequest)
		return
	}
	w.Write([]byte("success"))
}

// Run запускает сервер API
func Run() {
	var conn *amqp.Connection
	var err error
	fmt.Println("Кролик загружается")
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
	var db *sql.DB
	log.Println("Загрузка базы данных оркестратора")
	for {
		db, err = db_connect.OpenDb(context.Background(), os.Getenv("POSTGRES_STRING"))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Println("Базы данных оркестратора загружена")
	defer db.Close()
	server := Server{db: db, ch: ch}
	router := mux.NewRouter()

	// Необходим токен
	router.Handle("/expressions", AuthenticationMiddleware(http.HandlerFunc(server.getExpressions))).Methods("GET", "OPTIONS")
	router.Handle("/expressions/{id}", AuthenticationMiddleware(http.HandlerFunc(server.getExpressionById))).Methods("GET", "OPTIONS")
	router.Handle("/expressions", AuthenticationMiddleware(http.HandlerFunc(server.newExpression))).Methods("POST", "OPTIONS")
	router.Handle("/operations", AuthenticationMiddleware(http.HandlerFunc(server.getOperations))).Methods("GET", "OPTIONS")
	router.Handle("/operations", AuthenticationMiddleware(http.HandlerFunc(server.putOperations))).Methods("PUT", "OPTIONS")

	// Токен не нужен (агенты для всех общие)
	router.HandleFunc("/agents", server.getAgents).Methods("GET", "OPTIONS")
	router.HandleFunc("/login", server.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/register", server.register).Methods("POST", "OPTIONS")

	fmt.Println("API запущено на http://localhost:8080 (порт 8080)")
	corsHandler := handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	log.Fatal(http.ListenAndServe(":8080", corsHandler(router)))
}
