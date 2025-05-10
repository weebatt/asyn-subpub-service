package subpub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSubPub(t *testing.T) {
	t.Run("Basic Subscribe and Publish", func(t *testing.T) {
		sp := NewSubPub()
		var received []interface{}
		handler := func(msg interface{}) {
			received = append(received, msg)
		}

		sub, err := sp.Subscribe("test", handler)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		if err := sp.Publish("test", "msg1"); err != nil {
			t.Fatalf("Publish failed: %v", err)
		}
		if err := sp.Publish("test", "msg2"); err != nil {
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
	})

	t.Run("Subscribe to Closed SubPub", func(t *testing.T) {
		sp := NewSubPub()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := sp.Close(ctx); err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		_, err := sp.Subscribe("test", func(msg interface{}) {})
		if err == nil || err.Error() != "subpub is closed" {
			t.Errorf("Expected error 'subpub is closed', got %v", err)
		}
	})

	t.Run("Multiple Subscribers", func(t *testing.T) {
		sp := NewSubPub()
		var received1, received2 []interface{}
		handler1 := func(msg interface{}) {
			received1 = append(received1, msg)
		}
		handler2 := func(msg interface{}) {
			received2 = append(received2, msg)
		}

		sub1, err := sp.Subscribe("test", handler1)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
		sub2, err := sp.Subscribe("test", handler2)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		if err := sp.Publish("test", "msg"); err != nil {
			t.Fatalf("Publish failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		if len(received1) != 1 || received1[0] != "msg" {
			t.Errorf("Expected [msg] for subscriber 1, got %v", received1)
		}
		if len(received2) != 1 || received2[0] != "msg" {
			t.Errorf("Expected [msg] for subscriber 2, got %v", received2)
		}

		sub1.Unsubscribe()
		sub2.Unsubscribe()
	})

	t.Run("Publish to Non-Existent Subject", func(t *testing.T) {
		sp := NewSubPub()
		if err := sp.Publish("nonexistent", "msg"); err != nil {
			t.Errorf("Expected no error for nonexistent subject, got %v", err)
		}
	})

	t.Run("Full Buffer", func(t *testing.T) {
		sp := NewSubPub()
		var received []interface{}
		handler := func(msg interface{}) {
			received = append(received, msg)
			time.Sleep(100 * time.Millisecond)
		}

		_, err := sp.Subscribe("test", handler)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		for i := 0; i < 110; i++ {
			sp.Publish("test", "msg")
		}

		time.Sleep(200 * time.Millisecond)
		if len(received) > 100 {
			t.Errorf("Expected at most 100 messages, got %d", len(received))
		}
	})

	t.Run("Double Unsubscribe", func(t *testing.T) {
		sp := NewSubPub()
		sub, err := sp.Subscribe("test", func(msg interface{}) {})
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		sub.Unsubscribe()
		sub.Unsubscribe()
	})

	t.Run("Close with Timeout", func(t *testing.T) {
		sp := NewSubPub()
		_, err := sp.Subscribe("test", func(msg interface{}) {
			time.Sleep(time.Second)
		})
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		for i := 0; i < 10; i++ {
			if err := sp.Publish("test", "msg"); err != nil {
				t.Fatalf("Publish failed: %v", err)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err = sp.Close(ctx)
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("Concurrent Subscribe and Publish", func(t *testing.T) {
		sp := NewSubPub()
		var wg sync.WaitGroup
		numSubscribers := 10
		numMessages := 50

		for i := 0; i < numSubscribers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, err := sp.Subscribe("test", func(msg interface{}) {})
				if err != nil {
					t.Errorf("Subscribe failed for subscriber %d: %v", i, err)
				}
			}(i)
		}

		for i := 0; i < numMessages; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if err := sp.Publish("test", "msg"); err != nil {
					t.Errorf("Publish failed for message %d: %v", i, err)
				}
			}(i)
		}

		wg.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := sp.Close(ctx); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})
}
