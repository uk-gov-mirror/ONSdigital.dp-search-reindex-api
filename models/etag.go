package models

import (
	"time"

	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	"github.com/globalsign/mgo/bson"
)

// GenerateETagForJob generates a new eTag for a job resource
func GenerateETagForJob(updatedJob Job) (eTag string, err error) {
	// ignoring the metadata LastUpdated and currentEtag when generating new eTag
	zeroTime := time.Time{}.UTC()
	updatedJob.ETag = ""
	updatedJob.LastUpdated = zeroTime

	b, err := bson.Marshal(updatedJob)
	if err != nil {
		return "", err
	}

	eTag = dpresponse.GenerateETag(b, false)

	return eTag, nil
}
