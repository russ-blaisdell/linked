package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
	"github.com/russ-blaisdell/linked/internal/models"
)

//go:embed messenger_send.py
var messengerSendScript string

// MessagingService handles LinkedIn messaging operations.
type MessagingService struct {
	c          *client.Client
	profileURN string // lazily resolved fsd_profile URN for the authenticated user
}

// NewMessagingService returns a new MessagingService.
func NewMessagingService(c *client.Client) *MessagingService {
	return &MessagingService{c: c}
}

// getProfileURN returns the authenticated user's fsd_profile URN,
// fetching it from /me on first call and caching the result.
func (s *MessagingService) getProfileURN() (string, error) {
	if s.profileURN != "" {
		return s.profileURN, nil
	}
	var raw struct {
		Included []json.RawMessage `json:"included"`
	}
	if err := s.c.Get(client.EndpointMe, nil, &raw); err != nil {
		return "", fmt.Errorf("get profile URN: %w", err)
	}
	for _, inc := range raw.Included {
		var peek struct {
			DashEntityURN string `json:"dashEntityUrn"`
		}
		if json.Unmarshal(inc, &peek) == nil && peek.DashEntityURN != "" {
			s.profileURN = peek.DashEntityURN
			return s.profileURN, nil
		}
	}
	return "", fmt.Errorf("could not determine profile URN from /me")
}

// GraphQL response types for the messenger API.

type gqlMessengerText struct {
	Text string `json:"text"`
}

type gqlMessengerParticipant struct {
	HostIdentityURN string `json:"hostIdentityUrn"`
	ParticipantType *struct {
		Member *struct {
			FirstName *gqlMessengerText `json:"firstName"`
			LastName  *gqlMessengerText `json:"lastName"`
		} `json:"member"`
	} `json:"participantType"`
}

type gqlMessengerMessage struct {
	EntityURN   string `json:"entityUrn"`
	DeliveredAt int64  `json:"deliveredAt"`
	Body        *struct {
		Text string `json:"text"`
	} `json:"body"`
	Sender *struct {
		HostIdentityURN string `json:"hostIdentityUrn"`
		ParticipantType *struct {
			Member *struct {
				FirstName *gqlMessengerText `json:"firstName"`
				LastName  *gqlMessengerText `json:"lastName"`
			} `json:"member"`
		} `json:"participantType"`
	} `json:"sender"`
}

type gqlMessengerConversation struct {
	EntityURN    string `json:"entityUrn"`
	BackendURN   string `json:"backendUrn"`
	UnreadCount  int    `json:"unreadCount"`
	Read         bool   `json:"read"`
	LastActivity int64  `json:"lastActivityAt"`
	Participants []gqlMessengerParticipant `json:"conversationParticipants"`
	Messages     *struct {
		Elements []gqlMessengerMessage `json:"elements"`
	} `json:"messages"`
}

// ListConversations returns a paged list of conversations.
func (s *MessagingService) ListConversations(start, count int) (*models.PagedConversations, error) {
	return s.listConversationsByCategory("INBOX", start, count)
}

// ListUnread returns only unread conversations.
func (s *MessagingService) ListUnread(start, count int) (*models.PagedConversations, error) {
	result, err := s.listConversationsByCategory("INBOX", start, count)
	if err != nil {
		return nil, err
	}
	// Filter to unread only (the API doesn't have a direct unread filter).
	var unread []models.Conversation
	for _, c := range result.Items {
		if c.Unread {
			unread = append(unread, c)
		}
	}
	result.Items = unread
	result.Pagination.Total = len(unread)
	return result, nil
}

