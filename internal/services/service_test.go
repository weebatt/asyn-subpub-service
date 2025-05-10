package services

import (
	"asyn-subpub-service/internal/subpub"
	"asyn-subpub-service/pb/proto/api"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Run("Valid SubPub", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		server := NewServer(sp)
		if server == nil {
			t.Fatal("NewServer returned nil")
		}
		if server.subpub == nil {
			t.Fatal("subpub is nil")
		}
	})

	t.Run("Nil SubPub", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected panic for nil subpub")
			}
		}()
		NewServer(nil)
	})
}

func TestServerPublish(t *testing.T) {
	t.Run("Successful Publish", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		server := NewServer(sp)
		req := &pb.PublishRequest{Key: "test", Data: "hello"}
		resp, err := server.Publish(context.Background(), req)
		if err != nil {
			t.Fatalf("Publish failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response is nil")
		}
	})
}

func TestServerSubscribe(t *testing.T) {
	t.Run("Successful Subscribe", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		server := NewServer(sp)
		req := &pb.SubscribeRequest{Key: "test"}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		stream := &mockPubSubStream{
			send: func(event *pb.Event) error {
				if event.Data != "hello" {
					t.Errorf("Expected data: hello, got %s", event.Data)
				}
				return nil
			},
			ctx: ctx,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Subscribe(req, stream); err != nil {
				t.Errorf("Subscribe failed: %v", err)
			}
		}()

		go func() {
			time.Sleep(50 * time.Millisecond)
			server.Publish(context.Background(), &pb.PublishRequest{Key: "test", Data: "hello"})
		}()

		wg.Wait()
	})

	t.Run("Subscribe with Stream Error", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		server := NewServer(sp)
		req := &pb.SubscribeRequest{Key: "test"}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		stream := &mockPubSubStream{
			send: func(event *pb.Event) error {
				return status.Error(codes.Internal, "stream error")
			},
			ctx: ctx,
		}

		if err := server.Subscribe(req, stream); err != nil {
			t.Errorf("Subscribe failed: %v", err)
		}

		server.Publish(context.Background(), &pb.PublishRequest{Key: "test", Data: "hello"})
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("Subscribe with SubPub Error", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		sp.Close(ctx)

		server := NewServer(sp)
		req := &pb.SubscribeRequest{Key: "test"}
		stream := &mockPubSubStream{
			send: func(event *pb.Event) error { return nil },
			ctx:  context.Background(),
		}

		err := server.Subscribe(req, stream)
		if err == nil {
			t.Fatal("Expected error for subscribe with closed subpub")
		}
		if status.Code(err) != codes.Internal {
			t.Errorf("Expected Internal error, got %v", err)
		}
	})

	t.Run("Concurrent Subscribe and Publish", func(t *testing.T) {
		sp := subpub.NewSubPub(100)
		server := NewServer(sp)
		var wg sync.WaitGroup
		numGoroutines := 5

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				stream := &mockPubSubStream{
					send: func(event *pb.Event) error { return nil },
					ctx:  ctx,
				}
				if err := server.Subscribe(&pb.SubscribeRequest{Key: "test"}, stream); err != nil {
					t.Errorf("Subscribe failed: %v", err)
				}
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := server.Publish(context.Background(), &pb.PublishRequest{Key: "test", Data: "hello"})
				if err != nil {
					t.Errorf("Publish failed: %v", err)
				}
			}()
		}

		wg.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := sp.Close(ctx); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})
}

type mockPubSubStream struct {
	send func(*pb.Event) error
	ctx  context.Context
}

func (m *mockPubSubStream) Send(event *pb.Event) error {
	return m.send(event)
}

func (m *mockPubSubStream) Context() context.Context {
	return m.ctx
}

func (m *mockPubSubStream) SetHeader(_ metadata.MD) error {
	return nil
}

func (m *mockPubSubStream) SendHeader(_ metadata.MD) error {
	return nil
}

func (m *mockPubSubStream) SetTrailer(_ metadata.MD) {
}

func (m *mockPubSubStream) SendMsg(_ interface{}) error {
	return nil
}

func (m *mockPubSubStream) RecvMsg(_ interface{}) error {
	return nil
}
