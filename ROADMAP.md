# Go Todo API — Learning Roadmap

- [x] **Refactoring** — Split main.go into separate files/packages following Go project layout conventions
- [x] **Dockerizing** — Containerize the API so it's runnable locally
- [x] **AWS Lambda** — Deploy to AWS using Lambda + API Gateway with a native handler and custom Gin adapter
- [x] **Testing** — Write HTTP handler tests using Go's httptest and Gin's test utilities
- [x] **Error handling & validation** — Request structs, custom error responses, field-level validation
- [ ] **PostgreSQL** — Database storage with migrations, connection pooling. Revisit cmd/internal project layout at this stage
- [ ] **Authentication** — Lock the API down to a lightweight frontend deployed on AWS Amplify. Options to evaluate: API Gateway resource policies, Cognito, or API keys
- [ ] **Frontend** — Lightweight frontend deployed to AWS Amplify, consuming the API
- [ ] **Expand AWS usage** — SQS for async processing, EventBridge for scheduled tasks, S3 for file storage
- [ ] **Infrastructure as Code (Terraform)** — Rewrite the SAM/CloudFormation infrastructure in Terraform to learn IaC concepts applicable across cloud providers
