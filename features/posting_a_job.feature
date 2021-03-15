Feature: Posting a job

  Scenario: Job is posted succesfully

    When I POST "/jobs"
    """
    {
        "last_updated": "2021-03-15",
        "number_of_tasks": 0,
        "reindex_completed": "2021-03-15",
        "reindex_failed": "2021-03-15",
        "reindex_started": "2021-03-15",
        "search_index_name": "string",
        "state": "CREATED",
        "total_search_documents": 0
    }
    """
    Then I should receive the following JSON response with status "200":
    """
    {
        "id": "abc123",
        "last_updated": "2021-03-15",
        "links": {
            "tasks": "string",
            "self": "string"
        },
        "number_of_tasks": 0,
        "reindex_completed": "2021-03-15",
        "reindex_failed": "2021-03-15",
        "reindex_started": "2021-03-15",
        "search_index_name": "string",
        "state": "CREATED",
        "total_search_documents": 0
    }
    """