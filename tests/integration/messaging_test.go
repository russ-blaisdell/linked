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

func TestSendMessageReply(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SendMessageInput{
		ConversationURN: "urn:li:msg_conversation:(urn:li:fsd_profile:test-user-encoded-id,thread001)",
		Body:            "Test reply",
	}
	if err := li.Messaging.SendMessage(input); err != nil {
		t.Fatalf("SendMessage() error: %v", err)
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
