package main

import (
	"log"
	"net/http"
	"slices"
	"strconv"

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

func getTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, todos[todoIndex])
}

func createTodo(c *gin.Context) {
	var newTodo Todo

	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newTodo.ID = len(todos) + 1

	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

func updateTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	// Get a pointer to the todo in the slice so BindJSON modifies it directly
	updatedTodo := &todos[todoIndex]
	// BindJSON reads the request body and unmarshals it into the struct; requires a pointer
	if err := c.BindJSON(updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve the original ID so the client can't overwrite it
	updatedTodo.ID = id
	c.JSON(http.StatusOK, updatedTodo)
}

func deleteTodoByID(c *gin.Context) {
	// Convert the URL param from string to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Find the index of the todo matching the ID; returns -1 if not found
	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	// Remove the element at todoIndex from the slice
	todos = slices.Delete(todos, todoIndex, todoIndex+1)
	c.Status(http.StatusNoContent)
}

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
