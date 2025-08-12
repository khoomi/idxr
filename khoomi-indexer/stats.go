package indexer

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *Manager) Stats(ctx context.Context, collection string) ([]IndexStats, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), m.options.Timeout)
		defer cancel()
	}

	coll := m.db.Collection(collection)
	pipeline := mongo.Pipeline{
		{{Key: "$indexStats", Value: bson.D{}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}
	defer cursor.Close(ctx)

	var rawStats []bson.M
	if err = cursor.All(ctx, &rawStats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	stats := make([]IndexStats, 0, len(rawStats))
	for _, raw := range rawStats {
		stat := IndexStats{}

		if name, ok := raw["name"].(string); ok {
			stat.Name = name
		}

		if accesses, ok := raw["accesses"].(bson.M); ok {
			if ops, ok := accesses["ops"].(int64); ok {
				stat.Accesses = ops
			} else if ops, ok := accesses["ops"].(int32); ok {
				stat.Accesses = int64(ops)
			}

			if since, ok := accesses["since"].(interface{}); ok {
				if t, ok := since.(bson.M); ok {
					if dt, ok := t["$date"].(interface{}); ok {
						stat.Since = dt.(time.Time)
					}
				}
			}
		}

		if host, ok := raw["host"].(string); ok {
			stat.Host = host
		}

		if building, ok := raw["building"].(bool); ok {
			stat.Building = building
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

func (m *Manager) StatsAll(ctx context.Context) (map[string][]IndexStats, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), m.options.Timeout)
		defer cancel()
	}

	collectionNames := make(map[string]bool)
	for _, def := range m.indexes {
		collectionNames[def.Collection] = true
	}

	results := make(map[string][]IndexStats)
	for collName := range collectionNames {
		stats, err := m.Stats(ctx, collName)
		if err != nil {
			if m.options.ContinueOnError {
				results[collName] = []IndexStats{}
				continue
			}
			return nil, fmt.Errorf("failed to get stats for %s: %w", collName, err)
		}
		results[collName] = stats
	}

	return results, nil
}
