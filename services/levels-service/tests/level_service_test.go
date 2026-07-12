package levels_service_test

import (
	"context"
	"errors"
	"testing"

	"metarang/levels-service/internal/mocks"
	"metarang/levels-service/internal/service"
	pb "metarang/shared/pb/levels"
)

func TestLevelServiceGetUserLevelNoLevel(t *testing.T) {
	svc := service.NewLevelService(
		&mocks.MockLevelRepository{
			GetUserLatestLevelFunc: func(ctx context.Context, userID uint64) (*pb.Level, error) {
				return nil, errors.New("no level")
			},
		},
		&mocks.MockUserLogRepository{
			GetUserScoreFunc: func(ctx context.Context, userID uint64) (int32, error) { return 0, nil },
		},
		&mocks.MockCommercialClient{},
	)

	resp, err := svc.GetUserLevel(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.LatestLevel != nil {
		t.Fatalf("expected nil latest level")
	}
	if len(resp.PreviousLevels) != 0 {
		t.Fatalf("expected no previous levels")
	}
}

func TestLevelServiceGetUserLevelSuccess(t *testing.T) {
	svc := service.NewLevelService(
		&mocks.MockLevelRepository{
			GetUserLatestLevelFunc: func(ctx context.Context, userID uint64) (*pb.Level, error) {
				return &pb.Level{Id: 2, Score: 100}, nil
			},
			GetLevelsBelowScoreFunc: func(ctx context.Context, score int32) ([]*pb.Level, error) {
				return []*pb.Level{{Id: 1, Score: 50}}, nil
			},
			GetNextLevelFunc: func(ctx context.Context, currentScore int32) (*pb.Level, error) {
				return &pb.Level{Id: 3, Score: 200}, nil
			},
		},
		&mocks.MockUserLogRepository{
			GetUserScoreFunc: func(ctx context.Context, userID uint64) (int32, error) { return 120, nil },
		},
		&mocks.MockCommercialClient{},
	)

	resp, err := svc.GetUserLevel(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.ScorePercentageToNextLevel != 60 {
		t.Fatalf("expected 60 score percentage, got %d", resp.ScorePercentageToNextLevel)
	}
}

func TestLevelServiceClaimPrizeSuccess(t *testing.T) {
	addCalls := 0
	recorded := false

	svc := service.NewLevelService(
		&mocks.MockLevelRepository{
			GetLevelPrizeFunc: func(ctx context.Context, levelID uint64) (*pb.LevelPrize, error) {
				return &pb.LevelPrize{
					Id:           5,
					Psc:          "1000",
					Blue:         "2",
					Red:          "3",
					Yellow:       "4",
					Effect:       1,
					Satisfaction: "1.50",
				}, nil
			},
			HasUserReceivedPrizeFunc: func(ctx context.Context, userID, prizeID uint64) (bool, error) {
				return false, nil
			},
			GetVariableRateFunc: func(ctx context.Context, name string) (float64, error) {
				return 100, nil
			},
			RecordReceivedPrizeFunc: func(ctx context.Context, userID, prizeID uint64) error {
				recorded = true
				return nil
			},
		},
		&mocks.MockUserLogRepository{},
		&mocks.MockCommercialClient{
			AddBalanceFunc: func(ctx context.Context, userID uint64, asset string, amount float64) error {
				addCalls++
				return nil
			},
		},
	)

	if err := svc.ClaimPrize(context.Background(), 10, 2); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if addCalls != 6 {
		t.Fatalf("expected 6 add balance calls, got %d", addCalls)
	}
	if !recorded {
		t.Fatalf("expected prize to be recorded as received")
	}
}

func TestLevelServiceClaimPrizeAlreadyClaimed(t *testing.T) {
	svc := service.NewLevelService(
		&mocks.MockLevelRepository{
			GetLevelPrizeFunc: func(ctx context.Context, levelID uint64) (*pb.LevelPrize, error) {
				return &pb.LevelPrize{Id: 5}, nil
			},
			HasUserReceivedPrizeFunc: func(ctx context.Context, userID, prizeID uint64) (bool, error) {
				return true, nil
			},
		},
		&mocks.MockUserLogRepository{},
		&mocks.MockCommercialClient{},
	)

	if err := svc.ClaimPrize(context.Background(), 10, 2); err == nil {
		t.Fatalf("expected error for already claimed prize")
	}
}

func TestLevelServicePassThroughMethods(t *testing.T) {
	svc := service.NewLevelService(
		&mocks.MockLevelRepository{
			GetAllLevelsFunc: func(ctx context.Context) ([]*pb.Level, error) { return []*pb.Level{{Id: 1}}, nil },
			FindByIDFunc:     func(ctx context.Context, id uint64) (*pb.Level, error) { return &pb.Level{Id: id}, nil },
			FindBySlugFunc:   func(ctx context.Context, slug string) (*pb.Level, error) { return &pb.Level{Id: 2, Slug: slug}, nil },
			GetLevelGeneralInfoFunc: func(ctx context.Context, levelID uint64) (*pb.LevelGeneralInfo, error) {
				return &pb.LevelGeneralInfo{LevelId: levelID}, nil
			},
			GetLevelGemFunc: func(ctx context.Context, levelID uint64) (*pb.LevelGem, error) {
				return &pb.LevelGem{LevelId: levelID}, nil
			},
			GetLevelGiftFunc: func(ctx context.Context, levelID uint64) (*pb.LevelGift, error) {
				return &pb.LevelGift{LevelId: levelID}, nil
			},
			GetLevelLicensesFunc: func(ctx context.Context, levelID uint64) (*pb.LevelLicense, error) {
				return &pb.LevelLicense{LevelId: levelID}, nil
			},
			GetLevelPrizeFunc: func(ctx context.Context, levelID uint64) (*pb.LevelPrize, error) {
				return &pb.LevelPrize{LevelId: levelID}, nil
			},
		},
		&mocks.MockUserLogRepository{},
		&mocks.MockCommercialClient{},
	)

	if _, err := svc.GetAllLevels(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevel(context.Background(), 1, ""); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevel(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevelGeneralInfo(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevelGem(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevelGift(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevelLicenses(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := svc.GetLevelPrizes(context.Background(), 0, "gold"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
