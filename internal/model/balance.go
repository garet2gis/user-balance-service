package model

type BalanceModel struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type ReserveModel struct {
	UserID    string  `json:"user_id"`
	ServiceID string  `json:"service_id"`
	OrderID   string  `json:"order_id"`
	Cost      float64 `json:"cost"`
}

type ReportRow struct {
	ServiceName string `json:"service_name"`
	Cost        string `json:"cost"`
}

type HistoryRow struct {
	OrderID         string `json:"order_id"`
	ServiceName     string `json:"service_name"`
	CreateAt        string `json:"create_at"`
	Amount          int    `json:"amount"`
	TransactionType string `json:"transaction_type"`
}
