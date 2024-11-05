package users

import (
	"context"
	"errors"
	"time"

	"github.com/awhdesmond/user-service/pkg/common"
)

const (
	MaxYears = 150
)

var (
	ErrUsernameIsEmpty            = errors.New("username cannot be empty")
	ErrUsernameContainsNonLetters = errors.New("username contains non letters")
	ErrDoBFutureUsed              = errors.New("a date of birth in the future is used")
	ErrDoBTooOld                  = errors.New("date of birth is too old")
	ErrDoBInvalid                 = errors.New("invalid date of birth")
)

type Service interface {
	Upsert(ctx context.Context, username, dob string) error
	Read(ctx context.Context, username string) (string, error)
}

type service struct {
	store Store
	nowFn func() time.Time
}

func NewService(store Store, nowFn func() time.Time) Service {
	return &service{store: store, nowFn: nowFn}
}

func NewDefaultService(store Store) Service {
	return &service{store: store, nowFn: time.Now}
}

func (svc *service) validateUsername(username string) error {
	if username == "" {
		return ErrUsernameIsEmpty
	}

	if !common.StringContainsOnlyLetters(username) {
		return ErrUsernameContainsNonLetters
	}
	return nil
}

// Upsert saves/updates the given user’s name and date of birth in the database
func (svc *service) Upsert(ctx context.Context, username, dob string) error {
	if err := svc.validateUsername(username); err != nil {
		return err
	}

	dobDt, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return ErrDoBInvalid
	}

	today := time.Now()
	if dobDt.After(today) {
		return ErrDoBFutureUsed
	}
	if today.Year()-dobDt.Year() > MaxYears {
		return ErrDoBTooOld
	}

	return svc.store.Upsert(ctx, username, dobDt)
}

// Read retrieves a user and generates a Hello Birthday message
// based on the user's birthday
func (svc *service) Read(ctx context.Context, username string) (string, error) {
	if err := svc.validateUsername(username); err != nil {
		return "", err
	}

	user, err := svc.store.Read(ctx, username)
	if err != nil {
		return "", err
	}

	return user.GenerateDobMessage(svc.nowFn), nil
}
