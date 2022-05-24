Feature: Getting a job

  Scenario: Job exists in the Job Store and a get request returns it successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the If-Match header to the generated e-tag
    When I call GET /jobs/{id} using the generated id
    Then the response should contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the response ETag header should not be empty

  Scenario: Request made with no If-Match header ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    When I call GET /jobs/{id} using the generated id
    Then the response should contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the response ETag header should not be empty

  Scenario: Request made with empty If-Match header ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the "If-Match" header to ""
    When I call GET /jobs/{id} using the generated id
    Then the response should contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the response ETag header should not be empty

  Scenario: Request made with If-Match set to `*` ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the "If-Match" header to "*"
    When I call GET /jobs/{id} using the generated id
    Then the response should contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the response ETag header should not be empty

  Scenario: Request made with outdated or invalid etag returns an conflict error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 1
    And I set the "If-Match" header to "invalid"
    When I call GET /jobs/{id} using the generated id
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of resource
    """ 
    And the response header "E-Tag" should be ""

  Scenario: When request is made with empty job id and request returns StatusNotFound by gorilla/mux as handler is not found

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    When I GET "/jobs/"
    Then the HTTP status code should be "404"
    And I should receive the following response:
    """
      404 page not found
    """
    And the response header "E-Tag" should be ""
  
  Scenario: Job does not exist in the Job Store and a get request returns StatusNotFound

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"} using a valid UUID
    Then the HTTP status code should be "404"
    And the response header "E-Tag" should be ""

  Scenario: The connection to mongo DB is lost and a get request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    And the api version is undefined for incoming requests
    When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"} using a valid UUID
    Then the HTTP status code should be "500"
    And the response header "E-Tag" should be ""
