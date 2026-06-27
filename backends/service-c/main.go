package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// echoRequest and echoResponse mirror the proto messages using plain JSON
// encoding over gRPC's byte-stream transport. The gateway proxy does not need
// to know the schema, so this keeps service-c self-contained without protoc.
type echoRequest struct {
	Message string `json:"message"`
}

type echoResponse struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// echoServer implements the EchoService.Echo RPC.
type echoServer struct{}

func (s *echoServer) Echo(_ context.Context, req *echoRequest) (*echoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	return &echoResponse{
		Message:   req.Message,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// echoHandler adapts echoServer.Echo to grpc.UnaryHandler.
func echoHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := new(echoRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(*echoServer).Echo(ctx, req)
}

// jsonCodec encodes/decodes using JSON so no protoc-generated code is needed.
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)   { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                     { return "json" }

var echoServiceDesc = grpc.ServiceDesc{
	ServiceName: "echo.EchoService",
	HandlerType: (*echoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler:    echoHandler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	// Register the JSON codec so gRPC uses it instead of protobuf.
	// This keeps service-c independent of protoc-generated code.
	encoding.RegisterCodec(jsonCodec{})

	srv := grpc.NewServer()
	srv.RegisterService(&echoServiceDesc, &echoServer{})
	reflection.Register(srv)

	fmt.Printf("service-c gRPC listening on :%s\n", port)
	if err := srv.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %v\n", err)
		os.Exit(1)
	}
}
