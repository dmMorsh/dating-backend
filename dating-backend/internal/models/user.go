package models

type User struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	Password     string   `json:"password,omitempty"` // хэш, не возвращаем наружу
	Name         string   `json:"name"`
	Gender       string   `json:"gender"` // "male", "female", "other"
	Age          int      `json:"age"`
	Birthday     *SQLiteDate  `json:"birthday"`
	InterestedIn string   `json:"interested_in"` // кого ищет: "male", "female", "both"
	Bio          string   `json:"bio"`
	PhotoURL     string   `json:"photo_url"`
	Location     *string  `json:"location"` // город/район (для вывода)
	Latitude     *float64 `json:"latitude"` // для геолокации
	Longitude    *float64 `json:"longitude"`
	CreatedAt    string   `json:"created_at"`
	LastActive   string   `json:"last_active"`

	// Height      *int                   `json:"height,omitempty"`
	// Weight      *int                   `json:"weight,omitempty"`
	// Education   *string                `json:"education,omitempty"`
	// Children    *string                `json:"children,omitempty"`
	// Extras      map[string]interface{} `json:"extras,omitempty"` // хобби, питомцы, предпочтения
}