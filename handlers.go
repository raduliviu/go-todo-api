package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raduliviu/go-todo-api/store"
)

type Handler struct {
	store store.TodoStorer
}

func NewHandler(store store.TodoStorer) *Handler {
	return &Handler{store: store}
}

func errorResponse(msg string) gin.H {
	return gin.H{"error": msg}
}

func (h *Handler) getTodos(c *gin.Context) {
	todos, err := h.store.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to fetch todos"))
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (h *Handler) getTodoByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	todo, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to fetch todo"))
		return
	}

	c.JSON(http.StatusOK, todo)
}

func (h *Handler) createTodo(c *gin.Context) {
	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	todo := &store.Todo{
		Title:     req.Title,
		Completed: req.Completed,
	}

	if err := h.store.Create(c.Request.Context(), todo); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to create todo"))
		return
	}

	c.JSON(http.StatusCreated, todo)
}

func (h *Handler) updateTodoByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	todo, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to fetch todo"))
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	if err := h.store.Update(c.Request.Context(), todo); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to update todo"))
		return
	}

	c.JSON(http.StatusOK, todo)
}

func (h *Handler) deleteTodoByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("Invalid ID"))
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("Todo not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to delete todo"))
		return
	}

	c.Status(http.StatusNoContent)
}
