# Go Project Structure

In TypeScript/Node, there's no standard project layout — you'll see `src/`, `lib/`, `app/`, or whatever the team prefers. Go has stronger conventions, and they evolve as your project grows.

## Level 1: Flat layout

For small projects, keep everything in the root with `package main`:

```
go-todo-api/
  main.go
  handlers.go
  models.go
  go.mod
  go.sum
```

All files share the same package, so they can access each other's functions and types without imports. This is similar to how multiple `.ts` files in the same directory might all contribute to one module.

This is where our project is right now. It works well when you have a handful of files and one clear purpose.

## Level 2: cmd/ and internal/

As the project grows — say you add a database layer, middleware, configuration — the flat layout gets crowded. The Go community convention is:

```
go-todo-api/
  cmd/
    api/
      main.go           # entry point — just wires things together
  internal/
    handlers/
      todos.go          # HTTP handlers
    models/
      todo.go           # data types
    database/
      postgres.go       # DB connection and queries
  go.mod
  go.sum
```

### cmd/

Contains your application entry points. Each subdirectory is a separate binary. If your project only builds one binary, you'd have `cmd/api/main.go` (or `cmd/server/main.go`).

In TypeScript terms, think of `cmd/` as where your `index.ts` files live — they import and wire things together but contain minimal logic.

### internal/

This is a special directory in Go. Code inside `internal/` can only be imported by code in the parent module — it's **enforced by the compiler**, not by convention. This is Go's version of "private packages."

In TypeScript, the closest equivalent would be not exporting something from your package's `index.ts`, but Go's `internal/` is a stronger guarantee.

### Why separate packages?

Each subdirectory under `internal/` becomes its own package with its own namespace. This means:

- Types in `internal/models` are imported as `models.Todo`
- Functions in `internal/handlers` are imported as `handlers.GetTodos`
- Dependencies are explicit — you can see exactly what each package uses

In TypeScript, you'd achieve this with separate files and import statements. The concept is the same, Go just enforces it at the package level.

## When to move from flat to cmd/internal

There's no hard rule, but good signals are:

- You're adding a layer that feels like its own concern (database, auth, config)
- You want to enforce boundaries between parts of your code
- Files are getting hard to navigate
- You need multiple binaries (e.g., an API server and a CLI tool)

Don't restructure prematurely. A flat layout is perfectly fine for a small API. Move to `cmd/internal` when the flat layout starts feeling cramped — not before.
