package models

import dprequest "github.com/ONSdigital/dp-net/v2/request"

// JOB STATE - Possible values of a job's state
const (
	JobStateCreated    = "created" // this is the default value of state in a new job
	JobStateCompleted  = "completed"
	JobStateFailed     = "failed"
	JobStateInProgress = "in-progress"
)

var (
	// ValidJobStates is used for logging available job states
	ValidJobStates = []string{
		JobStateCreated,
		JobStateCompleted,
		JobStateFailed,
		JobStateInProgress,
	}

	// ValidJobStatesMap is used for searching available job states
	ValidJobStatesMap = map[string]bool{
		JobStateCreated:    true,
		JobStateCompleted:  true,
		JobStateFailed:     true,
		JobStateInProgress: true,
	}
)

// PATCH-OPERATIONS - Possible patch operations available in search-reindex-api
var (
	// ValidPatchOps is used for logging available patch operations
	ValidPatchOps = []string{
		dprequest.OpReplace.String(),
	}

	// ValidPatchOpsMap is used for searching available patch operations
	ValidPatchOpsMap = map[string]bool{
		dprequest.OpReplace.String(): true,
	}
)
