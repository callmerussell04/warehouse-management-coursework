package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warehouse-management-system/internal/dto"
	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

type CounterpartyService interface {
	Create(ctx context.Context, name, typeStr, phone, email string) (*model.Counterparty, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error)
	Update(ctx context.Context, id uuid.UUID, name, phone, email string) (*model.Counterparty, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, page, pageSize int) ([]*model.Counterparty, int, error)
}

type CounterpartyHandler struct {
	service CounterpartyService
	logger  *slog.Logger
}

func NewCounterpartyHandler(service CounterpartyService, logger *slog.Logger) *CounterpartyHandler {
	return &CounterpartyHandler{service: service, logger: logger}
}

func (h *CounterpartyHandler) Create(c *gin.Context) {
	var req dto.CreateCounterpartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	counterparty, err := h.service.Create(c.Request.Context(), req.Name, req.Type, req.PhoneNumber, req.Email)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapModelToResponse(counterparty))
}

func (h *CounterpartyHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	counterparty, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, h.mapModelToResponse(counterparty))
}

func (h *CounterpartyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var req dto.UpdateCounterpartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	counterparty, err := h.service.Update(c.Request.Context(), id, req.Name, req.PhoneNumber, req.Email)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, h.mapModelToResponse(counterparty))
}

func (h *CounterpartyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CounterpartyHandler) GetList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, errP := strconv.Atoi(pageStr)
	pageSize, errS := strconv.Atoi(pageSizeStr)

	if errP != nil || errS != nil || page < 1 || pageSize < 1 {
		RespondWithError(c, h.logger, customErrors.NewAppError(customErrors.ErrInvalidInput, "Parameters 'page' and 'pageSize' must be positive integers."))
		return
	}

	list, totalCount, err := h.service.GetList(c.Request.Context(), page, pageSize)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	counterpartyResponses := make([]dto.CounterpartyResponse, len(list))
	for i, counterparty := range list {
		counterpartyResponses[i] = h.mapModelToResponse(counterparty)
	}

	response := dto.PagedCounterparties{
		Paging: dto.Paging{
			Page:  page,
			Size:  pageSize,
			Total: totalCount,
		},
		Items: counterpartyResponses,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CounterpartyHandler) mapModelToResponse(counterparty *model.Counterparty) dto.CounterpartyResponse {
	return dto.CounterpartyResponse{
		ID:          counterparty.ID.String(),
		Name:        counterparty.Name,
		Type:        string(counterparty.Type),
		PhoneNumber: counterparty.PhoneNumber,
		Email:       counterparty.Email,
	}
}
