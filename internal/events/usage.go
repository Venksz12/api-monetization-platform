package events

import "time"

const UsageRecordedSubject = "usage.recorded"
const UsageDLQSubject = "deadletter.usage"

type UsageRecorded struct {
	EventID    string    `json:"event_id"`
	APIKeyID   string    `json:"api_key_id"`
	OwnerID    string    `json:"owner_id"`
	Path       string    `json:"path"`
	Units      int64     `json:"units"`
	OccurredAt time.Time `json:"occurred_at"`
}
