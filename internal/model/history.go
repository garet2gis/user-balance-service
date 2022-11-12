package model

type HistoryRow struct {
	// UUID заказа
	OrderID string `json:"order_id,omitempty"`
	// Название услуги
	ServiceName string `json:"service_name,omitempty"`
	// UUID отправителя
	UserIDFrom string `json:"user_id_from,omitempty"`
	// UUID получателя
	UserIDTo string `json:"user_id_to,omitempty"`
	// Время создания
	CreateAt string `json:"create_at"`
	// Сумма
	Amount float64 `json:"amount"`
	// Тип транзакции
	TransactionType string `json:"transaction_type"`
	// Комментарий
	Comment string `json:"comment"`
} // @name HistoryRow
