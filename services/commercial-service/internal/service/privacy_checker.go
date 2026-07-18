package service

// walletPrivacyMap maps asset name → privacy settings key.
// Assets absent from this map (e.g. "effect") are always visible.
var walletPrivacyMap = map[string]string{
	"psc":          "psc_transactions",
	"irr":          "irr_transactions",
	"red":          "red_transactions",
	"blue":         "blue_transactions",
	"yellow":       "yellow_transactions",
	"satisfaction": "satisfaction",
}

// IsAssetVisible returns true when the asset should be shown publicly.
// Rule: missing privacy key defaults to visible (1); value == 0 → hidden.
// Assets without a privacy key (e.g. effect) are always visible.
func IsAssetVisible(privacy map[string]int32, asset string) bool {
	key, ok := walletPrivacyMap[asset]
	if !ok || key == "" {
		return true
	}
	if privacy == nil {
		return true
	}
	val, exists := privacy[key]
	if !exists {
		return true
	}
	return val != 0
}

// FilterAllowedAssets returns only the assets that pass the privacy check.
func FilterAllowedAssets(privacy map[string]int32, assets []string) []string {
	out := make([]string, 0, len(assets))
	for _, asset := range assets {
		if IsAssetVisible(privacy, asset) {
			out = append(out, asset)
		}
	}
	return out
}
