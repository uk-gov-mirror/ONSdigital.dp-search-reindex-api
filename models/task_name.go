package models

import (
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
)

// TaskName - iota enum of possible task names
type TaskName int

// Possible values for a State of an image. It can only be one of the following:
const (
	TaskDatasetApi TaskName = iota
	TaskZebedee
)

var taskNameValues = []string{"dataset-api", "zebedee"}

// String returns the string representation of a state
func (t TaskName) String() string {
	return taskNameValues[t]
}

// ParseTaskName returns a TaskName from its string representation
func ParseTaskName(taskNameStr string) (TaskName, error) {
	for t, validTaskName := range taskNameValues {
		if taskNameStr == validTaskName {
			return TaskName(t), nil
		}
	}
	return -1, apierrors.ErrTaskInvalidName
}
