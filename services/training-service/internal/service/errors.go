package service

import "errors"

var (
	ErrNotAuthorized = errors.New("user not authorized")
	ErrNotFound      = errors.New("resource not found")
)
