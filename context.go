package ginger

import (
	"github.com/gin-gonic/gin"
	"github.com/ginger-go/sql"
)

type Context[T any] struct {
	GinContext *gin.Context
	Request    *T
	Page       *sql.Pagination
	Sort       *sql.Sort
	Response   interface{}
}

func (ctx *Context[T]) ClientIP() string {
	return ctx.GinContext.ClientIP()
}

func (ctx *Context[T]) UserAgent() string {
	return ctx.GinContext.Request.UserAgent()
}

func (ctx *Context[T]) OK(data interface{}, page ...*sql.Pagination) {
	var p *sql.Pagination
	if len(page) > 0 {
		p = page[0]
	}
	resp := &Response{
		Success:    true,
		Data:       data,
		Pagination: p,
	}
	ctx.Response = resp // for testing
	ctx.GinContext.JSON(200, resp)
}

func (ctx *Context[T]) Error(err Error) {
	resp := &Response{
		Success: false,
		Error: &ResponseError{
			Code:    err.Code(),
			Message: err.Error(),
		},
	}
	ctx.Response = resp // for testing
	ctx.GinContext.JSON(200, resp)
}
