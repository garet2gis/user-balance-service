package model

type HistoryRow struct {
	OrderID         string `json:"order_id"`
	ServiceName     string `json:"service_name"`
	CreateAt        string `json:"create_at"`
	Amount          int    `json:"amount"`
	TransactionType string `json:"transaction_type"`
}
