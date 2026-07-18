package repository_test

import (
	"context"
	"testing"
	"time"

	"metarang/commercial-service/internal/repository"
	"metarang/commercial-service/tests/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletHistoryRepository_GetCurrentBalance_Empty(t *testing.T) {
	db := testutil.OpenMySQLOrSkip(t)
	defer db.Close()

	repo := repository.NewWalletHistoryRepository(db)
	bal, err := repo.GetCurrentBalance(context.Background(), 999999999)
	require.NoError(t, err)
	assert.Equal(t, 0.0, bal.PSC)
}

func TestWalletHistoryRepository_SumsReturnZeroWithoutRows(t *testing.T) {
	db := testutil.OpenMySQLOrSkip(t)
	defer db.Close()

	repo := repository.NewWalletHistoryRepository(db)
	ctx := context.Background()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	userID := uint64(999999999)

	deposit, err := repo.SumDeposits(ctx, userID, "psc", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, deposit)

	withdraw, err := repo.SumWithdrawals(ctx, userID, "red", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, withdraw)

	hourly, err := repo.SumHourlyProfits(ctx, userID, "blue", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, hourly)

	sells, err := repo.SumTradeSells(ctx, userID, "psc", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, sells)

	buys, err := repo.SumTradeBuys(ctx, userID, "irr", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, buys)

	referral, err := repo.SumReferralBonuses(ctx, userID, start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, referral)

	first, err := repo.SumFirstOrderBonuses(ctx, userID, "psc", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, first)

	level, err := repo.SumLevelRewards(ctx, userID, "psc", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, level)

	feature, err := repo.SumFeaturePurchaseWithdrawals(ctx, userID, "psc", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, feature)

	building, err := repo.SumBuildingSatisfaction(ctx, userID, start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, building)
}

func TestWalletHistoryRepository_TradeAmountIgnoresNonMoneyAssets(t *testing.T) {
	db := testutil.OpenMySQLOrSkip(t)
	defer db.Close()

	repo := repository.NewWalletHistoryRepository(db)
	ctx := context.Background()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	sells, err := repo.SumTradeSells(ctx, 1, "red", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, sells)

	level, err := repo.SumLevelRewards(ctx, 1, "irr", start, end)
	require.NoError(t, err)
	assert.Equal(t, 0.0, level)
}
