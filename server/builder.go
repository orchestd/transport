package server

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/log"
	"bitbucket.org/HeilaSystems/transport/discoveryService"
	"github.com/gin-gonic/gin"
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
	SetWriteTimeout(d time.Duration) HttpBuilder
	SetReadTimeout(d time.Duration) HttpBuilder
	SetLogger(logger log.Logger) HttpBuilder
	AddInterceptors(...gin.HandlerFunc) HttpBuilder
	AddSystemHandlers(...IHandler) HttpBuilder
	Build(lifecycle fx.Lifecycle) gin.IRouter
	SetDiscoveryServiceProvider(dsp discoveryService.DiscoveryServiceProvider) HttpBuilder
}
