Feature: Updating the number of tasks for a particular job

  Scenario: Job exists in the Job Store and a put request updates its number of tasks successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    And I set the If-Match header to the generated job e-tag
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "204"
    And the response ETag header should not be empty
    
    Given I set the "If-Match" header to "*"
    And I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Request is made with invalid service auth token

    Given I use a service auth token "invalidServiceAuthToken"
    And zebedee does not recognise the service auth token
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    And I set the If-Match header to the generated job e-tag
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "401"
    And I should receive the following response:
    """
      error making get permissions request: unauthorized
    """ 
    And the response header "Content-Type" should be ""
    And the response header "E-Tag" should be ""
  
  Scenario: Request made with no If-Match header ignores the ETag check and updates number of tasks of job successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "204"
    And the response ETag header should not be empty
    And I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Request made with empty If-Match header ignores the ETag check and updates its number of tasks of job successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    And I set the "If-Match" header to ""
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "204"
    And the response ETag header should not be empty
    And I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Request made with If-Match set to `*` ignores the ETag check and updates its number of tasks of job successfully

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    And I set the "If-Match" header to "*"
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "204"
    And the response ETag header should not be empty
    And I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Request made with outdated or invalid etag returns an conflict error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    And I set the "If-Match" header to "invalid"
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of resource
    """ 
    And the response header "E-Tag" should be ""

  Scenario: Job does not exist in the Job Store and a put request, to update its number of tasks, returns StatusNotFound

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 0
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/number_of_tasks/{7} using a valid UUID
    Then the HTTP status code should be "404"
    And I should receive the following response:
    """
      failed to find the specified reindex job
    """ 
    And the response header "E-Tag" should be ""

  Scenario: A put request fails to update the number of tasks because it contains an invalid value of count

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{"seven"} using the generated id with an invalid count
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      number of tasks must be a positive integer
    """ 
    And the response header "E-Tag" should be ""

  Scenario: A put request fails to update the number of tasks because it contains a negative value of count

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{"-7"} using the generated id with a negative count
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      number of tasks must be a positive integer
    """ 
    And the response header "E-Tag" should be ""

  Scenario: Request results in no modification to job resource and returns an unmodified error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the number of existing jobs in the Job Store is 1
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{0} using the generated id
    Then the HTTP status code should be "304"
    And I should receive the following response:
    """
      no modification made on resource
    """ 
    And the response header "E-Tag" should be ""

  Scenario: The connection to mongo DB is lost and a put request returns an internal server error

    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search reindex api loses its connection to mongo DB
    And the api version is v1 for incoming requests
    When I call PUT /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/number_of_tasks/{7} using a valid UUID
    Then the HTTP status code should be "500"
    And I should receive the following response:
    """
      internal server error
    """ 
    And the response header "E-Tag" should be ""
