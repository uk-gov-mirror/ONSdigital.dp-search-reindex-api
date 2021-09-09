package url

import "fmt"

// Builder encapsulates the building of urls in a central place, with knowledge of the url structures and base host names.
type Builder struct {
	apiURL string
}

// NewBuilder returns a new instance of url.Builder
func NewBuilder(apiURL string) *Builder {
	return &Builder{
		apiURL: apiURL,
	}
}

// BuildJobURL returns the website URL for a specific reindex job
func (builder Builder) BuildJobURL(jobID string) string {
	return fmt.Sprintf("%s/jobs/%s",
		builder.apiURL, jobID)
}

// BuildJobTasksURL returns the website URL for a specific reindex job's tasks
func (builder Builder) BuildJobTasksURL(jobID string) string {
	return fmt.Sprintf("%s/jobs/%s/tasks",
		builder.apiURL, jobID)
}

// BuildJobTaskURL returns the website URL for a specific reindex job task
func (builder Builder) BuildJobTaskURL(jobID, taskName string) string {
	return fmt.Sprintf("%s/jobs/%s/tasks/%s",
		builder.apiURL, jobID, taskName)
}
