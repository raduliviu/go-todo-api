# Go TODO API

A learning project building a CRUD API for a TODO tracker using [Go](https://go.dev/) and [Gin](https://gin-gonic.com/).

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
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"id": 4, "title": "Submit new todo", "completed": false}'
```

### PATCH BY ID

```bash
curl -X PATCH http://localhost:8080/todos/1 -H "Content-Type: application/json" -d '{"completed": true}'
```

### DELETE BY ID

```bash
curl -X DELETE http://localhost:8080/todos/1 -H "Content-Type: application/json"
```
