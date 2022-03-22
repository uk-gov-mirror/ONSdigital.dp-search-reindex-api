package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

// RegisterSteps defines the steps within a specific JobsFeature cucumber test.
func (f *JobsFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a new task resource is created containing the following values:$`, f.aNewTaskResourceIsCreatedContainingTheFollowingValues)
	ctx.Step(`^each job should also contain the following values:$`, f.eachJobShouldAlsoContainTheFollowingValues)
	ctx.Step(`^each task should also contain the following values:$`, f.eachTaskShouldAlsoContainTheFollowingValues)
	ctx.Step(`^I am not identified by zebedee$`, f.iAmNotIdentifiedByZebedee)

	ctx.Step(`^I call GET \/jobs\/{id} using the generated id$`, f.iCallGETJobsidUsingTheGeneratedID)
	ctx.Step(`^I call GET \/jobs\/{"([^"]*)"} using a valid UUID$`, f.iCallGETJobsUsingAValidUUID)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks\/{"([^"]*)"}$`, f.iCallGETJobsidtasks)
	ctx.Step(`^I call GET \/jobs\/{"([^"]*)"}\/tasks\/{"([^"]*)"} using a valid UUID$`, f.iCallGETJobsTasksUsingAValidUUID)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks using the same id again$`, f.iCallGETJobsidtasksUsingTheSameIDAgain)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks\?offset="([^"]*)"&limit="([^"]*)"$`, f.iCallGETJobsidtasksoffsetLimit)
	ctx.Step(`^I GET "\/jobs\/{"([^"]*)"}\/tasks"$`, f.iGETJobsTasks)
	ctx.Step(`^I GET \/jobs\/{id}\/tasks using the generated id$`, f.iGETJobsidtasksUsingTheGeneratedID)

	ctx.Step(`^I call POST \/jobs\/{id}\/tasks to update the number_of_documents for that task$`, f.iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask)
	ctx.Step(`^I call POST \/jobs\/{id}\/tasks using the generated id$`, f.iCallPOSTJobsidtasksUsingTheGeneratedID)
	ctx.Step(`^I call POST \/jobs\/{id}\/tasks using the same id again$`, f.iCallPOSTJobsidtasksUsingTheSameIDAgain)

	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{(\d+)} using the generated id$`, f.iCallPUTJobsidnumberTofTasksUsingTheGeneratedID)
	ctx.Step(`^I call PUT \/jobs\/{"([^"]*)"}\/number_of_tasks\/{(\d+)} using a valid UUID$`, f.iCallPUTJobsNumberoftasksUsingAValidUUID)
	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{"([^"]*)"} using the generated id with an invalid count$`, f.iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount)
	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{"([^"]*)"} using the generated id with a negative count$`, f.iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount)

	ctx.Step(`^I have created a task for the generated job$`, f.iHaveCreatedATaskForTheGeneratedJob)
	ctx.Step(`^I have generated (\d+) jobs in the Job Store$`, f.iHaveGeneratedJobsInTheJobStore)
	ctx.Step(`^I would expect job_id, last_updated, and links to have this structure$`, f.iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure)
	ctx.Step(`^I would expect the response to be an empty list$`, f.iWouldExpectTheResponseToBeAnEmptyList)
	ctx.Step(`^I would expect the response to be an empty list of tasks$`, f.iWouldExpectTheResponseToBeAnEmptyListOfTasks)
	ctx.Step(`^I would expect there to be three or more jobs returned in a list$`, f.iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList)
	ctx.Step(`^I would expect there to be four jobs returned in a list$`, f.iWouldExpectThereToBeFourJobsReturnedInAList)
	ctx.Step(`^I would expect there to be (\d+) tasks returned in a list$`, f.iWouldExpectThereToBeTasksReturnedInAList)
	ctx.Step(`^in each job I would expect the response to contain values that have these structures$`, f.inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures)
	ctx.Step(`^in each task I would expect job_id, last_updated, and links to have this structure$`, f.inEachTaskIWouldExpectJobIDLastUpdatedAndLinksToHaveThisStructure)
	ctx.Step(`^no tasks have been created in the tasks collection$`, f.noTasksHaveBeenCreatedInTheTasksCollection)
	ctx.Step(`^the jobs should be ordered, by last_updated, with the oldest first$`, f.theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst)
	ctx.Step(`^the search reindex api loses its connection to mongo DB$`, f.theSearchReindexAPILosesItsConnectionToMongoDB)
	ctx.Step(`^the task resource should also contain the following values:$`, f.theTaskResourceShouldAlsoContainTheFollowingValues)
	ctx.Step(`^the tasks should be ordered, by last_updated, with the oldest first$`, f.theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst)
	ctx.Step(`^the reindex-requested event should contain the expected job ID and search index name$`, f.theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName)
	ctx.Step(`^the response should also contain the following values:$`, f.theResponseShouldAlsoContainTheFollowingValues)
	ctx.Step(`^the response should contain the new number of tasks$`, f.theResponseShouldContainTheNewNumberOfTasks)
	ctx.Step(`^the response should contain values that have these structures$`, f.theResponseShouldContainValuesThatHaveTheseStructures)
	ctx.Step(`^the response should contain a state of "([^"]*)"$`, f.theResponseShouldContainAStateOf)
	ctx.Step(`^the search reindex api loses its connection to the search api$`, f.theSearchReindexAPILosesItsConnectionToTheSearchAPI)
}

