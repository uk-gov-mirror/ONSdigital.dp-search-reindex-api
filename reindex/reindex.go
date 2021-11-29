package reindex

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/pkg/errors"
)

type NewIndexName struct {
	IndexName string
}

func CreateIndex(ctx context.Context, userAuthToken, serviceAuthToken, searchAPISearchURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, searchAPISearchURL, nil)
	if err != nil {
		return nil, errors.New("failed to create the request for post search")
	}

	headers.SetAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	return dphttp.NewClient().Do(ctx, req)
}
