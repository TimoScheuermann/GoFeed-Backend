package database

import "go.mongodb.org/mongo-driver/mongo"

type DBDriver struct {
	Collection mongo.Collection
}
