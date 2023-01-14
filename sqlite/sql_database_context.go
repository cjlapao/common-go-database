package sqlite

type SqlDatabaseContext struct {
	ConnectionString *SqlConnectionString
	CurrentTable     string
}

func (s *SqlDatabaseContext) CurrentDatabase() string {
	return s.ConnectionString.ConnectionString()
}
