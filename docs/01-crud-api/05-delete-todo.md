# Delete Todo

The last CRUD operation. We'll add `DELETE /todos/:id` — introducing `slices.Delete` and the `204 No Content` response pattern.

## The handler

By now, the first half of this handler should look familiar — parse the ID, find the index, handle errors:

```go
func deleteTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	todos = slices.Delete(todos, todoIndex, todoIndex+1)
	c.Status(http.StatusNoContent)
}
```

The new parts are the last two lines.

## Deleting from a slice

As covered in the [slices concept doc](../go-concepts/slices-and-arrays.md#delete), Go doesn't have a `.splice()` equivalent in the language itself. The standard library's `slices.Delete` handles it:

```typescript
// TypeScript
todos.splice(index, 1);
```

```go
// Go
todos = slices.Delete(todos, todoIndex, todoIndex+1) // (slice, start, end)
```

Remember that `slices.Delete` returns a new slice — you must reassign it. The range is half-open: `todoIndex` is included, `todoIndex+1` is excluded, so this removes exactly one element.

## 204 No Content

After deleting, there's nothing meaningful to return. The convention is `204 No Content` — it tells the client "the action succeeded, but there's no response body."

```go
c.Status(http.StatusNoContent)
```

Note we use `c.Status()` instead of `c.JSON()` since there's no body to send. In Express, this would be:

```typescript
res.sendStatus(204);
```

## Try it out

```bash
curl -X DELETE http://localhost:8080/todos/1
```

Verify it's gone:

```bash
curl http://localhost:8080/todos
```

## What's next

We now have a fully working CRUD API — all five routes are in place. In the next lesson, we'll refactor the code by splitting our single `main.go` into separate files, following Go project layout conventions.
