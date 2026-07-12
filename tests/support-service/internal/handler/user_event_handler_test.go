package handler_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbCommon "metarang/shared/pb/common"
	pb "metarang/shared/pb/support"

	"metarang/support-service/internal/handler"
	"metarang/support-service/internal/models"
	"metarang/support-service/internal/service"
	"metarang/support-service/tests/internal/testutil"
)

func TestUserEventHandler_CreateUserEvent_Success(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		CreateFunc: func(ctx context.Context, event *models.UserEvent) (*models.UserEvent, error) {
			e := *event
			e.ID = 60
			e.CreatedAt = time.Now()
			e.UpdatedAt = e.CreatedAt
			return &e, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	resp, err := client.CreateUserEvent(context.Background(), &pb.CreateUserEventRequest{
		UserId: 4, Title: "signup", Description: "d", EventDate: "2024-01-01",
	})
	if err != nil || resp.Id != 60 || resp.Title != "signup" {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestUserEventHandler_CreateUserEvent_Validation(t *testing.T) {
	repo := &testutil.MockUserEventRepo{}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	_, err := client.CreateUserEvent(context.Background(), &pb.CreateUserEventRequest{UserId: 1, Title: ""})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
}

func TestUserEventHandler_GetUserEvent_NotFound(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return nil, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	_, err := client.GetUserEvent(context.Background(), &pb.GetUserEventRequest{EventId: 1, UserId: 1})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("got %v", err)
	}
}

func TestUserEventHandler_GetUserEvent_WithReport(t *testing.T) {
	cit := "c1"
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return &models.UserEventWithReport{
				UserEvent: models.UserEvent{
					ID:        1,
					UserID:    2,
					Event:     "login",
					IP:        "1.1.1.1",
					Device:    "ios",
					Status:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Report: &models.UserEventReport{
					ID:                5,
					UserEventID:       1,
					SuspeciousCitizen: &cit,
					EventDescription:  "suspicious",
					Status:            0,
					Closed:            false,
					CreatedAt:         time.Now(),
					UpdatedAt:         time.Now(),
				},
				Responses: []models.UserEventReportResponse{{
					ID:                9,
					UserEventReportID: 5,
					Response:          "we will check",
					ResponserName:     "Admin",
					CreatedAt:         time.Now(),
					UpdatedAt:         time.Now(),
				}},
			}, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	resp, err := client.GetUserEvent(context.Background(), &pb.GetUserEventRequest{EventId: 1, UserId: 2})
	if err != nil || resp.Report == nil || resp.Report.SuspiciousCitizen != cit {
		t.Fatalf("err=%v report=%+v", err, resp.GetReport())
	}
}

func TestUserEventHandler_ReportUserEvent_Success(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		CreateReportFunc: func(ctx context.Context, report *models.UserEventReport) (*models.UserEventReport, error) {
			r := *report
			r.ID = 77
			r.CreatedAt = time.Now()
			r.UpdatedAt = r.CreatedAt
			return &r, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	resp, err := client.ReportUserEvent(context.Background(), &pb.ReportUserEventRequest{
		EventId: 3, SuspiciousCitizen: "u1", EventDescription: "bad",
	})
	if err != nil || resp.Id != 77 {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestUserEventHandler_GetUserEvent_PermissionDenied(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return &models.UserEventWithReport{
				UserEvent: models.UserEvent{ID: 1, UserID: 5, Event: "e"},
			}, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	_, err := client.GetUserEvent(context.Background(), &pb.GetUserEventRequest{EventId: 1, UserId: 9})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("got %v", err)
	}
}

func TestUserEventHandler_SendEventReportResponse_Success(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetReportByEventIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventReport, error) {
			return &models.UserEventReport{ID: 20, UserEventID: eventID}, nil
		},
		CreateReportResponseFunc: func(ctx context.Context, response *models.UserEventReportResponse) (*models.UserEventReportResponse, error) {
			r := *response
			r.ID = 300
			r.CreatedAt = time.Date(2025, 1, 2, 15, 0, 0, 0, time.UTC)
			return &r, nil
		},
		UpdateReportStatusFunc: func(ctx context.Context, reportID uint64, status int32) error {
			return nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	resp, err := client.SendEventReportResponse(context.Background(), &pb.SendEventReportResponseRequest{
		EventId:       8,
		Response:      "ok",
		ResponderName: "Admin",
	})
	if err != nil || resp.Id != 300 || resp.Response != "ok" {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestUserEventHandler_CloseUserEventReport_Success(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return &models.UserEventWithReport{UserEvent: models.UserEvent{ID: 1, UserID: 7}}, nil
		},
		GetReportByEventIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventReport, error) {
			return &models.UserEventReport{ID: 55, UserEventID: 1}, nil
		},
		CloseReportFunc: func(ctx context.Context, reportID uint64) error {
			return nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	_, err := client.CloseUserEventReport(context.Background(), &pb.CloseUserEventReportRequest{EventId: 1, UserId: 7})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserEventHandler_GetUserEvents_Success(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByUserIDFunc: func(ctx context.Context, userID uint64, page, perPage int32) ([]*models.UserEvent, int, error) {
			return []*models.UserEvent{{ID: 2, UserID: userID, Event: "login"}}, 1, nil
		},
	}
	svc := service.NewUserEventService(repo)
	conn, cleanup := testutil.DialBufConn(func(s *grpc.Server) {
		handler.RegisterUserEventHandler(s, svc)
	})
	defer cleanup()
	client := pb.NewUserEventReportServiceClient(conn)
	resp, err := client.GetUserEvents(context.Background(), &pb.GetUserEventsRequest{
		UserId:     4,
		Pagination: &pbCommon.PaginationRequest{Page: 1, PerPage: 10},
	})
	if err != nil || len(resp.Events) != 1 {
		t.Fatalf("err=%v", err)
	}
}
