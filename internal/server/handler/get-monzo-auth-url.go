package handler

import (
	"context"

	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) GetMonzoAuthURL(_ context.Context, _ *pb.GetMonzoAuthURLRequest) (*pb.GetMonzoAuthURLResponse, error) {
	authURL, err := f.monzoService.GetAuthorizationURL()
	if err != nil {
		return nil, err
	}

	return &pb.GetMonzoAuthURLResponse{
		AuthUrl: authURL,
	}, nil
}
