package bux

import (
	"database/sql"
	"strings"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/OrlovEvgeny/go-mcache"
	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
	"github.com/vmihailenco/taskq/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm/logger"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *clientOptions)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *clientOptions {

	// Set the default options
	return &clientOptions{

		// Incoming Transaction Checker (lookup external tx via miner for validity)
		itc: true,

		// Blank chainstate config
		chainstate: &chainstateOptions{
			ClientInterface: nil,
			options:         []chainstate.ClientOps{},
		},

		// Blank cache config
		cacheStore: &cacheStoreOptions{
			ClientInterface: nil,
			options:         []cachestore.ClientOps{},
		},

		// Blank Datastore config
		dataStore: &dataStoreOptions{
			ClientInterface: nil,
			options:         []datastore.ClientOps{},
		},

		// Blank model options (use the Base models)
		models: &modelOptions{
			modelNames:        modelNames(BaseModels...),
			models:            BaseModels,
			migrateModelNames: nil,
			migrateModels:     nil,
		},

		// Blank NewRelic config
		newRelic: &newRelicOptions{},

		// Blank Paymail config
		paymail: &paymailOptions{
			client: nil,
			serverConfig: &paymailServerOptions{
				Configuration: nil,
			},
		},

		// Blank TaskManager config
		taskManager: &taskManagerOptions{
			ClientInterface: nil,
			cronTasks: map[string]time.Duration{
				ModelDraftTransaction.String() + "_clean_up":   60 * time.Second,
				ModelIncomingTransaction.String() + "_process": 30 * time.Second,
				ModelSyncTransaction.String() + "_broadcast":   30 * time.Second,
				ModelSyncTransaction.String() + "_sync":        30 * time.Second,
			},
		},

		// Default user agent
		userAgent: defaultUserAgent,
	}
}

// modelNames will take a list of models and return the list of names
func modelNames(models ...interface{}) (names []string) {
	for _, modelInterface := range models {
		names = append(names, modelInterface.(ModelInterface).Name())
	}
	return
}

// modelExists will return true if the model is found
func (o *clientOptions) modelExists(modelName, list string) bool {
	m := o.models.modelNames
	if list == migrateList {
		m = o.models.migrateModelNames
	}
	for _, name := range m {
		if strings.EqualFold(name, modelName) {
			return true
		}
	}
	return false
}

// addModel will add the model if it does not exist already (load once)
func (o *clientOptions) addModel(model interface{}, list string) {
	name := model.(ModelInterface).Name()
	if !o.modelExists(name, list) {
		if list == migrateList {
			o.models.migrateModelNames = append(o.models.migrateModelNames, name)
			o.models.migrateModels = append(o.models.migrateModels, model)
			return
		}
		o.models.modelNames = append(o.models.modelNames, name)
		o.models.models = append(o.models.models, model)
	}
}

// addModels will add the models if they do not exist already (load once)
func (o *clientOptions) addModels(list string, models ...interface{}) {
	for _, modelInterface := range models {
		o.addModel(modelInterface, list)
	}
}

// DefaultModelOptions will set any default model options (from Client options->model)
func (c *Client) DefaultModelOptions(opts ...ModelOps) []ModelOps {

	// Set the Client from the bux.Client onto the model
	opts = append(opts, WithClient(c))

	// Return the new options
	return opts
}

// -----------------------------------------------------------------
// GENERAL
// -----------------------------------------------------------------

// WithUserAgent will overwrite the default useragent
func WithUserAgent(userAgent string) ClientOps {
	return func(c *clientOptions) {
		if len(userAgent) > 0 {
			c.userAgent = userAgent
		}
	}
}

