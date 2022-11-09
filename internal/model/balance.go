package model

import "github.com/jackc/pgx/v5/pgtype"

type BalanceDTO struct {
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment"`
}

type ReserveDTO struct {
	UserID    string  `json:"user_id"`
	ServiceID string  `json:"service_id"`
	OrderID   string  `json:"order_id"`
	Cost      float64 `json:"cost"`
	Comment   string  `json:"comment"`
}

type ReserveModel struct {
	UserID        string  `json:"user_id"`
	ReservationID string  `json:"reservation_id"`
	ServiceID     string  `json:"service_id"`
	OrderID       string  `json:"order_id"`
	Cost          float64 `json:"cost"`
	Comment       string  `json:"comment"`
	CreatedAt     pgtype.Timestamp
}
