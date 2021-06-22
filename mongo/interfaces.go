package mongo

import (
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/globalsign/mgo"
)

//go:generate moq -out mock/mongo_client.go -pkg mock . MongoClient

// MongoClient is an interface for a type that can connect to MongoDB
type MongoClient interface {
	GetMongoClient(db *mgo.Session, clientDatabaseCollection map[dpMongoHealth.Database][]dpMongoHealth.Collection) *dpMongoHealth.Client
}
