Feature: Posting a job

  Scenario: Task is created successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                           |
      | last_updated | Not in the future                              |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/dataset-api |
      | links: job   | http://{bind_address}/jobs/{id}                |
    And the task resource should also contain the following values:
      | number_of_documents               | 29                        |
      | task_name                         | dataset-api                  |
    And the HTTP status code should be "201"

  Scenario: The connection to mongo DB is lost and a post request returns an internal server error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
    And the search reindex api loses its connection to mongo DB
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "500"

  Scenario: Task is updated successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    And a new task resource is created containing the following values:
      | number_of_documents               | 29                        |
      | task_name                         | dataset-api                  |
    When I call POST /jobs/{id}/tasks to update the number_of_documents for that task
    """
    { "task_name": "dataset-api", "number_of_documents": 36 }
    """
    Then I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                           |
      | last_updated | Not in the future                              |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/dataset-api |
      | links: job   | http://{bind_address}/jobs/{id}                |
    And the task resource should also contain the following values:
      | number_of_documents               | 36                        |
      | task_name                         | dataset-api                  |
    And the HTTP status code should be "201"

  Scenario: Request body cannot be read returns a bad request error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And I have generated a job in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "400"

  Scenario: Job does not exist and an attempt to create a task for it returns a not found error
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And no jobs have been generated in the Job Store
    And I POST "/jobs/any-job-id/tasks"
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "404"

  Scenario: No authorisation header set returns a bad request error

    Given I am not authorised
    And I have generated a job in the Job Store
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "400"

  Scenario: Invalid service auth token returns unauthorised error

    Given I use a service auth token "invalidServiceAuthToken"
    And zebedee does not recognise the service auth token
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "401"

  Scenario: Valid user token returns bad request error

    Given I use an X Florence user token "validXFlorenceToken"
    And I am identified as "someone@somewhere.com"
    And zebedee recognises the user token as valid
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "400"

  Scenario: Invalid user token returns unauthorised error

    Given I use an X Florence user token "invalidXFlorenceToken"
    And I am not identified by zebedee
    And zebedee does not recognise the user token
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "401"
