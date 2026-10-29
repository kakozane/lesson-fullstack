package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"lesson06/backend/internal/pkg/apperr"
	"lesson06/backend/internal/product/admin/dto"
	"lesson06/backend/internal/product/entity"
	"lesson06/backend/internal/product/repository"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ProductService struct{ repo *repository.ProductRepository }

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}
func Categories() []string { return []string{"数码配件", "办公文具", "生活好物"} }
func validate(p *dto.SaveRequest) error {
	p.Name = strings.TrimSpace(p.Name)
	p.SKU = strings.TrimSpace(p.SKU)
	p.Description = strings.TrimSpace(p.Description)
	bad := func(message string) error { return apperr.New(400, message) }
	if utf8.RuneCountInString(p.Name) < 1 || utf8.RuneCountInString(p.Name) > 80 {
		return bad("商品名称需为 1～80 个字符")
	}
	if len(p.SKU) < 1 || len(p.SKU) > 32 {
		return bad("商品编码需为 1～32 位字母、数字或短横线")
	}
	for _, c := range p.SKU {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return bad("商品编码仅支持字母、数字和短横线")
		}
	}
	valid := false
	for _, c := range Categories() {
		if p.Category == c {
			valid = true
		}
	}
	if !valid {
		return bad("请选择有效分类")
	}
	if p.PriceCents < 1 || p.PriceCents > 99999999 {
		return bad("价格范围为 0.01～999999.99 元")
	}
	if p.Stock < 0 || p.Stock > 999999 {
		return bad("库存范围为 0～999999")
	}
	if p.Status == nil || (*p.Status != 0 && *p.Status != 1) {
		return bad("商品状态必须为 0 或 1")
	}
	if utf8.RuneCountInString(p.Description) > 500 {
		return bad("商品描述不能超过 500 个字符")
	}
	return nil
}
func storageError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(404, "商品不存在或已删除")
	}
	var duplicate *mysql.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return apperr.New(409, "商品编码已存在")
	}
	return err
}
func (s *ProductService) List(ctx context.Context, q dto.ListQuery) (*dto.PageResult, error) {
	if q.Page < 1 || q.Page > 1000000 || q.PageSize < 1 || q.PageSize > 100 {
		return nil, apperr.New(400, "分页参数不合法")
	}
	q.Keyword = strings.TrimSpace(q.Keyword)
	if utf8.RuneCountInString(q.Keyword) > 80 {
		return nil, apperr.New(400, "搜索词不能超过 80 个字符")
	}
	if q.Status != "" && q.Status != "0" && q.Status != "1" {
		return nil, apperr.New(400, "状态参数不合法")
	}
	items, total, page, err := s.repo.ListPaged(ctx, q.Keyword, q.Category, q.Status, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}
	result := &dto.PageResult{Items: []dto.Product{}, Total: total, Page: page, PageSize: q.PageSize}
	for _, p := range items {
		result.Items = append(result.Items, dto.Product{ID: p.ID, Name: p.Name, SKU: p.SKU, Category: p.Category, PriceCents: p.PriceCents, Stock: p.Stock, Status: p.Status, Description: p.Description, UpdatedAt: p.UpdatedAt.In(time.FixedZone("UTC+8", 8*60*60)).Format("2006-01-02 15:04:05")})
	}
	return result, nil
}
func toEntity(p dto.SaveRequest) *entity.Product {
	return &entity.Product{Name: p.Name, SKU: p.SKU, Category: p.Category, PriceCents: p.PriceCents, Stock: p.Stock, Status: *p.Status, Description: p.Description}
}
func (s *ProductService) Create(ctx context.Context, p dto.SaveRequest) error {
	if err := validate(&p); err != nil {
		return err
	}
	return storageError(s.repo.Create(ctx, toEntity(p)))
}
func (s *ProductService) Update(ctx context.Context, id int64, p dto.SaveRequest) error {
	if err := validate(&p); err != nil {
		return err
	}
	return storageError(s.repo.Update(ctx, id, toEntity(p)))
}
func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return storageError(s.repo.Delete(ctx, id))
}
func (s *ProductService) BatchStatus(ctx context.Context, p dto.BatchStatusRequest) error {
	if len(p.IDs) < 1 || len(p.IDs) > 100 || p.Status == nil || (*p.Status != 0 && *p.Status != 1) {
		return apperr.New(400, "请选择 1～100 个商品及有效状态")
	}
	seen := map[int64]bool{}
	for _, id := range p.IDs {
		if id < 1 || seen[id] {
			return apperr.New(400, "商品 ID 不合法或重复")
		}
		seen[id] = true
	}
	err := s.repo.BatchStatus(ctx, p.IDs, *p.Status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.New(409, "部分商品已删除，请刷新后重试")
	}
	return err
}
