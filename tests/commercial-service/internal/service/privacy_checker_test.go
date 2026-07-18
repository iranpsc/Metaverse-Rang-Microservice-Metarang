package service_test

import (
	"testing"

	"metarang/commercial-service/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestIsAssetVisible_VisibleAsset(t *testing.T) {
	privacy := map[string]int32{"psc_transactions": 1}
	assert.True(t, service.IsAssetVisible(privacy, "psc"))
}

func TestIsAssetVisible_HiddenAsset(t *testing.T) {
	privacy := map[string]int32{"psc_transactions": 0}
	assert.False(t, service.IsAssetVisible(privacy, "psc"))
}

func TestIsAssetVisible_MissingKeyDefaultsVisible(t *testing.T) {
	privacy := map[string]int32{}
	assert.True(t, service.IsAssetVisible(privacy, "irr"))
}

func TestIsAssetVisible_NilPrivacyDefaultsVisible(t *testing.T) {
	assert.True(t, service.IsAssetVisible(nil, "red"))
}

func TestIsAssetVisible_EffectAlwaysVisible(t *testing.T) {
	privacy := map[string]int32{"psc_transactions": 0}
	assert.True(t, service.IsAssetVisible(privacy, "effect"))
}

func TestIsAssetVisible_UnknownAssetAlwaysVisible(t *testing.T) {
	privacy := map[string]int32{"psc_transactions": 0}
	assert.True(t, service.IsAssetVisible(privacy, "unknown_token"))
}

func TestFilterAllowedAssets(t *testing.T) {
	privacy := map[string]int32{
		"psc_transactions": 1,
		"irr_transactions": 0,
		"red_transactions": 1,
	}
	assets := []string{"psc", "irr", "red", "effect"}
	got := service.FilterAllowedAssets(privacy, assets)
	assert.Equal(t, []string{"psc", "red", "effect"}, got)
}
