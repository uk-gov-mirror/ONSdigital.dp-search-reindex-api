//Package steps is used to define the steps that are used in the component test, which is written in godog (Go's version of cucumber).
package steps

import (
	"context"
	"encoding/json"
	"fmt"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	"github.com/ONSdigital/dp-search-reindex-api/service/mock"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/rdumont/assistdog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//JobsFeature is a type that contains all the requirements for running a godog (cucumber) feature that tests the /jobs endpoint.
type JobsFeature struct {
	ErrorFeature   componenttest.ErrorFeature
	svc            *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	ApiFeature     *componenttest.APIFeature
	responseBody   []byte
}

//NewJobsFeature returns a pointer to a new JobsFeature, which can then be used for testing the /jobs endpoint.
func NewJobsFeature() (*JobsFeature, error) {
	f := &JobsFeature{
		HTTPServer:     &http.Server{},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}
	svcErrors := make(chan error, 1)
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}
	initFunctions := &mock.InitialiserMock{
		DoGetHealthCheckFunc: f.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  f.DoGetHTTPServer,
	}
	ctx := context.Background()
	serviceList := service.NewServiceList(initFunctions)
	f.svc, err = service.Run(ctx, cfg, serviceList, "1", "", "", svcErrors)
	if err != nil {
		return nil, err
	}

	return f, nil
}

//InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (f *JobsFeature) InitAPIFeature() *componenttest.APIFeature {
	f.ApiFeature = componenttest.NewAPIFeature(f.InitialiseService)

	return f.ApiFeature
}

//RegisterSteps defines the steps within a specific JobsFeature cucumber test.
func (f *JobsFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I would expect id, last_updated, and links to have this structure$`, f.iWouldExpectIdLast_updatedAndLinksToHaveThisStructure)
	ctx.Step(`^the response should also contain the following values:$`, f.theResponseShouldAlsoContainTheFollowingValues)
	ctx.Step(`^I have generated a job in the Job Store$`, f.iHaveGeneratedAJobInTheJobStore)
	ctx.Step(`^I call GET \/jobs\/{id} using the generated id$`, f.iCallGETJobsidUsingTheGeneratedId)
	ctx.Step(`^I have generated three jobs in the Job Store$`, f.iHaveGeneratedThreeJobsInTheJobStore)
}

//Reset sets the resources within a specific JobsFeature back to their default values.
func (f *JobsFeature) Reset() *JobsFeature {
	return f
}

//Close stops the *service.Service, which is pointed to from within the specific JobsFeature, from running.
func (f *JobsFeature) Close() error {
	if f.svc != nil && f.ServiceRunning {
		f.svc.Close(context.Background())
		f.ServiceRunning = false
	}
	return nil
}

//InitialiseService returns the http.Handler that's contained within a specific JobsFeature.
func (f *JobsFeature) InitialiseService() (http.Handler, error) {
	return f.HTTPServer.Handler, nil
}

//DoGetHTTPServer takes a bind Address (string) and a router (http.Handler), which are used to set up an HTTPServer.
//The HTTPServer is in a specific JobsFeature and is returned.
func (f *JobsFeature) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	f.HTTPServer.Addr = bindAddr
	f.HTTPServer.Handler = router
	return f.HTTPServer
}

//DoGetHealthcheckOk returns a HealthChecker service for a specific JobsFeature.
func (f *JobsFeature) DoGetHealthcheckOk(cfg *config.Config, time string, commit string, version string) (service.HealthChecker, error) {
	versionInfo, _ := healthcheck.NewVersionInfo(time, commit, version)
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

//iWouldExpectIdLast_updatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific JobsFeature.
//It takes a table that contains the expected structures for id, last_updated, and links values. And it asserts whether or not these are found.
func (f *JobsFeature) iWouldExpectIdLast_updatedAndLinksToHaveThisStructure(table *godog.Table) error {
	f.responseBody, _ = ioutil.ReadAll(f.ApiFeature.HttpResponse.Body)
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		panic(err)
	}

	var response models.Job

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return err
	}

	_, err = uuid.FromString(response.ID)
	if err != nil {
		fmt.Println("Got uuid: " + response.ID)
		return err
	}

	if response.LastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + response.LastUpdated.String())
	}

	linksTasks := strings.Replace(expectedResult["links: tasks"], "{id}", response.ID, 1)

	assert.Equal(&f.ErrorFeature, linksTasks, response.Links.Tasks)

	linksSelf := strings.Replace(expectedResult["links: self"], "{id}", response.ID, 1)

	assert.Equal(&f.ErrorFeature, linksSelf, response.Links.Self)

	return f.ErrorFeature.StepError()
}

//theResponseShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
//It takes a table that contains the expected values for all the remaining attributes, of a Job resource, and it asserts whether or not these are found.
func (f *JobsFeature) theResponseShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		panic(err)
	}

	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(response.NumberOfTasks))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_completed"], response.ReindexCompleted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_failed"], response.ReindexFailed.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_started"], response.ReindexStarted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["search_index_name"], response.SearchIndexName)
	assert.Equal(&f.ErrorFeature, expectedResult["state"], response.State)
	assert.Equal(&f.ErrorFeature, expectedResult["total_search_documents"], strconv.Itoa(response.TotalSearchDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["total_inserted_search_documents"], strconv.Itoa(response.TotalInsertedSearchDocuments))

	return f.ErrorFeature.StepError()
}

//iHaveGeneratedAJobInTheJobStore is a feature step that can be defined for a specific JobsFeature.
//It calls POST /jobs with an empty body, which causes a default job resource to be generated.
//The newly created job resource is stored in the Job Store and also returned in the response body.
func (f *JobsFeature) iHaveGeneratedAJobInTheJobStore() error {
	//call POST /jobs
	f.callPostJobs()

	return f.ErrorFeature.StepError()
}

func (f *JobsFeature) callPostJobs() {
	var emptyBody = godog.DocString{}
	err := f.ApiFeature.IPostToWithBody("/jobs", &emptyBody)
	if err != nil {
		os.Exit(1)
	}
}

//iCallGETJobsidUsingTheGeneratedId is a feature step that can be defined for a specific JobsFeature.
//It gets the id from the response body, generated in the previous step, and then uses this to call GET /jobs/{id}.
func (f *JobsFeature) iCallGETJobsidUsingTheGeneratedId() error {
	f.responseBody, _ = ioutil.ReadAll(f.ApiFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return err
	}

	_, err = uuid.FromString(response.ID)
	if err != nil {
		fmt.Println("Got uuid: " + response.ID)
		return err
	}

	//call GET /jobs/{id}
	err = f.ApiFeature.IGet("/jobs/" + response.ID)
	if err != nil {
		os.Exit(1)
	}

	return f.ErrorFeature.StepError()
}

//iHaveGeneratedThreeJobsInTheJobStore is a feature step that can be defined for a specific JobsFeature.
//It calls POST /jobs with an empty body, three times, which causes three default job resources to be generated.
func (f *JobsFeature) iHaveGeneratedThreeJobsInTheJobStore() error {
	//call POST /jobs three times
	f.callPostJobs()
	f.callPostJobs()
	f.callPostJobs()

	return f.ErrorFeature.StepError()
}