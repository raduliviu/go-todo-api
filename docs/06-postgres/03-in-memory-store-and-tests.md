# The In-Memory Store and Tests

With `TodoStorer` as an interface, tests can use a fake implementation that lives entirely in memory. No Docker, no migrations, no connection strings.

## The inMemoryStore

```go
type inMemoryStore struct {
    todos  map[int64]*store.Todo
    nextID int64
}
```

A map keyed by ID, a counter for generating new IDs. The five methods manipulate the map directly:

- `GetAll` — collect all values, sort by ID (map iteration order is non-deterministic in Go), return as slice
- `GetByID` — look up by key, return `sql.ErrNoRows` if missing
- `Create` — assign `nextID`, insert into map, increment counter
- `Update` — look up by ID, overwrite fields, return `sql.ErrNoRows` if missing
- `Delete` — look up by ID, delete from map, return `sql.ErrNoRows` if missing

The `sql.ErrNoRows` sentinel is used for not-found in both the fake and the real store. That's the point — the handler calls `errors.Is(err, sql.ErrNoRows)` and gets the same result regardless of which implementation is behind the interface.

## Why map, not slice?

The original global state was a `[]Todo` slice. The in-memory store uses `map[int64]*Todo` instead.

With a slice, finding a todo by ID requires scanning every element — `O(n)`. With a map, it's a direct key lookup — `O(1)`. More importantly, delete operations on a slice require rebuilding the slice without the deleted element. On a map, it's a single `delete(m, key)` call.

A database table is conceptually a map (keyed by primary key), so the map matches the mental model better.

## How tests changed

Before:

```go
func TestGetTodos(t *testing.T) {
    resetTodos()
    router := setupRouter()
    ...
}
```

After:

```go
func TestGetTodos(t *testing.T) {
    s := newInMemoryStore()
    h := NewHandler(s)
    router := setupRouter(h)
    ...
}
```

Each test creates a fresh store, a fresh handler, and a fresh router. No shared global state, no need for a reset function. Tests are now fully isolated from each other — one test mutating the store has no effect on any other test.

## The seeded state

`newInMemoryStore()` seeds the map with three todos (IDs 1, 2, 3) matching the old `resetTodos` seed data. This means all existing tests pass without changes to their assertions — the store is different, but the data is the same.

`nextID` starts at 4, so the first `Create` call produces ID 4. Tests that call create and then check the returned ID depend on this.

## sql.ErrNoRows: the not-found sentinel

`database/sql` defines a sentinel error for "query returned no rows":

```go
var ErrNoRows = errors.New("sql: no rows in result set")
```

Bun returns this automatically from `Scan` when a `SELECT` finds nothing. For `UPDATE` and `DELETE`, Bun doesn't return it — Postgres treats zero-row updates/deletes as success. So we check `RowsAffected() == 0` ourselves and return `sql.ErrNoRows` manually.

The in-memory store returns the same sentinel for consistency. The handler checks it the same way in both cases:

```go
if errors.Is(err, sql.ErrNoRows) {
    c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
    return
}
```
