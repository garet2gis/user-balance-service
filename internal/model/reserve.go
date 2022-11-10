package model

type Reserve struct {
	UserID    string  `json:"user_id"`
	ServiceID string  `json:"service_id"`
	OrderID   string  `json:"order_id"`
	Cost      float64 `json:"cost"`
	Comment   string  `json:"comment"`
}
