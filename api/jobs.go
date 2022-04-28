package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/pagination"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

var (
	// NewID generates a random uuid and returns it as a string.
	NewID = func() string {
		return uuid.NewV4().String()
	}

	serverErrorMessage = "internal server error"
)

// CreateJobHandler generates a new Job resource and a new ElasticSearch index associated with it	.
func (api *API) CreateJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host
	id := NewID()

	log.Info(ctx, "creating new job resource in the data store")
	newJob, err := api.dataStore.CreateJob(ctx, id)
	if err != nil {
		log.Error(ctx, "creating and storing job failed", err)
		if err == mongo.ErrExistingJobInProgress {
			http.Error(w, "existing reindex job in progress", http.StatusConflict)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	log.Info(ctx, "creating new index in ElasticSearch via the Search API")
	searchAPISearchURL := api.cfg.SearchAPIURL + "/search"
	reindexResponse, errCreateIndex := api.reindex.CreateIndex(ctx, api.cfg.ServiceAuthToken, searchAPISearchURL, api.httpClient)
	if errCreateIndex != nil {
		log.Error(ctx, "error occurred when connecting to Search API", errCreateIndex)
		if !updateJobStateToFailed(ctx, w, &newJob, api) {
			return
		}
	} else if reindexResponse.StatusCode != 201 {
		log.Info(ctx, "unexpected status returned by the search api", log.Data{"status returned by search api": reindexResponse.Status})
		if !updateJobStateToFailed(ctx, w, &newJob, api) {
			return
		}
	} else {
		newJob, err = api.updateSearchIndexName(ctx, reindexResponse, newJob, id)
		if err != nil {
			log.Error(ctx, "error occurred in updateSearchIndexName function", err)
			if !updateJobStateToFailed(ctx, w, &newJob, api) {
				return
			}
		} else {
			// As the index name was updated successfully we can send a reindex-requested event
			traceID := dprequest.NewRequestID(16)
			reindexReqEvent := models.ReindexRequested{
				JobID:       newJob.ID,
				SearchIndex: newJob.SearchIndexName,
				TraceID:     traceID,
			}

			log.Info(ctx, "sending reindex-requested event", log.Data{"reindexRequestedEvent": reindexReqEvent})

			if err = api.producer.ProduceReindexRequested(ctx, reindexReqEvent); err != nil {
				log.Error(ctx, "error while attempting to send reindex-requested event to producer", err)
				http.Error(w, serverErrorMessage, http.StatusInternalServerError)
				return
			}
			log.Info(ctx, "reindex request has been processed")
		}
	}

	newJob.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, newJob.Links.Self)
	newJob.Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, newJob.Links.Tasks)

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(newJob)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// updateJobStateToFailed returns true if the job state was successfully updated to failed
func updateJobStateToFailed(ctx context.Context, w http.ResponseWriter, newJob *models.Job, api *API) bool {
	newJob.State = models.JobStateFailed
	log.Info(ctx, "updating job state to failed", log.Data{"job id": newJob.ID})
	setStateErr := api.dataStore.UpdateJobState(models.JobStateFailed, newJob.ID)
	if setStateErr != nil {
		log.Error(ctx, "setting state to failed has failed", setStateErr)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return false
	}
	return true
}

