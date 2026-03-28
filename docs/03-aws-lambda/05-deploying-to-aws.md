# Deploying to AWS

This document covers everything needed to go from local code to a live API on AWS.

## Prerequisites

### 1. AWS account

You need an AWS account. If you don't have one, create it at [aws.amazon.com](https://aws.amazon.com). A free tier account is sufficient for this project.

### 2. IAM user with programmatic access

You need an IAM user with:
- Programmatic access enabled (access key + secret key)
- `AdministratorAccess` policy attached, or at minimum permissions for: Lambda, API Gateway, IAM, S3, CloudFormation

To check your current user and attached policies:
```bash
aws iam get-user
aws iam list-attached-user-policies --user-name YOUR_USERNAME
```

### 3. AWS CLI installed and configured

Install:
```bash
brew install awscli
```

Configure with your credentials:
```bash
aws configure
```

You'll be prompted for:
- AWS Access Key ID
- AWS Secret Access Key
- Default region — enter `eu-central-1`
- Default output format — enter `json`

Verify it works:
```bash
aws sts get-caller-identity
```

This should return your account ID, user ID, and ARN. If it does, your credentials are valid.

### 4. SAM CLI installed

```bash
brew install aws-sam-cli
sam --version
```

### 5. Docker running

SAM uses Docker for `sam build` and `sam local`. Make sure Docker Desktop is running before using any SAM commands.

### 6. Code compiled for Lambda

Before deploying, always build first:
```bash
sam build
```

This cross-compiles the Go binary for Linux and places it in `.aws-sam/build/`. SAM deploys this artifact — not your source files directly.

---

## First deploy

Run the guided deploy:
```bash
sam deploy --guided
```

You'll be asked a series of questions:

| Prompt | What to enter | Why |
|---|---|---|
| Stack Name | `go-todo-api` | The name of the CloudFormation stack AWS will create |
| AWS Region | `eu-central-1` | The region to deploy to |
| Confirm changes before deploy | `y` | Always review what AWS is about to create |
| Allow SAM CLI IAM role creation | `Y` | SAM needs to create the Lambda execution role |
| Disable rollback | `N` | If deploy fails, AWS will revert to the last working state |
| TodoFunction has no authentication | `y` | Acceptable for a learning project |
| Save arguments to configuration file | `Y` | Saves choices to `samconfig.toml` for future deploys |
| SAM configuration file | (Enter) | Accept default `samconfig.toml` |
| SAM configuration environment | (Enter) | Accept default `default` |

After confirming the changeset, CloudFormation creates all resources in order:

1. `TodoFunctionRole` — IAM execution role
2. `TodoFunction` — the Lambda function
3. `ServerlessRestApi` — the API Gateway
4. `TodoFunctionApiEventPermissionProd` — permission allowing API Gateway to invoke Lambda
5. `ServerlessRestApiDeployment` — API Gateway deployment
6. `ServerlessRestApiProdStage` — the `Prod` stage

---

## Finding your API URL

### Via AWS CLI
```bash
aws apigateway get-rest-apis \
  --region eu-central-1 \
  --query 'items[?name==`go-todo-api`].id' \
  --output text
```

This returns the API ID. Your base URL is:
```
https://{API_ID}.execute-api.eu-central-1.amazonaws.com/Prod
```

### Via AWS Console

Go to the [API Gateway console](https://eu-central-1.console.aws.amazon.com/apigateway/home?region=eu-central-1), click `go-todo-api`, then **Stages → Prod**. The invoke URL is shown at the top.

---

## Testing the live API

```bash
BASE_URL=https://{API_ID}.execute-api.eu-central-1.amazonaws.com/Prod

curl $BASE_URL/todos
curl $BASE_URL/todos/1
curl -X POST $BASE_URL/todos -H "Content-Type: application/json" -d '{"title":"Buy milk","completed":false}'
curl -X PATCH $BASE_URL/todos/1 -H "Content-Type: application/json" -d '{"completed":true}'
curl -X DELETE $BASE_URL/todos/2
```

Hitting `$BASE_URL` directly (no path) returns a `Missing Authentication Token` error from API Gateway — this is expected, as there is no route defined for `/`.

---

## Subsequent deploys

Once `samconfig.toml` exists, future deploys are a single command:
```bash
sam build && sam deploy
```

No prompts — SAM uses the saved configuration.

---

## Warm vs cold containers

Lambda keeps containers alive between requests ("warm start"). This means in-memory state (the `todos` slice) persists across consecutive requests on the same container — you can POST a todo and immediately GET it back.

However, this is not reliable storage:
- After a period of inactivity (~15-20 minutes), Lambda discards the container. The next request triggers a cold start, and the slice resets to seed data.
- Under concurrent traffic, Lambda may spin up multiple containers. Each has its own isolated slice — a todo created in one container won't be visible to requests hitting another.

This is why a database is essential for Lambda. We'll add one in a later chapter.

---

## What was created in AWS

You can inspect all created resources in the [CloudFormation console](https://eu-central-1.console.aws.amazon.com/cloudformation/home?region=eu-central-1) under the `go-todo-api` stack.

| Resource | Purpose |
|---|---|
| Lambda function | Runs the Go binary on each request |
| API Gateway (REST API) | Public HTTPS endpoint, routes traffic to Lambda |
| API Gateway Stage (Prod) | The deployment stage — part of the URL path |
| IAM Role | Grants Lambda permission to write logs to CloudWatch |
| S3 Bucket | Stores the deployment artifact (managed by SAM) |
