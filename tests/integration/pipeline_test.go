package integration

import (
	"testing"

	"github.com/yourusername/api-monetization-platform/internal/events"
)

func TestUsageSubjectContract(t *testing.T) {
	if events.UsageRecordedSubject != "usage.recorded" {
		t.Fatalf("unexpected usage subject: %s", events.UsageRecordedSubject)
	}
	if events.UsageDLQSubject != "deadletter.usage" {
		t.Fatalf("unexpected DLQ subject: %s", events.UsageDLQSubject)
	}
}
