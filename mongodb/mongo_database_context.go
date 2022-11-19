package mongodb

type MongoDatabaseContext struct {
	CurrentDatabaseName string
	ConnectionString    string
	CurrentCollection   string
}
