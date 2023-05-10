package ginger

type Response struct {
	Success    bool                `json:"success"`
	Error      *ResponseError      `json:"error,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
	Data       interface{}         `json:"data,omitempty"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginationResponse struct {
	Page      int   `json:"page"`
	TotalPage int   `json:"total_page"`
	Size      int   `json:"size"`
	Total     int64 `json:"total"`
	HasNext   bool  `json:"has_next"`
}
