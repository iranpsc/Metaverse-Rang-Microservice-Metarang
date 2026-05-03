package service_test

import (
	"context"
	"testing"
	"time"

	"metargb/training-service/internal/models"
	"metargb/training-service/internal/repository"
	"metargb/training-service/internal/service"
	"metargb/training-service/internal/testutil"
)

func TestVideoService_GetVideos(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideosFunc = func(ctx context.Context, page, perPage int32, categoryID, subCategoryID *uint64) ([]*models.Video, int32, error) {
		return []*models.Video{{ID: 1, Title: "a"}}, 1, nil
	}
	mc := &testutil.MockCategoryRepo{}
	mu := &testutil.MockUserRepo{}
	svc := service.NewVideoService(mv, mc, mu)
	vids, total, err := svc.GetVideos(context.Background(), 1, 18, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(vids) != 1 {
		t.Fatalf("got %d videos total %d", len(vids), total)
	}
}

func TestVideoService_GetVideoBySlug_NotFound(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoBySlugFunc = func(ctx context.Context, slug string) (*models.Video, error) {
		return nil, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	_, err := svc.GetVideoBySlug(context.Background(), "x", nil, "127.0.0.1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVideoService_GetVideoBySlug_IncrementsView(t *testing.T) {
	var incremented bool
	mv := &testutil.MockVideoRepo{}
	slug := "s"
	mv.GetVideoBySlugFunc = func(ctx context.Context, s string) (*models.Video, error) {
		return &models.Video{ID: 10, Slug: &slug, CreatedAt: time.Now()}, nil
	}
	mv.IncrementViewFunc = func(ctx context.Context, videoID uint64, ipAddress string) error {
		incremented = videoID == 10 && ipAddress == "1.2.3.4"
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	v, err := svc.GetVideoBySlug(context.Background(), "s", nil, "1.2.3.4")
	if err != nil || v == nil || !incremented {
		t.Fatalf("err=%v inc=%v", err, incremented)
	}
}

func TestVideoService_SearchVideos_EmptyTerm(t *testing.T) {
	svc := service.NewVideoService(&testutil.MockVideoRepo{}, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	_, _, err := svc.SearchVideos(context.Background(), "", 1, 18)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVideoService_AddInteraction(t *testing.T) {
	var called bool
	mv := &testutil.MockVideoRepo{}
	mv.AddInteractionFunc = func(ctx context.Context, videoID, userID uint64, liked bool, ipAddress string) error {
		called = videoID == 1 && userID == 2 && liked && ipAddress == "ip"
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	if err := svc.AddInteraction(context.Background(), 1, 2, true, "ip"); err != nil || !called {
		t.Fatal(err, called)
	}
}

func TestVideoService_GetVideoBySlug_RepoError(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoBySlugFunc = func(ctx context.Context, slug string) (*models.Video, error) {
		return nil, context.Canceled
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	_, err := svc.GetVideoBySlug(context.Background(), "x", nil, "ip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVideoService_GetVideoByFileName(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoByFileNameFunc = func(ctx context.Context, fileName string) (*models.Video, error) {
		s := "sl"
		return &models.Video{ID: 3, Slug: &s, CreatedAt: time.Now()}, nil
	}
	var incID uint64
	mv.IncrementViewFunc = func(ctx context.Context, videoID uint64, ipAddress string) error {
		incID = videoID
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	v, err := svc.GetVideoByFileName(context.Background(), "frag", "ip")
	if err != nil || v == nil || incID != 3 {
		t.Fatal(err, incID)
	}
}

func TestVideoService_SearchVideos_OK(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.SearchVideosFunc = func(ctx context.Context, searchTerm string, page, perPage int32) ([]*models.Video, int32, error) {
		return []*models.Video{{ID: 1}}, 1, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	list, total, err := svc.SearchVideos(context.Background(), "term", 1, 18)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatal(err, total)
	}
}

func TestVideoService_GetVideoStats(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{ViewsCount: 9}, nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	st, err := svc.GetVideoStats(context.Background(), 1)
	if err != nil || st.ViewsCount != 9 {
		t.Fatal(err, st)
	}
}

func TestVideoService_IncrementView(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.IncrementViewFunc = func(ctx context.Context, videoID uint64, ipAddress string) error {
		return nil
	}
	svc := service.NewVideoService(mv, &testutil.MockCategoryRepo{}, &testutil.MockUserRepo{})
	if err := svc.IncrementView(context.Background(), 4, "ip"); err != nil {
		t.Fatal(err)
	}
}

func TestVideoService_GetVideoWithDetails_NoCategoryWhenSubMissing(t *testing.T) {
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{}, nil
	}
	mc := &testutil.MockCategoryRepo{}
	mc.GetSubCategoryByIDFunc = func(ctx context.Context, subCategoryID uint64) (*models.VideoSubCategory, error) {
		return nil, nil
	}
	svc := service.NewVideoService(mv, mc, &testutil.MockUserRepo{})
	video := &models.Video{ID: 1, CreatedAt: time.Now()}
	d, err := svc.GetVideoWithDetails(context.Background(), video)
	if err != nil || d.Category != nil {
		t.Fatalf("%+v", d)
	}
}

func TestVideoService_GetVideoWithDetails(t *testing.T) {
	code := "c1"
	mv := &testutil.MockVideoRepo{}
	mv.GetVideoStatsFunc = func(ctx context.Context, videoID uint64) (*models.VideoStats, error) {
		return &models.VideoStats{ViewsCount: 3}, nil
	}
	mc := &testutil.MockCategoryRepo{}
	mc.GetSubCategoryByIDFunc = func(ctx context.Context, subCategoryID uint64) (*models.VideoSubCategory, error) {
		return &models.VideoSubCategory{ID: 5, VideoCategoryID: 7}, nil
	}
	mc.GetCategoryByIDFunc = func(ctx context.Context, categoryID uint64) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: 7, Name: "Cat", Slug: "cat"}, nil
	}
	mu := &testutil.MockUserRepo{}
	mu.GetUserBasicByCodeFunc = func(ctx context.Context, code string) (*repository.UserBasic, error) {
		return &repository.UserBasic{ID: 99, Code: code, Name: "Creator"}, nil
	}
	svc := service.NewVideoService(mv, mc, mu)
	video := &models.Video{
		ID:                 1,
		VideoSubCategoryID: 5,
		CreatorCode:        code,
		CreatedAt:          time.Now(),
	}
	d, err := svc.GetVideoWithDetails(context.Background(), video)
	if err != nil || d.Creator == nil || d.Category == nil || d.SubCategory == nil || d.Stats.ViewsCount != 3 {
		t.Fatalf("details=%+v err=%v", d, err)
	}
}
