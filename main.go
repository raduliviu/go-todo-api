package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

func setupRouter(h *Handler) *gin.Engine {
	server := gin.Default()

	server.GET("/todos", h.getTodos)
	server.GET("/todos/:id", h.getTodoByID)
	server.POST("/todos", h.createTodo)
	server.PATCH("/todos/:id", h.updateTodoByID)
	server.DELETE("/todos/:id", h.deleteTodoByID)

	return server
}

func main() {
	h := NewHandler(nil)
	router := setupRouter(h)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(ginHandler(router))
	} else {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}

}
