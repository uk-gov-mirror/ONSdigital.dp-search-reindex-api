package store

import (
	"context"
	"fmt"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
)

type JobStorer struct {
	JobsMap	map[string]models.Job
}

// CreateJob creates a new Job resource and stores it in the JobsMap
func (js *JobStorer) CreateJob(ctx context.Context, id string, job_details models.Job) error{

	log.Event(ctx, "creating job (need to do this part here)", log.Data{"id": id})
	//Only create a new JobsMap if one does not exist already
	if js.JobsMap == nil {
		js.JobsMap = make(map[string]models.Job)
	}
	fmt.Println("adding job to the map..")
	js.JobsMap[id] = job_details
	fmt.Println("Job id: " + js.JobsMap[id].ID)

	fmt.Println("The jobs map now..")
	fmt.Println(js.JobsMap)

	return nil
}
