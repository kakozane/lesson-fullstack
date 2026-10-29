package service

import (
	"lesson06/backend/internal/task/admin/dto"
	"lesson06/backend/internal/task/entity"
	"strings"
	"testing"
	"time"
)

func TestValidateTask(t *testing.T) {
	good := dto.SaveRequest{Version: 1, TaskInput: dto.TaskInput{Title: "任务", Priority: "medium", Status: "todo", DueDate: "2028-02-29"}}
	for _, tc := range []struct {
		name   string
		change func(*dto.SaveRequest)
	}{
		{"empty title", func(p *dto.SaveRequest) { p.Title = "  " }},
		{"long title", func(p *dto.SaveRequest) { p.Title = strings.Repeat("中", 81) }},
		{"long description", func(p *dto.SaveRequest) { p.Description = strings.Repeat("中", 2001) }},
		{"long assignee", func(p *dto.SaveRequest) { p.Assignee = strings.Repeat("中", 31) }},
		{"status", func(p *dto.SaveRequest) { p.Status = "unknown" }},
		{"priority", func(p *dto.SaveRequest) { p.Priority = "urgent" }},
		{"invalid date", func(p *dto.SaveRequest) { p.DueDate = "2026-02-29" }},
		{"date format", func(p *dto.SaveRequest) { p.DueDate = "2026-2-1" }},
		{"missing version", func(p *dto.SaveRequest) { p.Version = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := good
			tc.change(&p)
			if validate(&p) == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	good.Title = "  整理任务  "
	good.Description = "  "
	good.Assignee = "  林悦  "
	if err := validate(&good); err != nil {
		t.Fatal(err)
	}
	if good.Title != "整理任务" || good.Description != "" || good.Assignee != "林悦" {
		t.Fatal("trim failed")
	}
	for _, status := range []string{"todo", "doing", "done"} {
		good.Status = status
		good.DueDate = ""
		if err := validate(&good); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBoardStatistics(t *testing.T) {
	// UTC 16:00 对应北京时间次日零点；今天到期不算逾期，已完成不算逾期。
	now := time.Date(2026, 10, 28, 16, 0, 0, 0, time.UTC)
	items := []entity.Task{
		{Status: "todo", DueDate: "2026-10-28"},
		{Status: "doing", DueDate: "2026-10-28"},
		{Status: "doing", DueDate: "2026-10-29"},
		{Status: "todo", DueDate: ""},
		{Status: "done", DueDate: "2026-10-27"},
		{Status: "todo", DueDate: "2026-10-30"},
	}
	b := buildBoard(items, 29, now)
	if b.Total != 6 || b.Doing != 2 || b.Completed != 1 || b.Overdue != 2 || b.Completion != 17 || b.Version != 29 || len(b.Items) != 6 {
		t.Fatalf("unexpected board: %+v", b)
	}
	before := buildBoard(items, 29, now.Add(-time.Second))
	if before.Overdue != 0 {
		t.Fatalf("before midnight overdue=%d", before.Overdue)
	}
	empty := buildBoard(nil, 30, now)
	if empty.Total != 0 || empty.Completion != 0 || empty.Overdue != 0 || empty.Items == nil {
		t.Fatalf("unexpected empty board: %+v", empty)
	}
	done := buildBoard([]entity.Task{{Status: "done"}}, 31, now)
	if done.Completion != 100 || done.Completed != 1 {
		t.Fatalf("unexpected completed board: %+v", done)
	}
}
