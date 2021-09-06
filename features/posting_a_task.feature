Feature: Posting a job

  Scenario: Task is created successfully

    Given I am authorised
    And I have generated a job in the Job Store
    When I call POST /jobs/{id}/tasks using the generated id
    """
    { "name_of_api": "florence", "number_of_documents": 29 }
    """
    Then I would expect job_id, last_updated, and links to have this structure
      | job_id       | UUID                                           |
      | last_updated | Not in the future                              |
      | links: self  | http://{bind_address}/jobs/{id}/tasks/florence |
      | links: job   | http://{bind_address}/jobs/{id}                |
    And the task resource should also contain the following values:
      | number_of_documents               | 29                        |
      | task                              | florence                  |

#
#      Scenario: The connection to mongo DB is lost and a post request returns an internal server error
#
#    Given the search reindex api loses its connection to mongo DB
#    When I POST "/jobs"
#    """
#    """
#    Then the HTTP status code should be "500"
