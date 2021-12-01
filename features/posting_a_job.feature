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

      Scenario: The connection to mongo DB is lost and a post request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    When I POST "/jobs"
    """
    """
    Then the HTTP status code should be "500"
