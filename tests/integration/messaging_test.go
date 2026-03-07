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

	if len(convs.Items) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(convs.Items))
	}

	first := convs.Items[0]
	if first.ID == "" {
		t.Error("conversation ID should not be empty")
	}
	if !first.Unread {
		t.Error("first conversation should be unread")
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

	msgs, err := li.Messaging.GetConversation("2-abc111", 0, 20)
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

	// All returned conversations should be marked unread or it's an empty list.
	for _, c := range convs.Items {
		if !c.Unread {
			t.Errorf("conversation %s is not unread", c.ID)
		}
	}
}

func TestSendMessageNewConversation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SendMessageInput{
		RecipientURNs: []string{"urn:li:member:987654321"},
		Body:          "Hello! Great to connect.",
	}

	if err := li.Messaging.SendMessage(input); err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}

	sent := s.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(sent))
	}
}

func TestSendMessageReply(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SendMessageInput{
		ConversationURN: "urn:li:msg_conversation:2-abc111",
		Body:            "Thanks for reaching out!",
	}

	if err := li.Messaging.SendMessage(input); err != nil {
		t.Fatalf("SendMessage() reply error: %v", err)
	}
}

func TestMarkRead(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Messaging.MarkRead("2-abc111"); err != nil {
		t.Fatalf("MarkRead() error: %v", err)
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
