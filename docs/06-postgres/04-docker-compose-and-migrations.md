# Docker Compose and SQL Migrations

## Docker Compose for local Postgres

`docker-compose.yml` defines a local Postgres instance for development:

```yaml
services:
  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: todo
      POSTGRES_PASSWORD: todo
      POSTGRES_DB: todo
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
volumes:
  postgres_data:
```

A few things worth understanding:

**`postgres:17-alpine`** — Alpine is a minimal Linux distribution. The Alpine variant of the Postgres image is significantly smaller than the default Debian-based one. For a local development container this doesn't matter much, but it's a good habit.

**`volumes: postgres_data`** — Without a named volume, all data would be lost every time the container stops. The named volume persists data to disk and is remounted when the container restarts. `docker compose down` removes containers but keeps volumes; `docker compose down -v` removes both.

**`ports: "5432:5432"`** — Maps port 5432 on your laptop to port 5432 inside the container. Format is `host:container`. This is what lets TablePlus and the Go app connect to Postgres as if it were running locally.

Useful commands:
```bash
docker compose up -d       # start in background
docker compose down        # stop and remove containers (data persists)
docker compose down -v     # stop and remove containers + volumes (data gone)
docker compose logs db     # view Postgres logs
```

## SQL migrations

A migration is a versioned, ordered SQL script that transforms the database schema from one state to the next. Each migration has an **up** file (apply the change) and a **down** file (reverse it).

```
db/migrations/
  000001_create_todos_table.up.sql
  000001_create_todos_table.down.sql
```

The naming convention `000001_<description>.up.sql` is required by golang-migrate. The numeric prefix determines execution order.

**Up migration:**

```sql
CREATE TABLE todos (
    id         BIGSERIAL    PRIMARY KEY,
    title      TEXT         NOT NULL,
    completed  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

- `BIGSERIAL` — auto-incrementing 64-bit integer. Maps to Go's `int64`.
- `TIMESTAMPTZ` — timestamp with timezone. `created_at` is tracked by the database automatically but not exposed in the Go struct — Bun won't SELECT it, but the DB records it for auditing.
- `DEFAULT FALSE` / `DEFAULT NOW()` — Postgres fills these in if the INSERT doesn't provide them.

**Down migration:**

```sql
DROP TABLE IF EXISTS todos;
```

`IF EXISTS` prevents an error if the migration is rolled back before it was ever applied.

## golang-migrate

golang-migrate reads migration files and applies them in order, keeping track of which have already run in a `schema_migrations` table it manages automatically.

Running `Up()` applies all pending migrations. If nothing is pending, it returns `migrate.ErrNoChange` — which we explicitly ignore, since "nothing to do" is not an error.

Migrations run at startup in `main.go`. On the first run, the todos table is created. On every subsequent run, golang-migrate sees the table is already up to date and does nothing.

## embed.FS

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS
```

The `//go:embed` directive bundles files from disk into the compiled binary at build time. At runtime, `migrationsFS` contains the full content of every `.sql` file — no external files needed.

This matters for Lambda deployment: the binary is uploaded to AWS without any surrounding directory structure. If migrations were loaded from disk at runtime, they wouldn't exist in the Lambda environment. Embedding them means the binary is self-contained.

The TypeScript equivalent would be using a bundler to inline file content at build time, though Node.js apps usually just read from disk at runtime and accept the coupling to the filesystem.
