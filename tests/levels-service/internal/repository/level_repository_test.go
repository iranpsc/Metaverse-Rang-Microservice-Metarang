package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptime "github.com/yaa110/go-persian-calendar"
)

func TestLevelRepository_FormatImageURL(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	t.Run("WithAdminPanelURL", func(t *testing.T) {
		repo := NewLevelRepository(db, "https://admin.example.com")

		// Test relative path
		url := repo.formatImageURL("image.jpg")
		assert.Equal(t, "https://admin.example.com/uploads/image.jpg", url)

		// Test path with uploads prefix
		url = repo.formatImageURL("uploads/image.jpg")
		assert.Equal(t, "https://admin.example.com/uploads/image.jpg", url)

		// Test full URL (should return as-is)
		url = repo.formatImageURL("https://example.com/image.jpg")
		assert.Equal(t, "https://example.com/image.jpg", url)
	})

	t.Run("WithoutAdminPanelURL", func(t *testing.T) {
		repo := NewLevelRepository(db, "")

		// Test relative path
		url := repo.formatImageURL("image.jpg")
		assert.Equal(t, "/uploads/image.jpg", url)

		// Test path with uploads prefix
		url = repo.formatImageURL("uploads/image.jpg")
		assert.Equal(t, "/uploads/image.jpg", url)
	})

	t.Run("EmptyURL", func(t *testing.T) {
		repo := NewLevelRepository(db, "https://admin.example.com")
		url := repo.formatImageURL("")
		assert.Equal(t, "", url)
	})
}

func TestLevelRepository_FormatJalaliDateTime(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "")

	t.Run("ValidDate", func(t *testing.T) {
		// Test with a known date: 2024-01-15 14:30:45
		testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
		formatted := repo.formatJalaliDateTime(testTime)

		// Verify format is Y/m/d H:i:s (Jalali)
		pt := ptime.New(testTime)
		expected := pt.Format("yyyy/MM/dd HH:mm:ss")
		assert.Equal(t, expected, formatted)
		assert.Contains(t, formatted, "/")
		assert.Contains(t, formatted, ":")
	})
}

func TestLevelRepository_GetLevelPrize_JalaliFormat(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "https://admin.example.com")
	ctx := context.Background()

	t.Run("PrizeWithCreatedAt", func(t *testing.T) {
		createdAt := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
		prizeRows := sqlmock.NewRows([]string{"id", "level_id", "psc", "blue", "red", "yellow", "effect", "satisfaction", "created_at"}).
			AddRow(1, 1, 1000, 5, 3, 2, 10, 50.75, createdAt)

		mock.ExpectQuery("SELECT id, level_id, psc, blue, red, yellow, effect, satisfaction, created_at").
			WithArgs(uint64(1)).
			WillReturnRows(prizeRows)

		prize, err := repo.GetLevelPrize(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, prize)

		// Verify satisfaction is formatted to 2 decimal places
		assert.Equal(t, "50.75", prize.Satisfaction)

		// Verify created_at is in Jalali format Y/m/d H:i:s
		pt := ptime.New(createdAt)
		expectedJalali := pt.Format("yyyy/MM/dd HH:mm:ss")
		assert.Equal(t, expectedJalali, prize.CreatedAt)
		assert.Contains(t, prize.CreatedAt, "/")
		assert.Contains(t, prize.CreatedAt, ":")
	})
}

func TestLevelRepository_GetLevelGeneralInfo_FileURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "https://admin.example.com")
	ctx := context.Background()

	t.Run("GeneralInfoWithFileURLs", func(t *testing.T) {
		generalInfoRows := sqlmock.NewRows([]string{"id", "level_id", "score", "rank", "description", "subcategories",
			"persian_font", "english_font", "file_volume", "used_colors", "points", "lines",
			"has_animation", "designer", "model_designer", "creation_date", "png_file", "fbx_file", "gif_file"}).
			AddRow(1, 1, 100, 1, "Description", 2, "Font1", "Font2", 1.5, "Colors", 100, 200, 1,
				"Designer", "Model Designer", "2024-01-01", "png.png", "fbx.fbx", "gif.gif")

		mock.ExpectQuery("SELECT id, level_id, score, rank").
			WithArgs(uint64(1)).
			WillReturnRows(generalInfoRows)

		info, err := repo.GetLevelGeneralInfo(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, info)

		// Verify file URLs are formatted with admin_panel_url
		assert.Equal(t, "https://admin.example.com/uploads/png.png", info.PngFile)
		assert.Equal(t, "https://admin.example.com/uploads/fbx.fbx", info.FbxFile)
		assert.Equal(t, "https://admin.example.com/uploads/gif.gif", info.GifFile)
	})
}

