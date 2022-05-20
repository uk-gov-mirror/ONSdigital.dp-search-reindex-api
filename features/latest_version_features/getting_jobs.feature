Feature: Getting a list of jobs

  Scenario: Three Jobs exist in the Job Store and a get request returns them successfully

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    When I GET "/jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
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

  Scenario: No Jobs exist in the Job Store and a get request returns an empty list

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    When I GET "/jobs"
    """
    """
    Then I would expect the response to be an empty list
    And the HTTP status code should be "200"

  Scenario: Six jobs exist and a get request with offset and limit correctly returns four

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 6
    When I GET "/jobs?offset=1&limit=4"
    """
    """
    Then I would expect there to be four jobs returned in a list
    And in each job I would expect the response to contain values that have these structures
      | id                | UUID                                    |
      | last_updated      | Not in the future                       |
      | links: tasks      | {host}/{latest_version}/jobs/{id}/tasks |
      | links: self       | {host}/{latest_version}/jobs/{id}       |
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

  Scenario: Three jobs exist and a get request with negative offset returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    When I GET "/jobs?offset=-2"
    """
    """
    Then the HTTP status code should be "400"

  Scenario: Three jobs exist and a get request with negative limit returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    When I GET "/jobs?limit=-3"
    """
    """
    Then the HTTP status code should be "400"

  Scenario: Three jobs exist and a get request with limit greater than the maximum returns an error

    Given the api version is undefined for incoming requests
    And the number of existing jobs in the Job Store is 3
    When I GET "/jobs?limit=1001"
    """
    """
    Then the HTTP status code should be "400"
