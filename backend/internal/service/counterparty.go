package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=CounterpartyRepository --output=../../mocks --outpkg=mocks --with-expecter=true
type CounterpartyRepository interface {
	Create(ctx context.Context, c *model.Counterparty) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error)
	Update(ctx context.Context, c *model.Counterparty) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, limit, offset int, typeFilter string) ([]*model.Counterparty, int, error)
}

type CounterpartyService struct {
	repo   CounterpartyRepository
	logger *slog.Logger
}

func NewCounterpartyService(repo CounterpartyRepository, logger *slog.Logger) *CounterpartyService {
	return &CounterpartyService{repo: repo, logger: logger}
}

func (s *CounterpartyService) Create(ctx context.Context, name, typeStr, phone, email string) (*model.Counterparty, error) {
	if name == "" {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Name cannot be empty.")
	}

	var cpType model.CounterpartyType
	switch typeStr {
	case "client":
		cpType = model.CounterpartyClient
	case "supplier":
		cpType = model.CounterpartySupplier
	default:
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid counterparty type.")
	}

	c := &model.Counterparty{
		ID:          uuid.New(),
		Name:        name,
		Type:        cpType,
		PhoneNumber: phone,
		Email:       email,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	s.logger.Info("counterparty created", "id", c.ID, "name", c.Name, "type", c.Type)
	return c, nil
}

func (s *CounterpartyService) GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CounterpartyService) Update(ctx context.Context, id uuid.UUID, name, phone, email string) (*model.Counterparty, error) {
	if name == "" {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Name cannot be empty.")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.Name = name
	c.PhoneNumber = phone
	c.Email = email

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	s.logger.Info("counterparty updated", "id", id)
	return c, nil
}

func (s *CounterpartyService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("counterparty deleted", "id", id)
	return nil
}

func (s *CounterpartyService) GetList(ctx context.Context, page, pageSize int, typeFilter string) ([]*model.Counterparty, int, error) {
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

	list, totalCount, err := s.repo.GetList(ctx, limit, offset, typeFilter)
	if err != nil {
		return nil, 0, err
	}

	s.logger.Info("counterparty list retrieved",
		"count", len(list),
		"total", totalCount,
		"page", page,
		"type_filter", typeFilter,
	)
	return list, totalCount, nil
}
