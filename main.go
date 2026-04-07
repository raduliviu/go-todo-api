package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/raduliviu/go-todo-api/db"
	"github.com/raduliviu/go-todo-api/store"
)

func setupRouter(h *Handler) *gin.Engine {
	server := gin.Default()

	allowOrigins := os.Getenv("ALLOWED_ORIGINS")

	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{allowOrigins},
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

func main() {
	dsn := os.Getenv("DATABASE_URL")
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
