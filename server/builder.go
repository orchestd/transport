package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"time"
)


type HttpBuilder interface {
	SetPort (port string) HttpBuilder
	SetWriteTimeout(d time.Duration) HttpBuilder
	SetReadTimeout(d time.Duration) HttpBuilder
	AddInterceptors(...gin.HandlerFunc) HttpBuilder
	Build(lifecycle fx.Lifecycle) gin.IRouter
}