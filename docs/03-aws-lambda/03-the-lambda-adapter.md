# The Lambda Adapter

Lambda doesn't speak HTTP natively. It receives a JSON event from API Gateway and expects a JSON response back. Gin, on the other hand, is a standard HTTP server. The adapter in `lambda.go` bridges the two.

## Why write it ourselves?

The common library for this (`aws-lambda-go-api-proxy/ginadapter`) has been proposed for deprecation in favour of Lambda Web Adapter. We tried Lambda Web Adapter first, but it doesn't work with `sam local start-api` — meaning you can't test Lambda behaviour locally without deploying to AWS.

Writing the adapter ourselves solves both problems: no deprecated dependency, and full local testability via `sam local`. It's also more educational — you can see exactly what the translation does rather than it being hidden in a library.

## What the adapter does

```
API Gateway JSON event
        ↓
  APIGatewayProxyRequest   (decoded by the Lambda SDK)
        ↓
     http.Request          (built by the adapter)
        ↓
    Gin router             (ServeHTTP)
        ↓
  ResponseRecorder         (captures response in memory)
        ↓
  APIGatewayProxyResponse  (returned to Lambda → API Gateway)
```

## The closure pattern

`ginHandler` takes a `*gin.Engine` and returns a function — this is a closure. The returned function is the actual Lambda handler, and it "closes over" `r`, capturing it from the outer scope.

```go
func ginHandler(r *gin.Engine) func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
        // r is accessible here, even though it was defined in the outer function
    }
}
```

This matters for performance: `r` (the Gin router with all its routes) is initialised once when the Lambda container starts, then reused across all invocations. Lambda keeps the container alive between requests ("warm starts"), so the router doesn't need to be rebuilt each time.

In `main.go` it's called as:

```go
lambda.Start(ginHandler(r))
```

`ginHandler(r)` evaluates once, returning the inner function with `r` already captured. `lambda.Start` calls that inner function on each incoming request.

## Request translation

### Body

API Gateway can base64-encode binary payloads (e.g. file uploads). The adapter checks `req.IsBase64Encoded` and decodes if needed before passing the body to Gin. For a JSON API this will almost always be false, but ignoring it would silently corrupt binary requests.

### Query parameters

`req.QueryStringParameters` is a `map[string]string`. `url.Values` is used to encode it correctly — it handles special characters and produces a properly formatted query string.

### Context

`http.NewRequestWithContext` threads the Lambda context into the HTTP request. This matters because the Lambda context carries the invocation's deadline — if Lambda is about to time out, any in-progress work can be cancelled rather than hanging.

### Headers

API Gateway headers are copied directly to the `http.Request`. `Header.Set` is used rather than `Header.Add` since API Gateway sends each header once.

## Response translation

`httptest.ResponseRecorder` is a standard library type that implements `http.ResponseWriter` — so Gin writes its response into it exactly as it would write to a real network connection, but the output is captured in memory.

`MultiValueHeaders` is used instead of `Headers` because `recorder.Header()` returns `map[string][]string`, which matches it directly. It also correctly handles headers that appear multiple times, like `Set-Cookie`.

## The result

`main.go` can now switch between two modes:

```go
if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
    lambda.Start(ginHandler(r))  // Lambda or sam local
} else {
    r.Run(":8080")               // local development
}
```

`AWS_LAMBDA_FUNCTION_NAME` is set automatically by both real Lambda and `sam local start-api`, so no manual configuration is needed to switch between modes.
