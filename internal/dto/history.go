package dto

type BalanceHistory struct {
	// UUID баланса пользователя
	UserID string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	// Порядок сортировки
	OrderBy string `json:"order_by" default:"desc" validate:"required,oneof='desc' 'asc'"`
	// Поле для сортировки
	OrderField string `json:"order_field" default:"create_date" validate:"required,oneof='create_date' 'amount'"`
	Limit      int64  `json:"limit,omitempty" validate:"gte=0"`
	Offset     int64  `json:"offset,omitempty" validate:"gte=0"`
} // @name BalanceHistory
