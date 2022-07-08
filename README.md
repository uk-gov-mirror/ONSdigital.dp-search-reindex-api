dp-search-reindex-api
=====================
Provides detail about search reindex jobs and enables creation and running of them

### Getting started

* Set up dependencies locally as follows:

In dp-compose repo run `docker-compose up -d` to run MongoDB on port 27017. 

NB. The above command will also run Site Wide ElasticSearch, on port 11200, which is required by the Search API.

Run `vault server -dev` (this is required by Zebedee)

In the zebedee repo run `./run.sh` to run Zebedee

In the dp-search-api repo set the ELASTIC_SEARCH_URL environment variable as follows (to use the Site Wide ElasticSearch):

`export ELASTIC_SEARCH_URL="http://localhost:11200"`

Also in the dp-search-api repo run `make debug`

Make sure that you have a valid local SERVICE_AUTH_TOKEN environment variable value;
if not then set one up by following these instructions: https://github.com/ONSdigital/zebedee

* Then in the dp-search-reindex-api repo run `make debug`

### Dependencies

* Requires MongoDB running on port 27017
* Requires Kafka running on port 9092  
* Requires Zebedee running on port 8082
* Requires Search API running on port 23900  
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable          | Default               | Description
| ------------------------------| --------------------- | -----------
| BIND_ADDR                     | localhost:25700       | The host and port to bind to (The http:// scheme prefix is added programmatically)
| DEFAULT_LIMIT                 | 20                    | The default number of items to be returned from a list endpoint
| DEFAULT_MAXIMUM_LIMIT         | 1000                  | The maximum number of items to be returned in any list endpoint (to prevent performance issues)
| DEFAULT_OFFSET                | 0                     | The number of items into the full list (i.e. the 0-based index) that a particular response is starting at
| GRACEFUL_SHUTDOWN_TIMEOUT     | 20s                   | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT  | 90s                   | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| HEALTHCHECK_INTERVAL          | 30s                   | Time between self-healthchecks (`time.Duration` format)
| KAFKA_ADDR                    | localhost:39092       | The kafka broker addresses (can be comma separated)
| KAFKA_REINDEX_REQUESTED_TOPIC | reindex-requested     | The name of the topic to produce messages for
| KAFKA_SEC_CA_CERTS            | _unset_               | CA cert chain for the server cert [[1]](#note1)
| KAFKA_SEC_CLIENT_CERT         | _unset_               | PEM for the client certificate [[1]](#note1)
| KAFKA_SEC_CLIENT_KEY          | _unset_               | PEM for the client key [[1]](#note1)
| KAFKA_SEC_PROTO               | _unset_               | if set to `TLS`, kafka connections will use TLS [[1]](#note1)
| KAFKA_SEC_SKIP_VERIFY         | false                 | ignores server certificate issues if `true` [[1]](#note1)
| KAFKA_VERSION                 | 1.0.2                 | The kafka version that this service expects to connect to
| LATEST_VERSION                | v1                    | The latest version of the Search Reindex API
| MAX_REINDEX_JOB_RUNTIME       | 3600s                 | The maximum amount of time that a reindex job is allowed to run before another reindex job can be started
| MONGODB_BIND_ADDR             | localhost:27017       | The MongoDB bind address (aka the cluster endpoint)
| MONGODB_CERT_CHAIN            | _unset_               | CA cert chain for the server cert
| MONGODB_COLLECTIONS           | JobsCollection: "jobs", LocksCollection: "jobs_locks", TasksCollection: "tasks" | The MongoDB collections
| MONGODB_CONNECT_TIMEOUT       | 5s                    | The timeout when connecting to MongoDB (`time.Duration` format)
| MONGODB_DATABASE              | search                | The MongoDB search database
| MONGODB_ENABLE_READ_CONCERN   | false                 | Switch to use (or not) majority read concern
| MONGODB_ENABLE_WRITE_CONCERN  | true                  | Switch to use (or not) majority write concern
| MONGODB_IS_SSL                | false                 | Switch to use (or not) TLS when connecting to mongodb
| MONGODB_PASSWORD              | _unset_               | The MongoDB Password
| MONGODB_QUERY_TIMEOUT         | 15s                   | The timeout for querying MongoDB (`time.Duration` format)
| MONGODB_REPLICA_SET           | _unset_               | The name of the MongoDB replica set
| MONGODB_USERNAME              | _unset_               | The MongoDB Username
| MONGODB_VERIFY_CERT           | false                 | Switch for whether the Mongo server certificate is to be validated or not (a major security breach not doing so)
| SEARCH_API_URL                | http://localhost:23900| The URL to the Search API (for creating new ElasticSearch indexes)
| SERVICE_AUTH_TOKEN            | _unset_               | This is required to identify the Search Reindex API when it calls the Search API POST /search endpoint
| TASK_NAME_VALUES              | dataset-api,zebedee   | The list of permissible values that can be used for the task_name when creating a new task for a reindex job
| ZEBEDEE_URL                   | http://localhost:8082 | The URL to Zebedee (for authorisation)

**Notes:**

<a name="note1"></a> 1. For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)

### Testing

* Run the component tests with this command `go test -component`
* Run the unit tests with this command `make test`
* For all details of the service endpoints use a swagger editor [such as this one](https://editor.swagger.io/) to view the [swagger specification](swagger.yaml)

When running the service (see 'Getting Started') then one can use command line tool (cURL) or REST API client (e.g. [Postman](https://www.postman.com/product/rest-client/)) to test the endpoints:
- POST: http://localhost:25700/jobs (should post a default job into the data store and return a JSON representation of it, should also create a new ElasticSearch index)
- GET: http://localhost:25700/jobs/ID NB. Use the id returned by the POST call above,e.g., http://localhost:25700/jobs/bc7b87de-abf5-45c5-8e3c-e2a575cab28a (should get a job from the data store)
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
- GET: http://localhost:25700/jobs/ID/tasks (should get all the tasks, for a particular job, from the data store)
- PATCH: http://localhost:25700/jobs/41f72483-aeee-4693-9fed-a45ebaa370f2 -H 'Authorization: Bearer fc4089e2e12937861377629b0cd96cf79298a4c5d329a2ebb96664c88df77b67' -H 'If-Match: *' -H 'Content-Type: application/json' -d '[{"op":"replace","path":"/state","value":"created"},{"op":"replace","path":"/total_search_documents","value":208},{"op":"replace","path":"/number_of_tasks","value":2}]'

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
