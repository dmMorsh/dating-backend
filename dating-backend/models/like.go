package models

type Like struct {
	ID       int64 `json:"id"`
	FromUser int64 `json:"from_user"`
	ToUser   int64 `json:"to_user"`
	Is_match bool  `json:"is_match"`
}
