# Context

`context.Context` is one of those things you see everywhere in Go but rarely get a proper explanation for. It's the first argument on almost every function that does I/O — database calls, HTTP requests, file reads — and it carries two things: a cancellation signal and request-scoped values.

## The problem it solves

In a web server, work happens on behalf of a request. If the client disconnects mid-flight, that work is wasted — but without some coordination mechanism, the server keeps going anyway, tying up resources (database connections, goroutines, memory) for a result nobody will ever read.

Context solves this by giving every piece of work a handle to check: "has the request that started me been cancelled?"

The TypeScript mental model: imagine every async function implicitly received an `AbortSignal` as its first argument. That's essentially what `ctx` is.

## What a context carries

**1. A deadline or cancellation signal**

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Any function receiving this ctx will be cancelled after 5 seconds
rows, err := db.QueryContext(ctx, "SELECT ...")
```

If the deadline fires, any I/O operation holding that context returns an error immediately rather than continuing to block.

**2. Request-scoped values**

```go
ctx = context.WithValue(ctx, "userID", 42)

// Somewhere deeper in the call stack:
userID := ctx.Value("userID").(int)
```

Used for things like auth tokens, trace IDs, and request metadata — values that need to flow through many layers without being explicitly threaded through every function signature. Use sparingly; it bypasses the type system.

## Where context comes from in Gin

Gin attaches a context to every HTTP request. You get it from the handler:

```go
func (h *Handler) getTodos(c *gin.Context) {
    todos, err := h.store.GetAll(c.Request.Context())
    // ...
}
```

`c.Request.Context()` returns the context tied to this specific HTTP request. It is automatically cancelled when:

- The client disconnects
- The response is fully sent
- A timeout middleware fires

## The cancellation chain

When you pass `c.Request.Context()` into a Bun call:

```go
s.db.NewSelect().Model(&todos).Scan(ctx)
```

Bun passes it to the Postgres driver, which passes it to the underlying TCP connection. If the context cancels at any point in that chain, the database query is abandoned and the connection is returned to the pool immediately.

```
HTTP request arrives
  → Gin creates a context
    → handler passes it to the store
      → store passes it to Bun
        → Bun passes it to pgdriver
          → pgdriver holds query open with this context

Client disconnects
  → context cancelled
    → pgdriver cancels the query
      → connection returned to pool
```

Without this, a slow query would keep running and hold a pool connection even after the client is long gone.

## context.Background()

When you don't have a request context — in `main.go`, in tests, in one-off scripts — you start with:

```go
ctx := context.Background()
```

This is a never-cancels, never-times-out root context. Think of it as "I have no real context here, give me a blank one." It's the root of every context tree in a Go program.

`context.TODO()` is also available and identical in behaviour — it's a signal to other developers (and linters) that you intend to wire up a real context later but haven't yet.

## The convention

Context is always the first argument, always named `ctx`:

```go
func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error)
func (s *TodoStore) GetByID(ctx context.Context, id int64) (*Todo, error)
```

This is not enforced by the language — it's a Go community convention followed so universally that deviating from it reads as a mistake.

## What context is not

Context is not a way to pass optional parameters or configuration into functions. If you find yourself using `context.WithValue` for anything other than request-scoped metadata (trace IDs, auth tokens), it's a sign the value should be an explicit function argument instead.
