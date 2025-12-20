package dto

type TurnoverReportItem struct {
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	StartBalance int64  `json:"start_balance"`
	Income       int64  `json:"income"`
	Expense      int64  `json:"expense"`
	EndBalance   int64  `json:"end_balance"`
}
