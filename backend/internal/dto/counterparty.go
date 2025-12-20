package dto

type PagedCounterparties struct {
	Paging Paging                 `json:"paging"`
	Items  []CounterpartyResponse `json:"items"`
}

type CreateCounterpartyRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Type        string `json:"type" binding:"required,oneof=client supplier"`
	PhoneNumber string `json:"phone_number" binding:"max=50"`
	Email       string `json:"email" binding:"omitempty,email,max=254"`
}

type UpdateCounterpartyRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	PhoneNumber string `json:"phone_number" binding:"max=50"`
	Email       string `json:"email" binding:"omitempty,email,max=254"`
}

type CounterpartyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
