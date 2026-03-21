package integration_test

import (
	"strings"
	"testing"
)

func TestListNotifications(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Notifications.List(0, 20)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one notification")
	}

	n := result.Items[0]
	if n.Body == "" {
		t.Error("notification Body should not be empty")
	}
	if n.CreatedAt == "" {
		t.Error("notification CreatedAt should not be empty")
	}
}

func TestNotificationReadStatus(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Notifications.List(0, 20)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	var readCount, unreadCount int
	for _, n := range result.Items {
		if n.Read {
			readCount++
		} else {
			unreadCount++
		}
	}

	if readCount == 0 {
		t.Error("expected at least one read notification")
	}
	if unreadCount == 0 {
		t.Error("expected at least one unread notification")
	}
}

func TestGetBadgeCount(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	badge, err := li.Notifications.GetBadgeCount()
	if err != nil {
		t.Fatalf("GetBadgeCount() error: %v", err)
	}
	if badge.UnreadCount != 3 {
		t.Errorf("expected 3 unread, got %d", badge.UnreadCount)
	}
}

func TestMarkNotificationReadNotSupported(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	err := li.Notifications.MarkRead("urn:li:notification:aaa111")
	if err == nil {
		t.Fatal("expected error for unsupported mark read")
	}
	if !strings.Contains(err.Error(), "not yet supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
