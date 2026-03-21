package api

import (
	"encoding/json"
	"fmt"
	"strings"

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
	EntityURN  string             `json:"entityUrn"`
	ID         string             `json:"id,omitempty"`
	FromMember voyagerMiniProfile `json:"fromMember,omitempty"`
	ToMember   voyagerMiniProfile `json:"toMember,omitempty"`
	Message    string             `json:"message,omitempty"`
	SentTime   int64              `json:"sentTime,omitempty"`
	Status     string             `json:"status,omitempty"`
}

// ListConnections returns all 1st-degree connections.
// Uses the dash connections endpoint + individual profile lookups.
func (s *ConnectionsService) ListConnections(start, count int) (*models.PagedConnections, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":        "search",
		"start":    fmt.Sprintf("%d", start),
		"count":    fmt.Sprintf("%d", count),
		"sortType": "RECENTLY_ADDED",
	}

	var raw struct {
		Data struct {
			Paging struct {
				Start int `json:"start"`
				Count int `json:"count"`
			} `json:"paging"`
			Elements []string `json:"*elements"`
		} `json:"data"`
		Included []json.RawMessage `json:"included"`
	}

	if err := s.c.Get(client.EndpointDashConnections, params, &raw); err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
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

	// Extract connection entities and collect member IDs to resolve.
	type connInfo struct {
		entityURN string
		memberID  string
		createdAt int64
	}
	var connections []connInfo
	for _, urn := range raw.Data.Elements {
		connRaw, ok := byURN[urn]
		if !ok {
			continue
		}
		var conn struct {
			EntityURN       string `json:"entityUrn"`
			ConnectedMember string `json:"connectedMember"`
			CreatedAt       int64  `json:"createdAt"`
		}
		if json.Unmarshal(connRaw, &conn) != nil {
			continue
		}
		memberID := urnToID(conn.ConnectedMember)
		connections = append(connections, connInfo{
			entityURN: conn.EntityURN,
			memberID:  memberID,
			createdAt: conn.CreatedAt,
		})
	}

	// Resolve profile names via individual dash profile lookups.
	type resolvedProfile struct {
		firstName, lastName, headline, profileID string
	}
	profiles := make(map[string]resolvedProfile)
	for _, ci := range connections {
		if ci.memberID == "" {
			continue
		}
		var profRaw struct {
			Included []json.RawMessage `json:"included"`
		}
		profParams := map[string]string{
			"q":              "memberIdentity",
			"memberIdentity": ci.memberID,
		}
		if err := s.c.Get(client.EndpointDashProfiles, profParams, &profRaw); err != nil {
			continue
		}
		for _, inc := range profRaw.Included {
			var p struct {
				FirstName        string `json:"firstName"`
				LastName         string `json:"lastName"`
				Headline         string `json:"headline"`
				PublicIdentifier string `json:"publicIdentifier"`
			}
			if json.Unmarshal(inc, &p) == nil && p.FirstName != "" {
				profiles[ci.memberID] = resolvedProfile{
					firstName: p.FirstName,
					lastName:  p.LastName,
					headline:  p.Headline,
					profileID: p.PublicIdentifier,
				}
				break
			}
		}
	}

	result := &models.PagedConnections{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   len(connections),
			HasMore: len(connections) >= count,
		},
	}
	for _, ci := range connections {
		p := profiles[ci.memberID]
		result.Items = append(result.Items, models.Connection{
			URN:       ci.entityURN,
			ProfileID: p.profileID,
			FirstName: p.firstName,
			LastName:  p.lastName,
			Headline:  p.headline,
		})
	}

	return result, nil
}

