package main

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

func errorResponse(msg string) gin.H {
	return gin.H{"error": msg}
}

func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func getTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
		return
	}

	c.JSON(http.StatusOK, todos[todoIndex])
}

func createTodo(c *gin.Context) {
	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	newTodo := Todo{
		ID:        len(todos) + 1,
		Title:     req.Title,
		Completed: req.Completed,
	}

	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

func updateTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if req.Title != nil {
		todos[todoIndex].Title = *req.Title
	}
	if req.Completed != nil {
		todos[todoIndex].Completed = *req.Completed
	}

	c.JSON(http.StatusOK, todos[todoIndex])
}

func deleteTodoByID(c *gin.Context) {
	// Convert the URL param from string to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	// Find the index of the todo matching the ID; returns -1 if not found
	todoIndex := slices.IndexFunc(todos, func(todo Todo) bool {
		return todo.ID == id
	})

	if todoIndex == -1 {
		c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
		return
	}

	// Remove the element at todoIndex from the slice
	todos = slices.Delete(todos, todoIndex, todoIndex+1)
	c.Status(http.StatusNoContent)
}
