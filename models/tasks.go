package models

// Tasks represents an array of Task resources and json representation for API
type Tasks struct {
	Count      int    `bson:"count,omitempty"        json:"count"`
	TaskList   []Task `bson:"jobs,omitempty"         json:"tasks"`
	Limit      int    `bson:"limit,omitempty"        json:"limit"`
	Offset     int    `bson:"offset,omitempty"       json:"offset"`
	TotalCount int    `bson:"total_count,omitempty"  json:"total_count"`
}
