# Postgres: Intro and the Store Pattern

Up to this point, todos live in a global slice. That works for learning, but it has problems you can't paper over: data disappears on restart, there's no concurrency safety, and there's no clean boundary between HTTP logic and data logic. This chapter replaces it with a real Postgres database.

The change introduces three ideas at once:

1. **The store/repository pattern** — a dedicated layer responsible for all database access
2. **Dependency injection via a Handler struct** — handlers receive their dependencies rather than reaching for globals
3. **Interfaces for testability** — the store is defined as an interface so tests can swap in a fast in-memory implementation

## The store pattern

In the old code, handlers directly manipulate the `todos` slice. The handler both owns the data and knows how to serve HTTP. That's two responsibilities in one place.

The store pattern splits them:

- **Handler**: knows HTTP — parsing requests, writing responses, status codes
- **Store**: knows data — queries, inserts, updates, deletes

A handler never touches a database directly. It calls the store, gets data back, and decides what HTTP response to send. The store never touches HTTP. Neither layer needs to know how the other works.

The TypeScript equivalent is a service/repository layer:

```typescript
// TypeScript mental model
class TodoService {
  async getAll(): Promise<Todo[]> { ... }
  async getById(id: number): Promise<Todo | null> { ... }
}

class TodoController {
  constructor(private service: TodoService) {}
  async getTodos(req, res) {
    const todos = await this.service.getAll()
    res.json(todos)
  }
}
```

In Go:

```go
// store/todo.go — data layer
func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) { ... }

// handlers.go — HTTP layer
func (h *Handler) getTodos(c *gin.Context) {
    todos, err := h.store.GetAll(c.Request.Context())
    ...
}
```

## The Handler struct

Previously, handler functions were standalone functions that accessed the global `todos` slice directly. With the store pattern, they need access to a store — and the way to give a function access to something in Go is to make it a method on a struct.

```go
type Handler struct {
    store TodoStorer
}

func NewHandler(s TodoStorer) *Handler {
    return &Handler{store: s}
}
```

`NewHandler` is a constructor — a plain function that returns a populated struct. Go has no `new` keyword for this; the convention is a function named `New<Type>`.

All five handler functions become methods on `*Handler`:

```go
// Before
func getTodos(c *gin.Context) { ... }

// After
func (h *Handler) getTodos(c *gin.Context) { ... }
```

The `(h *Handler)` part is the receiver — it gives the function access to `h.store`. Think of it as `this` in TypeScript methods.

## What's next

The next two lessons cover the interface that makes this testable and the in-memory implementation used in tests.
