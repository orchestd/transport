package http

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/log"
	"bitbucket.org/HeilaSystems/transport/discoveryService"
	"bitbucket.org/HeilaSystems/transport/server"
	"container/list"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"time"
)

type HttpServerSettings struct {
	Port                     *string
	WriteTimeOut             *time.Duration
	ReadTimeOut              *time.Duration
	Logger                   log.Logger
	apiInterceptors          []gin.HandlerFunc
	routerInterceptors       []gin.HandlerFunc
	systemHandlers           []server.IHandler
	DiscoveryServiceProvider discoveryService.DiscoveryServiceProvider
	Statics                  map[string]string
}

type defaultHttpServerConfigBuilder struct {
	ll *list.List
}

func Builder() server.HttpBuilder {
	return &defaultHttpServerConfigBuilder{ll: list.New()}
}

func (d *defaultHttpServerConfigBuilder) SetStatics(statics map[string]string) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.Statics = statics
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) SetPort(port string) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.Port = &port
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) SetWriteTimeout(duration time.Duration) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.WriteTimeOut = &duration
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) SetReadTimeout(duration time.Duration) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.ReadTimeOut = &duration
	})
	return d
}
func (d *defaultHttpServerConfigBuilder) SetLogger(logger log.Logger) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.Logger = logger
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) AddApiInterceptors(interceptors ...gin.HandlerFunc) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.apiInterceptors = append(cfg.apiInterceptors, interceptors...)
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) AddRouterInterceptors(interceptors ...gin.HandlerFunc) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.routerInterceptors = append(cfg.routerInterceptors, interceptors...)
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) AddSystemHandlers(handlers ...server.IHandler) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.systemHandlers = handlers
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) Build(lc fx.Lifecycle) gin.IRouter {
	httpCfg := &HttpServerSettings{}
	for e := d.ll.Front(); e != nil; e = e.Next() {
		f := e.Value.(func(cfg *HttpServerSettings))
		f(httpCfg)
	}

	return NewGinServer(httpCfg.DiscoveryServiceProvider, lc, httpCfg.Port, httpCfg.WriteTimeOut, httpCfg.ReadTimeOut,
		httpCfg.Logger, httpCfg.apiInterceptors, httpCfg.routerInterceptors, httpCfg.systemHandlers, httpCfg.Statics)
}

func (d *defaultHttpServerConfigBuilder) SetDiscoveryServiceProvider(ds discoveryService.DiscoveryServiceProvider) server.HttpBuilder {
	d.ll.PushBack(func(cfg *HttpServerSettings) {
		cfg.DiscoveryServiceProvider = ds
	})
	return d
}
