package entity

import "time"

type Task struct {
	ID          int64 `gorm:"primaryKey"`
	Title       string
	Description string
	Assignee    string
	Priority    string
	Status      string
	DueDate     string
	Position    int
	UpdatedAt   time.Time
}

func (Task) TableName() string { return "tasks" }

type BoardMeta struct {
	ID      int
	Version int64
}

func (BoardMeta) TableName() string { return "task_board_meta" }
