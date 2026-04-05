# Bun Query Builder

Bun is the query builder used in this project (and at work). It sits between raw `database/sql` and a full ORM: you write Go method chains that produce SQL, rather than writing SQL strings by hand or having a framework generate everything automatically.

## The bundebug hook

Before looking at queries, it's worth knowing how to see what SQL Bun generates:

```go
db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
```

This prints every SQL statement to stdout as it executes. When learning Bun, this is invaluable — you can see exactly what your method chain produces before trusting it.

## SELECT — GetAll

```go
func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) {
    var todos []Todo
    err := s.db.NewSelect().
        Model(&todos).
        OrderExpr("id ASC").
        Scan(ctx)
    return todos, err
}
```

Generates: `SELECT t."id", t."title", t."completed" FROM "todos" AS "t" ORDER BY id ASC`

- `Model(&todos)` — tells Bun the target type. It reads the struct tags to determine the table name and column names. Pass a pointer to a slice for multi-row results.
- `OrderExpr("id ASC")` — raw SQL expression for ORDER BY. Bun has a `.Order()` method too, but `OrderExpr` is useful when you need exact control.
- `Scan(ctx)` — executes the query and populates the slice. Returns `nil` if zero rows are found (empty slice is valid for GetAll).

## SELECT — GetByID

```go
func (s *TodoStore) GetByID(ctx context.Context, id int64) (*Todo, error) {
    var todo Todo
    err := s.db.NewSelect().Where("id = ?", id).Model(&todo).Scan(ctx)
    return &todo, err
}
```

Generates: `SELECT t."id", t."title", t."completed" FROM "todos" AS "t" WHERE id = $1`

- `Where("id = ?", id)` — the `?` placeholder is replaced with a properly escaped parameter (`$1` in Postgres). Never interpolate values directly into SQL strings — this is how SQL injection happens.
- `Model(&todo)` — pass a pointer to a single struct for single-row results.
- `Scan` returns `sql.ErrNoRows` if no row matches — the handler maps this to a 404.

## INSERT — Create

```go
func (s *TodoStore) Create(ctx context.Context, todo *Todo) error {
    _, err := s.db.NewInsert().Model(todo).Returning("id").Exec(ctx)
    return err
}
```

Generates: `INSERT INTO "todos" ("title", "completed") VALUES ($1, $2) RETURNING "id"`

- `Returning("id")` — Postgres executes the INSERT and immediately returns the generated `id`. Bun scans it back into `todo.ID`. After `Create` returns, the caller's `todo` struct has its ID populated.
- `Exec(ctx)` — used instead of `Scan` for write operations. Returns `(sql.Result, error)`.
- The first return value (rows affected) is discarded with `_` — for INSERT we only care whether it succeeded.

**Why not return the ID from Create?**

`Returning("id")` mutates the `todo` pointer that was passed in. The caller already has a reference to that struct, so after calling `Create`, they can read `todo.ID` directly. Returning it redundantly would be noise.

## UPDATE — Update

```go
func (s *TodoStore) Update(ctx context.Context, todo *Todo) error {
    result, err := s.db.NewUpdate().Model(todo).Column("title", "completed").WherePK().Exec(ctx)
    if err != nil {
        return err
    }
    numOfRows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if numOfRows == 0 {
        return sql.ErrNoRows
    }
    return nil
}
```

Generates: `UPDATE "todos" SET "title" = $1, "completed" = $2 WHERE "id" = $3`

- `.Column("title", "completed")` — restricts the UPDATE to only these columns. Without it, Bun would update every column including `id`, which is wrong.
- `.WherePK()` — adds a WHERE clause targeting the primary key field (`id`). Equivalent to `.Where("id = ?", todo.ID)`.
- **`RowsAffected()` check** — Postgres doesn't error when an UPDATE matches zero rows. It succeeds silently. We have to check the count ourselves and return `sql.ErrNoRows` manually so the handler can respond with a 404.

## DELETE — Delete

```go
func (s *TodoStore) Delete(ctx context.Context, id int64) error {
    result, err := s.db.NewDelete().Model((*Todo)(nil)).Where("id = ?", id).Exec(ctx)
    if err != nil {
        return err
    }
    numOfRows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if numOfRows == 0 {
        return sql.ErrNoRows
    }
    return nil
}
```

Generates: `DELETE FROM "todos" AS "t" WHERE (id = $1)`

- `Model((*Todo)(nil))` — unlike Update, we don't have a full struct instance here, just an ID. The nil pointer tells Bun which table to target without needing a real value. `(*Todo)(nil)` is a typed nil — a nil pointer of type `*Todo`.
- Same `RowsAffected()` pattern as Update — DELETE on a nonexistent row succeeds silently in Postgres.

## Scan vs Exec

| Method | Use for | Returns |
|--------|---------|---------|
| `Scan(ctx)` | SELECT | Populates model, returns error |
| `Exec(ctx)` | INSERT / UPDATE / DELETE | Returns `(sql.Result, error)` |

`Scan` is designed for reading rows into Go structs. `Exec` is designed for write operations where you care about success/failure and rows affected, not about reading data back.
