package reindex

import (
	"context"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
)

var searchAPISearchURL = "http://localhost:23900/search"

type NewIndexName struct {
	IndexName string
}

func CreateIndex(ctx context.Context) (*http.Response, error) {
	client := dphttp.NewClient()
	return client.Post(ctx, searchAPISearchURL, "", nil)
}
