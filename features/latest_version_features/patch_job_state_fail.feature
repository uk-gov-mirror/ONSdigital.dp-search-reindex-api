Feature: Patch job state - Failure
  
  Scenario: Request is made with invalid service auth token

    Given I use a service auth token "invalidServiceAuthToken"
    And zebedee does not recognise the service auth token
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "401"
    And I should receive the following response:
    """
      error making get permissions request: unauthorized
    """ 
    And the response header "Content-Type" should be ""
    And the response header "E-Tag" should be ""

  Scenario: Request is made with invalid or outdated etag in If-Match header
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the "If-Match" header to "invalid"
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of job resource
    """ 
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""
  
  Scenario: Request is made with no modification to job resource
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the "If-Match" header to the old e-tag

    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "304"
    And I should receive the following response:
    """
      new etag is same as existing etag
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with invalid job ID
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I PATCH "/jobs/invalid"
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "404"
    And I should receive the following response:
    """
      the job id could not be found in the jobs collection
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with empty patch body
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      empty request body given
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with invalid patch body
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      {"test": "invalid"}
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      patch operation is missing or invalid. Please, provide one of the following: [replace]
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with unknown patch operation
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "add", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      patch operation 'add' not supported. Supported op(s): [replace]
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with unknown path for the patch operation
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/unknown", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      provided path '/unknown' not supported
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""
  
  Scenario: Request is made which updates an invalid number of tasks which is the wrong type
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/number_of_tasks", "value": "invalid" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      wrong value type `string` for `/number_of_tasks`, expected an integer
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made which updates an unknown state
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "unknown" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      invalid job state `unknown` for `/state` - expected [created completed failed in-progress]
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made which updates an invalid state which is the wrong type
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": 1 }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      wrong value type `integer` for `/state`, expected string
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made which updates an invalid total search documents which is the wrong type
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the search api is working correctly
    And set the api version to undefined for incoming requests
    And I have generated 1 jobs in the Job Store
    And I set the If-Match header to the generated e-tag
    
    When I call PATCH /jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/total_search_documents", "value": "invalid" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      wrong value type `string` for `/total_search_documents`, expected an integer
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""