package service

import "metarang/financial-service/internal/sadad"

// SadadClient interface for payment gateway operations.
// Allows for easier testing with mocks.
type SadadClient interface {
	RequestPayment(params sadad.RequestParams) (*sadad.RequestResponse, error)
	VerifyPayment(params sadad.VerificationParams) (*sadad.VerificationResponse, error)
}
