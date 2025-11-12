package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SQLiteDate struct {
	time.Time
}

// --- JSON (for client .NET) ---
func (d SQLiteDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	// RFC3339 format for .NET DateTime
	return json.Marshal(d.Time.Format(time.RFC3339))
}

func (d *SQLiteDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	// try multiple layouts
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("invalid date: %s", s)
}

// --- SQL (for SQLite) ---
func (d SQLiteDate) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format("2006-01-02 15:04:05"), nil
}

func (d *SQLiteDate) Scan(value interface{}) error {
    if value == nil {
        d.Time = time.Time{}
        return nil
    }

    switch v := value.(type) {
    case string:
        t, err := time.Parse("2006-01-02 15:04:05", v)
        if err != nil {
            t2, err2 := time.Parse("2006-01-02", v)
            if err2 != nil {
                return err
            }
            d.Time = t2
            return nil
        }
        d.Time = t
        return nil

    case []byte:
        s := string(v)
        t, err := time.Parse("2006-01-02 15:04:05", s)
        if err != nil {
            t2, err2 := time.Parse("2006-01-02", s)
            if err2 != nil {
                return err
            }
            d.Time = t2
            return nil
        }
        d.Time = t
        return nil

    case time.Time:
        d.Time = v
        return nil

    default:
        return fmt.Errorf("unsupported type %T for SQLiteDate", value)
    }
}