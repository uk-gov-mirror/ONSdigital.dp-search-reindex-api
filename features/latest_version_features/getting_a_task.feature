Feature: Getting a task

  Scenario: Task exists in the tasks collection and a get request returns it successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I have created a task for the generated job
    """
    { "task_name": "dataset-api", "number_of_documents": 30 }
    """
    When I call GET /jobs/{id}/tasks/{"dataset-api"}
    Then the HTTP status code should be "200"
    And the response for getting task to look like this
      | job_id              | UUID                                                |
      | last_updated        | Not in the future                                   |
      | links: self         | {host}/{latest_version}/jobs/{id}/tasks/dataset-api |
      | links: job          | {host}/{latest_version}/jobs/{id}                   |
      | number_of_documents | 30                                                  |
      | task_name           | dataset-api                                         |

  Scenario: Job does not exist in the Job Store and a get task for job id request returns StatusNotFound

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/tasks/{"dataset-api"} using a valid UUID
    Then the HTTP status code should be "404"

  Scenario: Task does not exist in the tasks collection and a get task for job id request returns StatusNotFound

    Given no tasks have been created in the tasks collection
    And the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call GET /jobs/{id}/tasks/{"dataset-api"}
    Then the HTTP status code should be "404"

  Scenario: The connection to mongo DB is lost and a get request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    And the api version is undefined for incoming requests
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/tasks/{"dataset-api"} using a valid UUID
    Then the HTTP status code should be "500"
