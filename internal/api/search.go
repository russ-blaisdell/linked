package api

import (
	"encoding/json"
	"fmt"
	"net/url"

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

// GraphQL response types for the unified search endpoint.

type gqlSearchClusterCollection struct {
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
	Elements []struct {
		Items []struct {
			Item json.RawMessage `json:"item"`
		} `json:"items"`
	} `json:"elements"`
}

type gqlEntityResult struct {
	EntityURN       string `json:"entityUrn"`
	Title           *struct{ Text string `json:"text"` } `json:"title"`
	PrimarySubtitle *struct{ Text string `json:"text"` } `json:"primarySubtitle"`
	SecondarySubtitle *struct{ Text string `json:"text"` } `json:"secondarySubtitle"`
	Summary         *struct{ Text string `json:"text"` } `json:"summary"`
}

type gqlSearchFeedUpdate struct {
	Update *struct {
		Actor *struct {
			Name *struct{ Text string `json:"text"` } `json:"name"`
		} `json:"actor"`
		Commentary *struct {
			Text *struct{ Text string `json:"text"` } `json:"text"`
		} `json:"commentary"`
		Metadata *struct {
			ShareURN string `json:"shareUrn"`
			URN      string `json:"urn"`
		} `json:"metadata"`
	} `json:"update"`
}

// doSearch executes a search query against the unified GraphQL search endpoint.
func (s *SearchService) doSearch(resultType string, keywords string, start, count int) (*gqlSearchClusterCollection, error) {
	if count == 0 {
		count = client.DefaultCount
	}

	path := fmt.Sprintf(
		"%s?includeWebMetadata=true&variables=(start:%d,count:%d,origin:GLOBAL_SEARCH_HEADER,query:(keywords:%s,flagshipSearchIntent:SEARCH_SRP,queryParameters:List((key:resultType,value:List(%s)))))&queryId=%s",
		client.EndpointGraphQL, start, count,
		url.QueryEscape(keywords), resultType,
		client.EndpointSearchClustersQueryID,
	)

	var raw struct {
		Data *struct {
			Collection *gqlSearchClusterCollection `json:"searchDashClustersByAll"`
		} `json:"data"`
	}

	if err := s.c.GetGraphQL(path, &raw); err != nil {
		return nil, err
	}
	if raw.Data == nil || raw.Data.Collection == nil {
		return &gqlSearchClusterCollection{}, nil
	}
	return raw.Data.Collection, nil
}

// SearchPeople searches LinkedIn members.
func (s *SearchService) SearchPeople(input models.SearchPeopleInput) (*models.PagedSearchPeople, error) {
	col, err := s.doSearch("PEOPLE", input.Keywords, input.Start, input.Count)
	if err != nil {
		return nil, fmt.Errorf("search people: %w", err)
	}

	result := &models.PagedSearchPeople{
		Pagination: models.Pagination{
			Start:   input.Start,
			Count:   input.Count,
			Total:   col.Paging.Total,
			HasMore: (input.Start + input.Count) < col.Paging.Total,
		},
	}

	for _, cluster := range col.Elements {
		for _, item := range cluster.Items {
			var wrapper struct {
				EntityResult *gqlEntityResult `json:"entityResult"`
			}
			if json.Unmarshal(item.Item, &wrapper) != nil || wrapper.EntityResult == nil {
				continue
			}
			er := wrapper.EntityResult
			name := ""
			if er.Title != nil {
				name = er.Title.Text
			}
			headline := ""
			if er.PrimarySubtitle != nil {
				headline = er.PrimarySubtitle.Text
			}
			location := ""
			if er.SecondarySubtitle != nil {
				location = er.SecondarySubtitle.Text
			}
			result.Items = append(result.Items, models.SearchPeopleResult{
				Profile: models.Profile{
					URN:       extractProfileURN(er.EntityURN),
					FirstName: name,
					Headline:  headline,
					Location:  location,
				},
			})
		}
	}

	return result, nil
}

// SearchJobs searches LinkedIn job postings.
// Note: Jobs search via the unified search endpoint returns internal errors
// from LinkedIn. Use `jobs recommended` or `jobs saved` instead.
func (s *SearchService) SearchJobs(input models.SearchJobsInput) (*models.PagedJobs, error) {
	return nil, fmt.Errorf("job search is not available via the current API — use 'jobs recommended' instead")
}

// SearchCompanies searches LinkedIn companies.
func (s *SearchService) SearchCompanies(keywords string, start, count int) (*models.PagedCompanies, error) {
	col, err := s.doSearch("COMPANIES", keywords, start, count)
	if err != nil {
		return nil, fmt.Errorf("search companies: %w", err)
	}

	result := &models.PagedCompanies{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   col.Paging.Total,
			HasMore: (start + count) < col.Paging.Total,
		},
	}

	for _, cluster := range col.Elements {
		for _, item := range cluster.Items {
			var wrapper struct {
				EntityResult *gqlEntityResult `json:"entityResult"`
			}
			if json.Unmarshal(item.Item, &wrapper) != nil || wrapper.EntityResult == nil {
				continue
			}
			er := wrapper.EntityResult
			name := ""
			if er.Title != nil {
				name = er.Title.Text
			}
			industry := ""
			if er.PrimarySubtitle != nil {
				industry = er.PrimarySubtitle.Text
			}
			result.Items = append(result.Items, models.Company{
				URN:      extractCompanyURN(er.EntityURN),
				Name:     name,
				Industry: industry,
			})
		}
	}

	return result, nil
}

