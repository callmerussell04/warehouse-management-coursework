package dto

type CreateProductRequest struct {
	SKU  string `json:"sku" binding:"required,min=1,max=128"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type UpdateProductRequest struct {
	SKU  string `json:"sku" binding:"required,min=1,max=128"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type UpdateStockRequest struct {
	Quantity int64  `json:"quantity" binding:"required,min=1"`
	Type     string `json:"type" binding:"required,oneof=income expense"`
}

type ProductResponse struct {
	ID        string `json:"id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Quantity  int64  `json:"quantity"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PagedProducts struct {
	Paging Paging            `json:"paging"`
	Items  []ProductResponse `json:"items"`
}
