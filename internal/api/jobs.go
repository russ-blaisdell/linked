package api

import (
	"fmt"

	"strings"

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

// graphqlJobPostingCard is the card shape returned by GraphQL job endpoints.
type graphqlJobPostingCard struct {
	JobPostingTitle    string `json:"jobPostingTitle"`
	PrimaryDescription *struct {
		Text string `json:"text"`
	} `json:"primaryDescription"`
	JobPosting *struct {
		EntityURN string `json:"entityUrn"`
		Title     string `json:"title"`
	} `json:"jobPosting"`
}

// graphqlJobPostingDetail is the full detail shape from GetJob.
type graphqlJobPostingDetail struct {
	EntityURN  string `json:"entityUrn"`
	Title      string `json:"title"`
	Description *struct {
		Text string `json:"text"`
	} `json:"description"`
	CompanyDetails *struct {
		Name string `json:"name"`
	} `json:"companyDetails"`
	Location *struct {
		DefaultLocalizedName string `json:"defaultLocalizedName"`
	} `json:"location"`
	EmploymentStatus *struct {
		LocalizedName string `json:"localizedName"`
	} `json:"employmentStatus"`
	JobState       string `json:"jobState"`
	CompanyApplyURL string `json:"companyApplyUrl"`
	WorkRemoteAllowed bool `json:"workRemoteAllowed"`
}

// mapCardToJob converts a GraphQL job posting card to models.Job.
func mapCardToJob(card graphqlJobPostingCard) *models.Job {
	j := &models.Job{
		Title: card.JobPostingTitle,
	}
	if card.PrimaryDescription != nil {
		// primaryDescription.text is typically "Company Name · Location"
		parts := strings.SplitN(card.PrimaryDescription.Text, " · ", 2)
		if len(parts) > 0 {
			j.Company = models.Company{Name: strings.TrimSpace(parts[0])}
		}
		if len(parts) > 1 {
			j.Location = strings.TrimSpace(parts[1])
		}
	}
	if card.JobPosting != nil {
		j.URN = card.JobPosting.EntityURN
		j.ID = urnToID(card.JobPosting.EntityURN)
		if j.Title == "" {
			j.Title = card.JobPosting.Title
		}
	}
	return j
}

// GetJob returns details for a specific job posting.
func (s *JobsService) GetJob(jobID string) (*models.Job, error) {
	path := client.EndpointGraphQL +
		"?includeWebMetadata=true&variables=(jobPostingUrn:urn%3Ali%3Afsd_jobPosting%3A" + jobID +
		")&queryId=" + client.EndpointJobPostingsQueryID

	var raw struct {
		Data *struct {
			JobsDashJobPostingsByID *graphqlJobPostingDetail `json:"jobsDashJobPostingsById"`
		} `json:"data"`
	}
	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("get job %q: %w", jobID, err)
	}

	if raw.Data == nil || raw.Data.JobsDashJobPostingsByID == nil {
		return nil, fmt.Errorf("get job %q: no data returned", jobID)
	}

	detail := raw.Data.JobsDashJobPostingsByID
	j := &models.Job{
		URN:   detail.EntityURN,
		ID:    urnToID(detail.EntityURN),
		Title: detail.Title,
	}
	if detail.Description != nil {
		j.Description = detail.Description.Text
	}
	if detail.CompanyDetails != nil {
		j.Company = models.Company{Name: detail.CompanyDetails.Name}
	}
	if detail.Location != nil {
		j.Location = detail.Location.DefaultLocalizedName
	}
	if detail.EmploymentStatus != nil {
		j.EmploymentType = detail.EmploymentStatus.LocalizedName
	}
	j.ApplyURL = detail.CompanyApplyURL
	j.Remote = detail.WorkRemoteAllowed
	return j, nil
}

