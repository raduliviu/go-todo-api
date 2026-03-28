# The SAM Template

`template.yaml` is the infrastructure definition for the API. It tells AWS what to create — a Lambda function, an API Gateway, and an IAM role — and how they connect.

SAM (Serverless Application Model) is a layer on top of CloudFormation. When you deploy, SAM expands its higher-level types into the underlying CloudFormation resources they represent.

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
    Timeout: 30
```

Default settings applied to all functions. Here, a 30-second timeout: if the Lambda doesn't respond within 30 seconds, AWS terminates the invocation. Can be overridden per function.

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
    Events:
      ApiEvent:
        Type: Api
        Properties:
          Path: /{proxy+}
          Method: ANY
  Metadata:
    BuildMethod: makefile
```

`TodoFunction` is the logical name — used in the Makefile target and referenced elsewhere in the template.

`CodeUri: .` tells SAM where your source code is (the project root).

`Handler: bootstrap` is the binary Lambda executes — must match the filename the Makefile produces.

`Runtime: provided.al2023` means "bring your own binary on Amazon Linux 2023". Go no longer has a managed runtime on Lambda.

## API Gateway

`Events` is what creates the API Gateway. SAM sees this block and automatically provisions an API Gateway that triggers the Lambda on incoming HTTP requests.

`/{proxy+}` is a catch-all path — it matches all routes (`/todos`, `/todos/1`, etc.) and forwards them to the Lambda. Gin's router then handles the actual routing.

`Method: ANY` matches all HTTP methods (GET, POST, PATCH, DELETE).

## Build method

`BuildMethod: makefile` tells SAM to use the `Makefile` to build this function instead of its default build process. Without this, SAM wouldn't know to call `build-TodoFunction`.

## What SAM creates

Running `sam deploy` from this template provisions:

| Resource | What it is |
|---|---|
| Lambda function | Runs your Go binary on demand |
| API Gateway | Public HTTPS endpoint that triggers the Lambda |
| IAM execution role | Grants Lambda permission to write logs to CloudWatch |
| S3 bucket | Stores the deployment artifact (created by SAM automatically) |
