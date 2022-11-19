package migrations

type Migration interface {
	Name() string
	Order() int
	Up() bool
	Down() bool
}

type MigrationsRepository interface {
	CreateTable() error
	GetAppliedMigrations() ([]MigrationEntity, error)
	SaveMigrationStatus(migration MigrationEntity) error
}
