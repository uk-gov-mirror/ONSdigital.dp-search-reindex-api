Feature: Getting a task

  Scenario: Task exists in the tasks collection and a get request returns it successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
    And I have created a task for the generated job
    """
    { "task_name": "dp-api-router", "number_of_documents": 30 }
    """
    When I call GET /jobs/{id}/tasks/dp-api-router to get the task
    Then the HTTP status code should be "200"
    And I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/dp-api-router |
      | links: job   | http://{bind_address}/jobs/{id}                     |
    And the task resource should also contain the following values:
      | number_of_documents               | 30                             |
      | task_name                         | dp-api-router                  |

  Scenario: Job does not exist in the Job Store and a get task for job id request returns StatusNotFound

    Given no jobs have been generated in the Job Store
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/tasks/dp-api-router using a valid UUID
    Then the HTTP status code should be "404"

  Scenario: Task does not exist in the tasks collection and a get task for job id request returns StatusNotFound

    Given no tasks have been created in the tasks collection
    And I have generated a job in the Job Store
    When I call GET /jobs/{id}/tasks/dp-api-router to get the task
    Then the HTTP status code should be "404"

  Scenario: The connection to mongo DB is lost and a get request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/tasks/dp-api-router using a valid UUID
    Then the HTTP status code should be "500"
