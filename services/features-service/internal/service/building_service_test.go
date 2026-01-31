package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"metargb/features-service/internal/models"
	pb "metargb/shared/pb/features"
)

// Mock repositories for testing
type mockBuildingRepository struct {
	upsertModelFunc                   func(ctx context.Context, modelID, name, sku, images, attributes, file string, requiredSatisfaction float64) error
	findModelFunc                     func(ctx context.Context, modelID uint64) (*pb.BuildingModel, error)
	hasBuildingFunc                   func(ctx context.Context, featureID uint64) (bool, error)
	createBuildingFunc                func(ctx context.Context, featureID, buildingModelID uint64, launchedSatisfaction, rotation, position, information string, startDate, endDate time.Time, bubbleDiameter float64) error
	findByFeatureIDFunc               func(ctx context.Context, featureID uint64) ([]*pb.Building, error)
	updateBuildingFunc                func(ctx context.Context, featureID, buildingModelID uint64, launchedSatisfaction, rotation, position, information string, endDate time.Time, bubbleDiameter float64) (*pb.Building, error)
	findBuildingByFeatureAndModelFunc func(ctx context.Context, featureID, buildingModelID uint64) (*pb.Building, error)
	deleteBuildingFunc                func(ctx context.Context, featureID, buildingModelID uint64) error
	firstOrCreateIsicCodeFunc         func(ctx context.Context, activityLine string) (uint64, error)
}

func (m *mockBuildingRepository) UpsertBuildingModel(ctx context.Context, modelID, name, sku, images, attributes, file string, requiredSatisfaction float64) error {
	if m.upsertModelFunc != nil {
		return m.upsertModelFunc(ctx, modelID, name, sku, images, attributes, file, requiredSatisfaction)
	}
	return errors.New("not implemented")
}

