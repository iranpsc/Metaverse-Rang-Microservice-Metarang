package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"metarang/features-service/internal/models"

	ptime "github.com/yaa110/go-persian-calendar"
)

var colorAssets = map[string]string{
	"red":    "قرمز",
	"blue":   "آبی",
	"yellow": "زرد",
}

type tradeHistoryFeatureRepo interface {
	FindByID(ctx context.Context, id uint64) (*models.Feature, *models.FeatureProperties, error)
}

type tradeHistoryTradeRepo interface {
	ListByFeatureWithDetails(ctx context.Context, featureID uint64) ([]models.TradeHistoryTrade, error)
	FindSystemUserID(ctx context.Context) (uint64, error)
}

// FeatureTradeHistoryService builds the ownership timeline for a feature.
type FeatureTradeHistoryService struct {
	featureRepo tradeHistoryFeatureRepo
	tradeRepo   tradeHistoryTradeRepo
	now         func() time.Time
}

func NewFeatureTradeHistoryService(
	featureRepo tradeHistoryFeatureRepo,
	tradeRepo tradeHistoryTradeRepo,
) *FeatureTradeHistoryService {
	return &FeatureTradeHistoryService{
		featureRepo: featureRepo,
		tradeRepo:   tradeRepo,
		now:         time.Now,
	}
}

// Paginate returns a page of trade history for the feature owner.
func (s *FeatureTradeHistoryService) Paginate(
	ctx context.Context,
	featureID, requesterID uint64,
	page int,
) (*models.TradeHistoryPage, error) {
	if page < 1 {
		page = 1
	}

	feature, _, err := s.featureRepo.FindByID(ctx, featureID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrFeatureNotFound
		}
		return nil, fmt.Errorf("load feature: %w", err)
	}
	if feature == nil {
		return nil, models.ErrFeatureNotFound
	}
	if feature.OwnerID != requesterID {
		return nil, models.ErrNotFeatureOwner
	}

	systemUserID, err := s.tradeRepo.FindSystemUserID(ctx)
	var systemUserPtr *uint64
	if err == nil {
		systemUserPtr = &systemUserID
	}

	trades, err := s.tradeRepo.ListByFeatureWithDetails(ctx, featureID)
	if err != nil {
		return nil, fmt.Errorf("list trades: %w", err)
	}

	items := make([]models.TradeHistoryItem, 0, len(trades)+1)
	for _, trade := range trades {
		items = append(items, s.transformTrade(trade, systemUserPtr))
	}
	items = append(items, s.buildGenesisEntry(feature))

	total := len(items)
	perPage := models.TradeHistoryPerPage
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	offset := (page - 1) * perPage
	var pageItems []models.TradeHistoryItem
	if offset < total {
		end := offset + perPage
		if end > total {
			end = total
		}
		pageItems = items[offset:end]
	} else {
		pageItems = []models.TradeHistoryItem{}
	}

	result := &models.TradeHistoryPage{
		Items:       pageItems,
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
		Path:        fmt.Sprintf("/api/features/%d/trade-history", featureID),
	}
	if len(pageItems) > 0 {
		from := offset + 1
		to := offset + len(pageItems)
		result.From = &from
		result.To = &to
	}

	return result, nil
}

func (s *FeatureTradeHistoryService) transformTrade(
	trade models.TradeHistoryTrade,
	systemUserID *uint64,
) models.TradeHistoryItem {
	id := trade.ID
	var participantCode *string
	if trade.BuyerCode != "" {
		code := strings.ToUpper(trade.BuyerCode)
		participantCode = &code
	}

	return models.TradeHistoryItem{
		ID:               &id,
		Type:             models.TradeHistoryTypeTrade,
		ParticipantCode:  participantCode,
		ParticipantLabel: trade.BuyerName,
		DateTime:         formatTradeHistoryDateTime(trade.TradeTimestamp(s.now())),
		Price:            s.resolvePrice(trade, systemUserID),
	}
}

func (s *FeatureTradeHistoryService) buildGenesisEntry(feature *models.Feature) models.TradeHistoryItem {
	timestamp := feature.CreatedAt
	if timestamp.IsZero() {
		timestamp = s.now()
	}
	zero := int64(0)
	return models.TradeHistoryItem{
		ID:               nil,
		Type:             models.TradeHistoryTypeGenesis,
		ParticipantCode:  nil,
		ParticipantLabel: models.SystemOwnerLabel,
		DateTime:         formatTradeHistoryDateTime(timestamp),
		Price: models.TradeHistoryPrice{
			Type:     models.TradeHistoryPriceCurrency,
			PricePSC: &zero,
			PriceIRR: &zero,
		},
	}
}

func (s *FeatureTradeHistoryService) resolvePrice(
	trade models.TradeHistoryTrade,
	systemUserID *uint64,
) models.TradeHistoryPrice {
	if s.isSystemPurchase(trade, systemUserID) {
		color, amount := firstColorWithdraw(trade.Transactions)
		var colorPtr, colorNamePtr *string
		var amountPtr *int64
		if color != "" {
			colorPtr = &color
			if name, ok := colorAssets[color]; ok {
				colorNamePtr = &name
			}
		}
		amt := int64(amount)
		amountPtr = &amt
		return models.TradeHistoryPrice{
			Type:        models.TradeHistoryPriceColor,
			Color:       colorPtr,
			ColorName:   colorNamePtr,
			ColorAmount: amountPtr,
		}
	}

	psc := int64(trade.PSCAmount)
	irr := int64(trade.IRRAmount)
	return models.TradeHistoryPrice{
		Type:     models.TradeHistoryPriceCurrency,
		PricePSC: &psc,
		PriceIRR: &irr,
	}
}

func (s *FeatureTradeHistoryService) isSystemPurchase(
	trade models.TradeHistoryTrade,
	systemUserID *uint64,
) bool {
	if systemUserID != nil && trade.SellerID == *systemUserID {
		return true
	}
	if trade.PSCAmount == 0 && trade.IRRAmount == 0 && hasColorWithdraw(trade.Transactions) {
		return true
	}
	return false
}

func hasColorWithdraw(transactions []models.TradeHistoryTransaction) bool {
	color, _ := firstColorWithdraw(transactions)
	return color != ""
}

func firstColorWithdraw(transactions []models.TradeHistoryTransaction) (string, float64) {
	for _, tx := range transactions {
		if tx.Action != "withdraw" {
			continue
		}
		if _, ok := colorAssets[tx.Asset]; ok {
			return tx.Asset, tx.Amount
		}
	}
	return "", 0
}

func formatTradeHistoryDateTime(t time.Time) models.TradeHistoryDateTime {
	pt := ptime.New(t)
	monthName := pt.Month().String()
	year := pt.Year()
	timeStr := t.Format("15:04:05")
	dateStr := pt.Format("yyyy/MM/dd")
	return models.TradeHistoryDateTime{
		Date:      dateStr,
		MonthName: monthName,
		Year:      year,
		Time:      timeStr,
		Formatted: fmt.Sprintf("%s %d | %s", monthName, year, timeStr),
	}
}
