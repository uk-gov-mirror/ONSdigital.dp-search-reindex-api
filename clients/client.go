package clients

import "context"

//go:generate moq -out ./mocks/client.go -pkg mocks . Client

type Client interface {
	PostJob(ctx context.Context) ([]byte, error)
}

type Headers struct {
	ETag          string
	IfMatch       string
	UserAuthToken string
}

type Options struct {
	Offset int
	Limit  int
	Sort   string
}
