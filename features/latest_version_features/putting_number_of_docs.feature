Feature: Updating the number of docs for a particular task
  Scenario: Job exists in the Job Store and a put request updates its number of docs successfully
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And I have created a task for the generated job
    """
    { "task_name": "dataset-api", "number_of_documents": 7 }
    """
    And the api version is v1 for incoming requests
    And I set the If-Match header to the generated task e-tag
    When I call PUT /search-reindex-jobs/{id}/tasks/{task_name}/number-of-documents/{8} using the generated id
    Then the HTTP status code should be "204"
    And the response ETag header should not be empty

  Scenario: Job exists in the Job Store and returns not modified status
     Given I use a service auth token "validServiceAuthToken"
     And zebedee recognises the service auth token as valid
     And the number of existing jobs in the Job Store is 1
     And I have created a task for the generated job
     """
     { "task_name": "dataset-api", "number_of_documents": 7 }
     """
     And the api version is v1 for incoming requests
     And I set the If-Match header to the generated task e-tag
     When I call PUT /search-reindex-jobs/{id}/tasks/{task_name}/number-of-documents/{7} using the generated id
     Then the HTTP status code should be "304"

  Scenario: Job exists in the Job Store and returns not modified status
      Given I use a service auth token "validServiceAuthToken"
      And zebedee recognises the service auth token as valid
      And the number of existing jobs in the Job Store is 1
      And I have created a task for the generated job
      """
      { "task_name": "dataset-api", "number_of_documents": 7 }
      """
      And the api version is v1 for incoming requests
      And I set the "If-Match" header to "invalid"
      When I call PUT /search-reindex-jobs/{id}/tasks/{task_name}/number-of-documents/{7} using the generated id
      Then the HTTP status code should be "409"