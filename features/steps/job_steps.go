package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// setAPIVersionForPath is a feature step that sets the API version future steps will use when calling the API
func (f *SearchReindexAPIFeature) setAPIVersionForPath(apiVersion string) {
	f.apiVersion = apiVersion

	if apiVersion == "undefined" {
		f.apiVersion = ""
	}
}

// anExistingReindexJobIsInProgress is a feature step that generates a job where its state is in-progress
func (f *SearchReindexAPIFeature) anExistingReindexJobIsInProgress() error {
	err := f.theNoOfExistingJobsInTheJobStore(1)
	if err != nil {
		return fmt.Errorf("failed to generate an existing reindex job - error: %w", err)
	}

	currentTime := time.Now().UTC()

	updates := make(bson.M)
	updates[models.JobStateKey] = models.JobStateInProgress
	updates[models.JobReindexStartedKey] = currentTime
	updates[models.JobLastUpdatedKey] = currentTime

	err = f.MongoClient.UpdateJob(context.Background(), f.createdJob.ID, updates)
	if err != nil {
		return fmt.Errorf("failed to update state to in-progress of generated job: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// eachJobShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /search-reindex-jobs to make sure that each job contains the expected values of
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
// It gets the id from the response body, generated in the previous step, and then uses this to call GET /search-reindex-jobs/{id}.
func (f *SearchReindexAPIFeature) iCallGETJobsidUsingTheGeneratedID() error {
	err := f.CallGetJobByID(f.apiVersion, f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsUsingAValidUUID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It calls GET /search-reindex-jobs/{id} using the id passed in, which should be a valid UUID.
func (f *SearchReindexAPIFeature) iCallGETJobsUsingAValidUUID(id string) error {
	err := f.CallGetJobByID(f.apiVersion, id)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberTofTasksUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /search-reindex-jobs/{id}/number-of-tasks/{count}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberTofTasksUsingTheGeneratedID(count int) error {
	countStr := strconv.Itoa(count)

	err := f.PutNumberOfTasks(f.apiVersion, f.createdJob.ID, countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberTofTasksUsingTheGeneratedID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /search-reindex-jobs/{job_id}/tasks/{task_name}/number-of-documents/{count}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberTofDocsUsingTheGeneratedID(count int) error {
	countStr := strconv.Itoa(count)

	err := f.PutNumberOfDocs(f.apiVersion, f.createdJob.ID, f.createdTask.TaskName, countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfDocs: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsNumberoftasksUsingAValidUUID is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It uses the parameters passed in to call PUT /search-reindex-jobs/{id}/number-of-tasks/{count}
func (f *SearchReindexAPIFeature) iCallPUTJobsNumberoftasksUsingAValidUUID(id string, count int) error {
	countStr := strconv.Itoa(count)

	err := f.PutNumberOfTasks(f.apiVersion, id, countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /search-reindex-jobs/{id}/number-of-tasks/{invalidCount}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount(invalidCount string) error {
	err := f.PutNumberOfTasks(f.apiVersion, f.createdJob.ID, invalidCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /search-reindex-jobs/{id}/number-of-tasks/{negativeCount}
func (f *SearchReindexAPIFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount(negativeCount string) error {
	err := f.PutNumberOfTasks(f.apiVersion, f.createdJob.ID, negativeCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPATCHJobsIDUsingTheGeneratedID is a feature step that gets ID from the response body generated in the previous step and then calls PATCH /search-reindex-jobs/{id}
func (f *SearchReindexAPIFeature) iCallPATCHJobsIDUsingTheGeneratedID(patchReqBody *godog.DocString) error {
	path := getPath(f.apiVersion, fmt.Sprintf("/search-reindex-jobs/%s", f.createdJob.ID))

	err := f.APIFeature.IPatch(path, patchReqBody)
	if err != nil {
		return fmt.Errorf("failed to send patch request - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iSetIfMatchHeaderToValidETagForJobs gets the etag of the jobs resource which contains all the jobs
// and then sets If-Match header to that eTag
func (f *SearchReindexAPIFeature) iSetIfMatchHeaderToValidETagForJobs() error {
	ctx := context.Background()

	option := mongo.Options{
		Offset: f.Config.DefaultOffset,
		Limit:  f.Config.DefaultLimit,
	}

	jobs, err := f.MongoClient.GetJobs(ctx, option)
	if err != nil {
		return fmt.Errorf("failed to get jobs - err: %w", err)
	}

	jobsETag, err := models.GenerateETagForJobs(ctx, *jobs)
	if err != nil {
		return fmt.Errorf("failed to generate etag for jobs - err: %w", err)
	}

	err = f.APIFeature.ISetTheHeaderTo("If-Match", jobsETag)
	if err != nil {
		return fmt.Errorf("failed to set If-Match header - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iSetIfMatchHeaderToTheGeneratedJobETag is a feature step that gets the eTag of the job from the response body generated in the previous step
// and then sets If-Match header to that eTag
func (f *SearchReindexAPIFeature) iSetIfMatchHeaderToTheGeneratedJobETag() error {
	err := f.APIFeature.ISetTheHeaderTo("If-Match", f.createdJob.ETag)
	if err != nil {
		return fmt.Errorf("failed to set If-Match header - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iSetIfMatchHeaderToTheOldGeneratedETag is a feature step that gets the eTag from the response body generated in the previous step
// and then sets If-Match header to that eTag and then updates the same job resource again so that the set eTag is now outdated
func (f *SearchReindexAPIFeature) iSetIfMatchHeaderToTheOldGeneratedETag() error {
	err := f.APIFeature.ISetTheHeaderTo("If-Match", f.createdJob.ETag)
	if err != nil {
		return fmt.Errorf("failed to set If-Match header - err: %w", err)
	}

	path := getPath(f.apiVersion, fmt.Sprintf("/search-reindex-jobs/%s", f.createdJob.ID))

	patchReqBody := &messages.PickleDocString{
		Content: `[{ "op": "replace", "path": "/state", "value": "created" }]`,
	}

	err = f.APIFeature.IPatch(path, patchReqBody)
	if err != nil {
		return fmt.Errorf("failed to send patch request - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyList is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /search-reindex-jobs to make sure that an empty list (0 jobs) has been returned.
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
// It checks the response from calling GET /search-reindex-jobs to make sure that a list containing three or more jobs has been returned.
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
// It checks the response from calling GET /search-reindex-jobs to make sure that a list containing three or more jobs has been returned.
func (f *SearchReindexAPIFeature) iWouldExpectThereToBeFourJobsReturnedInAList() error {
	var err error

	f.responseBody, err = io.ReadAll(f.APIFeature.HttpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var response models.Jobs
	err = json.Unmarshal(f.responseBody, &response)
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
// It checks the response from calling GET /search-reindex-jobs to make sure that each job contains the expected types of values of id,
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

	for i := range response.JobList {
		job := response.JobList[i]
		err := f.checkStructure(&job, expectedResult)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}
	return f.ErrorFeature.StepError()
}

// theGeneratedIDForNewJobIsNotGoingToBeUnique is a feature step that ensures that a new job's id is not going to be unique
func (f *SearchReindexAPIFeature) theGeneratedIDForNewJobIsNotGoingToBeUnique() error {
	models.NewJobID = func() string {
		return "same_id"
	}

	return f.ErrorFeature.StepError()
}

// theJobShouldOnlyBeUpdatedWithTheFollowingFieldsAndValues is a feature step that checks the updated job in mongo with the expected result given via table
func (f *SearchReindexAPIFeature) theJobShouldOnlyBeUpdatedWithTheFollowingFieldsAndValues(table *godog.Table) error {
	ctx := context.Background()

	updatedJob, err := f.MongoClient.GetJob(ctx, f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("failed to get job from mongo - err: %w", err)
	}

	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	err = f.checkJobUpdates(f.createdJob, updatedJob, expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check job updates - err: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It checks the response from calling GET /search-reindex-jobs to make sure that the jobs are in ascending order of their last_updated
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

// theNoOfExistingJobsInTheJobStore is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It creates a job in mongo and assigns the created job to f.createdJob to be used in later feature steps
func (f *SearchReindexAPIFeature) theNoOfExistingJobsInTheJobStore(noOfJobs int) (err error) {
	ctx := context.Background()

	if noOfJobs < 0 {
		return fmt.Errorf("invalid number of jobs given - noOfJobs = %d", noOfJobs)
	}

	if noOfJobs == 0 {
		err = f.Reset(false)
		if err != nil {
			return fmt.Errorf("failed to reset the SearchReindexAPIFeature: %w", err)
		}
	}

	if noOfJobs > 0 {
		var job *models.Job
		for i := 1; i <= noOfJobs; i++ {
			// create a job in mongo
			job, err = models.NewJob(ctx, "")
			if err != nil {
				return fmt.Errorf("failed to create new job: %w", err)
			}

			err = f.MongoClient.CreateJob(ctx, *job)
			if err != nil {
				return fmt.Errorf("failed to insert new job in mongo: %w", err)
			}

			if i <= noOfJobs {
				time.Sleep(5 * time.Millisecond)
			}
		}

		f.createdJob = job
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
// It asserts that the job id and search index name that get returned by the POST /search-reindex-jobs endpoint match the ones that get sent in the
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

// theResponseETagHeaderShouldNotBeEmpty is a feature step that checks if the response ETag header should not be empty
func (f *SearchReindexAPIFeature) theResponseETagHeaderShouldNotBeEmpty() error {
	assert.NotEmpty(&f.ErrorFeature, f.APIFeature.HttpResponse.Header.Get(dpresponse.ETagHeader))
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
// After PUT /search-reindex-jobs/{id}/number-of-tasks/{number_of_tasks} has been called, followed by GET /search-reindex-jobs/{id},
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

	responseJob, err := f.getJobFromResponse()
	if err != nil {
		return err
	}
	f.createdJob = responseJob

	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	err = f.checkStructure(responseJob, expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theSearchReindexAPILosesItsConnectionToTheSearchAPI is a feature step that can be defined for a specific SearchReindexAPIFeature.
// It closes the connection to the search feature so as to mimic losing the connection to the Search API.
func (f *SearchReindexAPIFeature) theSearchReindexAPILosesItsConnectionToTheSearchAPI() error {
	f.fakeSearchAPI.Close()
	return f.ErrorFeature.StepError()
}

func (f *SearchReindexAPIFeature) successfulSearchAPIResponse() error {
	f.fakeSearchAPI.fakeHTTP.NewHandler().Post("/search").Reply(201).BodyString(`{ "IndexName": "ons1638363874110115"}`)
	return nil
}

func (f *SearchReindexAPIFeature) unsuccessfulSearchAPIResponse() error {
	f.fakeSearchAPI.fakeHTTP.NewHandler().Post("/search").Reply(500).BodyString(`internal server error`)
	return nil
}

func (f *SearchReindexAPIFeature) restartFakeSearchAPI() error {
	f.fakeSearchAPI.Restart()
	return nil
}
