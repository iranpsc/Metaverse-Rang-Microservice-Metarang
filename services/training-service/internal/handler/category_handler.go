package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "metargb/shared/pb/common"
	trainingpb "metargb/shared/pb/training"
	"metargb/training-service/internal/models"
	"metargb/training-service/internal/service"
)

type CategoryHandler struct {
	trainingpb.UnimplementedCategoryServiceServer
	categoryService *service.CategoryService
	videoService    *service.VideoService
}

func RegisterCategoryHandler(grpcServer *grpc.Server, categorySvc *service.CategoryService, videoSvc *service.VideoService) {
	handler := &CategoryHandler{
		categoryService: categorySvc,
		videoService:    videoSvc,
	}
	trainingpb.RegisterCategoryServiceServer(grpcServer, handler)
}

// GetCategories retrieves paginated categories
func (h *CategoryHandler) GetCategories(ctx context.Context, req *trainingpb.GetCategoriesRequest) (*trainingpb.CategoriesResponse, error) {
	page := int32(1)
	perPage := int32(30) // Default per API spec

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PerPage > 0 {
			perPage = req.Pagination.PerPage
		}
	}

	categories, total, err := h.categoryService.GetCategories(ctx, page, perPage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get categories: %v", err)
	}

	response := &trainingpb.CategoriesResponse{
		Categories: make([]*trainingpb.CategoryResponse, 0, len(categories)),
		Pagination: &commonpb.PaginationMeta{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			LastPage:    (total + perPage - 1) / perPage,
		},
	}

	for _, category := range categories {
		stats, _ := h.categoryService.GetCategoryStats(ctx, category.ID)
		catResp := buildCategoryProto(category, stats)
		response.Categories = append(response.Categories, catResp)
	}

	return response, nil
}

// GetCategory retrieves a category by slug
func (h *CategoryHandler) GetCategory(ctx context.Context, req *trainingpb.GetCategoryRequest) (*trainingpb.CategoryResponse, error) {
	details, err := h.categoryService.GetCategoryBySlug(ctx, req.Slug)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
	}

	resp := buildCategoryProto(details.Category, details.Stats)

	if len(details.SubCategories) > 0 {
		resp.SubCategories = make([]*trainingpb.SubCategoryInfo, 0, len(details.SubCategories))
		for _, subCat := range details.SubCategories {
			resp.SubCategories = append(resp.SubCategories, &trainingpb.SubCategoryInfo{
				Id:   subCat.ID,
				Name: subCat.Name,
				Slug: subCat.Slug,
			})
		}
	}

	return resp, nil
}

// GetSubCategory retrieves a subcategory by slugs
func (h *CategoryHandler) GetSubCategory(ctx context.Context, req *trainingpb.GetSubCategoryRequest) (*trainingpb.SubCategoryResponse, error) {
	details, err := h.categoryService.GetSubCategoryBySlugs(ctx, req.CategorySlug, req.SubCategorySlug)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "subcategory not found: %v", err)
	}

	return buildSubCategoryProto(details.SubCategory, details.Category, details.Stats), nil
}

// GetCategoryVideos retrieves videos for a category
func (h *CategoryHandler) GetCategoryVideos(ctx context.Context, req *trainingpb.GetCategoryVideosRequest) (*trainingpb.VideosResponse, error) {
	page := int32(1)
	perPage := int32(18) // Default per API spec

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PerPage > 0 {
			perPage = req.Pagination.PerPage
		}
	}

	videos, total, err := h.categoryService.GetCategoryVideos(ctx, req.CategorySlug, page, perPage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get category videos: %v", err)
	}

	response := &trainingpb.VideosResponse{
		Videos: make([]*trainingpb.VideoResponse, 0, len(videos)),
		Pagination: &commonpb.PaginationMeta{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			LastPage:    (total + perPage - 1) / perPage,
		},
	}

	for _, video := range videos {
		details, err := h.videoService.GetVideoWithDetails(ctx, video)
		if err != nil {
			continue
		}
		videoResp, err := buildVideoResponse(details)
		if err != nil {
			continue
		}
		response.Videos = append(response.Videos, videoResp)
	}

	return response, nil
}

func buildCategoryProto(category *models.VideoCategory, stats *models.CategoryStats) *trainingpb.CategoryResponse {
	if category == nil {
		return &trainingpb.CategoryResponse{}
	}
	resp := &trainingpb.CategoryResponse{
		Id:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ImageUrl:    buildUploadURL(category.Image),
	}
	if category.Icon != nil {
		resp.IconUrl = buildUploadURL(*category.Icon)
	}
	if stats != nil {
		resp.VideosCount = stats.VideosCount
	}
	return resp
}

func buildSubCategoryProto(subCategory *models.VideoSubCategory, category *models.VideoCategory, stats *models.SubCategoryStats) *trainingpb.SubCategoryResponse {
	if subCategory == nil {
		return &trainingpb.SubCategoryResponse{}
	}
	resp := &trainingpb.SubCategoryResponse{
		Id:          subCategory.ID,
		Name:        subCategory.Name,
		Slug:        subCategory.Slug,
		Description: subCategory.Description,
		ImageUrl:    buildUploadURL(subCategory.Image),
	}
	if subCategory.Icon != nil {
		resp.IconUrl = buildUploadURL(*subCategory.Icon)
	}
	if category != nil {
		resp.Category = &trainingpb.CategoryInfo{
			Id:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		}
	}
	if stats != nil {
		resp.VideosCount = stats.VideosCount
	}
	return resp
}
