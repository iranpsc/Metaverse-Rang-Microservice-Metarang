package service

import (
	"context"
	"testing"

	"metargb/support-service/internal/models"
	"metargb/support-service/internal/testutil"
)

func TestReportService_CreateGetList(t *testing.T) {
	var lastID uint64 = 5
	repo := &testutil.MockReportRepo{
		CreateFunc: func(ctx context.Context, report *models.Report) (*models.Report, error) {
			r := *report
			r.ID = lastID
			return &r, nil
		},
		CreateImageFunc: func(ctx context.Context, reportID uint64, url string) error {
			return nil
		},
		GetByIDFunc: func(ctx context.Context, reportID uint64) (*models.ReportWithImages, error) {
			return &models.ReportWithImages{
				Report: models.Report{
					ID:      reportID,
					UserID:  9,
					Subject: "displayError",
					Title:   "t",
					Content: "c",
					URL:     "https://ex.com",
				},
				Images: nil,
			}, nil
		},
		GetByUserIDFunc: func(ctx context.Context, userID uint64, page, perPage int32) ([]*models.Report, int, error) {
			return []*models.Report{{ID: 1, UserID: userID}}, 1, nil
		},
	}
	svc := NewReportService(repo)
	full, err := svc.CreateReport(context.Background(), 9, "displayError", "t", "c", "https://ex.com", []string{"a.png"})
	if err != nil || full.Report.URL != "https://ex.com" {
		t.Fatalf("err=%v full=%+v", err, full)
	}
	list, total, err := svc.GetReports(context.Background(), 9, 1, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list err=%v total=%d n=%d", err, total, len(list))
	}
}

func TestReportService_GetReportOwnership(t *testing.T) {
	repo := &testutil.MockReportRepo{
		GetByIDFunc: func(ctx context.Context, reportID uint64) (*models.ReportWithImages, error) {
			return &models.ReportWithImages{
				Report: models.Report{ID: 1, UserID: 99},
			}, nil
		},
	}
	svc := NewReportService(repo)
	_, err := svc.GetReport(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected unauthorized")
	}
	got, err := svc.GetReport(context.Background(), 1, 99)
	if err != nil || got == nil {
		t.Fatalf("err=%v got=%v", err, got)
	}
}
