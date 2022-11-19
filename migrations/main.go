package migrations

import (
	"context"
	"strings"
	"time"

	"github.com/cjlapao/common-go/guard"
	"github.com/cjlapao/common-go/log"
	"github.com/elliotchance/orderedmap/v2"
)

const (
	MIGRATION_TABLE_NAME = "_migrations"
)

type SqlMigrationService struct {
	Context           context.Context
	logger            *log.Logger
	Migrations        *orderedmap.OrderedMap[int, Migration]
	AppliedMigrations []MigrationEntity
	Repository        MigrationsRepository
}

func NewMigrationService(repo MigrationsRepository) *SqlMigrationService {
	logger := log.Get()
	if err := guard.EmptyOrNil(repo, "repository"); err != nil {
		logger.Exception(err, "error getting the repository")
		return nil
	}

	m := SqlMigrationService{
		Migrations:        orderedmap.NewOrderedMap[int, Migration](),
		AppliedMigrations: make([]MigrationEntity, 0),
		logger:            log.Get(),
		Repository:        repo,
	}

	if err := repo.CreateTable(); err != nil {
		return nil
	}

	return &m
}

func (m *SqlMigrationService) WasApplied(name string) bool {
	for _, migration := range m.AppliedMigrations {
		if strings.EqualFold(name, migration.Name) {
			m.logger.Info("Migration %v was already applied on %v", migration.Name, migration.ExecutedOn.Format(time.RFC3339))
			return true
		}
	}

	return false
}

func (m *SqlMigrationService) Register(migration Migration) {
	m.Migrations.Set(migration.Order(), migration)
}

func (m *SqlMigrationService) Run() error {
	logger := log.Get()

	if appliedMigrations, err := m.Repository.GetAppliedMigrations(); err != nil {
		return err
	} else {
		m.AppliedMigrations = append(m.AppliedMigrations, appliedMigrations...)
	}

	for el := m.Migrations.Front(); el != nil; el = el.Next() {
		migration := MigrationEntity{
			Name:   strings.ReplaceAll(el.Value.Name(), " ", "_"),
			Status: false,
		}
		if m.WasApplied(el.Value.Name()) {
			continue
		}

		if !el.Value.Up() {
			logger.Error("there was an error applying migration %v, running migration down", el.Value.Name())
			if !el.Value.Down() {
				logger.Error("there was an error applying migration down for %v, database might be inconsistent", el.Value.Name())
			}
			// Stopping migrations as they need to be run in order
			break
		} else {
			logger.Info("Migration %v was applied successfully", el.Value.Name())
			migration.Status = true
			migration.ExecutedOn = time.Now()
		}

		m.Repository.SaveMigrationStatus(migration)
	}

	return nil
}
