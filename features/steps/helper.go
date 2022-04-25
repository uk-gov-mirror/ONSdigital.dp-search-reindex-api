package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/schema"
	"github.com/cucumber/godog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// callPostJobs can be called by a feature step in order to call the POST /jobs endpoint
// Calling that endpoint results in the creation of a job, in the Job Store, containing a unique id and default values.
func (f *SearchReindexAPIFeature) callPostJobs() error {
	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPostToWithBody("/jobs", &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// CallGetJobByID can be called by a feature step in order to call the GET /jobs/{id} endpoint.
func (f *SearchReindexAPIFeature) CallGetJobByID(id string) error {
	err := f.APIFeature.IGet("/jobs/" + id)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return nil
}

// PutNumberOfTasks can be called by a feature step in order to call the PUT /jobs/{id}/number_of_tasks/{count} endpoint
func (f *SearchReindexAPIFeature) PutNumberOfTasks(countStr string) error {
	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPut("/jobs/"+f.createdJob.ID+"/number_of_tasks/"+countStr, &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPut: %w", err)
	}

	return nil
}

// PostTaskForJob can be called by a feature step in order to call the POST /jobs/{id}/tasks endpoint
func (f *SearchReindexAPIFeature) PostTaskForJob(jobID string, requestBody *godog.DocString) error {
	err := f.APIFeature.IPostToWithBody("/jobs/"+jobID+"/tasks", requestBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// GetTaskForJob can be called by a feature step in order to call the GET /jobs/{id}/tasks/{task name} endpoint
func (f *SearchReindexAPIFeature) GetTaskForJob(jobID, taskName string) error {
	err := f.APIFeature.IGet("/jobs/" + jobID + "/tasks/" + taskName)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// checkJobUpdates can be called by a feature step that checks every field of a job resource to see if any updates have been made and checks if the expected
// result have been updated to the relevant fields
func (f *SearchReindexAPIFeature) checkJobUpdates(oldJob, updatedJob models.Job, expectedResult map[string]string) (err error) {
	// get BSON tags for all fields of a job resource
	jobBSONTags := getJobBSONTags()

	for _, field := range jobBSONTags {
		if expectedResult[field] != "" {
			// if a change is expected to occur then check the update
			err = f.checkUpdateForJobField(field, oldJob, updatedJob, expectedResult)
			if err != nil {
				return fmt.Errorf("failed to check update for job field - err: %v", err)
			}
		} else {
			err = f.checkForNoChangeInJobField(field, oldJob, updatedJob)
			if err != nil {
				return fmt.Errorf("failed to check for no change in job field - err: %v", err)
			}
		}
	}

	return nil
}

// getJobBSONTags gets the bson tags of all the fields in a job resource
func getJobBSONTags() []string {
	var jobBSONTags []string

	val := reflect.ValueOf(models.Job{})
	for i := 0; i < val.Type().NumField(); i++ {
		valBSONTag := val.Type().Field(i).Tag.Get("bson")

		switch valBSONTag {
		case "_id":
			jobBSONTags = append(jobBSONTags, models.JobIDJSONKey)
		case "links":
			jobBSONTags = append(jobBSONTags, models.JobLinksSelfKey, models.JobLinksTasksKey)
		default:
			jobBSONTags = append(jobBSONTags, valBSONTag)
		}
	}

	return jobBSONTags
}

// checkUpdateForJobField checks for an update of a given field in a job resource
func (f *SearchReindexAPIFeature) checkUpdateForJobField(field string, oldJob, updatedJob models.Job, expectedResult map[string]string) error {
	timeDifferenceCheck := 1 * time.Second

	switch field {
	case models.JobETagKey:
		assert.NotEqual(&f.ErrorFeature, oldJob.ETag, updatedJob.ETag)
	case models.JobIDJSONKey:
		assert.NotEqual(&f.ErrorFeature, oldJob.ID, updatedJob.ID)
	case models.JobLastUpdatedKey:
		assert.WithinDuration(&f.ErrorFeature, time.Now(), updatedJob.LastUpdated, timeDifferenceCheck)
	case models.JobLinksTasksKey:
		assert.NotEqual(&f.ErrorFeature, oldJob.Links.Tasks, updatedJob.Links.Tasks)
	case models.JobLinksSelfKey:
		assert.NotEqual(&f.ErrorFeature, oldJob.Links.Self, updatedJob.Links.Self)
	case models.JobNoOfTasksKey:
		assert.Equal(&f.ErrorFeature, expectedResult[field], strconv.Itoa(updatedJob.NumberOfTasks))
	case models.JobReindexCompletedKey:
		assert.WithinDuration(&f.ErrorFeature, time.Now(), updatedJob.ReindexCompleted, timeDifferenceCheck)
	case models.JobReindexFailedKey:
		assert.WithinDuration(&f.ErrorFeature, time.Now(), updatedJob.ReindexFailed, timeDifferenceCheck)
	case models.JobReindexStartedKey:
		assert.WithinDuration(&f.ErrorFeature, time.Now(), updatedJob.ReindexStarted, timeDifferenceCheck)
	case models.JobSearchIndexNameKey:
		assert.Equal(&f.ErrorFeature, expectedResult[field], updatedJob.SearchIndexName)
	case models.JobStateKey:
		assert.Equal(&f.ErrorFeature, expectedResult[field], updatedJob.State)
	case models.JobTotalSearchDocumentsKey:
		assert.Equal(&f.ErrorFeature, expectedResult[field], strconv.Itoa(updatedJob.TotalSearchDocuments))
	case models.JobTotalInsertedSearchDocumentsKey:
		assert.Equal(&f.ErrorFeature, expectedResult[field], strconv.Itoa(updatedJob.TotalInsertedSearchDocuments))
	default:
		return fmt.Errorf("missing assertion for unexpected field: %v", field)
	}

	return nil
}

// checkForNoChangeInJobField checks for no change in the value of a given field in a job resource
func (f *SearchReindexAPIFeature) checkForNoChangeInJobField(field string, oldJob, updatedJob models.Job) error {
	switch field {
	case models.JobETagKey:
		assert.Equal(&f.ErrorFeature, oldJob.ETag, updatedJob.ETag)
	case models.JobIDJSONKey:
		assert.Equal(&f.ErrorFeature, oldJob.ID, updatedJob.ID)
	case models.JobLastUpdatedKey:
		assert.Equal(&f.ErrorFeature, oldJob.LastUpdated, updatedJob.LastUpdated)
	case models.JobLinksTasksKey:
		assert.Equal(&f.ErrorFeature, oldJob.Links.Tasks, updatedJob.Links.Tasks)
	case models.JobLinksSelfKey:
		assert.Equal(&f.ErrorFeature, oldJob.Links.Self, updatedJob.Links.Self)
	case models.JobNoOfTasksKey:
		assert.Equal(&f.ErrorFeature, oldJob.NumberOfTasks, updatedJob.NumberOfTasks)
	case models.JobReindexCompletedKey:
		assert.Equal(&f.ErrorFeature, oldJob.ReindexCompleted, updatedJob.ReindexCompleted)
	case models.JobReindexFailedKey:
		assert.Equal(&f.ErrorFeature, oldJob.ReindexFailed, updatedJob.ReindexFailed)
	case models.JobReindexStartedKey:
		assert.Equal(&f.ErrorFeature, oldJob.ReindexStarted, updatedJob.ReindexStarted)
	case models.JobSearchIndexNameKey:
		assert.Equal(&f.ErrorFeature, oldJob.SearchIndexName, updatedJob.SearchIndexName)
	case models.JobStateKey:
		assert.Equal(&f.ErrorFeature, oldJob.State, updatedJob.State)
	case models.JobTotalSearchDocumentsKey:
		assert.Equal(&f.ErrorFeature, oldJob.TotalSearchDocuments, updatedJob.TotalSearchDocuments)
	case models.JobTotalInsertedSearchDocumentsKey:
		assert.Equal(&f.ErrorFeature, oldJob.TotalInsertedSearchDocuments, updatedJob.TotalInsertedSearchDocuments)
	default:
		return fmt.Errorf("missing assertion for unexpected field: %v", field)
	}

	return nil
}

// checkStructure can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *SearchReindexAPIFeature) checkStructure(expectedResult map[string]string) error {
	_, err := uuid.FromString(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	if f.createdJob.LastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + f.createdJob.LastUpdated.String())
	}

	expectedLinksTasks := strings.Replace(expectedResult["links: tasks"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksTasks = strings.Replace(expectedLinksTasks, "{id}", f.createdJob.ID, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksTasks, f.createdJob.Links.Tasks)

	expectedLinksSelf := strings.Replace(expectedResult["links: self"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{id}", f.createdJob.ID, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksSelf, f.createdJob.Links.Self)

	re := regexp.MustCompile(`(ons)(\d*)`)
	wordWithExpectedPattern := re.FindString(f.createdJob.SearchIndexName)
	assert.Equal(&f.ErrorFeature, wordWithExpectedPattern, f.createdJob.SearchIndexName)

	return nil
}

// checkTaskStructure can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *SearchReindexAPIFeature) checkTaskStructure(id string, lastUpdated time.Time, expectedResult map[string]string, links *models.TaskLinks, taskName string) error {
	_, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("the jobID should be a uuid: %w", err)
	}

	if lastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + lastUpdated.String())
	}

	expectedLinksJob := strings.Replace(expectedResult["links: job"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksJob = strings.Replace(expectedLinksJob, "{id}", id, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksJob, links.Job)

	expectedLinksSelf := strings.Replace(expectedResult["links: self"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{id}", id, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{task_name}", taskName, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksSelf, links.Self)
	return nil
}

// checkValuesInJob can be called by a feature step in order to check that the values
// of certain attributes, in a job, are all equal to the expected ones.
func (f *SearchReindexAPIFeature) checkValuesInJob(expectedResult map[string]string, job models.Job) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(job.NumberOfTasks))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_completed"], job.ReindexCompleted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_failed"], job.ReindexFailed.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_started"], job.ReindexStarted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["state"], job.State)
	assert.Equal(&f.ErrorFeature, expectedResult["total_search_documents"], strconv.Itoa(job.TotalSearchDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["total_inserted_search_documents"], strconv.Itoa(job.TotalInsertedSearchDocuments))
}

// checkValuesInTask can be called by a feature step in order to check that the values
// of certain attributes, in a task, are all equal to the expected ones.
func (f *SearchReindexAPIFeature) checkValuesInTask(expectedResult map[string]string, task models.Task) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_documents"], strconv.Itoa(task.NumberOfDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["task_name"], task.TaskName)
}

// readOutputMessages reads the kafka messages that get sent to the SearchReindexAPIFeature KafkaProducer's output channel
func (f *SearchReindexAPIFeature) readOutputMessages() {
	go func() {
		for {
			select {
			case f.kafkaProducerOutputData <- <-f.KafkaMessageProducer.Channels().Output:
				log.Println("read")
			case <-f.quitReadingOutput:
				return
			}
		}
	}()
}

// readAndDeserializeKafkaProducerOutput reads the kafka message that get sent to the SearchReindexAPIFeature KafkaProducer's output channel
// and unmarshals the content in the form of ReindexRequested
func readAndDeserializeKafkaProducerOutput(kafkaProducerOutputData <-chan []byte) (*models.ReindexRequested, error) {
	reindexRequestedDataBytes := <-kafkaProducerOutputData
	reindexRequestedData := &models.ReindexRequested{}
	err := schema.ReindexRequestedEvent.Unmarshal(reindexRequestedDataBytes, reindexRequestedData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reindex kafka message - err: %w", err)
	}

	return reindexRequestedData, err
}

// getAndSetCreatedJobFromResponse gets the previously generated job and sets it to f.createdJob so that the job resource is accessible in each step
func (f *SearchReindexAPIFeature) getAndSetCreatedJobFromResponse() error {
	if (f.createdJob == models.Job{}) {
		response, err := f.getJobFromResponse()
		if err != nil {
			return fmt.Errorf("failed to get job from response: %w", err)
		}

		f.createdJob = *response
	}

	return nil
}

// getJobFromResponse reads the job JSON response from the SearchReindexAPIFeature's HTTP response body and unmarshals it to the form of Job
func (f *SearchReindexAPIFeature) getJobFromResponse() (*models.Job, error) {
	var err error
	f.responseBody, err = io.ReadAll(f.APIFeature.HttpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body - err: %w", err)
	}

	var jobResponse models.Job
	err = json.Unmarshal(f.responseBody, &jobResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	return &jobResponse, err
}
