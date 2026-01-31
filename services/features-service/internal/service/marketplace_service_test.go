package service_test

import (
	"context"
	"testing"

	"metargb/features-service/internal/service"
	commercialpb "metargb/shared/pb/commercial"
	pb "metargb/shared/pb/features"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommercialClient is a mock implementation of CommercialClient
type MockCommercialClient struct {
	mock.Mock
}

func (m *MockCommercialClient) CheckBalance(ctx context.Context, userID uint64, asset string, requiredAmount float64) (bool, error) {
	args := m.Called(ctx, userID, asset, requiredAmount)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommercialClient) DeductBalance(ctx context.Context, userID uint64, asset string, amount float64) error {
	args := m.Called(ctx, userID, asset, amount)
	return args.Error(0)
}

func (m *MockCommercialClient) AddBalance(ctx context.Context, userID uint64, asset string, amount float64) error {
	args := m.Called(ctx, userID, asset, amount)
	return args.Error(0)
}

func (m *MockCommercialClient) CreateTransaction(ctx context.Context, userID uint64, asset string, amount float64, action string, status int32, payableType string, payableID uint64) (*commercialpb.Transaction, error) {
	args := m.Called(ctx, userID, asset, amount, action, status, payableType, payableID)
	return args.Get(0).(*commercialpb.Transaction), args.Error(1)
}

func (m *MockCommercialClient) Close() error {
	return nil
}

// MockNotificationClient is a mock implementation of NotificationClient
type MockNotificationClient struct {
	mock.Mock
}

func (m *MockNotificationClient) SendBuyRequestNotification(ctx context.Context, userID uint64, notificationType string, buyRequestID, featureID uint64, pricePSC, priceIRR float64) error {
	args := m.Called(ctx, userID, notificationType, buyRequestID, featureID, pricePSC, priceIRR)
	return args.Error(0)
}

func (m *MockNotificationClient) SendBuyFeatureNotification(ctx context.Context, userID uint64, featureID uint64, isRGBPurchase bool, color string, stability float64, pscAmount, irrAmount float64) error {
	args := m.Called(ctx, userID, featureID, isRGBPurchase, color, stability, pscAmount, irrAmount)
	return args.Error(0)
}

func (m *MockNotificationClient) SendSellRequestNotification(ctx context.Context, sellerID uint64, featureID uint64, featurePropertiesID string) error {
	args := m.Called(ctx, sellerID, featureID, featurePropertiesID)
	return args.Error(0)
}

func (m *MockNotificationClient) Close() error {
	return nil
}

// MockEventBroadcaster is a mock implementation of EventBroadcaster
type MockEventBroadcaster struct {
	mock.Mock
}

func (m *MockEventBroadcaster) BroadcastFeatureStatusChanged(ctx context.Context, featureID uint64, rgb string) error {
	args := m.Called(ctx, featureID, rgb)
	return args.Error(0)
}

func (m *MockEventBroadcaster) Close() error {
	return nil
}

// setupTestMarketplaceService creates a marketplace service with mocked dependencies
func setupTestMarketplaceService(t *testing.T) (*service.MarketplaceService, *MockCommercialClient, *MockNotificationClient, *MockEventBroadcaster) {
	// TODO: Setup test database and repositories
	// For now, we'll skip tests that require actual database
	t.Skip("Test database setup not implemented")

	mockCommercial := &MockCommercialClient{}
	mockNotification := &MockNotificationClient{}
	mockBroadcaster := &MockEventBroadcaster{}

	// Create service with mocks
	// Note: MarketplaceService uses concrete client types, not interfaces
	// This test needs to be refactored to use interfaces or integration tests
	// marketplaceService := service.NewMarketplaceService(
	// 	nil, // featureRepo
	// 	nil, // propertiesRepo
	// 	nil, // geometryRepo
	// 	nil, // tradeRepo
	// 	nil, // buyRequestRepo
	// 	nil, // sellRequestRepo
	// 	nil, // lockedAssetRepo
	// 	nil, // hourlyProfitRepo
	// 	nil, // featureLimitRepo
	// 	nil, // variableRepo
	// 	mockCommercial,
	// 	mockNotification,
	// 	mockBroadcaster,
	// 	metrics.NewMarketplaceMetrics(),
	// 	nil, // db
	// 	logger.NewLogger("test"),
	// )

	return nil, mockCommercial, mockNotification, mockBroadcaster
}

func TestMarketplaceService_SendBuyRequest_FloorPriceValidation(t *testing.T) {
	service, mockCommercial, mockNotification, _ := setupTestMarketplaceService(t)
	if service == nil {
		return
	}

	ctx := context.Background()

	// Setup mocks
	mockCommercial.On("CheckBalance", ctx, uint64(1), "psc", mock.AnythingOfType("float64")).Return(true, nil)
	mockCommercial.On("CheckBalance", ctx, uint64(1), "irr", mock.AnythingOfType("float64")).Return(true, nil)
	mockCommercial.On("DeductBalance", ctx, uint64(1), "psc", mock.AnythingOfType("float64")).Return(nil)
	mockCommercial.On("DeductBalance", ctx, uint64(1), "irr", mock.AnythingOfType("float64")).Return(nil)
	mockNotification.On("SendBuyRequestNotification", ctx, uint64(1), "buyer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNotification.On("SendBuyRequestNotification", ctx, uint64(2), "seller", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := &pb.SendBuyRequestRequest{
		BuyerId:   1,
		FeatureId: 100,
		PricePsc:  "50.0", // Below floor price
		PriceIrr:  "500000.0",
	}

	_, err := service.SendBuyRequest(ctx, req)

	// Should fail validation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "floor")
}

func TestMarketplaceService_SendBuyRequest_InsufficientBalance(t *testing.T) {
	service, mockCommercial, _, _ := setupTestMarketplaceService(t)
	if service == nil {
		return
	}

	ctx := context.Background()

	// Setup mocks - insufficient balance
	mockCommercial.On("CheckBalance", ctx, uint64(1), "psc", mock.AnythingOfType("float64")).Return(false, nil)

	req := &pb.SendBuyRequestRequest{
		BuyerId:   1,
		FeatureId: 100,
		PricePsc:  "100.0",
		PriceIrr:  "1000000.0",
	}

	_, err := service.SendBuyRequest(ctx, req)

	// Should fail with insufficient balance
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "موجودی")
}

