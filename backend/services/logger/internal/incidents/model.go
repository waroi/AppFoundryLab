package incidents

type RunbookReference struct {
	ID       string `bson:"id" json:"id"`
	Title    string `bson:"title" json:"title"`
	Path     string `bson:"path" json:"path"`
	Reason   string `bson:"reason" json:"reason"`
	Priority string `bson:"priority" json:"priority"`
}

type Event struct {
	ID                  string             `bson:"id" json:"id"`
	EventType           string             `bson:"eventType" json:"eventType"`
	AlertCode           string             `bson:"alertCode" json:"alertCode"`
	Severity            string             `bson:"severity" json:"severity"`
	Status              string             `bson:"status" json:"status"`
	Source              string             `bson:"source" json:"source"`
	Title               string             `bson:"title" json:"title"`
	Summary             string             `bson:"summary" json:"summary"`
	Message             string             `bson:"message" json:"message"`
	RecommendedAction   string             `bson:"recommendedAction" json:"recommendedAction"`
	RecommendedSeverity string             `bson:"recommendedSeverity" json:"recommendedSeverity"`
	TriggeredAt         string             `bson:"triggeredAt" json:"triggeredAt"`
	FirstSeenAt         string             `bson:"firstSeenAt" json:"firstSeenAt"`
	LastSeenAt          string             `bson:"lastSeenAt" json:"lastSeenAt"`
	BreachCount         int                `bson:"breachCount" json:"breachCount"`
	TraceID             string             `bson:"traceId,omitempty" json:"traceId,omitempty"`
	ReportGeneratedAt   string             `bson:"reportGeneratedAt" json:"reportGeneratedAt"`
	ReportVersion       string             `bson:"reportVersion" json:"reportVersion"`
	Runbooks            []RunbookReference `bson:"runbooks" json:"runbooks"`
}

type Summary struct {
	TotalEvents     uint64 `json:"totalEvents"`
	ActiveEvents    uint64 `json:"activeEvents"`
	LatestEventAt   string `json:"latestEventAt"`
	LastEventStatus string `json:"lastEventStatus"`
}
