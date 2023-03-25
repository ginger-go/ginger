package ginger

import (
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	_gin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

var Middleware = new(middleware)

type middleware struct{}

func (m *middleware) CORS(config cors.Config) gin.HandlerFunc {
	return cors.New(config)
}

func (m *middleware) RateLimit(duration time.Duration, rate int64) gin.HandlerFunc {
	return _gin.NewMiddleware(limiter.New(memory.NewStore(), limiter.Rate{Period: duration, Limit: rate}))
}

func (m *middleware) Cache(duration time.Duration, handler gin.HandlerFunc) gin.HandlerFunc {
	return cache.CachePage(persistence.NewInMemoryStore(time.Second), duration, handler)
}
