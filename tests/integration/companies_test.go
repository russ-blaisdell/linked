package integration_test

import "testing"

func TestGetCompany(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	co, err := li.Companies.GetCompany("anthropic")
	if err != nil {
		t.Fatalf("GetCompany() error: %v", err)
	}

	if co.Name != "Anthropic" {
		t.Errorf("Name = %q, want %q", co.Name, "Anthropic")
	}
	if co.ID != "anthropic" {
		t.Errorf("ID = %q, want %q", co.ID, "anthropic")
	}
	if co.Industry == "" {
		t.Error("Industry should not be empty")
	}
}

func TestFollowUnfollowCompany(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	urn := "urn:li:company:9999"
	if err := li.Companies.FollowCompany(urn); err != nil {
		t.Fatalf("FollowCompany() error: %v", err)
	}
	if err := li.Companies.UnfollowCompany(urn); err != nil {
		t.Fatalf("UnfollowCompany() error: %v", err)
	}
}

func TestGetCompanyPosts(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Companies.GetCompanyPosts("urn:li:company:9999", 0, 20)
	if err != nil {
		t.Fatalf("GetCompanyPosts() error: %v", err)
	}
	_ = result
}
