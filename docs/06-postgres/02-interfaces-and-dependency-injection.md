# Interfaces and Dependency Injection

The `Handler` struct holds a `TodoStorer` — not a `*TodoStore`. That distinction is what makes the whole system testable without a real database.

## The interface

```go
type TodoStorer interface {
    GetAll(ctx context.Context) ([]Todo, error)
    GetByID(ctx context.Context, id int64) (*Todo, error)
    Create(ctx context.Context, todo *Todo) error
    Update(ctx context.Context, todo *Todo) error
    Delete(ctx context.Context, id int64) error
}
```

This is a contract: anything that implements all five methods satisfies `TodoStorer`. The handler doesn't care what's behind the interface — real Postgres, in-memory map, or a mock. As long as the methods match, it works.

## Go interfaces vs TypeScript interfaces

In TypeScript, a class explicitly declares which interfaces it implements:

```typescript
class TodoService implements ITodoService { ... }
```

Go has no `implements` keyword. Satisfaction is **implicit** — if a type has all the methods, it satisfies the interface automatically:

```go
// TodoStore never mentions TodoStorer anywhere
// But it satisfies it because it has all five methods
var _ TodoStorer = (*TodoStore)(nil) // optional compile-time check
```

This is called **structural typing** (or duck typing at the type system level). The interface and the implementation are completely decoupled — `TodoStore` doesn't need to import the package that defines `TodoStorer`.

The `var _ TodoStorer = (*TodoStore)(nil)` line is a common Go idiom for asserting interface satisfaction at compile time without allocating anything. `(*TodoStore)(nil)` is a nil pointer of type `*TodoStore` — it has no value, but the compiler still checks whether the type has all the required methods.

## Dependency injection

`NewHandler` takes a `TodoStorer`:

```go
func NewHandler(s TodoStorer) *Handler {
    return &Handler{store: s}
}
```

This is dependency injection — the handler doesn't create its own store, it receives one. That means in production you pass a real `*TodoStore`, and in tests you pass an `inMemoryStore`. The handler code is identical in both cases.

In TypeScript you'd typically use a DI framework (NestJS, tsyringe) for this. In Go, manual constructor injection is the norm — no framework needed, just pass the dependency as an argument.

## Why this matters for tests

Without an interface, testing a handler that hits Postgres requires a real Postgres connection. That means:

- Tests need Docker or a CI database
- Tests are slow (network round trips)
- Tests fail if the database is down

With an interface, tests pass in an in-memory implementation that runs in microseconds, needs no external processes, and can be seeded with exactly the data each test needs.

The real `TodoStore` is only exercised in integration tests or manual testing against a real database. Unit tests never touch it.
