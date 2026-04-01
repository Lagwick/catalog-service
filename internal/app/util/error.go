package util

import (
	"errors"
)

func ReplaceErr1(what, from, to error) error {
	switch {
	case errors.Is(what, from):
		return to
	default:
		return what
	}
}
