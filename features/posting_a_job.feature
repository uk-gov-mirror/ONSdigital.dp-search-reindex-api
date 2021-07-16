Feature: Posting a job

  Scenario: Job is posted succesfully

    When I POST "/jobs"
    """
    """
    Then I would expect id, last_updated, and links to have this structure
      | id           | UUID                                   |
      | last_updated | Not in the future                      |
      | links: tasks | http://localhost:12150/jobs/{id}/tasks |
      | links: self  | http://localhost:12150/jobs/{id}       |
    And the response should also contain the following values:
      | number_of_tasks                 | 0                         |
      | reindex_completed               | 0001-01-01T00:00:00Z      |
      | reindex_failed                  | 0001-01-01T00:00:00Z      |
      | reindex_started                 | 0001-01-01T00:00:00Z      |
      | search_index_name               | Default Search Index Name |
      | state                           | created                   |
      | total_search_documents          | 0                         |
      | total_inserted_search_documents | 0                         |

  Scenario: The connection to mongo DB is lost and a get request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    When I POST "/jobs"
    """
    """
    Then the HTTP status code should be "500"
