package ginger

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ginger-go/ginger/typescript"
	"github.com/ginger-go/sql"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron"
)

type Engine struct {
	GinEngine      *gin.Engine
	ModelConverter *typescript.ModelConverter
	ApiConverter   *typescript.ApiConverter
	CronWorker     *cron.Cron
}

func NewEngine() *Engine {
	return &Engine{
		GinEngine:      gin.Default(),
		ModelConverter: typescript.NewModelConverter(),
		ApiConverter:   typescript.NewApiConverter(),
		CronWorker:     cron.New(),
	}
}

func (e *Engine) Run(addr string) {
	e.CronWorker.Start()
	e.GinEngine.Run(addr)
}

func (e *Engine) RunServerOnly(addr string) {
	e.GinEngine.Run(addr)
}

func (e *Engine) RunCronOnly() {
	e.CronWorker.Start()
}

func (e *Engine) Use(middleware ...gin.HandlerFunc) {
	e.GinEngine.Use(middleware...)
}

func (e *Engine) GenerateTypescript() {
	os.RemoveAll("api")
	err := os.Mkdir("api", os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("api/model.ts", []byte(e.ModelConverter.ToString()), os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("api/api.ts", []byte(e.ApiConverter.ToString()), os.ModePerm)
	if err != nil {
		panic(err)
	}
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

func Cron(engine *Engine, spec string, job func()) {
	engine.CronWorker.AddFunc(spec, job)
}

func newGinServiceHandler[T any](handler Handler[T]) gin.HandlerFunc {
	handlerSetup := handler()
	return func(c *gin.Context) {
		ctx := &Context[T]{
			GinContext: c,
			Request:    GinRequest[T](c),
		}
		if handlerSetup.Pagination {
			ctx.Page = GinRequest[sql.Pagination](c)
		}
		if handlerSetup.Sort {
			ctx.Sort = GinRequest[sql.Sort](c)
		}
		resp, err := handlerSetup.Service(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.OK(resp, ctx.Page)
	}
}

var wsUpGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func newGinWSServiceHandler[T any](handler WSHandler[T]) gin.HandlerFunc {
	handlerSetup := handler()
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
		}
		err1 := handlerSetup.Service(ctx, ws)
		if err != nil {
			ctx.Error(err1)
		}
	}
}

func joinMiddlewareAndService(service gin.HandlerFunc, middleware ...gin.HandlerFunc) []gin.HandlerFunc {
	var funcs = make([]gin.HandlerFunc, 0)
	if len(middleware) > 0 {
		funcs = append(funcs, middleware...)
	}
	funcs = append(funcs, service)
	return funcs
}
