package model

type Balance struct {
	// Баланс пользователя
	Balance float64 `json:"balance" validate:"required"`
	// UUID баланса пользователя
	UserID string `json:"user_id" validate:"required"`
} // @name Balance