func TestLevelRepository_GetAllLevels_ImageURLFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "https://admin.example.com")
	ctx := context.Background()

	t.Run("LevelsWithImageURLs", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "score", "background_image", "image_url"}).
			AddRow(1, "Level 1", "level-1", 100, "bg1.jpg", "img1.jpg").
			AddRow(2, "Level 2", "level-2", 200, "", "")

		mock.ExpectQuery("SELECT l.id, l.name, l.slug, CAST\\(l.score AS UNSIGNED\\) as score").
			WillReturnRows(rows)

		levels, err := repo.GetAllLevels(ctx)
		require.NoError(t, err)
		require.Len(t, levels, 2)

		// Verify image URL is formatted with admin_panel_url
		assert.Equal(t, "https://admin.example.com/uploads/img1.jpg", levels[0].ImageUrl)

		// Verify empty image URL is not formatted
		assert.Equal(t, "", levels[1].ImageUrl)
	})
}

func TestLevelRepository_FindBySlug_ImageURLFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "https://admin.example.com")
	ctx := context.Background()

	t.Run("LevelWithImageURL", func(t *testing.T) {
		levelRows := sqlmock.NewRows([]string{"id", "name", "slug", "score", "background_image", "image_url"}).
			AddRow(1, "Level 1", "level-1", 100, "bg1.jpg", "img1.jpg")

		generalInfoRows := sqlmock.NewRows([]string{"id", "level_id", "score", "rank", "description", "subcategories",
			"persian_font", "english_font", "file_volume", "used_colors", "points", "lines",
			"has_animation", "designer", "model_designer", "creation_date", "png_file", "fbx_file", "gif_file"}).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT l.id, l.name, l.slug, CAST\\(l.score AS UNSIGNED\\) as score").
			WithArgs("level-1").
			WillReturnRows(levelRows)

		mock.ExpectQuery("SELECT id, level_id, score, rank").
			WithArgs(uint64(1)).
			WillReturnRows(generalInfoRows)

		level, err := repo.FindBySlug(ctx, "level-1")
		require.NoError(t, err)
		require.NotNil(t, level)

		// Verify image URL is formatted with admin_panel_url
		assert.Equal(t, "https://admin.example.com/uploads/img1.jpg", level.ImageUrl)
	})
}

func TestLevelRepository_GetLevelPrize_SatisfactionFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewLevelRepository(db, "")
	ctx := context.Background()

	testCases := []struct {
		name         string
		satisfaction float64
		expected     string
	}{
		{
			name:         "TwoDecimalPlaces",
			satisfaction: 50.75,
			expected:     "50.75",
		},
		{
			name:         "OneDecimalPlace",
			satisfaction: 50.5,
			expected:     "50.50",
		},
		{
			name:         "IntegerValue",
			satisfaction: 50.0,
			expected:     "50.00",
		},
		{
			name:         "ManyDecimalPlaces",
			satisfaction: 50.123456,
			expected:     "50.12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prizeRows := sqlmock.NewRows([]string{"id", "level_id", "psc", "blue", "red", "yellow", "effect", "satisfaction", "created_at"}).
				AddRow(1, 1, 1000, 5, 3, 2, 10, tc.satisfaction, nil)

			mock.ExpectQuery("SELECT id, level_id, psc, blue, red, yellow, effect, satisfaction, created_at").
				WithArgs(uint64(1)).
				WillReturnRows(prizeRows)

			prize, err := repo.GetLevelPrize(ctx, 1)
			require.NoError(t, err)
			require.NotNil(t, prize)

			assert.Equal(t, tc.expected, prize.Satisfaction)
		})
	}
}
