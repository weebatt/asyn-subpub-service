package services

import (
	"asyn-subpub-service/internal/subpub"
	"asyn-subpub-service/pb/proto/api"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)

type Server struct {
	pb.UnimplementedPubSubServer
	subpub subpub.SubPub
}

func NewServer(subpub subpub.SubPub) *Server {
	if subpub == nil {
		panic("subpub is nil")
	}
	return &Server{
		subpub: subpub,
	}
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	_, err := s.subpub.Subscribe(req.Key, func(msg interface{}) {
		event := &pb.Event{Data: msg.(string)}
		if err := stream.Send(event); err != nil {
			log.Printf("Error sending event: %v", err)
		}
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}
	<-stream.Context().Done()
	return nil
}

func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	err := s.subpub.Publish(req.Key, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish: %v", err)
	}
	return &emptypb.Empty{}, nil
}
