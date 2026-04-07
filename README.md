# Go TODO API

A learning project building a CRUD API for a TODO tracker using [Go](https://go.dev/) and [Gin](https://gin-gonic.com/).

## Running locally

**Prerequisites:** Docker, Go 1.21+

1. Copy the example env file and fill in your values:

   ```bash
   cp .env.example .env
   ```

2. Start Postgres:

   ```bash
   docker compose up -d
   ```

3. Run the API (migrations run automatically on startup):

   ```bash
   source .env && go run .
   ```

The API will be available at `http://localhost:8080`.

To stop Postgres: `docker compose down`

---

## CURL testing commands

Reference, for quick testing

### GET

```bash
curl -X GET http://localhost:8080/todos -H "Content-Type: application/json"
```

### GET BY ID

```bash
curl -X GET http://localhost:8080/todos/1 -H "Content-Type: application/json"
```

### POST

```bash
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title": "Submit new todo", "completed": false}'
```

### PATCH BY ID

```bash
curl -X PATCH http://localhost:8080/todos/1 -H "Content-Type: application/json" -d '{"completed": true}'
```

### DELETE BY ID

```bash
curl -X DELETE http://localhost:8080/todos/1 -H "Content-Type: application/json"
```
