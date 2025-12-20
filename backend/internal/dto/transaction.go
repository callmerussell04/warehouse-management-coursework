package dto

type TransactionResponse struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Type         string `json:"type"`
	Quantity     int64  `json:"quantity"`
	BalanceAfter int64  `json:"balance_after"`
	CreatedAt    string `json:"created_at"`
}

type PagedTransactions struct {
	Paging Paging                `json:"paging"`
	Items  []TransactionResponse `json:"items"`
}
