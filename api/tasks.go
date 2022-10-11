package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/pagination"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// CreateTaskHandler returns a function that generates a new TaskName resource containing default values in its fields.
func (api *API) CreateTaskHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host

	vars := mux.Vars(req)
	jobID := vars["id"]
	logData := log.Data{"job_id": jobID}

	taskToCreate := &models.TaskToCreate{}

	// get task to create from request body
	err := ReadJSONBody(req.Body, taskToCreate)
	if err != nil {
		logData["task_to_create"] = taskToCreate
		log.Error(ctx, "reading request body failed", err, logData)
		http.Error(w, apierrors.ErrInternalServer.Error(), http.StatusBadRequest)
		return
	}

	// validate task to create
	err = taskToCreate.Validate(api.taskNames)
	if err != nil {
		logData["task_to_create"] = taskToCreate.TaskName
		log.Error(ctx, "failed to validate taskToCreate", err, logData)
		http.Error(w, apierrors.ErrInvalidRequestBody.Error(), http.StatusBadRequest)
		return
	}

	// acquire job lock
	lockID, err := api.dataStore.AcquireJobLock(ctx, jobID)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.dataStore.UnlockJob(ctx, lockID)

	// check if job exists
	job, err := api.dataStore.GetJob(ctx, jobID)
	if (job == nil) || (err != nil) {
		if err == mongo.ErrJobNotFound {
			log.Error(ctx, "job not found", err, logData)
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Error(ctx, "failed to get job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// create new task
	newTask, err := models.NewTask(ctx, jobID, taskToCreate)
	if err != nil {
		logData["task_to_create"] = taskToCreate.TaskName
		log.Error(ctx, "failed to create task", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// insert new task in datastore
	err = api.dataStore.UpsertTask(ctx, jobID, taskToCreate.TaskName, *newTask)
	if err != nil {
		logData["new_task"] = newTask
		logData["task_to_create"] = taskToCreate
		log.Error(ctx, "failed to insert task to datastore", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// update links with host and version for json response
	newTask.Links.Job = fmt.Sprintf("%s/%s%s", host, v1, newTask.Links.Job)
	newTask.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, newTask.Links.Self)

	// set eTag on ETag response header
	dpresponse.SetETag(w, newTask.ETag)

	// write response
	err = dpresponse.WriteJSON(w, newTask, http.StatusCreated)
	if err != nil {
		logData["new_task"] = newTask
		logData["response_status_to_write"] = http.StatusCreated
		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// GetTaskHandler returns a function that gets a specific task, associated with an existing Job resource, using the job id and task name passed in.
func (api *API) GetTaskHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host

	vars := mux.Vars(req)
	jobID := vars["id"]
	taskName := vars["task_name"]

	logData := log.Data{
		"job_id":    jobID,
		"task_name": taskName,
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

	// check if job exists
	job, err := api.dataStore.GetJob(ctx, jobID)
	if (job == nil) || (err != nil) {
		if err == mongo.ErrJobNotFound {
			log.Error(ctx, "job not found", err, logData)
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Error(ctx, "failed to get job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// get task
	task, err := api.dataStore.GetTask(ctx, jobID, taskName)
	if err != nil {
		if err == mongo.ErrTaskNotFound {
			log.Error(ctx, "task not found", err, logData)
			http.Error(w, apierrors.ErrTaskNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Error(ctx, "error occurred when getting task", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// check eTags to see if it matches with the current state of the tasks resource
	if task.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = task.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, task.ETag)

	// update links with host and version for json response
	task.Links.Job = fmt.Sprintf("%s/%s%s", host, v1, task.Links.Job)
	task.Links.Self = fmt.Sprintf("%s/%s%s", host, v1, task.Links.Self)

	// write response
	err = dpresponse.WriteJSON(w, task, http.StatusOK)
	if err != nil {
		logData["task"] = task
		logData["response_status_to_write"] = http.StatusOK

		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// PutTaskNumOfDocsHandler returns a function that updates the number_of_documents in an existing Task resource, which is associated with a specific Job.
func (api *API) PutTaskNumOfDocsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	jobID := vars["job_id"]
	taskName := vars["task_name"]
	docCount := vars["count"]
	logData := log.Data{"job_id": jobID, "task_name": taskName, "count": docCount}

	// validate no of documents
	numOfDocs, err := strconv.Atoi(docCount)
	if err != nil {
		log.Error(ctx, "invalid path parameter - failed to convert count to integer", err, logData)
		http.Error(w, apierrors.ErrInvalidNumTasks.Error(), http.StatusBadRequest)
		return
	}

	if numOfDocs < 0 {
		err = errors.New("the count is negative")
		logData["no_of_documents"] = numOfDocs

		log.Error(ctx, "invalid path parameter - count should be a positive integer", err, logData)
		http.Error(w, apierrors.ErrInvalidNumDocs.Error(), http.StatusBadRequest)
		return
	}

	// get eTag from If-Match header
	eTag, err := headers.GetIfMatch(req)
	if err != nil {
		log.Warn(ctx, "ignoring eTag check", log.FormatErrors([]error{err}), logData)
		eTag = headers.IfMatchAnyETag
	}

	// acquire lock
	lockID, err := api.dataStore.AcquireJobLock(ctx, jobID)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.dataStore.UnlockJob(ctx, lockID)

	// get job from mongo
	_, err = api.dataStore.GetJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "failed to get job", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	task, err := api.dataStore.GetTask(ctx, jobID, taskName)
	if err != nil {
		log.Error(ctx, "failed to get task associated with the job", err, logData)
		if err == mongo.ErrTaskNotFound {
			http.Error(w, apierrors.ErrTaskNotFound.Error(), http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	// check eTags to see if it matches with the current state of the task resource
	if task.ETag != eTag && eTag != headers.IfMatchAnyETag {
		logData["current_etag"] = task.ETag
		logData["given_etag"] = eTag

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	updatedTask := *task
	updatedTask.NumberOfDocuments = numOfDocs

	// generate new etag for updated job
	updatedETag, err := models.GenerateETagForTask(ctx, updatedTask)
	if err != nil {
		logData["updated_task"] = updatedTask
		log.Error(ctx, "failed to generate new etag for task", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	updatedTask.LastUpdated = time.Now().UTC()
	// compare updatedETag with existing eTag to check for modifications
	if updatedETag == task.ETag {
		logData["updated_eTag"] = updatedETag
		logData["current_eTag"] = task.ETag
		log.Info(ctx, "no modifications made to job resource", logData)
		// set eTag on ETag response header
		dpresponse.SetETag(w, updatedETag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// update no of tasks, etag and last updated in job resource in mongo
	err = api.dataStore.UpsertTask(ctx, jobID, taskName, updatedTask)
	if err != nil {
		logData["updated_task"] = updatedTask
		log.Error(ctx, "failed to update no of tasks in job resource in mongo", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, updatedETag)

	w.WriteHeader(http.StatusNoContent)
}

// GetTasksHandler gets a list of existing Task resources, from the data store, sorted by their values of last_updated time (ascending)
func (api *API) GetTasksHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host
	offsetParam := req.URL.Query().Get("offset")
	limitParam := req.URL.Query().Get("limit")

	vars := mux.Vars(req)
	jobID := vars["id"]
	logData := log.Data{"job_id": jobID}

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
		log.Error(ctx, "failed to initialise pagination", err, logData)

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if job exists
	job, err := api.dataStore.GetJob(ctx, jobID)
	if (job == nil) || (err != nil) {
		if err == mongo.ErrJobNotFound {
			log.Error(ctx, "job not found", err, logData)
			http.Error(w, apierrors.ErrJobNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Error(ctx, "failed to get job", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	options := mongo.Options{
		Offset: offset,
		Limit:  limit,
	}

	// get tasks
	tasks, err := api.dataStore.GetTasks(ctx, jobID, options)
	if err != nil {
		logData["options"] = options
		log.Error(ctx, "getting list of tasks failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	tasksETag, err := models.GenerateETagForTasks(ctx, *tasks)
	if err != nil {
		log.Error(ctx, "failed to generate etag for tasks", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	// check eTags to see if it matches with the current state of the tasks resource
	if tasksETag != eTagFromIfMatch && eTagFromIfMatch != headers.IfMatchAnyETag {
		logData["current_etag"] = tasksETag
		logData["given_etag"] = eTagFromIfMatch

		err = apierrors.ErrConflictWithETag
		log.Error(ctx, "given and current etags do not match", err, logData)
		http.Error(w, apierrors.ErrConflictWithETag.Error(), http.StatusConflict)
		return
	}

	// update links with host and version for json response
	for i := range tasks.TaskList {
		tasks.TaskList[i].Links.Job = fmt.Sprintf("%s/%s%s", host, v1, tasks.TaskList[i].Links.Job)
		tasks.TaskList[i].Links.Self = fmt.Sprintf("%s/%s%s", host, v1, tasks.TaskList[i].Links.Self)
	}

	// set eTag on ETag response header
	dpresponse.SetETag(w, tasksETag)

	// write response
	err = dpresponse.WriteJSON(w, tasks, http.StatusOK)
	if err != nil {
		logData["tasks"] = tasks
		logData["response_status_to_write"] = http.StatusOK

		log.Error(ctx, "failed to write response", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}
