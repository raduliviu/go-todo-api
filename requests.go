package main

type CreateTodoRequest struct {
	Title     string `json:"title" binding:"required,min=1,max=255"`
	Completed bool   `json:"completed"`
}

type UpdateTodoRequest struct {
	Title     *string `json:"title" binding:"omitempty,min=1,max=255"`
	Completed *bool   `json:"completed"`
}
