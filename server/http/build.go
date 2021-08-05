package http

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/log"
	"bitbucket.org/HeilaSystems/transport/server"
	"container/list"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"time"
)

type httpServerSettings struct {
	Port                    *string
	WriteTimeOut            *time.Duration
	ReadTimeOut             *time.Duration
	Logger                  log.Logger
	interceptors            []gin.HandlerFunc
	systemHandlers  []server.IHandler
}

type defaultHttpServerConfigBuilder struct {
	ll *list.List
}

func Builder() server.HttpBuilder {
	return &defaultHttpServerConfigBuilder{ll: list.New()}
}

func (d *defaultHttpServerConfigBuilder) SetPort(port string) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.Port = &port
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) SetWriteTimeout(duration time.Duration) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.WriteTimeOut = &duration
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) SetReadTimeout(duration time.Duration) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.ReadTimeOut = &duration
	})
	return d
}
func (d *defaultHttpServerConfigBuilder) SetLogger(logger log.Logger) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.Logger = logger
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) AddInterceptors(interceptors ...gin.HandlerFunc) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.interceptors =  append(cfg.interceptors , interceptors...)
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) AddSystemHandlers(handlers ...server.IHandler) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.systemHandlers = handlers
	})
	return d
}


func (d *defaultHttpServerConfigBuilder) Build(lc fx.Lifecycle) gin.IRouter {
	httpCfg := &httpServerSettings{}
	for e := d.ll.Front(); e != nil; e = e.Next() {
		f := e.Value.(func(cfg *httpServerSettings))
		f(httpCfg)
	}
	return NewGinServer(lc, httpCfg.Port, httpCfg.WriteTimeOut, httpCfg.ReadTimeOut, httpCfg.Logger, httpCfg.interceptors, httpCfg.systemHandlers)
}