// WithNewRelic will set the NewRelic application client
func WithNewRelic(app *newrelic.Application) ClientOps {
	return func(c *clientOptions) {
		// Disregard if the app is nil
		if app == nil {
			return
		}

		// Set the app
		c.newRelic.app = app

		// Enable New relic on other services
		if c.chainstate != nil {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithNewRelic())
		}
		if c.cacheStore != nil {
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithNewRelic())
		}
		if c.dataStore != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithNewRelic())
		}
		if c.taskManager != nil {
			c.taskManager.options = append(c.taskManager.options, taskmanager.WithNewRelic())
		}

		// Enable the service
		c.newRelic.enabled = true
	}
}

// WithDebugging will set debugging in any applicable configuration
func WithDebugging() ClientOps {
	return func(c *clientOptions) {
		c.debug = true

		// Enable debugging on other services
		if c.chainstate != nil {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithDebugging())
		}
		if c.cacheStore != nil {
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithDebugging())
		}
		if c.dataStore != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithDebugging())
		}
		if c.taskManager != nil {
			c.taskManager.options = append(c.taskManager.options, taskmanager.WithDebugging())
		}
	}
}

// WithModels will add additional models (will NOT migrate using datastore)
//
// Pointers of structs (IE: &models.Xpub{})
func WithModels(models ...interface{}) ClientOps {
	return func(c *clientOptions) {
		if len(models) > 0 {
			c.addModels(modelList, models...)
		}
	}
}

// WithITCDisabled will disable (ITC) incoming transaction checking
func WithITCDisabled() ClientOps {
	return func(c *clientOptions) {
		c.itc = false
	}
}

// WithLogger will set the custom logger interface
func WithLogger(customLogger logger.Interface) ClientOps {
	return func(c *clientOptions) {
		if customLogger != nil {
			c.logger = customLogger

			// Enable debugging on other services
			if c.chainstate != nil {
				c.chainstate.options = append(c.chainstate.options, chainstate.WithLogger(c.logger))
			}
			if c.taskManager != nil {
				c.taskManager.options = append(c.taskManager.options, taskmanager.WithLogger(c.logger))
			}
			if c.dataStore != nil {
				c.dataStore.options = append(c.dataStore.options, datastore.WithLogger(c.logger))
			}
		}
	}
}

// -----------------------------------------------------------------
// CACHESTORE
// -----------------------------------------------------------------

// WithCustomCachestore will set the cachestore
func WithCustomCachestore(cacheStore cachestore.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if cacheStore != nil {
			c.cacheStore.ClientInterface = cacheStore
		}
	}
}

// WithMcache will set the cache client for both Read & Write clients
func WithMcache() ClientOps {
	return func(c *clientOptions) {
		c.cacheStore.options = append(c.cacheStore.options, cachestore.WithMcache())
	}
}

// WithMcacheConnection will set the cache client to an active mcache driver connection
func WithMcacheConnection(driver *mcache.CacheDriver) ClientOps {
	return func(c *clientOptions) {
		if driver != nil {
			c.cacheStore.options = append(
				c.cacheStore.options,
				cachestore.WithMcacheConnection(driver),
			)
		}
	}
}

// WithRistretto will set the cache client for both Read & Write clients
func WithRistretto(config *ristretto.Config) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithRistretto(config))
		}
	}
}

// WithRistrettoConnection will set the cache client to an active Ristretto connection
func WithRistrettoConnection(client *ristretto.Cache) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.cacheStore.options = append(
				c.cacheStore.options,
				cachestore.WithRistrettoConnection(client),
			)
		}
	}
}

// WithRedis will set the redis cache client for both Read & Write clients
//
// This will load new redis connections using the given parameters
func WithRedis(config *cachestore.RedisConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithRedis(config))
		}
	}
}

// WithRedisConnection will set the cache client to an active redis connection
func WithRedisConnection(activeClient *cache.Client) ClientOps {
	return func(c *clientOptions) {
		if activeClient != nil {
			c.cacheStore.options = append(
				c.cacheStore.options,
				cachestore.WithRedisConnection(activeClient),
			)
		}
	}
}

// -----------------------------------------------------------------
// DATASTORE
// -----------------------------------------------------------------

