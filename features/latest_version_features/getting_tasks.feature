Feature: Getting a list of tasks

  Scenario: Three Tasks exist in the Data Store and a get request returns them successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /jobs/{id}/tasks using the same id again
    Then I would expect there to be 3 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first

  Scenario: No Tasks exist in the Data Store and a get request returns an empty list

    Given no tasks have been created in the tasks collection
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    When I GET /jobs/{id}/tasks using the generated id
    Then I would expect the response to be an empty list of tasks
    And the HTTP status code should be "200"

  Scenario: Four tasks exist and a get request with offset and limit correctly returns two

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name4", "number_of_documents": 4 }
    """
    When I call GET /jobs/{id}/tasks?offset="1"&limit="2"
    Then I would expect there to be 2 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first

  Scenario: Three tasks exist and a get request with negative offset returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /jobs/{id}/tasks?offset="-2"&limit="20"
    Then the HTTP status code should be "400"

  Scenario: Three tasks exist and a get request with negative limit returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /jobs/{id}/tasks?offset="0"&limit="-3"
    Then the HTTP status code should be "400"

  Scenario: Three tasks exist and a get request with limit greater than the maximum returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /jobs/{id}/tasks?offset="0"&limit="1001"
    Then the HTTP status code should be "400"

  Scenario: Job does not exist and a get request returns StatusNotFound

    Given I have generated 0 jobs in the Job Store
    And set the api version to undefined for incoming requests
    When I GET "/jobs/a219584a-454a-4add-92c6-170359b0ee77/tasks"
    Then the HTTP status code should be "404"
