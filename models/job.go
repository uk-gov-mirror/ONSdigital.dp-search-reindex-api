package models

import (
	"time"
)

// Job represents a job metadata model and json representation for API
type Job struct {
	ID								string				`json:"id"`
	LastUpdated						time.Time			`json:"last_updated"`
	Links							*JobLinks			`json:"links"`
	NumberOfTasks 					int					`json:"number_of_tasks"`
	ReindexCompleted 				time.Time        	`json:"reindex_completed"`
	ReindexFailed					time.Time			`json:"reindex_failed"`
	ReindexStarted					time.Time			`json:"reindex_started"`
	SearchIndexName					string				`json:"search_index_name"`
	State							string				`json:"state"`
	TotalSearchDocuments			int					`json:"total_search_documents"`
	TotalInsertedSearchDocuments	int					`json:"total_inserted_search_documents"`
}

type JobLinks struct {
	Tasks 	string `json:"tasks"`
	Self    string `json:"self"`
}

