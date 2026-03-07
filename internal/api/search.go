package api

import (
	"fmt"
	"strings"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// SearchService handles LinkedIn search operations.
type SearchService struct {
	c *client.Client
}

// NewSearchService returns a new SearchService.
func NewSearchService(c *client.Client) *SearchService {
	return &SearchService{c: c}
}

// SearchPeople searches LinkedIn members.
func (s *SearchService) SearchPeople(input models.SearchPeopleInput) (*models.PagedSearchPeople, error) {
	if input.Count == 0 {
		input.Count = client.DefaultCount
	}

	params := map[string]string{
		"q":       "people",
		"start":   fmt.Sprintf("%d", input.Start),
		"count":   fmt.Sprintf("%d", input.Count),
		"filters": buildPeopleFilters(input),
	}
	if input.Keywords != "" {
		params["keywords"] = input.Keywords
	}

	var raw struct {
		Elements []struct {
			HitInfo struct {
				MiniProfile voyagerMiniProfile `json:"com.linkedin.voyager.search.SearchProfile"`
			} `json:"hitInfo"`
			Distance struct {
				Value string `json:"value"`
			} `json:"distance,omitempty"`
		} `json:"elements"`
		Total  int `json:"total,omitempty"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging,omitempty"`
	}

	if err := s.c.Get(client.EndpointSearch, params, &raw); err != nil {
		return nil, fmt.Errorf("search people: %w", err)
	}

	result := &models.PagedSearchPeople{
		Pagination: models.Pagination{
			Start:   input.Start,
			Count:   input.Count,
			Total:   raw.Paging.Total,
			HasMore: (input.Start + input.Count) < raw.Paging.Total,
		},
	}

	for _, el := range raw.Elements {
		mp := el.HitInfo.MiniProfile
		if mp.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, models.SearchPeopleResult{
			Profile: models.Profile{
				URN:       mp.EntityURN,
				ProfileID: mp.PublicID,
				FirstName: mp.FirstName,
				LastName:  mp.LastName,
				Headline:  mp.Occupation,
			},
			Distance: el.Distance.Value,
		})
	}

	return result, nil
}

// SearchJobs searches LinkedIn job postings.
func (s *SearchService) SearchJobs(input models.SearchJobsInput) (*models.PagedJobs, error) {
	if input.Count == 0 {
		input.Count = client.DefaultCount
	}

	params := map[string]string{
		"q":     "jobSearch",
		"start": fmt.Sprintf("%d", input.Start),
		"count": fmt.Sprintf("%d", input.Count),
	}
	if input.Keywords != "" {
		params["keywords"] = input.Keywords
	}
	if input.Location != "" {
		params["locationUnion"] = fmt.Sprintf(`(geoId:%s)`, input.Location)
	}

	filters := buildJobFilters(input)
	if filters != "" {
		params["filters"] = filters
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
		return nil, fmt.Errorf("search jobs: %w", err)
	}

	result := &models.PagedJobs{
		Pagination: models.Pagination{
			Start:   input.Start,
			Count:   input.Count,
			Total:   raw.Paging.Total,
			HasMore: (input.Start + input.Count) < raw.Paging.Total,
		},
	}

	for _, el := range raw.Elements {
		card := el.JobCardUnion.JobPostingCard
		if card.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, models.Job{
			URN:      card.EntityURN,
			Title:    card.Title,
			Company:  models.Company{Name: card.Company.Name, ID: card.Company.UniversalName},
			Location: card.FormattedLocation,
			Remote:   card.WorkRemoteAllowed,
		})
	}

	return result, nil
}

