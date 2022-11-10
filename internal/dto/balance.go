package dto

type BalanceRequest struct {
	// Баланс пользователя
	Amount float64 `json:"amount" validate:"required,gte=1"`
	// UUID баланса пользователя
	UserID string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	// UUID баланса пользователя
	Comment string `json:"comment,omitempty"`
} // @name BalanceRequest