func (s *MessagingService) listConversationsByCategory(category string, start, count int) (*models.PagedConversations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	profileURN, err := s.getProfileURN()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(category:%s,count:%d,start:%d,mailboxUrn:%s)&queryId=%s",
		client.EndpointGraphQL, category, count, start,
		url.QueryEscape(profileURN),
		client.EndpointMessengerConversationsQueryID,
	)

	var raw struct {
		Data *struct {
			Collection *struct {
				Elements []gqlMessengerConversation `json:"elements"`
				Paging   struct {
					Start int `json:"start"`
					Count int `json:"count"`
					Total int `json:"total"`
				} `json:"paging"`
			} `json:"messengerConversationsByCategory"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	result := &models.PagedConversations{
		Pagination: models.Pagination{Start: start, Count: count},
	}

	if raw.Data == nil || raw.Data.Collection == nil {
		return result, nil
	}

	col := raw.Data.Collection
	result.Pagination.Total = len(col.Elements)
	result.Pagination.HasMore = len(col.Elements) >= count

	for _, gc := range col.Elements {
		conv := s.mapGraphQLConversation(gc, profileURN)
		result.Items = append(result.Items, conv)
	}

	return result, nil
}

// GetConversation returns messages in a conversation thread.
func (s *MessagingService) GetConversation(conversationID string, start, count int) (*models.PagedMessages, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	profileURN, err := s.getProfileURN()
	if err != nil {
		return nil, err
	}

	// Build the full conversation URN from the thread ID.
	convURN := fmt.Sprintf("urn:li:msg_conversation:(%s,%s)", profileURN, conversationID)

	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(conversationUrn:%s,count:%d)&queryId=%s",
		client.EndpointGraphQL,
		url.QueryEscape(convURN),
		count,
		client.EndpointMessengerMessagesQueryID,
	)

	var raw struct {
		Data *struct {
			Messages *struct {
				Elements []gqlMessengerMessage `json:"elements"`
				Paging   struct {
					Start int `json:"start"`
					Count int `json:"count"`
					Total int `json:"total"`
				} `json:"paging"`
			} `json:"messengerMessagesByConversation"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("get conversation %q: %w", conversationID, err)
	}

	result := &models.PagedMessages{
		Pagination: models.Pagination{Start: start, Count: count},
	}

	if raw.Data == nil || raw.Data.Messages == nil {
		return result, nil
	}

	msgs := raw.Data.Messages
	result.Pagination.Total = len(msgs.Elements)
	result.Pagination.HasMore = len(msgs.Elements) >= count

	for _, gm := range msgs.Elements {
		result.Items = append(result.Items, mapGraphQLMessage(gm))
	}

	return result, nil
}

// SendMessage sends a message to an existing conversation or starts a new one.
// Uses a Python helper script because LinkedIn's messenger POST endpoint is
// extremely sensitive to TLS fingerprints and HTTP headers — Go's TLS stack
// (both default and utls) triggers session revocation, while Python's http.client
// works reliably. The Python script sends only the 4 essential headers.
func (s *MessagingService) SendMessage(input models.SendMessageInput) error {
	if input.ConversationURN == "" && len(input.RecipientURNs) == 0 {
		return fmt.Errorf("provide a conversation URN to reply or recipient URNs to start a new conversation")
	}
	if input.Body == "" {
		return fmt.Errorf("message body is required")
	}

	// Write the embedded Python script to a temp file.
	tmpFile, err := os.CreateTemp("", "linked-messenger-*.py")
	if err != nil {
		return fmt.Errorf("creating temp script: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(messengerSendScript); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp script: %w", err)
	}
	tmpFile.Close()

	// Get credentials path.
	credsPath, err := config.CredentialsPath("default")
	if err != nil {
		return fmt.Errorf("finding credentials: %w", err)
	}

	cmd := exec.Command("python3", tmpFile.Name(), credsPath, input.ConversationURN, input.Body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("send message: %s", string(output))
	}
	return nil
}

// MarkRead marks a conversation as read.
// TODO: Needs migration to new messaging protocol.
func (s *MessagingService) MarkRead(conversationID string) error {
	return fmt.Errorf("mark read is not yet supported (endpoint migration in progress)")
}

