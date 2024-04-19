package user_server

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

// Тут ничего интересного, просто возвращает html и проверка аунтефикации

func Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", homePage)
	router.HandleFunc("/agents", agentsPage)
	router.HandleFunc("/expressions", expressionsPage)
	router.HandleFunc("/operations", operationsPage)
	router.HandleFunc("/login", operationsPage)
	fmt.Println("User server запущен на  http://localhost (порт 80)")
	log.Fatal(http.ListenAndServe(":80", router))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(html)
}

func agentsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/agents.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(html)
}

func expressionsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/expressions.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(html)
}

func operationsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/operations.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(html)
}
