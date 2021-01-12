package http

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/configuration"
	"bitbucket.org/HeilaSystems/transport/client"
	"container/list"
	"fmt"
	"net/http"
)

type httpClientBuilderConfig struct {
	predefinedClient *http.Client
	interceptors     []client.HTTPClientInterceptor
	conf configuration.Config
}

type builderImpl struct {
	ll *list.List
}

func HTTPClientBuilder() client.HTTPClientBuilder {
	return &builderImpl{
		ll: list.New(),
	}
}

func (impl *builderImpl) SetConfig(conf configuration.Config) client.HTTPClientBuilder {
	impl.ll.PushBack(func(cfg *httpClientBuilderConfig) {
		cfg.conf = conf
	})
	return impl
}


func (impl *builderImpl) AddInterceptors(interceptors ...client.HTTPClientInterceptor) client.HTTPClientBuilder {
	impl.ll.PushBack(func(cfg *httpClientBuilderConfig) {
		if len(interceptors) > 0 {
			cfg.interceptors = append(cfg.interceptors, interceptors...)
		}
	})
	return impl
}

func (impl *builderImpl) WithPreconfiguredClient(client *http.Client) client.HTTPClientBuilder {
	impl.ll.PushBack(func(cfg *httpClientBuilderConfig) {
		cfg.predefinedClient = client
	})
	return impl
}

func (impl *builderImpl) Build() (client.HttpClient,error) {
	var client = &http.Client{}
	var conf configuration.Config
	if impl != nil {
		cfg := new(httpClientBuilderConfig)
		for e := impl.ll.Front(); e != nil; e = e.Next() {
			f := e.Value.(func(cfg *httpClientBuilderConfig))
			f(cfg)
		}
		if cfg.conf == nil {
			return nil, fmt.Errorf("Cannot initialize Http client without configuration dependency")
		}
		conf = cfg.conf
		if cfg.predefinedClient != nil {
			client = cfg.predefinedClient
		}
		if client.Transport == nil {
			client.Transport = http.DefaultTransport
		}


		client.Transport = prepareCustomRoundTripper(client.Transport, cfg.interceptors...)
	}
	return NewHttpClientWrapper(client,conf)
}

type customRoundTripper struct {
	inner             http.RoundTripper
	unitedInterceptor client.HTTPClientInterceptor
}

func prepareCustomRoundTripper(actual http.RoundTripper, interceptors ...client.HTTPClientInterceptor) http.RoundTripper {
	return &customRoundTripper{
		inner:             actual,
		unitedInterceptor: uniteInterceptors(interceptors),
	}
}

func (crt *customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return crt.unitedInterceptor(req, crt.inner.RoundTrip)
}

func uniteInterceptors(interceptors []client.HTTPClientInterceptor) client.HTTPClientInterceptor {
	if len(interceptors) == 0 {
		return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
			// That's why we needed an alias to http.RoundTripper.RoundTrip
			return handler(req)
		}
	}

	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		tailHandler := func(innerReq *http.Request) (*http.Response, error) {
			unitedInterceptor := uniteInterceptors(interceptors[1:])
			return unitedInterceptor(req, handler)
		}
		headInterceptor := interceptors[0]
		return headInterceptor(req, tailHandler)
	}
}
var _ client.HTTPClientBuilder = (*builderImpl)(nil)
