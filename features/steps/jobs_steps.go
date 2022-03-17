// Package steps is used to define the steps that are used in the component test, which is written in godog (Go's version of cucumber).
package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	clientssitesearch "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-authorisation/auth"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/event"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/schema"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	serviceMock "github.com/ONSdigital/dp-search-reindex-api/service/mock"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/rdumont/assistdog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// collection names
const jobsCol = "jobs"
const tasksCol = "tasks"

var (
	// create map of valid task_name values for testing (in place of TaskNameValues in config)
	taskName1, taskName2, taskName3, taskName4 = "dataset-api", "zebedee", "another-task-name3", "another-task-name4"

	taskNames = map[string]bool{
		taskName1: true,
		taskName2: true,
		taskName3: true,
		taskName4: true,
	}
)

// JobsFeature is a type that contains all the requirements for running a godog (cucumber) feature that tests the /jobs endpoint.
type JobsFeature struct {
	ErrorFeature            componentTest.ErrorFeature
	svc                     *service.Service
	errorChan               chan error
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	APIFeature              *componentTest.APIFeature
	responseBody            []byte
	MongoClient             *mongo.JobStore
	MongoFeature            *componentTest.MongoFeature
	AuthFeature             *componentTest.AuthorizationFeature
	SearchFeature           *SearchFeature
	KafkaProducer           service.KafkaProducer
	MessageProducer         kafka.IProducer
	quitReadingOutput       chan bool
	createdJob              models.Job
	kafkaProducerOutputData chan []byte
}

// NewJobsFeature returns a pointer to a new JobsFeature, which can then be used for testing the /jobs endpoint.
func NewJobsFeature(mongoFeature *componentTest.MongoFeature,
	authFeature *componentTest.AuthorizationFeature,
	searchFeature *SearchFeature) (*JobsFeature, error) {
	f := &JobsFeature{
		HTTPServer:     &http.Server{},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}
	svcErrors := make(chan error, 1)
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	mongodb := &mongo.JobStore{
		JobsCollection:  jobsCol,
		TasksCollection: tasksCol,
		Database:        utils.RandomDatabase(),
		URI:             mongoFeature.Server.URI(),
	}
	ctx := context.Background()
	if dbErr := mongodb.Init(ctx, cfg); dbErr != nil {
		return nil, fmt.Errorf("failed to initialise mongo DB: %w", dbErr)
	}

	f.MongoClient = mongodb

	f.AuthFeature = authFeature
	cfg.ZebedeeURL = f.AuthFeature.FakeAuthService.ResolveURL("")

	f.SearchFeature = searchFeature
	cfg.SearchAPIURL = f.SearchFeature.FakeSearchAPI.ResolveURL("")

	messageProducer := kafkatest.NewMessageProducer(true)
	messageProducer.CheckerFunc = funcCheck
	f.MessageProducer = messageProducer

	kafkaProducer := &event.ReindexRequestedProducer{
		Marshaller: schema.ReindexRequestedEvent,
		Producer:   messageProducer,
	}
	f.KafkaProducer = kafkaProducer

	f.kafkaProducerOutputData = make(chan []byte, 9999)
	f.quitReadingOutput = make(chan bool)
	f.readOutputMessages()

	err = runJobsFeatureService(ctx, f, cfg, svcErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to run JobsFeature service: %w", err)
	}

	return f, nil
}

// runJobsFeatureService uses the InitialiserMock to create the mock services that are required for testing the JobsFeature
// It then uses these to create the external serviceList, which it passes into the service.Run function along with the fake identity and search clients.
func runJobsFeatureService(ctx context.Context, f *JobsFeature, cfg *config.Config, svcErrors chan error) error {
	var err error
	initFunctions := &serviceMock.InitialiserMock{
		DoGetHealthCheckFunc:           f.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:            f.DoGetHTTPServer,
		DoGetMongoDBFunc:               f.DoGetMongoDB,
		DoGetAuthorisationHandlersFunc: f.DoGetAuthorisationHandlers,
		DoGetKafkaProducerFunc:         f.DoGetKafkaProducer,
	}

	serviceList := service.NewServiceList(initFunctions)
	testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)
	testSearchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

	f.svc, err = service.Run(ctx, cfg, serviceList, "1", "", "", svcErrors, testIdentityClient, taskNames, testSearchClient)
	return err
}

// InitAPIFeature initialises the APIFeature that's contained within a specific JobsFeature.
func (f *JobsFeature) InitAPIFeature() *componentTest.APIFeature {
	f.APIFeature = componentTest.NewAPIFeature(f.InitialiseService)

	return f.APIFeature
}

