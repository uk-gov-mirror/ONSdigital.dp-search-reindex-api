package models

import (
	"time"
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
func NewTask(jobID string, nameOfApi string) Task {

	return Task{
		JobID:       jobID,
		LastUpdated: time.Now().UTC(),
		Links: &TaskLinks{
			Self: "http://localhost:12150/jobs/" + jobID + "/tasks/" + nameOfApi,
			Job:  "http://localhost:12150/jobs/" + jobID,
		},
		NumberOfDocuments: 0,
		Task:              "zebedee",
	}
}
