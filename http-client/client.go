package http_client

import (
	"github.com/robert-pkg/micro-go/rpc/client/http"
)

// GetClient .
func GetClient(serviceName string) (*http.Client, error) {
	return cacheManager.GetClient(serviceName)
}

// Stop .
func Stop() {
	cacheManager.Stop()
}
