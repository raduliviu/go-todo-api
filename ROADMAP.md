# Go Todo API — Learning Roadmap

- [x] **Refactoring** — Split main.go into separate files/packages following Go project layout conventions
- [x] **Dockerizing** — Containerize the API so it's runnable locally
- [x] **AWS Lambda** — Deploy to AWS using Lambda + API Gateway with a native handler and custom Gin adapter
- [x] **Testing** — Write HTTP handler tests using Go's httptest and Gin's test utilities
- [x] **Error handling & validation** — Request structs, custom error responses, field-level validation
- [x] **PostgreSQL** — Bun query builder, pgdriver, golang-migrate, RDS on AWS, Secrets Manager for credentials
- [x] **Frontend** — Lightweight frontend deployed to AWS Amplify, consuming the API
- [ ] **VPC & Network Security** — Move Lambda and RDS into a private VPC, remove public RDS exposure, restrict port 5432 to Lambda's security group only
- [ ] **Infrastructure as Code (Terraform)** — Rewrite the SAM/CloudFormation infrastructure in Terraform before adding more AWS services
- [ ] **Authentication** — Lock the API down. Options: API Gateway resource policies, Cognito, or API keys. Built in Terraform
- [ ] **Expand AWS usage** — SQS for async processing, EventBridge for scheduled tasks, S3 for file storage. All built in Terraform
