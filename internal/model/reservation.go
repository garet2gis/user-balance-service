package model

type ReservationStatus string

const (
	Confirm ReservationStatus = "confirm"
	Cancel                    = "cancel"
)

type Reservation struct {
	// UUID баланса пользователя
	UserID string `json:"user_id" example:"7a13445c-d6df-4111-abc0-abb12f610069" validate:"required,uuid"`
	// UUID сервиса
	ServiceID string `json:"service_id" example:"34e16535-480c-43f8-95a9-b7a503499af0" validate:"required,uuid"`
	// UUID заказа
	OrderID string `json:"order_id" example:"983e8792-6736-41bd-9f1a-7c67f8501645" validate:"required,uuid"`
	// Стоимость услуги
	Cost float64 `json:"cost" validate:"gt=0,required"`
	// Дополнительный комментарий
	Comment string `json:"comment,omitempty"`
} // @name Reservation
