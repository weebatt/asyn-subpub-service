package services

import (
	"asyn-subpub-service/pb/proto/api"
	"context"
	"google.golang.org/grpc"
	_ "google.golang.org/protobuf/types/known/emptypb"
	"net"
	"testing"
	"time"
)

func TestServerPublishSubscribe(t *testing.T) {
	server := NewServer().UnimplementedPubSubServer
	grpcServer := grpc.NewServer()
	pb.RegisterPubSubServer(grpcServer, server)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewPubSubClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{Key: "test"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	_, err = client.Publish(ctx, &pb.PublishRequest{Key: "test", Data: "hello"})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("failed to receive: %v", err)
	}
	if event.Data != "hello" {
		t.Errorf("expected data: hello, got: %s", event.Data)
	}
}
