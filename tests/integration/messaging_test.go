package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestListConversations(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	convs, err := li.Messaging.ListConversations(0, 20)
	if err != nil {
		t.Fatalf("ListConversations() error: %v", err)
	}

	if len(convs.Items) == 0 {
		t.Fatal("expected at least one conversation")
	}

	first := convs.Items[0]
	if first.ID == "" {
		t.Error("conversation ID should not be empty")
	}
	if first.LastMessage == nil {
		t.Error("first conversation should have a last message")
	}
	if first.LastMessage.Body == "" {
		t.Error("last message body should not be empty")
	}
	if len(first.Participants) == 0 {
		t.Error("first conversation should have participants")
	}
}

func TestGetConversation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	msgs, err := li.Messaging.GetConversation("thread001", 0, 20)
	if err != nil {
		t.Fatalf("GetConversation() error: %v", err)
	}

	if len(msgs.Items) == 0 {
		t.Fatal("expected at least one message")
	}

	msg := msgs.Items[0]
	if msg.Body == "" {
		t.Error("message body should not be empty")
	}
	if msg.SentAt == "" {
		t.Error("message SentAt should not be empty")
	}
	if msg.SenderProfile.FirstName == "" {
		t.Error("sender FirstName should not be empty")
	}
}

func TestListUnread(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	convs, err := li.Messaging.ListUnread(0, 20)
	if err != nil {
		t.Fatalf("ListUnread() error: %v", err)
	}

	// All returned conversations should be unread.
	for _, c := range convs.Items {
		if !c.Unread {
			t.Errorf("conversation %s is not unread", c.ID)
		}
	}
}

func TestSendMessageValidation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	// SendMessage requires conversation URN or recipients.
	err := li.Messaging.SendMessage(models.SendMessageInput{Body: "test"})
	if err == nil {
		t.Fatal("expected error for missing conversation/recipients")
	}

	// SendMessage requires body.
	err = li.Messaging.SendMessage(models.SendMessageInput{ConversationURN: "urn:li:msg_conversation:test"})
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestPagination(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	convs, err := li.Messaging.ListConversations(0, 1)
	if err != nil {
		t.Fatalf("ListConversations() error: %v", err)
	}

	if convs.Pagination.Count != 1 {
		t.Errorf("Pagination.Count = %d, want 1", convs.Pagination.Count)
	}
}
