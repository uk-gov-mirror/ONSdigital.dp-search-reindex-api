package api

import (
	"context"
	"io"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"encoding/json"
)

//API provides a struct to wrap the api around
type API struct {
	Router *mux.Router
}

//Setup function sets up the api and returns an api
func Setup(ctx context.Context, r *mux.Router) *API {
	api := &API{
		Router: r,
	}

	r.HandleFunc("/jobs", JobHandler(ctx)).Methods("POST")
	return api
}

// ReadJSONBody reads the bytes from the provided body, and marshals it to the provided model interface.
func (api *API) ReadJSONBody(ctx context.Context, body io.ReadCloser, v interface{}) error {
	defer body.Close()

	// Get Body bytes
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return apierrors.ErrUnableToReadMessage
	}

	// Unmarshal body bytes to model
	if err := json.Unmarshal(payload, v); err != nil {
		return apierrors.ErrUnableToParseJSON
	}

	return nil
}
