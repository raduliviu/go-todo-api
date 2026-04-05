package store

import (
	"context"
	"database/sql"

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

func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) {
	var todos []Todo
	err := s.db.NewSelect().
		Model(&todos).
		OrderExpr("id ASC").
		Scan(ctx)
	return todos, err
}

func (s *TodoStore) GetByID(ctx context.Context, id int64) (*Todo, error) {
	var todo Todo
	err := s.db.NewSelect().Where("id = ?", id).Model(&todo).Scan(ctx)
	return &todo, err
}

func (s *TodoStore) Create(ctx context.Context, todo *Todo) error {
	_, err := s.db.NewInsert().Model(todo).Returning("id").Exec(ctx)
	return err
}

func (s *TodoStore) Update(ctx context.Context, todo *Todo) error {
	result, err := s.db.NewUpdate().Model(todo).Column("title", "completed").WherePK().Exec(ctx)
	if err != nil {
		return err
	}
	numOfRows, err := result.RowsAffected()
	if err == nil && numOfRows == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (s *TodoStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.NewDelete().Model((*Todo)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	numOfRows, err := result.RowsAffected()
	if err == nil && numOfRows == 0 {
		return sql.ErrNoRows
	}
	return err
}
