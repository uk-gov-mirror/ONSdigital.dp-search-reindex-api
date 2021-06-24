package models

// Jobs represents an array of Job resources and json representation for API
type Jobs struct {
	JobList []Job `json:"jobs"`
}
