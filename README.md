# idxr

MongoDB index management library for Go applications and in CLI.

## Features

- Declarative index definitions
- Batch index creation with error handling
- Index statistics and monitoring
- Migration support with rollback capability
- CLI tool for management operations
- Support for compound, text, and partial indexes

## Installation

```bash
go get github.com/khoomi/idxr
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/khoomi/idxr"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    client, _ := mongo.Connect(context.Background(), 
        options.Client().ApplyURI("mongodb://localhost:27017"))
    
    db := client.Database("khoomi")
    manager := indexer.NewManager(db)
    
    // Add a unique index
    manager.AddIndex("users", mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true).SetName("email_unique"),
    })
    
    // Add a compound index
    manager.AddCompoundIndex("orders", []string{"user_id", "created_at"})
    
    // Add a text search index
    manager.AddTextIndex("products", "name", "description", "tags")
    
    // Create all indexes
    result, err := manager.Create(context.Background())
    if err != nil {
        log.Printf("Some indexes failed: %v", err)
    }
    
    log.Printf("Created %d indexes, %d failed", 
        result.SuccessCount, result.FailedCount)
}
```

## CLI Usage

### Install the CLI tool

```bash
go install github.com/khoomi/idxr/cmd/idxr@latest
```

### Commands

```bash
# Create indexes
idxr -action create -db myapp

# List indexes for a collection
idxr -action list -db myapp -collection users

# Get index statistics
idxr -action stats -db myapp

# Drop indexes
idxr -action drop -db myapp users products

# JSON output
idxr -action stats -db myapp -json
```

### Environment Variables

- `DATABASE_URL`: MongoDB connection URI
- `DB_NAME`: Database name

## Migrations

```go
migrationManager := indexer.NewMigrationManager(db)

migrationManager.AddMigration(indexer.Migration{
    Version:     "001_initial",
    Description: "Initial indexes",
    Up: func(db *mongo.Database) error {
        manager := indexer.NewManager(db)
        manager.AddIndex("users", mongo.IndexModel{
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        })
        result, _ := manager.Create(context.Background())
        return nil
    },
    Down: func(db *mongo.Database) error {
        return indexer.NewManager(db).Drop(context.Background(), "users")
    },
})

// Run migrations
err := migrationManager.Run(context.Background())

// Check status
statuses, _ := migrationManager.Status(context.Background())

// Rollback to specific version
err = migrationManager.Rollback(context.Background(), "001_initial")
```

## Options

```go
opts := &indexer.Options{
    Timeout:         60 * time.Second,
    ContinueOnError: true,
    SkipIfExists:    true,
    Environment:     "production",
}

manager := indexer.NewManager(db, opts)
```

## Index Types

### Compound Index
```go
manager.AddCompoundIndex("collection", []string{"field1", "field2"})
```

### Text Search Index
```go
manager.AddTextIndex("collection", "title", "content", "tags")
```

### Partial Index
```go
manager.AddIndex("users", mongo.IndexModel{
    Keys: bson.D{{Key: "email", Value: 1}},
    Options: options.Index().
        SetPartialFilterExpression(bson.M{"active": true}),
})
```

### TTL Index
```go
manager.AddIndex("sessions", mongo.IndexModel{
    Keys: bson.D{{Key: "expires_at", Value: 1}},
    Options: options.Index().
        SetExpireAfterSeconds(3600),
})
```

## Statistics

```go
// Get stats for single collection
stats, err := manager.Stats(context.Background(), "users")
for _, stat := range stats {
    fmt.Printf("%s: %d accesses since %v\n", 
        stat.Name, stat.Accesses, stat.Since)
}

// Get stats for all managed collections
allStats, err := manager.StatsAll(context.Background())
```

## Error Handling

The library provides detailed error information:

```go
result, err := manager.Create(context.Background())

if result.FailedCount > 0 {
    for _, failure := range result.Failures {
        log.Printf("Failed: %s.%s - %v", 
            failure.Collection, 
            failure.IndexName, 
            failure.Error)
    }
}
```
