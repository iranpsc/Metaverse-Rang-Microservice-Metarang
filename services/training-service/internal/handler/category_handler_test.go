package handler_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "metargb/shared/pb/common"
	trainingpb "metargb/shared/pb/training"

	"metargb/training-service/internal/handler"
	"metargb/training-service/internal/models"
	"metargb/training-service/internal/service"
	"metargb/training-service/internal/testutil"
)

func TestCategoryHandler_GetCategory_NotFound(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoryBySlugFunc = func(ctx context.Context, slug string) (*models.VideoCategory, error) {
		return nil, nil
	}
	catSvc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	vidSvc := service.NewVideoService(&testutil.MockVideoRepo{}, mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCategoryHandler(s, catSvc, vidSvc)
	})
	defer cleanup()
	client := trainingpb.NewCategoryServiceClient(conn)
	_, err := client.GetCategory(context.Background(), &trainingpb.GetCategoryRequest{Slug: "x"})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("got %v", err)
	}
}

func TestCategoryHandler_GetCategoryVideos_PopulatesVideos(t *testing.T) {
	catID := uint64(7)
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoryBySlugFunc = func(ctx context.Context, slug string) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: catID, Slug: slug}, nil
	}
	mv := &testutil.MockVideoRepo{}
	mv.GetVideosFunc = func(ctx context.Context, page, perPage int32, categoryID, subCategoryID *uint64) ([]*models.Video, int32, error) {
		s := "vslug"
		return []*models.Video{{ID: 100, Title: "V", Slug: &s, CreatedAt: time.Now()}}, 1, nil
	}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{ViewsCount: 1}, nil
	}
	catSvc := service.NewCategoryService(mc, mv)
	vidSvc := service.NewVideoService(mv, mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCategoryHandler(s, catSvc, vidSvc)
	})
	defer cleanup()
	client := trainingpb.NewCategoryServiceClient(conn)
	resp, err := client.GetCategoryVideos(context.Background(), &trainingpb.GetCategoryVideosRequest{
		CategorySlug: "catslug",
		Pagination:   &commonpb.PaginationRequest{Page: 1, PerPage: 18},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Videos) != 1 || resp.Videos[0].Id != 100 {
		t.Fatalf("videos=%+v", resp.Videos)
	}
}

func TestCategoryHandler_GetSubCategory(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetSubCategoryBySlugsFunc = func(ctx context.Context, categorySlug, subCategorySlug string) (*models.VideoSubCategory, error) {
		return &models.VideoSubCategory{ID: 4, Name: "Sub", Slug: subCategorySlug, VideoCategoryID: 2}, nil
	}
	mc.GetCategoryByIDFunc = func(ctx context.Context, categoryID uint64) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: categoryID, Slug: "cslug"}, nil
	}
	mc.GetSubCategoryStatsFunc = func(ctx context.Context, subCategoryID uint64) (*models.SubCategoryStats, error) {
		return &models.SubCategoryStats{VideosCount: 5}, nil
	}
	catSvc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	vidSvc := service.NewVideoService(&testutil.MockVideoRepo{}, mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCategoryHandler(s, catSvc, vidSvc)
	})
	defer cleanup()
	client := trainingpb.NewCategoryServiceClient(conn)
	resp, err := client.GetSubCategory(context.Background(), &trainingpb.GetSubCategoryRequest{
		CategorySlug:    "c",
		SubCategorySlug: "s",
	})
	if err != nil || resp.Name != "Sub" {
		t.Fatal(err, resp)
	}
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoriesFunc = func(ctx context.Context, page, perPage int32) ([]*models.VideoCategory, int32, error) {
		return []*models.VideoCategory{{ID: 1, Name: "C", Slug: "c"}}, 1, nil
	}
	mc.GetCategoryStatsFunc = func(ctx context.Context, categoryID uint64) (*models.CategoryStats, error) {
		return &models.CategoryStats{VideosCount: 2}, nil
	}
	catSvc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	vidSvc := service.NewVideoService(&testutil.MockVideoRepo{}, mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCategoryHandler(s, catSvc, vidSvc)
	})
	defer cleanup()
	client := trainingpb.NewCategoryServiceClient(conn)
	resp, err := client.GetCategories(context.Background(), &trainingpb.GetCategoriesRequest{
		Pagination: &commonpb.PaginationRequest{Page: 1, PerPage: 30},
	})
	if err != nil || len(resp.Categories) != 1 {
		t.Fatal(err, resp)
	}
}
