package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestListReceivedRecommendations(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Recommendations.ListReceived(0, 20)
	if err != nil {
		t.Fatalf("ListReceived() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one received recommendation")
	}

	rec := result.Items[0]
	if rec.Body == "" {
		t.Error("recommendation body should not be empty")
	}
	if rec.RecommenderProfile.FirstName == "" {
		t.Error("recommender FirstName should not be empty")
	}
	if rec.Status != "VISIBLE" {
		t.Errorf("Status = %q, want VISIBLE", rec.Status)
	}
}

func TestListGivenRecommendations(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Recommendations.ListGiven(0, 20)
	if err != nil {
		t.Fatalf("ListGiven() error: %v", err)
	}
	_ = result
}

func TestRequestRecommendation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.RecommendationRequestInput{
		RecipientProfileURN: "urn:li:member:111000001",
		Message:             "Hi Bob, would you be willing to write me a recommendation?",
		Relationship:        "COLLEAGUE",
	}

	if err := li.Recommendations.RequestRecommendation(input); err != nil {
		t.Fatalf("RequestRecommendation() error: %v", err)
	}
}

func TestHideShowRecommendation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	urn := "urn:li:recommendation:rec001"

	if err := li.Recommendations.HideRecommendation(urn); err != nil {
		t.Fatalf("HideRecommendation() error: %v", err)
	}
	if err := li.Recommendations.ShowRecommendation(urn); err != nil {
		t.Fatalf("ShowRecommendation() error: %v", err)
	}
}
