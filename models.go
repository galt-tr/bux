package bux

import (
	"context"
	"reflect"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
)

// Model is the generic model field(s) and interface(s)
//
// gorm: https://gorm.io/docs/models.html
type Model struct {
	// ModelInterface `json:"-" toml:"-" yaml:"-" gorm:"-"` (@mrz: not needed, all models implement all methods)
	// ID string  `json:"id" toml:"id" yaml:"id" gorm:"primaryKey"`  (@mrz: custom per table)

	CreatedAt time.Time `json:"created_at" toml:"created_at" yaml:"created_at" gorm:"comment:The time that the record was originally created" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" toml:"updated_at" yaml:"updated_at" gorm:"comment:The time that the record was last updated" bson:"updated_at,omitempty"`
	Metadata  Metadata  `gorm:"type:json" json:"metadata,omitempty" bson:"metadata,omitempty"`

	// https://gorm.io/docs/indexes.html
	// DeletedAt gorm.DeletedAt `json:"deleted_at" toml:"deleted_at" yaml:"deleted_at" (@mrz: this was the original type)
	DeletedAt utils.NullTime `json:"deleted_at" toml:"deleted_at" yaml:"deleted_at" gorm:"index;comment:The time the record was marked as deleted" bson:"deleted_at,omitempty"`

	// Private fields
	client     ClientInterface // Interface of the parent Client that loaded this bux model
	name       ModelName       // Name of model (table name)
	newRecord  bool            // Determine if the record is new (create vs update)
	rawXpubKey string          // Used on "CREATE" on some instances
}

// ModelInterface is the interface that all models share
type ModelInterface interface {
	AfterCreated(ctx context.Context) (err error)
	AfterDeleted(ctx context.Context) (err error)
	AfterUpdated(ctx context.Context) (err error)
	BeforeCreating(ctx context.Context) (err error)
	BeforeUpdating(ctx context.Context) (err error)
	ChildModels() []ModelInterface
	Client() ClientInterface
	DebugLog(text string)
	Display() interface{}
	GetID() string
	GetModelName() string
	GetModelTableName() string
	GetOptions(isNewRecord bool) (opts []ModelOps)
	IsNew() bool
	Migrate(client datastore.ClientInterface) error
	Name() string
	New()
	NotNew()
	RawXpub() string
	RegisterTasks() error
	Save(ctx context.Context) (err error)
	SetOptions(opts ...ModelOps)
	SetRecordTime(bool)
}

// ModelName is the model name type
type ModelName string

// NewBaseModel create an empty base model
func NewBaseModel(name ModelName, opts ...ModelOps) (m *Model) {
	m = &Model{name: name}
	m.SetOptions(opts...)
	return
}

// DisplayModels process the (slice) of model(s) for display
func DisplayModels(models interface{}) interface{} {
	if models == nil {
		return nil
	}

	s := reflect.ValueOf(models)
	if s.IsNil() {
		return nil
	}
	if s.Kind() == reflect.Slice {
		for i := 0; i < s.Len(); i++ {
			s.Index(i).MethodByName("Display").Call([]reflect.Value{})
		}
	} else {
		s.MethodByName("Display").Call([]reflect.Value{})
	}

	return models
}

// String is the string version of the name
func (n ModelName) String() string {
	return string(n)
}

// IsEmpty tests if the model name is empty
func (n ModelName) IsEmpty() bool {
	return n == ModelNameEmpty
}
