package api

import (
	"fmt"
	"time"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// MessagingService handles LinkedIn messaging operations.
type MessagingService struct {
	c *client.Client
}

// NewMessagingService returns a new MessagingService.
func NewMessagingService(c *client.Client) *MessagingService {
	return &MessagingService{c: c}
}

// voyagerMessageEvent is the inner content of a messaging event.
type voyagerMessageEvent struct {
	AttributedBody struct {
		Text string `json:"text"`
	} `json:"attributedBody"`
}

type voyagerConversationEvent struct {
	EntityURN    string `json:"entityUrn"`
	CreatedAt    int64  `json:"createdAt"`
	DeliveredAt  int64  `json:"deliveredAt,omitempty"`
	EventContent struct {
		MessageEvent voyagerMessageEvent `json:"com.linkedin.voyager.messaging.event.MessageEvent"`
	} `json:"eventContent"`
	Actor struct {
		MiniProfile voyagerMiniProfile `json:"com.linkedin.voyager.messaging.MessagingMember"`
	} `json:"from"`
}

type voyagerConversation struct {
	EntityURN      string `json:"entityUrn"`
	Read           bool   `json:"read"`
	LastActivityAt int64  `json:"lastActivityAt"`
	Participants   []struct {
		MiniProfile voyagerMiniProfile `json:"com.linkedin.voyager.messaging.MessagingMember"`
	} `json:"participants"`
	Events []voyagerConversationEvent `json:"events,omitempty"`
}

// ListConversations returns a paged list of conversations.
func (s *MessagingService) ListConversations(start, count int) (*models.PagedConversations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"keyVersion": "LEGACY_INBOX",
		"start":      fmt.Sprintf("%d", start),
		"count":      fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerConversation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointConversations, params, &raw); err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	result := &models.PagedConversations{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}

	for _, vc := range raw.Elements {
		result.Items = append(result.Items, mapVoyagerConversation(vc))
	}

	return result, nil
}

// GetConversation returns messages in a conversation thread.
func (s *MessagingService) GetConversation(conversationID string, start, count int) (*models.PagedMessages, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	path := fmt.Sprintf(client.EndpointConversationEvents, conversationID)
	params := map[string]string{
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerConversationEvent `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(path, params, &raw); err != nil {
		return nil, fmt.Errorf("get conversation %q: %w", conversationID, err)
	}

	result := &models.PagedMessages{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}

	for _, ve := range raw.Elements {
		result.Items = append(result.Items, mapVoyagerEvent(ve))
	}

	return result, nil
}

// ListUnread returns only unread conversations.
func (s *MessagingService) ListUnread(start, count int) (*models.PagedConversations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"keyVersion": "LEGACY_INBOX",
		"q":          "unread",
		"start":      fmt.Sprintf("%d", start),
		"count":      fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerConversation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointConversations, params, &raw); err != nil {
		return nil, fmt.Errorf("list unread: %w", err)
	}

	result := &models.PagedConversations{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, vc := range raw.Elements {
		result.Items = append(result.Items, mapVoyagerConversation(vc))
	}
	return result, nil
}

// SendMessage sends a message to an existing conversation or starts a new one.
func (s *MessagingService) SendMessage(input models.SendMessageInput) error {
	payload := map[string]interface{}{
		"eventCreate": map[string]interface{}{
			"value": map[string]interface{}{
				"com.linkedin.voyager.messaging.create.MessageCreate": map[string]interface{}{
					"attributedBody": map[string]interface{}{
						"text":       input.Body,
						"attributes": []interface{}{},
					},
					"attachments": []interface{}{},
				},
			},
		},
	}

	var path string
	if input.ConversationURN != "" {
		convID := urnToID(input.ConversationURN)
		path = fmt.Sprintf(client.EndpointConversationEvents, convID)
	} else {
		recipients := make([]map[string]interface{}, 0, len(input.RecipientURNs))
		for _, urn := range input.RecipientURNs {
			recipients = append(recipients, map[string]interface{}{
				"com.linkedin.voyager.messaging.MessagingMember": map[string]interface{}{
					"miniProfile": map[string]interface{}{"entityUrn": urn},
				},
			})
		}
		payload["recipients"] = recipients
		path = client.EndpointConversations
	}

	return s.c.Post(path, payload, nil)
}

// MarkRead marks a conversation as read.
func (s *MessagingService) MarkRead(conversationID string) error {
	path := fmt.Sprintf("%s/%s", client.EndpointConversations, conversationID)
	return s.c.Put(path, map[string]interface{}{"read": true}, nil)
}

// mapVoyagerConversation converts a raw conversation to models.Conversation.
func mapVoyagerConversation(vc voyagerConversation) models.Conversation {
	conv := models.Conversation{
		URN:       vc.EntityURN,
		ID:        urnToID(vc.EntityURN),
		Unread:    !vc.Read,
		UpdatedAt: msToTime(vc.LastActivityAt),
	}
	for _, p := range vc.Participants {
		mp := p.MiniProfile
		if mp.EntityURN == "" {
			continue
		}
		conv.Participants = append(conv.Participants, models.Profile{
			URN:       mp.EntityURN,
			ProfileID: mp.PublicID,
			FirstName: mp.FirstName,
			LastName:  mp.LastName,
			Headline:  mp.Occupation,
		})
	}
	if len(vc.Events) > 0 {
		msg := mapVoyagerEvent(vc.Events[0])
		conv.LastMessage = &msg
	}
	return conv
}

// mapVoyagerEvent converts a raw conversation event to models.Message.
func mapVoyagerEvent(ve voyagerConversationEvent) models.Message {
	m := models.Message{
		URN:    ve.EntityURN,
		Body:   ve.EventContent.MessageEvent.AttributedBody.Text,
		SentAt: msToTime(ve.CreatedAt),
	}
	if ve.DeliveredAt > 0 {
		m.DeliveredAt = msToTime(ve.DeliveredAt)
	}
	mp := ve.Actor.MiniProfile
	m.SenderProfile = models.Profile{
		URN:       mp.EntityURN,
		ProfileID: mp.PublicID,
		FirstName: mp.FirstName,
		LastName:  mp.LastName,
	}
	return m
}

// msToTime converts a millisecond timestamp to an RFC3339 string.
func msToTime(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}
