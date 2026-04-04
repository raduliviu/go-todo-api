# Request Structs and Binding Validation

## The problem with binding into the model

Before this chapter, `createTodo` bound the request body directly into a `Todo`:

```go
var newTodo Todo
c.ShouldBindJSON(&newTodo)
```

The `Todo` struct has an `ID` field. That means a client could send `{"id": 999, "title": "x"}` and the handler would read it — the only reason it didn't matter is that the next line overwrote the ID. That's accidental protection, not intentional design.

More importantly, there was no way to enforce that `title` was present and non-empty. Any request slipped through.

## Two different jobs, two different structs

Think of it this way:

- `models.go` describes what a `Todo` **is** in your domain — the shape of stored data
- `requests.go` describes what a client **is allowed to send** — the shape of incoming API requests

In TypeScript, you'd call these a model and a DTO (Data Transfer Object). In Go, it's the same idea: separate types for separate concerns.

```go
// models.go — what a Todo is
type Todo struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

// requests.go — what the client may send when creating a Todo
type CreateTodoRequest struct {
    Title     string `json:"title"     binding:"required,min=1,max=255"`
    Completed bool   `json:"completed"`
}
```

No `ID` field in `CreateTodoRequest`. The client literally cannot send an ID — there's no field to bind it into. Protection by omission is stronger than protection by overwriting.

## Binding tags

The `binding:` struct tag is read by Gin's validator (which uses go-playground/validator under the hood — already in your dependencies via Gin). When `ShouldBindJSON` runs, it validates the struct against these tags automatically.

```go
binding:"required,min=1,max=255"
```

- `required` — the field must be present in the JSON *and* not be the zero value. For a string, the zero value is `""`, so both a missing `title` key and `{"title": ""}` are rejected.
- `min=1` — minimum length of 1. Alongside `required`, this makes the intent explicit rather than relying on `required`'s zero-value behavior.
- `max=255` — maximum length of 255. Prevents a client from sending a 100KB title and having it land in your storage layer unchecked.

**Why no `binding:` tag on `Completed`?**

In go-playground/validator, `required` on a `bool` means the field must be non-zero — and the zero value of `bool` is `false`. So `required` on `Completed` would reject `{"completed": false}`, which is a perfectly valid request. Leave bools without `required` unless you specifically need to enforce `true`.

## The handler after the change

```go
func createTodo(c *gin.Context) {
    var req CreateTodoRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
        return
    }

    newTodo := Todo{
        ID:        len(todos) + 1,
        Title:     req.Title,
        Completed: req.Completed,
    }

    todos = append(todos, newTodo)
    c.JSON(http.StatusCreated, newTodo)
}
```

The explicit struct literal `Todo{ID: ..., Title: req.Title, Completed: req.Completed}` is intentional. It documents exactly what the client controls — you're copying specific fields from the request into the model. There's no way a future field added to `CreateTodoRequest` can accidentally leak into the stored `Todo` without an explicit assignment here.

## What's next

The create handler is now clean. The update handler has a harder problem: PATCH requests are partial by design — the client sends only the fields they want to change. That requires a different technique, covered in the next lesson.
