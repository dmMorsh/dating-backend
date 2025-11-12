package utils

import (
	"time"
)

// GetAge calculates age from birthday
func GetAge(birthday *time.Time) int {
	if birthday == nil {
		return 0
	}
	now := time.Now()
	years := now.Year() - birthday.Year()
	// If birthday hasn't occurred yet this year, subtract one
	if now.Month() < birthday.Month() || (now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		years--
	}
	return years
}