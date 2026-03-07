package api

import (
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// JobsService handles LinkedIn job operations.
type JobsService struct {
	c *client.Client
}

// NewJobsService returns a new JobsService.
func NewJobsService(c *client.Client) *JobsService {
	return &JobsService{c: c}
}

type voyagerJobPosting struct {
	EntityURN   string `json:"entityUrn"`
	Title       string `json:"title"`
	Description struct {
		Text string `json:"text"`
	} `json:"description,omitempty"`
	CompanyDetails struct {
		Company struct {
			EntityURN     string `json:"entityUrn"`
			Name          string `json:"name"`
			UniversalName string `json:"universalName"`
		} `json:"com.linkedin.voyager.jobs.JobPostingCompany"`
	} `json:"companyDetails"`
	FormattedLocation string `json:"formattedLocation,omitempty"`
	WorkRemoteAllowed bool   `json:"workRemoteAllowed,omitempty"`
	WorkplaceTypes    []struct {
		TypeURN string `json:"workplaceTypeUrn"`
	} `json:"workplaceTypes,omitempty"`
	ListedAt    int64 `json:"listedAt"`
	ExpireAt    int64 `json:"expireAt,omitempty"`
	ApplyMethod struct {
		ExternalURL string `json:"com.linkedin.voyager.jobs.OffsiteApply,omitempty"`
	} `json:"applyMethod,omitempty"`
	JobState         string `json:"jobState,omitempty"`
	EmploymentStatus string `json:"formattedEmploymentStatus,omitempty"`
	ExperienceLevel  string `json:"formattedExperienceLevel,omitempty"`
}

// GetJob returns details for a specific job posting.
func (s *JobsService) GetJob(jobID string) (*models.Job, error) {
	path := fmt.Sprintf(client.EndpointJobPosting, jobID)
	var raw voyagerJobPosting
	if err := s.c.Get(path, nil, &raw); err != nil {
		return nil, fmt.Errorf("get job %q: %w", jobID, err)
	}
	return mapVoyagerJob(raw), nil
}

// ListSavedJobs returns the authenticated user's saved jobs.
func (s *JobsService) ListSavedJobs(start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "savedJob",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			SavedJobURN string            `json:"savedJobUrn,omitempty"`
			Job         voyagerJobPosting `json:"job,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointSavedJobs, params, &raw); err != nil {
		return nil, fmt.Errorf("list saved jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		j := mapVoyagerJob(el.Job)
		j.Saved = true
		result.Items = append(result.Items, *j)
	}
	return result, nil
}

// SaveJob saves a job for the authenticated user.
func (s *JobsService) SaveJob(jobID string) error {
	return s.c.Post(client.EndpointSavedJobs, map[string]interface{}{
		"jobId": jobID,
	}, nil)
}

// UnsaveJob removes a job from the user's saved jobs.
func (s *JobsService) UnsaveJob(jobID string) error {
	return s.c.Delete(fmt.Sprintf("%s/%s", client.EndpointSavedJobs, jobID))
}

// ListAppliedJobs returns jobs the user has applied to.
func (s *JobsService) ListAppliedJobs(start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "appliedJob",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			Job voyagerJobPosting `json:"job,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointAppliedJobs, params, &raw); err != nil {
		return nil, fmt.Errorf("list applied jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		j := mapVoyagerJob(el.Job)
		j.Applied = true
		result.Items = append(result.Items, *j)
	}
	return result, nil
}

// GetRecommendedJobs returns LinkedIn's recommended jobs for the authenticated user.
func (s *JobsService) GetRecommendedJobs(start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "recommendedJobs",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			JobPosting voyagerJobPosting `json:"jobPosting,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointJobRecommendations, params, &raw); err != nil {
		return nil, fmt.Errorf("get recommended jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		if el.JobPosting.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, *mapVoyagerJob(el.JobPosting))
	}
	return result, nil
}

// SearchJobsByCompany returns job postings for a specific company.
func (s *JobsService) SearchJobsByCompany(companyURN string, start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":       "jobSearch",
		"filters": fmt.Sprintf("List(company->%s)", urnToID(companyURN)),
		"start":   fmt.Sprintf("%d", start),
		"count":   fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			JobCardUnion struct {
				JobPostingCard struct {
					EntityURN string `json:"entityUrn"`
					Title     string `json:"title"`
					Company   struct {
						Name          string `json:"name"`
						UniversalName string `json:"universalName"`
					} `json:"company"`
					FormattedLocation string `json:"formattedLocation"`
					WorkRemoteAllowed bool   `json:"workRemoteAllowed"`
					PostedAt          int64  `json:"listedAt"`
				} `json:"com.linkedin.voyager.jobs.JobPostingCard"`
			} `json:"jobCardUnion"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging,omitempty"`
	}

	if err := s.c.Get(client.EndpointJobSearchDash, params, &raw); err != nil {
		return nil, fmt.Errorf("search jobs by company: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		card := el.JobCardUnion.JobPostingCard
		if card.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, models.Job{
			URN:      card.EntityURN,
			ID:       urnToID(card.EntityURN),
			Title:    card.Title,
			Company:  models.Company{Name: card.Company.Name, ID: card.Company.UniversalName},
			Location: card.FormattedLocation,
			Remote:   card.WorkRemoteAllowed,
			PostedAt: msToTime(card.PostedAt),
		})
	}
	return result, nil
}

// mapVoyagerJob converts a raw voyagerJobPosting to models.Job.
func mapVoyagerJob(vj voyagerJobPosting) *models.Job {
	co := vj.CompanyDetails.Company
	j := &models.Job{
		URN:             vj.EntityURN,
		ID:              urnToID(vj.EntityURN),
		Title:           vj.Title,
		Description:     vj.Description.Text,
		Location:        vj.FormattedLocation,
		Remote:          vj.WorkRemoteAllowed,
		PostedAt:        msToTime(vj.ListedAt),
		EmploymentType:  vj.EmploymentStatus,
		ExperienceLevel: vj.ExperienceLevel,
		Company: models.Company{
			URN:  co.EntityURN,
			ID:   co.UniversalName,
			Name: co.Name,
		},
	}
	if vj.ExpireAt > 0 {
		j.ExpiresAt = msToTime(vj.ExpireAt)
	}
	return j
}
