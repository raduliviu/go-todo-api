# Project Setup & Your First Route

If you're a TypeScript developer looking to pick up Go, building a REST API is a great starting point — the concepts map closely to what you already know. In this series, we'll build a todo API from scratch using Go and [Gin](https://github.com/gin-gonic/gin), a lightweight HTTP framework similar to Express.

## Initializing the project

In Node, you'd run `npm init`. In Go, the equivalent is:

```bash
go mod init github.com/your-username/go-todo-api
```

This creates a `go.mod` file — Go's version of `package.json`. It tracks your module name, Go version, and dependencies. There's also a `go.sum` file (like `package-lock.json`) that gets generated automatically when you install packages.

## Installing Gin

In Node, you'd `npm install express`. In Go:

```bash
go get github.com/gin-gonic/gin
```

`go get` fetches the package and adds it to your `go.mod`. There's no `node_modules` folder — Go stores downloaded packages in a global cache (`$GOPATH/pkg/mod`), so they're shared across projects.

## Writing main.go

Every Go program starts with a `main` function in a `main` package. Create a `main.go` file:

```go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func handleGetTodos(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":        1,
		"title":     "Learn Go",
		"completed": false,
	})
}

func main() {
	server := gin.Default()

	server.GET("/todos", handleGetTodos)
	err := server.Run(":8080")
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
```

Let's break down what's different from TypeScript.

### package main

Every `.go` file starts with a package declaration. `package main` is special — it tells the compiler this is an executable program, not a library. Think of it like the entry point of your app.

### Imports

Go doesn't have `import express from 'express'`. Instead, imports are grouped in parentheses. Standard library packages (like `net/http`) are referenced by path, and third-party packages use their full module path.

### Structs instead of interfaces

In TypeScript, you'd define a Todo type with an interface. In Go, you use a **struct** — a typed collection of fields:

```typescript
// TypeScript
interface Todo {
  id: number;
  title: string;
  completed: boolean;
}
```

```go
// Go
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}
```

The backtick annotations (`` `json:"id"` ``) are called **struct tags**. They tell the JSON encoder what keys to use when serializing. Without them, your JSON would have uppercase keys like `{"ID": 1}` since Go exports fields by starting them with a capital letter.

### Handler functions

In Express, a route handler looks like:

```typescript
app.get('/todos', (req, res) => {
  res.json({ id: 1, title: 'Learn Go' });
});
```

In Gin, the handler takes a `*gin.Context` (a [pointer](../go-concepts/pointers.md) to a Context struct) that combines both request and response into one object. `c.JSON()` is the equivalent of `res.json()`.

### Error handling

Notice there's no `try/catch`. Go doesn't have exceptions. Instead, functions return errors as values, and you check them explicitly. `server.Run()` returns an error if the server fails to start — we check it and log a fatal message if something goes wrong.

## Running the server

```bash
go run .
```

This compiles and runs your code in one step (like `npx ts-node`). Test it with curl:

```bash
curl http://localhost:8080/todos
```

You should see your hardcoded todo in the response.

### Adding mock data

Returning a single hardcoded JSON object isn't very useful. Let's add an in-memory [slice](../go-concepts/slices-and-arrays.md) of todos. A **slice** in Go is the equivalent of a JavaScript array — a dynamically-sized, ordered collection.

```go
var todos = []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a web server", Completed: false},
	{ID: 3, Title: "Write unit tests", Completed: false},
}

func handleGetTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}
```

`var todos = []Todo{...}` declares a package-level variable — a slice of `Todo` structs. This is our in-memory "database" for now. The handler now returns the full list instead of a single object.

## What's next

We have a running API with a single GET route. In the next lesson, we'll add the ability to fetch a single todo by its ID, which introduces route parameters and type conversion in Go.
