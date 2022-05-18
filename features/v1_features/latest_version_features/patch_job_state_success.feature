Feature: Patch job state - Success
  
    Scenario: Request is made with valid job ID and valid patch operations and returns 204 status code

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "created" },
            { "op": "replace", "path": "/total_search_documents", "value": 100 },
            { "op": "replace", "path": "/number_of_tasks", "value": 2 }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | state                  | created                   |
            | total_search_documents | 100                       |
            | number_of_tasks        | 2                         |
  
    Scenario: Request is made with no If-Match header which then sets If-Match header to IfMatchAnyETag (`*`) to ask the API to ignore the ETag check

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "in-progress" }
        ]
        """
        
        And the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | reindex_started        | now or earlier            |
            | state                  | in-progress               |
  
    Scenario: Request is made with empty If-Match header which then sets If-Match header to IfMatchAnyETag (`*`) to ask the API to ignore the ETag check

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the "If-Match" header to ""
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "in-progress" }
        ]
        """
        
        And the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | reindex_started        | now or earlier            |
            | state                  | in-progress               |
  
    Scenario: Request is made to update state to in-progress

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "in-progress" }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | reindex_started        | now or earlier            |
            | state                  | in-progress               |

    Scenario: Request is made to update state to failed

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "failed" }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | reindex_failed         | now or earlier            |
            | state                  | failed                    |
  
    Scenario: Request is made to update state to completed

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/state", "value": "completed" }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | reindex_completed      | now or earlier            |
            | state                  | completed                 |

    Scenario: Request is made to update total search documents

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/total_search_documents", "value": 100 }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | total_search_documents | 100                       |

    Scenario: Request is made to update number of tasks

        Given I use a service auth token "validServiceAuthToken"
        And zebedee recognises the service auth token as valid
        And the search api is working correctly
        And set the api version to v1 for incoming requests
        And I have generated 1 jobs in the Job Store
        And I set the If-Match header to the generated e-tag
        
        When I call PATCH /jobs/{id} using the generated id
        """
        [
            { "op": "replace", "path": "/number_of_tasks", "value": 2 }
        ]
        """
        
        Then the HTTP status code should be "204"
        And the response header "Content-Type" should be ""
        And the response ETag header should not be empty
        And the job should only be updated with the following fields and values
            | e_tag                  | new eTag                  |
            | last_updated           | now or earlier            |
            | number_of_tasks        | 2                         |
