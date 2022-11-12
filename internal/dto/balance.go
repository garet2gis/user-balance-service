package dto

type BalanceChangeRequest struct {
	// Баланс пользователя
	Amount float64 `json:"amount" validate:"gt=0,required"`
	// UUID баланса пользователя
	UserID string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	// Коментарий
	Comment string `json:"comment,omitempty"`
} // @name BalanceChangeRequest

type BalanceGetRequest struct {
	UserID string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
} // @name BalanceGetRequest

type TransferRequest struct {
	// Списание
	Amount float64 `json:"amount" validate:"gt=0,required"`
	// UUID баланса отправителя
	UserIDFrom string `json:"user_id_from"  example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	// UUID баланса получателя
	UserIDTo string `json:"user_id_to"  example:"7a13445c-d6df-4111-abc0-abb12f610068" validate:"required,uuid,necsfield=UserIDFrom"`
	// Коментарий
	Comment string `json:"comment,omitempty"`
} // @name TransferRequest
