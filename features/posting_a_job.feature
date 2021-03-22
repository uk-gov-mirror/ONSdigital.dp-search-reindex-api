Feature: Posting a job

  Scenario: Job is posted succesfully

    When I POST "/jobs"
    """
    """
    Then I would expect id, last_updated, and links to have these values
    """
    """
    And the response should also contain the following JSON:
    """
    {
        "id": "abc123",
        "last_updated": "2021-02-21T01:10:30Z",
        "links": {
            "tasks": "http://localhost:12150/jobs/abc123/tasks",
            "self": "http://localhost:12150/jobs/abc123"
        },
        "number_of_tasks": 0,
        "reindex_completed": "0001-01-01T00:00:00Z",
        "reindex_failed": "0001-01-01T00:00:00Z",
        "reindex_started": "0001-01-01T00:00:00Z",
        "search_index_name": "string",
        "state": "created",
        "total_search_documents": 0,
        "total_inserted_search_documents": 0
    }
    """