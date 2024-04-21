package user_server

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

// Тут рендерятся шаблоны html

type TemplateData struct {
	Page   string
	Logged bool
}

func Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", homePage)
	router.HandleFunc("/agents", agentsPage)
	router.HandleFunc("/expressions", expressionsPage)
	router.HandleFunc("/operations", operationsPage)
	router.HandleFunc("/login", loginPage)
	fmt.Println("User server запущен на  http://localhost (порт 80)")
	log.Fatal(http.ListenAndServe(":80", CheckTokenCookie(router, []string{"/", "/agents", "/login"})))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("templates/base.html", "templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{"Главная", r.Context().Value("logged").(bool)})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func agentsPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("templates/base.html", "templates/agents.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{"Мониторинг воркеров", r.Context().Value("logged").(bool)})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func expressionsPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("templates/base.html", "templates/expressions.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println(r.Context().Value("logged").(bool))
	err = ts.Execute(w, TemplateData{"Статусы выражений", r.Context().Value("logged").(bool)})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func operationsPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("templates/base.html", "templates/operations.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{"Длительность выполнения операций", r.Context().Value("logged").(bool)})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("templates/base.html", "templates/login.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{"Вход в аккаунт", r.Context().Value("logged").(bool)})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}
