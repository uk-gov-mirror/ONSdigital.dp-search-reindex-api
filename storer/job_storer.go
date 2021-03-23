package storer
//
//import (
//	"context"
//	"fmt"
//	models "github.com/ONSdigital/dp-search-reindex-api/models"
//	"github.com/ONSdigital/log.go/log"
//	uuid "github.com/satori/go.uuid"
//)
//
//var jobs_map := make(map[uuid.UUID]models.Job)
//
//// CreateJob creates a new Job resource
//func (m map[uuid.UUID]models.Job) CreateJob(ctx context.Context, id string, job_details models.Job) (bool, error) {
//
//	log.Event(ctx, "creating job", log.Data{"id": id})
//
//	m := jobs_map
//
//	fmt.Println(m)
//	fmt.Println(job_details)
//
//	return true, nil
//}
