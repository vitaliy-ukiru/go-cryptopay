package internal

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrNotMatchesParts = errors.New("invalid token parts")
	ErrInvalidAppID    = errors.New("invalid appID in token")
)

func ValidateToken(token string) (int, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return 0, ErrNotMatchesParts
	}
	appId, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, ErrInvalidAppID
	}
	return appId, nil
}
