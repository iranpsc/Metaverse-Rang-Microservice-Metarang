package handler_test

import (
	"testing"
	"time"

	"metargb/training-service/internal/handler"
	"metargb/training-service/internal/models"
	"metargb/training-service/internal/repository"
	"metargb/training-service/internal/service"
)

func TestVideoDetailsToProto(t *testing.T) {
	t.Setenv("APP_URL", "http://example.com")
	slug := "my-slug"
	d := &service.VideoDetails{
		Video: &models.Video{
			ID:          1,
			Title:       "T",
			Slug:        &slug,
			Description: "D",
			FileName:    "file.mp4",
			CreatorCode: "c",
			Image:       "thumb.jpg",
			CreatedAt:   time.Now(),
		},
		Creator:         &repository.UserBasic{ID: 2, Name: "N", Code: "c"},
		Category:        &models.VideoCategory{ID: 3, Name: "Cat", Slug: "cat"},
		SubCategory:     &models.VideoSubCategory{ID: 4, Name: "Sub", Slug: "sub"},
		Stats:           &models.VideoStats{ViewsCount: 1, LikesCount: 2, DislikesCount: 3, CommentsCount: 4},
		CreatedAtJalali: "1402/01/01",
	}
	p, err := handler.VideoDetailsToProto(d)
	if err != nil {
		t.Fatal(err)
	}
	if p.Id != 1 || p.Slug != slug || p.Stats.ViewsCount != 1 {
		t.Fatalf("proto mismatch %+v", p)
	}
}

func TestVideoDetailsToProto_Nil(t *testing.T) {
	_, err := handler.VideoDetailsToProto(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
