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
	UpdateQuantity(ctx context.Context, id uuid.UUID, delta int) (int, error)
	AddTransaction(ctx context.Context, t *model.InventoryTransaction) error
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

	var delta int
	var tType model.TransactionType

	switch transactionType {
	case "income":
		delta = amount
		tType = model.TransactionIncome
	case "expense":
		delta = -amount
		tType = model.TransactionExpense
	default:
		return customErrors.ErrInvalidInput
	}

	newQty, err := s.repo.UpdateQuantity(ctx, productID, delta)
	if err != nil {
		return err
	}

	t := &model.InventoryTransaction{
		ID:        uuid.New(),
		ProductID: productID,
		Type:      tType,
		Quantity:  amount,
		CreatedAt: time.Now(),
	}

	if err := s.repo.AddTransaction(ctx, t); err != nil {
		s.logger.Error("CRITICAL: failed to log transaction after stock update",
			"product_id", productID, "error", err)
		return err
	}

	s.logger.Info("stock updated", "product_id", productID, "type", transactionType, "new_quantity", newQty)
	return nil
}
