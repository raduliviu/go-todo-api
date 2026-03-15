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

var todos = []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a web server", Completed: false},
	{ID: 3, Title: "Write unit tests", Completed: false},
}

func handleGetTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func main() {
	server := gin.Default()

	server.GET("/todos", handleGetTodos)
	err := server.Run(":8080")
	if err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
