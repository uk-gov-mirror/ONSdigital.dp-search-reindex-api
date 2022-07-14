Feature: Getting a list of jobs

  Scenario: Three Jobs exist in the Job Store and a get request returns them successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the If-Match header to a valid e-tag to get jobs
    When I GET "/search-reindex-jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/search-reindex-jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/search-reindex-jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: No Jobs exist in the Job Store and a get request returns an empty list

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    And I set the If-Match header to a valid e-tag to get jobs
    When I GET "/search-reindex-jobs"
    """
    """
    Then I would expect the response to be an empty list
    And the HTTP status code should be "200"
    And the response ETag header should not be empty

  Scenario: Six jobs exist and a get request with offset and limit correctly returns four

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 6
    When I GET "/search-reindex-jobs?offset=1&limit=4"
    """
    """
    Then I would expect there to be four jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/search-reindex-jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/search-reindex-jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with no If-Match header ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    When I GET "/search-reindex-jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/search-reindex-jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/search-reindex-jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with empty If-Match header ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the "If-Match" header to "" 
    When I GET "/search-reindex-jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/search-reindex-jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/search-reindex-jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with If-Match set to `*` ignores the ETag check and returns jobs successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the "If-Match" header to "*" 
    When I GET "/search-reindex-jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/search-reindex-jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/search-reindex-jobs/{id}       |
      | search_index_name | ons{date_stamp}                         |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first
    And the response ETag header should not be empty

  Scenario: Request made with outdated or invalid etag returns an conflict error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the "If-Match" header to "invalid"
    When I GET "/search-reindex-jobs"
    Then the HTTP status code should be "409"
    And I should receive the following response:
    """
      etag does not match with current state of resource
    """ 
    And the response header "E-Tag" should be ""

  Scenario: Three jobs exist and a get request with negative offset returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the If-Match header to a valid e-tag to get jobs
    When I GET "/search-reindex-jobs?offset=-2"
    """
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""

  Scenario: Three jobs exist and a get request with negative limit returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the If-Match header to a valid e-tag to get jobs
    When I GET "/search-reindex-jobs?limit=-3"
    """
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""

  Scenario: Three jobs exist and a get request with limit greater than the maximum returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    And I set the If-Match header to a valid e-tag to get jobs
    When I GET "/search-reindex-jobs?limit=1001"
    """
    """
    Then the HTTP status code should be "400"
    And the response header "E-Tag" should be ""
