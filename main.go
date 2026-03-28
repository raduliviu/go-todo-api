package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	server := gin.Default()

	server.GET("/todos", getTodos)
	server.GET("/todos/:id", getTodoByID)
	server.POST("/todos", createTodo)
	server.PATCH("/todos/:id", updateTodoByID)
	server.DELETE("/todos/:id", deleteTodoByID)

	return server
}

func main() {
	router := setupRouter()

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(ginHandler(router))
	} else {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}

}