// RegisterSteps defines the steps within a specific JobsFeature cucumber test.
func (f *JobsFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the response should also contain the following values:$`, f.theResponseShouldAlsoContainTheFollowingValues)
	ctx.Step(`^I have generated a job in the Job Store$`, f.iHaveGeneratedAJobInTheJobStore)
	ctx.Step(`^I call GET \/jobs\/{id} using the generated id$`, f.iCallGETJobsidUsingTheGeneratedID)
	ctx.Step(`^I have generated three jobs in the Job Store$`, f.iHaveGeneratedThreeJobsInTheJobStore)
	ctx.Step(`^I would expect there to be three or more jobs returned in a list$`, f.iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList)
	ctx.Step(`^I would expect there to be four jobs returned in a list$`, f.iWouldExpectThereToBeFourJobsReturnedInAList)
	ctx.Step(`^each job should also contain the following values:$`, f.eachJobShouldAlsoContainTheFollowingValues)
	ctx.Step(`^the jobs should be ordered, by last_updated, with the oldest first$`, f.theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst)
	ctx.Step(`^no jobs have been generated in the Job Store$`, f.noJobsHaveBeenGeneratedInTheJobStore)
	ctx.Step(`^I call GET \/jobs\/{"([^"]*)"} using a valid UUID$`, f.iCallGETJobsUsingAValidUUID)
	ctx.Step(`^the response should contain the new number of tasks$`, f.theResponseShouldContainTheNewNumberOfTasks)
	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{(\d+)} using the generated id$`, f.iCallPUTJobsidnumberTofTasksUsingTheGeneratedID)
	ctx.Step(`^I would expect the response to be an empty list$`, f.iWouldExpectTheResponseToBeAnEmptyList)
	ctx.Step(`^I call PUT \/jobs\/{"([^"]*)"}\/number_of_tasks\/{(\d+)} using a valid UUID$`, f.iCallPUTJobsNumberoftasksUsingAValidUUID)
	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{"([^"]*)"} using the generated id with an invalid count$`, f.iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount)
	ctx.Step(`^I call PUT \/jobs\/{id}\/number_of_tasks\/{"([^"]*)"} using the generated id with a negative count$`, f.iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount)
	ctx.Step(`^the search reindex api loses its connection to mongo DB$`, f.theSearchReindexAPILosesItsConnectionToMongoDB)
	ctx.Step(`^I have generated six jobs in the Job Store$`, f.iHaveGeneratedSixJobsInTheJobStore)
	ctx.Step(`^I call POST \/jobs\/{id}\/tasks using the generated id$`, f.iCallPOSTJobsidtasksUsingTheGeneratedID)
	ctx.Step(`^I would expect job_id, last_updated, and links to have this structure$`, f.iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure)
	ctx.Step(`^the task resource should also contain the following values:$`, f.theTaskResourceShouldAlsoContainTheFollowingValues)
	ctx.Step(`^a new task resource is created containing the following values:$`, f.aNewTaskResourceIsCreatedContainingTheFollowingValues)
	ctx.Step(`^I call POST \/jobs\/{id}\/tasks to update the number_of_documents for that task$`, f.iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask)
	ctx.Step(`^I am not identified by zebedee$`, f.iAmNotIdentifiedByZebedee)
	ctx.Step(`^I have created a task for the generated job$`, f.iHaveCreatedATaskForTheGeneratedJob)
	ctx.Step(`^no tasks have been created in the tasks collection$`, f.noTasksHaveBeenCreatedInTheTasksCollection)
	ctx.Step(`^I call GET \/jobs\/{"([^"]*)"}\/tasks\/{"([^"]*)"} using a valid UUID$`, f.iCallGETJobsTasksUsingAValidUUID)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks\/{"([^"]*)"}$`, f.iCallGETJobsidtasks)
	ctx.Step(`^I call POST \/jobs\/{id}\/tasks using the same id again$`, f.iCallPOSTJobsidtasksUsingTheSameIDAgain)
	ctx.Step(`^I GET "\/jobs\/{"([^"]*)"}\/tasks"$`, f.iGETJobsTasks)
	ctx.Step(`^in each task I would expect job_id, last_updated, and links to have this structure$`, f.inEachTaskIWouldExpectJobIDLastUpdatedAndLinksToHaveThisStructure)
	ctx.Step(`^each task should also contain the following values:$`, f.eachTaskShouldAlsoContainTheFollowingValues)
	ctx.Step(`^the tasks should be ordered, by last_updated, with the oldest first$`, f.theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst)
	ctx.Step(`^I GET \/jobs\/{id}\/tasks using the generated id$`, f.iGETJobsidtasksUsingTheGeneratedID)
	ctx.Step(`^I would expect the response to be an empty list of tasks$`, f.iWouldExpectTheResponseToBeAnEmptyListOfTasks)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks using the same id again$`, f.iCallGETJobsidtasksUsingTheSameIDAgain)
	ctx.Step(`^I call GET \/jobs\/{id}\/tasks\?offset="([^"]*)"&limit="([^"]*)"$`, f.iCallGETJobsidtasksoffsetLimit)
	ctx.Step(`^I would expect there to be (\d+) tasks returned in a list$`, f.iWouldExpectThereToBeTasksReturnedInAList)
	ctx.Step(`^the response should contain values that have these structures$`, f.theResponseShouldContainValuesThatHaveTheseStructures)
	ctx.Step(`^in each job I would expect the response to contain values that have these structures$`, f.inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures)
	ctx.Step(`^the search reindex api loses its connection to the search api$`, f.theSearchReindexAPILosesItsConnectionToTheSearchAPI)
	ctx.Step(`^the response should contain a state of "([^"]*)"$`, f.theResponseShouldContainAStateOf)
	ctx.Step(`^the reindex-requested event should contain the expected job ID and search index name$`, f.theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName)
}

