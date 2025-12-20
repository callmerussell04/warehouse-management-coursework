package handler

import (
	"context"
	"log/slog"
	"net/http"

	"warehouse-management-system/internal/dto"
	"warehouse-management-system/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=OrderService --output=../../mocks --outpkg=mocks --with-expecter=true
type OrderService interface {
	Create(ctx context.Context, order *model.Order) (*model.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	Update(ctx context.Context, id uuid.UUID, newStatus string) (*model.Order, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, page, pageSize int) ([]*model.Order, int, error)
}

type OrderHandler struct {
	service OrderService
	logger  *slog.Logger
}

func NewOrderHandler(service OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

func orderModelToResponse(m *model.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID:             m.ID.String(),
		CounterpartyID: m.CounterpartyID.String(),
		OrderType:      string(m.OrderType),
		Status:         string(m.Status),
		OrderDate:      m.OrderDate,
		Destination:    m.Destination,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		Items:          make([]dto.OrderItemResponse, len(m.Items)),
	}
	for i, item := range m.Items {
		resp.Items[i] = dto.OrderItemResponse{
			ProductID: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}
	return resp
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	counterpartyID, err := uuid.Parse(req.CounterpartyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid counterparty UUID format"})
		return
	}

	items := make([]model.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product UUID format"})
			return
		}
		items[i] = model.OrderItem{ProductID: productID, Quantity: itemReq.Quantity}
	}

	destination := ""
	if req.Destination != nil {
		destination = *req.Destination
	}

	newOrder := &model.Order{
		CounterpartyID: counterpartyID,
		OrderType:      model.OrderType(req.OrderType),
		OrderDate:      req.OrderDate,
		Destination:    destination,
		Items:          items,
	}

	order, err := h.service.Create(c.Request.Context(), newOrder)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, orderModelToResponse(order))
}

func (h *OrderHandler) Get(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	order, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, orderModelToResponse(order))
}

func (h *OrderHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	var req dto.UpdateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	order, err := h.service.Update(c.Request.Context(), id, *req.Status)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, orderModelToResponse(order))
}

func (h *OrderHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
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

func (h *OrderHandler) GetList(c *gin.Context) {
	page, pageSize, err := parsePaging(c)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	orders, totalCount, err := h.service.GetList(c.Request.Context(), page, pageSize)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = orderModelToResponse(order)
	}

	response := dto.PagedOrders{
		TotalCount: totalCount,
		PageSize:   pageSize,
		Page:       page,
		Data:       orderResponses,
	}

	c.JSON(http.StatusOK, response)
}
