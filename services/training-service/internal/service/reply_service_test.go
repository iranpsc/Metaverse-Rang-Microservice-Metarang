package service_test

import (
	"context"
	"testing"
	"time"

	"metargb/training-service/internal/models"
	"metargb/training-service/internal/service"
	"metargb/training-service/internal/testutil"
)

func TestReplyService_AddReply_SelfBlocked(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 5}, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	_, err := svc.AddReply(context.Background(), 1, 5, "hi")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReplyService_UpdateReply_Unauthorized(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 2, UserID: 8}, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	_, err := svc.UpdateReply(context.Background(), 2, 9, "x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReplyService_AddReplyInteraction_SelfBlocked(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 3, UserID: 1}, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	err := svc.AddReplyInteraction(context.Background(), 3, 1, true, "ip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReplyService_AddReply_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 4, CommentableID: 100}, nil
	}
	mc.AddReplyFunc = func(ctx context.Context, parentCommentID, userID uint64, content string) (*models.Comment, error) {
		return &models.Comment{ID: 8, UserID: userID, Content: content, CreatedAt: time.Now()}, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	d, err := svc.AddReply(context.Background(), 1, 9, "reply text")
	if err != nil || d.Comment.ID != 8 {
		t.Fatal(err)
	}
}

func TestReplyService_UpdateReply_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	var getCalls int
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		getCalls++
		if getCalls == 1 {
			return &models.Comment{ID: 2, UserID: 5, CreatedAt: time.Now()}, nil
		}
		return &models.Comment{ID: 2, UserID: 5, Content: "new", CreatedAt: time.Now()}, nil
	}
	mc.UpdateReplyFunc = func(ctx context.Context, replyID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	d, err := svc.UpdateReply(context.Background(), 2, 5, "new")
	if err != nil || d.Comment.Content != "new" {
		t.Fatal(err)
	}
}

func TestReplyService_DeleteReply(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.DeleteReplyFunc = func(ctx context.Context, replyID, userID uint64) error {
		return nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	if err := svc.DeleteReply(context.Background(), 3, 4); err != nil {
		t.Fatal(err)
	}
}

func TestReplyService_AddReplyInteraction_OK(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 3, UserID: 1}, nil
	}
	mc.AddReplyInteractionFunc = func(ctx context.Context, replyID, userID uint64, liked bool, ipAddress string) error {
		return nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	if err := svc.AddReplyInteraction(context.Background(), 3, 2, true, "ip"); err != nil {
		t.Fatal(err)
	}
}

func TestReplyService_GetReplies(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetRepliesFunc = func(ctx context.Context, commentID uint64, page, perPage int32) ([]*models.Comment, int32, error) {
		return []*models.Comment{{ID: 4, UserID: 2, ParentID: ptrUint64(1)}}, 1, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	list, total, err := svc.GetReplies(context.Background(), 1, 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatal(err, total, len(list))
	}
}

func ptrUint64(v uint64) *uint64 {
	return &v
}
