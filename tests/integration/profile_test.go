package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestGetMe(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	profile, err := li.Profile.GetMe()
	if err != nil {
		t.Fatalf("GetMe() error: %v", err)
	}

	if profile.ProfileID != "jane-doe" {
		t.Errorf("ProfileID = %q, want %q", profile.ProfileID, "jane-doe")
	}
	if profile.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", profile.FirstName, "Jane")
	}
	if profile.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", profile.LastName, "Doe")
	}
	if profile.URN != "urn:li:member:123456789" {
		t.Errorf("URN = %q, want %q", profile.URN, "urn:li:member:123456789")
	}
	if len(profile.Experience) == 0 {
		t.Error("expected at least one experience entry")
	}
	if len(profile.Skills) == 0 {
		t.Error("expected at least one skill")
	}
}

func TestGetProfile(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	profile, err := li.Profile.GetProfile("jane-doe")
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}

	if profile.ProfileID != "jane-doe" {
		t.Errorf("ProfileID = %q, want %q", profile.ProfileID, "jane-doe")
	}
}

func TestGetContactInfo(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	info, err := li.Profile.GetContactInfo("jane-doe")
	if err != nil {
		t.Fatalf("GetContactInfo() error: %v", err)
	}

	if info.ProfileID != "jane-doe" {
		t.Errorf("ProfileID = %q, want %q", info.ProfileID, "jane-doe")
	}
	if len(info.Emails) == 0 {
		t.Error("expected at least one email")
	}
}

func TestUpdateProfile(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	update := models.ProfileUpdate{
		Headline: "Principal Engineer",
		Summary:  "Building the future",
	}

	if err := li.Profile.UpdateProfile("jane-doe", update); err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
}

func TestUpdateProfileEmpty(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	// Empty update should be a no-op (no network call).
	update := models.ProfileUpdate{}
	if err := li.Profile.UpdateProfile("jane-doe", update); err != nil {
		t.Fatalf("UpdateProfile() with empty update error: %v", err)
	}
}

func TestProfileExperience(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	profile, err := li.Profile.GetMe()
	if err != nil {
		t.Fatalf("GetMe() error: %v", err)
	}

	if len(profile.Experience) < 2 {
		t.Fatalf("expected 2 experience entries, got %d", len(profile.Experience))
	}

	current := profile.Experience[0]
	if current.Title != "Senior Software Engineer" {
		t.Errorf("first experience Title = %q, want %q", current.Title, "Senior Software Engineer")
	}
	if !current.Current {
		t.Error("first experience should be current (no end date)")
	}

	past := profile.Experience[1]
	if past.EndDate == "" {
		t.Error("second experience should have an end date")
	}
	if past.Current {
		t.Error("second experience should not be current")
	}
}

func TestProfileEducation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	profile, err := li.Profile.GetMe()
	if err != nil {
		t.Fatalf("GetMe() error: %v", err)
	}

	if len(profile.Education) == 0 {
		t.Fatal("expected education entries")
	}

	edu := profile.Education[0]
	if edu.SchoolName != "MIT" {
		t.Errorf("SchoolName = %q, want %q", edu.SchoolName, "MIT")
	}
	if edu.Degree != "B.S." {
		t.Errorf("Degree = %q, want %q", edu.Degree, "B.S.")
	}
}

func TestGetWhoViewed(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Profile.GetWhoViewed(0, 10)
	if err != nil {
		t.Fatalf("GetWhoViewed() error: %v", err)
	}

	// Viewer count comes from the analytics view endpoint (42 in fixture).
	if result.Pagination.Total != 42 {
		t.Errorf("Total = %d, want 42", result.Pagination.Total)
	}

	// Individual viewers come from the analytics object endpoint (2 in fixture).
	if len(result.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(result.Items))
	}

	first := result.Items[0]
	if first.Profile.ProfileID != "alice-smith" {
		t.Errorf("first viewer ProfileID = %q, want %q", first.Profile.ProfileID, "alice-smith")
	}
	if first.Profile.FirstName != "Alice" {
		t.Errorf("first viewer FirstName = %q, want %q", first.Profile.FirstName, "Alice")
	}
	if first.ViewedAt == "" {
		t.Error("first viewer ViewedAt should not be empty")
	}

	second := result.Items[1]
	if second.Profile.ProfileID != "bob-jones" {
		t.Errorf("second viewer ProfileID = %q, want %q", second.Profile.ProfileID, "bob-jones")
	}
}