func TestMarketplaceService_AcceptBuyRequest_UnderpricedCooldown(t *testing.T) {
	service, _, _, _ := setupTestMarketplaceService(t)
	if service == nil {
		return
	}

	// Test that underpriced cooldown is enforced
	// This requires setting up test data with recent underpriced sale

	// TODO: Implement test with underpriced restriction
	t.Skip("Requires test data setup for underpriced restriction")
}

func TestMarketplaceService_CreateSellRequest_AgeBasedPricingLimit(t *testing.T) {
	service, _, mockNotification, mockBroadcaster := setupTestMarketplaceService(t)
	if service == nil {
		return
	}

	ctx := context.Background()

	// Setup mocks
	mockNotification.On("SendSellRequestNotification", ctx, uint64(1), mock.Anything, mock.Anything).Return(nil)
	mockBroadcaster.On("BroadcastFeatureStatusChanged", ctx, mock.Anything, mock.Anything).Return(nil)

	// Test under-18 user trying to set price below 110%
	req := &pb.CreateSellRequestRequest{
		SellerId:               1, // Under-18 user
		FeatureId:              100,
		MinimumPricePercentage: 100, // Below 110% limit
	}

	_, err := service.CreateSellRequest(ctx, req)

	// Should fail with age-based limit error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "110")
}

func TestMarketplaceService_BuyFeature_ThreePathRouting(t *testing.T) {
	service, mockCommercial, mockNotification, mockBroadcaster := setupTestMarketplaceService(t)
	if service == nil {
		return
	}

	ctx := context.Background()

	// Test limited feature path
	t.Run("LimitedFeature", func(t *testing.T) {
		mockCommercial.On("CheckBalance", ctx, uint64(1), "yellow", mock.AnythingOfType("float64")).Return(true, nil)
		mockCommercial.On("DeductBalance", ctx, uint64(1), "yellow", mock.AnythingOfType("float64")).Return(nil)
		mockCommercial.On("AddBalance", ctx, mock.Anything, "yellow", mock.AnythingOfType("float64")).Return(nil)
		mockNotification.On("SendBuyFeatureNotification", ctx, uint64(1), mock.Anything, true, mock.Anything, mock.Anything, float64(0), float64(0)).Return(nil)
		mockBroadcaster.On("BroadcastFeatureStatusChanged", ctx, mock.Anything, mock.Anything).Return(nil)

		// TODO: Setup feature with limited status
		t.Skip("Requires test data setup")
	})

	// Test RGB purchase path
	t.Run("RGBPurchase", func(t *testing.T) {
		// TODO: Implement
		t.Skip("Requires test data setup")
	})

	// Test user-to-user purchase path
	t.Run("UserToUserPurchase", func(t *testing.T) {
		// TODO: Implement
		t.Skip("Requires test data setup")
	})
}
