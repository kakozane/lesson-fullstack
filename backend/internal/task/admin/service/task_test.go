package service

import (
	"lesson06/backend/internal/task/admin/dto"
	"strings"
	"testing"
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
