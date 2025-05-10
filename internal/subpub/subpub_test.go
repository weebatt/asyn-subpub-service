package subpub

import (
	"context"
	"go.uber.org/zap"
	"testing"
	"time"
)

const (
	key = "logger"
)

func TestSubPub(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.WithValue(context.Background(), key, zap.NewNop())

	sp := NewSubPub()
	var received []interface{}
	handler := func(msg interface{}) {
		received = append(received, msg)
	}

	sub, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if err := sp.Publish(ctx, "test", "msg1"); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if err := sp.Publish(ctx, "test", "msg2"); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if len(received) != 2 || received[0] != "msg1" || received[1] != "msg2" {
		t.Errorf("Expected [msg1, msg2], got %v", received)
	}

	sub.Unsubscribe()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := sp.Close(ctx); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
