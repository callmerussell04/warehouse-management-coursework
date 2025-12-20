package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warehouse-management-system/internal/dto"
	"warehouse-management-system/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ProductService --output=../../mocks --outpkg=mocks --with-expecter=true
type ProductService interface {
	Create(ctx context.Context, sku, name string) (*model.Product, error)
	GetList(ctx context.Context, page, pageSize int) ([]*model.Product, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	Update(ctx context.Context, id uuid.UUID, sku, name string) (*model.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangeStock(ctx context.Context, productID uuid.UUID, amount int64, transactionType string) error
	GetProductHistory(ctx context.Context, productID uuid.UUID, page, pageSize int, fromStr, toStr string) ([]*model.InventoryTransaction, int, error)
}

type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
}

func NewProductHandler(service ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	product, err := h.service.Create(c.Request.Context(), req.SKU, req.Name)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapModelToResponse(product))
}

func (h *ProductHandler) GetList(c *gin.Context) {
	page, pageSize, err := parsePaging(c)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	products, totalCount, err := h.service.GetList(c.Request.Context(), page, pageSize)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	productResponses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = h.mapModelToResponse(p)
	}

	response := dto.PagedProducts{
		Paging: dto.Paging{
			Page:  page,
			Size:  pageSize,
			Total: totalCount,
		},
		Items: productResponses,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, h.mapModelToResponse(product))
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	product, err := h.service.Update(c.Request.Context(), id, req.SKU, req.Name)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, h.mapModelToResponse(product))
}

func (h *ProductHandler) Delete(c *gin.Context) {
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

func (h *ProductHandler) UpdateStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var req dto.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	err = h.service.ChangeStock(c.Request.Context(), id, req.Quantity, req.Type)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "stock updated"})
}

func (h *ProductHandler) mapModelToResponse(p *model.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:        p.ID.String(),
		SKU:       p.SKU,
		Name:      p.Name,
		Quantity:  p.Quantity,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *ProductHandler) GetHistory(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	page, pageSize, err := parsePaging(c)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	history, total, err := h.service.GetProductHistory(c.Request.Context(), productID, page, pageSize, fromStr, toStr)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	dtos := make([]dto.TransactionResponse, len(history))
	for i, item := range history {
		dtos[i] = dto.TransactionResponse{
			ID:           item.ID.String(),
			ProductID:    item.ProductID.String(),
			ProductName:  item.ProductName,
			Type:         string(item.Type),
			Quantity:     item.Quantity,
			BalanceAfter: item.BalanceAfter,
			CreatedAt:    item.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, dto.PagedTransactions{
		Paging: dto.Paging{
			Page:  page,
			Size:  pageSize,
			Total: total,
		},
		Items: dtos,
	})
}
