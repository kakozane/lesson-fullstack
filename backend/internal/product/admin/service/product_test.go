package service

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"lesson06/backend/internal/pkg/apperr"
	"strings"
	"testing"

	"lesson06/backend/internal/product/admin/dto"
)

func validProduct() dto.SaveRequest {
	s := 0
	return dto.SaveRequest{Name: "测试", SKU: "TEST-01", Category: Categories()[0], PriceCents: 1, Stock: 0, Status: &s}
}
func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*dto.SaveRequest)
		valid  bool
	}{
		{"zero stock and offline", func(p *dto.SaveRequest) {}, true},
		{"blank name", func(p *dto.SaveRequest) { p.Name = "  " }, false},
		{"bad sku", func(p *dto.SaveRequest) { p.SKU = "a b" }, false},
		{"missing status", func(p *dto.SaveRequest) { p.Status = nil }, false},
		{"negative price", func(p *dto.SaveRequest) { p.PriceCents = -1 }, false},
		{"unknown category", func(p *dto.SaveRequest) { p.Category = "其他" }, false},
		{"long description", func(p *dto.SaveRequest) { p.Description = strings.Repeat("中", 501) }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := validProduct()
			tc.mutate(&p)
			got := validate(&p)
			if (got == nil) != tc.valid {
				t.Fatalf("validation = %v", got)
			}
		})
	}
}

func TestInvalidQueries(t *testing.T) {
	s := NewProductService(nil)
	for _, q := range []dto.ListQuery{{Page: 0, PageSize: 10}, {Page: 1, PageSize: 101}, {Page: 1, PageSize: 10, Status: "2"}, {Page: 1, PageSize: 10, Keyword: strings.Repeat("中", 81)}} {
		if _, err := s.List(context.Background(), q); err == nil {
			t.Fatalf("accepted query: %+v", q)
		}
	}
}
func TestInvalidBatch(t *testing.T) {
	zero := 0
	s := NewProductService(nil)
	for _, req := range []dto.BatchStatusRequest{{IDs: []int64{1, 1}, Status: &zero}, {IDs: []int64{}, Status: &zero}, {IDs: []int64{1}}, {IDs: []int64{-1}, Status: &zero}} {
		if err := s.BatchStatus(context.Background(), req); err == nil {
			t.Fatalf("accepted batch: %+v", req)
		}
	}
}

func TestStorageBusinessErrorsReturn400(t *testing.T) {
	for _, source := range []error{gorm.ErrRecordNotFound, &mysql.MySQLError{Number: 1062}} {
		var business *apperr.Error
		if err := storageError(source); !errors.As(err, &business) || business.Code != 400 {
			t.Fatalf("expected business code 400, got %v", err)
		}
	}
	unexpected := errors.New("database unavailable")
	if storageError(unexpected) != unexpected {
		t.Fatal("unexpected errors must remain server errors")
	}
}
