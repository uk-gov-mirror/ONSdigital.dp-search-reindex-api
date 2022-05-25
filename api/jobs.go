package api

import (
	"context"
	"fmt"
	"io"
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
)

var (
	serverErrorMessage = apierrors.ErrInternalServer.Error()
)

// CreateJobHandler generates a new job resource and a new elasticSearch index associated with it
func (api *API) CreateJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host
	logData := log.Data{}

	log.Info(ctx, "starting post operation of reindex job")

	// check if a new reindex job can be created
	err := api.dataStore.CheckInProgressJob(ctx)
	if err != nil {
		log.Error(ctx, "error occurred when checking to create a new reindex job", err)

		if err == mongo.ErrExistingJobInProgress {
			http.Error(w, apierrors.ErrExistingJobInProgress.Error(), http.StatusConflict)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	searchAPISearchURL := api.cfg.SearchAPIURL + "/search"

	// create new index in elasticsearch via Search API
	reindexResponse, err := api.reindex.CreateIndex(ctx, api.cfg.ServiceAuthToken, searchAPISearchURL, api.httpClient)
	if err != nil {
		log.Error(ctx, "error occurred when connecting to search API", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	} else if reindexResponse.StatusCode != 201 {
		logData["status_returned"] = reindexResponse.Status
		log.Info(ctx, "unexpected status returned by the search api", logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// get search index name from Search API response
	defer closeResponseBody(ctx, reindexResponse)
	indexName, err := api.reindex.GetIndexNameFromResponse(ctx, reindexResponse.Body)
	if err != nil {
		log.Error(ctx, "failed to get index name from search api", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// create a new job
	newJob, err := models.NewJob(ctx, indexName)
	if err != nil {
		logData["index_name"] = indexName
		log.Error(ctx, "failed to create new job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// checks if there exists a reindex job with the same id as the new job in the datastore
	err = api.dataStore.ValidateJobIDUnique(ctx, newJob.ID)
	if err != nil {
		logData["jobID"] = newJob.ID
		log.Error(ctx, "failed to validate job id is unique", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// insert new job in the datastore
	err = api.dataStore.CreateJob(ctx, *newJob)
	if err != nil {
		logData["new_job"] = newJob
		log.Error(ctx, "failed to create reindex job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	reindexReqEvent := models.ReindexRequested{
		JobID:       newJob.ID,
		SearchIndex: newJob.SearchIndexName,
		TraceID:     dprequest.NewRequestID(16),
	}

	// send a reindex-requested event
	if err = api.producer.ProduceReindexRequested(ctx, reindexReqEvent); err != nil {
		logData["reindex_requested_event"] = reindexReqEvent
		log.Error(ctx, "failed to send reindex-requested event to producer", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// update links for json response
	newJob.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, newJob.Links.Self)
	newJob.Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, newJob.Links.Tasks)

	// set eTag on ETag response header
	dpresponse.SetETag(w, newJob.ETag)

	// write response
	err = dpresponse.WriteJSON(w, newJob, http.StatusCreated)
	if err != nil {
		logData["new_job"] = newJob
		logData["response_status_to_write"] = http.StatusCreated

		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// GetJobHandler returns a function that gets an existing Job resource, from the Job Store, that's associated with the id passed in
func (api *API) GetJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host

	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"job_id": id}

	// get eTag from If-Match header
	eTag, err := headers.GetIfMatch(req)
	if err != nil {
		if err != headers.ErrHeaderNotFound {
			log.Error(ctx, "if-match header not found", err, logData)
		} else {
			log.Error(ctx, "unable to get eTag from if-match header", err, logData)
		}

		log.Info(ctx, "ignoring eTag check")
		eTag = headers.IfMatchAnyETag
	}

	// acquire job lock
	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.dataStore.UnlockJob(ctx, lockID)

	// get job from mongo
	job, err := api.dataStore.GetJob(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get job", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	// check eTags to see if it matches with the current state of the jobs resource
	if job.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = job.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, job.ETag)

	// update links for json response
	job.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, job.Links.Self)
	job.Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, job.Links.Tasks)

	// write response
	err = dpresponse.WriteJSON(w, job, http.StatusOK)
	if err != nil {
		logData["job"] = job
		logData["response_status_to_write"] = http.StatusOK

		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// GetJobsHandler gets a list of existing Job resources, from the Job Store, sorted by their values of last_updated time (ascending)
func (api *API) GetJobsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	log.Info(ctx, "starting operation to get all job resources")

	host := req.Host
	offsetParam := req.URL.Query().Get("offset")
	limitParam := req.URL.Query().Get("limit")
	logData := log.Data{}

	// get eTag from If-Match header
	eTagFromIfMatch, err := headers.GetIfMatch(req)
	if err != nil {
		if err != headers.ErrHeaderNotFound {
			log.Error(ctx, "if-match header not found", err, logData)
		} else {
			log.Error(ctx, "unable to get eTag from if-match header", err, logData)
		}

		log.Info(ctx, "ignoring eTag check")
		eTagFromIfMatch = headers.IfMatchAnyETag
	}

	// initialise pagination
	offset, limit, err := pagination.InitialisePagination(api.cfg, offsetParam, limitParam)
	if err != nil {
		logData["offset_parameter"] = offsetParam
		logData["limit_parameter"] = limitParam

		log.Error(ctx, "pagination validation failed", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	options := mongo.Options{
		Offset: offset,
		Limit:  limit,
	}

	// get jobs from mongo
	jobs, err := api.dataStore.GetJobs(ctx, options)
	if err != nil {
		logData["options"] = options
		log.Error(ctx, "getting list of jobs failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	logData["jobs_count"] = jobs.Count
	logData["jobs_limit"] = jobs.Limit
	logData["jobs_offset"] = jobs.Offset
	logData["jobs_total_count"] = jobs.TotalCount

	jobsETag, err := models.GenerateETagForJobs(ctx, *jobs)
	if err != nil {
		log.Error(ctx, "failed to generate etag for jobs", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// check eTags to see if it matches with the current state of the jobs resource
	if jobsETag != eTagFromIfMatch && eTagFromIfMatch != headers.IfMatchAnyETag {
		logData["current_etag"] = jobsETag
		logData["given_etag"] = eTagFromIfMatch

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	// update links for json response
	for i := range jobs.JobList {
		jobs.JobList[i].Links.Self = fmt.Sprintf("%s/%s%s", host, v1, jobs.JobList[i].Links.Self)
		jobs.JobList[i].Links.Tasks = fmt.Sprintf("%s/%s%s", host, v1, jobs.JobList[i].Links.Tasks)
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, jobsETag)

	// write response
	err = dpresponse.WriteJSON(w, jobs, http.StatusOK)
	if err != nil {
		logData["response_status_to_write"] = http.StatusOK

		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// PutNumTasksHandler returns a function that updates the number_of_tasks in an existing Job resource, which is associated with the id passed in.
func (api *API) PutNumTasksHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]
	count := vars["count"]
	logData := log.Data{"id": id, "count": count}

	// validate no of tasks
	numTasks, err := strconv.Atoi(count)
	if err != nil {
		log.Error(ctx, "invalid path parameter - failed to convert count to integer", err, logData)
		http.Error(w, apierrors.ErrInvalidNumTasks.Error(), http.StatusBadRequest)
		return
	}

	if numTasks < 0 {
		err = errors.New("the count is negative")
		logData["no_of_tasks"] = numTasks

		log.Error(ctx, "invalid path parameter - count should be a positive integer", err, logData)
		http.Error(w, apierrors.ErrInvalidNumTasks.Error(), http.StatusBadRequest)
		return
	}

	// get eTag from If-Match header
	eTag, err := headers.GetIfMatch(req)
	if err != nil {
		if err != headers.ErrHeaderNotFound {
			log.Error(ctx, "if-match header not found", err, logData)
		} else {
			log.Error(ctx, "unable to get eTag from if-match header", err, logData)
		}

		log.Info(ctx, "ignoring eTag check")
		eTag = headers.IfMatchAnyETag
	}

	// acquire lock
	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.dataStore.UnlockJob(ctx, lockID)

	// get job from mongo
	job, err := api.dataStore.GetJob(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get job", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	// check eTags to see if it matches with the current state of the job resource
	if job.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = job.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	updatedJob := *job
	updatedJob.NumberOfTasks = numTasks
	updatedJob.LastUpdated = time.Now().UTC()

	// generate new etag for updated job
	updatedETag, err := models.GenerateETagForJob(ctx, updatedJob)
	if err != nil {
		logData["updated_job"] = updatedJob
		log.Error(ctx, "failed to generate new etag for job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	updates := make(bson.M)
	updates[models.JobETagKey] = updatedETag
	updates[models.JobNoOfTasksKey] = updatedJob.NumberOfTasks
	updates[models.JobLastUpdatedKey] = updatedJob.LastUpdated

	// update no of tasks, etag and last updated in job resource in mongo
	err = api.dataStore.UpdateJob(ctx, id, updates)
	if err != nil {
		logData["updates"] = updates
		log.Error(ctx, "failed to update no of tasks in job resource in mongo", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, updatedETag)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
		if err != headers.ErrHeaderNotFound {
			log.Error(ctx, "if-match header not found", err, logData)
		} else {
			log.Error(ctx, "unable to get eTag from if-match header", err, logData)
		}

		log.Info(ctx, "ignoring eTag check")
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
		http.Error(w, apierrors.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	defer api.dataStore.UnlockJob(ctx, lockID)

	// get current job by jobID
	currentJob, err := api.dataStore.GetJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "unable to retrieve job with jobID given", err, logData)

		if err == mongo.ErrJobNotFound {
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
		} else {
			http.Error(w, apierrors.ErrInternalServer.Error(), http.StatusInternalServerError)
		}

		return
	}

	// check eTags to see if it matches with the current state of the jobs resource
	if currentJob.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = currentJob.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	// prepare patch updates to the specific job
	updatedJob, bsonUpdates, err := GetUpdatesFromJobPatches(ctx, patches, currentJob)
	if err != nil {
		logData["patches"] = patches
		logData["currentJob"] = currentJob

		log.Error(ctx, "failed to prepare patches for job", err, logData)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// generate eTag based on updatedJob created from patches
	newETag, err := models.GenerateETagForJob(ctx, updatedJob)
	if err != nil {
		logData["updated_job"] = updatedJob

		log.Error(ctx, "failed to new eTag for updated job", err, logData)
		http.Error(w, apierrors.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// compare newETag with existing eTag to check for modifications
	if newETag == currentJob.ETag {
		logData["new_eTag"] = currentJob.ETag
		logData["current_eTag"] = newETag

		newETagErr := fmt.Errorf("new eTag is same as existing eTag")
		log.Error(ctx, "no modifications made to job resource", newETagErr, logData)

		dpresponse.SetETag(w, newETag)
		http.Error(w, apierrors.ErrNewETagSame.Error(), http.StatusNotModified)
		return
	}
	bsonUpdates[models.JobETagKey] = newETag

	// update job with the request patches
	err = api.dataStore.UpdateJob(ctx, jobID, bsonUpdates)
	if err != nil {
		logData["bson_updates"] = bsonUpdates

		log.Error(ctx, "failed to update job in mongo with patch operations", err, logData)
		http.Error(w, apierrors.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, newETag)

	w.WriteHeader(http.StatusNoContent)

	log.Info(ctx, "successfully patched status of job", logData)
}

// GetUpdatesFromJobPatches returns an updated job resource and updated bson.M resource based on updates from the patches
func GetUpdatesFromJobPatches(ctx context.Context, patches []dprequest.Patch, currentJob *models.Job) (jobUpdates models.Job, bsonUpdates bson.M, err error) {
	// bsonUpdates keeps track of updates to be then applied on the mongo document
	bsonUpdates = make(bson.M)
	// jobUpdates keeps track of updates as type models.Job to be then used to generate newETag
	jobUpdates = *currentJob

	currentTime := time.Now().UTC()
	// prepare updates by iterating through patches
	for _, patch := range patches {
		jobUpdates, bsonUpdates, err = addJobPatchUpdate(ctx, patch.Path, patch.Value, jobUpdates, bsonUpdates, currentTime)
		if err != nil {
			logData := log.Data{
				"job_id":       jobUpdates.ID,
				"failed_patch": patch,
			}
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
func addJobPatchUpdate(ctx context.Context, patchPath string, patchValue interface{}, jobUpdates models.Job, bsonUpdates bson.M, currentTime time.Time) (models.Job, bson.M, error) {
	logData := log.Data{"job_id": jobUpdates.ID}

	switch patchPath {
	case models.JobNoOfTasksPath:
		noOfTasks, ok := patchValue.(float64)
		if !ok {
			logData["patch_value"] = patchValue
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected an integer", GetValueType(patchValue), patchPath)

			log.Error(ctx, "invalid type for no_of_tasks", err, logData)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobNoOfTasksKey] = int(noOfTasks)
		jobUpdates.NumberOfTasks = int(noOfTasks)
		return jobUpdates, bsonUpdates, nil

	case models.JobStatePath:
		state, ok := patchValue.(string)
		if !ok {
			logData["patch_value"] = patchValue
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected string", GetValueType(patchValue), patchPath)

			log.Error(ctx, "invalid type for state", err, logData)
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
			log.Error(ctx, "invalid state given", err, logData)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobStateKey] = state
		jobUpdates.State = state
		return jobUpdates, bsonUpdates, nil

	case models.JobTotalSearchDocumentsPath:
		totalSearchDocs, ok := patchValue.(float64)
		if !ok {
			logData["patch_value"] = patchValue
			err := fmt.Errorf("wrong value type `%s` for `%s`, expected an integer", GetValueType(patchValue), patchPath)

			log.Error(ctx, "invalid type for total_search_documents", err, logData)
			return models.Job{}, bsonUpdates, err
		}

		bsonUpdates[models.JobTotalSearchDocumentsKey] = int(totalSearchDocs)
		jobUpdates.TotalSearchDocuments = int(totalSearchDocs)
		return jobUpdates, bsonUpdates, nil

	default:
		err := fmt.Errorf("provided path '%s' not supported", patchPath)
		log.Error(ctx, "provided patch path not supported", err, logData)
		return models.Job{}, bsonUpdates, err
	}
}
