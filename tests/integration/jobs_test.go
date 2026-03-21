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
	if job.Location != "San Francisco, CA" {
		t.Errorf("Location = %q, want %q", job.Location, "San Francisco, CA")
	}
	if job.EmploymentType != "Full-time" {
		t.Errorf("EmploymentType = %q, want %q", job.EmploymentType, "Full-time")
	}
	if job.ApplyURL == "" {
		t.Error("ApplyURL should not be empty")
	}
}

func TestListSavedJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Jobs.ListSavedJobs(0, 20)
	if err != nil {
		t.Fatalf("ListSavedJobs() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one saved job")
	}

	for _, j := range result.Items {
		if !j.Saved {
			t.Errorf("job %s should be marked saved", j.ID)
		}
	}

	job := result.Items[0]
	if job.Title != "Senior Go Engineer" {
		t.Errorf("Title = %q, want %q", job.Title, "Senior Go Engineer")
	}
	if job.Company.Name != "TechCorp" {
		t.Errorf("Company.Name = %q, want %q", job.Company.Name, "TechCorp")
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
	// applied jobs fixture is empty -- just verify no error
	if len(result.Items) != 0 {
		t.Errorf("expected 0 applied jobs, got %d", len(result.Items))
	}
}

func TestGetRecommendedJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Jobs.GetRecommendedJobs(0, 20)
	if err != nil {
		t.Fatalf("GetRecommendedJobs() error: %v", err)
	}

	if len(result.Items) < 2 {
		t.Fatalf("expected at least 2 recommended jobs, got %d", len(result.Items))
	}

	job := result.Items[0]
	if job.Title != "Senior Go Engineer" {
		t.Errorf("Title = %q, want %q", job.Title, "Senior Go Engineer")
	}
	if job.ID != "987654321" {
		t.Errorf("ID = %q, want %q", job.ID, "987654321")
	}

	job2 := result.Items[1]
	if job2.Title != "Backend Developer" {
		t.Errorf("Title = %q, want %q", job2.Title, "Backend Developer")
	}
}

func TestSearchJobs(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.SearchJobsInput{
		Keywords: "golang engineer",
	}

	_, err := li.Search.SearchJobs(input)
	if err == nil {
		t.Fatal("expected error for unsupported job search")
	}
}
