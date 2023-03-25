package ginger

import "github.com/ginger-go/sql"

type Response struct {
	Success    bool            `json:"success"`
	Error      *ResponseError  `json:"error,omitempty"`
	Pagination *sql.Pagination `json:"pagination,omitempty"`
	Data       interface{}     `json:"data,omitempty"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
