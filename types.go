package indexer

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexDefinition struct {
	Collection string
	Index      mongo.IndexModel
}

type Manager struct {
	db      *mongo.Database
	indexes []IndexDefinition
	options *Options
}

type Options struct {
	Timeout         time.Duration
	ContinueOnError bool
	SkipIfExists    bool
	Environment     string
	Silent          bool
}

type Result struct {
	SuccessCount int
	FailedCount  int
	Failures     []FailureDetail
	Duration     time.Duration
}

type FailureDetail struct {
	Collection string
	IndexName  string
	Error      error
}

type IndexStats struct {
	Name     string
	Accesses int64
	Since    time.Time
	Host     string
	Building bool
}

type Migration struct {
	Version     string
	Description string
	Timestamp   time.Time
	Up          func(*mongo.Database) error
	Down        func(*mongo.Database) error
}

type MigrationStatus struct {
	Version   string    `bson:"version"`
	AppliedAt time.Time `bson:"applied_at"`
	Success   bool      `bson:"success"`
}

func DefaultOptions() *Options {
	return &Options{
		Timeout:         60 * time.Second,
		ContinueOnError: true,
		SkipIfExists:    true,
		Environment:     "development",
	}
}

func NewManager(db *mongo.Database, opts ...*Options) *Manager {
	var options *Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	return &Manager{
		db:      db,
		indexes: []IndexDefinition{},
		options: options,
	}
}

func (m *Manager) AddIndex(collection string, index mongo.IndexModel) *Manager {
	m.indexes = append(m.indexes, IndexDefinition{
		Collection: collection,
		Index:      index,
	})
	return m
}

func (m *Manager) AddTextIndex(collection string, fields ...string) *Manager {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}

	m.indexes = append(m.indexes, IndexDefinition{
		Collection: collection,
		Index: mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(collection + "_text_search"),
		},
	})
	return m
}

func (m *Manager) AddCompoundIndex(collection string, fields []string, opts ...*options.IndexOptions) *Manager {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}

	indexOpts := options.Index()
	if len(opts) > 0 {
		indexOpts = opts[0]
	}

	m.indexes = append(m.indexes, IndexDefinition{
		Collection: collection,
		Index: mongo.IndexModel{
			Keys:    keys,
			Options: indexOpts,
		},
	})
	return m
}
