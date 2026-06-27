package sadad

// SadadError handles Sadad error codes with Persian messages.
type SadadError struct {
	Code string
}

// NewSadadError creates a new Sadad error.
func NewSadadError(code string) *SadadError {
	return &SadadError{Code: code}
}

// Message returns the Persian error message for the error code.
func (e *SadadError) Message() string {
	switch e.Code {
	case "0":
		return "تراکنش موفق"
	case "-1":
		return "تراکنش ناموفق می باشد"
	case "101":
		return "پذیرنده نامعتبر است"
	case "102":
		return "ترمینال نامعتبر است"
	case "103":
		return "مبلغ نامعتبر است"
	case "104":
		return "شناسه سفارش تکراری است"
	case "105":
		return "امضای دیجیتال نامعتبر است"
	case "106":
		return "توکن نامعتبر است"
	case "107":
		return "تراکنش قبلا تایید شده است"
	default:
		return "خطای ناشناخته"
	}
}

// GetCode returns the error code.
func (e *SadadError) GetCode() string {
	return e.Code
}

// IsSuccess checks if the code indicates success.
func (e *SadadError) IsSuccess() bool {
	return e.Code == "0"
}
