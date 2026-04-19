package services

import (
	"testing"

	"cs-agent/internal/pkg/config"
)

func TestWxWorkNotifyBuildTextContent(t *testing.T) {
	svc := newWxWorkNotifyService()
	got := svc.buildTextContent("工单提醒", "这是一条测试消息")
	if got != "工单提醒\n\n这是一条测试消息" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestWxWorkNotifyDefaultRecipients(t *testing.T) {
	config.SetCurrent(&config.Config{
		WxWork: config.WxWorkConfig{
			Notify: config.WxWorkNotifyConfig{
				Enabled:   true,
				ToUsers:   []string{" user_a ", "user_a", ""},
				ToParties: []string{"2", "2"},
				ToTags:    []string{"tag-1"},
			},
		},
	})

	svc := newWxWorkNotifyService()
	recipients := svc.defaultRecipients()
	if len(recipients.ToUsers) != 1 || recipients.ToUsers[0] != "user_a" {
		t.Fatalf("unexpected users: %#v", recipients.ToUsers)
	}
	if len(recipients.ToParties) != 1 || recipients.ToParties[0] != "2" {
		t.Fatalf("unexpected parties: %#v", recipients.ToParties)
	}
	if len(recipients.ToTags) != 1 || recipients.ToTags[0] != "tag-1" {
		t.Fatalf("unexpected tags: %#v", recipients.ToTags)
	}
}

func TestWxWorkNotifyNormalizeDuplicateCheckInterval(t *testing.T) {
	svc := newWxWorkNotifyService()
	if got := svc.normalizeDuplicateCheckInterval(0); got != 1800 {
		t.Fatalf("expected default interval 1800, got %d", got)
	}
	if got := svc.normalizeDuplicateCheckInterval(20000); got != 14400 {
		t.Fatalf("expected capped interval 14400, got %d", got)
	}
	if got := svc.normalizeDuplicateCheckInterval(600); got != 600 {
		t.Fatalf("expected interval 600, got %d", got)
	}
}
