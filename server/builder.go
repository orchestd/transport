package server

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"time"
)


type HttpBuilder interface {
	SetPort (port string) HttpBuilder
	SetWriteTimeout(d time.Duration) HttpBuilder
	SetReadTimeout(d time.Duration) HttpBuilder
	SetLogger(logger log.Logger) HttpBuilder
	AddContextInterceptors(...gin.HandlerFunc)HttpBuilder
	AddInterceptors(...gin.HandlerFunc) HttpBuilder
	Build(lifecycle fx.Lifecycle) gin.IRouter
}