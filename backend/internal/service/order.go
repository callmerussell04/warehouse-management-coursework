package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"

	"github.com/google/uuid"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=OrderRepository --output=../../mocks --outpkg=mocks --with-expecter=true
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	Transition(ctx context.Context, id uuid.UUID, targetStatus model.OrderStatus) (*model.Order, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, limit, offset int) ([]*model.Order, int, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=OrderCounterpartyService --output=../../mocks --outpkg=mocks --with-expecter=true
type OrderCounterpartyService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error)
}

type OrderService struct {
	repo                OrderRepository
	counterpartyService OrderCounterpartyService
	logger              *slog.Logger
}

func NewOrderService(repo OrderRepository, counterpartyService OrderCounterpartyService, logger *slog.Logger) *OrderService {
	return &OrderService{
		repo:                repo,
		counterpartyService: counterpartyService,
		logger:              logger,
	}
}

func (s *OrderService) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	order.Destination = strings.TrimSpace(order.Destination)
	if order.CounterpartyID == uuid.Nil || utf8.RuneCountInString(order.Destination) > 255 {
		return nil, customErrors.ErrInvalidInput
	}
	if len(order.Items) == 0 || len(order.Items) > 100 {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "order must contain between 1 and 100 items")
	}
	seenProducts := make(map[uuid.UUID]struct{}, len(order.Items))
	for _, item := range order.Items {
		if item.ProductID == uuid.Nil || item.Quantity <= 0 {
			return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "order item must contain a valid product and positive quantity")
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "order contains duplicate products")
		}
		seenProducts[item.ProductID] = struct{}{}
	}
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
	targetStatus := model.OrderStatus(newStatus)
	switch targetStatus {
	case model.StatusProcessing, model.StatusCompleted, model.StatusCanceled:
		return s.repo.Transition(ctx, id, targetStatus)
	default:
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "invalid target order status")
	}
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
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	orders, totalCount, err := s.repo.GetList(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error("service: failed to list orders from repo", "error", err)
		return nil, 0, err
	}
	return orders, totalCount, nil
}
