package model

type DepositType string

const (
	Replenish DepositType = "replenish"
	Reduce                = "reduce"
)

type Balance struct {
	// Баланс пользователя
	Balance float64 `json:"balance" validate:"required"`
	// UUID баланса пользователя
	UserID string `json:"user_id" validate:"required"`
} // @name Balance
