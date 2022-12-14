package mongodb

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/cjlapao/common-go/execution_context"
	"github.com/cjlapao/common-go/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient interface {
	Database(string) MongoDatabaseClient
	Connect() error
	StartSession() (mongoSession, error)
}

type MongoDatabaseClient interface {
	Collection(name string)
}

type MongoCollectionClient interface {
	Find(context.Context, interface{}) (mongoCursor, error)
	FindOne(context.Context, interface{}) mongoSingleResult
	InsertOne(context.Context, interface{}) (interface{}, error)
}

type mongoClient struct {
	factory *MongoFactory
	cl      *mongo.Client
}

type mongoDatabase struct {
	name    string
	factory *MongoFactory
	db      *mongo.Database
}

type mongoCollection struct {
	name    string
	factory *MongoFactory
	coll    *mongo.Collection
}

// Repository Gets the repository for this collection
func (collection mongoCollection) Repository() MongoRepository {
	return collection.factory.NewRepository(collection.name)
}

// Repository Gets the repository for this collection
func (collection mongoCollection) OData() *ODataParser {
	return EmptyODataParser(&collection)
}

type mongoSession struct {
	mongo.Session
}

type mongoCursor struct {
	cursor *mongo.Cursor
}

func (cursor mongoCursor) Decode(destination interface{}) error {
	var destType = reflect.TypeOf(destination)
	if destType.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer type")
	}

	return cursor.cursor.Decode(destination)
}

func (cursor mongoCursor) DecodeAll(destination interface{}) error {
	ctx := context.Background()
	var destType = reflect.TypeOf(destination)
	if destType.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer type")
	}

	return cursor.cursor.All(ctx, destination)
}

func (cursor mongoCursor) Next(ctx context.Context) bool {
	return cursor.cursor.Next(ctx)
}

func (cursor mongoCursor) Current() interface{} {
	return cursor.cursor.Current
}

type mongoSingleResult struct {
	sr *mongo.SingleResult
}

func (cursor mongoSingleResult) Decode(destination interface{}) error {
	var destType = reflect.TypeOf(destination)
	if destType.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer type")
	}

	return cursor.sr.Decode(destination)
}

func (cursor mongoSingleResult) Err() error {
	return cursor.sr.Err()
}

type mongoInsertOneResult struct {
	result     *mongo.InsertOneResult
	InsertedId interface{}
}

func (result *mongoInsertOneResult) FromMongo(entity *mongo.InsertOneResult) {
	result.result = entity
	result.InsertedId = entity.InsertedID
}

type mongoInsertManyResult struct {
	result      *mongo.InsertManyResult
	InsertedIds []interface{}
}

func (result *mongoInsertManyResult) FromMongo(entity *mongo.InsertManyResult) {
	result.result = entity
	result.InsertedIds = entity.InsertedIDs
}

type mongoUpdateResult struct {
	result        *mongo.UpdateResult
	MatchedCount  int64       // The number of documents matched by the filter.
	ModifiedCount int64       // The number of documents modified by the operation.
	UpsertedCount int64       // The number of documents upserted by the operation.
	UpsertedID    interface{} // The _id field of the upserted document, or nil if no upsert was done.
}

func (result *mongoUpdateResult) FromMongo(entity *mongo.UpdateResult) {
	result.result = entity
	result.MatchedCount = entity.MatchedCount
	result.ModifiedCount = entity.ModifiedCount
	result.UpsertedCount = entity.UpsertedCount
	result.UpsertedID = entity.UpsertedID
}

type mongoBulkWriteResult struct {
	result        *mongo.BulkWriteResult
	InsertedCount int64                 // The number of documents inserted.
	MatchedCount  int64                 // The number of documents matched by filters in update and replace operations.
	ModifiedCount int64                 // The number of documents modified by update and replace operations.
	DeletedCount  int64                 // The number of documents deleted.
	UpsertedCount int64                 // The number of documents upserted by update and replace operations.
	UpsertedIDs   map[int64]interface{} // A map of operation index to the _id of each upserted document.
}

