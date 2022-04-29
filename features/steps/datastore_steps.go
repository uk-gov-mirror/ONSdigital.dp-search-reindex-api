package steps

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
)

// theTaskShouldHaveTheFollowingFieldsAndValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It takes a table that contains the expected structures and values and compares it to the task resource.
func (f *SearchReindexAPIFeature) theTaskResourceShouldHaveTheFollowingFieldsAndValuesInDatastore(table *godog.Table) error {
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	options := mongo.Options{
		Offset: f.Config.DefaultOffset,
		Limit: 1,
	}
	tasksList, err := f.MongoClient.GetTasks(context.Background(),  options, f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of tasks: %w", err)
	}

	task := tasksList.TaskList[0]

	err = f.checkTaskStructure(task, expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	f.checkValuesInTask(expectedResult, task)

	return f.ErrorFeature.StepError()
}
