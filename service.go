package ginger

import "github.com/gorilla/websocket"

type Service[T any] func(ctx *Context[T]) (interface{}, Error)

type WSService[T any] func(ctx *Context[T], ws *websocket.Conn) Error
