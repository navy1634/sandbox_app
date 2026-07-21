package domain

import "time"

type AuthUser struct {
	AccountID int64
	Email     string
	Name      string
	Picture   string
}

type AppAccount struct {
	ID        int       `json:"id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	Picture   string    `json:"picture,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
