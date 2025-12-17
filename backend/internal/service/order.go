package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, limit, offset int) ([]*model.Order, int, error)
}

type OrderCounterpartyService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error)
}

type OrderProductService interface {
	ChangeStock(ctx context.Context, productID uuid.UUID, amount int, transactionType string) error
}

type OrderService struct {
	repo                OrderRepository
	counterpartyService OrderCounterpartyService
	productService      OrderProductService
	logger              *slog.Logger
}

func NewOrderService(repo OrderRepository, counterpartyService OrderCounterpartyService, productService OrderProductService, logger *slog.Logger) *OrderService {
	return &OrderService{
		repo:                repo,
		counterpartyService: counterpartyService,
		productService:      productService,
		logger:              logger,
	}
}

func (s *OrderService) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	counterparty, err := s.counterpartyService.GetByID(ctx, order.CounterpartyID)
	if err != nil {
		return nil, fmt.Errorf("%w: counterparty not found or service error", err)
	}

	switch order.OrderType {
	case model.OrderInbound:
		if counterparty.Type != model.CounterpartySupplier {
			return nil, fmt.Errorf("%w: Inbound order must be from a Supplier", customErrors.ErrInvalidInput)
		}
		order.Destination = ""
	case model.OrderOutbound:
		if counterparty.Type != model.CounterpartyClient {
			return nil, fmt.Errorf("%w: Outbound order must be for a Client", customErrors.ErrInvalidInput)
		}
		if order.Destination == "" {
			return nil, fmt.Errorf("%w: Destination is required for Outbound order", customErrors.ErrInvalidInput)
		}
	default:
		return nil, fmt.Errorf("%w: Invalid OrderType specified", customErrors.ErrInvalidInput)
	}

	order.ID = uuid.New()

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = model.StatusPending

	if err := s.repo.Create(ctx, order); err != nil {
		s.logger.Error("service: failed to create order in repo", "error", err)
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) Update(ctx context.Context, id uuid.UUID, newStatus string) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	targetStatus := model.OrderStatus(newStatus)

	if order.Status == model.StatusCanceled || order.Status == model.StatusCompleted {
		return nil, fmt.Errorf("%w: cannot change status of a %s order", customErrors.ErrInvalidInput, order.Status)
	}

	if targetStatus == model.StatusCompleted && order.Status != model.StatusCompleted {

		transactionType := ""
		switch order.OrderType {
		case model.OrderInbound:
			transactionType = "income"
		case model.OrderOutbound:
			transactionType = "expense"
		default:
			s.logger.Error("Order type is invalid for stock change", "order_id", id, "type", order.OrderType)
		}

		if transactionType != "" {
			for i, item := range order.Items {
				err := s.productService.ChangeStock(ctx, item.ProductID, item.Quantity, transactionType)
				if err != nil {
					s.logger.Error("failed to update stock, attempting rollback", "order_id", id, "failed_item", item.ProductID)

					rollbackType := "income"
					if transactionType == "income" {
						rollbackType = "expense"
					}

					for j := 0; j < i; j++ {
						rollbackItem := order.Items[j]
						_ = s.productService.ChangeStock(ctx, rollbackItem.ProductID, rollbackItem.Quantity, rollbackType)
					}

					return nil, fmt.Errorf("failed to process order items: %w", err)
				}
			}
			s.logger.Info("Stock updated successfully", "order_id", id)
		}
	}

	order.Status = targetStatus
	order.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, order); err != nil {
		s.logger.Error("CRITICAL: Stock updated but Order Status update failed", "order_id", id, "error", err)
		return nil, fmt.Errorf("failed to update order status (data inconsistency potential): %w", err)
	}

	return order, nil
}

func (s *OrderService) Delete(ctx context.Context, id uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if order.Status != model.StatusCanceled && order.Status != model.StatusPending {
		return fmt.Errorf("%w: only pending or canceled orders can be deleted", customErrors.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *OrderService) GetList(ctx context.Context, page, pageSize int) ([]*model.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	orders, totalCount, err := s.repo.GetList(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error("service: failed to list orders from repo", "error", err)
		return nil, 0, err
	}
	return orders, totalCount, nil
}
