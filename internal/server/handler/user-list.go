package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) ListUser(ctx context.Context, _ *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	users, err := f.userService.UserList(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListUserResponse{
		Users: convertUserListToPb(users),
	}, nil
}
