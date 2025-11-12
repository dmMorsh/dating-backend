package models

type SimpleFilter struct {
	PageSize int64 `json:"page_size,omitempty" schema:"page_size"`

	Gender        *string  `json:"gender,omitempty" schema:"gender"`
	MinAge        *int64   `json:"min_age,omitempty" schema:"min_age"`
	MaxAge        *int64   `json:"max_age,omitempty" schema:"max_age"`
	MaxDistanceKm *float64 `json:"max_distance_km,omitempty" schema:"max_distance_km"`
	Latitude      *float64 `json:"latitude,omitempty" schema:"latitude"`
	Longitude     *float64 `json:"longitude,omitempty" schema:"longitude"`
	HasPhoto      *bool    `json:"has_photo,omitempty" schema:"has_photo"`
	InterestedIn  *string  `json:"interested_in,omitempty" schema:"interested_in"`
	LastSeenID    *int64   `json:"last_seen_id,omitempty" schema:"last_seen_id"`
	OnlineOnly    *bool    `json:"onlineOnly,omitempty" schema:"online_only"`
}