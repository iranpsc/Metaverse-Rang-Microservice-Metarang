package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"

	commonpb "metarang/shared/pb/common"
	trainingpb "metarang/shared/pb/training"

	"metarang/training-service/internal/handler"
	"metarang/training-service/internal/models"
	"metarang/training-service/internal/repository"
	"metarang/training-service/internal/service"
	"metarang/training-service/internal/testutil"
)

func TestCommentHandler_GetComments(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentsFunc = func(ctx context.Context, videoID uint64, page, perPage int32) ([]*models.Comment, int32, error) {
		return []*models.Comment{{ID: 1, UserID: 2, CommentableID: videoID, CreatedAt: time.Now()}}, 1, nil
	}
	mu := &testutil.MockUserRepo{}
	mu.GetUserByIDFunc = func(ctx context.Context, userID uint64) (*repository.UserBasic, error) {
		return &repository.UserBasic{ID: userID}, nil
	}
	svc := service.NewCommentService(mc, mu)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	resp, err := client.GetComments(context.Background(), &trainingpb.GetCommentsRequest{
		VideoId: 10,
		Pagination: &commonpb.PaginationRequest{
			Page: 1, PerPage: 10,
		},
	})
	if err != nil || len(resp.Comments) != 1 {
		t.Fatal(err, resp)
	}
}

func TestCommentHandler_UpdateComment(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	var getCalls int
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		getCalls++
		if getCalls == 1 {
			return &models.Comment{ID: 1, UserID: 3, CreatedAt: time.Now()}, nil
		}
		return &models.Comment{ID: 1, UserID: 3, Content: "x", CreatedAt: time.Now()}, nil
	}
	mc.UpdateCommentFunc = func(ctx context.Context, commentID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	_, err := client.UpdateComment(context.Background(), &trainingpb.UpdateCommentRequest{
		CommentId: 1, UserId: 3, Content: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.DeleteCommentFunc = func(ctx context.Context, commentID, userID uint64) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	_, err := client.DeleteComment(context.Background(), &trainingpb.DeleteCommentRequest{
		CommentId: 4, UserId: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentHandler_ReportComment(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 2, UserID: 5, CommentableID: 100}, nil
	}
	mc.ReportCommentFunc = func(ctx context.Context, videoID, commentID, userID uint64, content string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	_, err := client.ReportComment(context.Background(), &trainingpb.ReportCommentRequest{
		CommentId: 2, UserId: 9, Content: "report body text here long enough",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentHandler_AddCommentInteraction(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.GetCommentByIDFunc = func(ctx context.Context, commentID uint64) (*models.Comment, error) {
		return &models.Comment{ID: 3, UserID: 1}, nil
	}
	mc.AddCommentInteractionFunc = func(ctx context.Context, commentID, userID uint64, liked bool, ipAddress string) error {
		return nil
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	_, err := client.AddCommentInteraction(context.Background(), &trainingpb.AddCommentInteractionRequest{
		CommentId: 3, UserId: 4, Liked: true, IpAddress: "ip",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommentHandler_AddComment_ServiceError(t *testing.T) {
	mc := &testutil.MockCommentRepo{}
	mc.AddCommentFunc = func(ctx context.Context, videoID, userID uint64, content string) (*models.Comment, error) {
		return nil, errors.New("db")
	}
	svc := service.NewCommentService(mc, &testutil.MockUserRepo{})
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterCommentHandler(s, svc)
	})
	defer cleanup()
	client := trainingpb.NewCommentServiceClient(conn)
	_, err := client.AddComment(context.Background(), &trainingpb.AddCommentRequest{
		VideoId: 1, UserId: 2, Content: "hello",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
