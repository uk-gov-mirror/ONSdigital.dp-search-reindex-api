package models

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
