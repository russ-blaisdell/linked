package integration_test

import (
	"strings"
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
		t.Error("person name should not be empty")
	}
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

	if len(result.Items) == 0 {
		t.Fatal("expected at least one post result")
	}
	if result.Items[0].Body == "" {
		t.Error("post body should not be empty")
	}
}

func TestSearchJobsNotAvailable(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	_, err := li.Search.SearchJobs(models.SearchJobsInput{Keywords: "golang"})
	if err == nil {
		t.Fatal("expected error for unsupported job search")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}
