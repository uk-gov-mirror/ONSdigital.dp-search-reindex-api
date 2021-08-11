Feature: Getting a list of jobs

  Scenario: Three Jobs exist in the Job Store and a get request returns them successfully

    Given I have generated three jobs in the Job Store
    When I GET "/jobs"
    """
    """
    Then I would expect there to be three or more jobs returned in a list
    And in each job I would expect id, last_updated, and links to have this structure
      | id           | UUID                                   |
      | last_updated | Not in the future                      |
      | links: tasks | http://localhost:12150/jobs/{id}/tasks |
      | links: self  | http://localhost:12150/jobs/{id}       |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | search_index_name               | Default Search Index Name |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first

  Scenario: No Jobs exist in the Job Store and a get request returns an empty list

    Given no jobs have been generated in the Job Store
    When I GET "/jobs"
    """
    """
    Then I would expect the response to be an empty list
    And the HTTP status code should be "200"

  Scenario: Six jobs exist and a get request with offset and limit correctly returns four

    Given I have generated six jobs in the Job Store
    When I GET "/jobs?offset=1&limit=4"
    """
    """
    Then I would expect there to be four jobs returned in a list
    And in each job I would expect id, last_updated, and links to have this structure
      | id           | UUID                                   |
      | last_updated | Not in the future                      |
      | links: tasks | http://localhost:12150/jobs/{id}/tasks |
      | links: self  | http://localhost:12150/jobs/{id}       |
    And each job should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | search_index_name               | Default Search Index Name |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the jobs should be ordered, by last_updated, with the oldest first

  Scenario: Three jobs exist and a get request with negative offset returns an error

    Given I have generated three jobs in the Job Store
    When I GET "/jobs?offset=-2"
    """
    """
    Then the HTTP status code should be "400"

  Scenario: Three jobs exist and a get request with offset greater than three returns an error

    Given I have generated three jobs in the Job Store
    When I GET "/jobs?offset=4"
    """
    """
    Then the HTTP status code should be "400"

  Scenario: Three jobs exist and a get request with negative limit returns an error

    Given I have generated three jobs in the Job Store
    When I GET "/jobs?limit=-3"
    """
    """
    Then the HTTP status code should be "400"

  Scenario: Three jobs exist and a get request with limit greater than the maximum returns an error

    Given I have generated three jobs in the Job Store
    When I GET "/jobs?limit=1001"
    """
    """
    Then the HTTP status code should be "400"
