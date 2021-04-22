package models

//Jobs represents an array of Job resources and json representation for API
type Jobs struct {
	Job_List []Job `json:"jobs"`
}
