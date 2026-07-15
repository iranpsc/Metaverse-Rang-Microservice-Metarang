package handler_test

import (
	"context"
	"errors"
	"testing"

	"metarang/features-service/internal/handler"
	"metarang/features-service/internal/models"
	pb "metarang/shared/pb/features"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockTradeHistoryPort struct {
	paginate func(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error)
}

func (m *mockTradeHistoryPort) Paginate(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error) {
	if m.paginate != nil {
		return m.paginate(ctx, featureID, requesterID, page)
	}
	return nil, errors.New("not implemented")
}

func TestFeatureHandler_GetFeatureTradeHistory_Unauthenticated(t *testing.T) {
	h := handler.NewFeatureHandler(&mockFeaturePort{}, &mockTradeHistoryPort{})
	_, err := h.GetFeatureTradeHistory(context.Background(), &pb.GetFeatureTradeHistoryRequest{FeatureId: 1})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestFeatureHandler_GetFeatureTradeHistory_MissingFeatureID(t *testing.T) {
	h := handler.NewFeatureHandler(&mockFeaturePort{}, &mockTradeHistoryPort{})
	ctx := withUserID(context.Background(), 1)
	_, err := h.GetFeatureTradeHistory(ctx, &pb.GetFeatureTradeHistoryRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestFeatureHandler_GetFeatureTradeHistory_NotFound(t *testing.T) {
	m := &mockTradeHistoryPort{}
	m.paginate = func(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error) {
		return nil, models.ErrFeatureNotFound
	}
	h := handler.NewFeatureHandler(&mockFeaturePort{}, m)
	ctx := withUserID(context.Background(), 1)
	_, err := h.GetFeatureTradeHistory(ctx, &pb.GetFeatureTradeHistoryRequest{FeatureId: 9, Page: 1})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestFeatureHandler_GetFeatureTradeHistory_Forbidden(t *testing.T) {
	m := &mockTradeHistoryPort{}
	m.paginate = func(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error) {
		return nil, models.ErrNotFeatureOwner
	}
	h := handler.NewFeatureHandler(&mockFeaturePort{}, m)
	ctx := withUserID(context.Background(), 1)
	_, err := h.GetFeatureTradeHistory(ctx, &pb.GetFeatureTradeHistoryRequest{FeatureId: 9, Page: 1})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestFeatureHandler_GetFeatureTradeHistory_Success(t *testing.T) {
	zero := int64(0)
	code := "HM-2000003"
	id := uint64(42)
	from, to := 1, 1
	m := &mockTradeHistoryPort{}
	m.paginate = func(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error) {
		assert.Equal(t, uint64(10), featureID)
		assert.Equal(t, uint64(7), requesterID)
		assert.Equal(t, 2, page)
		return &models.TradeHistoryPage{
			Items: []models.TradeHistoryItem{
				{
					ID:               &id,
					Type:             models.TradeHistoryTypeTrade,
					ParticipantCode:  &code,
					ParticipantLabel: "کاربر",
					DateTime: models.TradeHistoryDateTime{
						Date:      "1405/02/12",
						MonthName: "اردیبهشت",
						Year:      1405,
						Time:      "12:16:00",
						Formatted: "اردیبهشت 1405 | 12:16:00",
					},
					Price: models.TradeHistoryPrice{
						Type:     models.TradeHistoryPriceCurrency,
						PricePSC: &zero,
						PriceIRR: &zero,
					},
				},
			},
			CurrentPage: 2,
			PerPage:     10,
			Total:       1,
			LastPage:    1,
			From:        &from,
			To:          &to,
			Path:        "/api/features/10/trade-history",
		}, nil
	}

	h := handler.NewFeatureHandler(&mockFeaturePort{}, m)
	ctx := withUserID(context.Background(), 7)
	resp, err := h.GetFeatureTradeHistory(ctx, &pb.GetFeatureTradeHistoryRequest{FeatureId: 10, Page: 2})
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, uint64(42), resp.Data[0].GetId())
	assert.Equal(t, "HM-2000003", resp.Data[0].GetParticipantCode())
	assert.Equal(t, "trade", resp.Data[0].Type)
	assert.Equal(t, int32(2), resp.Meta.CurrentPage)
	assert.Equal(t, "/api/features/10/trade-history?page=1", resp.Links.First)
	assert.Contains(t, resp.Links.Prev, "page=1")
}

func TestFeatureHandler_GetFeatureTradeHistory_DefaultPage(t *testing.T) {
	var gotPage int
	m := &mockTradeHistoryPort{}
	m.paginate = func(ctx context.Context, featureID, requesterID uint64, page int) (*models.TradeHistoryPage, error) {
		gotPage = page
		return &models.TradeHistoryPage{
			Items:       []models.TradeHistoryItem{},
			CurrentPage: 1,
			PerPage:     10,
			Total:       0,
			LastPage:    1,
			Path:        "/api/features/1/trade-history",
		}, nil
	}
	h := handler.NewFeatureHandler(&mockFeaturePort{}, m)
	ctx := withUserID(context.Background(), 1)
	_, err := h.GetFeatureTradeHistory(ctx, &pb.GetFeatureTradeHistoryRequest{FeatureId: 1, Page: 0})
	require.NoError(t, err)
	assert.Equal(t, 1, gotPage)
}
