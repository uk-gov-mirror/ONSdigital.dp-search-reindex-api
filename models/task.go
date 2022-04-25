package models

import (
	"fmt"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/url"
	"github.com/pkg/errors"
)

// Task represents a job metadata model and json representation for API
type Task struct {
	JobID             string     `bson:"job_id" json:"job_id"`
	LastUpdated       time.Time  `bson:"last_updated" json:"last_updated"`
	Links             *TaskLinks `bson:"links" json:"links"`
	NumberOfDocuments int        `bson:"number_of_documents" json:"number_of_documents"`
	TaskName          string     `bson:"task_name" json:"task_name"`
	ETag              string     `bson:"e_tag" json:"-"`
}

// TaskLinks is a type that contains links to the endpoints for returning a specific task (self), and the job that it is part of (job), respectively.
type TaskLinks struct {
	Self string `json:"self"`
	Job  string `json:"job"`
}

// ParseTaskName returns ErrTaskInvalidName if it cannot find a valid task name matching the given taskName string
func ParseTaskName(taskName string, taskNames map[string]bool) error {
	if isPresent := taskNames[taskName]; !isPresent {
		return apierrors.ErrTaskInvalidName
	}
	return nil
}

// NewTask returns a new Task resource that it creates and populates with default values.
func NewTask(jobID, taskName string, numDocuments int, bindAddress string) (Task, error) {
	urlBuilder := url.NewBuilder("http://" + bindAddress)
	self := urlBuilder.BuildJobTaskURL(jobID, taskName)
	job := urlBuilder.BuildJobURL(jobID)

	newTask := Task{
		JobID:       jobID,
		LastUpdated: time.Now().UTC(),
		Links: &TaskLinks{
			Self: self,
			Job:  job,
		},
		NumberOfDocuments: numDocuments,
		TaskName:          taskName,
	}

	taskETag, err := GenerateETagForTask(newTask)
	if err != nil {
		return Task{}, fmt.Errorf("%s: %w", errors.New("unable to generate eTag for new task"), err)
	}
	newTask.ETag = taskETag

	return newTask, nil
}

// TaskToCreate is a type that contains the details required for creating a Task type.
type TaskToCreate struct {
	TaskName          string `json:"task_name"`
	NumberOfDocuments int    `json:"number_of_documents"`
}

// Validate checks that the TaskToCreate contains a valid TaskName that is not an empty string.
func (task TaskToCreate) Validate(taskNames map[string]bool) error {
	if task.TaskName == "" {
		return apierrors.ErrEmptyTaskNameProvided
	}
	if err := ParseTaskName(task.TaskName, taskNames); err != nil {
		return apierrors.ErrTaskInvalidName
	}
	return nil
}
