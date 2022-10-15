package cryptopay

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// ApiError is error of CryptoPay API.
type ApiError struct {
	Code int    `json:"code"` // HTTP Code of error.
	Name string `json:"name"` // Name of error.
}

// AsApiError retrieves the ApiError from given error.
// If unsuccessfully returns nil.
func AsApiError(err error) *ApiError {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

func (a ApiError) Error() string {
	return fmt.Sprintf("crypto-pay/api: api response %d - %s", a.Code, a.Name)
}
