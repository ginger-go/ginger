package ginger

import (
	"math"
	"net/http/httptest"

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

type MockContextParams[T any] struct {
	Request   *T
	Page      *sql.Pagination
	Sort      *sql.Sort
	Method    string
	Path      string
	ClientIP  string
	UserAgent string
	Headers   map[string]string
}

func NewMockContext[T any](param MockContextParams[T]) *Context[T] {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	if param.Method == "" {
		param.Method = "GET"
	}
	if param.Path == "" {
		param.Path = "/"
	}
	ctx.Request = httptest.NewRequest(param.Method, param.Path, nil)
	if param.ClientIP != "" {
		ctx.Request.RemoteAddr = param.ClientIP
	}

	if param.UserAgent != "" {
		ctx.Request.Header.Set("User-Agent", param.UserAgent)
	} else {
		ctx.Request.Header.Set("User-Agent", "GingerMockContext")
	}
	if param.Headers != nil {
		for k, v := range param.Headers {
			ctx.Request.Header.Set(k, v)
		}
	}

	return &Context[T]{
		GinContext: ctx,
		Request:    param.Request,
		Page:       param.Page,
		Sort:       param.Sort,
	}
}

func (ctx *Context[T]) ClientIP() string {
	return ctx.GinContext.ClientIP()
}

func (ctx *Context[T]) UserAgent() string {
	return ctx.GinContext.Request.UserAgent()
}

func (ctx *Context[T]) OK(data interface{}, page ...*sql.Pagination) {
	var p *PaginationResponse
	if len(page) > 0 && page[0] != nil {
		totalPage := int(math.Ceil(float64(page[0].Total) / float64(page[0].Size)))
		hasNext := page[0].Page < totalPage
		p = &PaginationResponse{
			Page:      page[0].Page,
			TotalPage: totalPage,
			Size:      page[0].Size,
			Total:     page[0].Total,
			HasNext:   hasNext,
		}
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

func (ctx *Context[T]) ErrorWithStatus(status int, err Error) {
	resp := &Response{
		Success: false,
		Error: &ResponseError{
			Code:    err.Code(),
			Message: err.Error(),
		},
	}
	ctx.Response = resp // for testing
	ctx.GinContext.JSON(status, resp)
}
