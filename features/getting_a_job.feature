# Feature: Getting a job

#   Scenario: Job exists in the Job Store and a get request returns it successfully

#     Given the search api is working correctly
#     And I have generated 1 jobs in the Job Store
#     When I call GET /jobs/{id} using the generated id
#     Then the response should contain values that have these structures
#       | id                | UUID                                    |
#       | last_updated      | Not in the future                       |
#       | links: tasks      | http://{bind_address}/jobs/{id}/tasks   |
#       | links: self       | http://{bind_address}/jobs/{id}         |
#       | search_index_name | ons{date_stamp}                         |
#     And the response should also contain the following values:
#       | number_of_tasks                 | 0                         |
#       | reindex_completed               | 0001-01-01T00:00:00Z      |
#       | reindex_failed                  | 0001-01-01T00:00:00Z      |
#       | reindex_started                 | 0001-01-01T00:00:00Z      |
#       | state                           | created                   |
#       | total_search_documents          | 0                         |
#       | total_inserted_search_documents | 0                         |

#   Scenario: Job does not exist in the Job Store and a get request returns StatusNotFound

#     Given I have generated 0 jobs in the Job Store
#     When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"} using a valid UUID
#     Then the HTTP status code should be "404"

#   Scenario: The connection to mongo DB is lost and a get request returns an internal server error

#     Given the search reindex api loses its connection to mongo DB
#     When I call GET /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"} using a valid UUID
#     Then the HTTP status code should be "500"
