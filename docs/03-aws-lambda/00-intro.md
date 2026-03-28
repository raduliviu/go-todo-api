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

## The problem Lambda Web Adapter solves

Lambda doesn't speak HTTP natively. It receives a JSON event and expects a JSON response. A typical Lambda handler looks like this:

```go
func handler(event MyEvent) (MyResponse, error) { ... }
```

Your Gin app, on the other hand, is a standard HTTP server. It listens on a port and speaks HTTP. These two models are incompatible out of the box.

The **Lambda Web Adapter (LWA)** bridges the gap. It's an AWS-published extension that:

1. Starts your Go binary as a subprocess
2. Receives the Lambda JSON event from API Gateway
3. Translates it into a standard HTTP request
4. Forwards it to your app on `localhost:8080`
5. Returns the HTTP response back to Lambda as JSON

The result: your `main.go` stays exactly as written. No AWS imports, no handler rewrites, no coupling to Lambda's event format.

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
- A `template.yaml` that defines the full infrastructure
- A working local development setup via `sam local`
- A live API on AWS behind a real API Gateway URL

The in-memory todo store stays for now — the API will work end-to-end on Lambda, just without persistence between cold starts. We'll add a database in a later chapter.
