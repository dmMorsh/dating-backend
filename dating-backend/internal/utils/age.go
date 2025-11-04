package utils

import (
	"strings"
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

type JSONDate time.Time

func (jt *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		return nil
	}

	// try few formates
	layouts := []string{
		time.RFC3339,              // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05",     
		"2006-01-02",              
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			*jt = JSONDate(t)
			return nil
		}
	}
	return err
}

func (jt JSONDate) Time() time.Time {
	return time.Time(jt)
}