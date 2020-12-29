package http

import (
	"bitbucket.org/HeilaSystems/transport/server"
	"container/list"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"time"
)

type httpServerSettings struct {
	Port *string
	WriteTimeOut *time.Duration
	ReadTimeOut *time.Duration
	interceptors []server.HttpServerInterceptors
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

func (d *defaultHttpServerConfigBuilder) AddInterceptors(interceptors ...server.HttpServerInterceptors) server.HttpBuilder {
	d.ll.PushBack(func(cfg *httpServerSettings) {
		cfg.interceptors = interceptors
	})
	return d
}

func (d *defaultHttpServerConfigBuilder) Build(lc fx.Lifecycle) gin.IRouter {
	httpCfg := &httpServerSettings{}
	for e := d.ll.Front(); e != nil; e = e.Next() {
		f := e.Value.(func(cfg *httpServerSettings))
		f(httpCfg)
	}
	return NewGinServer(lc , httpCfg.Port,httpCfg.WriteTimeOut,httpCfg.ReadTimeOut,httpCfg.interceptors...)
}



