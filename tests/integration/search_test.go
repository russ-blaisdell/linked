package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestSearchPeople(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SearchPeopleInput{
		Keywords: "software engineer",
		Count:    20,
	}

	result, err := li.Search.SearchPeople(input)
	if err != nil {
		t.Fatalf("SearchPeople() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one person result")
	}

	person := result.Items[0]
	if person.Profile.FirstName == "" {
		t.Error("person FirstName should not be empty")
	}
	if person.Distance == "" {
		t.Error("person Distance should not be empty")
	}
}

func TestSearchPeopleWithFilters(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SearchPeopleInput{
		Keywords: "product manager",
		Company:  "google",
		Network:  []string{"FIRST", "SECOND"},
		Count:    10,
	}

	result, err := li.Search.SearchPeople(input)
	if err != nil {
		t.Fatalf("SearchPeople() with filters error: %v", err)
	}
	_ = result
}

func TestSearchCompanies(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Search.SearchCompanies("Anthropic", 0, 20)
	if err != nil {
		t.Fatalf("SearchCompanies() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one company result")
	}

	co := result.Items[0]
	if co.Name == "" {
		t.Error("company name should not be empty")
	}
}

func TestSearchPosts(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Search.SearchPosts("AI safety", 0, 20)
	if err != nil {
		t.Fatalf("SearchPosts() error: %v", err)
	}
	_ = result
}

func TestSearchPeoplePagination(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	r1, err := li.Search.SearchPeople(models.SearchPeopleInput{Keywords: "engineer", Start: 0, Count: 5})
	if err != nil {
		t.Fatalf("first page error: %v", err)
	}

	r2, err := li.Search.SearchPeople(models.SearchPeopleInput{Keywords: "engineer", Start: 5, Count: 5})
	if err != nil {
		t.Fatalf("second page error: %v", err)
	}

	if r1.Pagination.Start != 0 {
		t.Errorf("page 1 start = %d, want 0", r1.Pagination.Start)
	}
	if r2.Pagination.Start != 5 {
		t.Errorf("page 2 start = %d, want 5", r2.Pagination.Start)
	}
}
