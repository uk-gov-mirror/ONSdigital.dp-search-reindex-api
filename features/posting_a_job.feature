Feature: Posting a job

  Scenario: Job is posted successfully

    Given the search api is working correctly
    When I POST "/jobs"
    """
    """
    Then the response should contain values that have these structures
      | id                | UUID                                  |
      | last_updated      | Not in the future                     |
      | links: tasks      | http://{bind_address}/jobs/{id}/tasks |
      | links: self       | http://{bind_address}/jobs/{id}       |
      | search_index_name | ons{date_stamp}                       |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |
    And the HTTP status code should be "201"
    And the reindex-requested event should contain the expected job ID and search index name

  Scenario: The connection to mongo DB is lost and a post request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    When I POST "/jobs"
    """
    """
    Then the HTTP status code should be "500"

  Scenario: The connection to search API is lost and a post request returns a job state of failed

    Given the search reindex api loses its connection to the search api
    When I POST "/jobs"
    """
    """
    Then the response should contain a state of "failed"
    And the HTTP status code should be "201"

  Scenario: The search API is not pointing to the correct version of ES and a post request returns a job state of failed

    Given the search reindex api is not working correctly because it is pointing to the old version of ES
    When I POST "/jobs"
    """
    """
    Then the response should contain a state of "failed"
    And the HTTP status code should be "201"
