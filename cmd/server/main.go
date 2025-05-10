package main

import (
	"asyn-subpub-service/internal/config"
	"asyn-subpub-service/internal/services"
	"asyn-subpub-service/internal/subpub"
	pb "asyn-subpub-service/pb/proto/api"
	"asyn-subpub-service/pkg/logger"
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Initialize context and logger
	ctx := context.Background()
	ctx, _ = logger.New(ctx)

	// Initialize config
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.New(configPath)
	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal("failed reading config", zap.Error(err))
		return err
	}

	// Create listener
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Server.GRPCPort))
	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal("failed to listen", zap.Error(err))
		return err
	}

	// Initialize subpub and gRPC server
	subPub := subpub.NewSubPub(cfg.SubPub.BufferSize)
	s := grpc.NewServer()
	pb.RegisterPubSubServer(s, services.NewServer(subPub))

	// Start server in a goroutine
	go func() {
		logger.GetLoggerFromContext(ctx).Info("Server listening on", zap.String("port", strconv.Itoa(cfg.Server.GRPCPort)))
		if err := s.Serve(lis); err != nil {
			logger.GetLoggerFromContext(ctx).Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Handle signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Perform graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := subPub.Close(ctx); err != nil {
		logger.GetLoggerFromContext(ctx).Fatal("failed to close subPub", zap.Error(err))
		return err
	}
	s.GracefulStop()
	log.Println("Server stopped gracefully")
	return nil
}
