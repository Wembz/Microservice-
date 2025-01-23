package discovery

import (
	"context"
	"math/rand"
	"fmt"
	"time"
)

//Creating a registry interface

type Registry interface {
	Register(ctx context.Context, instanceID, serverName, hostPort string) error
	Deregister(ctx context.Context, instanceID, serverName string) error
	Discover(ctx context.Context, serverName string) ([]string, error)
	HealthCheck(instanceID, serverName string) error

}

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
