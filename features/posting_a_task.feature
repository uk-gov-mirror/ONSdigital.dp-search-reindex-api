Feature: Posting a job

  Scenario: Task is created successfully

    Given I am authorised
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "name_of_api": "florence", "number_of_documents": 29 }
    """
#    Then I would expect id, last_updated, and links to have this structure
#      | job_id       | UUID                                           |
#      | last_updated | Not in the future                              |
#      | links: self  | http://{bind_address}/jobs/{id}/tasks/florence |
#      | links: job   | http://{bind_address}/jobs/{id}                |

#    And the response should also contain the following values:
#      | number_of_tasks                 | 0                         |
#      | reindex_completed               | 0001-01-01T00:00:00Z      |
#      | reindex_failed                  | 0001-01-01T00:00:00Z      |
#      | reindex_started                 | 0001-01-01T00:00:00Z      |
#      | search_index_name               | Default Search Index Name |
#      | state                           | created                   |
#      | total_search_documents          | 0                         |
#      | total_inserted_search_documents | 0                         |
#
#      Scenario: The connection to mongo DB is lost and a post request returns an internal server error
#
#    Given the search reindex api loses its connection to mongo DB
#    When I POST "/jobs"
#    """
#    """
#    Then the HTTP status code should be "500"
