Feature: Posting a job

  Scenario: Job is posted succesfully

    When I POST "/jobs"
    """
    """
    Then I should receive the following JSON response with status "200":
    """
    {
        "id": "abc123",
        "last_updated": "2021-02-21T01:10:30Z",
        "links": {
            "tasks": "string",
            "self": "string"
        },
        "number_of_tasks": 0,
        "reindex_completed": "2021-02-21T01:10:30Z",
        "reindex_failed": "2021-02-21T01:10:30Z",
        "reindex_started": "2021-02-21T01:10:30Z",
        "search_index_name": "string",
        "state": "CREATED",
        "total_search_documents": 0,
        "total_inserted_search_documents": 0
    }
    """