# AWS Secrets Manager

## The problem with credentials in config files

Before this chapter, `samconfig.toml` contained the full database connection string including the password:

```
DatabaseUrl="postgres://todo:hunter2@go-todo-api-db...rds.amazonaws.com:5432/todo"
```

This is a problem. `samconfig.toml` is version-controlled. One accidental commit puts the database password in git history — permanently, even if you delete it in a later commit. Anyone with access to the repo has the password.

AWS Secrets Manager stores credentials outside of your codebase. Your app fetches them at runtime using its AWS identity. The config file only contains an ARN — a resource identifier, not a credential.

## What Secrets Manager stores

Secrets are stored as key/value JSON. When you choose the "Credentials for Amazon RDS database" type, the JSON contains everything needed to connect:

```json
{
  "username": "todo",
  "password": "...",
  "host": "go-todo-api-db.xxxx.eu-central-1.rds.amazonaws.com",
  "port": 5432,
  "dbname": "todo",
  "engine": "postgres"
}
```

AWS knows the secret is tied to a specific RDS instance, which enables automatic password rotation — AWS generates a new password, updates the database, and updates the secret, with no downtime.

## The AWS SDK for Go v2

The AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2`) is the standard way to call AWS services from Go code. The pattern is consistent across all services:

1. Load configuration (region, credentials)
2. Create a service client
3. Call the API

```go
cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
svc := secretsmanager.NewFromConfig(cfg)
result, err := svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
    SecretId: aws.String(secretArn),
})
```

`LoadDefaultConfig` resolves credentials automatically — on Lambda it uses the function's IAM role, locally it uses your `~/.aws` credentials. You don't handle credentials in code.

`result.SecretString` is a `*string` containing the raw JSON. You unmarshal it yourself.

## getDSN

The full pattern in `main.go`:

```go
func getDSN() (string, error) {
    secretArn := os.Getenv("SECRET_ARN")
    localDsn := os.Getenv("DATABASE_URL")
    if secretArn == "" {
        return localDsn, nil
    }

    region := "eu-central-1"

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        return "", err
    }

    svc := secretsmanager.NewFromConfig(cfg)

    input := &secretsmanager.GetSecretValueInput{
        SecretId:     aws.String(secretArn),
        VersionStage: aws.String("AWSCURRENT"), // defaults to AWSCURRENT if unspecified
    }

    result, err := svc.GetSecretValue(context.TODO(), input)
    if err != nil {
        return "", err
    }

    var secretString string = *result.SecretString

    type dbCredentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Host     string `json:"host"`
        Port     int    `json:"port"`
        DBName   string `json:"dbname"`
    }
    var creds dbCredentials

    if err := json.Unmarshal([]byte(secretString), &creds); err != nil {
        return "", err
    }

    dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require", creds.Username, creds.Password, creds.Host, creds.Port, creds.DBName)
    return dsn, nil
}
```

A few things worth noting:

**Branch on `SECRET_ARN`, not on environment.** The local/Lambda distinction is implicit — locally `SECRET_ARN` is not set, so you fall back to `DATABASE_URL`. On Lambda, `SECRET_ARN` is injected by the SAM template. No hardcoded environment names.

**Return errors, don't `log.Fatal`.** `getDSN` returns `(string, error)` — it should not terminate the process. The caller in `main` does that if needed. Functions that return errors should use them.

**Build the DSN from the parsed fields.** `result.SecretString` is JSON, not a connection string. Unmarshal it into a struct, then construct the URL with `fmt.Sprintf`. `sslmode=require` is needed for RDS — it enforces TLS.

## IAM policy in the SAM template

Lambda functions have no permissions by default. To allow the function to call `secretsmanager:GetSecretValue`, you add a policy to `template.yaml`:

```yaml
Policies:
  - AWSSecretsManagerGetSecretValuePolicy:
      SecretArn: !Ref SecretArn
```

`AWSSecretsManagerGetSecretValuePolicy` is a SAM policy template — a shorthand that expands into the correct IAM policy granting `secretsmanager:GetSecretValue` on the specific secret ARN. Scoping it to a single secret (not `*`) follows least-privilege: the function can only read this one secret, nothing else.

## What samconfig.toml now contains

```toml
parameter_overrides = "SecretArn=\"arn:aws:secretsmanager:eu-central-1:...:secret:go-todo-api/db-credentials-xxx\" AllowedOrigins=\"...\""
```

An ARN is a resource identifier — it tells AWS which secret to fetch, but contains no credentials itself. This file is now safe to commit.
