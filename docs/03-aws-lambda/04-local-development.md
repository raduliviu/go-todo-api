# Local Development with SAM

`sam local start-api` simulates the full Lambda + API Gateway stack on your machine using Docker. It's the closest you can get to real Lambda behaviour without deploying to AWS.

## Ports

SAM local exposes your API on port 3000 by default, not 8080:

```
you → localhost:3000 → SAM local proxy → Lambda container → your app on :8080
```

Your app still listens on 8080 internally — that doesn't change. Port 3000 is SAM's external proxy port. SAM reserves 8080 for its own internal runtime API (the communication channel between SAM and the Lambda container), so it can't also expose your app there.

This is the same concept as Docker's `-p` flag — the external port and the internal port are independent.

You can change the external port with the `--port` flag:

```bash
sam local start-api --port 9000
```

## Cold starts

SAM local spins up a fresh Docker container for every request. This means:

- Gin's debug output (route registration) appears on every invocation
- Each request shows an `Init Duration` in the REPORT line

In real Lambda, the container is kept alive between requests ("warm start"). `Init Duration` only appears on cold starts — the first invocation after a container is created. Subsequent requests on the same container skip init entirely and only pay the `Duration` cost.

SAM local's per-request containers make cold start behaviour more visible than you'd see in production.

## In-memory state

Because each SAM local request gets a fresh container, the in-memory `todos` slice resets on every invocation. A POST that creates a todo is immediately lost — the next GET starts from the original seed data.

In real Lambda this is partially better: a warm container persists the slice between requests. But if Lambda scales to multiple containers to handle concurrent traffic, each has its own slice. Any cold start also resets it.

This is the fundamental limitation of in-memory storage in a serverless environment. It's why a database is essential — Lambda functions are stateless by design. We'll add one in a later chapter.

## Workflow

| Goal | Command |
|---|---|
| Quick local dev (no Lambda simulation) | `go run .` → `localhost:8080` |
| Test Lambda behaviour locally | `sam build && sam local start-api` → `localhost:3000` |
| Deploy to AWS | `sam deploy` |

Run `sam build` after any code change before using `sam local` — SAM local uses the compiled binary in `.aws-sam/build/`, not your source files directly.
