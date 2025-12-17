package dto

type TurnoverReportItem struct {
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	StartBalance int    `json:"start_balance"`
	Income       int    `json:"income"`
	Expense      int    `json:"expense"`
	EndBalance   int    `json:"end_balance"`
}
