package grpcpool

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"testing"

	"time"

	assertpkg "github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":19800"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// start a gRPC hello world server
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
}

func TestGetAndBack(t *testing.T) {
	wg := sync.WaitGroup{}
	assert := assertpkg.New(t)
	sa := ServiceArg{
		Service: "hello",
		Target:  "127.0.0.1:19800",
		Opts:    []grpc.DialOption{grpc.WithInsecure()},
	}
	err := Create(context.Background(), grpc.Dial, runtime.NumCPU(), runtime.NumCPU()*2, sa)
	if !assert.NoError(err, "gRPC.Create") {
		t.Fatal(err)
	}
	connCount := runtime.NumCPU() * 20
	wg.Add(connCount)
	for i := 0; i < connCount; i++ {
		go func(t *testing.T, wg *sync.WaitGroup) {
			conn, err := Get(context.Background(), sa.Service)
			if !assert.NoError(err, "gRPC.Get") {
				t.Fatal(err)
			}
			client := pb.NewGreeterClient(conn)
			r := &pb.HelloRequest{
				Name: "client",
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			res, err := client.SayHello(ctx, r)
			if !assert.NoError(err, "SayHello") {
				t.Fatal(err)
			}
			if !assert.EqualValues(res.Message, "Hello "+r.Name, "SayHello") {
				t.Fatal()
			}
			PutBack(context.Background(), sa.Service, conn)
			wg.Done()
		}(t, &wg)
	}
	wg.Wait()
	if !assert.EqualValues(runtime.NumCPU()*2, Len(context.Background(), sa.Service), "max idle connection") {
		t.Failed()
	}
	Close(context.Background())
}
