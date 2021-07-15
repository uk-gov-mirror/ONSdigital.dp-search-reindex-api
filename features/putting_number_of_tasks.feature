Feature: Updating the number of tasks for a particular job

  Scenario: Job exists in the Job Store and a put request updates its number of tasks successfully

    Given I have generated a job in the Job Store
    When I call PUT /jobs/{id}/number_of_tasks/{7} using the generated id
    Then the HTTP status code should be "200"
    Given I call GET /jobs/{id} using the generated id
    Then the response should contain the new number of tasks
      | number_of_tasks | 7 |

  Scenario: Job does not exist in the Job Store and a put request, to update its number of tasks, returns StatusNotFound

    Given no jobs have been generated in the Job Store
    When I call PUT /jobs/{"a219584a-454a-4add-92c6-170359b0ee77"}/number_of_tasks/{7} using a valid UUID
    Then the HTTP status code should be "404"
