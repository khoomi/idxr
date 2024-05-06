package util

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func GetLoginHistorySortBson(sort string) bson.D {
	value := -1
	var key string

	switch sort {
	case "date_asc":
		key = "date"
	case "date_desc":
		key = "date"
	default:
		key = "date"
	}

	if strings.Contains(sort, "asc") {
		value = 1
	}
	return bson.D{{Key: key, Value: value}}
}