// ListSavedJobs returns the authenticated user's saved jobs.
func (s *JobsService) ListSavedJobs(start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(count:%d,start:%d,jobCollectionSlug:saved,query:(origin:GENERIC_JOB_COLLECTIONS_LANDING))&queryId=%s",
		client.EndpointGraphQL, count, start, client.EndpointJobCardsQueryID)

	var raw struct {
		Data *struct {
			JobsDashJobCardsByJobSearchV2 *struct {
				Elements []struct {
					JobCard *struct {
						JobPostingCardWrapper *struct {
							JobPostingCard graphqlJobPostingCard `json:"jobPostingCard"`
						} `json:"jobPostingCardWrapper"`
					} `json:"jobCard"`
				} `json:"elements"`
				Paging struct {
					Start int `json:"start"`
					Count int `json:"count"`
					Total int `json:"total"`
				} `json:"paging"`
			} `json:"jobsDashJobCardsByJobSearchV2"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("list saved jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start: start,
			Count: count,
		},
	}

	if raw.Data != nil && raw.Data.JobsDashJobCardsByJobSearchV2 != nil {
		col := raw.Data.JobsDashJobCardsByJobSearchV2
		result.Pagination.Total = col.Paging.Total
		result.Pagination.HasMore = (start + count) < col.Paging.Total
		for _, el := range col.Elements {
			if el.JobCard == nil || el.JobCard.JobPostingCardWrapper == nil {
				continue
			}
			j := mapCardToJob(el.JobCard.JobPostingCardWrapper.JobPostingCard)
			j.Saved = true
			result.Items = append(result.Items, *j)
		}
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
	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(count:%d,start:%d,jobCollectionSlug:applied,query:(origin:GENERIC_JOB_COLLECTIONS_LANDING))&queryId=%s",
		client.EndpointGraphQL, count, start, client.EndpointJobCardsQueryID)

	var raw struct {
		Data *struct {
			JobsDashJobCardsByJobSearchV2 *struct {
				Elements []struct {
					JobCard *struct {
						JobPostingCardWrapper *struct {
							JobPostingCard graphqlJobPostingCard `json:"jobPostingCard"`
						} `json:"jobPostingCardWrapper"`
					} `json:"jobCard"`
				} `json:"elements"`
				Paging struct {
					Start int `json:"start"`
					Count int `json:"count"`
					Total int `json:"total"`
				} `json:"paging"`
			} `json:"jobsDashJobCardsByJobSearchV2"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("list applied jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start: start,
			Count: count,
		},
	}

	if raw.Data != nil && raw.Data.JobsDashJobCardsByJobSearchV2 != nil {
		col := raw.Data.JobsDashJobCardsByJobSearchV2
		result.Pagination.Total = col.Paging.Total
		result.Pagination.HasMore = (start + count) < col.Paging.Total
		for _, el := range col.Elements {
			if el.JobCard == nil || el.JobCard.JobPostingCardWrapper == nil {
				continue
			}
			j := mapCardToJob(el.JobCard.JobPostingCardWrapper.JobPostingCard)
			j.Applied = true
			result.Items = append(result.Items, *j)
		}
	}
	return result, nil
}

// GetRecommendedJobs returns LinkedIn's recommended jobs for the authenticated user.
func (s *JobsService) GetRecommendedJobs(start, count int) (*models.PagedJobs, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	path := fmt.Sprintf("%s?variables=()&queryId=%s",
		client.EndpointGraphQL, client.EndpointJobsFeedQueryID)

	var raw struct {
		Data *struct {
			JobsDashJobsFeedAll *struct {
				Elements []struct {
					EntitiesResolutionResults []struct {
						JobPostingCard *graphqlJobPostingCard `json:"jobPostingCard"`
					} `json:"entitiesResolutionResults"`
				} `json:"elements"`
				Paging struct {
					Start int `json:"start"`
					Count int `json:"count"`
					Total int `json:"total"`
				} `json:"paging"`
			} `json:"jobsDashJobsFeedAll"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, fmt.Errorf("get recommended jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start: start,
			Count: count,
		},
	}

	if raw.Data != nil && raw.Data.JobsDashJobsFeedAll != nil {
		col := raw.Data.JobsDashJobsFeedAll
		result.Pagination.Total = col.Paging.Total
		result.Pagination.HasMore = (start + count) < col.Paging.Total
		for _, el := range col.Elements {
			for _, res := range el.EntitiesResolutionResults {
				if res.JobPostingCard == nil {
					continue
				}
				j := mapCardToJob(*res.JobPostingCard)
				if j.URN == "" && j.Title == "" {
					continue
				}
				result.Items = append(result.Items, *j)
			}
		}
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
