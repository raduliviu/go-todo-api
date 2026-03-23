# Get Todo by ID

In the previous lesson we built a `GET /todos` route that returns all todos. Now we'll add `GET /todos/:id` to fetch a single todo — introducing route parameters and type conversion.

## Route parameters

In Express, you'd define a route parameter like this:

```typescript
app.get('/todos/:id', (req, res) => {
  const id = req.params.id; // string
});
```

In Gin, it's the same `:id` syntax. You extract it with `c.Param()`:

```go
server.GET("/todos/:id", getTodoByID)

func getTodoByID(c *gin.Context) {
	id := c.Param("id") // string
}
```

In both cases, the parameter comes back as a **string**. But our todo IDs are integers, so we need to convert.

## Type conversion: strconv.Atoi

In TypeScript, you'd write `parseInt(id)` or `Number(id)`. In Go, there's no implicit type coercion — you use the `strconv` (string conversion) package:

```typescript
// TypeScript
const id = parseInt(req.params.id);
if (isNaN(id)) { ... }
```

```go
// Go
id, err := strconv.Atoi(c.Param("id")) // Atoi = "ASCII to integer"
if err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
	return
}
```

A few things to notice here:

### Multiple return values

`strconv.Atoi` returns **two values**: the converted integer and an error. This is a core Go pattern — functions that can fail return `(result, error)` instead of throwing exceptions.

In TypeScript, `parseInt("abc")` silently returns `NaN`. In Go, you get an explicit error you must handle.

### The `:=` operator

`:=` declares and assigns a variable in one step. It's shorthand for:

```go
var id int
var err error
id, err = strconv.Atoi(c.Param("id"))
```

You'll use `:=` everywhere in Go. It infers the type from the right-hand side, similar to TypeScript's `const id = ...` with type inference.

### Early returns

Notice the `return` after sending the error response. Without it, the function would continue executing and try to look up the todo — Go won't stop for you. This is different from Express middleware where `next()` controls flow. In Gin handlers, you manage flow with explicit returns.

## Finding the todo

With a valid ID, we loop through the slice to find a match:

```go
for _, todo := range todos {
	if todo.ID == id {
		c.JSON(http.StatusOK, todo)
		return
	}
}
c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
```

There's no `.find()` method like in TypeScript:

```typescript
const todo = todos.find(t => t.id === id);
if (!todo) { res.status(404).json({ error: 'Todo not found' }); }
```

In Go, you write the loop. When the todo is found, we respond and `return` immediately. If the loop finishes without finding anything, we fall through to the 404 response.

## The full handler

```go
func getTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	for _, todo := range todos {
		if todo.ID == id {
			c.JSON(http.StatusOK, todo)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}
```

## Try it out

```bash
curl http://localhost:8080/todos/1
```

Try an invalid ID to see the error handling in action:

```bash
curl http://localhost:8080/todos/abc
```

## What's next

We can read todos — now let's create them. The next lesson introduces `POST` requests, JSON request body parsing, and how Gin binds JSON to structs.
