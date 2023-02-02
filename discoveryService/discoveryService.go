package discoveryService

import (
	"github.com/orchestd/servicereply"
)

type DiscoveryServiceProvider interface {
	Register() servicereply.ServiceReply
	GetAddress(serviceName string) servicereply.ServiceReply
}
