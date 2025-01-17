package tester

import (
	"database/sql/driver"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/BuxOrg/bux/datastore"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	embeddedPostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/tryvium-travels/memongo"
)

// AnyTime will fill the need for any timestamp field
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// AnyGUID will fill the need for any GUID field
type AnyGUID struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyGUID) Match(v driver.Value) bool {
	str, ok := v.(string)
	return ok && len(str) > 0
}

// CreatePostgresServer will create a new Postgresql server
func CreatePostgresServer(port uint32) (*embeddedPostgres.EmbeddedPostgres, error) {

	// Create the new database
	postgres := embeddedPostgres.NewDatabase(embeddedPostgres.DefaultConfig().Port(port))
	if postgres == nil {
		return nil, ErrFailedLoadingPostgresql
	}

	// Start the database
	if err := postgres.Start(); err != nil {
		return nil, err
	}

	// Return the database
	return postgres, nil
}

// CreateMongoServer will create a new mongo server
func CreateMongoServer(version string) (*memongo.Server, error) {
	mongoServer, err := memongo.StartWithOptions(
		&memongo.Options{MongoVersion: version, ShouldUseReplica: true, DownloadURL: os.Getenv("BUX_MONGODB_DOWNLOAD_URL")},
	)
	if err != nil {
		return nil, err
	}

	return mongoServer, nil
}

// CreateMySQL will make a new MySQL server
func CreateMySQL(host, databaseName, username, password string, port uint32) (*server.Server, error) {
	engine := sqle.NewDefault(
		sql.NewDatabaseProvider(
			CreateMySQLTestDatabase(databaseName),
			information_schema.NewInformationSchemaDatabase(),
		))
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", host, port),
		Auth:     auth.NewNativeSingle(username, password, auth.AllPermissions),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateMySQLTestDatabase is a dummy database for MySQL
func CreateMySQLTestDatabase(databaseName string) *memory.Database {
	return memory.NewDatabase(databaseName)
}

// SQLiteTestConfig will return a test-version of SQLite
func SQLiteTestConfig(t *testing.T, debug, shared bool) *datastore.SQLiteConfig {
	return &datastore.SQLiteConfig{
		CommonConfig: datastore.CommonConfig{
			Debug:              debug,
			MaxIdleConnections: 1,
			MaxOpenConnections: 1,
			TablePrefix:        RandomTablePrefix(t),
		},
		DatabasePath: "",
		Shared:       shared,
	}
}
