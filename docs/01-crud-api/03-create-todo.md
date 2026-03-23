# Create Todo

We can list todos and fetch one by ID. Now let's add `POST /todos` to create new ones — introducing JSON request body parsing and how Gin binds incoming data to structs.

## Registering the route

In Express:

```typescript
app.post('/todos', (req, res) => { ... });
```

In Gin:

```go
server.POST("/todos", postTodo)
```

Same pattern as GET, just a different HTTP method.

## Parsing the request body

In Express with the JSON middleware, `req.body` is already a parsed object:

```typescript
app.post('/todos', (req, res) => {
  const { title, completed } = req.body;
});
```

In Go, you need to explicitly bind the JSON body to a struct. This is where [pointers](../go-concepts/pointers.md) show up again:

```go
func postTodo(c *gin.Context) {
	var newTodo Todo

	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}
```

Let's break this down.

### var newTodo Todo

This declares an empty `Todo` struct. In Go, variables are initialized to their **zero values** — `0` for ints, `""` for strings, `false` for bools. So `newTodo` starts as `{ID: 0, Title: "", Completed: false}`.

### c.BindJSON(&newTodo)

`BindJSON` reads the request body and populates the struct fields. We pass `&newTodo` (a pointer) because the function needs to *write into* our variable. Without the `&`, Gin would receive a copy and our `newTodo` would stay empty.

The struct tags we defined earlier tell Gin how to match JSON keys to struct fields:

```go
type Todo struct {
	ID        int    `json:"id"`        // matches "id" in JSON
	Title     string `json:"title"`     // matches "title" in JSON
	Completed bool   `json:"completed"` // matches "completed" in JSON
}
```

### Inline error handling

Notice the `if err := ...; err != nil` pattern. This is an **inline declaration** — it declares `err`, calls the function, and checks the result in one statement. It's equivalent to:

```go
err := c.BindJSON(&newTodo)
if err != nil {
	// ...
}
```

The inline version is idiomatic Go. It keeps `err` scoped to the `if` block, which is cleaner when you have multiple error checks in a row.

### Appending to the slice

```go
todos = append(todos, newTodo)
```

`append` adds the new todo to our [slice](../go-concepts/slices-and-arrays.md) and returns the updated slice. We respond with `http.StatusCreated` (201) and the new todo.

## Auto-incrementing the ID

There's a problem — the client is responsible for sending the ID. Let's have the server assign it instead:

```go
newTodo.ID = len(todos) + 1

todos = append(todos, newTodo)
c.JSON(http.StatusCreated, newTodo)
```

`len(todos)` gives us the current count, so `+ 1` makes the new ID one higher than the last. This is a simple approach for in-memory storage — a real database would handle ID generation for you.

## Try it out

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Deploy to AWS", "completed": false}'
```

Then verify it was added:

```bash
curl http://localhost:8080/todos
```

## What's next

We can create and read todos. Next up: updating an existing todo with `PATCH`.
