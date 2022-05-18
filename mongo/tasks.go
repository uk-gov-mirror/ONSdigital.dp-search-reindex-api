package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Options struct {
	Offset int
	Limit  int
}

// CreateTask creates a new task, for the given API and job ID, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateTask(ctx context.Context, jobID, taskName string, numDocuments int) (models.Task, error) {
	log.Info(ctx, "creating task in mongo DB", log.Data{"jobID": jobID, "taskName": taskName, "numDocuments": numDocuments})

	// If an empty job id was passed in, return an error with a message.
	if jobID == "" {
		return models.Task{}, ErrEmptyIDProvided
	}

	s := m.Session.Copy()
	defer s.Close()

	// Check that the jobs collection contains the job that the task will be part of
	var jobToFind models.Job
	jobToFind.ID = jobID
	_, err := m.findJob(s, jobID, jobToFind)
	if err != nil {
		log.Error(ctx, "error finding job for task", err)
		if err == mgo.ErrNotFound {
			return models.Task{}, ErrJobNotFound
		}
		return models.Task{}, fmt.Errorf("an unexpected error has occurred: %w", err)
	}

	newTask, err := models.NewTask(jobID, taskName, numDocuments)
	if err != nil {
		return models.Task{}, fmt.Errorf("error creating task in mongo DB: %w", err)
	}

	err = m.UpsertTask(jobID, taskName, newTask)
	if err != nil {
		return models.Task{}, fmt.Errorf("error overwriting task in mongo DB: %w", err)
	}
	log.Info(ctx, "creating or overwriting task in tasks collection", log.Data{"Task details: ": newTask})

	return newTask, err
}

// GetTask retrieves the details of a particular task, from the collection, specified by its task name and associated job id
func (m *JobStore) GetTask(ctx context.Context, jobID, taskName string) (models.Task, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting task from the data store", log.Data{"jobID": jobID, "taskName": taskName})
	var task models.Task

	// If an empty jobID or taskName was passed in, return an error with a message.
	if jobID == "" {
		return task, ErrEmptyIDProvided
	} else if taskName == "" {
		return task, ErrEmptyTaskNameProvided
	}

	_, err := m.findJob(s, jobID, models.Job{})
	if err != nil {
		if err == mgo.ErrNotFound {
			return task, ErrJobNotFound
		}
		return task, err
	}

	err = s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID, "task_name": taskName}).One(&task)
	if err != nil {
		if err == mgo.ErrNotFound {
			return task, ErrTaskNotFound
		}
		return task, err
	}

	return task, err
}

// GetTasks retrieves all the tasks, from the collection, and lists them in order of last_updated
func (m *JobStore) GetTasks(ctx context.Context, options Options, jobID string) (models.Tasks, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting list of tasks")
	results := models.Tasks{}

	// If an empty jobID or taskName was passed in, return an error with a message.
	if jobID == "" {
		return results, ErrEmptyIDProvided
	}

	_, err := m.findJob(s, jobID, models.Job{})
	if err != nil {
		if err == mgo.ErrNotFound {
			return results, ErrJobNotFound
		}
		return results, err
	}

	numTasks := 0
	numTasks, err = s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Count()
	if err != nil {
		log.Error(ctx, "error counting tasks for given job id", err, log.Data{"job_id": jobID})
		return results, err
	}
	log.Info(ctx, "number of tasks found in tasks collection", log.Data{"numTasks": numTasks})

	if numTasks == 0 {
		log.Info(ctx, "there are no tasks in the data store - so the list is empty")
		results.TaskList = make([]models.Task, 0)
		return results, nil
	}

	// Get the requested tasks from the tasks collection, using the given job_id, offset, and limit, and order them by last_updated
	tasksQuery := s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Skip(options.Offset).Limit(options.Limit).Sort("last_updated")
	tasks := make([]models.Task, numTasks)
	if err := tasksQuery.All(&tasks); err != nil {
		return results, err
	}

	results.TaskList = tasks
	results.Count = len(tasks)
	results.Limit = options.Limit
	results.Offset = options.Offset
	results.TotalCount = numTasks
	log.Info(ctx, "list of tasks - sorted by last_updated", log.Data{"Sorted tasks: ": results.TaskList})

	return results, nil
}

// PutNumberOfTasks updates the number_of_tasks in a particular job, from the collection, specified by its id
func (m *JobStore) PutNumberOfTasks(ctx context.Context, id string, numTasks int) error {
	log.Info(ctx, "putting number of tasks", log.Data{"id": id, "numTasks": numTasks})

	updates := make(bson.M)
	updates["number_of_tasks"] = numTasks
	updates["last_updated"] = time.Now().UTC()

	err := m.UpdateJob(ctx, id, updates)
	if err != nil {
		logData := log.Data{
			"job_id":  id,
			"updates": updates,
		}
		log.Error(ctx, "failed to update job with number of tasks", err, logData)

		return err
	}

	return nil
}

// UpsertTask creates a new task document or overwrites an existing one
func (m *JobStore) UpsertTask(jobID, taskName string, task models.Task) error {
	s := m.Session.Copy()
	defer s.Close()

	selector := bson.M{
		"task_name": taskName,
		"job_id":    jobID,
	}

	task.LastUpdated = time.Now().UTC()

	update := bson.M{
		"$set": task,
	}

	_, err := s.DB(m.Database).C(m.TasksCollection).Upsert(selector, update)
	return err
}
