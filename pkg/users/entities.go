package users

import (
	"fmt"
	"math"
	"time"
)

type User struct {
	// Username is unique
	Username string    `json:"username" db:"username"`
	DoB      time.Time `json:"dateOfBirth" db:"date_of_birth"`
}

func (u User) CalcDaysToBirthday(nowFn func() time.Time) int {
	// Use server's local time.
	// NOTE: Might have some edge cases not handled as we are not storing date's timezone
	today := nowFn()

	if u.DoB.Month() == today.Month() && u.DoB.Day() == today.Day() {
		return 0
	}

	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	birthdayThisYear := time.Date(today.Year(), u.DoB.Month(), u.DoB.Day(), 0, 0, 0, 0, time.UTC)

	// Birthday has not yet passed in the current year
	if birthdayThisYear.After(todayDate) {
		return int(math.Ceil(birthdayThisYear.Sub(today).Hours() / 24))
	}

	// Birthday has already passed in the current year,
	// we need to increase the year
	birthdayNextYear := time.Date(today.Year()+1, u.DoB.Month(), u.DoB.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Ceil(birthdayNextYear.Sub(today).Hours() / 24))

}

func (u User) GenerateDobMessage(nowFn func() time.Time) string {
	numDaysToBirthday := u.CalcDaysToBirthday(nowFn)
	if numDaysToBirthday == 0 {
		return fmt.Sprintf("Hello, %s! Happy birthday!", u.Username)
	}

	return fmt.Sprintf("Hello, %s! Your birthday is in %d day(s)", u.Username, numDaysToBirthday)
}
