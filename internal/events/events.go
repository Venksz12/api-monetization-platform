package events

import "time"

const UsageSubject = "usage.recorded"

type UsageRecorded struct {
	EventID    string    `json:"event_id"`
	APIKeyID   string    `json:"api_key_id"`
	OwnerID    string    `json:"owner_id"`
	Path       string    `json:"path"`
	Units      int64     `json:"units"`
	OccurredAt time.Time `json:"occurred_at"`
}
