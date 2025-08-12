package indexer

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const migrationCollection = "_index_migrations"

type MigrationManager struct {
	db         *mongo.Database
	migrations []Migration
}

func NewMigrationManager(db *mongo.Database) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: []Migration{},
	}
}

func (mm *MigrationManager) AddMigration(migration Migration) *MigrationManager {
	mm.migrations = append(mm.migrations, migration)
	return mm
}

func (mm *MigrationManager) Run(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
	}

	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})

	coll := mm.db.Collection(migrationCollection)
	
	_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create migration index: %w", err)
	}

	for _, migration := range mm.migrations {
		applied, err := mm.isApplied(ctx, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}

		if applied {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Running migration %s: %s", migration.Version, migration.Description)
		
		start := time.Now()
		err = migration.Up(mm.db)
		duration := time.Since(start)

		status := MigrationStatus{
			Version:   migration.Version,
			AppliedAt: time.Now(),
			Success:   err == nil,
		}

		if err != nil {
			log.Printf("Migration %s failed after %v: %v", migration.Version, duration, err)
			if _, saveErr := coll.InsertOne(ctx, status); saveErr != nil {
				log.Printf("Failed to save migration status: %v", saveErr)
			}
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		if _, err = coll.InsertOne(ctx, status); err != nil {
			return fmt.Errorf("failed to save migration status: %w", err)
		}

		log.Printf("Migration %s completed successfully in %v", migration.Version, duration)
	}

	return nil
}

func (mm *MigrationManager) Rollback(ctx context.Context, targetVersion string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
	}

	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version > mm.migrations[j].Version
	})

	coll := mm.db.Collection(migrationCollection)
	
	for _, migration := range mm.migrations {
		if migration.Version <= targetVersion {
			break
		}

		applied, err := mm.isApplied(ctx, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}

		if !applied {
			continue
		}

		if migration.Down == nil {
			return fmt.Errorf("migration %s does not support rollback", migration.Version)
		}

		log.Printf("Rolling back migration %s", migration.Version)
		
		if err := migration.Down(mm.db); err != nil {
			return fmt.Errorf("rollback of migration %s failed: %w", migration.Version, err)
		}

		if _, err := coll.DeleteOne(ctx, bson.M{"version": migration.Version}); err != nil {
			return fmt.Errorf("failed to remove migration status: %w", err)
		}

		log.Printf("Successfully rolled back migration %s", migration.Version)
	}

	return nil
}

func (mm *MigrationManager) Status(ctx context.Context) ([]MigrationStatus, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	coll := mm.db.Collection(migrationCollection)
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "version", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %w", err)
	}
	defer cursor.Close(ctx)

	var statuses []MigrationStatus
	if err = cursor.All(ctx, &statuses); err != nil {
		return nil, fmt.Errorf("failed to decode migration statuses: %w", err)
	}

	return statuses, nil
}

func (mm *MigrationManager) isApplied(ctx context.Context, version string) (bool, error) {
	coll := mm.db.Collection(migrationCollection)
	count, err := coll.CountDocuments(ctx, bson.M{"version": version, "success": true})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}