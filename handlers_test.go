package main

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raduliviu/go-todo-api/store"
)

type inMemoryStore struct {
	todos  map[int64]*store.Todo
	nextID int64
}

func newInMemoryStore() *inMemoryStore {
	s := &inMemoryStore{
		todos:  make(map[int64]*store.Todo),
		nextID: 4,
	}
	s.todos[1] = &store.Todo{ID: 1, Title: "Learn Go", Completed: false}
	s.todos[2] = &store.Todo{ID: 2, Title: "Build a web server", Completed: false}
	s.todos[3] = &store.Todo{ID: 3, Title: "Write unit tests", Completed: false}
	return s
}

func (s *inMemoryStore) GetAll(ctx context.Context) ([]store.Todo, error) {
	result := make([]store.Todo, 0, len(s.todos))
	for _, t := range s.todos {
		result = append(result, *t)
	}
	slices.SortFunc(result, func(a, b store.Todo) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return result, nil
}

func (s *inMemoryStore) GetByID(ctx context.Context, id int64) (*store.Todo, error) {
	t, ok := s.todos[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copy := *t
	return &copy, nil
}

func (s *inMemoryStore) Create(ctx context.Context, todo *store.Todo) error {
	todo.ID = s.nextID
	s.nextID++
	copy := *todo
	s.todos[todo.ID] = &copy
	return nil
}

func (s *inMemoryStore) Update(ctx context.Context, todo *store.Todo) error {
	if _, ok := s.todos[todo.ID]; !ok {
		return sql.ErrNoRows
	}
	copy := *todo
	s.todos[todo.ID] = &copy
	return nil
}

func (s *inMemoryStore) Delete(ctx context.Context, id int64) error {
	if _, ok := s.todos[id]; !ok {
		return sql.ErrNoRows
	}
	delete(s.todos, id)
	return nil
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
	s := newInMemoryStore()
	h := NewHandler(s)
	router := setupRouter(h)

	w := performRequest(router, http.MethodGet, "/todos", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var got []store.Todo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if len(got) != 3 {
		t.Errorf("count: got %d, want 3", len(got))
	}
}

func TestGetTodoByID(t *testing.T) {
	s := newInMemoryStore()
	h := NewHandler(s)
	router := setupRouter(h)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantTodo   *store.Todo
		wantError  string
	}{
		{
			name:       "existing id returns correct todo",
			path:       "/todos/1",
			wantStatus: http.StatusOK,
			wantTodo:   &store.Todo{ID: 1, Title: "Learn Go", Completed: false},
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
				var got store.Todo
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
		wantTodo   *store.Todo
	}{
		{
			name:       "valid body creates todo and returns 201",
			body:       []byte(`{"title": "Deploy to prod", "completed": false}`),
			wantStatus: http.StatusCreated,
			wantTodo:   &store.Todo{ID: 4, Title: "Deploy to prod", Completed: false},
		},
		{
			name:       "malformed JSON returns 400",
			body:       []byte(`{bad json}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing title returns 400",
			body:       []byte(`{"completed": false}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty title returns 400",
			body:       []byte(`{"title": "", "completed": false}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "title too long returns 400",
			body:       []byte(`{"title": "` + string(bytes.Repeat([]byte("a"), 256)) + `", "completed": false}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "omitted completed defaults to false",
			body:       []byte(`{"title":"New task"}`),
			wantStatus: http.StatusCreated,
			wantTodo:   &store.Todo{ID: 4, Title: "New task", Completed: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newInMemoryStore()
			h := NewHandler(s)
			router := setupRouter(h)

			w := performRequest(router, http.MethodPost, "/todos", tc.body)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantTodo != nil {
				var got store.Todo
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
		wantTodo   *store.Todo
		wantError  string
	}{
		{
			name:       "valid update patches todo",
			path:       "/todos/1",
			body:       []byte(`{"title": "Learn Go well", "completed": true}`),
			wantStatus: http.StatusOK,
			wantTodo:   &store.Todo{ID: 1, Title: "Learn Go well", Completed: true},
		},
		{
			name:       "client cannot override id",
			path:       "/todos/1",
			body:       []byte(`{"id": 999, "title": "sneaky", "completed": false}`),
			wantStatus: http.StatusOK,
			wantTodo:   &store.Todo{ID: 1, Title: "sneaky", Completed: false},
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
		{
			name:       "partial update only changes title",
			path:       "/todos/1",
			body:       []byte(`{"title": "Learn Go deeply"}`),
			wantStatus: http.StatusOK,
			wantTodo:   &store.Todo{ID: 1, Title: "Learn Go deeply", Completed: false},
		},
		{
			name:       "partial update only changes completed",
			path:       "/todos/1",
			body:       []byte(`{"completed": true}`),
			wantStatus: http.StatusOK,
			wantTodo:   &store.Todo{ID: 1, Title: "Learn Go", Completed: true},
		},
		{
			name:       "empty title in update returns 400",
			path:       "/todos/1",
			body:       []byte(`{"title": ""}`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newInMemoryStore()
			h := NewHandler(s)
			router := setupRouter(h)

			w := performRequest(router, http.MethodPatch, tc.path, tc.body)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}

			if tc.wantTodo != nil {
				var got store.Todo
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
			s := newInMemoryStore()
			h := NewHandler(s)
			router := setupRouter(h)

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
