package indexer

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *Manager) Create(ctx context.Context) (*Result, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), m.options.Timeout)
		defer cancel()
	}

	start := time.Now()
	result := &Result{
		Failures: []FailureDetail{},
	}

	for _, def := range m.indexes {
		if m.options.SkipIfExists {
			exists, err := m.indexExists(ctx, def.Collection, def.Index.Options.Name)
			if err == nil && exists {
				log.Printf("Index %s on %s already exists, skipping", *def.Index.Options.Name, def.Collection)
				result.SuccessCount++
				continue
			}
		}

		collection := m.db.Collection(def.Collection)
		indexName, err := collection.Indexes().CreateOne(ctx, def.Index)

		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				log.Printf("Warning: Cannot create unique index on %s due to duplicate data", def.Collection)
			} else {
				log.Printf("Failed to create index on %s: %v", def.Collection, err)
			}

			result.FailedCount++
			name := ""
			if def.Index.Options != nil && def.Index.Options.Name != nil {
				name = *def.Index.Options.Name
			}
			result.Failures = append(result.Failures, FailureDetail{
				Collection: def.Collection,
				IndexName:  name,
				Error:      err,
			})

			if !m.options.ContinueOnError {
				result.Duration = time.Since(start)
				return result, err
			}
			continue
		}

		log.Printf("Created index %s on collection %s", indexName, def.Collection)
		result.SuccessCount++
	}

	result.Duration = time.Since(start)

	if result.FailedCount > 0 {
		return result, fmt.Errorf("%d indexes failed to create", result.FailedCount)
	}

	return result, nil
}

func (m *Manager) Drop(ctx context.Context, collections ...string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), m.options.Timeout)
		defer cancel()
	}

	targetCollections := collections
	if len(targetCollections) == 0 {
		colls := make(map[string]bool)
		for _, def := range m.indexes {
			colls[def.Collection] = true
		}
		for coll := range colls {
			targetCollections = append(targetCollections, coll)
		}
	}

	for _, collName := range targetCollections {
		collection := m.db.Collection(collName)
		if _, err := collection.Indexes().DropAll(ctx); err != nil {
			if !m.options.ContinueOnError {
				return fmt.Errorf("failed to drop indexes for %s: %w", collName, err)
			}
			log.Printf("Failed to drop indexes for %s: %v", collName, err)
		} else {
			log.Printf("Dropped all indexes for collection %s", collName)
		}
	}

	return nil
}

func (m *Manager) List(ctx context.Context, collection string) ([]bson.M, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), m.options.Timeout)
		defer cancel()
	}

	coll := m.db.Collection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}

	return indexes, nil
}

func (m *Manager) indexExists(ctx context.Context, collection string, indexName *string) (bool, error) {
	if indexName == nil {
		return false, nil
	}

	indexes, err := m.List(ctx, collection)
	if err != nil {
		return false, err
	}

	for _, idx := range indexes {
		if name, ok := idx["name"].(string); ok && name == *indexName {
			return true, nil
		}
	}

	return false, nil
}

func (m *Manager) LoadFromDefinitions(definitions []IndexDefinition) *Manager {
	m.indexes = append(m.indexes, definitions...)
	return m
}

func (m *Manager) Clear() *Manager {
	m.indexes = []IndexDefinition{}
	return m
}
