package db_migrations

type Migration interface {
	Name() string
	Order() int
	Up() bool
	Down() bool
}