// SearchPosts searches LinkedIn posts/content.
func (s *SearchService) SearchPosts(keywords string, start, count int) (*models.PagedPosts, error) {
	col, err := s.doSearch("CONTENT", keywords, start, count)
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}

	result := &models.PagedPosts{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   col.Paging.Total,
			HasMore: (start + count) < col.Paging.Total,
		},
	}

	for _, cluster := range col.Elements {
		for _, item := range cluster.Items {
			var wrapper struct {
				SearchFeedUpdate *gqlSearchFeedUpdate `json:"searchFeedUpdate"`
			}
			if json.Unmarshal(item.Item, &wrapper) != nil || wrapper.SearchFeedUpdate == nil {
				continue
			}
			sfu := wrapper.SearchFeedUpdate
			if sfu.Update == nil {
				continue
			}

			authorName := ""
			if sfu.Update.Actor != nil && sfu.Update.Actor.Name != nil {
				authorName = sfu.Update.Actor.Name.Text
			}
			body := ""
			if sfu.Update.Commentary != nil && sfu.Update.Commentary.Text != nil {
				body = sfu.Update.Commentary.Text.Text
			}
			urn := ""
			if sfu.Update.Metadata != nil {
				urn = sfu.Update.Metadata.ShareURN
				if urn == "" {
					urn = sfu.Update.Metadata.URN
				}
			}

			result.Items = append(result.Items, models.Post{
				URN:  urn,
				Body: body,
				AuthorProfile: models.Profile{
					FirstName: authorName,
				},
			})
		}
	}

	return result, nil
}

// extractProfileURN pulls the fsd_profile URN from the entityResultViewModel URN.
// Input:  "urn:li:fsd_entityResultViewModel:(urn:li:fsd_profile:ABC123,SEARCH_SRP,DEFAULT)"
// Output: "urn:li:fsd_profile:ABC123"
func extractProfileURN(entityResultURN string) string {
	const prefix = "urn:li:fsd_profile:"
	for i := 0; i < len(entityResultURN)-len(prefix); i++ {
		if entityResultURN[i:i+len(prefix)] == prefix {
			// Find end — either comma or closing paren
			end := i + len(prefix)
			for end < len(entityResultURN) && entityResultURN[end] != ',' && entityResultURN[end] != ')' {
				end++
			}
			return entityResultURN[i:end]
		}
	}
	return entityResultURN
}

// extractCompanyURN pulls the fsd_company URN from the entityResultViewModel URN.
func extractCompanyURN(entityResultURN string) string {
	const prefix = "urn:li:fsd_company:"
	for i := 0; i < len(entityResultURN)-len(prefix); i++ {
		if entityResultURN[i:i+len(prefix)] == prefix {
			end := i + len(prefix)
			for end < len(entityResultURN) && entityResultURN[end] != ',' && entityResultURN[end] != ')' {
				end++
			}
			return entityResultURN[i:end]
		}
	}
	return entityResultURN
}
