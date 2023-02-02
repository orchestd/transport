package server

import (
	"github.com/gin-gonic/gin"
	"github.com/orchestd/dependencybundler/interfaces/log"
	"github.com/orchestd/transport/discoveryService"
	"go.uber.org/fx"
	"time"
)

type HTTPType string

const (
	MethodGet    HTTPType = "GET"
	MethodPost   HTTPType = "POST"
	MethodPut    HTTPType = "PUT"
	MethodDelete HTTPType = "DELETE"
)

type Handler struct {
	HttpType HTTPType
	Method   string
	Handler  []gin.HandlerFunc
}

func (h *Handler) GetHttpType() HTTPType {
	return h.HttpType
}

func (h *Handler) GetMethod() string {
	return h.Method
}

func (h *Handler) GetHandler() []gin.HandlerFunc {
	return h.Handler
}

func NewHttpHandler(httpType HTTPType, method string, handler ...gin.HandlerFunc) func() IHandler {
	return func() IHandler {
		return &Handler{
			HttpType: httpType,
			Method:   method,
			Handler:  handler,
		}
	}
}

type IHandler interface {
	GetHttpType() HTTPType
	GetMethod() string
	GetHandler() []gin.HandlerFunc
}

type HttpBuilder interface {
	SetPort(port string) HttpBuilder
	SetStatics(statics map[string]string) HttpBuilder
	SetWriteTimeout(d time.Duration) HttpBuilder
	SetReadTimeout(d time.Duration) HttpBuilder
	SetLogger(logger log.Logger) HttpBuilder
	AddApiInterceptors(...gin.HandlerFunc) HttpBuilder
	AddRouterInterceptors(...gin.HandlerFunc) HttpBuilder
	AddSystemHandlers(...IHandler) HttpBuilder
	Build(lifecycle fx.Lifecycle) gin.IRouter
	SetDiscoveryServiceProvider(dsp discoveryService.DiscoveryServiceProvider) HttpBuilder
}
