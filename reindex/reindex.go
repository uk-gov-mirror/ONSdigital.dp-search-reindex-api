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

// Reindex is a type that contains an implementation of the Indexer interface, which can be used for calling the Search API.
type Reindex struct {
}

type NewIndexName struct {
	IndexName string
}

// CreateIndex calls the Search API via the Do function of the dp-net/http/Clienter. It passes in the ServiceAuthToken to identify itself, as the Search Reindex API, to the Search API.
func (r *Reindex) CreateIndex(ctx context.Context, userAuthToken, serviceAuthToken, searchAPISearchURL string, httpClient dphttp.Clienter) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, searchAPISearchURL, nil)
	if err != nil {
		return nil, errors.New("failed to create the request for post search")
	}

	headers.SetAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	return httpClient.Do(ctx, req)
}

// GetIndexNameFromResponse unmarshalls the response body, which is passed into the function, and extracts the IndexName, which it then returns.
func (r *Reindex) GetIndexNameFromResponse(ctx context.Context, body io.ReadCloser) (string, error) {
	b, err := ioutil.ReadAll(body)
	logData := log.Data{"response_body": string(b)}
	readBodyFailedMsg := "failed to read response body"
	unmarshalBodyFailedMsg := "failed to unmarshal response body"

	if err != nil {
		log.Error(ctx, readBodyFailedMsg, err, logData)
		return "", errors.New(readBodyFailedMsg)
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var newIndexName NewIndexName

	if err = json.Unmarshal(b, &newIndexName); err != nil {
		log.Error(ctx, unmarshalBodyFailedMsg, err, logData)
		return "", errors.New(unmarshalBodyFailedMsg)
	}

	return newIndexName.IndexName, nil
}
