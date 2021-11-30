package reindex

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
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

func GetIndexNameFromResponse(ctx context.Context, body io.ReadCloser) (string, error) {
	b, err := ioutil.ReadAll(body)
	logData := log.Data{"response_body": string(b)}

	if err != nil {
		log.Error(ctx, "failed to read response body", err, logData)
		return "", errors.New("failed to read response body")
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var newIndexName NewIndexName

	if err = json.Unmarshal(b, &newIndexName); err != nil {
		log.Error(ctx, "failed to unmarshal response body", err, logData)
		return "", errors.New("failed to unmarshal response body")
	}

	return newIndexName.IndexName, nil
}
