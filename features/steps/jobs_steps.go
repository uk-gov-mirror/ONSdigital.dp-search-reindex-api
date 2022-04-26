package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

// eachJobShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected values of
// all the remaining attributes of a job.
func (f *SearchReindexAPIFeature) eachJobShouldAlsoContainTheFollowingValues(table *godog.Table) error {
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

// iAmNotIdentifiedByZebedee is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It enables the fake Zebedee service to return a 401 unauthorized error, and the message "user not authenticated", when the /identity endpoint is called.
func (f *SearchReindexAPIFeature) iAmNotIdentifiedByZebedee() error {
	f.AuthFeature.FakeAuthService.NewHandler().Get("/identity").Reply(401).BodyString(`{ "message": "user not authenticated"}`)
	return nil
}

// iCallGETJobsidUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call GET /jobs/{id}.
func (f *SearchReindexAPIFeature) iCallGETJobsidUsingTheGeneratedID() error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return err
	}

	err = f.CallGetJobByID(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsUsingAValidUUID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls GET /jobs/{id} using the id passed in, which should be a valid UUID.
func (f *SearchReindexAPIFeature) iCallGETJobsUsingAValidUUID(id string) error {
	err := f.CallGetJobByID(id)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberTofTasksUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberTofTasksUsingTheGeneratedID(count int) error {
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

// iCallPUTJobsNumberoftasksUsingAValidUUID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It uses the parameters passed in to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *SearchReindexAPIFeature) iCallPUTJobsNumberoftasksUsingAValidUUID(idStr string, count int) error {
	countStr := strconv.Itoa(count)
	f.createdJob.ID = idStr

	err := f.PutNumberOfTasks(countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{invalidCount}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount(invalidCount string) error {
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

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{negativeCount}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount(negativeCount string) error {
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

// iCallPATCHJobsIDUsingTheGeneratedID is a feature step that gets ID from the response body generated in the previous step and then calls PATCH /jobs/{id}
func (f *SearchReindexAPIFeature) iCallPATCHJobsIDUsingTheGeneratedID(patchReqBody *godog.DocString) error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return err
	}

	patchPathWithID := fmt.Sprintf("/jobs/%s", f.createdJob.ID)
	err = f.APIFeature.IPatch(patchPathWithID, patchReqBody)
	if err != nil {
		return fmt.Errorf("failed to send patch request - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iHaveGeneratedJobsInTheJobStore is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls POST /jobs with an empty body which causes job resources to be generated.
func (f *SearchReindexAPIFeature) iHaveGeneratedJobsInTheJobStore(noOfJobs int) error {
	if noOfJobs < 0 {
		return fmt.Errorf("invalid number of jobs given - noOfJobs = %d", noOfJobs)
	}

	if noOfJobs == 0 {
		err := f.Reset(false)
		if err != nil {
			return fmt.Errorf("failed to reset the SearchReindexAPIFeature: %w", err)
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

// iSetIfMatchHeaderToTheGeneratedETag is a feature step that gets the eTag from the response body generated in the previous step
// and then sets If-Match header to that eTag
func (f *SearchReindexAPIFeature) iSetIfMatchHeaderToTheGeneratedETag() error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return fmt.Errorf("failed to read response - err: %w", err)
	}

	err = f.APIFeature.ISetTheHeaderTo("If-Match", f.createdJob.ETag)
	if err != nil {
		return fmt.Errorf("failed to set If-Match header - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iSetIfMatchHeaderToTheOldGeneratedETag is a feature step that gets the eTag from the response body generated in the previous step
// and then sets If-Match header to that eTag and then updates the same job resource again so that the set eTag is now outdated
func (f *SearchReindexAPIFeature) iSetIfMatchHeaderToTheOldGeneratedETag() error {
	err := f.getAndSetCreatedJobFromResponse()
	if err != nil {
		return fmt.Errorf("failed to read response - err: %w", err)
	}

	err = f.APIFeature.ISetTheHeaderTo("If-Match", f.createdJob.ETag)
	if err != nil {
		return fmt.Errorf("failed to set If-Match header - err: %w", err)
	}

	patchPathWithID := fmt.Sprintf("/jobs/%s", f.createdJob.ID)
	patchReqBody := &messages.PickleDocString{
		Content: `[{ "op": "replace", "path": "/state", "value": "created" }]`,
	}

	err = f.APIFeature.IPatch(patchPathWithID, patchReqBody)
	if err != nil {
		return fmt.Errorf("failed to send patch request - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It takes a table that contains the expected structures for job_id, last_updated, and links values. And it asserts whether or not these are found.
func (f *SearchReindexAPIFeature) iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure(table *godog.Table) error {

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

	_, err = f.MongoFeature.GetDocumentWithID("tasks", response.JobID)
	if err != nil {
		return fmt.Errorf("*********error : %w", err)
	}

	// var response2 models.Task
	// err = singleResult.Decode(&response2)
	// if err != nil {
	// 	return fmt.Errorf("error while decoding response for a particular document id from mongo: %w", err)
	// }

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

// iWouldExpectTheResponseToBeAnEmptyList is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that an empty list (0 jobs) has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectTheResponseToBeAnEmptyList() error {
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

// iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList() error {
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

// iWouldExpectThereToBeFourJobsReturnedInAList is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectThereToBeFourJobsReturnedInAList() error {
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

// iWouldExpectThereToBeTasksReturnedInAList is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response to make sure that a list containing the expected number of tasks has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectThereToBeTasksReturnedInAList(numTasksExpected int) error {
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

// inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected types of values of id,
// last_updated, links, and search_index_name.
func (f *SearchReindexAPIFeature) inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures(table *godog.Table) error {
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

// inEachTaskIWouldExpectIdLast_updatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that each task contains the expected types of values of job_id,
// last_updated, and links.
func (f *SearchReindexAPIFeature) inEachTaskIWouldExpectJobIDLastUpdatedAndLinksToHaveThisStructure(table *godog.Table) error {
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

// theJobShouldOnlyBeUpdatedWithTheFollowingFieldsAndValues is a feature step that
func (f *SearchReindexAPIFeature) theJobShouldOnlyBeUpdatedWithTheFollowingFieldsAndValues(table *godog.Table) error {
	err := f.CallGetJobByID(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("failed to get job with ID %s - err: %w", f.createdJob.ID, err)
	}

	updatedJobResponse, err := f.getJobFromResponse()
	if err != nil {
		return fmt.Errorf("failed to get job from response - err: %w", err)
	}

	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	err = f.checkJobUpdates(f.createdJob, *updatedJobResponse, expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check job updates - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /jobs to make sure that the jobs are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *SearchReindexAPIFeature) theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
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

// theSearchReindexAPILosesItsConnectionToMongoDB is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It loses the connection to mongo DB by setting the mongo database to an invalid setting (in the Reset function).
func (f *SearchReindexAPIFeature) theSearchReindexAPILosesItsConnectionToMongoDB() error {
	err := f.Reset(true)
	if err != nil {
		return fmt.Errorf("failed to reset the SearchReindexAPIFeature: %w", err)
	}
	return nil
}

// theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It asserts that the job id and search index name that get returned by the POST /jobs endpoint match the ones that get sent in the
// reindex-requested event
func (f *SearchReindexAPIFeature) theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName() error {
	reindexRequestedData, err := readAndDeserializeKafkaProducerOutput(f.kafkaProducerOutputData)
	if err != nil {
		return err
	}
	assert.Equal(&f.ErrorFeature, f.createdJob.ID, reindexRequestedData.JobID)
	assert.Equal(&f.ErrorFeature, f.createdJob.SearchIndexName, reindexRequestedData.SearchIndex)
	return nil
}

// theResponseETagHeaderShouldBeANewETag is a feature step that checks if the response ETag header returns a new eTag and not the old eTag
func (f *SearchReindexAPIFeature) theResponseETagHeaderShouldBeANewETag() error {
	assert.NotEqual(&f.ErrorFeature, f.createdJob.ETag, f.APIFeature.HttpResponse.Header.Get(dpresponse.ETagHeader))
	return f.ErrorFeature.StepError()
}

// theResponseShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It takes a table that contains the expected values for all the remaining attributes, of a Job resource, and it asserts whether or not these are found.
func (f *SearchReindexAPIFeature) theResponseShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	f.checkValuesInJob(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainAStateOf is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It unmarshalls the response into a Job and gets the state. It checks that the state is the same as the expected one.
func (f *SearchReindexAPIFeature) theResponseShouldContainAStateOf(expectedState string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	assert.Equal(&f.ErrorFeature, expectedState, response.State)

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainTheNewNumberOfTasks is a feature step that can be defined for a specific SearchReindexAPIFeature.
// After PUT /jobs/{id}/number_of_tasks/{number_of_tasks} has been called, followed by GET /jobs/{id},
// this function checks that the job returned contains the correct number_of_tasks value.
func (f *SearchReindexAPIFeature) theResponseShouldContainTheNewNumberOfTasks(table *godog.Table) error {
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

// theResponseShouldContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It takes a table that contains the expected structures for job_id, last_updated, links, and search_index_name values.
// And it asserts whether or not these are found.
func (f *SearchReindexAPIFeature) theResponseShouldContainValuesThatHaveTheseStructures(table *godog.Table) error {
	var err error

	err = f.getAndSetCreatedJobFromResponse()
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

// theSearchReindexAPILosesItsConnectionToTheSearchAPI is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It closes the connection to the search feature so as to mimic losing the connection to the Search API.
func (f *SearchReindexAPIFeature) theSearchReindexAPILosesItsConnectionToTheSearchAPI() error {
	f.SearchFeature.Close()
	return f.ErrorFeature.StepError()
}
