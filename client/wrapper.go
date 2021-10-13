package client

import (
	"bitbucket.org/HeilaSystems/servicereply"
	"bitbucket.org/HeilaSystems/transport/discoveryService"
	"context"
)

type HttpClient interface {
	InternalClient
	Post(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	Get(c context.Context, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	Put(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
	Delete(c context.Context, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply

	SetDiscoveryServiceProvider(dsp discoveryService.DiscoveryServiceProvider)
}
type InternalClient interface {
	Call(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) servicereply.ServiceReply
}
