package integration_test

import "testing"

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
	if n.ID == "" {
		t.Error("notification ID should not be empty")
	}
	if n.Type == "" {
		t.Error("notification Type should not be empty")
	}
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

	// Fixture has one read and two unread notifications.
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

func TestMarkNotificationRead(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Notifications.MarkRead("urn:li:notification:aaa111"); err != nil {
		t.Fatalf("MarkRead() error: %v", err)
	}
}