// iAmNotIdentifiedByZebedee is a feature step that can be defined for a specific JobsFeature.
// It enables the fake Zebedee service to return a 401 unauthorized error, and the message "user not authenticated", when the /identity endpoint is called.
func (f *JobsFeature) iAmNotIdentifiedByZebedee() error {
	f.AuthFeature.FakeAuthService.NewHandler().Get("/identity").Reply(401).BodyString(`{ "message": "user not authenticated"}`)
	return nil
}

// iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected structures for job_id, last_updated, and links values. And it asserts whether or not these are found.
func (f *JobsFeature) iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure(table *godog.Table) error {
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

	jobID := response.JobID
	lastUpdated := response.LastUpdated
	links := response.Links
	taskName := response.TaskName

	err = f.checkTaskStructure(jobID, lastUpdated, expectedResult, links, taskName)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected structures for job_id, last_updated, links, and search_index_name values.
// And it asserts whether or not these are found.
func (f *JobsFeature) theResponseShouldContainValuesThatHaveTheseStructures(table *godog.Table) error {
	var err error

	err = f.readResponse()
	if err != nil {
		return err
	}

	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	err = f.checkStructure(expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// aNewTaskResourceIsCreatedContainingTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It checks that a task has been created containing the expected values of number_of_documents and task_name that are passed in via the table.
func (f *JobsFeature) aNewTaskResourceIsCreatedContainingTheFollowingValues(table *godog.Table) error {
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

// theResponseShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected values for all the remaining attributes, of a Job resource, and it asserts whether or not these are found.
func (f *JobsFeature) theResponseShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	f.checkValuesInJob(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// theTaskResourceShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected values for all the remaining attributes, of a TaskName resource, and it asserts whether or not these are found.
func (f *JobsFeature) theTaskResourceShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Task

	_ = json.Unmarshal(f.responseBody, &response)

	f.checkValuesInTask(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// eachTaskShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It gets the list of tasks from the response and checks that each task contains the expected number of documents and a valid task name.
// NB. The valid task names are listed in the taskNames variable.
func (f *JobsFeature) eachTaskShouldAlsoContainTheFollowingValues(table *godog.Table) error {
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

// theResponseShouldContainAStateOf is a feature step that can be defined for a specific JobsFeature.
// It unmarshalls the response into a Job and gets the state. It checks that the state is the same as the expected one.
func (f *JobsFeature) theResponseShouldContainAStateOf(expectedState string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	assert.Equal(&f.ErrorFeature, expectedState, response.State)

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call GET /jobs/{id}.
func (f *JobsFeature) iCallGETJobsidUsingTheGeneratedID() error {
	err := f.readResponse()
	if err != nil {
		return err
	}

	err = f.GetJobByID(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the generated job id, and passes it the request body.
func (f *JobsFeature) iCallPOSTJobsidtasksUsingTheGeneratedID(body *godog.DocString) error {
	var response models.Job

	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the existing job id, and passes it the request body.
func (f *JobsFeature) iCallPOSTJobsidtasksUsingTheSameIDAgain(body *godog.DocString) error {
	err := f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}
	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasks is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks/{task name} via GetTaskForJob, using the generated job id, and passes it the task name.
func (f *JobsFeature) iCallGETJobsidtasks(taskName string) error {
	err := f.readResponse()
	if err != nil {
		return err
	}
	err = f.GetTaskForJob(f.createdJob.ID, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via PostTaskForJob using the generated job id
func (f *JobsFeature) iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask(body *godog.DocString) error {
	err := f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iHaveGeneratedJobsInTheJobStore is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs with an empty body which causes job resources to be generated.
func (f *JobsFeature) iHaveGeneratedJobsInTheJobStore(noOfJobs int) error {
	if noOfJobs < 0 {
		return fmt.Errorf("invalid number of jobs given - noOfJobs = %d", noOfJobs)
	}

	if noOfJobs == 0 {
		err := f.Reset(false)
		if err != nil {
			return fmt.Errorf("failed to reset the JobsFeature: %w", err)
		}
	}

	for i := 1; i < noOfJobs+1; i++ {
		// call POST /jobs
		err := f.callPostJobs()
		if err != nil {
			return fmt.Errorf("error occurred in callPostJobs at iteration %d: %w", i, err)
		}

		if i < noOfJobs {
			time.Sleep(5 * time.Millisecond)
		}
	}

	return f.ErrorFeature.StepError()
}

// iHaveCreatedATaskForTheGeneratedJob is a feature step that can be defined for a specific JobsFeature.
// It gets the job id from the response to calling POST /jobs and uses it to call POST /jobs/{job id}/tasks/{task name}
// in order to create a task for that job. It passes the taskToCreate request body to the POST endpoint.
func (f *JobsFeature) iHaveCreatedATaskForTheGeneratedJob(taskToCreate *godog.DocString) error {
	err := f.readResponse()
	if err != nil {
		return err
	}

	err = f.APIFeature.IPostToWithBody("/jobs/"+f.createdJob.ID+"/tasks", taskToCreate)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *JobsFeature) iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound >= 3, "The list should contain three or more jobs but it only contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeTasksReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response to make sure that a list containing the expected number of tasks has been returned.
func (f *JobsFeature) iWouldExpectThereToBeTasksReturnedInAList(numTasksExpected int) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numTasksFound := len(response.TaskList)
	assert.True(&f.ErrorFeature, numTasksFound == numTasksExpected, "The list should contain "+strconv.Itoa(numTasksExpected)+
		" tasks but it contains "+strconv.Itoa(numTasksFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeFourJobsReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *JobsFeature) iWouldExpectThereToBeFourJobsReturnedInAList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound == 4, "The list should contain four jobs but it contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected types of values of id,
// last_updated, links, and search_index_name.
func (f *JobsFeature) inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures(table *godog.Table) error {
	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Jobs

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for j := range response.JobList {
		f.createdJob = response.JobList[j]
		err := f.checkStructure(expectedResult)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}
	return f.ErrorFeature.StepError()
}

// inEachTaskIWouldExpectIdLast_updatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that each task contains the expected types of values of job_id,
// last_updated, and links.
func (f *JobsFeature) inEachTaskIWouldExpectJobIDLastUpdatedAndLinksToHaveThisStructure(table *godog.Table) error {
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

	for j := range response.TaskList {
		task := response.TaskList[j]
		err := f.checkTaskStructure(task.JobID, task.LastUpdated, expectedResult, task.Links, task.TaskName)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}

	return f.ErrorFeature.StepError()
}

// eachJobShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected values of
// all the remaining attributes of a job.
func (f *JobsFeature) eachJobShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Jobs

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for i := 0; i < len(response.JobList); i++ {
		f.checkValuesInJob(expectedResult, response.JobList[i])
	}

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainTheNewNumberOfTasks is a feature step that can be defined for a specific JobsFeature.
// After PUT /jobs/{id}/number_of_tasks/{number_of_tasks} has been called, followed by GET /jobs/{id},
// this function checks that the job returned contains the correct number_of_tasks value.
func (f *JobsFeature) theResponseShouldContainTheNewNumberOfTasks(table *godog.Table) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(response.NumberOfTasks))

	return f.ErrorFeature.StepError()
}

// theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that the jobs are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *JobsFeature) theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	jobList := response.JobList
	timeToCheck := jobList[0].LastUpdated

	for j := 1; j < len(jobList); j++ {
		index := strconv.Itoa(j - 1)
		nextIndex := strconv.Itoa(j)
		nextTime := jobList[j].LastUpdated
		assert.True(&f.ErrorFeature, timeToCheck.Before(nextTime),
			"The value of last_updated at job_list["+index+"] should be earlier than that at job_list["+nextIndex+"]")
		timeToCheck = nextTime
	}
	return f.ErrorFeature.StepError()
}

// theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that the tasks are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *JobsFeature) theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
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

// noTasksHaveBeenCreatedInTheTasksCollection is a feature step that can be defined for a specific JobsFeature.
// It resets the tasks collection to its default value, which means that it will contain no tasks.
func (f *JobsFeature) noTasksHaveBeenCreatedInTheTasksCollection() error {
	err := f.Reset(false)
	if err != nil {
		return fmt.Errorf("failed to reset the JobsFeature: %w", err)
	}
	return nil
}

// iCallGETJobsUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id} using the id passed in, which should be a valid UUID.
func (f *JobsFeature) iCallGETJobsUsingAValidUUID(id string) error {
	err := f.GetJobByID(id)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsTasksUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks/{task_name} using the id and taskName passed in, which should both be valid.
func (f *JobsFeature) iCallGETJobsTasksUsingAValidUUID(id, taskName string) error {
	err := f.GetTaskForJob(id, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in GetTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that an empty list (0 jobs) has been returned.
func (f *JobsFeature) iWouldExpectTheResponseToBeAnEmptyList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound == 0, "The list should contain no jobs but it contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyListOfTasks is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/jobID/tasks to make sure that an empty list (0 tasks) has been returned.
func (f *JobsFeature) iWouldExpectTheResponseToBeAnEmptyListOfTasks() error {
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

// iCallPUTJobsidnumberTofTasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *JobsFeature) iCallPUTJobsidnumberTofTasksUsingTheGeneratedID(count int) error {
	countStr := strconv.Itoa(count)
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID
	err = f.PutNumberOfTasks(countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsNumberoftasksUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It uses the parameters passed in to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *JobsFeature) iCallPUTJobsNumberoftasksUsingAValidUUID(idStr string, count int) error {
	countStr := strconv.Itoa(count)
	f.createdJob.ID = idStr

	err := f.PutNumberOfTasks(countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{invalidCount}
func (f *JobsFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount(invalidCount string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PutNumberOfTasks(invalidCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{negativeCount}
func (f *JobsFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount(negativeCount string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PutNumberOfTasks(negativeCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theSearchReindexAPILosesItsConnectionToMongoDB is a feature step that can be defined for a specific JobsFeature.
// It loses the connection to mongo DB by setting the mongo database to an invalid setting (in the Reset function).
func (f *JobsFeature) theSearchReindexAPILosesItsConnectionToMongoDB() error {
	err := f.Reset(true)
	if err != nil {
		return fmt.Errorf("failed to reset the JobsFeature: %w", err)
	}
	return nil
}

// iCallGETJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{id}/tasks using the existing value of id.
func (f *JobsFeature) iCallGETJobsidtasksUsingTheSameIDAgain() error {
	// call GET /jobs/{id}/tasks
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasksoffsetLimit is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks?offset={offset}&limit={limit} using the existing value of id.
func (f *JobsFeature) iCallGETJobsidtasksoffsetLimit(offset, limit string) error {
	// call GET /jobs/{id}/tasks?offset={offset}&limit={limit}
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks?offset=" + offset + "&limit=" + limit)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iGETJobsTasks is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{jobID}/tasks using the existing value of id as the jobID value.
func (f *JobsFeature) iGETJobsTasks() error {
	// call GET /jobs/{jobID}/tasks
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}
	return f.ErrorFeature.StepError()
}

// iGETJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{jobID}/tasks using the response.ID, from the previously returned Job, as the id value.
func (f *JobsFeature) iGETJobsidtasksUsingTheGeneratedID() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID
	err = f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theSearchReindexAPILosesItsConnectionToTheSearchAPI is a feature step that can be defined for a specific JobsFeature.
// It closes the connection to the search feature so as to mimic losing the connection to the Search API.
func (f *JobsFeature) theSearchReindexAPILosesItsConnectionToTheSearchAPI() error {
	f.SearchFeature.Close()
	return f.ErrorFeature.StepError()
}

// theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName is a feature step that can be defined for a specific JobsFeature.
// It asserts that the job id and search index name that get returned by the POST /jobs endpoint match the ones that get sent in the
// reindex-requested event
func (f *JobsFeature) theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName() error {
	reindexRequestedData, err := readAndDeserializeKafkaProducerOutput(f.kafkaProducerOutputData)
	if err != nil {
		return err
	}
	assert.Equal(&f.ErrorFeature, f.createdJob.ID, reindexRequestedData.JobID)
	assert.Equal(&f.ErrorFeature, f.createdJob.SearchIndexName, reindexRequestedData.SearchIndex)
	return nil
}
