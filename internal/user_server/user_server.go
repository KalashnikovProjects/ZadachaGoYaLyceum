package user_server

import (
	"fmt"
	"net/http"
	"os"
)

// Тут ничего интересного, просто возвращает html

func Run() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/agents", agentsPage)
	http.HandleFunc("/expressions", expressionsPage)
	http.HandleFunc("/operations", operationsPage)
	fmt.Println("User server запущен на  http://localhost (порт 80)")
	http.ListenAndServe(":80", nil)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func agentsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/agents.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func expressionsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/expressions.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func operationsPage(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/operations.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}