// iAmNotIdentifiedByZebedee is a feature step that can be defined for a specific JobsFeature.
// It enables the fake Zebedee service to return a 401 unauthorized error, and the message "user not authenticated", when the /identity endpoint is called.
func (f *JobsFeature) iAmNotIdentifiedByZebedee() error {
	f.AuthFeature.FakeAuthService.NewHandler().Get("/identity").Reply(401).BodyString(`{ "message": "user not authenticated"}`)
	return nil
}

// Reset sets the resources within a specific JobsFeature back to their default values.
func (f *JobsFeature) Reset(mongoFail bool) error {
	if mongoFail {
		f.MongoClient.Database = "lost database connection"
	} else {
		f.MongoClient.Database = utils.RandomDatabase()
	}
	if f.Config == nil {
		cfg, err := config.Get()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		f.Config = cfg
	}

	return nil
}

// Close stops the *service.Service, which is pointed to from within the specific JobsFeature, from running.
func (f *JobsFeature) Close() error {
	close(f.quitReadingOutput)
	if f.svc != nil && f.ServiceRunning {
		err := f.svc.Close(context.Background())
		if err != nil {
			return fmt.Errorf("failed to close JobsFeature service: %w", err)
		}
		f.ServiceRunning = false
	}
	return nil
}

// InitialiseService returns the http.Handler that's contained within a specific JobsFeature.
func (f *JobsFeature) InitialiseService() (http.Handler, error) {
	return f.HTTPServer.Handler, nil
}

// DoGetHTTPServer takes a bind Address (string) and a router (http.Handler), which are used to set up an HTTPServer.
// The HTTPServer is in a specific JobsFeature and is returned.
func (f *JobsFeature) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	f.HTTPServer.Addr = bindAddr
	f.HTTPServer.Handler = router
	return f.HTTPServer
}

// DoGetHealthcheckOk returns a mock HealthChecker service for a specific JobsFeature.
func (f *JobsFeature) DoGetHealthcheckOk(cfg *config.Config, curTime, commit, version string) (service.HealthChecker, error) {
	return &serviceMock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

// DoGetMongoDB returns a MongoDB, for the component test, which has a random database name and different URI to the one used by the API under test.
func (f *JobsFeature) DoGetMongoDB(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
	return f.MongoClient, nil
}

// DoGetAuthorisationHandlers returns the mock AuthHandler that was created in the NewJobsFeature function.
func (f *JobsFeature) DoGetAuthorisationHandlers(ctx context.Context, cfg *config.Config) api.AuthHandler {
	authClient := auth.NewPermissionsClient(dpHTTP.NewClient())
	authVerifier := auth.DefaultPermissionsVerifier()

	// for checking caller permissions when we only have a user/service token
	permissions := auth.NewHandler(
		auth.NewPermissionsRequestBuilder(cfg.ZebedeeURL),
		authClient,
		authVerifier,
	)
	return permissions
}

// DoGetKafkaProducer returns a mock kafka producer.
func (f *JobsFeature) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
	return f.KafkaProducer, nil
}

// iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected structures for job_id, last_updated, and links values. And it asserts whether or not these are found.
func (f *JobsFeature) iWouldExpectJobIDLastupdatedAndLinksToHaveThisStructure(table *godog.Table) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	var response models.Task

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	jobID := response.JobID
	lastUpdated := response.LastUpdated
	links := response.Links
	taskName := response.TaskName

	err = f.checkTaskStructure(jobID, lastUpdated, expectedResult, links, taskName)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected structures for job_id, last_updated, links, and search_index_name values.
