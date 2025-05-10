package main

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("Graceful Shutdown", func(t *testing.T) {
		yamlContent := `
			server:
			  GRPC_PORT: 0
			subpub:
			  buffer_size: 100`
		tmpFile, err := ioutil.TempFile("", "config-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		os.Setenv("CONFIG_PATH", tmpFile.Name())
		defer os.Unsetenv("CONFIG_PATH")

		done := make(chan error)
		go func() {
			done <- run()
		}()

		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Run did not complete in time")
		}
	})
}