// GetJobHandler returns a function that gets an existing Job resource, from the Job Store, that's associated with the id passed in.
func (api *API) GetJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host
	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"job_id": id}

	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.unlockJob(ctx, lockID)

	job, err := api.dataStore.GetJob(req.Context(), id)
	if err != nil {
		log.Error(ctx, "getting job failed", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, "Failed to find job in job store", http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	job.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, job.Links.Self)
	job.Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, job.Links.Tasks)

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(job)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// GetJobsHandler gets a list of existing Job resources, from the Job Store, sorted by their values of
// last_updated time (ascending).
func (api *API) GetJobsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log.Info(ctx, "Entering handler function, which calls GetJobs and returns a list of existing Job resources held in the JobStore.")
	host := req.Host
	offsetParam := req.URL.Query().Get("offset")
	limitParam := req.URL.Query().Get("limit")

	offset, limit, err := api.setUpPagination(offsetParam, limitParam)
	if err != nil {
		log.Error(ctx, "pagination validation failed", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobs, err := api.dataStore.GetJobs(ctx, offset, limit)
	if err != nil {
		log.Error(ctx, "getting list of jobs failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	for i := range jobs.JobList {
		jobs.JobList[i].Links.Self = fmt.Sprintf("%s/%s%s", host, v1, jobs.JobList[i].Links.Self)
		jobs.JobList[i].Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, jobs.JobList[i].Links.Tasks)
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(jobs)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

func (api *API) setUpPagination(offsetParam, limitParam string) (offset, limit int, err error) {
	paginator := pagination.NewPaginator(api.cfg.DefaultLimit, api.cfg.DefaultOffset, api.cfg.DefaultMaxLimit)
	return paginator.ValidatePaginationParameters(offsetParam, limitParam)
}

// unlockJob unlocks the provided job lockID
func (api *API) unlockJob(ctx context.Context, lockID string) {
	api.dataStore.UnlockJob(lockID)
	log.Info(ctx, "job lockID has unlocked successfully")
}

// PutNumTasksHandler returns a function that updates the number_of_tasks in an existing Job resource, which is associated with the id passed in.
func (api *API) PutNumTasksHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]
	count := vars["count"]
	logData := log.Data{"id": id, "count": count}

	numTasks, err := strconv.Atoi(count)
	if err != nil {
		log.Error(ctx, "invalid path parameter - failed to convert count to integer", err, logData)
		http.Error(w, "invalid path parameter - failed to convert count to integer", http.StatusBadRequest)
		return
	}

	floatNumTasks := float64(numTasks)
	isNegative := math.Signbit(floatNumTasks)
	if isNegative {
		err = errors.New("the count is negative")
	}
	if err != nil {
		log.Error(ctx, "invalid path parameter - count should be a positive integer", err, logData)
		http.Error(w, "invalid path parameter - count should be a positive integer", http.StatusBadRequest)
		return
	}

	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.unlockJob(ctx, lockID)

	err = api.dataStore.PutNumberOfTasks(ctx, id, numTasks)
	if err != nil {
		log.Error(ctx, "putting number of tasks failed", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, "Failed to find job in job store", http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// updateSearchIndexName calls the GetIndexNameFromResponse function, in the reindex package, to get the index name that was returned by the Search API.
// It then calls the UpdateIndexName function, in the mongo package, to update the search_index_name value in the relevant Job Resource in the data store.
func (api *API) updateSearchIndexName(ctx context.Context, reindexResponse *http.Response, newJob models.Job, id string) (models.Job, error) {
	defer closeResponseBody(ctx, reindexResponse)
	indexName, err := api.reindex.GetIndexNameFromResponse(ctx, reindexResponse.Body)
	if err != nil {
		log.Error(ctx, "failed to get index name from response", err)
		if newJob != (models.Job{}) {
			newJob.State = models.JobStateFailed
			log.Info(ctx, "updating job state to failed", log.Data{"job id": newJob.ID})
			setStateErr := api.dataStore.UpdateJobState(models.JobStateFailed, newJob.ID)
			if setStateErr != nil {
				log.Error(ctx, "setting state to failed has failed", setStateErr)
				return newJob, setStateErr
			}
		}
		return newJob, err
	}

	newJob.SearchIndexName = indexName
	log.Info(ctx, "updating search index name", log.Data{"job id": id, "indexName": indexName})
	err = api.dataStore.UpdateIndexName(indexName, id)

	return newJob, err
}

// PatchJobStatusHandler updates the status of a job
func (api *API) PatchJobStatusHandler(w http.ResponseWriter, req *http.Request) {
	defer dphttp.DrainBody(req)

	ctx := req.Context()
	vars := mux.Vars(req)

	// get jobID
	jobID := vars["id"]
	logData := log.Data{"job_id": jobID}

	log.Info(ctx, "starting patch operations to job", logData)

	// get eTag from If-Match header
	eTag, err := headers.GetIfMatch(req)
	if err != nil {
		log.Error(ctx, "unable to retrieve eTag from If-Match header - setting to ignore eTag check", err)
		eTag = headers.IfMatchAnyETag
	}

	// set supported patch operations
	supportedPatchOps := []dprequest.PatchOp{dprequest.OpReplace}

	// get patches from request body
	patches, err := dprequest.GetPatches(req.Body, supportedPatchOps)
	if err != nil {
		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			logData["body"] = "unavailable"
		} else {
			logData["body"] = body
		}
		logData["supported_ops"] = supportedPatchOps

		log.Error(ctx, "error obtaining patch from request body", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// acquire instance lock so that the job update is atomic
	lockID, err := api.dataStore.AcquireJobLock(ctx, jobID)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer api.unlockJob(ctx, lockID)

	// get current job by jobID
	currentJob, err := api.dataStore.GetJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "unable to retrieve job with jobID given", err, logData)

		if err == mongo.ErrJobNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	// check eTag to see if request conflicts with the current state of the job resource
	if currentJob.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = currentJob.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithJobETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// prepare patch updates to the specific job
	updatedJob, bsonUpdates, err := GetUpdatesFromJobPatches(ctx, patches, currentJob, logData)
	if err != nil {
		logData["patches"] = patches
		logData["currentJob"] = currentJob

		log.Error(ctx, "failed to prepare patches for job", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// generate eTag based on updatedJob created from patches
	newETag, err := models.GenerateETagForJob(updatedJob)
	if err != nil {
		logData["updated_job"] = updatedJob

		log.Error(ctx, "failed to new eTag for updated job", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// compare newETag with existing eTag to check for modifications
	if newETag == currentJob.ETag {
		logData["new_eTag"] = currentJob.ETag
		logData["current_eTag"] = newETag

		newETagErr := fmt.Errorf("new eTag is same as existing eTag")
		log.Error(ctx, "no modifications made to job resource", newETagErr)

		dpresponse.SetETag(w, newETag)
		http.Error(w, newETagErr.Error(), http.StatusNotModified)
		return
	}
	bsonUpdates[models.JobETagKey] = newETag

	// update job with the request patches
	err = api.dataStore.UpdateJobWithPatches(jobID, bsonUpdates)
	if err != nil {
		logData["bson_updates"] = bsonUpdates

		log.Error(ctx, "failed to update job in mongo with patch operations", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, newETag)

	w.WriteHeader(http.StatusNoContent)

	log.Info(ctx, "successfully patched status of job", logData)
}

// GetUpdatesFromJobPatches returns an updated job resource and updated bson.M resource based on updates from the patches
func GetUpdatesFromJobPatches(ctx context.Context, patches []dprequest.Patch, currentJob models.Job, logData log.Data) (jobUpdates models.Job, bsonUpdates bson.M, err error) {
	// bsonUpdates keeps track of updates to be then applied on the mongo document
	bsonUpdates = make(bson.M)
	// jobUpdates keeps track of updates as type models.Job to be then used to generate newETag
	jobUpdates = currentJob

	currentTime := time.Now().UTC()
	// prepare updates by iterating through patches
	for _, patch := range patches {
		jobUpdates, bsonUpdates, err = addJobPatchUpdate(patch.Path, patch.Value, jobUpdates, bsonUpdates, currentTime)
		if err != nil {
			logData["failed_patch"] = patch
			log.Error(ctx, "failed to add update from job patch", err, logData)
			return models.Job{}, bsonUpdates, err
		}
	}

	bsonUpdates[models.JobLastUpdatedKey] = currentTime
	jobUpdates.LastUpdated = currentTime

	return jobUpdates, bsonUpdates, nil
}

// addJobPatchUpdate looks at the path given in the patch and then depending on the path, it will retrieve the value given in the patch. If successful, the value will be
// updated and stored relative to its corresponding path within jobUpdates and bsonUpdates respectively in preparation of updating the job resource later on.
func addJobPatchUpdate(patchPath string, patchValue interface{}, jobUpdates models.Job, bsonUpdates bson.M, currentTime time.Time) (models.Job, bson.M, error) {
	switch patchPath {
	case models.JobNoOfTasksPath:
		noOfTasks, ok := patchValue.(float64)
		if !ok {
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected an integer", GetValueType(patchValue), patchPath)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobNoOfTasksKey] = int(noOfTasks)
		jobUpdates.NumberOfTasks = int(noOfTasks)
		return jobUpdates, bsonUpdates, nil

	case models.JobStatePath:
		state, ok := patchValue.(string)
		if !ok {
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected string", GetValueType(patchValue), patchPath)
			return models.Job{}, bsonUpdates, err
		}

		if state == models.JobStateInProgress {
			bsonUpdates[models.JobReindexStartedKey] = currentTime
			jobUpdates.ReindexStarted = currentTime
		}
		if state == models.JobStateFailed {
			bsonUpdates[models.JobReindexFailedKey] = currentTime
			jobUpdates.ReindexFailed = currentTime
		}
		if state == models.JobStateCompleted {
			bsonUpdates[models.JobReindexCompletedKey] = currentTime
			jobUpdates.ReindexCompleted = currentTime
		}

		if !models.ValidJobStatesMap[state] {
			err := fmt.Errorf("invalid job state `%s` for `%s` - expected %v", state, patchPath, models.ValidJobStates)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobStateKey] = state
		jobUpdates.State = state
		return jobUpdates, bsonUpdates, nil

	case models.JobTotalSearchDocumentsPath:
		totalSearchDocs, ok := patchValue.(float64)
		if !ok {
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected an integer", GetValueType(patchValue), patchPath)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobTotalSearchDocumentsKey] = int(totalSearchDocs)
		jobUpdates.TotalSearchDocuments = int(totalSearchDocs)
		return jobUpdates, bsonUpdates, nil

	default:
		err := fmt.Errorf("provided path '%s' not supported", patchPath)
		return models.Job{}, bsonUpdates, err
	}
}
