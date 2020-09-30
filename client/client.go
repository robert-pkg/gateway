package client

import (
	"github.com/robert-pkg/micro-go/rpc/client/grpc"
)

// GetClient .
func GetClient(serviceName string) (*grpc.Client, error) {
	return cacheManager.GetClient(serviceName)
}

// Stop .
func Stop() {
	cacheManager.Stop()
}
