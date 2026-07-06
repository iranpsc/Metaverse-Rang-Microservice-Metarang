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

func TestVideoHandler_GetVideo_NotFound(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoBySlugFunc = func(ctx context.Context, slug string) (*models.Video, error) {
		return nil, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	_, err := client.GetVideo(context.Background(), &trainingpb.GetVideoRequest{Slug: "missing"})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("got %v", err)
	}
}

func TestVideoHandler_SearchVideos_InvalidEmptyQuery(t *testing.T) {
	svc := service.NewVideoService(&testutil.MockVideoRepo{}, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	_, err := client.SearchVideos(context.Background(), &trainingpb.SearchVideosRequest{
		Query: "",
		Pagination: &commonpb.PaginationRequest{
			Page:    1,
			PerPage: 18,
		},
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
}

func TestVideoHandler_GetVideos_OK(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideosFunc = func(ctx context.Context, page, perPage int32, categoryID, subCategoryID *uint64) ([]*models.Video, int32, error) {
		s := "sl"
		return []*models.Video{{ID: 1, Title: "T", Slug: &s, CreatedAt: time.Now()}}, 1, nil
	}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{}, nil
	}
	mc := &testutil.MockCategoryRepo{}
	mc.GetSubCategoryByIDFunc = func(ctx context.Context, subCategoryID uint64) (*models.VideoSubCategory, error) {
		return nil, nil
	}
	svc := service.NewVideoService(mv, mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	resp, err := client.GetVideos(context.Background(), &trainingpb.GetVideosRequest{
		Pagination: &commonpb.PaginationRequest{Page: 1, PerPage: 18},
	})
	if err != nil || len(resp.Videos) != 1 || resp.Pagination.Total != 1 {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestVideoHandler_SearchVideos_OK(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.SearchVideosFunc = func(ctx context.Context, searchTerm string, page, perPage int32) ([]*models.Video, int32, error) {
		s := "sl"
		return []*models.Video{{ID: 1, Title: "T", Slug: &s, CreatedAt: time.Now()}}, 1, nil
	}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{}, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	resp, err := client.SearchVideos(context.Background(), &trainingpb.SearchVideosRequest{
		Query: "hello",
		Pagination: &commonpb.PaginationRequest{
			Page: 1, PerPage: 18,
		},
	})
	if err != nil || len(resp.Videos) != 1 {
		t.Fatal(err, resp)
	}
}

func TestVideoHandler_GetVideoByFileName(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoByFileNameFunc = func(ctx context.Context, fileName string) (*models.Video, error) {
		s := "sl"
		return &models.Video{ID: 7, Slug: &s, CreatedAt: time.Now()}, nil
	}
	mv.IncrementViewFunc = func(ctx context.Context, videoID uint64, ipAddress string) error { return nil }
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{}, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	_, err := client.GetVideoByFileName(context.Background(), &trainingpb.GetVideoByFileNameRequest{
		FileName: "part",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVideoHandler_IncrementView(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.IncrementViewFunc = func(ctx context.Context, videoID uint64, ipAddress string) error {
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	_, err := client.IncrementView(context.Background(), &trainingpb.IncrementViewRequest{
		VideoId: 99, IpAddress: "ip",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVideoHandler_AddInteraction(t *testing.T) {
	var saved bool
	mv := &testutil.MockVideoRepo{}
	mv.AddInteractionFunc = func(ctx context.Context, videoID, userID uint64, liked bool, ipAddress string) error {
		saved = videoID == 9 && userID == 8
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterVideoHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewVideoServiceClient(conn)
	_, err := client.AddInteraction(context.Background(), &trainingpb.AddInteractionRequest{
		VideoId: 9, UserId: 8, Liked: true, IpAddress: "ip",
	})
	if err != nil || !saved {
		t.Fatal(err, saved)
	}
}
