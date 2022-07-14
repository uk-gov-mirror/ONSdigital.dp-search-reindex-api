Feature: Getting a list of tasks

  Scenario: Three Tasks exist in the Data Store and a get request returns them successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    And I set the If-Match header to a valid e-tag to get tasks
    When I call GET /search-reindex-jobs/{id}/tasks using the same id again
    Then I would expect there to be 3 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/search-reindex-jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/search-reindex-jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with no If-Match header ignores the ETag check and returns tasks successfully 

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /search-reindex-jobs/{id}/tasks using the same id again
    Then I would expect there to be 3 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/search-reindex-jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/search-reindex-jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with empty If-Match header ignores the ETag check and returns tasks successfully 

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    And I set the "If-Match" header to ""
    When I call GET /search-reindex-jobs/{id}/tasks using the same id again
    Then I would expect there to be 3 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/search-reindex-jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/search-reindex-jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with If-Match set to `*` ignores the ETag check and returns tasks successfully 

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    And I set the "If-Match" header to "*"
    When I call GET /search-reindex-jobs/{id}/tasks using the same id again
    Then I would expect there to be 3 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/search-reindex-jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/search-reindex-jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with outdated or invalid etag returns an conflict error 

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I set the "If-Match" header to "invalid"
    When I call GET /search-reindex-jobs/{id}/tasks using the same id again
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of resource
    """ 
    And the response header "E-Tag" should be ""

  Scenario: No Tasks exist in the Data Store and a get request returns an empty list

    Given no tasks have been created in the tasks collection
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I GET /search-reindex-jobs/{id}/tasks using the generated id
    Then I would expect the response to be an empty list of tasks
    And the HTTP status code should be "200"

  Scenario: Four tasks exist and a get request with offset and limit correctly returns two

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name4", "number_of_documents": 4 }
    """
    When I call GET /search-reindex-jobs/{id}/tasks?offset="1"&limit="2"
    Then I would expect there to be 2 tasks returned in a list
    And the response for getting task to look like this
      | job_id       | UUID                                                |
      | last_updated | Not in the future                                   |
      | links: self  | {host}/{latest_version}/search-reindex-jobs/{id}/tasks/{task_name} |
      | links: job   | {host}/{latest_version}/search-reindex-jobs/{id}                   |
    And each task should also contain the following values:
      | number_of_documents | 4                                            |
      | task_name           | {task_name}                                  |
    And the tasks should be ordered, by last_updated, with the oldest first

  Scenario: Three tasks exist and a get request with negative offset returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /search-reindex-jobs/{id}/tasks?offset="-2"&limit="20"
    Then the HTTP status code should be "400"

  Scenario: Three tasks exist and a get request with negative limit returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /search-reindex-jobs/{id}/tasks?offset="0"&limit="-3"
    Then the HTTP status code should be "400"

  Scenario: Three tasks exist and a get request with limit greater than the maximum returns an error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I call POST /search-reindex-jobs/{id}/tasks using the generated id
    """
    { "task_name": "zebedee", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "dataset-api", "number_of_documents": 4 }
    """
    And I call POST /search-reindex-jobs/{id}/tasks using the same id again
    """
    { "task_name": "another-task-name3", "number_of_documents": 4 }
    """
    When I call GET /search-reindex-jobs/{id}/tasks?offset="0"&limit="1001"
    Then the HTTP status code should be "400"

  Scenario: Job does not exist and a get request returns StatusNotFound

    Given the number of existing jobs in the Job Store is 0
    And the api version is v1 for incoming requests
    When I GET "/search-reindex-jobs/a219584a-454a-4add-92c6-170359b0ee77/tasks"
    Then the HTTP status code should be "404"
