package sql

import (
	"context"
	"fmt"

	"github.com/cjlapao/common-go-database/migrations"
	"github.com/cjlapao/common-go/log"
	"github.com/google/uuid"
)

type SqlMigrationsRepo struct {
	context  context.Context
	logger   *log.Logger
	database *SqlFactory
}

func NewSqlMigrationRepo() *SqlMigrationsRepo {
	return &SqlMigrationsRepo{
		context:  context.Background(),
		logger:   log.Get(),
		database: Get().GlobalDatabase(),
	}
}

func (m *SqlMigrationsRepo) CreateTable() error {

	globalDb := m.database.Connect()

	if globalDb == nil {
		err := fmt.Errorf("error connecting to database %v", m.database.DatabaseContext.CurrentDatabase())
		m.logger.Error(err.Error())
		return err
	}

	defer globalDb.Close()

	_, err := globalDb.ExecContext(`
  CREATE TABLE IF NOT EXISTS ` + migrations.MIGRATION_TABLE_NAME + `(  
    id CHAR(36) NOT NULL PRIMARY KEY COMMENT 'Primary Key',
    executed_on DATETIME COMMENT 'Time of execution',
    name CHAR(150) NOT NULL COMMENT 'Migration Name',
    status BOOLEAN COMMENT 'Migration Status'
) DEFAULT CHARSET UTF8 COMMENT '';
`)

	if err != nil {
		err := fmt.Errorf("error creating migrations table on database %v", m.database.DatabaseContext.CurrentDatabase())
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

func (m *SqlMigrationsRepo) GetAppliedMigrations() ([]migrations.MigrationEntity, error) {
	var queryResult []migrations.MigrationEntity
	globalDb := m.database.Connect()

	defer globalDb.Close()

	result, err := globalDb.QueryContext(`
SELECT 
  * 
FROM 
  ` + migrations.MIGRATION_TABLE_NAME + `
WHERE 
  status = TRUE
ORDER BY 
  executed_on ASC;
`)

	if err != nil {
		return nil, err
	}

	queryResult = make([]migrations.MigrationEntity, 0)

	for result.Next() {
		var migration migrations.MigrationEntity

		err := result.Scan(&migration.ID, &migration.ExecutedOn, &migration.Name, &migration.Status)

		if err != nil {
			return nil, err
		}

		queryResult = append(queryResult, migration)
	}
	return queryResult, nil
}

func (m *SqlMigrationsRepo) SaveMigrationStatus(migration migrations.MigrationEntity) error {
	globalDb := m.database.Connect()

	migration.ID = uuid.NewString()

	defer globalDb.Close()

	_, err := globalDb.QueryContext(`
INSERT INTO
`+migrations.MIGRATION_TABLE_NAME+`(id, executed_on, name, status)
VALUES(?, ?, ?, ?)
`,
		migration.ID,
		migration.ExecutedOn,
		migration.Name,
		migration.Status,
	)

	if err != nil {
		return err
	}

	return nil
}
