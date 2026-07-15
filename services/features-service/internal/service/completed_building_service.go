package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"metarang/features-service/internal/models"
)

type completedBuildingRepo interface {
	FindCompleted(ctx context.Context, now time.Time, limit, offset int) ([]models.CompletedBuildingRow, error)
	CountCompleted(ctx context.Context, now time.Time) (int, error)
}

// CompletedBuildingService lists construction-completed buildings (Laravel BuildFeatureController@completedBuildings).
type CompletedBuildingService struct {
	repo completedBuildingRepo
	now  func() time.Time
}

func NewCompletedBuildingService(repo completedBuildingRepo) *CompletedBuildingService {
	return &CompletedBuildingService{
		repo: repo,
		now:  time.Now,
	}
}

// Paginate returns a page of completed buildings (10 per page).
func (s *CompletedBuildingService) Paginate(ctx context.Context, page int) (*models.CompletedBuildingPage, error) {
	if page < 1 {
		page = 1
	}

	now := s.now()
	perPage := models.CompletedBuildingPerPage

	total, err := s.repo.CountCompleted(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("count completed buildings: %w", err)
	}

	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	offset := (page - 1) * perPage
	rows, err := s.repo.FindCompleted(ctx, now, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("list completed buildings: %w", err)
	}

	items := make([]models.CompletedBuilding, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapCompletedBuilding(row))
	}

	result := &models.CompletedBuildingPage{
		Items:       items,
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
		Path:        models.CompletedBuildingPath,
	}
	if len(items) > 0 {
		from := offset + 1
		to := offset + len(items)
		result.From = &from
		result.To = &to
	}

	return result, nil
}

func mapCompletedBuilding(row models.CompletedBuildingRow) models.CompletedBuilding {
	name, area, density := extractCompletedBuildingAttributes(row.AttributesJSON)
	return models.CompletedBuilding{
		ID:                  row.ID,
		FeatureID:           row.FeatureID,
		FeaturePropertiesID: strings.ToUpper(row.FeaturePropertiesID),
		Name:                name,
		BuildingTotalArea:   area,
		Density:             density,
	}
}

func extractCompletedBuildingAttributes(attributesJSON string) (name, area, density *string) {
	if attributesJSON == "" {
		return nil, nil, nil
	}

	var attrs []map[string]interface{}
	if err := json.Unmarshal([]byte(attributesJSON), &attrs); err != nil {
		return nil, nil, nil
	}

	return attributeStringPtr(attrs, "name"),
		attributeStringPtr(attrs, "area"),
		attributeStringPtr(attrs, "density")
}

func attributeStringPtr(attrs []map[string]interface{}, slug string) *string {
	for _, attr := range attrs {
		s, ok := attr["slug"].(string)
		if !ok || s != slug {
			continue
		}
		if attr["value"] == nil {
			return nil
		}
		switch v := attr["value"].(type) {
		case string:
			return &v
		case float64:
			formatted := strconv.FormatFloat(v, 'f', -1, 64)
			return &formatted
		case bool:
			formatted := strconv.FormatBool(v)
			return &formatted
		default:
			formatted := fmt.Sprint(v)
			return &formatted
		}
	}
	return nil
}
