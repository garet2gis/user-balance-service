package dto

type BalanceHistory struct {
	UserID     string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	OrderBy    string `json:"order_by" validate:"required,oneof='asc' 'desc'"`
	OrderField string `json:"order_field" validate:"required,oneof='create_date' 'amount'"`
	Limit      int64  `json:"limit,omitempty" validate:"gte=0"`
	Offset     int64  `json:"offset,omitempty" validate:"gte=0"`
} // @name BalanceHistory