// And it asserts whether or not these are found.
func (f *JobsFeature) theResponseShouldContainValuesThatHaveTheseStructures(table *godog.Table) error {
	var err error

	err = f.readResponse()
	if err != nil {
		return err
	}

	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	err = f.checkStructure(expectedResult)
	if err != nil {
		return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// aNewTaskResourceIsCreatedContainingTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It checks that a task has been created containing the expected values of number_of_documents and task_name that are passed in via the table.
func (f *JobsFeature) aNewTaskResourceIsCreatedContainingTheFollowingValues(table *godog.Table) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	assist := assistdog.NewDefault()

	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}

	var response models.Task

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.checkValuesInTask(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// theResponseShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected values for all the remaining attributes, of a Job resource, and it asserts whether or not these are found.
func (f *JobsFeature) theResponseShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	f.checkValuesInJob(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// theTaskResourceShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It takes a table that contains the expected values for all the remaining attributes, of a TaskName resource, and it asserts whether or not these are found.
func (f *JobsFeature) theTaskResourceShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Task

	_ = json.Unmarshal(f.responseBody, &response)

	f.checkValuesInTask(expectedResult, response)

	return f.ErrorFeature.StepError()
}

// eachTaskShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It gets the list of tasks from the response and checks that each task contains the expected number of documents and a valid task name.
// NB. The valid task names are listed in the taskNames variable.
func (f *JobsFeature) eachTaskShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("unable to parse the table of values: %w", err)
	}
	var response models.Tasks

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for _, task := range response.TaskList {
		assert.Equal(&f.ErrorFeature, expectedResult["number_of_documents"], strconv.Itoa(task.NumberOfDocuments))
		assert.True(&f.ErrorFeature, taskNames[task.TaskName])
	}

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainAStateOf is a feature step that can be defined for a specific JobsFeature.
// It unmarshalls the response into a Job and gets the state. It checks that the state is the same as the expected one.
func (f *JobsFeature) theResponseShouldContainAStateOf(expectedState string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	assert.Equal(&f.ErrorFeature, expectedState, response.State)

	return f.ErrorFeature.StepError()
}

// iHaveGeneratedAJobInTheJobStore is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs with an empty body, which causes a default job resource to be generated.
// The newly created job resource is stored in the Job Store and also returned in the response body.
func (f *JobsFeature) iHaveGeneratedAJobInTheJobStore() error {
	// call POST /jobs
	err := f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs: %w", err)
	}
	return f.ErrorFeature.StepError()
}

