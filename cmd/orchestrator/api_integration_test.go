package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

/*
Интеграционные тесты для api
Используют собранный через docker сервис
Чтобы запустить все тесты с автоматическим запуском docker'а можно запустить тест TestAll
*/

type testCase struct {
	name     string
	method   string
	url      string
	body     string
	expected int
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
	t.Run("User System test", TestUserSystem)
	t.Run("Auth middleware test", TestAuthMiddleware)
}

// TestUserSystem интеграционный тест системы пользователей
func TestUserSystem(t *testing.T) {
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
			expected: http.StatusUnauthorized,
		},
		{
			name:     "login with incorrect password",
			method:   "POST",
			url:      "http://localhost:8080/login",
			body:     `{"name": "my_name", "password": "12345f6awfawf78"}`,
			expected: http.StatusUnauthorized,
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

func TestAuthMiddleware(t *testing.T) {
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
