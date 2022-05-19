Feature: Updating the number of tasks for a particular job

  Scenario: Job exists in the Job Store and a put request updates its number of tasks successfully

    Given the number of existing jobs in the Job Store is 1
    And the api version is undefined for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "200"
    Given I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Job does not exist in the Job Store and a put request, to update its number of tasks, returns StatusNotFound

    Given the number of existing jobs in the Job Store is 0
    And the api version is undefined for incoming requests
    When I call PUT /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/number_of_tasks/{7} using a valid UUID
    Then the HTTP status code should be "404"

  Scenario: A put request fails to update the number of tasks because it contains an invalid value of count

    Given the number of existing jobs in the Job Store is 1
    And the api version is undefined for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{"seven"} using the generated id with an invalid count
    Then the HTTP status code should be "400"

  Scenario: A put request fails to update the number of tasks because it contains a negative value of count

    Given the number of existing jobs in the Job Store is 1
    And the api version is undefined for incoming requests
    When I call PUT /jobs/{id}/number_of_tasks/{"-7"} using the generated id with a negative count
    Then the HTTP status code should be "400"

  Scenario: The connection to mongo DB is lost and a put request returns an internal server error

    Given the search reindex api loses its connection to mongo DB
    And the api version is undefined for incoming requests
    When I call PUT /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/number_of_tasks/{7} using a valid UUID
    Then the HTTP status code should be "500"
