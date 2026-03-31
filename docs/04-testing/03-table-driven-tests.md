# Table-Driven Tests

Table-driven testing is the standard Go pattern for testing multiple cases of the same function. Instead of writing a separate test function for each scenario, you define a list of cases as data and loop over them.

## The problem it solves

Imagine testing `GET /todos/:id` without this pattern:

```go
func TestGetTodoByIDFound(t *testing.T) { ... }
func TestGetTodoByIDNotFound(t *testing.T) { ... }
func TestGetTodoByIDInvalidID(t *testing.T) { ... }
```

Three functions, lots of repeated setup. And when you add a fourth case, you add a fourth function. The test logic is scattered.

## The table-driven approach

Instead, define a slice of anonymous structs — one per case — and loop:

```go
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
            // ... rest of assertions
        })
    }
}
```

### The anonymous struct

`tests` is a slice of an anonymous struct — a struct type defined inline without a name. You only need this type in one place, so there's no reason to name it. Each field describes one dimension of a test case:

- `name` — human-readable label, shows up in test output
- `path` — the input that varies between cases
- `wantStatus`, `wantTodo`, `wantError` — the expected outputs

The naming convention `want*` is idiomatic Go for "what I expect". You'll see it in virtually every Go test codebase.

### Optional fields with zero values

Notice that `wantTodo` is `*Todo` (a pointer), not `Todo`. This lets it be `nil` — meaning "don't check the body for this case". For error cases, `wantTodo` is `nil` and `wantError` is set instead.

For value types like `string`, the zero value is `""`. So `wantError: ""` is equivalent to "don't check the error". The assertion code checks `if tc.wantError != ""` before trying to unmarshal the error body.

This is a common Go pattern: use the zero value of a type as a sentinel meaning "not applicable".

### t.Run: named subtests

```go
t.Run(tc.name, func(t *testing.T) { ... })
```

`t.Run` creates a **subtest** with its own name. This does two things:

1. The output names each case individually — when one fails, you know exactly which one
2. Each subtest gets its own `t`, so `t.Fatalf` inside a subtest only stops that case, not the whole test function

The output looks like:

```bash
=== RUN   TestGetTodoByID/existing_id_returns_correct_todo
=== RUN   TestGetTodoByID/non-existent_id_returns_404
=== RUN   TestGetTodoByID/non-numeric_id_returns_400
```

Compare this to Jest's `describe` + `it`:

```typescript
describe('GET /todos/:id', () => {
    it('existing id returns correct todo', () => { ... })
    it('non-existent id returns 404', () => { ... })
})
```

The structure is the same. Go just expresses it as data (a slice of structs) rather than nested function calls.

### Adding a new case

To add a fourth case, you add one struct literal to the `tests` slice. No new functions, no new loops. The assertion logic runs automatically.

```go
{
    name:       "id zero returns 404",
    path:       "/todos/0",
    wantStatus: http.StatusNotFound,
    wantError:  "Todo not found",
},
```

This is why the pattern is idiomatic Go: it's the minimal amount of structure that scales.

## Asserting JSON responses

### Struct comparison for success cases

When the handler returns a single todo, we unmarshal the body into a `Todo` struct and compare directly:

```go
var got Todo
if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
    t.Fatalf("unmarshal failed: %v", err)
}
if got != *tc.wantTodo {
    t.Errorf("body: got %+v, want %+v", got, *tc.wantTodo)
}
```

Why not compare the raw JSON strings? Because JSON serialization doesn't guarantee key order. `{"id":1,"title":"..."}` and `{"title":"...","id":1}` are the same object but different strings. String comparison would make your tests fragile.

Unmarshalling into a struct and comparing structs is key-order safe.

`%+v` in the format string prints a struct with field names: `{ID:1 Title:Learn Go Completed:false}`. Much more readable than `%v` alone when a test fails.

### map[string]string for error cases

Error responses from `gin.H{"error": "..."}` serialise to `{"error":"..."}`. We unmarshal into a `map[string]string` and check the `"error"` key:

```go
var got map[string]string
if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
    t.Fatalf("unmarshal error body failed: %v", err)
}
if got["error"] != tc.wantError {
    t.Errorf("error: got %q, want %q", got["error"], tc.wantError)
}
```

`%q` prints a quoted string — `"Todo not found"` instead of `Todo not found`. Useful when comparing strings so whitespace or empty values are obvious.

## Mutating tests: reset inside the loop

For read-only handlers (`GET`), resetting once before the loop is enough — the cases don't affect each other's state.

For mutating handlers (`POST`, `PATCH`, `DELETE`), each case changes the `todos` slice. If you reset once before the loop, case 2 starts with whatever case 1 left behind. The fix: move `resetTodos()` and `setupRouter()` inside the subtest:

```go
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        resetTodos()           // fresh state for every case
        router := setupRouter()

        w := performRequest(router, ...)
        // ...
    })
}
```

Each subtest now runs in complete isolation. Creating a new router per subtest is cheap — Gin router setup is fast.

## What's next

With the pattern understood, the next lessons apply it to each handler and highlight anything specific to that handler's behaviour.
