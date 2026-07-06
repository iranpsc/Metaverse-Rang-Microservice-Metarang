package handler_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	commonpb "metargb/shared/pb/common"
	trainingpb "metargb/shared/pb/training"

	"metargb/training-service/internal/handler"
	"metargb/training-service/internal/models"
	"metargb/training-service/internal/service"
	"metargb/training-service/internal/testutil"
)

func TestReplyHandler_GetReplies(t *testing.T) {
	pid := uint64(99)
	mc := &testutil.MockCommentRepo{}
	mc.GetRepliesFunc = func(ctx context.Context, commentID uint64, page, perPage int32) ([]*models.Comment, int32, error) {
		return []*models.Comment{{ID: 2, UserID: 3, ParentID: &pid, CommentableID: 1, CreatedAt: time.Now()}}, 1, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterReplyHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewReplyServiceClient(conn)
	resp, err := client.GetReplies(context.Background(), &trainingpb.GetRepliesRequest{
		CommentId: 1,
		Pagination: &commonpb.PaginationRequest{
			Page: 1, PerPage: 10,
		},
	})
	if err != nil || len(resp.Replies) != 1 {
		t.Fatal(err, resp)
	}
}

func TestReplyHandler_AddReply(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 1, UserID: 6, CommentableID: 50}, nil
	}
	mc.AddReplyFunc = func(ctx context.Context, parentCommentID, userID uint64, content string) (*models.Comment, error) {
		return &models.Comment{ID: 12, UserID: userID, Content: content, CreatedAt: time.Now()}, nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterReplyHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewReplyServiceClient(conn)
	resp, err := client.AddReply(context.Background(), &trainingpb.AddReplyRequest{
		ParentCommentId: 1, UserId: 7, Content: "reply",
	})
	if err != nil || resp.Id != 12 {
		t.Fatal(err, resp)
	}
}

func TestReplyHandler_UpdateReply(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	var getCalls int
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		getCalls++
		if getCalls == 1 {
			return &models.Comment{ID: 8, UserID: 9, CreatedAt: time.Now()}, nil
		}
		return &models.Comment{ID: 8, UserID: 9, Content: "z", CreatedAt: time.Now()}, nil
	}
	mc.UpdateReplyFunc = func(ctx context.Context, replyID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterReplyHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewReplyServiceClient(conn)
	_, err := client.UpdateReply(context.Background(), &trainingpb.UpdateReplyRequest{
		ReplyId: 8, UserId: 9, Content: "z",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReplyHandler_DeleteReply(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.DeleteReplyFunc = func(ctx context.Context, replyID, userID uint64) error {
		return nil
	}
	svc := service.NewReplyService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterReplyHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewReplyServiceClient(conn)
	_, err := client.DeleteReply(context.Background(), &trainingpb.DeleteReplyRequest{
		ReplyId: 5, UserId: 9,
	})
	if err != nil {
		t.Fatal(err)
	}
}
