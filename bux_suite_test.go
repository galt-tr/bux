package bux

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/tester"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dolthub/go-mysql-server/server"
	embeddedPostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/mrz1836/go-logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tryvium-travels/memongo"
)

const (
	// testLocalConnectionURL   = "redis://localhost:6379"
	defaultDatabaseName      = "bux"
	defaultNewRelicTx        = "testing-transaction"
	defaultNewRelicApp       = "testing-app"
	mySQLHost                = "localhost"
	mySQLPassword            = ""
	mySQLTestPort            = uint32(3307)
	mySQLUsername            = "root"
	postgresqlTestHost       = "localhost"
	postgresqlTestName       = "postgres"
	postgresqlTestPort       = uint32(61333)
	postgresqlTestUser       = "postgres"
	postgresTestPassword     = "postgres"
	testIdleTimeout          = 240 * time.Second
	testMaxActiveConnections = 0
	testMaxConnLifetime      = 60 * time.Second
	testMaxIdleConnections   = 10
	testQueueName            = "test_queue"
)

// dbTestCase is a database test case
type dbTestCase struct {
	name     string
	database datastore.Engine
}

// dbTestCases is the list of supported databases
var dbTestCases = []dbTestCase{
	{name: "[mongo] [in-memory]", database: datastore.MongoDB},
	{name: "[mysql] [in-memory]", database: datastore.MySQL},
	{name: "[postgresql] [in-memory]", database: datastore.PostgreSQL},
	{name: "[sqlite] [in-memory]", database: datastore.SQLite},
}

// EmbeddedDBTestSuite is for testing the entire package using real/mocked services
type EmbeddedDBTestSuite struct {
	suite.Suite
	MongoServer      *memongo.Server                    // In-memory  Mongo server
	MySQLServer      *server.Server                     // In-memory MySQL server
	PostgresqlServer *embeddedPostgres.EmbeddedPostgres // In-memory Postgresql server
	quit             chan interface{}                   // Channel for exiting
	wg               sync.WaitGroup                     // Workgroup for managing goroutines
}

// serveMySQL will serve the MySQL server and exit if quit
func (ts *EmbeddedDBTestSuite) serveMySQL() {
	defer ts.wg.Done()

	for {
		err := ts.MySQLServer.Start()
		if err != nil {
			select {
			case <-ts.quit:
				logger.Data(2, logger.DEBUG, "MySQL channel has closed")
				return
			default:
				logger.Data(2, logger.ERROR, "mysql server error: "+err.Error())
			}
		}
	}
}

// SetupSuite runs at the start of the suite
func (ts *EmbeddedDBTestSuite) SetupSuite() {

	var err error

	// Create the MySQL server
	if ts.MySQLServer, err = tester.CreateMySQL(
		mySQLHost, defaultDatabaseName, mySQLUsername, mySQLPassword, mySQLTestPort,
	); err != nil {
		require.NoError(ts.T(), err)
	}

	// Don't block, serve the MySQL instance
	ts.quit = make(chan interface{})
	ts.wg.Add(1)
	go ts.serveMySQL()

	// Create the Mongo server
	if ts.MongoServer, err = tester.CreateMongoServer(mongoTestVersion); err != nil {
		require.NoError(ts.T(), err)
	}

	// Create the Postgresql server
	if ts.PostgresqlServer, err = tester.CreatePostgresServer(postgresqlTestPort); err != nil {
		require.NoError(ts.T(), err)
	}

	// Fail-safe! If a test completes or fails, this is triggered
	// Embedded servers are still running on the ports given, and causes a conflict re-running tests
	ts.T().Cleanup(func() {
		ts.TearDownSuite()
	})
}

// TearDownSuite runs after the suite finishes
func (ts *EmbeddedDBTestSuite) TearDownSuite() {

	// Stop the Mongo server
	if ts.MongoServer != nil {
		ts.MongoServer.Stop()
	}

	// Stop the postgresql server
	if ts.PostgresqlServer != nil {
		_ = ts.PostgresqlServer.Stop()
	}

	// Stop the MySQL server
	if ts.MySQLServer != nil {
		/*
			defer ts.wg.Done()
			if ts.quit != nil {
				close(ts.quit)
			}
		*/
		_ = ts.MySQLServer.Close()
	}
}

// SetupTest runs before each test
func (ts *EmbeddedDBTestSuite) SetupTest() {
	// Nothing needed here (yet)
}

// TearDownTest runs after each test
func (ts *EmbeddedDBTestSuite) TearDownTest() {
	// Nothing needed here (yet)
}

