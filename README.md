# gRPC Connection Pool
[![Actions Status](https://github.com/wuxinwei/grpcpool/workflows/build/badge.svg)](https://github.com/wuxinwei/grppool/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/wuxinwei/grpcpool)](https://goreportcard.com/report/github.com/wuxinwei/grpcpool)
[![codecov](https://codecov.io/gh/wuxinwei/grpcpool/branch/master/graph/badge.svg)](https://codecov.io/gh/wuxinwei/grpcpool)

It's a simple implement of grpc connection pool, based on buffered channel.

# Requirement
- Go 1.9+

# Usage

```go
import "github.com/wuxinwei/grpcpool"

// create a grpc pool
sa := ServiceArg
Create(grpc.Dial, 5, 10, sa)

// got a connection from pool, and create a grpc client
conn, _ := grpcpool.Get()
cli := pb.NewHelloClient(conn)

// do whatever you want with grpc client

// after you finish your work, remember to put the conn back into the pool
grpcpool.PutBack(conn)

// you can close your pool do your own purpose
grpcpool.Close()
```
