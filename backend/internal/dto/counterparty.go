package dto

type PagedCounterparties struct {
	Paging Paging                 `json:"paging"`
	Items  []CounterpartyResponse `json:"items"`
}

type CreateCounterpartyRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=client supplier"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type UpdateCounterpartyRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type CounterpartyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
