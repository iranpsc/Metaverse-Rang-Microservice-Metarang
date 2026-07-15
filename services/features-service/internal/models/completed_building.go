package models

const (
	CompletedBuildingPerPage = 10
	CompletedBuildingPath    = "/api/features/build/completed"
)

// CompletedBuildingRow is a DB row for a construction-completed building.
type CompletedBuildingRow struct {
	ID                  uint64
	FeatureID           uint64
	FeaturePropertiesID string
	AttributesJSON      string
}

// CompletedBuilding is the API resource for a completed building.
type CompletedBuilding struct {
	ID                  uint64
	FeatureID           uint64
	FeaturePropertiesID string
	Name                *string
	BuildingTotalArea   *string
	Density             *string
}

// CompletedBuildingPage is a paginated list of completed buildings.
type CompletedBuildingPage struct {
	Items       []CompletedBuilding
	CurrentPage int
	PerPage     int
	Total       int
	LastPage    int
	From        *int
	To          *int
	Path        string
}
