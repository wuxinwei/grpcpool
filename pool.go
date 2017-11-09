package grpcpool

import (
	"errors"
	"sync"

	"context"

	"google.golang.org/grpc"
)

var (
	pool sync.Map // map[serviceName]connPool
)

// ServiceArg is that specified gRPC service configuration
type ServiceArg struct {
	Service string
	Target  string // address:port, e.g: 127.0.0.1:8088
	Opts    []grpc.DialOption
}

// Create a bundle of connection pool instance
func Create(ctx context.Context, cg ConnGenerator, initialConn, maxConn int, serviceArgs ...ServiceArg) error {
	if cg == nil || initialConn <= 0 || maxConn <= 0 || initialConn > maxConn || len(serviceArgs) < 0 {
		return errors.New("invalid arguments")
	}
	for _, serviceArg := range serviceArgs {
		if _, ok := pool.Load(serviceArg.Service); !ok {
			cp := &connPool{
				conns:  make(chan *grpc.ClientConn, maxConn),
				cg:     cg,
				target: serviceArg.Target,
				opts:   serviceArg.Opts,
			}
			for i := 0; i < initialConn; i++ {
				c, err := cg(serviceArg.Target, serviceArg.Opts...)
				if err != nil {
					return err
				}
				cp.conns <- c
			}
			pool.Store(serviceArg.Service, cp)
		}
	}
	return nil
}

// Get is that try to get a grpc connection from grpc pool by specified service name
func Get(ctx context.Context, service string) (*grpc.ClientConn, error) {
	if val, ok := pool.Load(service); ok {
		return val.(*connPool).get()
	}
	return nil, errors.New("Invalid service: " + service)
}

// PutBack is that give back a specific gRPC service connection to gRPC pool
func PutBack(ctx context.Context, service string, conn *grpc.ClientConn) error {
	if val, ok := pool.Load(service); ok {
		return val.(*connPool).putBack(conn)
	}
	return errors.New("Invalid service: " + service)
}

// Len is that get length of specific grpc service connection pool
func Len(ctx context.Context, service string) int {
	if val, ok := pool.Load(service); ok {
		return val.(*connPool).len()
	}
	return 0
}

// Close the gRPC pool totally
func Close(ctx context.Context) {
	pool.Range(func(key, value interface{}) bool {
		connPool, ok := value.(*connPool)
		if !ok {
			return ok
		}
		connPool.close()
		for conn := range connPool.conns {
			if err := conn.Close(); err != nil {
				return false
			}
		}
		return true
	})
}
