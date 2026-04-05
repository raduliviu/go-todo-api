# Wiring main.go

The final step is connecting everything in `main.go`. This is pure plumbing — no new concepts, just assembling the pieces in order.

```go
func main() {
    dsn := os.Getenv("DATABASE_URL")
    database := db.NewDB(dsn)
    defer database.Close()

    if err := db.RunMigrations(database); err != nil {
        log.Fatalf("failed to run migrations: %v", err)
    }

    todoStore := store.NewTodoStore(database)
    h := NewHandler(todoStore)
    router := setupRouter(h)

    if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
        lambda.Start(ginHandler(router))
    } else {
        if err := router.Run(":8080"); err != nil {
            log.Fatalf("failed to run server: %v", err)
        }
    }
}
```

## The dependency chain

Each line produces something the next line needs:

```
DATABASE_URL (env var)
  → db.NewDB(dsn)        → *bun.DB
  → db.RunMigrations()   → schema in place
  → store.NewTodoStore() → *TodoStore (implements TodoStorer)
  → NewHandler()         → *Handler
  → setupRouter()        → *gin.Engine
  → router.Run()         → HTTP server
```

This is dependency injection by hand. No framework, no reflection, no magic — just constructors passing results to the next constructor.

## defer database.Close()

`defer` schedules a function call to run when the surrounding function returns. `database.Close()` shuts down the connection pool — flushing any in-flight connections and releasing resources.

In practice, `main` only returns on a clean shutdown (e.g., the server exits). For a Lambda function it's even less relevant since the process is terminated by the runtime. But it's correct practice to always close resources you open, and `defer` is the idiomatic Go way to ensure cleanup happens even if the function exits early.

## log.Fatalf for startup errors

If `RunMigrations` fails, the server should not start. `log.Fatalf` prints the error and calls `os.Exit(1)` — immediate termination with a non-zero exit code. This is the right pattern for startup failures: there's nothing useful the server can do without a valid schema.

## DATABASE_URL

The DSN (Data Source Name) is read from the environment, not hardcoded. This is standard practice:

- Different environments (local, staging, production) use different databases
- Credentials never appear in source code or git history
- Deployment platforms (AWS Lambda, Docker, Kubernetes) inject secrets as environment variables

Locally, you set it in `.env` (which is gitignored) and either `source .env` or prefix the run command:

```bash
DATABASE_URL=postgres://todo:todo@localhost:5432/todo?sslmode=disable go run .
```

## models.go is gone

With the `Todo` struct now living in `store/todo.go` and all data access going through the store, `models.go` — which held the global `todos` slice and the old `Todo` struct — has no remaining purpose and can be deleted.

Deleting a file is as significant as adding one. The compiler will tell you immediately if anything still depends on it.
