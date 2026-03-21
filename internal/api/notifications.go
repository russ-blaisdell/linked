package api

import (
	"encoding/json"
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// NotificationsService handles LinkedIn notification operations.
type NotificationsService struct {
	c *client.Client
}

// NewNotificationsService returns a new NotificationsService.
func NewNotificationsService(c *client.Client) *NotificationsService {
	return &NotificationsService{c: c}
}

// List returns recent notifications for the authenticated user.
// Uses the dash notification cards endpoint with normalized response format.
func (s *NotificationsService) List(start, count int) (*models.PagedNotifications, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	path := fmt.Sprintf(
		"%s?decorationId=com.linkedin.voyager.dash.deco.identity.notifications.CardsCollectionWithInjectionsNoPills-24&count=%d&start=%d&q=filterVanityName",
		client.EndpointDashNotificationCards, count, start,
	)

	var raw struct {
		Data struct {
			Elements []string `json:"*elements"`
			Paging   struct {
				Start int `json:"start"`
				Count int `json:"count"`
			} `json:"paging"`
		} `json:"data"`
		Included []json.RawMessage `json:"included"`
	}

	if err := s.c.Get(path, nil, &raw); err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	// Index included entities.
	byURN := make(map[string]json.RawMessage)
	for _, inc := range raw.Included {
		var peek struct {
			EntityURN string `json:"entityUrn"`
		}
		if json.Unmarshal(inc, &peek) == nil && peek.EntityURN != "" {
			byURN[peek.EntityURN] = inc
		}
	}

	result := &models.PagedNotifications{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   len(raw.Data.Elements),
			HasMore: len(raw.Data.Elements) >= count,
		},
	}

	for _, urn := range raw.Data.Elements {
		cardRaw, ok := byURN[urn]
		if !ok {
			continue
		}
		var card struct {
			EntityURN   string `json:"entityUrn"`
			Headline    *struct{ Text string `json:"text"` } `json:"headline"`
			BodyText    *struct{ Text string `json:"text"` } `json:"bodyText"`
			Read        bool  `json:"read"`
			PublishedAt int64 `json:"publishedAt"`
		}
		if json.Unmarshal(cardRaw, &card) != nil {
			continue
		}

		headline := ""
		if card.Headline != nil {
			headline = card.Headline.Text
		}

		n := models.Notification{
			URN:       card.EntityURN,
			ID:        urnToID(card.EntityURN),
			Body:      headline,
			Read:      card.Read,
			CreatedAt: msToTime(card.PublishedAt),
		}
		result.Items = append(result.Items, n)
	}
	return result, nil
}

// MarkRead marks a single notification as read.
// TODO: Needs migration to new endpoint.
func (s *NotificationsService) MarkRead(notificationURN string) error {
	return fmt.Errorf("mark read is not yet supported (endpoint migration in progress)")
}

// MarkAllRead marks all notifications as read.
// TODO: Needs migration to new endpoint.
func (s *NotificationsService) MarkAllRead() error {
	return fmt.Errorf("mark all read is not yet supported (endpoint migration in progress)")
}

// GetBadgeCount returns the count of unread notifications.
func (s *NotificationsService) GetBadgeCount() (*models.NotificationBadge, error) {
	var raw struct {
		Data struct {
			Elements []struct {
				BadgingItem string `json:"badgingItem"`
				Count       int    `json:"count"`
			} `json:"elements"`
		} `json:"data"`
	}
	if err := s.c.Get(client.EndpointDashBadgingCounts, nil, &raw); err != nil {
		return nil, fmt.Errorf("get notification badge: %w", err)
	}
	for _, elem := range raw.Data.Elements {
		if elem.BadgingItem == "NOTIFICATIONS" {
			return &models.NotificationBadge{UnreadCount: elem.Count}, nil
		}
	}
	return &models.NotificationBadge{UnreadCount: 0}, nil
}
