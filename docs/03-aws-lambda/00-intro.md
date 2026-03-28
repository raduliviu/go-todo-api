# Hosting on AWS Lambda

So far the API runs locally and inside a Docker container. This chapter gets it running on AWS — publicly accessible, no server to manage.

## The architecture

```
Client → API Gateway → Lambda → your Go binary
```

Three services, each with a distinct responsibility:

**API Gateway** is the public front door. It receives HTTP requests from the internet and forwards them to Lambda. It handles HTTPS termination, request routing, and throttling. You never interact with it directly in code — SAM provisions it from `template.yaml`.

**Lambda** is the compute layer. Instead of a long-running server, Lambda runs your binary on demand — one invocation per request. AWS manages the infrastructure entirely: no EC2 instances, no OS patches, no capacity planning. You pay only for the time your code actually runs.

**IAM** (Identity and Access Management) controls permissions. Every Lambda function runs under an IAM role that defines what AWS services it's allowed to call. SAM creates a basic role automatically — later, when we add DynamoDB or SQS, we'll extend it.

## Bridging Lambda and Gin

Lambda doesn't speak HTTP natively. It receives a JSON event from API Gateway and expects a JSON response back:

```go
func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) { ... }
```

Gin, on the other hand, is a standard HTTP server. These two models are incompatible out of the box.

We bridge them with a custom adapter in `lambda.go` that translates API Gateway events into `http.Request` objects, runs them through Gin's router, and converts the response back. This keeps all existing handler code unchanged while teaching the actual Lambda programming model.

An alternative — the Lambda Web Adapter (LWA) — does this translation transparently without code changes. We tried it first but switched away: LWA doesn't work with `sam local start-api`, meaning you can't test Lambda behaviour locally without deploying to AWS. The native handler approach gives you full local testability.

## Dual-mode main.go

Because the binary now needs to behave differently in two environments, `main.go` checks an environment variable to decide which mode to start in:

```go
if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
    lambda.Start(ginHandler(r))  // Lambda or sam local
} else {
    r.Run(":8080")               // local development
}
```

`AWS_LAMBDA_FUNCTION_NAME` is set automatically by both real Lambda and `sam local` — no manual configuration needed. `go run .` continues to work exactly as before.

## The tooling: SAM

SAM (Serverless Application Model) is AWS's official tool for building and deploying Lambda applications. It provides:

- A `template.yaml` format for defining infrastructure (Lambda, API Gateway, IAM) as code
- `sam build` — compiles your code for the Lambda runtime
- `sam local start-api` — runs the full stack locally, simulating API Gateway + Lambda on your machine
- `sam deploy` — packages and deploys everything to AWS

SAM is a layer on top of CloudFormation, AWS's general infrastructure-as-code service. When you deploy, SAM expands your template into CloudFormation resources and lets CloudFormation do the actual provisioning.

## What we're building

By the end of this chapter:

- A `Makefile` that cross-compiles the Go binary for Lambda's Linux environment
- A `lambda.go` adapter that translates between Lambda events and Gin's HTTP handler
- A `template.yaml` that defines the full infrastructure
- A dual-mode `main.go` that runs as an HTTP server locally and as a Lambda handler in AWS
- A live API on AWS behind a real API Gateway URL

The in-memory todo store stays for now — the API will work end-to-end on Lambda, just without persistence between cold starts. We'll add a database in a later chapter.
