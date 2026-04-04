# ShouldBindJSON and Consistent Error Responses

## The problem with `BindJSON`

When we first wrote `createTodo`, the error handling looked like this:

```go
if err := c.BindJSON(&newTodo); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

This looks like you're in control of the error response. You're not.

`c.BindJSON` internally calls `c.AbortWithStatus(400)` before it returns the error to you. By the time your `c.JSON(...)` line runs, the status code has already been written to the response and your call silently does nothing. You're writing to a response that's already been sent.

The TypeScript mental model: `BindJSON` is like an Express middleware that catches the error and sends a response before your handler runs. You think you're in control, but the framework has already acted.

## `ShouldBindJSON`: the same thing, but you're in charge

`c.ShouldBindJSON` is identical in every other way — same JSON parsing, same validation — but it returns the error without touching the response. You decide what to send back.

```go
if err := c.ShouldBindJSON(&newTodo); err != nil {
    c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
    return
}
```

Now the flow is clear: if parsing or validation fails, you write the error response. If it succeeds, you continue. No surprises.

**Rule of thumb:** use `ShouldBind*` variants in handlers. Reserve `Bind*` (which calls `MustBindWith`) only if you have a specific reason to let Gin take over on failure, which is rarely the case.

## The `errorResponse` helper

Before this change, error responses were written inconsistently:

```go
// Some places used friendly messages
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})

// Other places leaked raw Go error strings
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
```

Both produce `{"error": "..."}`, but there's nothing enforcing that shape. A future handler could easily write `gin.H{"message": "..."}` or `gin.H{"err": "..."}` and break API clients silently.

The fix is a one-liner helper:

```go
func errorResponse(msg string) gin.H {
    return gin.H{"error": msg}
}
```

Now every error response in the codebase goes through this function. The shape is enforced at the call site — if you use `errorResponse()`, you get `{"error": "..."}`. There's no way to accidentally write a different key.

This is a Go pattern worth internalizing: even a trivial helper is worth extracting when it enforces a contract. You're not adding abstraction for its own sake — you're making the wrong thing harder to write.

## What's next

With consistent error responses in place, the next lesson introduces request structs — a clean separation between what a `Todo` is in the domain and what a client is allowed to send.
