# CORS in Gin

## What CORS is

Browsers enforce the **Same-Origin Policy**: a page at `http://localhost:5173` is not allowed to make requests to `http://localhost:8080` unless the server explicitly says it's okay. This is a browser security rule — `curl` and Postman are unaffected, which is why the API worked fine for manual testing before the frontend existed.

**CORS (Cross-Origin Resource Sharing)** is the mechanism by which a server opts in to cross-origin requests. It works via HTTP headers — the server includes `Access-Control-Allow-Origin` in its response, and the browser decides whether to allow the JavaScript to read the response.

## The middleware

```go
server.Use(cors.New(cors.Config{
    AllowOrigins: []string{allowOrigins},
    AllowMethods: []string{"GET", "POST", "PATCH", "DELETE"},
    AllowHeaders: []string{"Content-Type"},
}))
```

`server.Use(...)` registers middleware that runs on every request before any route handler. The cors middleware intercepts the request, checks the `Origin` header, and adds the appropriate `Access-Control-*` headers to the response.

`AllowMethods` and `AllowHeaders` matter for **preflight requests** — browsers send an `OPTIONS` request before any non-simple request (POST, PATCH, DELETE, or requests with custom headers) to ask the server what it allows. The cors middleware handles these automatically.

## Environment-driven origin

```go
allowOrigins := os.Getenv("ALLOWED_ORIGINS")
```

The allowed origin is read from an environment variable rather than hardcoded. This means:

- Locally: `ALLOWED_ORIGINS=http://localhost:5173`
- Production: `ALLOWED_ORIGINS=https://your-app.amplifyapp.com`

Same binary, different config per environment. Never hardcode origins — the Amplify URL won't be known until deployment, and hardcoding `localhost` into a production binary would leave the API open to local development frontends.

## What happens in production

When the frontend is deployed to Amplify, `ALLOWED_ORIGINS` gets set to the Amplify domain in the Lambda environment variables. The API will then only accept requests from that domain — any other origin (including `localhost`) will be blocked by the browser.
