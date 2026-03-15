package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func handleGetTodos(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":        1,
		"title":     "Learn Go",
		"completed": false,
	})
}

func main() {
	server := gin.Default()

	server.GET("/todos", handleGetTodos)
	err := server.Run(":8080")
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
