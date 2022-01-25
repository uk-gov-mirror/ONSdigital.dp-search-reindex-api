package schema

import (
	"github.com/ONSdigital/dp-kafka/v2/avro"
)

var reindexRequested = `{
	"type": "record",
	"name": "reindex-requested",
	"fields": [
		{"name": "job_id", "type": "string", "default": ""},
		{"name": "search_index", "type": "string", "default": ""},
		{"name": "trace_id", "type": "string", "default": ""}
	]
}`

// ReindexRequestedEvent is the Avro schema for Reindex Requested messages.
var ReindexRequestedEvent = &avro.Schema{
	Definition: reindexRequested,
}
