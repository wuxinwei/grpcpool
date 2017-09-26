package grpcpool
import (
    "google.golang.org/grpc"
    "errors"
)

// connPool is that grpc connection pool by buffered channel
type connPool struct {
    grpcConns chan *grpc.ClientConn
    cg ConnGenerator
    target string
    opts []grpc.DialOption
    
}

// ConnGenerator is function type to generate a grpc connection function
type ConnGenerator func (target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)

func (c *connPool) get() (*grpc.ClientConn, error){
    select {
    case conn := <- c.grpcConns :
        if conn == nil {
            return nil, errors.New("connection is closed")
        }
        return conn, nil
    default :
        // channel is empty
        conn, err := c.cg(c.target, c.opts...)
        if err != nil {
            return nil, err
        }
        return conn, nil
    }
}

func (c *connPool) putBack(conn *grpc.ClientConn) error {
    select {
    case c.grpcConns <- conn:
        return nil
    default:
        // channel if full
        return conn.Close()
    }
}

func (c *connPool) len() int {
    return len(c.grpcConns)
}
