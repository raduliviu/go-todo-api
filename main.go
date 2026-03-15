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

func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func postTodo(c *gin.Context) {
	var newTodo Todo

	err := c.BindJSON(&newTodo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

func main() {
	server := gin.Default()

	server.GET("/todos", getTodos)
	server.POST("/todos", postTodo)
	if err := server.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
