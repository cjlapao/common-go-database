package db_migrations

type MigrationsRepository interface {
	CreateTable() error
	GetAppliedMigrations() ([]MigrationEntity, error)
	SaveMigrationStatus(migration MigrationEntity) error
}
