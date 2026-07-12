package levels_service_test

import (
	"context"
	"testing"

	"metarang/levels-service/internal/mocks"
	"metarang/levels-service/internal/service"
	pb "metarang/shared/pb/levels"
)

func TestChallengeServiceGetQuestion(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			GetRandomUnansweredQuestionFunc: func(ctx context.Context, userID uint64) (*pb.Question, error) {
				return &pb.Question{Id: 11}, nil
			},
			IncrementViewsFunc: func(ctx context.Context, questionID uint64) error { return nil },
		},
		&mocks.MockCommercialClient{},
	)

	question, found, err := svc.GetQuestion(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !found || question == nil || question.Id != 11 {
		t.Fatalf("expected question id 11")
	}
}

func TestChallengeServiceSubmitAnswerCorrect(t *testing.T) {
	addCalls := 0

	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			ValidateAnswerFunc:          func(ctx context.Context, questionID, answerID uint64) (bool, error) { return true, nil },
			HasUserAnsweredQuestionFunc: func(ctx context.Context, userID, questionID uint64) (bool, error) { return false, nil },
			RecordUserAnswerFunc:        func(ctx context.Context, userID, questionID, answerID uint64) error { return nil },
			IncrementParticipantsFunc:   func(ctx context.Context, questionID uint64) error { return nil },
			CheckAnswerFunc:             func(ctx context.Context, answerID, questionID uint64) (bool, string, error) { return true, "1200", nil },
			GetQuestionByIDFunc: func(ctx context.Context, questionID uint64) (*pb.Question, error) {
				return &pb.Question{Id: questionID}, nil
			},
			GetVariableRateFunc: func(ctx context.Context, name string) (float64, error) { return 100, nil },
		},
		&mocks.MockCommercialClient{
			AddBalanceFunc: func(ctx context.Context, userID uint64, asset string, amount float64) error {
				addCalls++
				return nil
			},
		},
	)

	correct, prize, question, err := svc.SubmitAnswer(context.Background(), 1, 2, 3)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !correct || prize != "1200" || question == nil {
		t.Fatalf("expected correct answer with prize")
	}
	if addCalls != 1 {
		t.Fatalf("expected one wallet update, got %d", addCalls)
	}
}

func TestChallengeServiceSubmitAnswerAlreadyAnswered(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			ValidateAnswerFunc:          func(ctx context.Context, questionID, answerID uint64) (bool, error) { return true, nil },
			HasUserAnsweredQuestionFunc: func(ctx context.Context, userID, questionID uint64) (bool, error) { return true, nil },
		},
		&mocks.MockCommercialClient{},
	)

	if _, _, _, err := svc.SubmitAnswer(context.Background(), 1, 2, 3); err == nil {
		t.Fatalf("expected already answered error")
	}
}

func TestChallengeServiceGetTimings(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			GetChallengeIntervalsFunc: func(ctx context.Context) (int32, int32, int32, error) { return 10, 20, 30, nil },
			GetUserAnswerCountsFunc:   func(ctx context.Context, userID uint64) (int32, int32, error) { return 2, 1, nil },
			GetTotalParticipantsFunc:  func(ctx context.Context) (int32, error) { return 9, nil },
		},
		&mocks.MockCommercialClient{},
	)

	resp, err := svc.GetTimings(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.DisplayAdInterval != 10 || resp.CorrectAnswers != 2 || resp.Participants != 9 {
		t.Fatalf("unexpected timings response: %+v", resp)
	}
}

func TestChallengeServiceGetQuestionNotFound(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			GetRandomUnansweredQuestionFunc: func(ctx context.Context, userID uint64) (*pb.Question, error) {
				return nil, nil
			},
		},
		&mocks.MockCommercialClient{},
	)

	q, found, err := svc.GetQuestion(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if found || q != nil {
		t.Fatalf("expected no question")
	}
}

func TestChallengeServiceGetQuestionRepoError(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			GetRandomUnansweredQuestionFunc: func(ctx context.Context, userID uint64) (*pb.Question, error) {
				return nil, assertErr{}
			},
		},
		&mocks.MockCommercialClient{},
	)

	if _, _, err := svc.GetQuestion(context.Background(), 1); err == nil {
		t.Fatalf("expected repo error")
	}
}

func TestChallengeServiceSubmitAnswerInvalidAnswer(t *testing.T) {
	svc := service.NewChallengeService(
		&mocks.MockChallengeRepository{
			ValidateAnswerFunc: func(ctx context.Context, questionID, answerID uint64) (bool, error) { return false, nil },
		},
		&mocks.MockCommercialClient{},
	)

	if _, _, _, err := svc.SubmitAnswer(context.Background(), 1, 2, 3); err == nil {
		t.Fatalf("expected invalid answer error")
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "assert error" }