// StarConversation stars (bookmarks) a conversation.
func (s *MessagingService) StarConversation(conversationID string) error {
	return fmt.Errorf("star conversation is not yet supported (endpoint migration in progress)")
}

// UnstarConversation removes the star from a conversation.
func (s *MessagingService) UnstarConversation(conversationID string) error {
	return fmt.Errorf("unstar conversation is not yet supported (endpoint migration in progress)")
}

// ArchiveConversation archives a conversation.
func (s *MessagingService) ArchiveConversation(conversationID string) error {
	return fmt.Errorf("archive conversation is not yet supported (endpoint migration in progress)")
}

// UnarchiveConversation restores an archived conversation.
func (s *MessagingService) UnarchiveConversation(conversationID string) error {
	return fmt.Errorf("unarchive conversation is not yet supported (endpoint migration in progress)")
}

// DeleteMessage deletes a specific message from a conversation.
func (s *MessagingService) DeleteMessage(conversationID, messageURN string) error {
	return fmt.Errorf("delete message is not yet supported (endpoint migration in progress)")
}

// DeleteConversation deletes an entire conversation.
func (s *MessagingService) DeleteConversation(conversationID string) error {
	return fmt.Errorf("delete conversation is not yet supported (endpoint migration in progress)")
}

// mapGraphQLConversation converts a GraphQL conversation to models.Conversation.
func (s *MessagingService) mapGraphQLConversation(gc gqlMessengerConversation, myURN string) models.Conversation {
	conv := models.Conversation{
		URN:       gc.EntityURN,
		ID:        extractThreadID(gc.BackendURN),
		Unread:    gc.UnreadCount > 0,
		UpdatedAt: msToTime(gc.LastActivity),
	}

	for _, p := range gc.Participants {
		if p.HostIdentityURN == myURN {
			continue // skip self
		}
		profile := models.Profile{URN: p.HostIdentityURN}
		if p.ParticipantType != nil && p.ParticipantType.Member != nil {
			m := p.ParticipantType.Member
			if m.FirstName != nil {
				profile.FirstName = m.FirstName.Text
			}
			if m.LastName != nil {
				profile.LastName = m.LastName.Text
			}
		}
		conv.Participants = append(conv.Participants, profile)
	}

	if gc.Messages != nil && len(gc.Messages.Elements) > 0 {
		msg := mapGraphQLMessage(gc.Messages.Elements[0])
		conv.LastMessage = &msg
	}

	return conv
}

// mapGraphQLMessage converts a GraphQL message to models.Message.
func mapGraphQLMessage(gm gqlMessengerMessage) models.Message {
	msg := models.Message{
		URN:    gm.EntityURN,
		SentAt: msToTime(gm.DeliveredAt),
	}
	if gm.Body != nil {
		msg.Body = gm.Body.Text
	}
	if gm.Sender != nil {
		msg.SenderProfile.URN = gm.Sender.HostIdentityURN
		if gm.Sender.ParticipantType != nil && gm.Sender.ParticipantType.Member != nil {
			m := gm.Sender.ParticipantType.Member
			if m.FirstName != nil {
				msg.SenderProfile.FirstName = m.FirstName.Text
			}
			if m.LastName != nil {
				msg.SenderProfile.LastName = m.LastName.Text
			}
		}
	}
	return msg
}

// extractThreadID extracts the thread ID from a backendUrn like
// "urn:li:messagingThread:2-XXXXX"
func extractThreadID(backendURN string) string {
	parts := splitURN(backendURN)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return backendURN
}

func splitURN(urn string) []string {
	// Split "urn:li:messagingThread:2-XXX" into parts by ":"
	result := make([]string, 0, 4)
	start := 0
	for i, ch := range urn {
		if ch == ':' {
			result = append(result, urn[start:i])
			start = i + 1
		}
	}
	result = append(result, urn[start:])
	return result
}

// msToTime converts a millisecond timestamp to an RFC3339 string.
func msToTime(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}
