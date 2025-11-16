package entity

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID
	Username string
	TeamName string
	IsActive bool
}
