package models

import (
	"context"
	"time"

	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// GenerateETagForJob generates a new eTag for a job resource
func GenerateETagForJob(ctx context.Context, updatedJob Job) (eTag string, err error) {
	// ignoring the metadata LastUpdated and currentEtag when generating new eTag
	zeroTime := time.Time{}.UTC()
	updatedJob.ETag = ""
	updatedJob.LastUpdated = zeroTime

	b, err := bson.Marshal(updatedJob)
	if err != nil {
		logData := log.Data{
			"updated_job": updatedJob,
		}
		log.Error(ctx, "failed to marshal job", err, logData)
		return "", err
	}

	eTag = dpresponse.GenerateETag(b, false)

	return eTag, nil
}

// GenerateETagForJobs generates a new eTag for a jobs resource
func GenerateETagForJobs(ctx context.Context, jobs Jobs) (eTag string, err error) {
	b, err := bson.Marshal(jobs)
	if err != nil {
		log.Error(ctx, "failed to marshal jobs", err)
		return "", err
	}

	eTag = dpresponse.GenerateETag(b, false)

	return eTag, nil
}

// GenerateETagForTask generates a new eTag for a task resource
func GenerateETagForTask(ctx context.Context, task Task) (eTag string, err error) {
	// ignoring the metadata LastUpdated and currentEtag when generating new eTag
	zeroTime := time.Time{}.UTC()
	task.ETag = ""
	task.LastUpdated = zeroTime

	b, err := bson.Marshal(task)
	if err != nil {
		logData := log.Data{
			"task": task,
		}
		log.Error(ctx, "failed to marshal task", err, logData)
		return "", err
	}

	eTag = dpresponse.GenerateETag(b, false)

	return eTag, nil
}

// GenerateETagForTasks generates a new eTag for a tasks resource
func GenerateETagForTasks(ctx context.Context, tasks Tasks) (eTag string, err error) {
	b, err := bson.Marshal(tasks)
	if err != nil {
		log.Error(ctx, "failed to marshal tasks", err)
		return "", err
	}

	eTag = dpresponse.GenerateETag(b, false)

	return eTag, nil
}
