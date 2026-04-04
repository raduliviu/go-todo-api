package store

import (
	"context"

	"github.com/uptrace/bun"
)

type Todo struct {
	bun.BaseModel `bun:"table:todos,alias:t"`
	ID            int64  `bun:"id,pk,autoincrement" json:"id"`
	Title         string `bun:"title,notnull" json:"title"`
	Completed     bool   `bun:"completed,notnull" json:"completed"`
}

type TodoStorer interface {
	GetAll(ctx context.Context) ([]Todo, error)
	GetByID(ctx context.Context, id int64) (*Todo, error)
	Create(ctx context.Context, todo *Todo) error
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id int64) error
}

type TodoStore struct {
	db *bun.DB
}

func NewTodoStore(db *bun.DB) *TodoStore {
	return &TodoStore{db: db}
}
