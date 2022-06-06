Feature: Posting a task

  Scenario: Task is created successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "201"
    And check the database has the following task document stored in task collection
      | job_id              | UUID                         |
      | last_updated        | Not in the future            |
      | links: self         | /jobs/{id}/tasks/dataset-api |
      | links: job          | /jobs/{id}                   |
      | number_of_documents | 29                           |
      | task_name           | dataset-api                  |
    And the response ETag header should not be empty

  Scenario: Task is updated successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    And a new task resource is created containing the following values:
      | number_of_documents               | 29          |
      | task_name                         | dataset-api |
    When I call POST /jobs/{id}/tasks to update the number_of_documents for that task
    """
    { "task_name": "dataset-api", "number_of_documents": 36 }
    """
    Then the HTTP status code should be "201"
    And check the database has the following task document stored in task collection
      | job_id              | UUID                         |
      | last_updated        | Not in the future            |
      | links: self         | /jobs/{id}/tasks/dataset-api |
      | links: job          | /jobs/{id}                   |
      | number_of_documents | 36                           |
      | task_name           | dataset-api                  |
    And the response ETag header should not be empty

  Scenario: Task is created successfully when provided with valid user token

    Given I use an X Florence user token "validXFlorenceToken"
    And I am identified as "someone@somewhere.com"
    And zebedee recognises the user token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "201"
    And the response ETag header should not be empty
  
  Scenario: No authorisation header set returns a bad request error

    Given I am not authorised
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""

  Scenario: Invalid service auth token returns unauthorised error

    Given I use a service auth token "invalidServiceAuthToken"
    And zebedee does not recognise the service auth token
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "401"
    And the response header "E-Tag" should be ""

  Scenario: Invalid user token returns unauthorised error

    Given I use an X Florence user token "invalidXFlorenceToken"
    And I am not identified by zebedee
    And zebedee does not recognise the user token
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "401"
    And the response header "E-Tag" should be ""

  Scenario: Invalid request body returns a bad request error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""
  
  Scenario: The connection to mongo DB is lost and a post request returns an internal server error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And the search reindex api loses its connection to mongo DB
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "500"
    And the response header "E-Tag" should be ""

  Scenario: Job does not exist and an attempt to create a task for it returns a not found error
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 0
    And I POST "/jobs/any-job-id/tasks"
    """
    { "task_name": "dataset-api", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "404"
    And the response header "E-Tag" should be ""

  Scenario: Invalid task name returns bad request error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /jobs/{id}/tasks using the generated id
    """
    { "task_name": "florence", "number_of_documents": 29 }
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""
