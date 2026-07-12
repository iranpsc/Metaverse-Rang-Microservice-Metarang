package service

import (
	"context"

	"metarang/notifications-service/internal/errs"
	"metarang/notifications-service/internal/models"
)

type noopEmailChannel struct{}

// NewEmailChannel returns a placeholder email channel implementation.
func NewEmailChannel() EmailChannel {
	return &noopEmailChannel{}
}

func (c *noopEmailChannel) SendEmail(ctx context.Context, payload models.EmailPayload) (string, error) {
	return "", errs.ErrNotImplemented
}
