package repository

import (
	"context"
	"database/sql"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lesson06/backend/internal/task/entity"
)

var ErrConflict = errors.New("board version conflict")
var ErrLimit = errors.New("board task limit")
var ErrAnchor = errors.New("invalid destination anchor")

type TaskRepository struct{ db *gorm.DB }

func New(db *gorm.DB) *TaskRepository { return &TaskRepository{db: db} }
func snapshot(tx *gorm.DB) ([]entity.Task, int64, error) {
	var meta entity.BoardMeta
	if err := tx.First(&meta, 1).Error; err != nil {
		return nil, 0, err
	}
	items := []entity.Task{}
	err := tx.Order("status, position, id").Find(&items).Error
	return items, meta.Version, err
}
func (r *TaskRepository) List(ctx context.Context) (items []entity.Task, version int64, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items, version, err = snapshot(tx)
		return err
	}, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	return
}

// 全看板共用版本和行锁，让排序与增删改在同一事务内串行，避免旧页面覆盖新数据。
func (r *TaskRepository) mutate(ctx context.Context, expected int64, apply func(*gorm.DB) error) (items []entity.Task, version int64, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meta entity.BoardMeta
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&meta, 1).Error; e != nil {
			return e
		}
		if meta.Version != expected {
			return ErrConflict
		}
		if e := apply(tx); e != nil {
			return e
		}
		if e := tx.Model(&meta).Update("version", meta.Version+1).Error; e != nil {
			return e
		}
		var e error
		items, version, e = snapshot(tx)
		return e
	})
	return
}
func tail(tx *gorm.DB, status string) (int, error) {
	var n int
	err := tx.Model(&entity.Task{}).Where("status = ?", status).Select("COALESCE(MAX(position), -1) + 1").Scan(&n).Error
	return n, err
}
func (r *TaskRepository) Save(ctx context.Context, version, id int64, item *entity.Task) ([]entity.Task, int64, error) {
	return r.mutate(ctx, version, func(tx *gorm.DB) error {
		if id == 0 {
			var count int64
			if e := tx.Model(&entity.Task{}).Count(&count).Error; e != nil {
				return e
			}
			// shortcut: 单看板最多 200 条；任务量增长时改为按项目与列分页。
			if count >= 200 {
				return ErrLimit
			}
			position, e := tail(tx, item.Status)
			if e != nil {
				return e
			}
			item.Position = position
			return tx.Create(item).Error
		}
		var old entity.Task
		if e := tx.First(&old, id).Error; e != nil {
			return e
		}
		position := old.Position
		if old.Status != item.Status {
			var e error
			position, e = tail(tx, item.Status)
			if e != nil {
				return e
			}
		}
		return tx.Model(&old).Updates(map[string]any{"title": item.Title, "description": item.Description, "assignee": item.Assignee, "priority": item.Priority, "status": item.Status, "due_date": item.DueDate, "position": position, "updated_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error
	})
}
func (r *TaskRepository) Delete(ctx context.Context, version, id int64) ([]entity.Task, int64, error) {
	return r.mutate(ctx, version, func(tx *gorm.DB) error {
		result := tx.Delete(&entity.Task{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
func (r *TaskRepository) Move(ctx context.Context, version, id int64, status string, beforeID int64) ([]entity.Task, int64, error) {
	return r.mutate(ctx, version, func(tx *gorm.DB) error {
		var item entity.Task
		if e := tx.First(&item, id).Error; e != nil {
			return e
		}
		target := []entity.Task{}
		if e := tx.Where("status = ? AND id <> ?", status, id).Order("position, id").Find(&target).Error; e != nil {
			return e
		}
		at := len(target)
		if beforeID != 0 {
			at = -1
			for i, t := range target {
				if t.ID == beforeID {
					at = i
					break
				}
			}
			if at < 0 {
				return ErrAnchor
			}
		}
		ids := make([]int64, 0, len(target)+1)
		for i, t := range target {
			if i == at {
				ids = append(ids, id)
			}
			ids = append(ids, t.ID)
		}
		if at == len(target) {
			ids = append(ids, id)
		}
		for i, taskID := range ids {
			values := map[string]any{"position": i, "updated_at": gorm.Expr("updated_at")}
			if taskID == id {
				values["status"] = status
				values["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")
			}
			if e := tx.Model(&entity.Task{}).Where("id = ?", taskID).UpdateColumns(values).Error; e != nil {
				return e
			}
		}
		return nil
	})
}
