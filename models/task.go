package models

import (
	"errors"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/url"
)

// Task represents a job metadata model and json representation for API
type Task struct {
	JobID             string     `bson:"job_id" json:"job_id"`
	LastUpdated       time.Time  `bson:"last_updated" json:"last_updated"`
	Links             *TaskLinks `bson:"links" json:"links"`
	NumberOfDocuments int        `bson:"number_of_documents" json:"number_of_documents"`
	Task              string     `bson:"task" json:"task"`
}

// TaskLinks is a type that contains links to the endpoints for returning a specific task (self), and the job that it is part of (job), respectively.
type TaskLinks struct {
	Self string `json:"self"`
	Job  string `json:"job"`
}

// NewTask returns a new Task resource that it creates and populates with default values.
func NewTask(jobID string, nameOfApi string, numDocuments int) (Task, error) {
	cfg, err := config.Get()
	if err != nil {
		err = errors.New("unable to retrieve service configuration")
	}
	urlBuilder := url.NewBuilder("http://" + cfg.BindAddr)
	self := urlBuilder.BuildJobTaskURL(jobID, nameOfApi)
	job := urlBuilder.BuildJobURL(jobID)

	return Task{
		JobID:       jobID,
		LastUpdated: time.Now().UTC(),
		Links: &TaskLinks{
			Self: self,
			Job:  job,
		},
		NumberOfDocuments: numDocuments,
		Task:              nameOfApi,
	}, err
}
