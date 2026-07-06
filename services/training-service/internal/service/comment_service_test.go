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

func TestCommentService_AddComment_Validation(t *testing.T) {
	svc := service.NewCommentService(&testutil.MockCommentRepo{}, &testutil.MockUserRepo{})
	_, err := svc.AddComment(context.Background(), 1, 1, "")
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	long := make([]byte, 2001)
	for i := range long {
		long[i] = 'a'
	}
	_, err = svc.AddComment(context.Background(), 1, 1, string(long))
	if err == nil {
		t.Fatal("expected error for too long content")
	}
}

func TestCommentService_UpdateComment_Unauthorized(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 5}, nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	_, err := svc.UpdateComment(context.Background(), 1, 99, "hello")
	if err == nil {
		t.Fatal("expected unauthorized")
	}
}

func TestCommentService_AddCommentInteraction_SelfBlocked(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 5}, nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	err := svc.AddCommentInteraction(context.Background(), 1, 5, true, "ip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCommentService_ReportComment_SelfBlocked(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 5}, nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	err := svc.ReportComment(context.Background(), 10, 1, 5, "reason")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCommentService_AddComment_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.AddCommentFunc = func(ctx context.Context, videoID, userID uint64, content string) (*models.Comment, error) {
		return &models.Comment{ID: 9, Content: content}, nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	d, err := svc.AddComment(context.Background(), 1, 2, "hello")
	if err != nil || d.Comment.ID != 9 {
		t.Fatal(err)
	}
}

func TestCommentService_UpdateComment_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	var getCalls int
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		getCalls++
		if getCalls == 1 {
			return &models.Comment{ID: 1, UserID: 5, CreatedAt: time.Now()}, nil
		}
		return &models.Comment{ID: 1, UserID: 5, Content: "updated", CreatedAt: time.Now()}, nil
	}
	mc.UpdateCommentFunc = func(ctx context.Context, commentID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	d, err := svc.UpdateComment(context.Background(), 1, 5, "updated")
	if err != nil || d.Comment.Content != "updated" {
		t.Fatal(err)
	}
}

func TestCommentService_DeleteComment(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.DeleteCommentFunc = func(ctx context.Context, commentID, userID uint64) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	if err := svc.DeleteComment(context.Background(), 3, 9); err != nil {
		t.Fatal(err)
	}
}

func TestCommentService_ReportComment_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 2, UserID: 6}, nil
	}
	mc.ReportCommentFunc = func(ctx context.Context, videoID, commentID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	if err := svc.ReportComment(context.Background(), 10, 2, 7, "spam"); err != nil {
		t.Fatal(err)
	}
}

func TestCommentService_AddCommentInteraction_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 2, UserID: 6}, nil
	}
	mc.AddCommentInteractionFunc = func(ctx context.Context, commentID, userID uint64, liked bool, ipAddress string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	if err := svc.AddCommentInteraction(context.Background(), 2, 8, false, "ip"); err != nil {
		t.Fatal(err)
	}
}

func TestCommentService_GetCommentByID(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 5, UserID: 1, CreatedAt: time.Now()}, nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	d, err := svc.GetCommentByID(context.Background(), 5)
	if err != nil || d.Comment.ID != 5 {
		t.Fatal(err)
	}
}

func TestCommentService_GetComments(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentsFunc = func(ctx context.Context, videoID uint64, page, perPage int32) ([]*models.Comment, int32, error) {
		return []*models.Comment{{ID: 1, UserID: 2, CommentableID: videoID, CreatedAt: time.Now()}}, 1, nil
	}
	mu := &testutil.MockUserRepo{}
	mu.GetUserByIDFunc = func(ctx context.Context, userID uint64) (*repository.UserBasic, error) {
		return &repository.UserBasic{ID: userID, Name: "U"}, nil
	}
	svc := service.NewCommentService(mc, mu)
	list, total, err := svc.GetComments(context.Background(), 9, 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list=%d total=%d err=%v", len(list), total, err)
	}
}
