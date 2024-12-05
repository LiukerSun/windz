package response

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message" example:"operation successful"`
}
