package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"metargb/training-service/internal/service"
)

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrNotAuthorized) {
		return status.Errorf(codes.PermissionDenied, "%s", err.Error())
	}
	return status.Errorf(codes.InvalidArgument, "%s", err.Error())
}