// getInvitationSummaryCount fetches the true total of pending invitations via
// the lightweight summary endpoint (avoids undercounting from paged results).
func (s *ConnectionsService) getInvitationSummaryCount() int {
	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(types:List(PENDING_INVITATION_COUNT,UNSEEN_INVITATION_COUNT))&queryId=%s",
		client.EndpointGraphQL,
		client.EndpointInvitationsSummaryQueryID,
	)

	var raw struct {
		Data *struct {
			Summary *struct {
				Elements []struct {
					NumPendingInvitations int `json:"numPendingInvitations"`
				} `json:"elements"`
			} `json:"relationshipsDashInvitationsSummaryByInvitationSummaryTypes"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return 0
	}
	if raw.Data != nil && raw.Data.Summary != nil && len(raw.Data.Summary.Elements) > 0 {
		return raw.Data.Summary.Elements[0].NumPendingInvitations
	}
	return 0
}

// ListPendingInvitations returns received connection invitations awaiting action.
// Uses the GraphQL endpoint (the old REST endpoint returns empty results).
func (s *ConnectionsService) ListPendingInvitations(start, count int) (*models.PagedInvitations, error) {
	if count == 0 {
		count = client.DefaultCount
	}

	// Get the true total from the summary endpoint first.
	summaryTotal := s.getInvitationSummaryCount()

	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(includeInsights:true,q:receivedInvitation,start:%d,count:%d)&queryId=%s",
		client.EndpointGraphQL, start, count,
		client.EndpointInvitationViewsQueryID,
	)

	var raw struct {
		Data *struct {
			Collection *graphqlInvitationCollection `json:"relationshipsDashInvitationViewsByReceived"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}

	var col *graphqlInvitationCollection
	if raw.Data != nil {
		col = raw.Data.Collection
	}

	result := parseInvitationCollection(col, start, count, "INBOUND", nil)
	if summaryTotal > result.Pagination.Total {
		result.Pagination.Total = summaryTotal
		result.Pagination.HasMore = (start + len(result.Items)) < summaryTotal
	}
	return result, nil
}

