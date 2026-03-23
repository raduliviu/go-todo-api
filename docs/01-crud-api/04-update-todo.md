# Update Todo

We can create and read todos. Now let's add `PATCH /todos/:id` to update an existing one. This lesson introduces `slices.IndexFunc` and a practical use of pointers for modifying slice elements in place.

## Why PATCH and not PUT?

Quick refresher: `PUT` replaces the entire resource, `PATCH` partially updates it. Since we want to allow updating just the title or just the completed status without sending the full object, `PATCH` is the right choice.

## Finding a todo by index

In lesson 02, we used a `for` loop with `range` to find a todo. There's a cleaner way using the standard library's `slices` package:

```typescript
// TypeScript
const index = todos.findIndex(t => t.id === id);
```

```go
// Go
import "slices"

todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
	return todo.ID == id
})
```

`slices.IndexFunc` takes a slice and a function that returns `bool`. It returns the index of the first match, or `-1` if nothing matches — just like JavaScript's `findIndex`.

The function argument `func(todo Todo) bool` is an **anonymous function** (a closure). Same concept as an arrow function in TypeScript, different syntax:

```typescript
// TypeScript arrow function
(t) => t.id === id
```

```go
// Go anonymous function
func(todo Todo) bool {
	return todo.ID == id
}
```

Go requires explicit types, the `func` keyword, and braces — there's no shorthand.

## Updating in place with pointers

Here's where it gets interesting. We need to modify the todo *in the slice*, not a copy of it. Remember the [range copies gotcha](../go-concepts/slices-and-arrays.md#one-gotcha-range-copies)?

```go
updatedTodo := &todos[todoIndex]
```

`&todos[todoIndex]` gives us a [pointer](../go-concepts/pointers.md) to the actual element in the slice. Any changes to `updatedTodo` now modify the original directly.

Then we bind the request body into it:

```go
if err := c.BindJSON(updatedTodo); err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	return
}
```

Notice we don't need `&` here — `updatedTodo` is already a pointer. `BindJSON` overwrites the fields with whatever the client sent.

## Preserving the ID

There's a subtle issue: the request body could include an `id` field, which would overwrite the todo's ID. We don't want the client to be able to change IDs:

```go
updatedTodo.ID = id
```

We reset the ID to the original value after binding. This is a simple guard — in a real API, you'd typically use separate request/response structs to avoid this entirely.

## The full handler

```go
func updateTodoByID(c *gin.Context) {
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

	updatedTodo := &todos[todoIndex]
	if err := c.BindJSON(updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedTodo.ID = id
	c.JSON(http.StatusOK, updatedTodo)
}
```

## Try it out

```bash
curl -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "completed": true}'
```

Then verify the change:

```bash
curl http://localhost:8080/todos/1
```

## What's next

One CRUD operation left: deleting a todo. The next lesson introduces `slices.Delete` and the `204 No Content` response pattern.
