package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/raduliviu/go-todo-api/db"
	"github.com/raduliviu/go-todo-api/store"
)

func setupRouter(h *Handler) *gin.Engine {
	server := gin.Default()

	allowOrigins := os.Getenv("ALLOWED_ORIGINS")

	splitOrigins := strings.Split(allowOrigins, ",")

	server.Use(cors.New(cors.Config{
		AllowOrigins: splitOrigins,
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type"},
	}))

	server.GET("/todos", h.getTodos)
	server.GET("/todos/:id", h.getTodoByID)
	server.POST("/todos", h.createTodo)
	server.PATCH("/todos/:id", h.updateTodoByID)
	server.DELETE("/todos/:id", h.deleteTodoByID)

	return server
}

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

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretArn),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		return "", err
	}

	// Decrypts secret using the associated KMS key.
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

func main() {
	dsn, err := getDSN()
	if err != nil {
		log.Fatalf("failed to get DSN: %v", err)
	}
	database := db.NewDB(dsn)
	defer database.Close()
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	todoStore := store.NewTodoStore(database)
	h := NewHandler(todoStore)
	router := setupRouter(h)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(ginHandler(router))
	} else {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}

}