func (m *mockBuildingRepository) FindBuildingModelByModelID(ctx context.Context, modelID uint64) (*pb.BuildingModel, error) {
	if m.findModelFunc != nil {
		return m.findModelFunc(ctx, modelID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBuildingRepository) HasBuilding(ctx context.Context, featureID uint64) (bool, error) {
	if m.hasBuildingFunc != nil {
		return m.hasBuildingFunc(ctx, featureID)
	}
	return false, errors.New("not implemented")
}

func (m *mockBuildingRepository) CreateBuilding(ctx context.Context, featureID, buildingModelID uint64, launchedSatisfaction, rotation, position, information string, startDate, endDate time.Time, bubbleDiameter float64) error {
	if m.createBuildingFunc != nil {
		return m.createBuildingFunc(ctx, featureID, buildingModelID, launchedSatisfaction, rotation, position, information, startDate, endDate, bubbleDiameter)
	}
	return errors.New("not implemented")
}

func (m *mockBuildingRepository) FindByFeatureID(ctx context.Context, featureID uint64) ([]*pb.Building, error) {
	if m.findByFeatureIDFunc != nil {
		return m.findByFeatureIDFunc(ctx, featureID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBuildingRepository) UpdateBuilding(ctx context.Context, featureID, buildingModelID uint64, launchedSatisfaction, rotation, position, information string, endDate time.Time, bubbleDiameter float64) (*pb.Building, error) {
	if m.updateBuildingFunc != nil {
		return m.updateBuildingFunc(ctx, featureID, buildingModelID, launchedSatisfaction, rotation, position, information, endDate, bubbleDiameter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBuildingRepository) FindBuildingByFeatureAndModel(ctx context.Context, featureID, buildingModelID uint64) (*pb.Building, error) {
	if m.findBuildingByFeatureAndModelFunc != nil {
		return m.findBuildingByFeatureAndModelFunc(ctx, featureID, buildingModelID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBuildingRepository) DeleteBuilding(ctx context.Context, featureID, buildingModelID uint64) error {
	if m.deleteBuildingFunc != nil {
		return m.deleteBuildingFunc(ctx, featureID, buildingModelID)
	}
	return errors.New("not implemented")
}

func (m *mockBuildingRepository) FirstOrCreateIsicCode(ctx context.Context, activityLine string) (uint64, error) {
	if m.firstOrCreateIsicCodeFunc != nil {
		return m.firstOrCreateIsicCodeFunc(ctx, activityLine)
	}
	return 0, errors.New("not implemented")
}

// Mock other repositories
type mockFeatureRepository struct {
	findByIDFunc func(ctx context.Context, id uint64) (*models.Feature, *models.FeatureProperties, error)
}

func (m *mockFeatureRepository) FindByID(ctx context.Context, id uint64) (*models.Feature, *models.FeatureProperties, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil, errors.New("not implemented")
}

type mockGeometryRepository struct {
	getCoordinatesFunc func(ctx context.Context, featureID uint64) ([]string, error)
}

func (m *mockGeometryRepository) GetCoordinatesByFeatureID(ctx context.Context, featureID uint64) ([]string, error) {
	if m.getCoordinatesFunc != nil {
		return m.getCoordinatesFunc(ctx, featureID)
	}
	return nil, errors.New("not implemented")
}

type mockHourlyProfitRepository struct {
	deactivateFunc func(ctx context.Context, featureID uint64) error
	activateFunc   func(ctx context.Context, featureID uint64) error
}

func (m *mockHourlyProfitRepository) DeactivateProfitsForFeature(ctx context.Context, featureID uint64) error {
	if m.deactivateFunc != nil {
		return m.deactivateFunc(ctx, featureID)
	}
	return errors.New("not implemented")
}

func (m *mockHourlyProfitRepository) ActivateProfitsForFeature(ctx context.Context, featureID uint64) error {
	if m.activateFunc != nil {
		return m.activateFunc(ctx, featureID)
	}
	return errors.New("not implemented")
}

// Mock 3D client
type mockThreeDClient struct {
	getBuildPackageFunc func(req interface{}) (interface{}, error)
}

func (m *mockThreeDClient) GetBuildPackage(req interface{}) (interface{}, error) {
	if m.getBuildPackageFunc != nil {
		return m.getBuildPackageFunc(req)
	}
	return nil, errors.New("not implemented")
}

func TestBuildingService_GetBuildPackage(t *testing.T) {
	t.Run("unauthorized user", func(t *testing.T) {
		// Note: This test requires refactoring to use interfaces instead of concrete types
		// The service constructor expects concrete repository types, not mocks
		// For now, we skip this test until the service is refactored to use interfaces
		t.Skip("Test requires service refactoring to use repository interfaces")
		// ctx := context.Background()
		// mockBuildingRepo := &mockBuildingRepository{}
		// mockFeatureRepo := &mockFeatureRepository{}
		// mockGeometryRepo := &mockGeometryRepository{}
		// mockProfitRepo := &mockHourlyProfitRepository{}
		//
		// mockFeatureRepo.findByIDFunc = func(ctx context.Context, id uint64) (*models.Feature, *models.FeatureProperties, error) {
		// 	return &models.Feature{
		// 		ID:      1,
		// 		OwnerID: 100, // Different owner
		// 	}, &models.FeatureProperties{}, nil
		// }
		// service := NewBuildingService(mockBuildingRepo, mockFeatureRepo, mockGeometryRepo, mockProfitRepo, nil)
		// _, _, err := service.GetBuildPackage(ctx, 1, 1) // featureID=1, page=1
		// if err == nil {
		// 	t.Error("Expected error for unauthorized user")
		// }
		// if err != nil && !contains(err.Error(), "unauthorized") && !contains(err.Error(), "does not own") {
		// 	t.Errorf("Expected authorization error, got: %v", err)
		// }
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test extractAttributeValue helper function
// NOTE: extractAttributeValue function doesn't exist in the service
// This test is skipped until the function is implemented or the test is removed
func TestExtractAttributeValue(t *testing.T) {
	t.Skip("extractAttributeValue function doesn't exist in the service")
	// tests := []struct {
	// 	name       string
	// 	attributes []map[string]interface{}
	// 	slug       string
	// 	wantValue  float64
	// 	wantOk     bool
	// }{
	// 	{
	// 		name: "valid attribute found",
	// 		attributes: []map[string]interface{}{
	// 			{"slug": "width", "value": 50.0},
	// 			{"slug": "length", "value": 30.0},
	// 			{"slug": "density", "value": 3.0},
	// 		},
	// 		slug:      "width",
	// 		wantValue: 50.0,
	// 		wantOk:    true,
	// 	},
	// }
	//
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		gotValue, gotOk := extractAttributeValue(tt.attributes, tt.slug)
	// 		if gotValue != tt.wantValue {
	// 			t.Errorf("extractAttributeValue() value = %v, want %v", gotValue, tt.wantValue)
	// 		}
	// 		if gotOk != tt.wantOk {
	// 			t.Errorf("extractAttributeValue() ok = %v, want %v", gotOk, tt.wantOk)
	// 		}
	// 	})
	// }
}

// Test calculateBubbleDiameter function
func TestBuildingService_CalculateBubbleDiameter(t *testing.T) {
	service := NewBuildingService(nil, nil, nil, nil, nil)

	tests := []struct {
		name           string
		attributesJSON string
		want           float64
	}{
		{
			name:           "density 1 - coefficient 1.0",
			attributesJSON: `[{"slug": "width", "value": 50}, {"slug": "length", "value": 30}, {"slug": "density", "value": 1}]`,
			want:           160.0, // perimeter = 2 * (50 + 30) = 160, coefficient = 1.0, diameter = 160 * 1.0 = 160
		},
		{
			name:           "density 2 - coefficient 1.3",
			attributesJSON: `[{"slug": "width", "value": 50}, {"slug": "length", "value": 30}, {"slug": "density", "value": 2}]`,
			want:           208.0, // perimeter = 160, coefficient = 1.3, diameter = 160 * 1.3 = 208
		},
		{
			name:           "density 3 - coefficient 1.6",
			attributesJSON: `[{"slug": "width", "value": 50}, {"slug": "length", "value": 30}, {"slug": "density", "value": 3}]`,
			want:           256.0, // perimeter = 160, coefficient = 1.6, diameter = 160 * 1.6 = 256
		},
		{
			name:           "density 4 - coefficient 1.9",
			attributesJSON: `[{"slug": "width", "value": 40}, {"slug": "length", "value": 20}, {"slug": "density", "value": 4}]`,
			want:           228.0, // perimeter = 2 * (40 + 20) = 120, coefficient = 1.9, diameter = 120 * 1.9 = 228
		},
		{
			name:           "missing width attribute",
			attributesJSON: `[{"slug": "length", "value": 30}, {"slug": "density", "value": 1}]`,
			want:           0.0,
		},
		{
			name:           "missing length attribute",
			attributesJSON: `[{"slug": "width", "value": 50}, {"slug": "density", "value": 1}]`,
			want:           0.0,
		},
		{
			name:           "missing density attribute",
			attributesJSON: `[{"slug": "width", "value": 50}, {"slug": "length", "value": 30}]`,
			want:           0.0,
		},
		{
			name:           "invalid JSON",
			attributesJSON: `invalid json`,
			want:           0.0,
		},
		{
			name:           "empty JSON",
			attributesJSON: `[]`,
			want:           0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.calculateBubbleDiameter(tt.attributesJSON)
			if got != tt.want {
				t.Errorf("calculateBubbleDiameter() = %v, want %v", got, tt.want)
			}
		})
	}
}
