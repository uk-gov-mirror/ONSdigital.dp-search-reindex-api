package models

import (
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/url"
	"github.com/pkg/errors"
)

// Job represents a job metadata model and json representation for API
type Job struct {
	ID                           string    `bson:"_id" json:"id"`
	LastUpdated                  time.Time `bson:"last_updated" json:"last_updated"`
	Links                        *JobLinks `bson:"links" json:"links"`
	NumberOfTasks                int       `bson:"number_of_tasks" json:"number_of_tasks"`
	ReindexCompleted             time.Time `bson:"reindex_completed" json:"reindex_completed"`
	ReindexFailed                time.Time `bson:"reindex_failed" json:"reindex_failed"`
	ReindexStarted               time.Time `bson:"reindex_started" json:"reindex_started"`
	SearchIndexName              string    `bson:"search_index_name" json:"search_index_name"`
	State                        string    `bson:"state" json:"state"`
	TotalSearchDocuments         int       `bson:"total_search_documents" json:"total_search_documents"`
	TotalInsertedSearchDocuments int       `bson:"total_inserted_search_documents" json:"total_inserted_search_documents"`
}

// JobLinks is a type that contains links to the endpoints for returning a specific job (self), and the tasks that it contains (tasks), respectively.
type JobLinks struct {
	Tasks string `json:"tasks"`
	Self  string `json:"self"`
}

// NewJob returns a new Job resource that it creates and populates with default values.
func NewJob(id string) (Job, error) {
	zeroTime := time.Time{}.UTC()
	cfg, err := config.Get()
	if err != nil {
		err = errors.New("unable to retrieve service configuration")
	}
	urlBuilder := url.NewBuilder("http://" + cfg.BindAddr)
	self := urlBuilder.BuildJobURL(id)
	tasks := urlBuilder.BuildJobTasksURL(id)

	return Job{
		ID:          id,
		LastUpdated: time.Now().UTC(),
		Links: &JobLinks{
			Tasks: tasks,
			Self:  self,
		},
		NumberOfTasks:                0,
		ReindexCompleted:             zeroTime,
		ReindexFailed:                zeroTime,
		ReindexStarted:               zeroTime,
		SearchIndexName:              "Default Search Index Name",
		State:                        "created",
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}, err
}
