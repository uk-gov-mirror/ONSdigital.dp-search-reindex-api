package models

import (
	"fmt"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/url"
	"github.com/pkg/errors"
)

// Paths of fields in a job resource
const (
	JobNoOfTasksPath            = "/number_of_tasks"
	JobStatePath                = "/state"
	JobTotalSearchDocumentsPath = "/total_search_documents"
)

// BSON keys for each field in the job resource so that there is a source of truth
const (
	JobETagBSONKey                 = "e_tag"
	JobLastUpdatedBSONKey          = "last_updated"
	JobReindexCompletedBSONKey     = "reindex_completed"
	JobReindexFailedBSONKey        = "reindex_failed"
	JobNoOfTasksBSONKey            = "number_of_tasks"
	JobStateBSONKey                = "state"
	JobTotalSearchDocumentsBSONKey = "total_search_documents"
)

// Job represents a job metadata model and json representation for API
type Job struct {
	ETag                         string    `bson:"e_tag"                            json:"e_tag"`
	ID                           string    `bson:"_id"                              json:"id"`
	LastUpdated                  time.Time `bson:"last_updated"                     json:"last_updated"`
	Links                        *JobLinks `bson:"links"                            json:"links"`
	NumberOfTasks                int       `bson:"number_of_tasks"                  json:"number_of_tasks"`
	ReindexCompleted             time.Time `bson:"reindex_completed"                json:"reindex_completed"`
	ReindexFailed                time.Time `bson:"reindex_failed"                   json:"reindex_failed"`
	ReindexStarted               time.Time `bson:"reindex_started"                  json:"reindex_started"`
	SearchIndexName              string    `bson:"search_index_name"                json:"search_index_name"`
	State                        string    `bson:"state"                            json:"state"`
	TotalSearchDocuments         int       `bson:"total_search_documents"           json:"total_search_documents"`
	TotalInsertedSearchDocuments int       `bson:"total_inserted_search_documents"  json:"total_inserted_search_documents"`
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
		return Job{}, fmt.Errorf("%s: %w", errors.New("unable to retrieve service configuration"), err)
	}
	urlBuilder := url.NewBuilder("http://" + cfg.BindAddr)
	self := urlBuilder.BuildJobURL(id)
	tasks := urlBuilder.BuildJobTasksURL(id)

	newJob := Job{
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
		State:                        JobStateCreated,
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}

	jobETag, err := GenerateETagForJob(newJob)
	if err != nil {
		return Job{}, fmt.Errorf("%s: %w", errors.New("unable to generate eTag for new job"), err)
	}
	newJob.ETag = jobETag

	return newJob, nil
}
