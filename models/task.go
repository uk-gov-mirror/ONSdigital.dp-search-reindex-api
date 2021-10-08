package models

import (
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"strings"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/url"
)

// Task represents a job metadata model and json representation for API
type Task struct {
	JobID             string     `bson:"job_id" json:"job_id"`
	LastUpdated       time.Time  `bson:"last_updated" json:"last_updated"`
	Links             *TaskLinks `bson:"links" json:"links"`
	NumberOfDocuments int        `bson:"number_of_documents" json:"number_of_documents"`
	TaskName          string     `bson:"task_name" json:"task_name"`
}

// TaskLinks is a type that contains links to the endpoints for returning a specific task (self), and the job that it is part of (job), respectively.
type TaskLinks struct {
	Self string `json:"self"`
	Job  string `json:"job"`
}

// TaskName - iota enum of possible task names
type TaskName int

// ParseTaskName returns a TaskName from its string representation
func ParseTaskName(taskNameStr string) (TaskName, error) {
	cfg, err := config.Get()
	if err != nil {
		return -1, fmt.Errorf("%s: %w", errors.New("unable to retrieve service configuration"), err)
	}
	taskNameValues := strings.Split(cfg.TaskNameValues, ",")
	for t, validTaskName := range taskNameValues {
		if taskNameStr == validTaskName {
			return TaskName(t), nil
		}
	}
	return -1, apierrors.ErrTaskInvalidName
}

// NewTask returns a new Task resource that it creates and populates with default values.
func NewTask(jobID string, taskName string, numDocuments int, bindAddress string) Task {
	urlBuilder := url.NewBuilder("http://" + bindAddress)
	self := urlBuilder.BuildJobTaskURL(jobID, taskName)
	job := urlBuilder.BuildJobURL(jobID)

	return Task{
		JobID:       jobID,
		LastUpdated: time.Now().UTC(),
		Links: &TaskLinks{
			Self: self,
			Job:  job,
		},
		NumberOfDocuments: numDocuments,
		TaskName:          taskName,
	}
}

// TaskToCreate is a type that contains the details required for creating a Task type.
type TaskToCreate struct {
	TaskName          string `json:"task_name"`
	NumberOfDocuments int    `json:"number_of_documents"`
}

// Validate checks that the TaskToCreate contains a valid TaskName that is not an empty string.
func (task TaskToCreate) Validate() error {
	if task.TaskName == "" {
		return apierrors.ErrEmptyTaskNameProvided
	}
	if _, err := ParseTaskName(task.TaskName); err != nil {
		return apierrors.ErrTaskInvalidName
	}
	return nil
}
