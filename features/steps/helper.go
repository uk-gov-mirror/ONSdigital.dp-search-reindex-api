package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

// callPostJobs is a utility method that can be called by a feature step in order to call the POST /jobs endpoint
// Calling that endpoint results in the creation of a job, in the Job Store, containing a unique id and default values.
func (f *JobsFeature) callPostJobs() error {
	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPostToWithBody("/jobs", &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// GetJobByID is a utility function that is used for calling the GET /jobs/{id} endpoint.
// It checks that the id string is a valid UUID before calling the endpoint.
func (f *JobsFeature) GetJobByID(id string) error {
	_, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call GET /jobs/{id}
	err = f.APIFeature.IGet("/jobs/" + id)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}
	return nil
}

// PutNumberOfTasks is a utility function that is used for calling the PUT /jobs/{id}/number_of_tasks/{count}
// It checks that the id string is a valid UUID before calling the endpoint.
func (f *JobsFeature) PutNumberOfTasks(countStr string) error {
	var emptyBody = godog.DocString{}
	_, err := uuid.FromString(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call PUT /jobs/{id}/number_of_tasks/{count}
	err = f.APIFeature.IPut("/jobs/"+f.createdJob.ID+"/number_of_tasks/"+countStr, &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPut: %w", err)
	}
	return nil
}

// PostTaskForJob is a utility function that is used for calling POST /jobs/{id}/tasks
// The endpoint requires authorisation and a request body.
func (f *JobsFeature) PostTaskForJob(jobID string, requestBody *godog.DocString) error {
	_, err := uuid.FromString(jobID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call POST /jobs/{id}/tasks
	err = f.APIFeature.IPostToWithBody("/jobs/"+jobID+"/tasks", requestBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}
	return nil
}

// GetTaskForJob is a utility function that is used for calling GET /jobs/{id}/tasks/{task name}
func (f *JobsFeature) GetTaskForJob(jobID, taskName string) error {
	_, err := uuid.FromString(jobID)
	if err != nil {
		return fmt.Errorf("the job id should be a uuid: %w", err)
	}

	// call GET /jobs/{jobID}/tasks/{taskName}
	err = f.APIFeature.IGet("/jobs/" + jobID + "/tasks/" + taskName)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}
	return nil
}

// checkStructure is a utility function that can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *JobsFeature) checkStructure(expectedResult map[string]string) error {
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

// checkTaskStructure is a utility function that can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *JobsFeature) checkTaskStructure(id string, lastUpdated time.Time, expectedResult map[string]string, links *models.TaskLinks, taskName string) error {
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

// checkValuesInJob is a utility function that can be called by a feature step in order to check that the values
// of certain attributes, in a job, are all equal to the expected ones.
func (f *JobsFeature) checkValuesInJob(expectedResult map[string]string, job models.Job) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(job.NumberOfTasks))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_completed"], job.ReindexCompleted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_failed"], job.ReindexFailed.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_started"], job.ReindexStarted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["state"], job.State)
	assert.Equal(&f.ErrorFeature, expectedResult["total_search_documents"], strconv.Itoa(job.TotalSearchDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["total_inserted_search_documents"], strconv.Itoa(job.TotalInsertedSearchDocuments))
}

// checkValuesInTask is a utility function that can be called by a feature step in order to check that the values
// of certain attributes, in a task, are all equal to the expected ones.
func (f *JobsFeature) checkValuesInTask(expectedResult map[string]string, task models.Task) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_documents"], strconv.Itoa(task.NumberOfDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["task_name"], task.TaskName)
}

// readOutputMessages is a utility method to read the kafka messages that get sent to the producer's output channel
func (f *JobsFeature) readOutputMessages() {
	go func() {
		for {
			select {
			case f.kafkaProducerOutputData <- <-f.MessageProducer.Channels().Output:
				log.Println("read")
			case <-f.quitReadingOutput:
				return
			}
		}
	}()
}

func readAndDeserializeKafkaProducerOutput(kafkaProducerOutputData <-chan []byte) (*models.ReindexRequested, error) {
	for reindexRequestedDataBytes := range kafkaProducerOutputData {
		reindexRequestedData := &models.ReindexRequested{}
		err := schema.ReindexRequestedEvent.Unmarshal(reindexRequestedDataBytes, reindexRequestedData)
		return reindexRequestedData, err
	}
	return nil, nil
}

// readResponse is a utility method to read the JSON response that gets returned by the POST /job endpoint
func (f *JobsFeature) readResponse() error {
	if (f.createdJob == models.Job{}) {
		f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
		response := models.Job{}
		err := json.Unmarshal(f.responseBody, &response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json response: %w", err)
		}

		f.createdJob = response
		return err
	}
	return nil
}
