package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) ListCategory(ctx context.Context, _ *pb.ListCategoryRequest) (*pb.ListCategoryResponse, error) {
	categories, err := f.categoryService.CategoryList(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListCategoryResponse{
		Category: convertCategoryListToPb(categories),
	}, nil
}
