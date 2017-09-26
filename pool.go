package grpcpool

import (
    "sync"
    "google.golang.org/grpc"
    "errors"
)

var (
    grpcPool sync.Map // map[service_name]connPool
)

type ServiceArg struct {
    Service string
    Target string
    Opts []grpc.DialOption
}

// Create is that create a bundle of connection pool instance
func Create(cg ConnGenerator, initialConn int, maxConn int, serviceArgs ...ServiceArg) error {
    if cg == nil || initialConn < 0 || maxConn < 0 || initialConn > maxConn || len(serviceArgs) < 0 {
        return errors.New("Invalid arguments")
    }
    for _, serviceArg := range serviceArgs {
        if _, ok := grpcPool.Load(serviceArg.Service); !ok {
            cp := &connPool{
                grpcConns: make(chan *grpc.ClientConn, maxConn),
                cg:        cg,
                target:    serviceArg.Target,
                opts:      serviceArg.Opts,
            }
            for i := 0; i < initialConn; i++ {
                c, err := cg(serviceArg.Target, serviceArg.Opts...)
                if err != nil {
                    return err
                }
                cp.grpcConns <- c
            }
            grpcPool.Store(serviceArg.Service, cp)
        }
    }
    return nil
}

// Get is that try to get a grpc connection from grpc pool by specified service name
func Get(service string) (*grpc.ClientConn, error){
    if val, ok := grpcPool.Load(service); ok {
        return val.(*connPool).get()
    }
    return nil, errors.New("Invalid service: "+ service)
}

// PutBack is that give back a specific grpc service connection to grpc pool
func PutBack(service string, conn *grpc.ClientConn) error {
    if val, ok := grpcPool.Load(service); ok {
        return val.(*connPool).putBack(conn)
    }
    return errors.New("Invalid service: "+ service)
}

// Len is that get length of specific grpc service connection pool
func Len(service string) int {
    if val, ok := grpcPool.Load(service); ok {
        return val.(*connPool).len()
    }
    return 0
}
