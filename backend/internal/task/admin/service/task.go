package service

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"lesson06/backend/internal/pkg/apperr"
	"lesson06/backend/internal/task/admin/dto"
	"lesson06/backend/internal/task/entity"
	"lesson06/backend/internal/task/repository"
	"strings"
	"time"
	"unicode/utf8"
)

type TaskService struct{ repo *repository.TaskRepository }

func New(repo *repository.TaskRepository) *TaskService { return &TaskService{repo: repo} }
func validStatus(status string) bool {
	return status == "todo" || status == "doing" || status == "done"
}
func validate(p *dto.SaveRequest) error {
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.Assignee = strings.TrimSpace(p.Assignee)
	if p.Version < 1 {
		return apperr.New(400, "看板版本不合法")
	}
	if n := utf8.RuneCountInString(p.Title); n < 1 || n > 80 {
		return apperr.New(400, "任务标题需为 1～80 个字符")
	}
	if utf8.RuneCountInString(p.Description) > 2000 || utf8.RuneCountInString(p.Assignee) > 30 {
		return apperr.New(400, "描述最多 2000 字，负责人最多 30 字")
	}
	if !validStatus(p.Status) {
		return apperr.New(400, "任务状态不合法")
	}
	if p.Priority != "low" && p.Priority != "medium" && p.Priority != "high" {
		return apperr.New(400, "任务优先级不合法")
	}
	if p.DueDate != "" {
		if _, err := time.Parse("2006-01-02", p.DueDate); err != nil {
			return apperr.New(400, "截止日期格式应为 YYYY-MM-DD")
		}
	}
	return nil
}
func result(items []entity.Task, version int64, err error) (*dto.Board, error) {
	if errors.Is(err, repository.ErrConflict) {
		return nil, apperr.New(409, "看板已被其他操作更新，请刷新后重试")
	}
	if errors.Is(err, repository.ErrLimit) {
		return nil, apperr.New(409, "本课堂看板最多支持 200 条任务")
	}
	if errors.Is(err, repository.ErrAnchor) {
		return nil, apperr.New(400, "目标位置不合法，请刷新后重试")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.New(404, "任务不存在或已删除")
	}
	if err != nil {
		return nil, err
	}
	board := &dto.Board{Items: []dto.Task{}, Version: version}
	for _, t := range items {
		board.Items = append(board.Items, dto.Task{ID: t.ID, TaskInput: dto.TaskInput{Title: t.Title, Description: t.Description, Assignee: t.Assignee, Priority: t.Priority, Status: t.Status, DueDate: t.DueDate}, UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05")})
	}
	return board, nil
}
func (s *TaskService) List(ctx context.Context) (*dto.Board, error) { return result(s.repo.List(ctx)) }
func (s *TaskService) Save(ctx context.Context, id int64, p dto.SaveRequest) (*dto.Board, error) {
	if err := validate(&p); err != nil {
		return nil, err
	}
	return result(s.repo.Save(ctx, p.Version, id, &entity.Task{Title: p.Title, Description: p.Description, Assignee: p.Assignee, Priority: p.Priority, Status: p.Status, DueDate: p.DueDate}))
}
func (s *TaskService) Delete(ctx context.Context, id, version int64) (*dto.Board, error) {
	if version < 1 {
		return nil, apperr.New(400, "看板版本不合法")
	}
	return result(s.repo.Delete(ctx, version, id))
}
func (s *TaskService) Move(ctx context.Context, id int64, p dto.MoveRequest) (*dto.Board, error) {
	if p.Version < 1 || !validStatus(p.Status) || p.BeforeID < 0 || p.BeforeID == id {
		return nil, apperr.New(400, "移动参数不合法")
	}
	return result(s.repo.Move(ctx, p.Version, id, p.Status, p.BeforeID))
}
