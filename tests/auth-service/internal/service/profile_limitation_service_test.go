package service_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"metarang/auth-service/internal/models"
	"metarang/auth-service/internal/repository"
	"metarang/auth-service/internal/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProfileLimitationTestService(t *testing.T) (service.ProfileLimitationService, *sql.DB) {
	dsn := "root@tcp(localhost:3306)/metarang_db_test?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("Database ping failed: %v", err)
	}

	_, _ = db.Exec(`
		ALTER TABLE profile_limitations
		ADD UNIQUE KEY profile_limitations_limiter_limited_unique (limiter_user_id, limited_user_id)
	`)

	_, _ = db.Exec("DELETE FROM profile_limitations")
	_, _ = db.Exec("DELETE FROM users WHERE id IN (1, 2, 3)")

	_, _ = db.Exec("INSERT INTO users (id, name, email, phone, password, code, created_at, updated_at) VALUES (1, 'User 1', 'user1@test.com', '09123456789', 'password', 'USER1', NOW(), NOW())")
	_, _ = db.Exec("INSERT INTO users (id, name, email, phone, password, code, created_at, updated_at) VALUES (2, 'User 2', 'user2@test.com', '09123456790', 'password', 'USER2', NOW(), NOW())")
	_, _ = db.Exec("INSERT INTO users (id, name, email, phone, password, code, created_at, updated_at) VALUES (3, 'User 3', 'user3@test.com', '09123456791', 'password', 'USER3', NOW(), NOW())")

	limitationRepo := repository.NewProfileLimitationRepository(db)
	userRepo := repository.NewUserRepository(db, "")
	svc := service.NewProfileLimitationService(limitationRepo, userRepo)

	return svc, db
}

func noteString(s string) service.NoteUpdate {
	return service.NoteUpdate{Present: true, Value: &s}
}

func noteClear() service.NoteUpdate {
	return service.NoteUpdate{Present: true, Value: nil}
}

func noteOmit() service.NoteUpdate {
	return service.NoteUpdate{Present: false}
}

func TestProfileLimitationService_Create(t *testing.T) {
	svc, db := setupProfileLimitationTestService(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		options := models.DefaultOptions()
		options.Follow = false
		options.SendMessage = false

		limitation, err := svc.Create(ctx, 1, 2, options, noteString("Test note"))
		require.NoError(t, err)
		assert.NotZero(t, limitation.ID)
		assert.Equal(t, uint64(1), limitation.LimiterUserID)
		assert.Equal(t, uint64(2), limitation.LimitedUserID)
		assert.False(t, limitation.Options.Follow)
		assert.True(t, limitation.Note.Valid)
		assert.Equal(t, "Test note", limitation.Note.String)
	})

	t.Run("duplicate creation fails", func(t *testing.T) {
		options := models.DefaultOptions()
		_, err := svc.Create(ctx, 1, 2, options, noteOmit())
		assert.ErrorIs(t, err, service.ErrProfileLimitationAlreadyExists)
	})

	t.Run("invalid limited user fails with 404 contract", func(t *testing.T) {
		options := models.DefaultOptions()
		_, err := svc.Create(ctx, 1, 99999, options, noteOmit())
		assert.ErrorIs(t, err, service.ErrUserNotFound)
	})

	t.Run("note too long fails", func(t *testing.T) {
		options := models.DefaultOptions()
		long := make([]byte, 501)
		for i := range long {
			long[i] = 'a'
		}
		_, err := svc.Create(ctx, 1, 3, options, noteString(string(long)))
		assert.ErrorIs(t, err, service.ErrNoteTooLong)
	})
}

