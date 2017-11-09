package grpcpool

import (
	"errors"

	"google.golang.org/grpc"
)

// connPool is that gRPC connection pool by buffered channel
type connPool struct {
	conns  chan *grpc.ClientConn
	cg     ConnGenerator
	target string
	opts   []grpc.DialOption
}

// ConnGenerator is function type to generate a grpc connection function
type ConnGenerator func(target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)

func (c *connPool) get() (*grpc.ClientConn, error) {
	select {
	case conn := <-c.conns:
		if conn == nil {
			return nil, errors.New("connection is closed")
		}
		return conn, nil
	default:
		// channel is empty
		conn, err := c.cg(c.target, c.opts...)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (c *connPool) putBack(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("conn is nil")
	}
	select {
	case c.conns <- conn:
		return nil
	default:
		// channel if full
		return conn.Close()
	}
}

func (c *connPool) len() int {
	return len(c.conns)
}

func (c *connPool) close() {
	close(c.conns)
}
