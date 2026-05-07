package levels_service_test

import (
	"context"
	"testing"
	"time"

	"metargb/levels-service/internal/mocks"
	"metargb/levels-service/internal/service"
	pb "metargb/shared/pb/levels"
)

func TestActivityServiceLogActivity(t *testing.T) {
	svc := service.NewActivityService(
		&mocks.MockActivityRepository{
			CreateActivityFunc: func(ctx context.Context, req *pb.LogActivityRequest) (uint64, error) {
				return 12, nil
			},
			CreateUserEventFunc: func(ctx context.Context, userID uint64, event, ip, device string, status int8) error {
				return nil
			},
		},
		&mocks.MockUserLogRepository{},
		&mocks.MockLevelRepository{},
		&mocks.MockCommercialClient{},
	)

	id, err := svc.LogActivity(context.Background(), &pb.LogActivityRequest{UserId: 1, EventType: "login"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != 12 {
		t.Fatalf("expected activity id 12, got %d", id)
	}
}

func TestActivityServiceUpdateActivityScoreLevelUp(t *testing.T) {
	addCalls := 0
	recorded := false
	attached := false

	svc := service.NewActivityService(
		&mocks.MockActivityRepository{
			GetVariableRateFunc: func(ctx context.Context, name string) (float64, error) { return 100, nil },
		},
		&mocks.MockUserLogRepository{
			CalculateScoreFunc: func(ctx context.Context, userID uint64) (int32, error) { return 250, nil },
			UpdateScoreFunc:    func(ctx context.Context, userID uint64, score int32) error { return nil },
		},
		&mocks.MockLevelRepository{
			GetNextLevelForScoreFunc: func(ctx context.Context, userID uint64, score int32) (*pb.Level, error) {
				return &pb.Level{Id: 3}, nil
			},
			AttachLevelToUserFunc: func(ctx context.Context, userID, levelID uint64) error {
				attached = true
				return nil
			},
			GetLevelPrizeFunc: func(ctx context.Context, levelID uint64) (*pb.LevelPrize, error) {
				return &pb.LevelPrize{
					Id:           9,
					Psc:          "1000",
					Blue:         "1",
					Red:          "1",
					Yellow:       "1",
					Effect:       1,
					Satisfaction: "1.0",
				}, nil
			},
			HasUserReceivedPrizeFunc: func(ctx context.Context, userID, prizeID uint64) (bool, error) { return false, nil },
			RecordReceivedPrizeFunc: func(ctx context.Context, userID, prizeID uint64) error {
				recorded = true
				return nil
			},
		},
		&mocks.MockCommercialClient{
			AddBalanceFunc: func(ctx context.Context, userID uint64, asset string, amount float64) error {
				addCalls++
				return nil
			},
		},
	)

	_, levelUp, newLevelID, err := svc.UpdateActivityScore(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !levelUp || newLevelID != 3 {
		t.Fatalf("expected level-up to level 3")
	}
	if !attached || !recorded || addCalls != 6 {
		t.Fatalf("expected attach+record and 6 wallet calls")
	}
}

func TestActivityServiceRecordTrade(t *testing.T) {
	updatedCount := ""
	svc := service.NewActivityService(
		&mocks.MockActivityRepository{
			GetVariableRateFunc: func(ctx context.Context, name string) (float64, error) { return 30000, nil },
			GetSignificantTradeCountFunc: func(ctx context.Context, userID uint64, minIrrAmount, minPscAmount float64) (int32, error) {
				return 5, nil
			},
		},
		&mocks.MockUserLogRepository{
			UpdateTransactionsCountFunc: func(ctx context.Context, userID uint64, count string) error {
				updatedCount = count
				return nil
			},
			CalculateScoreFunc: func(ctx context.Context, userID uint64) (int32, error) { return 10, nil },
			UpdateScoreFunc:    func(ctx context.Context, userID uint64, score int32) error { return nil },
		},
		&mocks.MockLevelRepository{
			GetNextLevelForScoreFunc: func(ctx context.Context, userID uint64, score int32) (*pb.Level, error) { return nil, nil },
		},
		&mocks.MockCommercialClient{},
	)

	if err := svc.RecordTrade(context.Background(), 1, "8000000", "0"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updatedCount != "10" {
		t.Fatalf("expected transactions_count 10, got %s", updatedCount)
	}
}

func TestActivityServiceOtherFlows(t *testing.T) {
	svc := service.NewActivityService(
		&mocks.MockActivityRepository{
			FindByUserIDFunc: func(ctx context.Context, userID uint64, limit int32) ([]*pb.UserActivity, error) {
				return []*pb.UserActivity{{Id: 1}}, nil
			},
			GetLatestActivityFunc: func(ctx context.Context, userID uint64) (*pb.UserActivity, error) {
				return &pb.UserActivity{Id: 2, Start: time.Now().Add(-30 * time.Minute).Format(time.RFC3339)}, nil
			},
			UpdateActivityFunc:          func(ctx context.Context, activityID uint64, endTime time.Time, totalMinutes int32) error { return nil },
			GetTotalActivityMinutesFunc: func(ctx context.Context, userID uint64) (int32, error) { return 60, nil },
		},
		&mocks.MockUserLogRepository{
			GetUserLogFunc:           func(ctx context.Context, userID uint64) (*pb.UserLog, error) { return &pb.UserLog{UserId: userID}, nil },
			IncrementDepositFunc:     func(ctx context.Context, userID uint64, amount string) error { return nil },
			GetTotalFollowersFunc:    func(ctx context.Context, userID uint64) (int32, error) { return 9, nil },
			UpdateFollowersCountFunc: func(ctx context.Context, userID uint64, totalFollowers int32) error { return nil },
			UpdateActivityHoursFunc:  func(ctx context.Context, userID uint64, totalMinutes int32) error { return nil },
			CalculateScoreFunc:       func(ctx context.Context, userID uint64) (int32, error) { return 1, nil },
			UpdateScoreFunc:          func(ctx context.Context, userID uint64, score int32) error { return nil },
		},
		&mocks.MockLevelRepository{
			GetNextLevelForScoreFunc: func(ctx context.Context, userID uint64, score int32) (*pb.Level, error) { return nil, nil },
		},
		&mocks.MockCommercialClient{},
	)

	if _, _, err := svc.GetUserActivities(context.Background(), 1, 5); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := svc.RecordDeposit(context.Background(), 1, "100000"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := svc.RecordFollower(context.Background(), 1); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := svc.LogLogout(context.Background(), 1, "127.0.0.1"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
