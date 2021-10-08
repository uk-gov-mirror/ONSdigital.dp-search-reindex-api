package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
)

// TaskName - iota enum of possible task names
type TaskName int

// ParseTaskName returns a TaskName from its string representation
func ParseTaskName(taskNameStr string) (TaskName, error) {
	cfg, err := config.Get()
	if err != nil {
		return -1, fmt.Errorf("%s: %w", errors.New("unable to retrieve service configuration"), err)
	}
	allNames := cfg.TaskNameValues
	taskNameValues := strings.Split(allNames, ",")
	for t, validTaskName := range taskNameValues {
		if taskNameStr == validTaskName {
			return TaskName(t), nil
		}
	}
	return -1, apierrors.ErrTaskInvalidName
}
