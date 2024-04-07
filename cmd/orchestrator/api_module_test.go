package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type testCase struct {
	name     string
	method   string
	url      string
	body     string
	expected int
}

// TestUserSystem модульный тест системы регистрации
func TestUserSystem(t *testing.T) {
	// Сначала нужно запустить бд и orchestrator

	http.Post("http://localhost:8080/register", "application/json", strings.NewReader(`{"name": "my_name", "password": "12345678"}`))
	testCases := []testCase{
		{
			name:     "register user again (name already taken)",
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
			}
		})
	}
}
