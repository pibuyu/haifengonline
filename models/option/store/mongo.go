package store

import "go.mongodb.org/mongo-driver/mongo"

const (
	defaultCollection = "dirties"
)

type MongoConfig struct {
	Address    string
	Port       string
	Username   string
	Password   string
	Database   string
	Collection string
	FieldName  string
}

type doc struct {
	Id   string `bson:"_id"`
	Word string `bson:"word"`
}

type MongoModel struct {
	store     *mongo.Collection
	fieldName string

	addChan chan string
	delChan chan string
}
