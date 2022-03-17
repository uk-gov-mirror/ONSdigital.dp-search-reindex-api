package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongodb "github.com/ONSdigital/dp-mongodb"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/globalsign/mgo"
)

// JobStore is a type that contains an implementation of the MongoJobStorer interface, which can be used for creating
// and getting Job resources. It also represents a simplistic MongoDB configuration, with session,
// health and lock clients
type JobStore struct {
	Session         *mgo.Session
	URI             string
	Database        string
	JobsCollection  string
	LocksCollection string
	TasksCollection string
	client          *dpMongoHealth.Client
	healthClient    *dpMongoHealth.CheckMongoClient
	lockClient      *dpMongoLock.Lock
	cfg             *config.Config
}

// Init creates a new mgo.Session with a strong consistency and a write mode of "majority".
func (m *JobStore) Init(ctx context.Context, cfg *config.Config) (err error) {
	m.cfg = cfg
	if m.Session != nil {
		return errors.New("session already exists")
	}

	// Create session
	if m.Session, err = mgo.Dial(m.URI); err != nil {
		return err
	}
	m.Session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	m.Session.SetMode(mgo.Strong, true)

	databaseCollectionBuilder := make(map[dpMongoHealth.Database][]dpMongoHealth.Collection)
	databaseCollectionBuilder[dpMongoHealth.Database(m.Database)] = []dpMongoHealth.Collection{dpMongoHealth.Collection(m.JobsCollection),
		dpMongoHealth.Collection(m.LocksCollection), dpMongoHealth.Collection(m.TasksCollection)}
	// Create client and healthClient from session
	m.client = dpMongoHealth.NewClientWithCollections(m.Session, databaseCollectionBuilder)
	m.healthClient = &dpMongoHealth.CheckMongoClient{
		Client:      *m.client,
		Healthcheck: m.client.Healthcheck,
	}

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Session, m.Database, m.JobsCollection)

	return nil
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *JobStore) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// Close closes the mongo session and returns any error
func (m *JobStore) Close(ctx context.Context) error {
	m.lockClient.Close(ctx)
	return dpMongodb.Close(ctx, m.Session)
}
