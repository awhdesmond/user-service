package common

import (
	"errors"
	"unicode"
)

type Errorer interface {
	Error() error
}

type GenericJSON map[string]interface{}

func StringContainsOnlyLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func ErrorContains(slice []error, target error) bool {
	for _, err := range slice {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
