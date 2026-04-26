# Connecting Lambda to RDS

## Why localhost fails in Lambda

When running locally, `DATABASE_URL` points to `localhost:5432` — your Docker Postgres. That works because both the Go app and the container run on your machine.

Lambda is different. Each invocation runs in an isolated container managed by AWS. There is no local Postgres process. When the function tries to connect to `localhost:5432`, it's looking for a database on the same machine as the Lambda container — which doesn't exist. The connection is refused immediately, and the function crashes before handling any request.

The fix is a real remote database that Lambda can reach over the network.

## AWS RDS

RDS (Relational Database Service) is a managed Postgres instance running on AWS infrastructure. AWS handles backups, patching, and restarts. From your app's perspective it's just a Postgres server with a hostname.

### Free tier

New AWS accounts (created after July 15, 2025) get 6 months of `db.t4g.micro` free. Older accounts get 12 months.

### Public vs private access

RDS instances can be publicly accessible (reachable from the internet) or private (only reachable from within a VPC). For simplicity, this chapter uses a public endpoint — Lambda connects directly by hostname without any VPC configuration.

The trade-off: the database port is open to the internet. Anyone who discovers the endpoint can attempt to connect. Valid credentials are still required to get in, but it's a broader attack surface than necessary. The proper fix — putting Lambda and RDS in the same private VPC and restricting the security group to Lambda only — is covered in chapter 09.

## Security groups

A security group is a stateful firewall attached to an AWS resource. It defines which traffic is allowed in (inbound rules) and out (outbound rules).

By default, an RDS instance's security group has no inbound rules — all connections are blocked. To allow Lambda to connect, you add an inbound rule:

- **Type:** PostgreSQL
- **Port:** 5432
- **Source:** `0.0.0.0/0` (anywhere)

Security groups are separate from RDS — they live in EC2 and can be reused across resources. When you find the default security group isn't a clickable link in the RDS console, you navigate to EC2 → Security Groups to edit it.

## Creating the database

RDS provisions the Postgres server but not the individual database inside it. After the instance is available, you connect with `psql` and create it manually:

```bash
psql -h <rds-endpoint> -U todo -d postgres
```

```sql
CREATE DATABASE todo;
```

The `postgres` database is a system database that always exists — you connect to it first, then create your own. After this, migrations run automatically at Lambda startup via `db.RunMigrations`.

## The nil slice bug

After connecting to an empty database, the API was returning `null` instead of `[]` for an empty todo list. This caused the frontend to crash.

In Go, there are two ways to declare an empty slice:

```go
var todos []Todo        // nil slice   → JSON: null
todos := make([]Todo, 0) // empty slice → JSON: []
```

A nil slice and an empty slice behave identically in most Go code — `len`, `append`, and `range` all work on both. But `encoding/json` treats them differently: a nil slice serialises to `null`, an empty slice serialises to `[]`.

The fix is in `store/todo.go`, at the point where the slice is declared before the database query:

```go
func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) {
    todos := make([]Todo, 0)  // never nil, even with zero rows
    err := s.db.NewSelect().
        Model(&todos).
        OrderExpr("id ASC").
        Scan(ctx)
    return todos, err
}
```

This is the right layer for the fix. The store is responsible for returning a valid collection — the handler shouldn't need to compensate for a nil coming from below.
