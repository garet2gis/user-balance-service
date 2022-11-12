package dto

type ReportRequest struct {
	// Баланс пользователя
	Year int `json:"year" example:"2022" validate:"required,gte=2000"`
	// UUID баланса пользователя
	Month int `json:"month"  example:"11" validate:"required,gte=1,lte=12"`
} // @name ReportRequest

type ReportResponse struct {
	// Ссылка на скачивание файла
	FileURL string `json:"file_url" validate:"required,url"`
} // @name ReportResponse
