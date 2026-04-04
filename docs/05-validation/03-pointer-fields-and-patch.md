# Pointer Fields and PATCH Semantics

## The silent data loss bug

Before this chapter, `updateTodoByID` bound the request body directly into the stored todo:

```go
updatedTodo := &todos[todoIndex]
c.ShouldBindJSON(updatedTodo)
```

This looks reasonable, but it has a subtle bug. When Go's JSON decoder processes a struct, any field absent from the JSON is set to its **zero value** — `""` for strings, `false` for bools. So if a client sends:

```json
{ "title": "New title" }
```

The decoder sets `Title` to `"New title"` — correct. But it also sets `Completed` to `false` — silently overwriting whatever was stored. A completed todo becomes incomplete because the client didn't include the `completed` field.

This is the difference between a true PATCH (update only what was sent) and what we had (overwrite everything with what was sent, defaulting absent fields to zero).

## Pointer fields: the fix

The root cause is that `false` and "not provided" are indistinguishable for a plain `bool`. Go needs a way to represent "this field was absent from the request".

Pointers give you that. A pointer's zero value is `nil` — distinct from any actual value.

```go
type UpdateTodoRequest struct {
    Title     *string `json:"title"     binding:"omitempty,min=1,max=255"`
    Completed *bool   `json:"completed"`
}
```

Now:

- Client sends `{"title": "New title"}` → `req.Title` is a `*string` pointing to `"New title"`, `req.Completed` is `nil`
- Client sends `{"completed": true}` → `req.Title` is `nil`, `req.Completed` is a `*bool` pointing to `true`
- Client sends `{"completed": false}` → `req.Title` is `nil`, `req.Completed` is a `*bool` pointing to `false`
- Client sends `{}` → both are `nil`

In TypeScript terms: `*string` is `string | undefined`. `nil` means the key was absent from the JSON.

## `omitempty` on Title

```go
binding:"omitempty,min=1,max=255"
```

`omitempty` means "skip all other validation rules if this field was not provided". Without it, `min=1` would fire even when `Title` is `nil` — rejecting requests that don't include a title at all, which is valid for a partial update.

The validation chain with `omitempty` reads as: "if title was provided, it must be between 1 and 255 characters; if it wasn't provided, that's fine."

## The nil-check pattern in the handler

```go
var req UpdateTodoRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
    return
}

if req.Title != nil {
    todos[todoIndex].Title = *req.Title
}
if req.Completed != nil {
    todos[todoIndex].Completed = *req.Completed
}

c.JSON(http.StatusOK, todos[todoIndex])
```

Each field is applied only if the client sent it. `*req.Title` dereferences the pointer to get the underlying string value — only safe to do after confirming it's not `nil`.

This is explicit and readable: anyone reading the handler can see exactly which fields are optional and how absent fields are handled.

## Why not just use `map[string]interface{}`?

You could decode the request into a map and check which keys are present. Some APIs do this. But you'd lose type safety, validation tags, and the ability to name fields clearly. Pointer fields give you all of that while keeping the struct approach.

## What the tests verify

The partial update test cases make this concrete:

```go
// Only title sent — completed must stay false (its stored value)
{"title": "Learn Go deeply"} → {ID:1, Title:"Learn Go deeply", Completed:false}

// Only completed sent — title must stay "Learn Go" (its stored value)
{"completed": true} → {ID:1, Title:"Learn Go", Completed:true}
```

Before pointer fields, the first case would have returned `Completed: false` — which happens to be correct, but only because the seed data has `Completed: false`. If the stored todo had `Completed: true`, that test would have caught the bug.