func (result *mongoBulkWriteResult) FromMongo(entity *mongo.BulkWriteResult) {
	result.result = entity
	result.InsertedCount = entity.InsertedCount
	result.MatchedCount = entity.MatchedCount
	result.ModifiedCount = entity.ModifiedCount
	result.UpsertedCount = entity.UpsertedCount
	result.UpsertedIDs = entity.UpsertedIDs
}

type mongoDeleteResult struct {
	result       *mongo.DeleteResult
	DeletedCount int64 // The number of documents deleted.
}

func (result *mongoDeleteResult) FromMongo(entity *mongo.DeleteResult) {
	result.result = entity
	result.DeletedCount = entity.DeletedCount
}

// MongoFactory MongoFactory Entity
type MongoFactory struct {
	Context         *execution_context.Context
	Client          *mongoClient
	Database        *mongoDatabase
	DatabaseContext *MongoDatabaseContext
	Logger          *log.Logger
}

// NewFactory Creates a brand new factory for a specific connection string
// this will create and attach a mongo client that it will use for all connections
// returns a pointer to a MongoFactory object
func NewFactory(connectionString string) *MongoFactory {
	factory := MongoFactory{}
	factory.DatabaseContext = &MongoDatabaseContext{
		ConnectionString: connectionString,
	}

	factory.Logger = log.Get()
	factory.Context = execution_context.Get()

	factory.GetClient()
	factory.Logger.Info("MongoDB Factory initiated successfully.")
	return &factory
}

func (f *MongoFactory) WithDatabase(databaseName string) *MongoFactory {
	f.GetDatabase(databaseName)
	f.Logger.Info("MongoDB Factory database %v initiated successfully.", databaseName)
	return f
}

// GetClient This will either return an already initiated client or the current
// active client in the factory, this will avoid having unclosed clients
// if you need a brand new client please use the NewFactory method to create a brand
// new factory.
// returns a mongoClient object
func (f *MongoFactory) GetClient() *mongoClient {

	if f.Client != nil {
		return f.Client
	}

	connectionString := f.DatabaseContext.ConnectionString
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		f.Logger.LogError(err)
		return nil
	}

	// Check the connection
	err = client.Ping(ctx, nil)

	if err != nil {
		f.Logger.LogError(err)
		return nil
	}

	f.Client = &mongoClient{
		factory: f,
		cl:      client,
	}

	f.Logger.Debug("Client connection created successfully")
	return f.Client
}

// GetDatabase Get a database from the current cluster and sets it in the database context
// returns a mongoDatabase object
func (f *MongoFactory) GetDatabase(databaseName string) *mongoDatabase {
	if f.Client == nil {
		f.Client = f.GetClient()
	}

	database := f.Client.cl.Database(databaseName)
	if database == nil {
		f.Logger.Error("There was an error getting the database %v", databaseName)
		return nil
	}

	f.DatabaseContext.CurrentDatabaseName = databaseName
	f.Database = &mongoDatabase{
		factory: f,
		db:      database,
		name:    databaseName,
	}

	f.Logger.Debug("Database was retrieved successfully")
	return f.Database
}

// GetCollection Get a collection from the current database
// returns a mongoCollection object
func (f *MongoFactory) GetCollection(collectionName string) *mongoCollection {
	if f.Client == nil {
		f.Client = f.GetClient()
	}

	if f.Database == nil {
		f.Database = f.GetDatabase(f.DatabaseContext.CurrentDatabaseName)
	}

	collection := f.Database.db.Collection(collectionName)
	if collection == nil {
		f.Logger.Error("There was an error getting the collection %v", collectionName)
		return nil
	}

	f.DatabaseContext.CurrentCollection = collectionName
	f.Logger.Debug("Collection was retrieved successfully")
	return &mongoCollection{
		factory: f,
		coll:    collection,
		name:    collectionName,
	}
}

// StartSession Starts a session in the mongodb client
func (f *MongoFactory) StartSession() (mongo.Session, error) {
	session, err := f.Client.cl.StartSession()
	return &mongoSession{Session: session}, err
}
