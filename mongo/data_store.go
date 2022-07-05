package mongo

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	//dpMongodb "github.com/ONSdigital/dp-mongodb"
	mongolock "github.com/ONSdigital/dp-mongodb/v3/dplock"
	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	//"github.com/globalsign/mgo"
)

// JobStore is a type that contains an implementation of the MongoJobStorer interface, which can be used for creating
// and getting Job resources. It also represents a simplistic MongoDB configuration, with session,
// health and lock clients
type JobStore struct {
	//Session         *mgo.Session
	//URI             string
	//Database        string
	//JobsCollection  string
	//LocksCollection string
	//TasksCollection string
	//client          *dpMongoHealth.Client
	//healthClient    *dpMongoHealth.CheckMongoClient
	//lockClient      *dpMongoLock.Lock
	cfg *config.Config
	//mongodriver.MongoDriverConfig
	config.MongoConfig

	Connection   *mongodriver.MongoConnection
	HealthClient *mongohealth.CheckMongoClient
	LockClient   *mongolock.Lock
}

// Options contains information for pagination which includes offset and limit
type Options struct {
	Offset int
	Limit  int
}

// Init returns an initialised Mongo object encapsulating a connection to the mongo server/cluster with the given configuration,
// a health client to check the health of the mongo server/cluster, and a lock client
func (m *JobStore) Init(ctx context.Context) (err error) {
	m.Connection, err = mongodriver.Open(&m.MongoDriverConfig)
	if err != nil {
		return err
	}

	databaseCollectionBuilder := map[mongohealth.Database][]mongohealth.Collection{
		(mongohealth.Database)(m.Database): {
			mongohealth.Collection(m.ActualCollectionName(config.JobsCollection)),
			mongohealth.Collection(m.ActualCollectionName(config.TasksCollection)),
		},
	}
	m.HealthClient = mongohealth.NewClientWithCollections(m.Connection, databaseCollectionBuilder)
	m.LockClient = mongolock.New(ctx, m.Connection, m.ActualCollectionName(config.JobsCollection))

	return nil
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *JobStore) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.HealthClient.Checker(ctx, state)
}

// Close disconnects the mongo session
func (m *JobStore) Close(ctx context.Context) error {
	m.LockClient.Close(ctx)
	return m.Connection.Close(ctx)
}
