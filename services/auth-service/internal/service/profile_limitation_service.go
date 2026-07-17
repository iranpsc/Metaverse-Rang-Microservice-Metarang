package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"metarang/auth-service/internal/models"
	"metarang/auth-service/internal/repository"
)

var (
	ErrProfileLimitationNotFound      = errors.New("profile limitation not found")
	ErrProfileLimitationAlreadyExists = errors.New("profile limitation already exists for this user pair")
	ErrInvalidOptions                 = errors.New("invalid options: all six boolean keys are required")
	ErrNoteTooLong                    = errors.New("note must be 500 characters or less")
	ErrUnauthorized                   = errors.New("unauthorized: you can only modify limitations you created")
)

// NoteUpdate carries note presence for update/create semantics.
// Present=false: leave note unchanged (update) or unset (create).
// Present=true with Value=nil: clear the note.
// Present=true with Value set: store that string.
type NoteUpdate struct {
	Present bool
	Value   *string
}

type ProfileLimitationService interface {
	Create(ctx context.Context, limiterUserID, limitedUserID uint64, options models.ProfileLimitationOptions, note NoteUpdate) (*models.ProfileLimitation, error)
	Update(ctx context.Context, limitationID, limiterUserID uint64, options models.ProfileLimitationOptions, note NoteUpdate) (*models.ProfileLimitation, error)
	Delete(ctx context.Context, limitationID, limiterUserID uint64) error
	GetByID(ctx context.Context, limitationID uint64) (*models.ProfileLimitation, error)
	GetBetweenUsers(ctx context.Context, callerUserID, targetUserID uint64) (*models.ProfileLimitation, error)
}

type profileLimitationService struct {
	limitationRepo repository.ProfileLimitationRepository
	userRepo       repository.UserRepository
}

func NewProfileLimitationService(limitationRepo repository.ProfileLimitationRepository, userRepo repository.UserRepository) ProfileLimitationService {
	return &profileLimitationService{
		limitationRepo: limitationRepo,
		userRepo:       userRepo,
	}
}

func (s *profileLimitationService) Create(ctx context.Context, limiterUserID, limitedUserID uint64, options models.ProfileLimitationOptions, note NoteUpdate) (*models.ProfileLimitation, error) {
	limitedUser, err := s.userRepo.FindByID(ctx, limitedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check limited user: %w", err)
	}
	if limitedUser == nil {
		return nil, ErrUserNotFound
	}

	limiterUser, err := s.userRepo.FindByID(ctx, limiterUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check limiter user: %w", err)
	}
	if limiterUser == nil {
		return nil, ErrUserNotFound
	}

	// Early readable check; unique constraint is the race-safe guarantee.
	exists, err := s.limitationRepo.ExistsForLimiterAndLimited(ctx, limiterUserID, limitedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing limitation: %w", err)
	}
	if exists {
		return nil, ErrProfileLimitationAlreadyExists
	}

	if err := validateNoteUpdate(note); err != nil {
		return nil, err
	}

	limitation := &models.ProfileLimitation{
		LimiterUserID: limiterUserID,
		LimitedUserID: limitedUserID,
		Options:       options,
		Note:          noteToNullString(note),
	}

	if err := s.limitationRepo.Create(ctx, limitation); err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrProfileLimitationAlreadyExists
		}
		return nil, fmt.Errorf("failed to create profile limitation: %w", err)
	}

	return limitation, nil
}

func (s *profileLimitationService) Update(ctx context.Context, limitationID, limiterUserID uint64, options models.ProfileLimitationOptions, note NoteUpdate) (*models.ProfileLimitation, error) {
	limitation, err := s.limitationRepo.FindByID(ctx, limitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get limitation: %w", err)
	}
	if limitation == nil {
		return nil, ErrProfileLimitationNotFound
	}

	if limitation.LimiterUserID != limiterUserID {
		return nil, ErrUnauthorized
	}

	if err := validateNoteUpdate(note); err != nil {
		return nil, err
	}

	limitation.Options = options
	if note.Present {
		limitation.Note = noteToNullString(note)
	}

	if err := s.limitationRepo.Update(ctx, limitation); err != nil {
		return nil, fmt.Errorf("failed to update profile limitation: %w", err)
	}

	return limitation, nil
}

func (s *profileLimitationService) Delete(ctx context.Context, limitationID, limiterUserID uint64) error {
	limitation, err := s.limitationRepo.FindByID(ctx, limitationID)
	if err != nil {
		return fmt.Errorf("failed to get limitation: %w", err)
	}
	if limitation == nil {
		return ErrProfileLimitationNotFound
	}

	if limitation.LimiterUserID != limiterUserID {
		return ErrUnauthorized
	}

	if err := s.limitationRepo.Delete(ctx, limitationID); err != nil {
		return fmt.Errorf("failed to delete profile limitation: %w", err)
	}

	return nil
}

func (s *profileLimitationService) GetByID(ctx context.Context, limitationID uint64) (*models.ProfileLimitation, error) {
	limitation, err := s.limitationRepo.FindByID(ctx, limitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get limitation: %w", err)
	}
	if limitation == nil {
		return nil, ErrProfileLimitationNotFound
	}
	return limitation, nil
}

func (s *profileLimitationService) GetBetweenUsers(ctx context.Context, callerUserID, targetUserID uint64) (*models.ProfileLimitation, error) {
	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check target user: %w", err)
	}
	if targetUser == nil {
		return nil, ErrUserNotFound
	}

	limitation, err := s.limitationRepo.FindBetweenUsers(ctx, callerUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get limitation between users: %w", err)
	}
	return limitation, nil
}

func validateNoteUpdate(note NoteUpdate) error {
	if !note.Present || note.Value == nil {
		return nil
	}
	if len(*note.Value) > 500 {
		return ErrNoteTooLong
	}
	return nil
}

func noteToNullString(note NoteUpdate) sql.NullString {
	if !note.Present || note.Value == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *note.Value, Valid: true}
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
