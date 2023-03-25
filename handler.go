package ginger

type Handler[T any] func() HandlerResponse[T]

type HandlerResponse[T any] struct {
	Service    Service[T]
	Response   interface{}
	Pagination bool
	Sort       bool
}

type WSHandler[T any] func() WSHandlerResponse[T]

type WSHandlerResponse[T any] struct {
	Service WSService[T]
}
