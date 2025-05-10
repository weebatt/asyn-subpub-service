package config

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("Valid YAML file", func(t *testing.T) {
		yamlContent := `
			server:
			  GRPC_PORT: 50051
			subpub:
			  BUFFER_SIZE: 100`

		tmpFile, err := ioutil.TempFile("", "config-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		log.Println(tmpFile.Name())

		cfg, err := New(tmpFile.Name())
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		if cfg.Server.GRPCPort != 50051 {
			t.Errorf("Expected GRPC_PORT 50051, got %d", cfg.Server.GRPCPort)
		}
		if cfg.SubPub.BufferSize != 100 {
			t.Errorf("Expected BUFFER_SIZE 100, got %d", cfg.SubPub.BufferSize)
		}
	})

	t.Run("Missing file with env vars", func(t *testing.T) {
		os.Setenv("GRPC_PORT", "50052")
		os.Setenv("BUFFER_SIZE", "300")
		defer os.Unsetenv("GRPC_PORT")
		defer os.Unsetenv("BUFFER_SIZE")

		cfg, err := New("nonexistent.yaml")
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		log.Println(cfg.Server.GRPCPort, cfg.SubPub.BufferSize)

		if cfg.Server.GRPCPort != 50052 {
			t.Errorf("Expected GRPC_PORT 50052, got %d", cfg.Server.GRPCPort)
		}
		if cfg.SubPub.BufferSize != 300 {
			t.Errorf("Expected BUFFER_SIZE 300, got %d", cfg.SubPub.BufferSize)
		}
	})

	t.Run("Invalid YAML file", func(t *testing.T) {
		yamlContent := `
			server:
			  GRPC_PORT: invalid
			subpub:
			  BUFFER_SIZE: 20`
		tmpFile, err := ioutil.TempFile("", "config-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		cfg, err := New(tmpFile.Name())
		if err != nil {
			t.Fatalf("Expected error for invalid YAML, got cfg: %+v", cfg)
		}
	})
}
