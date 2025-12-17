package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	GetList(ctx context.Context, limit, offset int) ([]*model.Product, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStock(ctx context.Context, productID uuid.UUID, amount int, transType model.TransactionType) error
	GetProductHistory(ctx context.Context, productID uuid.UUID, limit, offset int, from, to time.Time) ([]*model.InventoryTransaction, int, error)
}

type ProductService struct {
	repo   ProductRepository
	logger *slog.Logger
}

func NewProductService(repo ProductRepository, logger *slog.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ProductService) Create(ctx context.Context, sku, name string) (*model.Product, error) {
	if sku == "" || name == "" {
		return nil, customErrors.ErrInvalidInput
	}

	p := &model.Product{
		ID:        uuid.New(),
		SKU:       sku,
		Name:      name,
		Quantity:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	s.logger.Info("product created", "id", p.ID, "sku", p.SKU)
	return p, nil
}

func (s *ProductService) GetList(ctx context.Context, page, pageSize int) ([]*model.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	products, totalCount, err := s.repo.GetList(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	s.logger.Info("product list retrieved", "count", len(products), "total", totalCount, "page", page)
	return products, totalCount, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, sku, name string) (*model.Product, error) {
	if sku == "" || name == "" {
		return nil, customErrors.ErrInvalidInput
	}

	p := &model.Product{
		ID:   id,
		SKU:  sku,
		Name: name,
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	s.logger.Info("product updated", "id", id)
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("product deleted", "id", id)
	return nil
}

func (s *ProductService) ChangeStock(ctx context.Context, productID uuid.UUID, amount int, transactionType string) error {
	if amount <= 0 {
		return customErrors.ErrInvalidInput
	}

	var tType model.TransactionType
	switch transactionType {
	case "income":
		tType = model.TransactionIncome
	case "expense":
		tType = model.TransactionExpense
	default:
		return customErrors.ErrInvalidInput
	}

	if err := s.repo.UpdateStock(ctx, productID, amount, tType); err != nil {
		s.logger.Error("failed to update stock", "product_id", productID, "error", err)
		return err
	}

	return nil
}

func (s *ProductService) GetProductHistory(ctx context.Context, productID uuid.UUID, page, pageSize int, fromStr, toStr string) ([]*model.InventoryTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.DateOnly, fromStr)
		if err != nil {
			from, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				return nil, 0, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid 'from' format")
			}
		}
	}
	if toStr != "" {
		to, err = time.Parse(time.DateOnly, toStr)
		if err != nil {
			to, err = time.Parse(time.RFC3339, toStr)
			if err != nil {
				return nil, 0, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid 'to' format")
			}
		}
		if to.Hour() == 0 && to.Minute() == 0 && to.Second() == 0 {
			to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	return s.repo.GetProductHistory(ctx, productID, limit, offset, from, to)
}
