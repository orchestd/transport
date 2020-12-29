package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"time"
)

type HttpServerInterceptors func(next http.Handler) http.Handler
type HttpBuilder interface {
	SetPort (port string) HttpBuilder
	SetWriteTimeout(d time.Duration) HttpBuilder
	SetReadTimeout(d time.Duration) HttpBuilder
	AddInterceptors(...HttpServerInterceptors) HttpBuilder
	Build(lifecycle fx.Lifecycle) gin.IRouter
}