package sqlite

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cjlapao/common-go/execution_context"
	"github.com/cjlapao/common-go/log"
)

type sqlDatabase struct {
	context        context.Context
	defaultTimeout time.Duration
	factory        *SqlFactory
	db             *sql.DB
}

func (db *sqlDatabase) Close() error {
	return db.db.Close()
}

func (db *sqlDatabase) Query(query string, args ...any) (*sql.Rows, error) {
	return db.db.Query(query, args...)
}

func (db *sqlDatabase) QueryContext(query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(db.context, query, args...)
}

func (db *sqlDatabase) QueryRow(query string, args ...any) *sql.Row {
	return db.db.QueryRow(query, args...)
}

func (db *sqlDatabase) QueryRowContext(query string, args ...any) *sql.Row {
	return db.db.QueryRowContext(db.context, query, args...)
}

func (db *sqlDatabase) ExecContext(query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(db.context, query, args...)
}

type SqlFactory struct {
	Context         *execution_context.Context
	Database        *sqlDatabase
	Timeout         time.Duration
	DatabaseContext *SqlDatabaseContext
	Logger          *log.Logger
}

func NewFactory(databasePath string) *SqlFactory {
	factory := SqlFactory{
		DatabaseContext: &SqlDatabaseContext{},
		Context:         execution_context.Get(),
		Logger:          log.Get(),
		Timeout:         time.Second * 30,
	}

	connStrBuilder := SqlConnectionString{}
	if err := connStrBuilder.Parse(databasePath); err != nil {
		return nil
	}

	factory.DatabaseContext.ConnectionString = &connStrBuilder

	return &factory
}

func (f *SqlFactory) WithDatabase(databaseName string) *SqlFactory {
	f.DatabaseContext.ConnectionString.WithDatabase(databaseName)
	return f
}

func (f *SqlFactory) WithTimeout(seconds int) *SqlFactory {
	f.Timeout = time.Duration(seconds * int(time.Second))
	return f
}

func (f *SqlFactory) Ping() error {
	dbConn, err := sql.Open("sqlite3", f.DatabaseContext.ConnectionString.ConnectionString())

	if err != nil {
		f.Logger.Exception(err, "error getting connection to the database")
		return err
	}

	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		f.Logger.Exception(err, "error pinging the server")
		return err
	}

	f.Logger.Info("Ping successfully")
	return nil
}

func Named(name string, value any) sql.NamedArg {
	return sql.Named(name, value)
}

func (f *SqlFactory) Connect() *sqlDatabase {
	dbConn, err := sql.Open("mysql", f.DatabaseContext.ConnectionString.ConnectionString())

	if err != nil {
		f.Logger.Exception(err, "error getting connection to the database")
		return nil
	}

	f.Database = &sqlDatabase{
		factory:        f,
		context:        context.Background(),
		defaultTimeout: f.Timeout,
		db:             dbConn,
	}

	f.Logger.Debug("Database connection created successfully")

	return f.Database
}
