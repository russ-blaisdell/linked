package api

import (
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
	Industry      struct {
		LocalizedName string `json:"localizedName"`
	} `json:"industry,omitempty"`
	Website               string `json:"websiteUrl,omitempty"`
	Headquarters          struct {
		City    string `json:"city,omitempty"`
		Country string `json:"country,omitempty"`
	} `json:"headquarter,omitempty"`
	StaffCount            int    `json:"staffCount,omitempty"`
	StaffCountRange       struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"staffCountRange,omitempty"`
	Logo struct {
		Image struct {
			Artifacts []struct {
				FileIdentifyingURLPathSegment string `json:"fileIdentifyingUrlPathSegment"`
			} `json:"artifacts"`
			RootURL string `json:"rootUrl"`
		} `json:"com.linkedin.voyager.common.MediaProcessorImage"`
	} `json:"logoResolutionResult,omitempty"`
}

// GetCompany returns information about a company by universalName or ID.
func (s *CompaniesService) GetCompany(companyID string) (*models.Company, error) {
	params := map[string]string{
		"q":              "universalName",
		"universalName":  companyID,
		"decorationId":   "com.linkedin.voyager.deco.organization.web.WebFullCompanyMain-12",
	}

	var raw struct {
		Elements []voyagerCompany `json:"elements"`
	}

	if err := s.c.Get(client.EndpointCompanies, params, &raw); err != nil {
		return nil, fmt.Errorf("get company %q: %w", companyID, err)
	}

	if len(raw.Elements) == 0 {
		return nil, fmt.Errorf("company not found: %s", companyID)
	}

	return mapVoyagerCompany(raw.Elements[0]), nil
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
		"q":           "company",
		"companyUrn":  companyURN,
		"start":       fmt.Sprintf("%d", start),
		"count":       fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			Value struct {
				EntityURN string `json:"entityUrn"`
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

// mapVoyagerCompany converts a raw company to models.Company.
func mapVoyagerCompany(vc voyagerCompany) *models.Company {
	co := &models.Company{
		URN:         vc.EntityURN,
		ID:          vc.UniversalName,
		Name:        vc.Name,
		Headline:    vc.Tagline,
		Description: vc.Description,
		Industry:    vc.Industry.LocalizedName,
		Website:     vc.Website,
	}
	if vc.Headquarters.City != "" {
		co.Headquarters = fmt.Sprintf("%s, %s", vc.Headquarters.City, vc.Headquarters.Country)
	}
	if vc.StaffCount > 0 {
		co.EmployeeCount = fmt.Sprintf("%d", vc.StaffCount)
	} else if vc.StaffCountRange.Start > 0 {
		co.EmployeeCount = fmt.Sprintf("%d-%d", vc.StaffCountRange.Start, vc.StaffCountRange.End)
	}
	img := vc.Logo.Image
	if len(img.Artifacts) > 0 {
		co.LogoURL = img.RootURL + img.Artifacts[len(img.Artifacts)-1].FileIdentifyingURLPathSegment
	}
	return co
}
