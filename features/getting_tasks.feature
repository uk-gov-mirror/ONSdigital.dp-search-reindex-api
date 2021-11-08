Feature: Getting a list of tasks

  Scenario: Three Tasks exist in the Data Store and a get request returns them successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
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
    And in each task I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                              |
      | last_updated | Not in the future                                 |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/{task_name} |
      | links: job   | http://{bind_address}/jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                          |
      | task_name           | {task_name}                                |
    And the tasks should be ordered, by last_updated, with the oldest first

  Scenario: No Tasks exist in the Data Store and a get request returns an empty list

    Given no tasks have been created in the tasks collection
    And I have generated a job in the Job Store
    When I GET /jobs/{id}/tasks using the generated id
    Then I would expect the response to be an empty list of tasks
    And the HTTP status code should be "200"

  Scenario: Four tasks exist and a get request with offset and limit correctly returns two

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
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
    And in each task I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                              |
      | last_updated | Not in the future                                 |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/{task_name} |
      | links: job   | http://{bind_address}/jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                          |
      | task_name           | {task_name}                                |
    And the tasks should be ordered, by last_updated, with the oldest first

  Scenario: Three tasks exist and a get request with negative offset returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
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

  Scenario: Three tasks exist and a get request with offset greater than three returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
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
    When I call GET /jobs/{id}/tasks?offset="4"&limit="20"
    Then the HTTP status code should be "400"

  Scenario: Three tasks exist and a get request with negative limit returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
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
    And I have generated a job in the Job Store
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
