# Refactoring: Splitting main.go

Our CRUD API works, but everything lives in a single `main.go` file — struct definitions, mock data, five handler functions, and the server setup. That's getting unwieldy. Let's do an intermediate refactor by splitting it into separate files.

This is not the final structure. As the project grows (database layer, middleware, config), we'll revisit and adopt the `cmd/internal` layout described in the [project structure concept doc](../go-concepts/project-structure.md). For now, a flat split by responsibility is enough.

## The split

We'll go from one file to three:

| File | Responsibility |
|------|----------------|
| `main.go` | Server setup and route registration |
| `models.go` | The `Todo` struct and mock data |
| `handlers.go` | All HTTP handler functions |

In TypeScript, this would be like going from a single `index.ts` to `index.ts`, `types.ts`, and `routes.ts`.

## models.go

Extract the struct and the mock data:

```go
package main

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos = []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a web server", Completed: false},
	{ID: 3, Title: "Write unit tests", Completed: false},
}
```

## handlers.go

Move all the handler functions here:

```go
package main

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func getTodoByID(c *gin.Context) { ... }
func createTodo(c *gin.Context) { ... }
func updateTodoByID(c *gin.Context) { ... }
func deleteTodoByID(c *gin.Context) { ... }
```

## main.go

What's left is clean and focused — just server setup:

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	server.GET("/todos", getTodos)
	server.GET("/todos/:id", getTodoByID)
	server.POST("/todos", createTodo)
	server.PATCH("/todos/:id", updateTodoByID)
	server.DELETE("/todos/:id", deleteTodoByID)

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
```

## Why this works without imports

If you're coming from TypeScript, you'd expect to need import statements between these files. In Go, **all files in the same package can see each other's exported and unexported identifiers**. Since all three files declare `package main`, `handlers.go` can reference `Todo` and `todos` from `models.go` without any imports.

This is fundamentally different from TypeScript, where every file is its own module and you must explicitly `import` from other files.

## What didn't change

The API itself is identical — same routes, same behavior, same responses. This is purely a code organization change. You can verify by running `go run .` (which compiles all `.go` files in the directory) and testing your endpoints.

## What's next

This flat layout works for now, but it won't scale. When we add the database layer later in the project, we'll migrate to a `cmd/internal` package structure. For now, on to the next chapter: [Dockerizing the API](../02-dockerizing/).
