package api

import (
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// ConnectionsService handles LinkedIn connection operations.
type ConnectionsService struct {
	c *client.Client
}

// NewConnectionsService returns a new ConnectionsService.
func NewConnectionsService(c *client.Client) *ConnectionsService {
	return &ConnectionsService{c: c}
}

type voyagerConnection struct {
	EntityURN   string             `json:"entityUrn"`
	MiniProfile voyagerMiniProfile `json:"miniProfile"`
	ConnectedAt int64              `json:"connectedAt"`
}

type voyagerInvitation struct {
	EntityURN   string             `json:"entityUrn"`
	ID          string             `json:"id,omitempty"`
	FromMember  voyagerMiniProfile `json:"fromMember,omitempty"`
	ToMember    voyagerMiniProfile `json:"toMember,omitempty"`
	Message     string             `json:"message,omitempty"`
	SentTime    int64              `json:"sentTime,omitempty"`
	Status      string             `json:"status,omitempty"`
}

// ListConnections returns all 1st-degree connections.
func (s *ConnectionsService) ListConnections(start, count int) (*models.PagedConnections, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":          "viewer",
		"start":      fmt.Sprintf("%d", start),
		"count":      fmt.Sprintf("%d", count),
		"decoration": "(mini_profile~(entityUrn,firstName,lastName,occupation,publicIdentifier,picture~(rootUrl,artifacts~(fileIdentifyingUrlPathSegment))))",
	}

	var raw struct {
		Elements []voyagerConnection `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointConnections, params, &raw); err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}

	result := &models.PagedConnections{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}

	for _, vc := range raw.Elements {
		mp := vc.MiniProfile
		result.Items = append(result.Items, models.Connection{
			URN:       vc.EntityURN,
			ProfileID: mp.PublicID,
			FirstName: mp.FirstName,
			LastName:  mp.LastName,
			Headline:  mp.Occupation,
		})
	}

	return result, nil
}

// ListPendingInvitations returns received connection invitations awaiting action.
func (s *ConnectionsService) ListPendingInvitations(start, count int) (*models.PagedInvitations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":           "receivedInvitation",
		"invitationType": "CONNECTION",
		"start":       fmt.Sprintf("%d", start),
		"count":       fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerInvitation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointInvitations, params, &raw); err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}

	return mapInvitations(raw.Elements, "INBOUND", raw.Paging.Start, raw.Paging.Count, raw.Paging.Total), nil
}

// ListSentInvitations returns outbound connection invitations.
func (s *ConnectionsService) ListSentInvitations(start, count int) (*models.PagedInvitations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "sentInvitation",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerInvitation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointSentInvitations, params, &raw); err != nil {
		return nil, fmt.Errorf("list sent invitations: %w", err)
	}

	return mapInvitations(raw.Elements, "OUTBOUND", raw.Paging.Start, raw.Paging.Count, raw.Paging.Total), nil
}

// SendConnectionRequest sends a connection request to a profile.
func (s *ConnectionsService) SendConnectionRequest(input models.ConnectionRequestInput) error {
	payload := map[string]interface{}{
		"invitee": map[string]interface{}{
			"com.linkedin.voyager.growth.invitation.InviteeProfile": map[string]interface{}{
				"profileId": urnToID(input.ProfileURN),
			},
		},
		"trackingID": generateTrackingID(),
	}
	if input.Note != "" {
		payload["message"] = input.Note
	}
	return s.c.Post(client.EndpointConnections, payload, nil)
}

// AcceptInvitation accepts a received connection invitation.
func (s *ConnectionsService) AcceptInvitation(invitationID, sharedSecret string) error {
	payload := map[string]interface{}{
		"invitationId":   invitationID,
		"sharedSecret":   sharedSecret,
		"action":         "accept",
	}
	return s.c.Post(fmt.Sprintf(client.EndpointInvitationHandle, invitationID), payload, nil)
}

// IgnoreInvitation ignores a received connection invitation.
func (s *ConnectionsService) IgnoreInvitation(invitationID string) error {
	payload := map[string]interface{}{
		"invitationId": invitationID,
		"action":       "ignore",
	}
	return s.c.Post(fmt.Sprintf(client.EndpointInvitationHandle, invitationID), payload, nil)
}

// WithdrawInvitation withdraws a sent connection invitation.
func (s *ConnectionsService) WithdrawInvitation(invitationURN string) error {
	return s.c.Delete(fmt.Sprintf(client.EndpointInvitationHandle, urnToID(invitationURN)))
}

// FollowProfile follows a LinkedIn member (without connecting).
func (s *ConnectionsService) FollowProfile(profileURN string) error {
	return s.c.Post(client.EndpointFollowEntity, map[string]interface{}{
		"followedEntityUrn": profileURN,
	}, nil)
}

// UnfollowProfile unfollows a LinkedIn member.
func (s *ConnectionsService) UnfollowProfile(profileURN string) error {
	return s.c.Delete(fmt.Sprintf("%s/%s", client.EndpointFollowEntity, urnToID(profileURN)))
}

// mapInvitations converts raw invitations to the paged model.
func mapInvitations(elements []voyagerInvitation, direction string, start, count, total int) *models.PagedInvitations {
	result := &models.PagedInvitations{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   total,
			HasMore: (start + count) < total,
		},
	}
	for _, vi := range elements {
		inv := models.Invitation{
			URN:       vi.EntityURN,
			ID:        vi.ID,
			Message:   vi.Message,
			Status:    vi.Status,
			Direction: direction,
			SentAt:    msToTime(vi.SentTime),
		}
		inv.FromProfile = models.Profile{
			URN:       vi.FromMember.EntityURN,
			ProfileID: vi.FromMember.PublicID,
			FirstName: vi.FromMember.FirstName,
			LastName:  vi.FromMember.LastName,
			Headline:  vi.FromMember.Occupation,
		}
		inv.ToProfile = models.Profile{
			URN:       vi.ToMember.EntityURN,
			ProfileID: vi.ToMember.PublicID,
			FirstName: vi.ToMember.FirstName,
			LastName:  vi.ToMember.LastName,
			Headline:  vi.ToMember.Occupation,
		}
		result.Items = append(result.Items, inv)
	}
	return result
}
