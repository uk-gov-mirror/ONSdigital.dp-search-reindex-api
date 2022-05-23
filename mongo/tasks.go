package mongo

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CreateTask creates a new task, for the given API and job ID, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateTask(ctx context.Context, jobID, taskName string, numDocuments int) (models.Task, error) {
	logData := log.Data{
		"job_id":          jobID,
		"task_name":       taskName,
		"no_of_documents": numDocuments,
	}

	log.Info(ctx, "creating task in mongo DB", logData)

	// check if jobs exists
	_, err := m.findJob(ctx, jobID)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Error(ctx, "job not found", err, logData)
			return models.Task{}, ErrJobNotFound
		}

		log.Error(ctx, "error occurred when finding job in mongo", err)
		return models.Task{}, err
	}

	// create new task
	newTask, err := models.NewTask(ctx, jobID, taskName, numDocuments)
	if err != nil {
		log.Error(ctx, "failed to create task in mongo", err, logData)
		return models.Task{}, err
	}

	logData["new_task"] = newTask

	// upsert new task into mongo
	err = m.UpsertTask(ctx, jobID, taskName, newTask)
	if err != nil {
		log.Error(ctx, "failed to upsert task in mongo", err, logData)
		return models.Task{}, err
	}

	log.Info(ctx, "creating or overwriting task in tasks collection", logData)

	return newTask, nil
}

// GetTask retrieves the details of a particular task, from the collection, specified by its task name and associated job id
func (m *JobStore) GetTask(ctx context.Context, jobID, taskName string) (models.Task, error) {
	logData := log.Data{
		"jobID":    jobID,
		"taskName": taskName,
	}

	log.Info(ctx, "getting task from the data store", logData)

	if jobID == "" {
		err := ErrEmptyIDProvided
		log.Error(ctx, "empty job id given", err, logData)
		return models.Task{}, err
	}

	if taskName == "" {
		err := ErrEmptyTaskNameProvided
		log.Error(ctx, "empty task name given", err, logData)
		return models.Task{}, err
	}

	// check if job exists
	_, err := m.findJob(ctx, jobID)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Error(ctx, "job not found in mongo", err, logData)
			return models.Task{}, ErrJobNotFound
		}

		log.Error(ctx, "error occurred when finding job in mongo", err, logData)
		return models.Task{}, err
	}

	s := m.Session.Copy()
	defer s.Close()

	var task models.Task

	// find task in mongo using job_id and task_name
	err = s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID, "task_name": taskName}).One(&task)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Error(ctx, "task not found in mongo", err, logData)
			return models.Task{}, ErrTaskNotFound
		}

		log.Error(ctx, "error occurred when finding task in mongo", err, logData)
		return models.Task{}, err
	}

	return task, nil
}

// GetTasks retrieves all the tasks, from the collection, and lists them in order of last_updated
func (m *JobStore) GetTasks(ctx context.Context, options Options, jobID string) (models.Tasks, error) {
	logData := log.Data{
		"job_id":  jobID,
		"options": options,
	}

	log.Info(ctx, "getting list of tasks", logData)

	if jobID == "" {
		err := ErrEmptyIDProvided
		log.Error(ctx, "empty job id given", err, logData)
		return models.Tasks{}, err
	}

	// check if job exists
	_, err := m.findJob(ctx, jobID)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Error(ctx, "job not found in mongo", err, logData)
			return models.Tasks{}, ErrJobNotFound
		}

		log.Error(ctx, "failed to find job in mongo", err, logData)
		return models.Tasks{}, err
	}

	s := m.Session.Copy()
	defer s.Close()

	// find no of tasks related to job_id
	numTasks, err := s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Count()
	if err != nil {
		logData["database"] = m.Database
		logData["tasks_collections"] = m.TasksCollection

		log.Error(ctx, "failed to count tasks for given job id", err, logData)
		return models.Tasks{}, err
	}

	logData["no_of_tasks"] = numTasks
	log.Info(ctx, "number of tasks found in tasks collection", logData)

	if numTasks == 0 {
		log.Info(ctx, "there are no tasks in the data store")
		return models.Tasks{}, nil
	}

	// get the requested tasks from the tasks collection using the given job_id, offset, and limit, and order them by last_updated
	tasksQuery := s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Skip(options.Offset).Limit(options.Limit).Sort("last_updated")

	taskList := make([]models.Task, numTasks)
	if err := tasksQuery.All(&taskList); err != nil {
		log.Error(ctx, "failed to populate task list", err, logData)
		return models.Tasks{}, err
	}

	// create tasks
	tasks := models.Tasks{
		Count:      len(taskList),
		TaskList:   taskList,
		Limit:      options.Limit,
		Offset:     options.Offset,
		TotalCount: numTasks,
	}

	log.Info(ctx, "list of tasks - sorted by last_updated", log.Data{"sorted_tasks: ": taskList})

	return tasks, nil
}

// PutNumberOfTasks updates the number_of_tasks in a particular job, from the collection, specified by its id
func (m *JobStore) PutNumberOfTasks(ctx context.Context, id string, numTasks int) error {
	logData := log.Data{
		"job_id":   id,
		"numTasks": numTasks,
	}

	log.Info(ctx, "putting number of tasks", logData)

	updates := make(bson.M)
	updates["number_of_tasks"] = numTasks
	updates["last_updated"] = time.Now().UTC()

	err := m.UpdateJob(ctx, id, updates)
	if err != nil {
		logData["updates"] = updates
		log.Error(ctx, "failed to update job with number of tasks", err, logData)
		return err
	}

	return nil
}

// UpsertTask creates a new task document or overwrites an existing one
func (m *JobStore) UpsertTask(ctx context.Context, jobID, taskName string, task models.Task) error {
	s := m.Session.Copy()
	defer s.Close()

	logData := log.Data{
		"job_id":    jobID,
		"task_name": taskName,
		"task":      task,
	}

	selector := bson.M{
		"task_name": taskName,
		"job_id":    jobID,
	}

	task.LastUpdated = time.Now().UTC()

	update := bson.M{
		"$set": task,
	}

	_, err := s.DB(m.Database).C(m.TasksCollection).Upsert(selector, update)
	if err != nil {
		log.Error(ctx, "failed to upsert task in mongo", err, logData)
		return err
	}

	return nil
}