// createTestClient will make a new test client
//
// NOTE: you need to close the client: ts.Close()
func (ts *EmbeddedDBTestSuite) createTestClient(ctx context.Context, database datastore.Engine,
	tablePrefix string, mockDB, mockRedis bool, opts ...ClientOps) (*TestingClient, error) {

	var err error

	// Start the suite
	tc := &TestingClient{
		tablePrefix: tablePrefix,
		ctx:         ctx,
	}

	// Are we mocking SQL?
	if mockDB {

		// Create new SQL mocked connection
		if tc.SQLConn, tc.MockSQLDB, err = sqlmock.New(
			sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
		); err != nil {
			return nil, err
		}

		// Switch on database types
		if database == datastore.SQLite {
			opts = append(opts, WithSQLite(&datastore.SQLiteConfig{
				CommonConfig: datastore.CommonConfig{
					MaxConnectionIdleTime: 0,
					MaxConnectionTime:     0,
					MaxIdleConnections:    1,
					MaxOpenConnections:    1,
					TablePrefix:           tablePrefix,
				},
				ExistingConnection: tc.SQLConn,
			}))
		} else if database == datastore.MySQL {
			opts = append(opts, WithSQLConnection(datastore.MySQL, tc.SQLConn, tablePrefix))
		} else if database == datastore.PostgreSQL {
			opts = append(opts, WithSQLConnection(datastore.PostgreSQL, tc.SQLConn, tablePrefix))
		} else { // todo: finish more Datastore support (missing: Mongo)
			// "https://medium.com/@victor.neuret/mocking-the-official-mongo-golang-driver-5aad5b226a78"
			return nil, ErrDatastoreNotSupported
		}

	} else {

		// Load the in-memory version of the database
		if database == datastore.SQLite {
			opts = append(opts, WithSQLite(&datastore.SQLiteConfig{
				CommonConfig: datastore.CommonConfig{
					MaxIdleConnections: 1,
					MaxOpenConnections: 1,
					TablePrefix:        tablePrefix,
				},
				Shared: true,
			}))
		} else if database == datastore.MongoDB {

			// Sanity check
			if ts.MongoServer == nil {
				return nil, ErrLoadServerFirst
			}

			// Add the new Mongo connection
			opts = append(opts, WithMongoDB(&datastore.MongoDBConfig{
				CommonConfig: datastore.CommonConfig{
					MaxIdleConnections: 1,
					MaxOpenConnections: 1,
					TablePrefix:        tablePrefix,
				},
				URI:          ts.MongoServer.URIWithRandomDB(),
				DatabaseName: memongo.RandomDatabase(),
			}))

		} else if database == datastore.PostgreSQL {

			// Sanity check
			if ts.PostgresqlServer == nil {
				return nil, ErrLoadServerFirst
			}

			// Add the new Postgresql connection
			opts = append(opts, WithSQL(datastore.PostgreSQL, &datastore.SQLConfig{
				CommonConfig: datastore.CommonConfig{
					MaxIdleConnections: 1,
					MaxOpenConnections: 1,
					TablePrefix:        tablePrefix,
				},
				Host:                      postgresqlTestHost,
				Name:                      postgresqlTestName,
				User:                      postgresqlTestUser,
				Password:                  postgresTestPassword,
				Port:                      fmt.Sprintf("%d", postgresqlTestPort),
				SkipInitializeWithVersion: true,
			}))

		} else if database == datastore.MySQL {

			// Sanity check
			if ts.MySQLServer == nil {
				return nil, ErrLoadServerFirst
			}

			// Add the new Postgresql connection
			opts = append(opts, WithSQL(datastore.MySQL, &datastore.SQLConfig{
				CommonConfig: datastore.CommonConfig{
					MaxIdleConnections: 1,
					MaxOpenConnections: 1,
					TablePrefix:        tablePrefix,
				},
				Host:                      mySQLHost,
				Name:                      defaultDatabaseName,
				User:                      mySQLUsername,
				Password:                  mySQLPassword,
				Port:                      fmt.Sprintf("%d", mySQLTestPort),
				SkipInitializeWithVersion: true,
			}))

		} else {
			return nil, ErrDatastoreNotSupported
		}
	}

	// Custom for SQLite and Mocking (cannot ignore the version check that GORM does)
	if mockDB && database == datastore.SQLite {
		tc.MockSQLDB.ExpectQuery(
			"select sqlite_version()",
		).WillReturnRows(tc.MockSQLDB.NewRows([]string{"version"}).FromCSVString(sqliteTestVersion))
	}

	// Are we mocking redis?
	if mockRedis {
		tc.redisClient, tc.redisConn = tester.LoadMockRedis(
			testIdleTimeout,
			testMaxConnLifetime,
			testMaxActiveConnections,
			testMaxIdleConnections,
		)
		opts = append(opts, WithRedisConnection(tc.redisClient))
	}

	// Add a custom user agent (future: make this passed into the function via opts)
	opts = append(opts, WithUserAgent("bux test suite"))

	// Create the client
	if tc.client, err = NewClient(ctx, opts...); err != nil {
		return nil, err
	}

	// Return the suite
	return tc, nil
}

// genericDBClient is a helpful wrapper for getting the same type of client
//
// NOTE: you need to close the client: ts.Close()
func (ts *EmbeddedDBTestSuite) genericDBClient(t *testing.T, database datastore.Engine, taskManagerEnabled bool) *TestingClient {
	prefix := tester.RandomTablePrefix(t)

	var opts []ClientOps
	opts = append(opts,
		WithDebugging(),
		WithAutoMigrate(BaseModels...),
		WithRistretto(cachestore.DefaultRistrettoConfig()))
	if taskManagerEnabled {
		opts = append(opts, WithTaskQ(taskmanager.DefaultTaskQConfig(prefix+"_queue"), taskmanager.FactoryMemory))
	} else {
		opts = append(opts, WithCustomTaskManager(&taskManagerMockBase{}))
	}

	tc, err := ts.createTestClient(
		tester.GetNewRelicCtx(
			t, defaultNewRelicApp, defaultNewRelicTx,
		),
		database, prefix,
		false, false,
		opts...,
	)
	require.NoError(t, err)
	require.NotNil(t, tc)
	return tc
}

// genericMockedDBClient is a helpful wrapper for getting the same type of client
//
// NOTE: you need to close the client: ts.Close()
func (ts *EmbeddedDBTestSuite) genericMockedDBClient(t *testing.T, database datastore.Engine) *TestingClient {
	prefix := tester.RandomTablePrefix(t)
	tc, err := ts.createTestClient(
		tester.GetNewRelicCtx(
			t, defaultNewRelicApp, defaultNewRelicTx,
		),
		database, prefix,
		true, true, WithDebugging(),
		WithCustomTaskManager(&taskManagerMockBase{}),
		WithRistretto(cachestore.DefaultRistrettoConfig()),
	)
	require.NoError(t, err)
	require.NotNil(t, tc)
	return tc
}
