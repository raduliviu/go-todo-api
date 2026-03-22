package main

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
