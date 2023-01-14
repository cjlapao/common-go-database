package sqlite

import (
	"strings"

	"github.com/cjlapao/common-go/execution_context"
	"github.com/cjlapao/common-go/guard"
	"github.com/cjlapao/common-go/log"
)

// Global Sql service to keep single service for consumers
var globalSqlService *SqlService

// Global global database factory to keep a single Sql client
var globalFactory *SqlFactory

// Global tenant database factory to keep a single Sql client
var tenantFactory *SqlFactory

// SqlServiceOptions structure
type SqlServiceOptions struct {
	ConnectionString   string
	GlobalDatabaseName string
}

// SqlService structure
type SqlService struct {
	ConnectionString   string
	GlobalDatabaseName string
	TenantDatabaseName string
	logger             *log.Logger
}

// New Creates a Sql service using the default configuration
// This uses the environment variables to define the connection
// and the database name, the variables are:
// SQLDB_CONNECTION_STRING: for the connection string
// SQLDB_DATABASENAME: for the database name
// returns a SQLDBService pointer
func New() *SqlService {
	ctx := execution_context.Get()
	connStr := ctx.Configuration.GetString("SQL_CONNECTION_STRING")
	globalDatabaseName := ctx.Configuration.GetString("SQL_DATABASE_NAME")
	options := SqlServiceOptions{
		ConnectionString:   connStr,
		GlobalDatabaseName: globalDatabaseName,
	}

	return NewWithOptions(options)
}

// NewWithOptions Creates a SQL service passing the options object
// returns a SqlService pointer
func NewWithOptions(options SqlServiceOptions) *SqlService {
	service := SqlService{
		ConnectionString:   options.ConnectionString,
		GlobalDatabaseName: options.GlobalDatabaseName,
		logger:             log.Get(),
	}

	if options.ConnectionString != "" && options.GlobalDatabaseName != "" {
		globalFactory = NewFactory(service.ConnectionString).WithDatabase(service.GlobalDatabaseName)
	}

	globalSqlService = &service
	return globalSqlService
}

// Init initiates the SqlDb service and global database factory
// returns a SqlDbService pointer
func Init() *SqlService {
	ctx := execution_context.Get()
	logger := ctx.Services.Logger
	if globalSqlService != nil {
		if globalSqlService.ConnectionString != "" && globalSqlService.GlobalDatabaseName != "" {
			logger.Info("Initiating Sql Service for global database %v", globalSqlService.GlobalDatabaseName)
			globalFactory = NewFactory(globalSqlService.ConnectionString).WithDatabase(globalSqlService.GlobalDatabaseName)
			logger.Info("Sql Service for global database %v initiated successfully", globalSqlService.GlobalDatabaseName)
		}
		return globalSqlService
	}

	return New()
}

// Get Gets the current global service
// returns a SqlService pointer
func Get() *SqlService {
	if globalSqlService != nil {
		return globalSqlService
	}

	return New()
}

// WithDatabase Sets the global database name
// returns a SqlService pointer
func (service *SqlService) WithDatabase(databaseName string) *SqlService {
	guard.FatalEmptyOrNil(databaseName, "Database name is empty")
	if !strings.EqualFold(service.GlobalDatabaseName, databaseName) {
		service.GlobalDatabaseName = databaseName
		Init()
	}

	return service
}

// GlobalDatabase Gets the global database factory and initiate it ready for consumption.
// This will try to only keep a client per session to avoid starvation of the clients
// returns a SqlDbFactory pointer
func (service *SqlService) GlobalDatabase() *SqlFactory {
	if globalFactory == nil {
		service.logger.Info("Global factory not initiated, creating instance now.")
		Init()
	}

	return globalFactory
}

// TenantDatabase Gets the tenant database factory and initiate it ready for consumption
// if there is no tenant set this will bring the global database and we will treat it as
// a single tenant system.
// This will try to only keep a client per session to avoid starvation of the clients
// returns a SqlFactory pointer
func (service *SqlService) TenantDatabase() *SqlFactory {
	ctx := execution_context.Get()
	tenantId := ""
	if ctx.Authorization != nil {
		tenantId = ctx.Authorization.TenantId
	}

	if tenantId == "" || strings.ToLower(tenantId) == "global" {
		return service.GlobalDatabase()
	}
	if !strings.EqualFold(tenantId, service.TenantDatabaseName) {
		service.TenantDatabaseName = tenantId
		service.logger.Info("Initiating Sql Service for tenant database %v", service.TenantDatabaseName)
		tenantFactory = NewFactory(service.ConnectionString).WithDatabase(service.TenantDatabaseName)
		service.logger.Info("Sql Service for tenant database %v initiated successfully", service.TenantDatabaseName)
	}

	return tenantFactory
}
