package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	server.GET("/todos", getTodos)
	server.GET("/todos/:id", getTodoByID)
	server.POST("/todos", createTodo)
	server.PATCH("/todos/:id", updateTodoByID)
	server.DELETE("/todos/:id", deleteTodoByID)

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
