package ingest

type RequestLog struct {
	Path       string `bson:"path" json:"path"`
	Method     string `bson:"method" json:"method"`
	IP         string `bson:"ip" json:"ip"`
	TraceID    string `bson:"traceId,omitempty" json:"traceId,omitempty"`
	DurationMS int64  `bson:"durationMs" json:"durationMs"`
	StatusCode int    `bson:"statusCode" json:"statusCode"`
	OccurredAt string `bson:"occurredAt" json:"occurredAt"`
}
