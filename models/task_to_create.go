package models

// TaskToCreate is a type that contains the details required for creating a Task type.
type TaskToCreate struct {
	NameOfApi         string `json:"name_of_api"`
	NumberOfDocuments int    `json:"number_of_documents"`
}