// callPostJobs is a utility method that can be called by a feature step in order to call the POST jobs/ endpoint
// Calling that endpoint results in the creation of a job, in the Job Store, containing a unique id and default values.
func (f *JobsFeature) callPostJobs() error {
	var emptyBody = godog.DocString{}
	err := f.APIFeature.IPostToWithBody("/jobs", &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return nil
}

// iCallGETJobsidUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call GET /jobs/{id}.
func (f *JobsFeature) iCallGETJobsidUsingTheGeneratedID() error {
	err := f.readResponse()
	if err != nil {
		return err
	}

	err = f.GetJobByID(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the generated job id, and passes it the request body.
func (f *JobsFeature) iCallPOSTJobsidtasksUsingTheGeneratedID(body *godog.DocString) error {
	var response models.Job

	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via the PostTaskForJob, using the existing job id, and passes it the request body.
func (f *JobsFeature) iCallPOSTJobsidtasksUsingTheSameIDAgain(body *godog.DocString) error {
	err := f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}
	// make sure there's a time interval before any more tasks are posted
	time.Sleep(5 * time.Millisecond)

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasks is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks/{task name} via GetTaskForJob, using the generated job id, and passes it the task name.
func (f *JobsFeature) iCallGETJobsidtasks(taskName string) error {
	err := f.readResponse()
	if err != nil {
		return err
	}
	err = f.GetTaskForJob(f.createdJob.ID, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs/{id}/tasks via PostTaskForJob using the generated job id
func (f *JobsFeature) iCallPOSTJobsidtasksToUpdateTheNumberofdocumentsForThatTask(body *godog.DocString) error {
	err := f.PostTaskForJob(f.createdJob.ID, body)
	if err != nil {
		return fmt.Errorf("error occurred in PostTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iHaveGeneratedThreeJobsInTheJobStore is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs with an empty body, three times, which causes three default job resources to be generated.
func (f *JobsFeature) iHaveGeneratedThreeJobsInTheJobStore() error {
	// call POST /jobs three times
	err := f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs first time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs second time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs third time: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iHaveGeneratedSixJobsInTheJobStore is a feature step that can be defined for a specific JobsFeature.
// It calls POST /jobs with an empty body, three times, which causes three default job resources to be generated.
func (f *JobsFeature) iHaveGeneratedSixJobsInTheJobStore() error {
	// call POST /jobs five times
	err := f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs first time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs second time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs third time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs fourth time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs fifth time: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	err = f.callPostJobs()
	if err != nil {
		return fmt.Errorf("error occurred in callPostJobs sixth time: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iHaveCreatedATaskForTheGeneratedJob is a feature step that can be defined for a specific JobsFeature.
// It gets the job id from the response to calling POST /jobs and uses it to call POST /jobs/{job id}/tasks/{task name}
// in order to create a task for that job. It passes the taskToCreate request body to the POST endpoint.
func (f *JobsFeature) iHaveCreatedATaskForTheGeneratedJob(taskToCreate *godog.DocString) error {
	err := f.readResponse()
	if err != nil {
		return err
	}

	err = f.APIFeature.IPostToWithBody("/jobs/"+f.createdJob.ID+"/tasks", taskToCreate)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *JobsFeature) iWouldExpectThereToBeThreeOrMoreJobsReturnedInAList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound >= 3, "The list should contain three or more jobs but it only contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeTasksReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response to make sure that a list containing the expected number of tasks has been returned.
func (f *JobsFeature) iWouldExpectThereToBeTasksReturnedInAList(numTasksExpected int) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numTasksFound := len(response.TaskList)
	assert.True(&f.ErrorFeature, numTasksFound == numTasksExpected, "The list should contain "+strconv.Itoa(numTasksExpected)+
		" tasks but it contains "+strconv.Itoa(numTasksFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectThereToBeFourJobsReturnedInAList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that a list containing three or more jobs has been returned.
func (f *JobsFeature) iWouldExpectThereToBeFourJobsReturnedInAList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound == 4, "The list should contain four jobs but it contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected types of values of id,
// last_updated, links, and search_index_name.
func (f *JobsFeature) inEachJobIWouldExpectTheResponseToContainValuesThatHaveTheseStructures(table *godog.Table) error {
	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Jobs

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for j := range response.JobList {
		f.createdJob = response.JobList[j]
		err := f.checkStructure(expectedResult)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}
	return f.ErrorFeature.StepError()
}

// inEachTaskIWouldExpectIdLast_updatedAndLinksToHaveThisStructure is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that each task contains the expected types of values of job_id,
// last_updated, and links.
func (f *JobsFeature) inEachTaskIWouldExpectJobIDLastUpdatedAndLinksToHaveThisStructure(table *godog.Table) error {
	assist := assistdog.NewDefault()
	expectedResult, err := assist.ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Tasks

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for j := range response.TaskList {
		task := response.TaskList[j]
		err := f.checkTaskStructure(task.JobID, task.LastUpdated, expectedResult, task.Links, task.TaskName)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}

	return f.ErrorFeature.StepError()
}

// eachJobShouldAlsoContainTheFollowingValues is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that each job contains the expected values of
// all the remaining attributes of a job.
func (f *JobsFeature) eachJobShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Jobs

	err = json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for i := 0; i < len(response.JobList); i++ {
		f.checkValuesInJob(expectedResult, response.JobList[i])
	}

	return f.ErrorFeature.StepError()
}

// theResponseShouldContainTheNewNumberOfTasks is a feature step that can be defined for a specific JobsFeature.
// After PUT /jobs/{id}/number_of_tasks/{number_of_tasks} has been called, followed by GET /jobs/{id},
// this function checks that the job returned contains the correct number_of_tasks value.
func (f *JobsFeature) theResponseShouldContainTheNewNumberOfTasks(table *godog.Table) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Job

	_ = json.Unmarshal(f.responseBody, &response)

	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(response.NumberOfTasks))

	return f.ErrorFeature.StepError()
}

// theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that the jobs are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *JobsFeature) theJobsShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	jobList := response.JobList
	timeToCheck := jobList[0].LastUpdated

	for j := 1; j < len(jobList); j++ {
		index := strconv.Itoa(j - 1)
		nextIndex := strconv.Itoa(j)
		nextTime := jobList[j].LastUpdated
		assert.True(&f.ErrorFeature, timeToCheck.Before(nextTime),
			"The value of last_updated at job_list["+index+"] should be earlier than that at job_list["+nextIndex+"]")
		timeToCheck = nextTime
	}
	return f.ErrorFeature.StepError()
}

// theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/id/tasks to make sure that the tasks are in ascending order of their last_updated
// times i.e. the most recently updated is last in the list.
func (f *JobsFeature) theTasksShouldBeOrderedByLastupdatedWithTheOldestFirst() error {
	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	taskList := response.TaskList
	timeToCheck := taskList[0].LastUpdated

	for j := 1; j < len(taskList); j++ {
		index := strconv.Itoa(j - 1)
		nextIndex := strconv.Itoa(j)
		nextTime := taskList[j].LastUpdated
		assert.True(&f.ErrorFeature, timeToCheck.Before(nextTime),
			"The value of last_updated at taskList["+index+"] should be earlier than that at taskList["+nextIndex+"]")
		timeToCheck = nextTime
	}
	return f.ErrorFeature.StepError()
}

// noJobsHaveBeenGeneratedInTheJobStore is a feature step that can be defined for a specific JobsFeature.
// It resets the Job Store to its default values, which means that it will contain no jobs.
func (f *JobsFeature) noJobsHaveBeenGeneratedInTheJobStore() error {
	err := f.Reset(false)
	if err != nil {
		return fmt.Errorf("failed to reset the JobsFeature: %w", err)
	}
	return nil
}

// noTasksHaveBeenCreatedInTheTasksCollection is a feature step that can be defined for a specific JobsFeature.
// It resets the tasks collection to its default value, which means that it will contain no tasks.
func (f *JobsFeature) noTasksHaveBeenCreatedInTheTasksCollection() error {
	err := f.Reset(false)
	if err != nil {
		return fmt.Errorf("failed to reset the JobsFeature: %w", err)
	}
	return nil
}

// iCallGETJobsUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id} using the id passed in, which should be a valid UUID.
func (f *JobsFeature) iCallGETJobsUsingAValidUUID(id string) error {
	err := f.GetJobByID(id)
	if err != nil {
		return fmt.Errorf("error occurred in GetJobByID: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsTasksUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks/{task_name} using the id and taskName passed in, which should both be valid.
func (f *JobsFeature) iCallGETJobsTasksUsingAValidUUID(id, taskName string) error {
	err := f.GetTaskForJob(id, taskName)
	if err != nil {
		return fmt.Errorf("error occurred in GetTaskForJob: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyList is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs to make sure that an empty list (0 jobs) has been returned.
func (f *JobsFeature) iWouldExpectTheResponseToBeAnEmptyList() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Jobs
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numJobsFound := len(response.JobList)
	assert.True(&f.ErrorFeature, numJobsFound == 0, "The list should contain no jobs but it contains "+strconv.Itoa(numJobsFound))

	return f.ErrorFeature.StepError()
}

// iWouldExpectTheResponseToBeAnEmptyListOfTasks is a feature step that can be defined for a specific JobsFeature.
// It checks the response from calling GET /jobs/jobID/tasks to make sure that an empty list (0 tasks) has been returned.
func (f *JobsFeature) iWouldExpectTheResponseToBeAnEmptyListOfTasks() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Tasks
	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numTasksFound := len(response.TaskList)
	assert.True(&f.ErrorFeature, numTasksFound == 0, "The list should contain no tasks but it contains "+strconv.Itoa(numTasksFound))

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberTofTasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *JobsFeature) iCallPUTJobsidnumberTofTasksUsingTheGeneratedID(count int) error {
	countStr := strconv.Itoa(count)
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID
	err = f.PutNumberOfTasks(countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsNumberoftasksUsingAValidUUID is a feature step that can be defined for a specific JobsFeature.
// It uses the parameters passed in to call PUT /jobs/{id}/number_of_tasks/{count}
func (f *JobsFeature) iCallPUTJobsNumberoftasksUsingAValidUUID(idStr string, count int) error {
	countStr := strconv.Itoa(count)
	f.createdJob.ID = idStr

	err := f.PutNumberOfTasks(countStr)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{invalidCount}
func (f *JobsFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithAnInvalidCount(invalidCount string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PutNumberOfTasks(invalidCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount is a feature step that can be defined for a specific JobsFeature.
// It gets the id from the response body, generated in the previous step, and then uses this to call PUT /jobs/{id}/number_of_tasks/{negativeCount}
func (f *JobsFeature) iCallPUTJobsidnumberoftasksUsingTheGeneratedIDWithANegativeCount(negativeCount string) error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID

	err = f.PutNumberOfTasks(negativeCount)
	if err != nil {
		return fmt.Errorf("error occurred in PutNumberOfTasks: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theSearchReindexAPILosesItsConnectionToMongoDB is a feature step that can be defined for a specific JobsFeature.
// It loses the connection to mongo DB by setting the mongo database to an invalid setting (in the Reset function).
func (f *JobsFeature) theSearchReindexAPILosesItsConnectionToMongoDB() error {
	err := f.Reset(true)
	if err != nil {
		return fmt.Errorf("failed to reset the JobsFeature: %w", err)
	}
	return nil
}

// iCallGETJobsidtasksUsingTheSameIDAgain is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{id}/tasks using the existing value of id.
func (f *JobsFeature) iCallGETJobsidtasksUsingTheSameIDAgain() error {
	// call GET /jobs/{id}/tasks
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iCallGETJobsidtasksoffsetLimit is a feature step that can be defined for a specific JobsFeature.
// It calls GET /jobs/{id}/tasks?offset={offset}&limit={limit} using the existing value of id.
func (f *JobsFeature) iCallGETJobsidtasksoffsetLimit(offset, limit string) error {
	// call GET /jobs/{id}/tasks?offset={offset}&limit={limit}
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks?offset=" + offset + "&limit=" + limit)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// iGETJobsTasks is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{jobID}/tasks using the existing value of id as the jobID value.
func (f *JobsFeature) iGETJobsTasks() error {
	// call GET /jobs/{jobID}/tasks
	err := f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}
	return f.ErrorFeature.StepError()
}

// iGETJobsidtasksUsingTheGeneratedID is a feature step that can be defined for a specific JobsFeature.
// It calls /jobs/{jobID}/tasks using the response.ID, from the previously returned Job, as the id value.
func (f *JobsFeature) iGETJobsidtasksUsingTheGeneratedID() error {
	f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)

	var response models.Job

	err := json.Unmarshal(f.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	f.createdJob.ID = response.ID
	err = f.APIFeature.IGet("/jobs/" + f.createdJob.ID + "/tasks")
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}

	return f.ErrorFeature.StepError()
}

// theSearchReindexAPILosesItsConnectionToTheSearchAPI is a feature step that can be defined for a specific JobsFeature.
// It closes the connection to the search feature so as to mimic losing the connection to the Search API.
func (f *JobsFeature) theSearchReindexAPILosesItsConnectionToTheSearchAPI() error {
	f.SearchFeature.Close()
	return f.ErrorFeature.StepError()
}

// theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName is a feature step that can be defined for a specific JobsFeature.
// It asserts that the job id and search index name that get returned by the POST /jobs endpoint match the ones that get sent in the
// reindex-requested event
func (f *JobsFeature) theReindexrequestedEventShouldContainTheExpectedJobIDAndSearchIndexName() error {
	reindexRequestedData, err := readAndDeserializeKafkaProducerOutput(f.kafkaProducerOutputData)
	if err != nil {
		return err
	}
	assert.Equal(&f.ErrorFeature, f.createdJob.ID, reindexRequestedData.JobID)
	assert.Equal(&f.ErrorFeature, f.createdJob.SearchIndexName, reindexRequestedData.SearchIndex)
	return nil
}

// GetJobByID is a utility function that is used for calling the GET /jobs/{id} endpoint.
// It checks that the id string is a valid UUID before calling the endpoint.
func (f *JobsFeature) GetJobByID(id string) error {
	_, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call GET /jobs/{id}
	err = f.APIFeature.IGet("/jobs/" + id)
	if err != nil {
		return fmt.Errorf("error occurred in IGet: %w", err)
	}
	return nil
}

// PutNumberOfTasks is a utility function that is used for calling the PUT /jobs/{id}/number_of_tasks/{count}
// It checks that the id string is a valid UUID before calling the endpoint.
func (f *JobsFeature) PutNumberOfTasks(countStr string) error {
	var emptyBody = godog.DocString{}
	_, err := uuid.FromString(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call PUT /jobs/{id}/number_of_tasks/{count}
	err = f.APIFeature.IPut("/jobs/"+f.createdJob.ID+"/number_of_tasks/"+countStr, &emptyBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPut: %w", err)
	}
	return nil
}

// PostTaskForJob is a utility function that is used for calling POST /jobs/{id}/tasks
// The endpoint requires authorisation and a request body.
func (f *JobsFeature) PostTaskForJob(jobID string, requestBody *godog.DocString) error {
	_, err := uuid.FromString(jobID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	// call POST /jobs/{id}/tasks
	err = f.APIFeature.IPostToWithBody("/jobs/"+jobID+"/tasks", requestBody)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}
	return nil
}

// GetTaskForJob is a utility function that is used for calling GET /jobs/{id}/tasks/{task name}
func (f *JobsFeature) GetTaskForJob(jobID, taskName string) error {
	_, err := uuid.FromString(jobID)
	if err != nil {
		return fmt.Errorf("the job id should be a uuid: %w", err)
	}

	// call GET /jobs/{jobID}/tasks/{taskName}
	err = f.APIFeature.IGet("/jobs/" + jobID + "/tasks/" + taskName)
	if err != nil {
		return fmt.Errorf("error occurred in IPostToWithBody: %w", err)
	}
	return nil
}

// checkStructure is a utility function that can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *JobsFeature) checkStructure(expectedResult map[string]string) error {
	_, err := uuid.FromString(f.createdJob.ID)
	if err != nil {
		return fmt.Errorf("the id should be a uuid: %w", err)
	}

	if f.createdJob.LastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + f.createdJob.LastUpdated.String())
	}

	expectedLinksTasks := strings.Replace(expectedResult["links: tasks"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksTasks = strings.Replace(expectedLinksTasks, "{id}", f.createdJob.ID, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksTasks, f.createdJob.Links.Tasks)

	expectedLinksSelf := strings.Replace(expectedResult["links: self"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{id}", f.createdJob.ID, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksSelf, f.createdJob.Links.Self)

	re := regexp.MustCompile(`(ons)(\d*)`)
	wordWithExpectedPattern := re.FindString(f.createdJob.SearchIndexName)
	assert.Equal(&f.ErrorFeature, wordWithExpectedPattern, f.createdJob.SearchIndexName)

	return nil
}

// checkTaskStructure is a utility function that can be called by a feature step to assert that a job contains the expected structure in its values of
// id, last_updated, and links. It confirms that last_updated is a current or past time, and that the tasks and self links have the correct paths.
func (f *JobsFeature) checkTaskStructure(id string, lastUpdated time.Time, expectedResult map[string]string, links *models.TaskLinks, taskName string) error {
	_, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("the jobID should be a uuid: %w", err)
	}

	if lastUpdated.After(time.Now()) {
		return errors.New("expected LastUpdated to be now or earlier but it was: " + lastUpdated.String())
	}

	expectedLinksJob := strings.Replace(expectedResult["links: job"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksJob = strings.Replace(expectedLinksJob, "{id}", id, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksJob, links.Job)

	expectedLinksSelf := strings.Replace(expectedResult["links: self"], "{bind_address}", f.Config.BindAddr, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{id}", id, 1)
	expectedLinksSelf = strings.Replace(expectedLinksSelf, "{task_name}", taskName, 1)

	assert.Equal(&f.ErrorFeature, expectedLinksSelf, links.Self)
	return nil
}

// checkValuesInJob is a utility function that can be called by a feature step in order to check that the values
// of certain attributes, in a job, are all equal to the expected ones.
func (f *JobsFeature) checkValuesInJob(expectedResult map[string]string, job models.Job) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_tasks"], strconv.Itoa(job.NumberOfTasks))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_completed"], job.ReindexCompleted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_failed"], job.ReindexFailed.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["reindex_started"], job.ReindexStarted.Format(time.RFC3339))
	assert.Equal(&f.ErrorFeature, expectedResult["state"], job.State)
	assert.Equal(&f.ErrorFeature, expectedResult["total_search_documents"], strconv.Itoa(job.TotalSearchDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["total_inserted_search_documents"], strconv.Itoa(job.TotalInsertedSearchDocuments))
}

// checkValuesInTask is a utility function that can be called by a feature step in order to check that the values
// of certain attributes, in a task, are all equal to the expected ones.
func (f *JobsFeature) checkValuesInTask(expectedResult map[string]string, task models.Task) {
	assert.Equal(&f.ErrorFeature, expectedResult["number_of_documents"], strconv.Itoa(task.NumberOfDocuments))
	assert.Equal(&f.ErrorFeature, expectedResult["task_name"], task.TaskName)
}

// funcCheck is used to mock the kafka producer CheckerFunc
func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}

// readOutputMessages is a utility method to read the kafka messages that get sent to the producer's output channel
func (f *JobsFeature) readOutputMessages() {
	go func() {
		for {
			select {
			case f.kafkaProducerOutputData <- <-f.MessageProducer.Channels().Output:
				log.Println("read")
			case <-f.quitReadingOutput:
				return
			}
		}
	}()
}

func readAndDeserializeKafkaProducerOutput(kafkaProducerOutputData <-chan []byte) (*models.ReindexRequested, error) {
	for reindexRequestedDataBytes := range kafkaProducerOutputData {
		reindexRequestedData := &models.ReindexRequested{}
		err := schema.ReindexRequestedEvent.Unmarshal(reindexRequestedDataBytes, reindexRequestedData)
		return reindexRequestedData, err
	}
	return nil, nil
}

// readResponse is a utility method to read the JSON response that gets returned by the POST /job endpoint
func (f *JobsFeature) readResponse() error {
	if (f.createdJob == models.Job{}) {
		f.responseBody, _ = io.ReadAll(f.APIFeature.HttpResponse.Body)
		response := models.Job{}
		err := json.Unmarshal(f.responseBody, &response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json response: %w", err)
		}

		f.createdJob = response
		return err
	}
	return nil
}
