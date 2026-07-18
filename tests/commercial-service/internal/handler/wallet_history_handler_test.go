package handler_test

import (
	"context"
	"testing"

	"metarang/commercial-service/internal/handler"
	"metarang/commercial-service/internal/models"
	"metarang/commercial-service/internal/service"
	pb "metarang/shared/pb/commercial"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockWalletHistoryService struct {
	summary *service.WalletHistorySummaryResult
	chart   *service.WalletHistoryChartResult
	err     error
}

func (m *mockWalletHistoryService) GetSummary(ctx context.Context, userID uint64, periodStr string, assets []string, privacy map[string]int32) (*service.WalletHistorySummaryResult, error) {
	return m.summary, m.err
}

func (m *mockWalletHistoryService) GetChart(ctx context.Context, userID uint64, periodStr string, assets []string, privacy map[string]int32) (*service.WalletHistoryChartResult, error) {
	return m.chart, m.err
}

func TestWalletHistoryHandler_SummaryRequiresUserID(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{})
	_, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_SummaryRequiresPeriod(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{})
	_, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_ChartRequiresPeriod(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{})
	_, err := h.GetWalletHistoryChart(context.Background(), &pb.GetWalletHistoryChartRequest{
		UserId: 1,
		Period: "",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_SummaryInvalidPeriod(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{})
	_, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "hourly",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_SummaryHappyPath(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{
		summary: &service.WalletHistorySummaryResult{
			Period: "daily",
			Cards: []models.WalletHistorySummaryCard{
				{Asset: "psc", CurrentBalance: 10, PeriodIncome: 2, PeriodSpending: 1, GrowthPercent: 50, Direction: "up"},
			},
		},
	})
	resp, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "daily",
	})
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "psc", resp.Data[0].Asset)
	assert.Equal(t, 10.0, resp.Data[0].CurrentBalance)
}

func TestWalletHistoryHandler_ChartHappyPath(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{
		chart: &service.WalletHistoryChartResult{
			Period: "weekly",
			Charts: map[string]models.WalletAssetChart{
				"psc": {
					Income:   []models.WalletChartPoint{{Label: "a", Amount: 1}},
					Spending: []models.WalletChartPoint{{Label: "a", Amount: 0}},
				},
			},
		},
	})
	resp, err := h.GetWalletHistoryChart(context.Background(), &pb.GetWalletHistoryChartRequest{
		UserId: 1,
		Period: "weekly",
	})
	require.NoError(t, err)
	require.Contains(t, resp.Data, "psc")
	assert.Equal(t, 1.0, resp.Data["psc"].Income[0].Amount)
}

func TestWalletHistoryHandler_SummaryServiceError(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{err: assert.AnError})
	_, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "daily",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestWalletHistoryHandler_ChartInvalidPeriod(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{})
	_, err := h.GetWalletHistoryChart(context.Background(), &pb.GetWalletHistoryChartRequest{
		UserId: 1,
		Period: "bad",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_SummaryPrivacyRestrictedCard(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{
		summary: &service.WalletHistorySummaryResult{
			Period: "daily",
			Cards: []models.WalletHistorySummaryCard{
				{Asset: "irr", PrivacyRestricted: true},
			},
		},
	})
	resp, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "daily",
	})
	require.NoError(t, err)
	require.True(t, resp.Data[0].PrivacyRestricted)
}

func TestWalletHistoryHandler_ChartServiceError(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{err: assert.AnError})
	_, err := h.GetWalletHistoryChart(context.Background(), &pb.GetWalletHistoryChartRequest{
		UserId: 1,
		Period: "daily",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestWalletHistoryHandler_SummaryInvalidPeriodFromService(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{err: errInvalidPeriod{}})
	_, err := h.GetWalletHistorySummary(context.Background(), &pb.GetWalletHistorySummaryRequest{
		UserId: 1,
		Period: "daily",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestWalletHistoryHandler_ChartInvalidPeriodFromService(t *testing.T) {
	h := handler.NewWalletHistoryHandler(&mockWalletHistoryService{err: errInvalidPeriod{}})
	_, err := h.GetWalletHistoryChart(context.Background(), &pb.GetWalletHistoryChartRequest{
		UserId: 1,
		Period: "daily",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

type errInvalidPeriod struct{}

func (errInvalidPeriod) Error() string { return "invalid period [x] provided" }
