package service

import (
	"context"
	"fmt"
	"strconv"

	"metarang/levels-service/internal/client"
	pb "metarang/shared/pb/levels"
)

func parseNumericString(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse numeric value %q: %w", value, err)
	}
	return parsed, nil
}

func applyLevelPrizeBalances(
	ctx context.Context,
	commercialClient client.CommercialClient,
	userID uint64,
	prize *pb.LevelPrize,
	pscRate float64,
) error {
	pscAmount, err := parseNumericString(prize.Psc)
	if err != nil {
		return err
	}
	if pscRate <= 0 {
		return fmt.Errorf("invalid psc rate: %.4f", pscRate)
	}

	blueAmount, err := parseNumericString(prize.Blue)
	if err != nil {
		return err
	}
	redAmount, err := parseNumericString(prize.Red)
	if err != nil {
		return err
	}
	yellowAmount, err := parseNumericString(prize.Yellow)
	if err != nil {
		return err
	}
	satisfactionAmount, err := parseNumericString(prize.Satisfaction)
	if err != nil {
		return err
	}

	if err := commercialClient.AddBalance(ctx, userID, "psc", pscAmount/pscRate); err != nil {
		return err
	}
	if err := commercialClient.AddBalance(ctx, userID, "blue", blueAmount); err != nil {
		return err
	}
	if err := commercialClient.AddBalance(ctx, userID, "red", redAmount); err != nil {
		return err
	}
	if err := commercialClient.AddBalance(ctx, userID, "yellow", yellowAmount); err != nil {
		return err
	}
	if err := commercialClient.AddBalance(ctx, userID, "effect", float64(prize.Effect)); err != nil {
		return err
	}
	if err := commercialClient.AddBalance(ctx, userID, "satisfaction", satisfactionAmount); err != nil {
		return err
	}

	return nil
}
