package entity

import "github.com/google/uuid"

type Team struct {
	TeamName string
	Members  []uuid.UUID
}
