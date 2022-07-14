Feature: Patch job state - Failure
  
  Scenario: Request is made with invalid service auth token

    Given I use a service auth token "invalidServiceAuthToken"
    And zebedee does not recognise the service auth token
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the "If-Match" header to "invalid"
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of resource
    """ 
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""
  
  Scenario: Request is made with no modification to job resource
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the "If-Match" header to the old e-tag

    When I call PATCH /search-reindex-jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "304"
    And I should receive the following response:
    """
      no modification made on resource
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with invalid job ID
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I PATCH "/search-reindex-jobs/invalid"
    """
    [
      { "op": "replace", "path": "/state", "value": "created" }
    ]
    """
    
    Then the HTTP status code should be "404"
    And I should receive the following response:
    """
      failed to find the specified reindex job
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made with empty patch body
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
    """
    [
      { "op": "replace", "path": "/number-of-tasks", "value": "invalid" }
    ]
    """
    
    Then the HTTP status code should be "400"
    And I should receive the following response:
    """
      wrong value type `string` for `/number-of-tasks`, expected an integer
    """
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And the response header "E-Tag" should be ""

  Scenario: Request is made which updates an unknown state
    Given I use a service auth token "validServiceAuthToken"
    And zebedee recognises the service auth token as valid
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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
    And the api version is v1 for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated job e-tag
    
    When I call PATCH /search-reindex-jobs/{id} using the generated id
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