package ginger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginger-go/ginger/typescript"
	"github.com/ginger-go/sql"
	"github.com/gorilla/websocket"
)

type Engine struct {
	GinEngine      *gin.Engine
	ModelConverter *typescript.ModelConverter
	ApiConverter   *typescript.ApiConverter
}

func NewEngine() *Engine {
	return &Engine{
		GinEngine:      gin.Default(),
		ModelConverter: typescript.NewModelConverter(),
		ApiConverter:   typescript.NewApiConverter(),
	}
}

func (e *Engine) Run(addr string) error {
	return e.GinEngine.Run(addr)
}

func (e *Engine) Use(middleware ...gin.HandlerFunc) {
	e.GinEngine.Use(middleware...)
}

func GET[T any](engine *Engine, route string, handler Handler[T], middleware ...gin.HandlerFunc) {
	setup := handler()
	engine.ModelConverter.Add(new(T))
	engine.ModelConverter.Add(setup.Response)
	engine.ApiConverter.Add("GET", route, new(T), setup.Response, handler, setup.Pagination, setup.Sort)
	engine.GinEngine.GET(route, joinMiddlewareAndService(newGinServiceHandler(handler), middleware...)...)
}

func POST[T any](engine *Engine, route string, handler Handler[T], middleware ...gin.HandlerFunc) {
	setup := handler()
	engine.ModelConverter.Add(new(T))
	engine.ModelConverter.Add(setup.Response)
	engine.ApiConverter.Add("POST", route, new(T), setup.Response, handler, setup.Pagination, setup.Sort)
	engine.GinEngine.POST(route, joinMiddlewareAndService(newGinServiceHandler(handler), middleware...)...)
}

func PUT[T any](engine *Engine, route string, handler Handler[T], middleware ...gin.HandlerFunc) {
	setup := handler()
	engine.ModelConverter.Add(new(T))
	engine.ModelConverter.Add(setup.Response)
	engine.ApiConverter.Add("PUT", route, new(T), setup.Response, handler, setup.Pagination, setup.Sort)
	engine.GinEngine.PUT(route, joinMiddlewareAndService(newGinServiceHandler(handler), middleware...)...)
}

func DELETE[T any](engine *Engine, route string, handler Handler[T], middleware ...gin.HandlerFunc) {
	setup := handler()
	engine.ModelConverter.Add(new(T))
	engine.ModelConverter.Add(setup.Response)
	engine.ApiConverter.Add("DELETE", route, new(T), setup.Response, handler, setup.Pagination, setup.Sort)
	engine.GinEngine.DELETE(route, joinMiddlewareAndService(newGinServiceHandler(handler), middleware...)...)
}

func WS[T any](engine *Engine, route string, handler WSHandler[T], middleware ...gin.HandlerFunc) {
	engine.GinEngine.GET(route, joinMiddlewareAndService(newGinWSServiceHandler(handler), middleware...)...)
}

func newGinServiceHandler[T any](handler Handler[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context[T]{
			GinContext: c,
			Request:    GinRequest[T](c),
			Page:       GinRequest[sql.Pagination](c),
			Sort:       GinRequest[sql.Sort](c),
		}
		handlerSetup := handler()
		resp, err := handlerSetup.Service(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.OK(resp, *ctx.Page)
	}
}

var wsUpGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func newGinWSServiceHandler[T any](handler WSHandler[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := wsUpGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer ws.Close()
		ctx := &Context[T]{
			GinContext: c,
			Request:    GinRequest[T](c),
			Page:       GinRequest[sql.Pagination](c),
			Sort:       GinRequest[sql.Sort](c),
		}
		handlerSetup := handler()
		err1 := handlerSetup.Service(ctx, ws)
		if err != nil {
			ctx.Error(err1)
		}
	}
}

func joinMiddlewareAndService(service gin.HandlerFunc, middleware ...gin.HandlerFunc) []gin.HandlerFunc {
	var funcs = []gin.HandlerFunc{service}
	funcs = append(funcs, middleware...)
	return funcs
}
