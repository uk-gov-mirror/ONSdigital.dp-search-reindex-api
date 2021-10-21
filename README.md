dp-search-reindex-api
=====================
Provides detail about search reindex jobs and enables creation and running of them

### Getting started

* Set up dependencies locally as follows:

In dp-compose run `docker-compose up -d` to run MongoDB on port 27017

In any directory run `vault server -dev` as Zebedee has a dependency on Vault

In the zebedee directory run `./run.sh` to run Zebedee

* Then in the dp-search-reindex-api run `make debug`

### Dependencies

* Requires MongoDB running on port 27017 test
* Requires Zebedee running on port 8082
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default               | Description
| ---------------------------- | --------------------- | -----------
| BIND_ADDR                    | localhost:25700       | The host and port to bind to (The http:// scheme prefix is added programmatically)
| GRACEFUL_SHUTDOWN_TIMEOUT    | 20s                   | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                   | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                   | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| MAX_REINDEX_JOB_RUNTIME      | 3600s                 | The maximum amount of time that a reindex job is allowed to run before another reindex job can be started
| MONGODB_BIND_ADDR            | localhost:27017       | The MongoDB bind address
| MONGODB_JOBS_COLLECTION      | jobs                  | MongoDB jobs collection
| MONGODB_LOCKS_COLLECTION     | jobs_locks            | MongoDB locks collection
| MONGODB_TASKS_COLLECTION     | tasks                 | MongoDB tasks collection
| MONGODB_DATABASE             | search                | The MongoDB search database
| DEFAULT_MAXIMUM_LIMIT        | 1000                  | The maximum number of items to be returned in any list endpoint (to prevent performance issues)
| DEFAULT_LIMIT                | 20                    | The default number of items to be returned from a list endpoint
| DEFAULT_OFFSET               | 0                     | The number of items into the full list (i.e. the 0-based index) that a particular response is starting at
| ZEBEDEE_URL                  | http://localhost:8082 | The URL to Zebedee (for authorisation)

### Testing

* Run the component tests with this command `go test -component`
* Run the unit tests with this command `make test`
* For all details of the service endpoints use a swagger editor [such as this one](https://editor.swagger.io/) to view the [swagger specification](swagger.yaml)

When running the service (see 'Getting Started') then one can use command line tool (cURL) or REST API client (e.g. [Postman](https://www.postman.com/product/rest-client/)) to test the endpoints:
- POST: http://localhost:25700/jobs (should post a default job into the data store and return a JSON representation of it)
- GET: http://localhost:25700/jobs/ID NB. Use the id returned by the POST call above e.g. http://localhost:25700/jobs/bc7b87de-abf5-45c5-8e3c-e2a575cab28a (should get a job from the data store)
- GET: http://localhost:25700/jobs (should get all the jobs from the data store)
- GET: http://localhost:25700/jobs?offset=1&limit=2 (should get no more than 2 jobs, from the data store, starting from index 1)
- PUT: http://localhost:25700/jobs/ID/number_of_tasks/10 (should put a value of 10 in the number_of_tasks field for the job with that particular id)
- GET: http://localhost:25700/health (should show the health check details)
- POST: http://localhost:25700/jobs/ID/tasks (should post a task into the data store and return a JSON representation of it) NB. This endpoint requires authorisation. Choose 'Bearer Token' as the authorisation type in Postman. The token to use can be found by this command `echo $SERVICE_AUTH_TOKEN`
The endpoint also requires a body, which should contain the task name, and the number of documents e.g.
`{
"task_name": "dataset-api",
"number_of_documents": 29
}`
- GET: http://localhost:25700/jobs/ID/tasks/TASK_NAME (should get a task from the data store)

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
