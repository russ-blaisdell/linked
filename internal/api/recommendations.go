package api

import (
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// RecommendationsService handles LinkedIn recommendation operations.
type RecommendationsService struct {
	c *client.Client
}

// NewRecommendationsService returns a new RecommendationsService.
func NewRecommendationsService(c *client.Client) *RecommendationsService {
	return &RecommendationsService{c: c}
}

type voyagerRecommendation struct {
	EntityURN   string             `json:"entityUrn"`
	ID          string             `json:"id,omitempty"`
	Recommender voyagerMiniProfile `json:"recommender,omitempty"`
	Recommendee voyagerMiniProfile `json:"recommendee,omitempty"`
	Relationship string            `json:"relationship,omitempty"`
	RecommendationText string      `json:"recommendationText,omitempty"`
	CreatedAt   int64              `json:"createdAt,omitempty"`
	Status      string             `json:"status,omitempty"`
}

// ListReceived returns recommendations received by the authenticated user.
func (s *RecommendationsService) ListReceived(start, count int) (*models.PagedRecommendations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "received",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerRecommendation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointRecommendations, params, &raw); err != nil {
		return nil, fmt.Errorf("list received recommendations: %w", err)
	}

	return mapPagedRecommendations(raw.Elements, raw.Paging.Start, raw.Paging.Count, raw.Paging.Total), nil
}

// ListGiven returns recommendations the authenticated user has written.
func (s *RecommendationsService) ListGiven(start, count int) (*models.PagedRecommendations, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "given",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []voyagerRecommendation `json:"elements"`
		Paging   struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointRecommendations, params, &raw); err != nil {
		return nil, fmt.Errorf("list given recommendations: %w", err)
	}

	return mapPagedRecommendations(raw.Elements, raw.Paging.Start, raw.Paging.Count, raw.Paging.Total), nil
}

// RequestRecommendation sends a recommendation request to a connection.
func (s *RecommendationsService) RequestRecommendation(input models.RecommendationRequestInput) error {
	payload := map[string]interface{}{
		"recommendee": map[string]interface{}{
			"entityUrn": input.RecipientProfileURN,
		},
		"message":      input.Message,
		"relationship": input.Relationship,
	}
	return s.c.Post(client.EndpointRecommendations, payload, nil)
}

// HideRecommendation hides a received recommendation from the profile.
func (s *RecommendationsService) HideRecommendation(recommendationURN string) error {
	path := fmt.Sprintf("%s/%s", client.EndpointRecommendations, urnToID(recommendationURN))
	return s.c.Put(path, map[string]interface{}{"status": "HIDDEN"}, nil)
}

// ShowRecommendation makes a previously hidden recommendation visible again.
func (s *RecommendationsService) ShowRecommendation(recommendationURN string) error {
	path := fmt.Sprintf("%s/%s", client.EndpointRecommendations, urnToID(recommendationURN))
	return s.c.Put(path, map[string]interface{}{"status": "VISIBLE"}, nil)
}

// mapPagedRecommendations converts raw recommendations to the paged model.
func mapPagedRecommendations(elements []voyagerRecommendation, start, count, total int) *models.PagedRecommendations {
	result := &models.PagedRecommendations{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   total,
			HasMore: (start + count) < total,
		},
	}
	for _, vr := range elements {
		result.Items = append(result.Items, models.Recommendation{
			URN: vr.EntityURN,
			ID:  vr.ID,
			RecommenderProfile: models.Profile{
				URN:       vr.Recommender.EntityURN,
				ProfileID: vr.Recommender.PublicID,
				FirstName: vr.Recommender.FirstName,
				LastName:  vr.Recommender.LastName,
				Headline:  vr.Recommender.Occupation,
			},
			RecommendeeProfile: models.Profile{
				URN:       vr.Recommendee.EntityURN,
				ProfileID: vr.Recommendee.PublicID,
				FirstName: vr.Recommendee.FirstName,
				LastName:  vr.Recommendee.LastName,
				Headline:  vr.Recommendee.Occupation,
			},
			Body:         vr.RecommendationText,
			Relationship: vr.Relationship,
			Status:       vr.Status,
			CreatedAt:    msToTime(vr.CreatedAt),
		})
	}
	return result
}
