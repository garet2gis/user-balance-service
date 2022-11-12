package dto

import "github.com/jackc/pgx/v5/pgtype"

type ReserveDB struct {
	UserID        string  `json:"user_id"`
	ReservationID string  `json:"reservation_id"`
	ServiceID     string  `json:"service_id"`
	OrderID       string  `json:"order_id"`
	Cost          float64 `json:"cost"`
	Comment       string  `json:"comment"`
	CreatedAt     pgtype.Timestamp
}
