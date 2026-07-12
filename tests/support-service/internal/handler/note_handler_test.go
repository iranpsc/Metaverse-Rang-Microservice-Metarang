package handler_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "metarang/shared/pb/support"

	"metarang/support-service/internal/handler"
	"metarang/support-service/internal/models"
	"metarang/support-service/internal/service"
	"metarang/support-service/tests/internal/testutil"
)

func TestNoteHandler_CreateNote_Validation(t *testing.T) {
	repo := &testutil.MockNoteRepo{}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	_, err := client.CreateNote(context.Background(), &pb.CreateNoteRequest{UserId: 1, Title: "", Content: "c"})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
}

func TestNoteHandler_CreateNote_Success(t *testing.T) {
	repo := &testutil.MockNoteRepo{
		CreateFunc: func(ctx context.Context, note *models.Note) (*models.Note, error) {
			n := *note
			n.ID = 9
			n.UpdatedAt = time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
			return &n, nil
		},
		GetByIDFunc: func(ctx context.Context, noteID uint64) (*models.Note, error) {
			return &models.Note{
				ID:          noteID,
				UserID:      1,
				Title:       "T",
				Content:     "Body",
				Attachments: []string{"http://x"},
				UpdatedAt:   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	resp, err := client.CreateNote(context.Background(), &pb.CreateNoteRequest{
		UserId:      1,
		Title:       "T",
		Content:     "Body",
		Attachments: []string{"http://x"},
	})
	if err != nil || resp.Id != 9 || resp.Title != "T" {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestNoteHandler_GetNote_PermissionDenied(t *testing.T) {
	repo := &testutil.MockNoteRepo{
		CheckUserOwnershipFunc: func(ctx context.Context, noteID, userID uint64) (bool, error) {
			return false, nil
		},
	}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	_, err := client.GetNote(context.Background(), &pb.GetNoteRequest{NoteId: 1, UserId: 2})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("got %v", err)
	}
}

func TestNoteHandler_GetNotes_Success(t *testing.T) {
	repo := &testutil.MockNoteRepo{
		GetByUserIDFunc: func(ctx context.Context, userID uint64) ([]*models.Note, error) {
			return []*models.Note{{
				ID:          1,
				UserID:      userID,
				Title:       "A",
				Content:     "B",
				Attachments: []string{"x"},
				UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}}, nil
		},
	}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	resp, err := client.GetNotes(context.Background(), &pb.GetNotesRequest{UserId: 3})
	if err != nil || len(resp.Notes) != 1 || resp.Notes[0].Title != "A" {
		t.Fatalf("err=%v notes=%+v", err, resp)
	}
}

func TestNoteHandler_UpdateNote_TooManyAttachments(t *testing.T) {
	repo := &testutil.MockNoteRepo{}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	atts := make([]string, 6)
	for i := range atts {
		atts[i] = "u"
	}
	_, err := client.UpdateNote(context.Background(), &pb.UpdateNoteRequest{
		NoteId: 1, UserId: 1, Title: strings.Repeat("t", 10), Content: "c", Attachments: atts,
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
}

func TestNoteHandler_UpdateNote_Success(t *testing.T) {
	n := &models.Note{ID: 2, UserID: 1, Title: "old", Content: "old", UpdatedAt: time.Now()}
	repo := &testutil.MockNoteRepo{
		CheckUserOwnershipFunc: func(ctx context.Context, noteID, userID uint64) (bool, error) {
			return true, nil
		},
		GetByIDFunc: func(ctx context.Context, noteID uint64) (*models.Note, error) {
			return n, nil
		},
		UpdateFunc: func(ctx context.Context, note *models.Note) error {
			*n = *note
			return nil
		},
	}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	resp, err := client.UpdateNote(context.Background(), &pb.UpdateNoteRequest{
		NoteId: 2, UserId: 1, Title: "new", Content: "nc", Attachments: []string{},
	})
	if err != nil || resp.Title != "new" {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestNoteHandler_DeleteNote_Success(t *testing.T) {
	repo := &testutil.MockNoteRepo{
		CheckUserOwnershipFunc: func(ctx context.Context, noteID, userID uint64) (bool, error) {
			return true, nil
		},
		DeleteFunc: func(ctx context.Context, noteID uint64) error {
			return nil
		},
	}
	svc := service.NewNoteService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterNoteHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewNoteServiceClient(conn)
	_, err := client.DeleteNote(context.Background(), &pb.DeleteNoteRequest{NoteId: 3, UserId: 1})
	if err != nil {
		t.Fatal(err)
	}
}
