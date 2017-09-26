package grpcpool

import (
	"context"
	"log"
	"net"
	"testing"

	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func init() {
	go func() {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &server{})
		// Register reflection service on gRPC server.
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	time.Sleep(time.Second * 1)
}

func TestInit(t *testing.T) {
	sa := ServiceArg{
		Service: "clientHello",
		Target:  "127.0.0.1:19772",
		Opts:    []grpc.DialOption{grpc.WithInsecure()},
	}
	if err := Create(grpc.Dial, 3, 10, sa); err != nil {
		t.Errorf("Want: nil, Got: %s", err)
	}
}

func TestGet(t *testing.T) {
	sa := ServiceArg{
		Service: "hello",
		Target:  "127.0.0.1:19772",
		Opts:    []grpc.DialOption{grpc.WithInsecure()},
	}
	if err := Create(grpc.Dial, 3, 10, sa); err != nil {
		t.Errorf("Want: nil, Got: %s", err)
	}
	conn, err := Get("hello")
	if err != nil {
		t.Errorf("Want: nil, Got: %s", err)
	}
	client := pb.NewGreeterClient(conn)
	client.SayHello(context.Background(), &pb.HelloRequest{Name: "hello"})

	if err := PutBack("hello", conn); err != nil {
		t.Errorf("Want: nil, Got: %s", err)
	}
}

func TestBack(t *testing.T) {
}

func TestLen(t *testing.T) {
}