// SearchCompanies searches LinkedIn companies.
func (s *SearchService) SearchCompanies(keywords string, start, count int) (*models.PagedCompanies, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":        "company",
		"keywords": keywords,
		"start":    fmt.Sprintf("%d", start),
		"count":    fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			HitInfo struct {
				Company struct {
					EntityURN     string `json:"entityUrn"`
					UniversalName string `json:"universalName"`
					Name          string `json:"name"`
					Industry      struct {
						LocalizedName string `json:"localizedName"`
					} `json:"industry,omitempty"`
				} `json:"com.linkedin.voyager.search.SearchCompany"`
			} `json:"hitInfo"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging,omitempty"`
	}

	if err := s.c.Get(client.EndpointSearch, params, &raw); err != nil {
		return nil, fmt.Errorf("search companies: %w", err)
	}

	result := &models.PagedCompanies{
		Pagination: models.Pagination{Start: start, Count: count, Total: raw.Paging.Total, HasMore: (start + count) < raw.Paging.Total},
	}
	for _, el := range raw.Elements {
		co := el.HitInfo.Company
		if co.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, models.Company{
			URN:      co.EntityURN,
			ID:       co.UniversalName,
			Name:     co.Name,
			Industry: co.Industry.LocalizedName,
		})
	}
	return result, nil
}

// SearchPosts searches LinkedIn posts/content.
func (s *SearchService) SearchPosts(keywords string, start, count int) (*models.PagedPosts, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":        "blended",
		"keywords": keywords,
		"filters":  "List(resultType->CONTENT)",
		"start":    fmt.Sprintf("%d", start),
		"count":    fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			Items []struct {
				Item struct {
					EntityUrn string `json:"entityUrn"`
					Activity  struct {
						EntityUrn  string `json:"entityUrn"`
						UpdateText struct {
							Text string `json:"text"`
						} `json:"updateText,omitempty"`
						TotalLikes int   `json:"totalLikes,omitempty"`
						CreatedAt  int64 `json:"createdAt,omitempty"`
					} `json:"com.linkedin.voyager.search.SearchActivity,omitempty"`
				} `json:"item"`
			} `json:"items,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging,omitempty"`
	}

	if err := s.c.Get(client.EndpointSearchBlended, params, &raw); err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}

	result := &models.PagedPosts{
		Pagination: models.Pagination{Start: start, Count: count, Total: raw.Paging.Total, HasMore: (start + count) < raw.Paging.Total},
	}
	for _, group := range raw.Elements {
		for _, el := range group.Items {
			activity := el.Item.Activity
			urn := activity.EntityUrn
			if urn == "" {
				urn = el.Item.EntityUrn
			}
			if urn == "" {
				continue
			}
			result.Items = append(result.Items, models.Post{
				URN:       urn,
				Body:      activity.UpdateText.Text,
				LikeCount: activity.TotalLikes,
				PostedAt:  msToTime(activity.CreatedAt),
			})
		}
	}
	return result, nil
}

// buildPeopleFilters constructs the Voyager filter string for people search.
func buildPeopleFilters(input models.SearchPeopleInput) string {
	var parts []string
	if len(input.Network) > 0 {
		parts = append(parts, fmt.Sprintf("network->%s", strings.Join(input.Network, "|")))
	}
	if input.Company != "" {
		parts = append(parts, fmt.Sprintf("currentCompany->%s", input.Company))
	}
	if input.Title != "" {
		parts = append(parts, fmt.Sprintf("title->%s", input.Title))
	}
	if input.School != "" {
		parts = append(parts, fmt.Sprintf("school->%s", input.School))
	}
	if len(parts) == 0 {
		return "List()"
	}
	return fmt.Sprintf("List(%s)", strings.Join(parts, ","))
}

// buildJobFilters constructs the Voyager filter string for job search.
func buildJobFilters(input models.SearchJobsInput) string {
	var parts []string
	if input.Remote {
		parts = append(parts, "workplaceType->2") // 2 = REMOTE
	}
	if input.ExperienceLevel != "" {
		parts = append(parts, fmt.Sprintf("experienceLevel->%s", input.ExperienceLevel))
	}
	if input.EmploymentType != "" {
		parts = append(parts, fmt.Sprintf("employmentType->%s", input.EmploymentType))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("List(%s)", strings.Join(parts, ","))
}