// WithCustomDatastore will set the datastore
func WithCustomDatastore(dataStore datastore.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if dataStore != nil {
			c.dataStore.ClientInterface = dataStore
		}
	}
}

// WithAutoMigrate will enable auto migrate database mode (given models)
//
// Pointers of structs (IE: &models.Xpub{})
func WithAutoMigrate(migrateModels ...interface{}) ClientOps {
	return func(c *clientOptions) {
		if len(migrateModels) > 0 {
			c.addModels(modelList, migrateModels...)
			c.addModels(migrateList, migrateModels...)
		}
	}
}

// WithSQLite will set the Datastore to use SQLite
func WithSQLite(config *datastore.SQLiteConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithSQLite(config))
		}
	}
}

// WithSQL will load a Datastore using either an SQL database config or existing connection
func WithSQL(engine datastore.Engine, config *datastore.SQLConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil && !engine.IsEmpty() {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithSQL(engine, []*datastore.SQLConfig{config}),
			)
		}
	}
}

// WithSQLConnection will set the Datastore to an existing connection for MySQL or PostgreSQL
func WithSQLConnection(engine datastore.Engine, sqlDB *sql.DB, tablePrefix string) ClientOps {
	return func(c *clientOptions) {
		if sqlDB != nil && !engine.IsEmpty() {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithSQLConnection(engine, sqlDB, tablePrefix),
			)
		}
	}
}

// WithMongoDB will set the Datastore to use MongoDB
func WithMongoDB(config *datastore.MongoDBConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithMongo(config))
		}
	}
}

// WithMongoConnection will set the Datastore to an existing connection for MongoDB
func WithMongoConnection(database *mongo.Database, tablePrefix string) ClientOps {
	return func(c *clientOptions) {
		if database != nil {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithMongoConnection(database, tablePrefix),
			)
		}
	}
}

// -----------------------------------------------------------------
// PAYMAIL
// -----------------------------------------------------------------

// WithPaymailClient will set a custom paymail client
func WithPaymailClient(client paymail.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.paymail.client = client
		}
	}
}

// WithPaymailServer will set the server configuration for Paymail
func WithPaymailServer(config *server.Configuration, defaultFromPaymail, defaultNote string) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.paymail.serverConfig.Configuration = config
		}
		if len(defaultFromPaymail) > 0 {
			c.paymail.serverConfig.DefaultFromPaymail = defaultFromPaymail
		}
		if len(defaultNote) > 0 {
			c.paymail.serverConfig.DefaultNote = defaultNote
		}
	}
}

// -----------------------------------------------------------------
// TASK MANAGER
// -----------------------------------------------------------------

// WithCustomTaskManager will set the taskmanager
func WithCustomTaskManager(taskManager taskmanager.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if taskManager != nil {
			c.taskManager.ClientInterface = taskManager
		}
	}
}

// WithTaskQ will set the task manager to use TaskQ & in-memory
func WithTaskQ(config *taskq.QueueOptions, factory taskmanager.Factory) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.taskManager.options = append(
				c.taskManager.options,
				taskmanager.WithTaskQ(config, factory),
			)
		}
	}
}

// WithTaskQUsingRedis will set the task manager to use TaskQ & Redis
func WithTaskQUsingRedis(config *taskq.QueueOptions, redisOptions *redis.Options) ClientOps {
	return func(c *clientOptions) {
		if config != nil {

			// Create a new redis client
			if config.Redis == nil {
				config.Redis = redis.NewClient(redisOptions)
			}

			c.taskManager.options = append(
				c.taskManager.options,
				taskmanager.WithTaskQ(config, taskmanager.FactoryRedis),
			)
		}
	}
}

// -----------------------------------------------------------------
// CHAIN-STATE
// -----------------------------------------------------------------

// WithCustomChainstate will set the chainstate
func WithCustomChainstate(chainState chainstate.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if chainState != nil {
			c.chainstate.ClientInterface = chainState
		}
	}
}

// todo: finish these options for loading chainstate!
