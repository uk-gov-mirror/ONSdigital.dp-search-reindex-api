package store

import (
	"context"
	"fmt"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
	"time"
)

type JobStorer struct {
	JobsMap map[string]models.Job
}

// CreateJob creates a new Job resource and stores it in the JobsMap
func (js *JobStorer) CreateJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "creating job", log.Data{"id": id})
	newJob := models.Job{
		ID:          id,
		LastUpdated: time.Now().UTC(),
		Links: &models.JobLinks{
			Tasks: "http://localhost:12150/jobs/" + id + "/tasks",
			Self:  "http://localhost:12150/jobs/" + id,
		},
		NumberOfTasks:                0,
		ReindexCompleted:             time.Time{}.UTC(),
		ReindexFailed:                time.Time{}.UTC(),
		ReindexStarted:               time.Time{}.UTC(),
		SearchIndexName:              "Default Search Index Name",
		State:                        "created",
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}
	//Only create a new JobsMap if one does not exist already
	if js.JobsMap == nil {
		js.JobsMap = make(map[string]models.Job)
	}
	fmt.Println("adding job to the map..")
	js.JobsMap[id] = newJob
	fmt.Println("Job id: " + js.JobsMap[id].ID)

	fmt.Println("The jobs map now..")
	fmt.Println(js.JobsMap)

	return newJob, nil
}