// ListSentInvitations returns outbound connection invitations.
// Uses the GraphQL sent invitation views endpoint.
func (s *ConnectionsService) ListSentInvitations(start, count int) (*models.PagedInvitations, error) {
	if count == 0 {
		count = client.DefaultCount
	}

	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(invitationType:CONNECTION,start:%d,count:%d)&queryId=%s",
		client.EndpointGraphQL, start, count,
		client.EndpointSentInvitationViewsQueryID,
	)

	var raw struct {
		Data *struct {
			Collection *graphqlInvitationCollection `json:"relationshipsDashSentInvitationViewsByInvitationType"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("list sent invitations: %w", err)
	}

	var col *graphqlInvitationCollection
	if raw.Data != nil {
		col = raw.Data.Collection
	}
	return parseInvitationCollection(col, start, count, "OUTBOUND", nil), nil
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
		"invitationId": invitationID,
		"sharedSecret": sharedSecret,
		"action":       "accept",
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

// RemoveConnection removes an existing 1st-degree connection.
func (s *ConnectionsService) RemoveConnection(profileURN string) error {
	path := fmt.Sprintf(client.EndpointConnection, urnToID(profileURN))
	return s.c.Delete(path)
}

// GetMutualConnections returns shared connections between the authenticated user and another member.
func (s *ConnectionsService) GetMutualConnections(profileURN string, start, count int) (*models.PagedMutualConnections, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":         "memberAndConnection",
		"memberUrn": profileURN,
		"start":     fmt.Sprintf("%d", start),
		"count":     fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			MiniProfile voyagerMiniProfile `json:"miniProfile,omitempty"`
			Count       int                `json:"count,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointMutualConnections, params, &raw); err != nil {
		return nil, fmt.Errorf("get mutual connections: %w", err)
	}

	result := &models.PagedMutualConnections{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		mp := el.MiniProfile
		result.Items = append(result.Items, models.MutualConnection{
			Profile: models.Profile{
				URN:       mp.EntityURN,
				ProfileID: mp.PublicID,
				FirstName: mp.FirstName,
				LastName:  mp.LastName,
				Headline:  mp.Occupation,
			},
			Count: el.Count,
		})
	}
	return result, nil
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

// GraphQL invitation response types (Dash model).
type graphqlInvitationCollection struct {
	Paging   struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
	Elements []graphqlInvitationView `json:"elements"`
}

type graphqlInvitationView struct {
	Invitation    *graphqlInvitation `json:"invitation"`
	Title         *graphqlText       `json:"title"`
	Subtitle      *graphqlText       `json:"subtitle"`
	Insight       *graphqlInsight    `json:"insight"`
	SentTimeLabel string             `json:"sentTimeLabel"`
}

type graphqlInsight struct {
	Text                                       *graphqlText `json:"text"`
	SharedConnectionDetailTargetResolutionResult *struct {
		EntityURN string `json:"entityUrn"`
	} `json:"sharedConnectionDetailTargetResolutionResult"`
}

type graphqlInvitation struct {
	EntityURN    string `json:"entityUrn"`
	InvitationID int64  `json:"invitationId"`
	Message      string `json:"message"`
	SharedSecret string `json:"sharedSecret"`
	Inviter      *struct {
		MemberProfileURN *struct {
			EntityURN        string `json:"entityUrn"`
			ObjectURN        string `json:"objectUrn"`
			PublicIdentifier string `json:"publicIdentifier"`
			FirstName        string `json:"firstName"`
			LastName         string `json:"lastName"`
		} `json:"memberProfileUrn"`
	} `json:"genericInviter"`
}

type graphqlText struct {
	Text string `json:"text"`
}

// profileNameResolver looks up a profile URN and returns "FirstName LastName",
// or empty string on failure. Used to resolve single-mutual-connection names.
type profileNameResolver func(profileURN string) string

// parseInvitationCollection converts elements from the GraphQL response.
// If resolver is non-nil, it is called to resolve "1 mutual connection" insight
// text into the actual connection name (e.g. "Jonathan Cook is a mutual connection").
func parseInvitationCollection(col *graphqlInvitationCollection, start, count int, direction string, resolver profileNameResolver) *models.PagedInvitations {
	if col == nil {
		return &models.PagedInvitations{
			Pagination: models.Pagination{Start: start, Count: count},
		}
	}

	total := col.Paging.Total
	// LinkedIn sometimes reports total=0 even when elements exist.
	if total == 0 && len(col.Elements) > 0 {
		total = len(col.Elements)
	}

	result := &models.PagedInvitations{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   total,
			HasMore: (start + len(col.Elements)) < total,
		},
	}

	// Track invitations that need mutual connection name resolution.
	needsResolve := make(map[int]string)

	for _, view := range col.Elements {
		if view.Invitation == nil {
			continue
		}

		inv := models.Invitation{
			URN:       view.Invitation.EntityURN,
			ID:        fmt.Sprintf("%d", view.Invitation.InvitationID),
			Direction: direction,
			Status:    "PENDING",
			Message:   view.Invitation.Message,
			SentAt:    view.SentTimeLabel,
		}

		// Inviter name from title, headline from subtitle.
		if view.Title != nil {
			inv.FromProfile.FirstName = strings.TrimSpace(view.Title.Text)
		}
		if view.Subtitle != nil {
			inv.FromProfile.Headline = view.Subtitle.Text
		}

		// Richer profile data from the inline genericInviter.
		if gi := view.Invitation.Inviter; gi != nil && gi.MemberProfileURN != nil {
			mp := gi.MemberProfileURN
			inv.FromProfile = models.Profile{
				URN:       mp.EntityURN,
				ProfileID: mp.PublicIdentifier,
				FirstName: mp.FirstName,
				LastName:  mp.LastName,
			}
		}
		// Headline comes from subtitle, not the profile object.
		if view.Subtitle != nil {
			inv.FromProfile.Headline = view.Subtitle.Text
		}
		// Mutual connection insight (e.g. "1 mutual connection" or "Rob Tanzola and 2 others").
		if view.Insight != nil && view.Insight.Text != nil {
			inv.Insight = view.Insight.Text.Text
			// For exactly "1 mutual connection", mark for name resolution.
			if inv.Insight == "1 mutual connection" &&
				view.Insight.SharedConnectionDetailTargetResolutionResult != nil &&
				view.Insight.SharedConnectionDetailTargetResolutionResult.EntityURN != "" {
				needsResolve[len(result.Items)] = view.Insight.SharedConnectionDetailTargetResolutionResult.EntityURN
			}
		}

		result.Items = append(result.Items, inv)
	}

	// Resolve single-mutual connection names.
	if resolver != nil {
		for idx, urn := range needsResolve {
			if name := resolver(urn); name != "" {
				result.Items[idx].Insight = name + " is a mutual connection"
			}
		}
	}

	return result
}
