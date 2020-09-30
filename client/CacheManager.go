package client

import (
	"sync"

	"github.com/robert-pkg/micro-go/rpc/client/grpc"
)

var (
	cacheManager CacheManager
)

// CacheManager .
type CacheManager struct {
	clientMap sync.Map

	mutex sync.Mutex
}

// Stop .
func (cm *CacheManager) Stop() {

	cm.clientMap.Range(func(key, value interface{}) bool {
		if c, ok := value.(*grpc.Client); ok {
			c.Stop()
		}

		return true
	})

}

// GetClient .
func (cm *CacheManager) GetClient(serviceName string) (*grpc.Client, error) {

	if c, ok := cm.clientMap.Load(serviceName); ok {
		// 大部分情况下，这里直接返回了
		return c.(*grpc.Client), nil
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if c, ok := cm.clientMap.Load(serviceName); ok {
		// 由于并发，现在有 client了
		return c.(*grpc.Client), nil
	}

	c, err := grpc.NewClient(serviceName)
	if err != nil {
		return nil, err
	}

	cm.clientMap.Store(serviceName, c)
	return c, nil
}
