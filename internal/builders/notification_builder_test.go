package builders

import (
	"testing"

	"cs-agent/internal/models"
)

func TestBuildNotificationListReturnsEmptySlice(t *testing.T) {
	results := BuildNotificationList([]models.Notification{})

	if results == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(results))
	}
}
