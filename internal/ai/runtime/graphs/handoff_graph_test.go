package graphs

import (
	"testing"

	"cs-agent/internal/models"
)

func TestHandoffGraphBuildReason(t *testing.T) {
	graph := NewHandoffGraph(&models.Conversation{ID: 1}, &models.AIAgent{Name: "AI"})

	reason, err := graph.buildReason(`{"reason":"  用户需要人工确认  "}`)
	if err != nil {
		t.Fatalf("buildReason returned error: %v", err)
	}
	if reason != "用户需要人工确认" {
		t.Fatalf("unexpected reason: %q", reason)
	}
}

func TestHandoffGraphBuildReasonFallback(t *testing.T) {
	graph := NewHandoffGraph(&models.Conversation{ID: 1}, &models.AIAgent{Name: "AI"})

	reason, err := graph.buildReason(`{}`)
	if err != nil {
		t.Fatalf("buildReason returned error: %v", err)
	}
	if reason != "用户需要转人工支持" {
		t.Fatalf("unexpected fallback reason: %q", reason)
	}
}

func TestHandoffGraphBuildSuccessReply(t *testing.T) {
	graph := NewHandoffGraph(&models.Conversation{ID: 1}, &models.AIAgent{Name: "AI"})

	got := graph.buildSuccessReply()
	want := "已为你转接人工客服，请稍候。，请稍候。"
	if got != want {
		t.Fatalf("unexpected success reply: %q", got)
	}
}
