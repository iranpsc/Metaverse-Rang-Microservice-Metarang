package service_test

import (
	"context"
	"testing"

	"metarang/training-service/internal/models"
	"metarang/training-service/internal/service"
	"metarang/training-service/internal/testutil"
)

func TestCategoryService_GetCategoryBySlug_NotFound(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoryBySlugFunc = func(ctx context.Context, slug string) (*models.VideoCategory, error) {
		return nil, nil
	}
	svc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	_, err := svc.GetCategoryBySlug(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCategoryService_GetCategoryVideos(t *testing.T) {
	catID := uint64(3)
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoryBySlugFunc = func(ctx context.Context, slug string) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: catID, Slug: slug}, nil
	}
	mv := &testutil.MockVideoRepo{}
	mv.GetVideosFunc = func(ctx context.Context, page, perPage int32, categoryID, subCategoryID *uint64) ([]*models.Video, int32, error) {
		if categoryID == nil || *categoryID != catID {
			t.Fatalf("unexpected category filter %+v", categoryID)
		}
		return []*models.Video{{ID: 1}}, 1, nil
	}
	svc := service.NewCategoryService(mc, mv)
	videos, total, err := svc.GetCategoryVideos(context.Background(), "c", 1, 18)
	if err != nil || total != 1 || len(videos) != 1 {
		t.Fatal(err, total, len(videos))
	}
}

func TestCategoryService_GetCategoryBySlug_OK(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoryBySlugFunc = func(ctx context.Context, slug string) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: 1, Slug: slug}, nil
	}
	mc.GetSubCategoriesByCategoryIDFunc = func(ctx context.Context, categoryID uint64) ([]*models.VideoSubCategory, error) {
		return []*models.VideoSubCategory{{ID: 2, Name: "S"}}, nil
	}
	mc.GetSubCategoryStatsByCategoryIDFunc = func(ctx context.Context, categoryID uint64) (map[uint64]*models.SubCategoryStats, error) {
		return map[uint64]*models.SubCategoryStats{2: {VideosCount: 1}}, nil
	}
	mc.GetCategoryStatsFunc = func(ctx context.Context, categoryID uint64) (*models.CategoryStats, error) {
		return &models.CategoryStats{VideosCount: 3}, nil
	}
	svc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	d, err := svc.GetCategoryBySlug(context.Background(), "cat")
	if err != nil || d.Category.ID != 1 || len(d.SubCategories) != 1 {
		t.Fatal(err)
	}
}

func TestCategoryService_GetSubCategoryBySlugs(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetSubCategoryBySlugsFunc = func(ctx context.Context, categorySlug, subCategorySlug string) (*models.VideoSubCategory, error) {
		return &models.VideoSubCategory{ID: 9, VideoCategoryID: 3, Slug: subCategorySlug}, nil
	}
	mc.GetCategoryByIDFunc = func(ctx context.Context, categoryID uint64) (*models.VideoCategory, error) {
		return &models.VideoCategory{ID: categoryID}, nil
	}
	mc.GetSubCategoryStatsFunc = func(ctx context.Context, subCategoryID uint64) (*models.SubCategoryStats, error) {
		return &models.SubCategoryStats{VideosCount: 2}, nil
	}
	svc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	d, err := svc.GetSubCategoryBySlugs(context.Background(), "c", "s")
	if err != nil || d.SubCategory.ID != 9 {
		t.Fatal(err)
	}
}

func TestCategoryService_GetCategories(t *testing.T) {
	mc := &testutil.MockCategoryRepo{}
	mc.GetCategoriesFunc = func(ctx context.Context, page, perPage int32) ([]*models.VideoCategory, int32, error) {
		return []*models.VideoCategory{{ID: 1, Name: "N"}}, 1, nil
	}
	svc := service.NewCategoryService(mc, &testutil.MockVideoRepo{})
	list, total, err := svc.GetCategories(context.Background(), 1, 30)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatal(err)
	}
}
