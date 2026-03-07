package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestGetJob(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	job, err := li.Jobs.GetJob("987654321")
	if err != nil {
		t.Fatalf("GetJob() error: %v", err)
	}

	if job.Title != "Senior Go Engineer" {
		t.Errorf("Title = %q, want %q", job.Title, "Senior Go Engineer")
	}
	if job.Company.Name != "TechCorp" {
		t.Errorf("Company.Name = %q, want %q", job.Company.Name, "TechCorp")
	}
	if !job.Remote {
		t.Error("job should be remote")
	}
	if job.Description == "" {
		t.Error("job description should not be empty")
	}
	if job.PostedAt == "" {
		t.Error("PostedAt should not be empty")
	}
}

func TestListSavedJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Jobs.ListSavedJobs(0, 20)
	if err != nil {
		t.Fatalf("ListSavedJobs() error: %v", err)
	}

	for _, j := range result.Items {
		if !j.Saved {
			t.Errorf("job %s should be marked saved", j.ID)
		}
	}
}

func TestSaveUnsaveJob(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Jobs.SaveJob("987654321"); err != nil {
		t.Fatalf("SaveJob() error: %v", err)
	}
	if err := li.Jobs.UnsaveJob("987654321"); err != nil {
		t.Fatalf("UnsaveJob() error: %v", err)
	}
}

func TestListAppliedJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Jobs.ListAppliedJobs(0, 20)
	if err != nil {
		t.Fatalf("ListAppliedJobs() error: %v", err)
	}
	// applied jobs fixture is empty — just verify no error
	_ = result
}

func TestSearchJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SearchJobsInput{
		Keywords: "golang engineer",
		Remote:   true,
	}

	result, err := li.Search.SearchJobs(input)
	if err != nil {
		t.Fatalf("SearchJobs() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one job result")
	}

	job := result.Items[0]
	if job.Title == "" {
		t.Error("job title should not be empty")
	}
}
