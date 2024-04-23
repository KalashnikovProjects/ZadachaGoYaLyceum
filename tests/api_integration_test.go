package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/agent"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/api"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	name     string
	method   string
	url      string
	body     string
	expected int
}

type testCaseExpression struct {
	name           string
	method         string
	url            string
	body           string
	expectedStatus int
	auth           bool
	token          string

	willCreated bool
	id          int
	resultState string
	result      float64
}

func TestApi(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := RunPostgresContainer(ctx)
	if err != nil {
		t.Errorf("error running postgres container: %v", err)
		return
	}
	t.Cleanup(func() {
		if pgContainer.Terminate(context.Background()) != nil {
			t.Errorf("error terminate postgres container: %v", err)
			return
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Logf("error creating connection string: %v", err)
		return
	}
	t.Setenv("POSTGRES_STRING", connStr)
	t.Setenv("HMAC", "GGGGGGGGGG231241GEAW")
	t.Setenv("AGENT_COUNT", "5")
	t.Setenv("AGENT_ADDR", "localhost:9090")

	t.Log("Запускается api сервер и агент...")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go agent.ManagerAgent(ctx)
	go api.Run(ctx)
	time.Sleep(10 * time.Second)
	t.Run("User system tests", UserSystemUnderTest)
	t.Run("Auth middleware tests", AuthMiddlewareUnderTest)
	t.Run("Expressions endpoint tests", ExpressionsUnderTest)
}

// TestUserSystem интеграционный тест системы пользователей
func UserSystemUnderTest(t *testing.T) {
	// Сначала нужно запустить бд и orchestrator в докере

	http.Post("http://localhost:8080/register", "application/json", strings.NewReader(`{"name": "my_name", "password": "12345678"}`))
	testCases := []testCase{
		{
			name:     "name already taken",
			method:   "POST",
			url:      "http://localhost:8080/register",
			body:     `{"name": "my_name", "password": "abc"}`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "login with correct credentials",
			method:   "POST",
			url:      "http://localhost:8080/login",
			body:     `{"name": "my_name", "password": "12345678"}`,
			expected: http.StatusOK,
		},
		{
			name:     "login with incorrect name",
			method:   "POST",
			url:      "http://localhost:8080/login",
			body:     `{"name": "my_nafme", "password": "12345678"}`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "login with incorrect password",
			method:   "POST",
			url:      "http://localhost:8080/login",
			body:     `{"name": "my_name", "password": "12345f6awfawf78"}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			if err != nil {
				t.Errorf("failed to create request: %v", err)
				return
			}
			resp.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			res, err := client.Do(resp)
			if err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}
			defer res.Body.Close()
			text, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.expected {
				t.Errorf("unexpected status code: got %d, want %d, %s", res.StatusCode, tc.expected, text)
				return
			}
		})
	}
}

func AuthMiddlewareUnderTest(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/expressions")
	if err != nil {
		t.Errorf("error with request: %v", err)
		return
	}

	if resp.StatusCode != 401 {
		t.Errorf("/expressions status code must be: %d, got %d", 401, resp.StatusCode)
		return
	}

	resp, err = http.Get("http://localhost:8080/operations")
	if err != nil {
		t.Errorf("error with request: %v", err)
		return
	}

	if resp.StatusCode != 401 {
		t.Errorf("/operations status code must be: %d, got %d", 401, resp.StatusCode)
		return
	}

	resp, err = http.Get("http://localhost:8080/agents")
	if err != nil {
		t.Errorf("error with request: %v", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("/agents is not require a authorithation. Error: %s", resp.Body)
		return
	}
}

// TestExpressions интеграционные тесты для эндпоинтов /expressions
func ExpressionsUnderTest(t *testing.T) {
	// Регистрация пользователя для получения токена
	resp, err := http.Post("http://localhost:8080/register", "application/json", strings.NewReader(`{"name": "test_user", "password": "password123"}`))
	if err != nil {
		t.Errorf("failed to register user: %v", err)
		return
	}
	defer resp.Body.Close()

	// Получение токена
	resp, err = http.Post("http://localhost:8080/login", "application/json", strings.NewReader(`{"name": "test_user", "password": "password123"}`))
	if err != nil {
		t.Errorf("failed to login: %v", err)
		return
	}
	defer resp.Body.Close()
	token, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read token: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/operations", strings.NewReader(`{"plus": 1, "minus": 1, "division": 1, "multiplication": 1}`))
	if err != nil {
		t.Errorf("failed to register user: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+string(token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("failed to put operations: %v", err)
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, 200)
		return
	}
	defer resp.Body.Close()

	req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/operations", strings.NewReader(``))
	if err != nil {
		t.Errorf("failed to register user: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+string(token))
	req.Header.Set("Content-Type", "application/json")

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("failed to register user: %v", err)
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, 200)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
		return
	}
	var bodyBilder entities.OperationsTime
	err = json.Unmarshal(body, &bodyBilder)
	if err != nil {
		t.Errorf("failed to unmarshal body: %v", err)
		return
	}
	if bodyBilder.Plus != 1 || bodyBilder.Minus != 1 || bodyBilder.Division != 1 || bodyBilder.Multiplication != 1 {
		t.Errorf("operations wrong answer, expected {plus: 1, minus: 1, division: 1, multiplication: 1}, got %s", body)
		return
	}
	createCases := []testCaseExpression{
		{
			name:           "create new expression",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "2+3",
			expectedStatus: http.StatusOK,
			auth:           true,
			token:          string(token),

			willCreated: true,
			resultState: "done",
			result:      5,
		},
		{
			name:           "create bad expression",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "2+asdsad3",
			expectedStatus: http.StatusBadRequest,
			auth:           true,
			token:          string(token),

			willCreated: false,
		},
		{
			name:           "create big expression",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "2+ (2* 5 / 1 + 15 - 1)",
			expectedStatus: http.StatusOK,
			auth:           true,
			token:          string(token),

			willCreated: true,
			resultState: "process",
			result:      0,
		},
		{
			name:           "create digit expression",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "2",
			expectedStatus: http.StatusOK,
			auth:           true,
			token:          string(token),

			willCreated: true,
			resultState: "done",
			result:      2,
		},
		{
			name:           "create empty expression",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			auth:           true,
			token:          string(token),

			willCreated: false,
		},
		{
			name:           "no auth",
			method:         "POST",
			url:            "http://localhost:8080/expressions",
			body:           "",
			expectedStatus: http.StatusUnauthorized,
			auth:           false,
			token:          "",

			willCreated: false,
		},
	}
	created := 0

	for i, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			if err != nil {
				t.Errorf("failed to create request: %v", err)
				return
			}
			if tc.auth {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}
			defer res.Body.Close()

			if tc.willCreated {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("failed to read body: %v", err)
					return
				}
				ibody, err := strconv.Atoi(string(body))
				if err != nil {
					t.Errorf("failed to read body: %v", err)
					return
				}
				createCases[i].id = ibody
				created++
			}
			if res.StatusCode != tc.expectedStatus {
				t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, tc.expectedStatus)
				return
			}
		})
	}
	time.Sleep(3 * time.Second)
	req, err = http.NewRequest("GET", "http://localhost:8080/expressions", strings.NewReader(""))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

	client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, 200)
		return
	}
	var response []entities.Expression
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Errorf("failed to unmarshal body: %v", err)
		return
	}
	log.Println(response)
	if len(response) != created {
		t.Errorf("not enought results, need %d, got %d", created, len(response))
		return
	}
	for _, ex := range createCases {
		if !ex.willCreated {
			continue
		}
		t.Run(fmt.Sprintf("post processing expression %s", ex.name), func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/expressions/%d", ex.id), strings.NewReader(""))
			if err != nil {
				t.Errorf("failed to create request: %v", err)
				return
			}
			req.Header.Set("Authorization", "Bearer "+string(token))

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}
			if res.StatusCode != 200 {
				t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, 200)
				return
			}
			var respon entities.Expression
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("failed to read body: %v", err)
				return
			}
			err = json.Unmarshal(body, &respon)
			if err != nil {
				t.Errorf("failed to unmarshal body: %v", err)
				return
			}
			if ex.resultState != respon.Status || ex.result != respon.Result {
				t.Errorf("wrong answer, expected: status: want %s, got %s; result: want %f, got %f", ex.resultState, respon.Status, ex.result, respon.Result)
				return
			}
		})
	}
}
