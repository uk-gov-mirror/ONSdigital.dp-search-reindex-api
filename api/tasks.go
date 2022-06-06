package api

import (
	"fmt"
	"net/http"

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

// GetTasksHandler gets a list of existing Task resources, from the data store, sorted by their values of last_updated time (ascending)
func (api *API) GetTasksHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	host := req.Host
	offsetParam := req.URL.Query().Get("offset")
	limitParam := req.URL.Query().Get("limit")

	vars := mux.Vars(req)
	jobID := vars["id"]
	logData := log.Data{"job_id": jobID}

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

	// update links with host and version for json response
	for i := range tasks.TaskList {
		tasks.TaskList[i].Links.Job = fmt.Sprintf("%s/%s%s", host, v1, tasks.TaskList[i].Links.Job)
		tasks.TaskList[i].Links.Self = fmt.Sprintf("%s/%s%s", host, v1, tasks.TaskList[i].Links.Self)
	}

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
