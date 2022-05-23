package reindex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
)

// Reindex is a type that contains an implementation of the Indexer interface, which can be used for calling the Search API.
type Reindex struct {
}

type NewIndexName struct {
	IndexName string
}

// CreateIndex calls the Search API via the Do function of the dp-net/v2/http/Clienter. It passes in the ServiceAuthToken to identify itself, as the Search Reindex API, to the Search API.
func (r *Reindex) CreateIndex(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dphttp.Clienter) (*http.Response, error) {
	log.Info(ctx, "creating new index in elasticSearch via the search api")
	logData := log.Data{}

	req, err := http.NewRequest(http.MethodPost, searchAPISearchURL, http.NoBody)
	if err != nil {
		log.Error(ctx, "failed to create request for post search", err)
		return nil, err
	}

	if err := headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		logData["service_auth_token"] = serviceAuthToken
		logData["request"] = *req
		log.Error(ctx, "error setting service auth token", err, logData)
		return nil, ErrSettingServiceAuth
	}

	return httpClient.Do(ctx, req)
}

// GetIndexNameFromResponse unmarshalls the response body, which is passed into the function, and extracts the IndexName, which it then returns
func (r *Reindex) GetIndexNameFromResponse(ctx context.Context, body io.ReadCloser) (string, error) {
	log.Info(ctx, "getting index name from search api response")

	b, err := io.ReadAll(body)
	logData := log.Data{"response_body": string(b)}
	if err != nil {
		log.Error(ctx, "failed to read response body", err, logData)
		return "", ErrReadBodyFailed
	}

	if len(b) == 0 {
		log.Error(ctx, "empty response body given", err, logData)
		return "", ErrResponseBodyEmpty
	}

	var newIndexName NewIndexName

	if err = json.Unmarshal(b, &newIndexName); err != nil {
		log.Error(ctx, "failed to unmarshal response body", err, logData)
		return "", ErrUnmarshalBodyFailed
	}

	return newIndexName.IndexName, nil
}
