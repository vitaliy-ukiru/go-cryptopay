package go_crypto_pay

import (
	"errors"
	"testing"
)

func TestGetApiError(t *testing.T) {
	if GetApiError(errors.New("some error")) != nil {
		t.Error("get from errors.New")
	}
	if GetApiError(error(&ApiError{})) == nil {
		t.Error("don't get from ApiError{}")
	}
	if (ApiError{Code: 500}).Error() != GetApiError(error(&ApiError{Code: 500})).Error() {

	}
}



