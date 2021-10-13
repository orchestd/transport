package discoveryService

import (
	"bitbucket.org/HeilaSystems/servicereply"
)

type DiscoveryServiceProvider interface {
	Register() servicereply.ServiceReply
	GetAddress(serviceName string) servicereply.ServiceReply
}
