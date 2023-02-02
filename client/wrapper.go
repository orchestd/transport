package client

import (
	"context"
	"github.com/orchestd/servicereply"
	"github.com/orchestd/transport/discoveryService"
)

type HttpClient interface {
	InternalClient
	Post(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	Get(c context.Context, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	ExternalGet(c context.Context, host, handler string, payload map[string]string, target interface{},
		headers map[string]string, contentType string) servicereply.ServiceReply
	Put(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	Delete(c context.Context, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	PostForm(c context.Context, uri string, postData, headers map[string]string) ([]byte, servicereply.ServiceReply)
	SetDiscoveryServiceProvider(dsp discoveryService.DiscoveryServiceProvider)
}

type InternalClient interface {
	Call(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
}
