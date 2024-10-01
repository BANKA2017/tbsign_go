package _type

type ApiTemplate struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Version string `json:"version"`
}

type IsManagerPreCheckResponse struct {
	IsManager bool   `json:"is_manager"`
	Role      string `json:"role"`
}
