package steps

import (
	"context"
	"encoding/json"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	"github.com/ONSdigital/dp-search-reindex-api/service/mock"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type JobsFeature struct {
	ErrorFeature   componenttest.ErrorFeature
	svc            *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	ApiFeature     *componenttest.APIFeature
}

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

func (f *JobsFeature) InitAPIFeature() *componenttest.APIFeature {
	f.ApiFeature = componenttest.NewAPIFeature(f.InitialiseService)

	return f.ApiFeature
}

func (f *JobsFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I would expect id, last_updated, and links to have this structure$`, f.iWouldExpectIdLast_updatedAndLinksToHaveThisStructure)
	ctx.Step(`^the response should also contain the following JSON:$`, f.theResponseShouldAlsoContainTheFollowingJSON)

}
func (f *JobsFeature) Reset() *JobsFeature {
	return f
}
func (f *JobsFeature) Close() error {
	if f.svc != nil && f.ServiceRunning {
		f.svc.Close(context.Background())
		f.ServiceRunning = false
	}
	return nil
}
func (f *JobsFeature) InitialiseService() (http.Handler, error) {
	return f.HTTPServer.Handler, nil
}
func (f *JobsFeature) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	f.HTTPServer.Addr = bindAddr
	f.HTTPServer.Handler = router
	return f.HTTPServer
}

func (f *JobsFeature) DoGetHealthcheckOk(cfg *config.Config, time string, commit string, version string) (service.HealthChecker, error) {
	versionInfo, _ := healthcheck.NewVersionInfo(time, commit, version)
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (f *JobsFeature) iWouldExpectIdLast_updatedAndLinksToHaveThisStructure(expectedStructure *godog.DocString) error {
	//responseBody := f.ApiFeature.HttpResponse.Body

	//body, _ := ioutil.ReadAll(responseBody)

	//assert.JSONEq(&f.ErrorFeature, expectedStructure.Content, string(body))

	return f.ErrorFeature.StepError()
}

func (f *JobsFeature) theResponseShouldAlsoContainTheFollowingJSON(table *godog.Table) error {
	responseBody := f.ApiFeature.HttpResponse.Body
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(responseBody)

	var response models.Job

	_ = json.Unmarshal(body, &response)

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
