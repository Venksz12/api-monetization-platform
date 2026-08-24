package events

import (
	"encoding/json"
	"time"
)

type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Producer      string          `json:"producer"`
	RequestID     string          `json:"request_id"`
	CorrelationID string          `json:"correlation_id"`
	Payload       json.RawMessage `json:"payload"`
}

func NewEnvelope(id, typ, producer, requestID, correlationID string, payload any) (Envelope, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		EventID: id, EventType: typ, Version: 1, OccurredAt: time.Now().UTC(),
		Producer: producer, RequestID: requestID, CorrelationID: correlationID, Payload: b,
	}, nil
}
