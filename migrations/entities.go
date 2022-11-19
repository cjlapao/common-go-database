package migrations

import "time"

type MigrationEntity struct {
	ID         string    `json:"id"`
	ExecutedOn time.Time `json:"executed_on"`
	Name       string    `json:"name"`
	Status     bool      `json:"status"`
}
