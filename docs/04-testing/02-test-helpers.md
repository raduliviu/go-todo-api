# Test Helpers: resetTodos and performRequest

Before writing any actual tests, we need two helper functions. Both exist to eliminate repetition — you'll call them in every single test.

## resetTodos: taming global state

Our `todos` variable is a global slice. It's the in-memory "database". The problem with global state in tests is that one test can mutate it and leave it in a different shape for the next test. That makes tests **order-dependent** — a nightmare to debug.

The fix is simple: before each test, reset it back to a known state.

```go
func resetTodos() {
    todos = []Todo{
        {ID: 1, Title: "Learn Go", Completed: false},
        {ID: 2, Title: "Build a web server", Completed: false},
        {ID: 3, Title: "Write unit tests", Completed: false},
    }
}
```

This mirrors the seed data in `models.go` exactly. Every test starts from the same world: 3 todos, IDs 1–3. No surprises.

The TypeScript equivalent mental model is resetting a module-level variable before each test — except in Go you don't need any mocking library to do it, you just assign directly.

## performRequest: fake HTTP without a server

In TypeScript/Express testing, you might use `supertest` to fire requests at your app without starting a real server. Go has this built in via the `net/http/httptest` package.

Here's the helper:

```go
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
```

There are three concepts here worth understanding individually.

### bytes.Buffer: an in-memory byte stream

HTTP request bodies are **streams** — the server reads bytes from them as it processes the request. In a real HTTP call, that stream comes from a network socket. In a test, we need to fake it.

`bytes.NewBuffer(body)` wraps a `[]byte` slice in a `Buffer` — an in-memory object that implements the `io.Reader` interface. That's the interface `http.Request` expects for its body. It reads from memory instead of a network connection, but the handler doesn't know the difference.

In TypeScript terms: you're creating a `ReadableStream` from a string, except Go's version is lower-level and more explicit.

### httptest.NewRequest: a fake HTTP request

```go
req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
```

This builds an `*http.Request` struct — the same type Gin receives in production. It has a method, a URL path, headers, and a body reader. No network socket, no DNS, no TCP connection. Just a struct in memory.

Note: we use `httptest.NewRequest`, not `http.NewRequest`. The difference is error handling: `http.NewRequest` returns `(req, error)` and you'd have to handle the error. `httptest.NewRequest` panics on an invalid URL instead. That's the right behaviour in a test — a bad URL is always a programmer mistake, not a runtime condition.

### httptest.ResponseRecorder: a fake HTTP response writer

```go
w := httptest.NewRecorder()
```

When Gin calls `c.JSON(200, ...)`, it's writing to an `http.ResponseWriter`. In production, that's a real network socket — it sends bytes over TCP to the client. In a test, we need to capture those bytes instead of sending them anywhere.

`httptest.ResponseRecorder` implements `http.ResponseWriter`, but instead of sending bytes over the network, it stores them in memory. After the handler runs, you can read:

- `w.Code` — the HTTP status code
- `w.Body` — a `*bytes.Buffer` containing the response body
- `w.Header()` — the response headers

Think of it as intercepting the response before it goes anywhere, so you can inspect it.

### Putting it together

```go
router.ServeHTTP(w, req)
```

This is the line that runs the full Gin request cycle — routing, middleware, handler — using `req` as the input and `w` as the output. After this line returns, `w.Code` and `w.Body` contain exactly what the handler wrote.

The return value is `*httptest.ResponseRecorder`, so callers can inspect `w.Code` and `w.Body` directly.

## What's next

With these helpers in place, the next lesson covers how to structure multiple test cases cleanly using Go's table-driven test pattern.