func TestProfileLimitationService_Update(t *testing.T) {
	svc, db := setupProfileLimitationTestService(t)
	defer db.Close()

	ctx := context.Background()

	options := models.DefaultOptions()
	limitation, err := svc.Create(ctx, 1, 2, options, noteString("Original note"))
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		newOptions := models.DefaultOptions()
		newOptions.Follow = false

		updated, err := svc.Update(ctx, limitation.ID, 1, newOptions, noteString("Updated note"))
		require.NoError(t, err)
		assert.False(t, updated.Options.Follow)
		assert.Equal(t, "Updated note", updated.Note.String)
	})

	t.Run("omitted note retains existing", func(t *testing.T) {
		newOptions := models.DefaultOptions()
		updated, err := svc.Update(ctx, limitation.ID, 1, newOptions, noteOmit())
		require.NoError(t, err)
		assert.True(t, updated.Note.Valid)
		assert.Equal(t, "Updated note", updated.Note.String)
	})

	t.Run("explicit clear note", func(t *testing.T) {
		newOptions := models.DefaultOptions()
		updated, err := svc.Update(ctx, limitation.ID, 1, newOptions, noteClear())
		require.NoError(t, err)
		assert.False(t, updated.Note.Valid)
	})

	t.Run("unauthorized update fails", func(t *testing.T) {
		newOptions := models.DefaultOptions()
		_, err := svc.Update(ctx, limitation.ID, 2, newOptions, noteOmit())
		assert.ErrorIs(t, err, service.ErrUnauthorized)
	})

	t.Run("not found fails", func(t *testing.T) {
		newOptions := models.DefaultOptions()
		_, err := svc.Update(ctx, 99999, 1, newOptions, noteOmit())
		assert.ErrorIs(t, err, service.ErrProfileLimitationNotFound)
	})
}

func TestProfileLimitationService_Delete(t *testing.T) {
	svc, db := setupProfileLimitationTestService(t)
	defer db.Close()

	ctx := context.Background()
	options := models.DefaultOptions()
	limitation, err := svc.Create(ctx, 1, 2, options, noteOmit())
	require.NoError(t, err)

	t.Run("successful delete", func(t *testing.T) {
		err := svc.Delete(ctx, limitation.ID, 1)
		assert.NoError(t, err)
		_, err = svc.GetByID(ctx, limitation.ID)
		assert.ErrorIs(t, err, service.ErrProfileLimitationNotFound)
	})

	t.Run("unauthorized delete fails", func(t *testing.T) {
		limitation2, err := svc.Create(ctx, 1, 3, options, noteOmit())
		require.NoError(t, err)
		err = svc.Delete(ctx, limitation2.ID, 2)
		assert.ErrorIs(t, err, service.ErrUnauthorized)
	})
}

func TestProfileLimitationService_GetBetweenUsers(t *testing.T) {
	svc, db := setupProfileLimitationTestService(t)
	defer db.Close()

	ctx := context.Background()
	options := models.DefaultOptions()
	limitation, err := svc.Create(ctx, 1, 2, options, noteString("Test note"))
	require.NoError(t, err)

	t.Run("find from limiter perspective", func(t *testing.T) {
		found, err := svc.GetBetweenUsers(ctx, 1, 2)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, limitation.ID, found.ID)
	})

	t.Run("find from limited perspective", func(t *testing.T) {
		found, err := svc.GetBetweenUsers(ctx, 2, 1)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, limitation.ID, found.ID)
	})

	t.Run("empty when users exist but no limitation", func(t *testing.T) {
		found, err := svc.GetBetweenUsers(ctx, 2, 3)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("missing target user returns not found", func(t *testing.T) {
		_, err := svc.GetBetweenUsers(ctx, 1, 99999)
		assert.ErrorIs(t, err, service.ErrUserNotFound)
	})
}

func TestProfileLimitationService_ConcurrentDuplicateCreate(t *testing.T) {
	svc, db := setupProfileLimitationTestService(t)
	defer db.Close()

	ctx := context.Background()
	options := models.DefaultOptions()

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Create(ctx, 1, 2, options, noteString("race"))
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	var success, alreadyExists, other int
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, service.ErrProfileLimitationAlreadyExists):
			alreadyExists++
		default:
			other++
			t.Logf("unexpected error: %v", err)
		}
	}

	assert.Equal(t, 1, success)
	assert.Equal(t, 1, alreadyExists)
	assert.Equal(t, 0, other)
}
