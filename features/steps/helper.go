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

const testHost = "foo"

// CallGetJobByID can be called by a feature step in order to call the GET /search-reindex-jobs/{id} endpoint.
func (f *SearchReindexAPIFeature) CallGetJobByID(version, id string) error {
	path := getPath(version, fmt.Sprintf("/search-reindex-jobs/%s", id))

	err := f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return nil
}

// PutNumberOfTasks can be called by a feature step in order to call the PUT /search-reindex-jobs/{id}/number-of-tasks/{count} endpoint
func (f *SearchReindexAPIFeature) PutNumberOfTasks(version, id, countStr string) error {
	path := getPath(version, fmt.Sprintf("/search-reindex-jobs/%s/number-of-tasks/%s", id, countStr))

	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPut(path, &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPut: %w", err)
	}

	return nil
}

// PutNumberOfTasks can be called by a feature step in order to call the PUT /search-reindex-jobs/{id}/number-of-tasks/{count} endpoint
func (f *SearchReindexAPIFeature) PutNumberOfDocs(version, id, taskName, countStr string) error {
	path := getPath(version, fmt.Sprintf("/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", id, taskName, countStr))

	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPut(path, &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPut: %w", err)
	}

	return nil
}

// PostTaskForJob can be called by a feature step in order to call the POST /search-reindex-jobs/{id}/tasks endpoint
func (f *SearchReindexAPIFeature) PostTaskForJob(version, jobID string, requestBody *godog.DocString) error {
	path := getPath(version, fmt.Sprintf("/search-reindex-jobs/%s/tasks", jobID))

	err := f.APIFeature.IPostToWithBody(path, requestBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// GetTaskForJob can be called by a feature step in order to call the GET /search-reindex-jobs/{id}/tasks/{task name} endpoint
func (f *SearchReindexAPIFeature) GetTaskForJob(version, jobID, taskName string) error {
	path := getPath(version, fmt.Sprintf("/search-reindex-jobs/%s/tasks/%s", jobID, taskName))

	err := f.APIFeature.IGet(path)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// checkJobUpdates can be called by a feature step that checks every field of a job resource to see if any updates have been made and checks if the expected
// result have been updated to the relevant fields
func (f *SearchReindexAPIFeature) checkJobUpdates(oldJob, updatedJob *models.Job, expectedResult map[string]string) (err error) {
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
func (f *SearchReindexAPIFeature) checkUpdateForJobField(field string, oldJob, updatedJob *models.Job, expectedResult map[string]string) error {
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
func (f *SearchReindexAPIFeature) checkForNoChangeInJobField(field string, oldJob, updatedJob *models.Job) error {
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
func (f *SearchReindexAPIFeature) checkStructure(responseJob *models.Job, expectedResult map[string]string) error {
	_, err := uuid.FromString(responseJob.ID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	if responseJob.LastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + responseJob.LastUpdated.String())
	}

	replacer := strings.NewReplacer("{host}", testHost, "{latest_version}", f.Config.LatestVersion, "{id}", responseJob.ID)

	expectedLinksTasks := replacer.Replace(expectedResult["links: tasks"])
	assert.Equal(&f.ErrorFeature, expectedLinksTasks, responseJob.Links.Tasks)

	expectedLinksSelf := replacer.Replace(expectedResult["links: self"])
	assert.Equal(&f.ErrorFeature, expectedLinksSelf, responseJob.Links.Self)

	re := regexp.MustCompile(`(ons)(\d*)`)
	wordWithExpectedPattern := re.FindString(responseJob.SearchIndexName)
	assert.Equal(&f.ErrorFeature, wordWithExpectedPattern, responseJob.SearchIndexName)

	return nil
}

// checkTaskStructure can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *SearchReindexAPIFeature) checkTaskStructure(task models.Task, expectedResult map[string]string) error {
	_, err := uuid.FromString(task.JobID)
	if err != nil {
		return fmt.Errorf("the jobID should be a uuid: %w", err)
	}

	if task.LastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + task.LastUpdated.String())
	}

	replacer := strings.NewReplacer("{host}", testHost, "{latest_version}", f.Config.LatestVersion, "{id}", task.JobID, "{task_name}", task.TaskName)

	expectedLinksJob := replacer.Replace(expectedResult["links: job"])
	assert.Equal(&f.ErrorFeature, expectedLinksJob, task.Links.Job)

	expectedLinksSelf := replacer.Replace(expectedResult["links: self"])
	assert.Equal(&f.ErrorFeature, expectedLinksSelf, task.Links.Self)

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

func getPath(version, path string) string {
	if version != "" {
		path = fmt.Sprintf("/%s%s", version, path)
	}

	return path
}
