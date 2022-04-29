package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

// aNewTaskResourceIsCreatedContainingTheFollowingValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks that a task has been created containing the expected values of number_of_documents and task_name that are passed in via the table.
func (f *SearchReindexAPIFeature) aNewTaskResourceIsCreatedContainingTheFollowingValues(table *godog.Table) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	var response models.Task

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.checkValuesInTask(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// eachTaskShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the list of tasks from the response and checks that each task contains the expected number of documents and a valid task name.
// NB. The valid task names are listed in the taskNames variable.
func (f *SearchReindexAPIFeature) eachTaskShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Tasks

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for _, task := range response.TaskList {
		assert.Equal(&f.ErrorFeature, expectedResult["number_of_documents"], strconv.Itoa(task.NumberOfDocuments))
		assert.True(&f.ErrorFeature, taskNames[task.TaskName])
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasks is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls GET /jobs/{id}/tasks/{task name} via GetTaskForJob, using the generated job id, and passes it the task name.
func (f *SearchReindexAPIFeature) iCallGETJobsidtasks(taskName string) error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return err
	}
	err = f.GetTaskForJob(f.apiVersion, f.createdJob.ID, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsTasksUsingAValidUUID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls GET /jobs/{id}/tasks/{task_name} using the id and taskName passed in, which should both be valid.
func (f *SearchReindexAPIFeature) iCallGETJobsTasksUsingAValidUUID(id, taskName string) error {
	err := f.GetTaskForJob(f.apiVersion, id, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in GetTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls /jobs/{id}/tasks using the existing value of id.
func (f *SearchReindexAPIFeature) iCallGETJobsidtasksUsingTheSameIDAgain() error {
	// call GET /jobs/{id}/tasks
	path := getPath(f.apiVersion, fmt.Sprintf("/jobs/%s/tasks", f.createdJob.ID))

	err := f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasksoffsetLimit is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls GET /jobs/{id}/tasks?offset={offset}&limit={limit} using the existing value of id.
func (f *SearchReindexAPIFeature) iCallGETJobsidtasksoffsetLimit(offset, limit string) error {
	// call GET /jobs/{id}/tasks?offset={offset}&limit={limit}
	path := getPath(f.apiVersion, fmt.Sprintf("/jobs/%s/tasks?offset=%s&limit=%s", f.createdJob.ID, offset, limit))

	err := f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iGETJobsTasks is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls /jobs/{jobID}/tasks using the existing value of id as the jobID value.
func (f *SearchReindexAPIFeature) iGETJobsTasks() error {
	// call GET /jobs/{jobID}/tasks
	path := getPath(f.apiVersion, fmt.Sprintf("/jobs/%s/tasks", f.createdJob.ID))

	err := f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iGETJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls /jobs/{jobID}/tasks using the response.ID, from the previously returned Job, as the id value.
func (f *SearchReindexAPIFeature) iGETJobsidtasksUsingTheGeneratedID() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	path := getPath(f.apiVersion, fmt.Sprintf("/jobs/%s/tasks", f.createdJob.ID))

	err = f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls POST /jobs/{id}/tasks via PostTaskForJob using the generated job id
func (f *SearchReindexAPIFeature) iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask(body *godog.DocString) error {
	err := f.PostTaskForJob(f.apiVersion, f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the generated job id, and passes it the request body.
func (f *SearchReindexAPIFeature) iCallPOSTJobsidtasksUsingTheGeneratedID(body *godog.DocString) error {
	var response models.Job

	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PostTaskForJob(f.apiVersion, f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the existing job id, and passes it the request body.
func (f *SearchReindexAPIFeature) iCallPOSTJobsidtasksUsingTheSameIDAgain(body *godog.DocString) error {
	err := f.PostTaskForJob(f.apiVersion, f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}
	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iHaveCreatedATaskForTheGeneratedJob is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the job id from the response to calling POST /jobs and uses it to call POST /jobs/{job id}/tasks/{task name}
// in order to create a task for that job. It passes the taskToCreate request body to the POST endpoint.
func (f *SearchReindexAPIFeature) iHaveCreatedATaskForTheGeneratedJob(taskToCreate *godog.DocString) error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return err
	}

	path := getPath(f.apiVersion, fmt.Sprintf("/jobs/%s/tasks", f.createdJob.ID))
	err = f.APIFeature.IPostToWithBody(path, taskToCreate)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// inEachTaskIWouldExpectIdLast_updatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that each task contains the expected types of values of job_id,
// last_updated, and links.
func (f *SearchReindexAPIFeature) expectTaskToLookLikeThis(table *godog.Table) error {
	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	var response models.Tasks

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for i := range response.TaskList {
		err := f.checkTaskStructure(response.TaskList[i], expectedResult)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyListOfTasks is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs/jobID/tasks to make sure that an empty list (0 tasks) has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectTheResponseToBeAnEmptyListOfTasks() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numTasksFound := len(response.TaskList)
	assert.True(&f.ErrorFeature, numTasksFound == 0, "The list should contain no tasks but it contains "+strconv.Itoa(numTasksFound))

	return f.ErrorFeature.StepError()
}

// noTasksHaveBeenCreatedInTheTasksCollection is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It resets the tasks collection to its default value, which means that it will contain no tasks.
func (f *SearchReindexAPIFeature) noTasksHaveBeenCreatedInTheTasksCollection() error {
	err := f.Reset(false)
	if err != nil {
		return fmt.Errorf("failed to reset the SearchReindexAPIFeature: %w", err)
	}
	return nil
}

// theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that the tasks are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *SearchReindexAPIFeature) theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	taskList := response.TaskList
	timeToCheck := taskList[0].LastUpdated

	for j := 1; j < len(taskList); j++ {
		index := strconv.Itoa(j - 1)
		nextIndex := strconv.Itoa(j)
		nextTime := taskList[j].LastUpdated
		assert.True(&f.ErrorFeature, timeToCheck.Before(nextTime),
			"The value of last_updated at taskList["+index+"] should be earlier than that at taskList["+nextIndex+"]")
		timeToCheck = nextTime
	}
	return f.ErrorFeature.StepError()
}

// theTaskShouldHaveTheFollowingFieldsAndValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It takes a table that contains the expected structures and values and compares it to the task resource.
func (f *SearchReindexAPIFeature) theTaskShouldHaveTheFollowingFieldsAndValues(table *godog.Table) error {
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	options := mongo.Options{
		Offset: f.Config.DefaultOffset,
		Limit: 1,
	}
	tasksList, err := f.MongoClient.GetTasks(context.Background(), options, f.createdJob.ID)
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
