package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	indexer "github.com/khoomi/idxr"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		action      = flag.String("action", "create", "Action: create, drop, list, stats")
		uri         = flag.String("uri", "", "MongoDB URI (defaults to env DATABASE_URL)")
		dbName      = flag.String("db", "", "Database name (defaults to env DB_NAME)")
		collection  = flag.String("collection", "", "Collection name (for list/stats)")
		timeout     = flag.Duration("timeout", 60*time.Second, "Operation timeout")
		continueErr = flag.Bool("continue-on-error", true, "Continue on error")
		skipExists  = flag.Bool("skip-if-exists", true, "Skip existing indexes")
		jsonOutput  = flag.Bool("json", false, "Output in JSON format")
	)
	flag.Parse()

	mongoURI := *uri
	if mongoURI == "" {
		mongoURI = os.Getenv("DATABASE_URL")
		if mongoURI == "" {
			mongoURI = "mongodb://localhost:27017"
		}
	}

	database := *dbName
	if database == "" {
		database = os.Getenv("DB_NAME")
		if database == "" {
			log.Fatal("Database name required (use -db flag or DB_NAME env var)")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect:", err)
		}
	}()

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	db := client.Database(database)

	opts := &indexer.Options{
		Timeout:         *timeout,
		ContinueOnError: *continueErr,
		SkipIfExists:    *skipExists,
	}

	manager := indexer.NewManager(db, opts)

	switch *action {
	case "create":
		if !*jsonOutput {
			fmt.Printf("Creating indexes in database: %s\n", database)
		}

		result, err := manager.Create(context.Background())
		if *jsonOutput {
			outputJSON(map[string]any{
				"success": err == nil,
				"result":  result,
				"error":   errorString(err),
			})
		} else {
			if err != nil {
				log.Printf("Index creation completed with errors: %v", err)
			}
			fmt.Printf("\nResults:\n")
			fmt.Printf("  Success: %d\n", result.SuccessCount)
			fmt.Printf("  Failed: %d\n", result.FailedCount)
			fmt.Printf("  Duration: %v\n", result.Duration)

			if len(result.Failures) > 0 {
				fmt.Printf("\nFailures:\n")
				for _, f := range result.Failures {
					fmt.Printf("  - %s.%s: %v\n", f.Collection, f.IndexName, f.Error)
				}
			}
		}

	case "drop":
		collections := flag.Args()
		if !*jsonOutput {
			fmt.Printf("Dropping indexes in database: %s\n", database)
			if len(collections) > 0 {
				fmt.Printf("Collections: %v\n", collections)
			} else {
				fmt.Println("Collections: all")
			}
		}

		err := manager.Drop(context.Background(), collections...)
		if *jsonOutput {
			outputJSON(map[string]any{
				"success": err == nil,
				"error":   errorString(err),
			})
		} else {
			if err != nil {
				log.Fatal("Failed to drop indexes:", err)
			}
			fmt.Println("Indexes dropped successfully")
		}

	case "list":
		if *collection == "" {
			log.Fatal("Collection name required for list action (-collection flag)")
		}

		indexes, err := manager.List(context.Background(), *collection)
		if err != nil {
			log.Fatal("Failed to list indexes:", err)
		}

		if *jsonOutput {
			outputJSON(indexes)
		} else {
			fmt.Printf("Indexes for collection %s:\n", *collection)
			for _, idx := range indexes {
				if name, ok := idx["name"].(string); ok {
					fmt.Printf("  - %s\n", name)
					if key, ok := idx["key"]; ok {
						fmt.Printf("    Keys: %v\n", key)
					}
					if unique, ok := idx["unique"].(bool); ok && unique {
						fmt.Printf("    Unique: true\n")
					}
				}
			}
		}

	case "stats":
		if *collection == "" {
			stats, err := manager.StatsAll(context.Background())
			if err != nil {
				log.Fatal("Failed to get stats:", err)
			}

			if *jsonOutput {
				outputJSON(stats)
			} else {
				for coll, collStats := range stats {
					fmt.Printf("\n=== %s ===\n", coll)
					for _, stat := range collStats {
						fmt.Printf("  %s:\n", stat.Name)
						fmt.Printf("    Accesses: %d\n", stat.Accesses)
						fmt.Printf("    Since: %v\n", stat.Since)
						if stat.Building {
							fmt.Printf("    Status: BUILDING\n")
						}
					}
				}
			}
		} else {
			stats, err := manager.Stats(context.Background(), *collection)
			if err != nil {
				log.Fatal("Failed to get stats:", err)
			}

			if *jsonOutput {
				outputJSON(stats)
			} else {
				fmt.Printf("Index stats for %s:\n", *collection)
				for _, stat := range stats {
					fmt.Printf("  %s:\n", stat.Name)
					fmt.Printf("    Accesses: %d\n", stat.Accesses)
					fmt.Printf("    Since: %v\n", stat.Since)
					if stat.Building {
						fmt.Printf("    Status: BUILDING\n")
					}
				}
			}
		}

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: create, drop, list, stats")
		os.Exit(1)
	}
}

func outputJSON(data any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatal("Failed to encode JSON:", err)
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
