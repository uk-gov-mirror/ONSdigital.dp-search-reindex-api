package models

// Jobs represents an array of Job resources and json representation for API
type Jobs struct {
	Count      int   `bson:"count,omitempty"        json:"count"`
	JobList    []Job `bson:"jobs,omitempty"         json:"jobs"`
	Limit      int   `bson:"limit,omitempty"        json:"limit"`
	Offset     int   `bson:"offset,omitempty"       json:"offset"`
	TotalCount int   `bson:"total_count,omitempty"  json:"total_count"`
}
