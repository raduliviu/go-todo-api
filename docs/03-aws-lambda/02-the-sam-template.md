# The SAM Template

`template.yaml` is the infrastructure definition for the API. It tells AWS what to create — a Lambda function, an API Gateway, and an IAM role — and how they connect.

SAM (Serverless Application Model) is a layer on top of CloudFormation. When you deploy, AWS expands SAM's higher-level types into the underlying CloudFormation resources they represent.

## Header

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
```

`AWSTemplateFormatVersion` is a fixed string — always this value, not a date you set.

`Transform` tells CloudFormation this is a SAM template. Without it, CloudFormation wouldn't understand SAM-specific resource types like `AWS::Serverless::Function`.

## Globals

```yaml
Globals:
  Function:
    Timeout: 10
```

Default settings applied to all functions. Here, a 10-second timeout: if the Lambda doesn't respond within 10 seconds, AWS terminates the invocation. Can be overridden per function.

## The function

```yaml
TodoFunction:
  Type: AWS::Serverless::Function
  Properties:
    CodeUri: .
    Handler: bootstrap
    Runtime: provided.al2023
    Architectures:
      - x86_64
```

`TodoFunction` is the logical name — used in the Makefile target and referenced elsewhere in the template.

`CodeUri: .` tells SAM where your source code is (the project root).

`Handler: bootstrap` is the binary Lambda executes — must match the filename the Makefile produces.

`Runtime: provided.al2023` means "bring your own binary on Amazon Linux 2023". Go no longer has a managed runtime on Lambda.

## Lambda Web Adapter

```yaml
Layers:
  - arn:aws:lambda:eu-central-1:753240598075:layer:LambdaAdapterLayerX86:22
Environment:
  Variables:
    PORT: "8080"
    AWS_LAMBDA_EXEC_WRAPPER: /opt/bootstrap
```

The Lambda Web Adapter (LWA) is an AWS-published extension that bridges Lambda and standard HTTP servers. Without it, your Gin app would need to be rewritten to understand Lambda's event format.

`AWS_LAMBDA_EXEC_WRAPPER: /opt/bootstrap` tells Lambda to run LWA first. LWA then starts your `bootstrap` binary as a subprocess.

`PORT: "8080"` tells LWA which port to forward requests to — matching `r.Run(":8080")` in `main.go`.

This is why `main.go` needed no changes: LWA translates Lambda events into standard HTTP requests your existing Gin router already understands.

## API Gateway

```yaml
Events:
  ApiEvent:
    Type: Api
    Properties:
      Path: /{proxy+}
      Method: ANY
```

SAM automatically provisions an API Gateway from this block. No separate resource needed.

`/{proxy+}` is a catch-all path — it matches all routes (`/todos`, `/todos/1`, etc.) and forwards them to the Lambda. Gin's router then handles the actual routing.

`Method: ANY` matches all HTTP methods (GET, POST, PATCH, DELETE).

## Build method

```yaml
Metadata:
  BuildMethod: makefile
```

Tells SAM to use the `Makefile` to build this function instead of its default build process. Without this, SAM wouldn't know to call `build-TodoFunction`.

## What SAM creates

Running `sam deploy` from this template provisions:

| Resource | What it is |
|---|---|
| Lambda function | Runs your Go binary on demand |
| API Gateway | Public HTTPS endpoint that triggers the Lambda |
| IAM execution role | Grants Lambda permission to write logs to CloudWatch |
| S3 bucket | Stores the deployment artifact (created by SAM automatically) |
