package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func resetTodos() {
	todos = []Todo{
		{ID: 1, Title: "Learn Go", Completed: false},
		{ID: 2, Title: "Build a web server", Completed: false},
		{ID: 3, Title: "Write unit tests", Completed: false},
	}
}

func performRequest(router *gin.Engine, method, path string, body []byte) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestGetTodos(t *testing.T) {
	resetTodos()
	router := setupRouter()

	w := performRequest(router, http.MethodGet, "/todos", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var got []Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if len(got) != 3 {
		t.Errorf("count: got %d, want 3", len(got))
	}
}

func TestGetTodoByID(t *testing.T) {
	resetTodos()
	router := setupRouter()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantTodo   *Todo
		wantError  string
	}{
		{
			name:       "existing id returns correct todo",
			path:       "/todos/1",
			wantStatus: http.StatusOK,
			wantTodo:   &Todo{ID: 1, Title: "Learn Go", Completed: false},
		},
		{
			name:       "non-existent id returns 404",
			path:       "/todos/999",
			wantStatus: http.StatusNotFound,
			wantError:  "Todo not found",
		},
		{
			name:       "non-numeric id returns 400",
			path:       "/todos/abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := performRequest(router, http.MethodGet, tc.path, nil)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantTodo != nil {
				var got Todo
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal failed: %v", err)
				}
				if got != *tc.wantTodo {
					t.Errorf("body: got %+v, want %+v", got, *tc.wantTodo)
				}
			}

			if tc.wantError != "" {
				var got map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal error body failed: %v", err)
				}
				if got["error"] != tc.wantError {
					t.Errorf("error: got %q, want %q", got["error"], tc.wantError)
				}
			}
		})
	}
}

func TestCreateTodo(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		wantStatus int
		wantTodo   *Todo
	}{
		{
			name:       "valid body creates todo and returns 201",
			body:       []byte(`{"title": "Deploy to prod", "completed": false}`),
			wantStatus: http.StatusCreated,
			wantTodo:   &Todo{ID: 4, Title: "Deploy to prod", Completed: false},
		},
		{
			name:       "malformed JSON returns 400",
			body:       []byte(`{bad json}`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetTodos()
			router := setupRouter()

			w := performRequest(router, http.MethodPost, "/todos", tc.body)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantTodo != nil {
				var got Todo
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal failed: %v", err)
				}
				if got != *tc.wantTodo {
					t.Errorf("body: got %+v, want %+v", got, *tc.wantTodo)
				}
			}
		})
	}
}

func TestUpdateTodoByID(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       []byte
		wantStatus int
		wantTodo   *Todo
		wantError  string
	}{
		{
			name:       "valid update patches todo",
			path:       "/todos/1",
			body:       []byte(`{"title": "Learn Go well", "completed": true}`),
			wantStatus: http.StatusOK,
			wantTodo:   &Todo{ID: 1, Title: "Learn Go well", Completed: true},
		},
		{
			name:       "client cannot override id",
			path:       "/todos/1",
			body:       []byte(`{"id": 999, "title": "sneaky", "completed": false}`),
			wantStatus: http.StatusOK,
			wantTodo:   &Todo{ID: 1, Title: "sneaky", Completed: false},
		},
		{
			name:       "non-existent id returns 404",
			path:       "/todos/999",
			body:       []byte(`{"title": "x", "completed": false}`),
			wantStatus: http.StatusNotFound,
			wantError:  "Todo not found",
		},
		{
			name:       "non-numeric id returns 400",
			path:       "/todos/abc",
			body:       []byte(`{"title": "x", "completed": false}`),
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid ID",
		},
		{
			name:       "malformed JSON returns 400",
			path:       "/todos/1",
			body:       []byte(`{bad}`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetTodos()
			router := setupRouter()

			w := performRequest(router, http.MethodPatch, tc.path, tc.body)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantTodo != nil {
				var got Todo
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal failed: %v", err)
				}
				if got != *tc.wantTodo {
					t.Errorf("body: got %+v, want %+v", got, *tc.wantTodo)
				}
			}

			if tc.wantError != "" {
				var got map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal error body failed: %v", err)
				}
				if got["error"] != tc.wantError {
					t.Errorf("error: got %q, want %q", got["error"], tc.wantError)
				}
			}
		})
	}
}

func TestDeleteTodoByID(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "existing id deletes todo and returns 204",
			path:       "/todos/2",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "non-existent id returns 404",
			path:       "/todos/999",
			wantStatus: http.StatusNotFound,
			wantError:  "Todo not found",
		},
		{
			name:       "non-numeric id returns 400",
			path:       "/todos/abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetTodos()
			router := setupRouter()

			w := performRequest(router, http.MethodDelete, tc.path, nil)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantStatus == http.StatusNoContent && w.Body.Len() != 0 {
				t.Errorf("expected empty body for 204, got: %s", w.Body.String())
			}

			if tc.wantError != "" {
				var got map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal error body failed: %v", err)
				}
				if got["error"] != tc.wantError {
					t.Errorf("error: got %q, want %q", got["error"], tc.wantError)
				}
			}
		})
	}
}
