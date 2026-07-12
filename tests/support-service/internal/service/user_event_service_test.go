package service_test

import (
	"context"
	"testing"

	"metarang/support-service/internal/models"
	"metarang/support-service/internal/service"
	"metarang/support-service/tests/internal/testutil"
)

func TestUserEventService_CreateAndList(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		CreateFunc: func(ctx context.Context, event *models.UserEvent) (*models.UserEvent, error) {
			e := *event
			e.ID = 50
			return &e, nil
		},
		GetByUserIDFunc: func(ctx context.Context, userID uint64, page, perPage int32) ([]*models.UserEvent, int, error) {
			return []*models.UserEvent{{ID: 50, UserID: userID, Event: "login"}}, 1, nil
		},
	}
	svc := service.NewUserEventService(repo)
	ev, err := svc.CreateUserEvent(context.Background(), 3, "login", "d", "2024-01-01")
	if err != nil || ev.ID != 50 {
		t.Fatalf("create err=%v ev=%+v", err, ev)
	}
	list, total, err := svc.GetUserEvents(context.Background(), 3, 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list err=%v total=%d n=%d", err, total, len(list))
	}
}

func TestUserEventService_ReportUserEvent(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		CreateReportFunc: func(ctx context.Context, report *models.UserEventReport) (*models.UserEventReport, error) {
			r := *report
			r.ID = 88
			return &r, nil
		},
	}
	svc := service.NewUserEventService(repo)
	rep, err := svc.ReportUserEvent(context.Background(), 10, "citizen", "desc")
	if err != nil || rep.ID != 88 {
		t.Fatalf("err=%v rep=%+v", err, rep)
	}
}

func TestUserEventService_GetUserEventScoped(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return &models.UserEventWithReport{
				UserEvent: models.UserEvent{ID: 1, UserID: 42, Event: "login"},
			}, nil
		},
	}
	svc := service.NewUserEventService(repo)
	_, err := svc.GetUserEvent(context.Background(), 1, 99)
	if err == nil {
		t.Fatal("expected unauthorized")
	}
	ev, err := svc.GetUserEvent(context.Background(), 1, 42)
	if err != nil || ev.Event != "login" {
		t.Fatalf("err=%v ev=%+v", err, ev)
	}
}

func TestUserEventService_SendEventReportResponse(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetReportByEventIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventReport, error) {
			return &models.UserEventReport{ID: 7, UserEventID: eventID}, nil
		},
		CreateReportResponseFunc: func(ctx context.Context, response *models.UserEventReportResponse) (*models.UserEventReportResponse, error) {
			response.ID = 100
			return response, nil
		},
		UpdateReportStatusFunc: func(ctx context.Context, reportID uint64, status int32) error {
			return nil
		},
	}
	svc := service.NewUserEventService(repo)
	created, err := svc.SendEventReportResponse(context.Background(), 3, "admin", "hello")
	if err != nil || created.ID != 100 {
		t.Fatalf("err=%v created=%+v", err, created)
	}
}

func TestUserEventService_CloseUserEventReport(t *testing.T) {
	repo := &testutil.MockUserEventRepo{
		GetByIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
			return &models.UserEventWithReport{
				UserEvent: models.UserEvent{ID: 1, UserID: 5},
			}, nil
		},
		GetReportByEventIDFunc: func(ctx context.Context, eventID uint64) (*models.UserEventReport, error) {
			return &models.UserEventReport{ID: 9, UserEventID: 1}, nil
		},
		CloseReportFunc: func(ctx context.Context, reportID uint64) error {
			if reportID != 9 {
				t.Fatalf("report id %d", reportID)
			}
			return nil
		},
	}
	svc := service.NewUserEventService(repo)
	if err := svc.CloseUserEventReport(context.Background(), 1, 5); err != nil {
		t.Fatal(err)
	}
}
