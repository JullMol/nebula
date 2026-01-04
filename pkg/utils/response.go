package utils

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(data interface{}, msg string) APIResponse {
	return APIResponse{
		Status:  "success",
		Message: msg,
		Data:    data,
	}
}

func ErrorResponse(err string) APIResponse {
	return APIResponse{
		Status: "error",
		Error:  err,
	}
}