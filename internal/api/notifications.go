package api

import (
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
func (s *NotificationsService) List(start, count int) (*models.PagedNotifications, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "me",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			EntityURN   string `json:"entityUrn"`
			ReadAt      *int64 `json:"readAt,omitempty"`
			CreatedAt   int64  `json:"createdAt"`
			NotificationType string `json:"notificationType,omitempty"`
			HeadlineText struct {
				Text string `json:"text"`
			} `json:"headlineText,omitempty"`
			EntityEmbeddedObject struct {
				Urn string `json:"urn,omitempty"`
			} `json:"entityEmbeddedObject,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointNotifications, params, &raw); err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	result := &models.PagedNotifications{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		n := models.Notification{
			URN:       el.EntityURN,
			ID:        urnToID(el.EntityURN),
			Type:      el.NotificationType,
			Body:      el.HeadlineText.Text,
			Read:      el.ReadAt != nil,
			CreatedAt: msToTime(el.CreatedAt),
			EntityURN: el.EntityEmbeddedObject.Urn,
		}
		result.Items = append(result.Items, n)
	}
	return result, nil
}

// MarkRead marks a notification as read.
func (s *NotificationsService) MarkRead(notificationURN string) error {
	path := fmt.Sprintf("%s/%s", client.EndpointNotifications, urnToID(notificationURN))
	return s.c.Put(path, map[string]interface{}{"read": true}, nil)
}
