package api

import (
	"encoding/json"
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// CompaniesService handles LinkedIn company operations.
type CompaniesService struct {
	c *client.Client
}

// NewCompaniesService returns a new CompaniesService.
func NewCompaniesService(c *client.Client) *CompaniesService {
	return &CompaniesService{c: c}
}

type voyagerCompany struct {
	EntityURN     string `json:"entityUrn"`
	UniversalName string `json:"universalName"`
	Name          string `json:"name"`
	Tagline       string `json:"tagline,omitempty"`
	Description   string `json:"description,omitempty"`
	Industries    []string `json:"industries,omitempty"`
	Website       string `json:"companyPageUrl,omitempty"`
	Headquarters  struct {
		City    string `json:"city,omitempty"`
		Country string `json:"country,omitempty"`
	} `json:"headquarter,omitempty"`
	StaffCount      int    `json:"staffCount,omitempty"`
	CompanyType     struct {
		LocalizedName string `json:"localizedName"`
	} `json:"companyType,omitempty"`
	Specialities []string `json:"specialities,omitempty"`
}

// GetCompany returns information about a company by universalName or ID.
// The response uses normalized format (data.*elements + included).
func (s *CompaniesService) GetCompany(companyID string) (*models.Company, error) {
	params := map[string]string{
		"q":             "universalName",
		"universalName": companyID,
	}

	var raw struct {
		Data struct {
			Elements []string `json:"*elements"`
		} `json:"data"`
		Included []json.RawMessage `json:"included"`
	}

	if err := s.c.Get(client.EndpointCompanies, params, &raw); err != nil {
		return nil, fmt.Errorf("get company %q: %w", companyID, err)
	}

	// Find the company entity in included.
	for _, inc := range raw.Included {
		var vc voyagerCompany
		if json.Unmarshal(inc, &vc) != nil || vc.Name == "" {
			continue
		}
		return mapVoyagerCompany(vc), nil
	}

	return nil, fmt.Errorf("company not found: %s", companyID)
}

// FollowCompany follows a company.
func (s *CompaniesService) FollowCompany(companyURN string) error {
	return s.c.Post(client.EndpointFollowEntity, map[string]interface{}{
		"followedEntityUrn": companyURN,
	}, nil)
}

// UnfollowCompany unfollows a company.
func (s *CompaniesService) UnfollowCompany(companyURN string) error {
	return s.c.Delete(fmt.Sprintf("%s/%s", client.EndpointFollowEntity, urnToID(companyURN)))
}

// GetCompanyPosts returns recent posts from a company.
func (s *CompaniesService) GetCompanyPosts(companyURN string, start, count int) (*models.PagedPosts, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":          "company",
		"companyUrn": companyURN,
		"start":      fmt.Sprintf("%d", start),
		"count":      fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			Value struct {
				EntityURN  string `json:"entityUrn"`
				Commentary struct {
					Text struct {
						Text string `json:"text"`
					} `json:"text"`
				} `json:"commentary,omitempty"`
				SocialDetail struct {
					LikeCount    int `json:"likeCount,omitempty"`
					CommentCount int `json:"commentCount,omitempty"`
					ShareCount   int `json:"shareCount,omitempty"`
				} `json:"socialDetail,omitempty"`
				CreatedAt int64 `json:"createdAt,omitempty"`
			} `json:"com.linkedin.voyager.feed.render.UpdateV2"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointFeed, params, &raw); err != nil {
		return nil, fmt.Errorf("get company posts for %q: %w", companyURN, err)
	}

	result := &models.PagedPosts{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		v := el.Value
		if v.EntityURN == "" {
			continue
		}
		result.Items = append(result.Items, models.Post{
			URN:          v.EntityURN,
			Body:         v.Commentary.Text.Text,
			LikeCount:    v.SocialDetail.LikeCount,
			CommentCount: v.SocialDetail.CommentCount,
			ShareCount:   v.SocialDetail.ShareCount,
			PostedAt:     msToTime(v.CreatedAt),
		})
	}
	return result, nil
}

// GetCompanyEmployees searches for employees of a company.
func (s *CompaniesService) GetCompanyEmployees(companyID string, start, count int) (*models.PagedSearchPeople, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":       "people",
		"filters": fmt.Sprintf("List(currentCompany->%s)", companyID),
		"start":   fmt.Sprintf("%d", start),
		"count":   fmt.Sprintf("%d", count),
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
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging,omitempty"`
	}

	if err := s.c.Get(client.EndpointSearch, params, &raw); err != nil {
		return nil, fmt.Errorf("get company employees for %q: %w", companyID, err)
	}

	result := &models.PagedSearchPeople{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
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

// mapVoyagerCompany converts a raw company to models.Company.
func mapVoyagerCompany(vc voyagerCompany) *models.Company {
	industry := ""
	if len(vc.Industries) > 0 {
		industry = vc.Industries[0]
	}
	co := &models.Company{
		URN:         vc.EntityURN,
		ID:          vc.UniversalName,
		Name:        vc.Name,
		Headline:    vc.Tagline,
		Description: vc.Description,
		Industry:    industry,
		Website:     vc.Website,
	}
	if vc.Headquarters.City != "" {
		co.Headquarters = fmt.Sprintf("%s, %s", vc.Headquarters.City, vc.Headquarters.Country)
	}
	if vc.StaffCount > 0 {
		co.EmployeeCount = fmt.Sprintf("%d", vc.StaffCount)
	}
	if vc.CompanyType.LocalizedName != "" {
		co.Industry = vc.CompanyType.LocalizedName
		if industry != "" {
			co.Industry = industry
		}
	}
	return co
}
