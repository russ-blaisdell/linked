package api

import (
	"encoding/json"
	"fmt"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// ProfileService handles LinkedIn profile operations.
type ProfileService struct {
	c *client.Client
}

// NewProfileService returns a new ProfileService.
func NewProfileService(c *client.Client) *ProfileService {
	return &ProfileService{c: c}
}

// voyagerMiniProfile is embedded in many Voyager responses.
type voyagerMiniProfile struct {
	EntityURN string `json:"entityUrn"`
	PublicID  string `json:"publicIdentifier"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Occupation string `json:"occupation"`
	Picture   *struct {
		Artifacts []struct {
			FileIdentifyingURLPathSegment string `json:"fileIdentifyingUrlPathSegment"`
		} `json:"artifacts"`
		RootURL string `json:"rootUrl"`
	} `json:"picture,omitempty"`
}

// voyagerFullProfile is the shape returned by the profile endpoint.
type voyagerFullProfile struct {
	MiniProfile         voyagerMiniProfile `json:"miniProfile"`
	Summary             string             `json:"summary,omitempty"`
	Industry            struct {
		LocalizedName string `json:"localizedName"`
	} `json:"industry,omitempty"`
	ProfileTopPosition []struct {
		Title           string `json:"title"`
		CompanyName     string `json:"companyName,omitempty"`
		GeoLocationName string `json:"geoLocationName,omitempty"`
		Description     string `json:"description,omitempty"`
		DateRange       *voyagerDateRange `json:"dateRange,omitempty"`
	} `json:"profileTopPosition,omitempty"`
	Education []struct {
		SchoolName   string `json:"schoolName,omitempty"`
		DegreeName   string `json:"degreeName,omitempty"`
		FieldOfStudy string `json:"fieldOfStudy,omitempty"`
		Description  string `json:"description,omitempty"`
		DateRange    *voyagerDateRange `json:"dateRange,omitempty"`
	} `json:"education,omitempty"`
	Skills []struct {
		Name         string `json:"name"`
		NumEndorsers int    `json:"numEndorsers,omitempty"`
	} `json:"skills,omitempty"`
	Languages []struct {
		Name        string `json:"name"`
		Proficiency string `json:"proficiency,omitempty"`
	} `json:"languages,omitempty"`
}

type voyagerDateRange struct {
	Start *struct{ Year int `json:"year"` } `json:"start,omitempty"`
	End   *struct{ Year int `json:"year"` } `json:"end,omitempty"`
}

// voyagerNormalizedEnvelope is the standard Voyager response wrapper.
type voyagerNormalizedEnvelope struct {
	Data     json.RawMessage   `json:"data"`
	Included []json.RawMessage `json:"included"`
}

// GetMe returns the authenticated user's own profile.
func (s *ProfileService) GetMe() (*models.Profile, error) {
	var raw voyagerNormalizedEnvelope
	if err := s.c.Get(client.EndpointMe, nil, &raw); err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}
	return extractProfileFromEnvelope(raw)
}

// GetProfile returns a profile by public profile ID (e.g. "john-doe").
func (s *ProfileService) GetProfile(profileID string) (*models.Profile, error) {
	path := fmt.Sprintf("%s/%s", client.EndpointProfiles, profileID)
	var raw voyagerNormalizedEnvelope
	if err := s.c.Get(path, nil, &raw); err != nil {
		return nil, fmt.Errorf("get profile %q: %w", profileID, err)
	}
	return extractProfileFromEnvelope(raw)
}

// GetContactInfo returns contact information for a profile.
func (s *ProfileService) GetContactInfo(profileID string) (*models.ContactInfo, error) {
	path := fmt.Sprintf(client.EndpointProfileContact, profileID)
	var raw struct {
		EmailAddress   *string `json:"emailAddress"`
		PhoneNumbers   []struct {
			Number string `json:"number"`
		} `json:"phoneNumbers"`
		TwitterHandles []struct {
			Name string `json:"name"`
		} `json:"twitterHandles"`
		Websites []struct {
			URL string `json:"url"`
		} `json:"websites"`
		Address *string `json:"address"`
	}
	if err := s.c.Get(path, nil, &raw); err != nil {
		return nil, fmt.Errorf("get contact info %q: %w", profileID, err)
	}

	info := &models.ContactInfo{ProfileID: profileID}
	if raw.EmailAddress != nil {
		info.Emails = []string{*raw.EmailAddress}
	}
	for _, p := range raw.PhoneNumbers {
		info.PhoneNumbers = append(info.PhoneNumbers, p.Number)
	}
	for _, t := range raw.TwitterHandles {
		info.TwitterHandles = append(info.TwitterHandles, t.Name)
	}
	for _, w := range raw.Websites {
		info.Websites = append(info.Websites, w.URL)
	}
	if raw.Address != nil {
		info.Address = *raw.Address
	}
	return info, nil
}

// UpdateProfile updates mutable fields on the authenticated user's profile.
func (s *ProfileService) UpdateProfile(profileID string, update models.ProfileUpdate) error {
	path := fmt.Sprintf("%s/%s", client.EndpointProfiles, profileID)
	payload := map[string]interface{}{}
	if update.Headline != "" {
		payload["headline"] = update.Headline
	}
	if update.Summary != "" {
		payload["summary"] = update.Summary
	}
	if update.Location != "" {
		payload["location"] = update.Location
	}
	if len(payload) == 0 {
		return nil
	}
	return s.c.Put(path, payload, nil)
}

// extractProfileFromEnvelope tries to find a profile in a normalized Voyager envelope.
func extractProfileFromEnvelope(env voyagerNormalizedEnvelope) (*models.Profile, error) {
	// Try included items first (most profile endpoints return data here).
	for _, raw := range env.Included {
		var vp voyagerFullProfile
		if err := json.Unmarshal(raw, &vp); err != nil {
			continue
		}
		if vp.MiniProfile.EntityURN != "" {
			return mapVoyagerProfile(vp), nil
		}
	}
	// Fall back to data field.
	var vp voyagerFullProfile
	if err := json.Unmarshal(env.Data, &vp); err != nil {
		return nil, fmt.Errorf("no profile found in response")
	}
	if vp.MiniProfile.EntityURN == "" {
		return nil, fmt.Errorf("empty profile in response")
	}
	return mapVoyagerProfile(vp), nil
}

// mapVoyagerProfile converts a raw voyagerFullProfile to models.Profile.
func mapVoyagerProfile(vp voyagerFullProfile) *models.Profile {
	p := &models.Profile{
		URN:       vp.MiniProfile.EntityURN,
		ProfileID: vp.MiniProfile.PublicID,
		FirstName: vp.MiniProfile.FirstName,
		LastName:  vp.MiniProfile.LastName,
		Headline:  vp.MiniProfile.Occupation,
		Summary:   vp.Summary,
		Industry:  vp.Industry.LocalizedName,
	}

	if vp.MiniProfile.Picture != nil && len(vp.MiniProfile.Picture.Artifacts) > 0 {
		last := vp.MiniProfile.Picture.Artifacts[len(vp.MiniProfile.Picture.Artifacts)-1]
		p.PhotoURL = vp.MiniProfile.Picture.RootURL + last.FileIdentifyingURLPathSegment
	}

	for _, pos := range vp.ProfileTopPosition {
		exp := models.Experience{
			Title:       pos.Title,
			CompanyName: pos.CompanyName,
			Location:    pos.GeoLocationName,
			Description: pos.Description,
		}
		if pos.DateRange != nil {
			if pos.DateRange.Start != nil {
				exp.StartDate = fmt.Sprintf("%d", pos.DateRange.Start.Year)
			}
			if pos.DateRange.End != nil {
				exp.EndDate = fmt.Sprintf("%d", pos.DateRange.End.Year)
			} else {
				exp.Current = true
			}
		}
		p.Experience = append(p.Experience, exp)
	}

	for _, edu := range vp.Education {
		e := models.Education{
			SchoolName:   edu.SchoolName,
			Degree:       edu.DegreeName,
			FieldOfStudy: edu.FieldOfStudy,
			Description:  edu.Description,
		}
		if edu.DateRange != nil {
			if edu.DateRange.Start != nil {
				e.StartDate = fmt.Sprintf("%d", edu.DateRange.Start.Year)
			}
			if edu.DateRange.End != nil {
				e.EndDate = fmt.Sprintf("%d", edu.DateRange.End.Year)
			}
		}
		p.Education = append(p.Education, e)
	}

	for _, sk := range vp.Skills {
		p.Skills = append(p.Skills, models.Skill{Name: sk.Name, Endorsements: sk.NumEndorsers})
	}
	for _, lg := range vp.Languages {
		p.Languages = append(p.Languages, models.Language{Name: lg.Name, Proficiency: lg.Proficiency})
	}

	return p
}
