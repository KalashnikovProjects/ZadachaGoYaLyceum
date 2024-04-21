package main

import (
	"encoding/json"
	"fmt"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/entities"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

/*
Интеграционные тесты для api
Собирают проект через docker
TestAll запускает докер и все остальные под-тесты
*/

type testCase struct {
	name     string
	method   string
	url      string
	body     string
	expected int
}

type testCaseWithTokens struct {
	name     string
	method   string
	url      string
	body     string
	expected int
	auth     bool
	token    string
}

func runDockerComposeUp() error {
	cmd := exec.Command("docker-compose", "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %v\nOutput: %s\n", err, output)
	}

	return nil
}

func runDockerComposeDown() error {
	cmd := exec.Command("docker-compose", "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %v\nOutput: %s\n", err, output)
	}
	return nil
}

// TestAll сам поднимает докер и запускает все тесты
func TestAll(t *testing.T) {
	t.Log("Starting docker-compose up...")

	err := runDockerComposeUp()
	if err != nil {
		t.Fatalf("Error running docker-compose: %v\n", err)
		return
	}

	defer func() {
		err = runDockerComposeDown()
		if runDockerComposeDown() != nil {
			t.Logf("Error stopping docker-compose: %v\n", err)
			return
		}
		t.Log("Docker compose down")
	}()

	t.Log("Docker compose up")
	t.Log("Waiting for services start")
	time.Sleep(time.Minute)
	t.Log("Running tests")
	t.Run("User System test", UserSystemUnderTest)
	t.Run("Auth middleware test", AuthMiddlewareUnderTest)
	t.Run("Expressions test", ExpressionsUnderTest)
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

	var createdExpressions []int
	createCases := []testCaseWithTokens{
		{
			name:     "create new expression",
			method:   "POST",
			url:      "http://localhost:8080/expressions",
			body:     "2+3",
			expected: http.StatusOK,
			auth:     true,
			token:    string(token),
		},
		{
			name:     "create bad expression",
			method:   "POST",
			url:      "http://localhost:8080/expressions",
			body:     "2+asdsad3",
			expected: http.StatusBadRequest,
			auth:     true,
			token:    string(token),
		},
		{
			name:     "create big expression",
			method:   "POST",
			url:      "http://localhost:8080/expressions",
			body:     "2+ (2* 5 + 1)",
			expected: http.StatusOK,
			auth:     true,
			token:    string(token),
		},
		{
			name:     "create empty expression",
			method:   "POST",
			url:      "http://localhost:8080/expressions",
			body:     "",
			expected: http.StatusBadRequest,
			auth:     true,
			token:    string(token),
		},
		{
			name:     "no auth",
			method:   "POST",
			url:      "http://localhost:8080/expressions",
			body:     "",
			expected: http.StatusUnauthorized,
			auth:     false,
			token:    "",
		},
	}

	for _, tc := range createCases {
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
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("failed to read body: %v", err)
				return
			}
			ibody, err := strconv.Atoi(string(body))
			if err != nil && ibody != 0 {
				createdExpressions = append(createdExpressions, ibody)
			}
			if res.StatusCode != tc.expected {
				t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, tc.expected)
				return
			}
		})
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/expressions", strings.NewReader(""))
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
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, 200)
		return
	}
	var response []entities.Expression
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Errorf("failed to unmarshal body: %v", err)
		return
	}
	if len(response) < len(createdExpressions) {
		t.Errorf("not enought results: %v", err)
		return
	}
	for _, id := range createdExpressions {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/expressions/%d", id), strings.NewReader(""))
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
	}
}
