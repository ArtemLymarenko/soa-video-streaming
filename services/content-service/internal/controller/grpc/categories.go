package grpc

import (
	"context"
	pb "soa-video-streaming/pkg/pb/content"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/service"
)

type CategoryController struct {
	pb.CategoryServiceServer

	service *service.CategoryService
}

func NewCategoryController(service *service.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

func convertCategories(categories []entity.Category) []*pb.Category {
	res := make([]*pb.Category, len(categories))
	for i, c := range categories {
		res[i] = &pb.Category{
			Id:   string(c.ID),
			Name: c.Name,
		}
	}
	return res
}

func (c *CategoryController) GetCategoriesByTimestamp(
	ctx context.Context,
	req *pb.GetCategoriesByTimestampRequest,
) (*pb.GetCategoriesByTimestampResponse, error) {
	categories, err := c.service.GetByTimestamp(ctx, req.GetFrom(), req.GetTo())
	if err != nil {
		return nil, err
	}

	return &pb.GetCategoriesByTimestampResponse{
		Categories: convertCategories(categories),
	}, nil
}

func (c *CategoryController) GetMaxTimestamp(ctx context.Context, _ *pb.GetMaxTimestampRequest) (*pb.GetMaxTimestampResponse, error) {
	maxTimestamp, err := c.service.GetMaxTimestamp(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetMaxTimestampResponse{
		MaxTimestamp: maxTimestamp,
	}, nil
}
