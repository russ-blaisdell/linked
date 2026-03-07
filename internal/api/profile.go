package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	EntityURN  string `json:"entityUrn"`
	PublicID   string `json:"publicIdentifier"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Occupation string `json:"occupation"`
	Picture    *struct {
		Artifacts []struct {
			FileIdentifyingURLPathSegment string `json:"fileIdentifyingUrlPathSegment"`
		} `json:"artifacts"`
		RootURL string `json:"rootUrl"`
	} `json:"picture,omitempty"`
}

// voyagerFullProfile is the shape returned by the profile endpoint.
type voyagerFullProfile struct {
	MiniProfile        voyagerMiniProfile `json:"miniProfile"`
	Summary            string             `json:"summary,omitempty"`
	Industry           struct {
		LocalizedName string `json:"localizedName"`
	} `json:"industry,omitempty"`
	ProfileTopPosition []struct {
		Title           string           `json:"title"`
		CompanyName     string           `json:"companyName,omitempty"`
		GeoLocationName string           `json:"geoLocationName,omitempty"`
		Description     string           `json:"description,omitempty"`
		DateRange       *voyagerDateRange `json:"dateRange,omitempty"`
	} `json:"profileTopPosition,omitempty"`
	Education []struct {
		SchoolName   string           `json:"schoolName,omitempty"`
		DegreeName   string           `json:"degreeName,omitempty"`
		FieldOfStudy string           `json:"fieldOfStudy,omitempty"`
		Description  string           `json:"description,omitempty"`
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

// voyagerIDResponse is returned when creating a new profile section item.
type voyagerIDResponse struct {
	ID string `json:"id,omitempty"`
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

// UpdateProfile updates mutable top-level fields on the authenticated user's profile.
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

// --- Experience ---

// AddExperience adds a new position to the authenticated user's profile.
func (s *ProfileService) AddExperience(profileID string, input models.ExperienceInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfilePositions, profileID)
	payload := buildExperiencePayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add experience: %w", err)
	}
	return resp.ID, nil
}

// UpdateExperience updates an existing position.
func (s *ProfileService) UpdateExperience(profileID, positionID string, input models.ExperienceInput) error {
	path := fmt.Sprintf(client.EndpointProfilePosition, profileID, positionID)
	return s.c.Put(path, buildExperiencePayload(input), nil)
}

// DeleteExperience removes a position from the profile.
func (s *ProfileService) DeleteExperience(profileID, positionID string) error {
	path := fmt.Sprintf(client.EndpointProfilePosition, profileID, positionID)
	return s.c.Delete(path)
}

func buildExperiencePayload(input models.ExperienceInput) map[string]interface{} {
	payload := map[string]interface{}{
		"title":       input.Title,
		"companyName": input.CompanyName,
	}
	if input.CompanyURN != "" {
		payload["company"] = map[string]interface{}{"entityUrn": input.CompanyURN}
	}
	if input.Location != "" {
		payload["geoLocationName"] = input.Location
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.StartYear > 0 {
		start := map[string]interface{}{"year": input.StartYear}
		if input.StartMonth > 0 {
			start["month"] = input.StartMonth
		}
		payload["dateRange"] = map[string]interface{}{"start": start}
	}
	if !input.Current && input.EndYear > 0 {
		end := map[string]interface{}{"year": input.EndYear}
		if input.EndMonth > 0 {
			end["month"] = input.EndMonth
		}
		dr, _ := payload["dateRange"].(map[string]interface{})
		if dr == nil {
			dr = map[string]interface{}{}
		}
		dr["end"] = end
		payload["dateRange"] = dr
	}
	return payload
}

// --- Education ---

// AddEducation adds a new education entry to the profile.
func (s *ProfileService) AddEducation(profileID string, input models.EducationInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileEducations, profileID)
	payload := buildEducationPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add education: %w", err)
	}
	return resp.ID, nil
}

// UpdateEducation updates an existing education entry.
func (s *ProfileService) UpdateEducation(profileID, educationID string, input models.EducationInput) error {
	path := fmt.Sprintf(client.EndpointProfileEducation, profileID, educationID)
	return s.c.Put(path, buildEducationPayload(input), nil)
}

// DeleteEducation removes an education entry from the profile.
func (s *ProfileService) DeleteEducation(profileID, educationID string) error {
	path := fmt.Sprintf(client.EndpointProfileEducation, profileID, educationID)
	return s.c.Delete(path)
}

func buildEducationPayload(input models.EducationInput) map[string]interface{} {
	payload := map[string]interface{}{
		"schoolName":   input.SchoolName,
		"degreeName":   input.Degree,
		"fieldOfStudy": input.FieldOfStudy,
	}
	if input.SchoolURN != "" {
		payload["school"] = map[string]interface{}{"entityUrn": input.SchoolURN}
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.StartYear > 0 || input.EndYear > 0 {
		dr := map[string]interface{}{}
		if input.StartYear > 0 {
			dr["start"] = map[string]interface{}{"year": input.StartYear}
		}
		if input.EndYear > 0 {
			dr["end"] = map[string]interface{}{"year": input.EndYear}
		}
		payload["dateRange"] = dr
	}
	return payload
}

// --- Skills ---

// AddSkill adds a skill to the profile.
func (s *ProfileService) AddSkill(profileID string, input models.SkillInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileSkillsPath, profileID)
	payload := map[string]interface{}{"name": input.Name}
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add skill: %w", err)
	}
	return resp.ID, nil
}

// DeleteSkill removes a skill from the profile.
func (s *ProfileService) DeleteSkill(profileID, skillID string) error {
	path := fmt.Sprintf(client.EndpointProfileSkillPath, profileID, skillID)
	return s.c.Delete(path)
}

// --- Certifications ---

// AddCertification adds a certification to the profile.
func (s *ProfileService) AddCertification(profileID string, input models.CertificationInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileCertifications, profileID)
	payload := buildCertificationPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add certification: %w", err)
	}
	return resp.ID, nil
}

// UpdateCertification updates an existing certification.
func (s *ProfileService) UpdateCertification(profileID, certID string, input models.CertificationInput) error {
	path := fmt.Sprintf(client.EndpointProfileCertification, profileID, certID)
	return s.c.Put(path, buildCertificationPayload(input), nil)
}

// DeleteCertification removes a certification from the profile.
func (s *ProfileService) DeleteCertification(profileID, certID string) error {
	path := fmt.Sprintf(client.EndpointProfileCertification, profileID, certID)
	return s.c.Delete(path)
}

func buildCertificationPayload(input models.CertificationInput) map[string]interface{} {
	payload := map[string]interface{}{
		"name":      input.Name,
		"authority": input.Authority,
	}
	if input.LicenseNum != "" {
		payload["licenseNumber"] = input.LicenseNum
	}
	if input.URL != "" {
		payload["url"] = input.URL
	}
	if input.StartYear > 0 || input.EndYear > 0 {
		dr := map[string]interface{}{}
		if input.StartYear > 0 {
			start := map[string]interface{}{"year": input.StartYear}
			if input.StartMonth > 0 {
				start["month"] = input.StartMonth
			}
			dr["start"] = start
		}
		if input.EndYear > 0 {
			end := map[string]interface{}{"year": input.EndYear}
			if input.EndMonth > 0 {
				end["month"] = input.EndMonth
			}
			dr["end"] = end
		}
		payload["timePeriod"] = dr
	}
	return payload
}

// --- Languages ---

// AddLanguage adds a language to the profile.
func (s *ProfileService) AddLanguage(profileID string, input models.LanguageInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileLanguages, profileID)
	payload := map[string]interface{}{
		"name":        input.Name,
		"proficiency": input.Proficiency,
	}
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add language: %w", err)
	}
	return resp.ID, nil
}

// DeleteLanguage removes a language from the profile.
func (s *ProfileService) DeleteLanguage(profileID, languageID string) error {
	path := fmt.Sprintf(client.EndpointProfileLanguage, profileID, languageID)
	return s.c.Delete(path)
}

// --- Volunteer Experience ---

// AddVolunteer adds a volunteer experience to the profile.
func (s *ProfileService) AddVolunteer(profileID string, input models.VolunteerInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileVolunteer, profileID)
	payload := buildVolunteerPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add volunteer experience: %w", err)
	}
	return resp.ID, nil
}

// UpdateVolunteer updates an existing volunteer experience.
func (s *ProfileService) UpdateVolunteer(profileID, volunteerID string, input models.VolunteerInput) error {
	path := fmt.Sprintf(client.EndpointProfileVolunteerItem, profileID, volunteerID)
	return s.c.Put(path, buildVolunteerPayload(input), nil)
}

// DeleteVolunteer removes a volunteer experience from the profile.
func (s *ProfileService) DeleteVolunteer(profileID, volunteerID string) error {
	path := fmt.Sprintf(client.EndpointProfileVolunteerItem, profileID, volunteerID)
	return s.c.Delete(path)
}

func buildVolunteerPayload(input models.VolunteerInput) map[string]interface{} {
	payload := map[string]interface{}{
		"role":             input.Role,
		"organizationName": input.Organization,
	}
	if input.Cause != "" {
		payload["cause"] = input.Cause
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.StartYear > 0 {
		start := map[string]interface{}{"year": input.StartYear}
		if input.StartMonth > 0 {
			start["month"] = input.StartMonth
		}
		dr := map[string]interface{}{"start": start}
		if !input.Current && input.EndYear > 0 {
			end := map[string]interface{}{"year": input.EndYear}
			if input.EndMonth > 0 {
				end["month"] = input.EndMonth
			}
			dr["end"] = end
		}
		payload["dateRange"] = dr
	}
	return payload
}

// --- Projects ---

// AddProject adds a project to the profile.
func (s *ProfileService) AddProject(profileID string, input models.ProjectInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileProjects, profileID)
	payload := buildProjectPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add project: %w", err)
	}
	return resp.ID, nil
}

// UpdateProject updates an existing project.
func (s *ProfileService) UpdateProject(profileID, projectID string, input models.ProjectInput) error {
	path := fmt.Sprintf(client.EndpointProfileProject, profileID, projectID)
	return s.c.Put(path, buildProjectPayload(input), nil)
}

// DeleteProject removes a project from the profile.
func (s *ProfileService) DeleteProject(profileID, projectID string) error {
	path := fmt.Sprintf(client.EndpointProfileProject, profileID, projectID)
	return s.c.Delete(path)
}

func buildProjectPayload(input models.ProjectInput) map[string]interface{} {
	payload := map[string]interface{}{"title": input.Title}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.URL != "" {
		payload["url"] = input.URL
	}
	if input.StartYear > 0 {
		start := map[string]interface{}{"year": input.StartYear}
		if input.StartMonth > 0 {
			start["month"] = input.StartMonth
		}
		dr := map[string]interface{}{"start": start}
		if !input.Current && input.EndYear > 0 {
			end := map[string]interface{}{"year": input.EndYear}
			if input.EndMonth > 0 {
				end["month"] = input.EndMonth
			}
			dr["end"] = end
		}
		payload["dateRange"] = dr
	}
	return payload
}

// --- Publications ---

// AddPublication adds a publication to the profile.
func (s *ProfileService) AddPublication(profileID string, input models.PublicationInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfilePublications, profileID)
	payload := buildPublicationPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add publication: %w", err)
	}
	return resp.ID, nil
}

// UpdatePublication updates an existing publication.
func (s *ProfileService) UpdatePublication(profileID, pubID string, input models.PublicationInput) error {
	path := fmt.Sprintf(client.EndpointProfilePublication, profileID, pubID)
	return s.c.Put(path, buildPublicationPayload(input), nil)
}

// DeletePublication removes a publication from the profile.
func (s *ProfileService) DeletePublication(profileID, pubID string) error {
	path := fmt.Sprintf(client.EndpointProfilePublication, profileID, pubID)
	return s.c.Delete(path)
}

func buildPublicationPayload(input models.PublicationInput) map[string]interface{} {
	payload := map[string]interface{}{"name": input.Name}
	if input.Publisher != "" {
		payload["publisher"] = input.Publisher
	}
	if input.URL != "" {
		payload["url"] = input.URL
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.Year > 0 {
		date := map[string]interface{}{"year": input.Year}
		if input.Month > 0 {
			date["month"] = input.Month
		}
		payload["date"] = date
	}
	return payload
}

// --- Honors & Awards ---

// AddHonor adds an honor or award to the profile.
func (s *ProfileService) AddHonor(profileID string, input models.HonorInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileHonors, profileID)
	payload := buildHonorPayload(input)
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add honor: %w", err)
	}
	return resp.ID, nil
}

// DeleteHonor removes an honor from the profile.
func (s *ProfileService) DeleteHonor(profileID, honorID string) error {
	path := fmt.Sprintf(client.EndpointProfileHonor, profileID, honorID)
	return s.c.Delete(path)
}

func buildHonorPayload(input models.HonorInput) map[string]interface{} {
	payload := map[string]interface{}{"title": input.Title}
	if input.Issuer != "" {
		payload["issuer"] = input.Issuer
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.Year > 0 {
		date := map[string]interface{}{"year": input.Year}
		if input.Month > 0 {
			date["month"] = input.Month
		}
		payload["issueDate"] = date
	}
	return payload
}

// --- Courses ---

// AddCourse adds a course to the profile.
func (s *ProfileService) AddCourse(profileID string, input models.CourseInput) (string, error) {
	path := fmt.Sprintf(client.EndpointProfileCourses, profileID)
	payload := map[string]interface{}{"name": input.Name}
	if input.Number != "" {
		payload["number"] = input.Number
	}
	if input.Occupation != "" {
		payload["occupation"] = input.Occupation
	}
	var resp voyagerIDResponse
	if err := s.c.Post(path, payload, &resp); err != nil {
		return "", fmt.Errorf("add course: %w", err)
	}
	return resp.ID, nil
}

// DeleteCourse removes a course from the profile.
func (s *ProfileService) DeleteCourse(profileID, courseID string) error {
	path := fmt.Sprintf(client.EndpointProfileCourse, profileID, courseID)
	return s.c.Delete(path)
}

// --- Open to Work ---

// SetOpenToWork sets the user's job-seeking preferences.
func (s *ProfileService) SetOpenToWork(profileURN string, input models.OpenToWorkInput) error {
	payload := map[string]interface{}{
		"member":          map[string]interface{}{"entityUrn": profileURN},
		"jobTypes":        input.JobTypes,
		"preferenceTypes": input.PreferenceTypes,
	}
	if input.Title != "" {
		payload["jobTitle"] = input.Title
	}
	if len(input.Locations) > 0 {
		locs := make([]map[string]interface{}, 0, len(input.Locations))
		for _, l := range input.Locations {
			locs = append(locs, map[string]interface{}{"name": l})
		}
		payload["locations"] = locs
	}
	return s.c.Post(client.EndpointJobSeekingProfiles, payload, nil)
}

// ClearOpenToWork removes the user's open-to-work status.
func (s *ProfileService) ClearOpenToWork(profileID string) error {
	path := fmt.Sprintf(client.EndpointJobSeekingProfile, profileID)
	return s.c.Delete(path)
}

// --- Who Viewed Profile ---

// GetWhoViewed returns recent profile viewers.
func (s *ProfileService) GetWhoViewed(start, count int) (*models.PagedProfileViewers, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "viewedByMember",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			EntityURN  string             `json:"entityUrn"`
			ViewedAt   int64              `json:"viewedAt,omitempty"`
			ViewCount  int                `json:"viewCount,omitempty"`
			MiniProfile voyagerMiniProfile `json:"miniProfile,omitempty"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointWVMPCards, params, &raw); err != nil {
		return nil, fmt.Errorf("get who viewed: %w", err)
	}

	result := &models.PagedProfileViewers{
		Pagination: models.Pagination{
			Start:   start,
			Count:   count,
			Total:   raw.Paging.Total,
			HasMore: (start + count) < raw.Paging.Total,
		},
	}
	for _, el := range raw.Elements {
		mp := el.MiniProfile
		result.Items = append(result.Items, models.ProfileViewer{
			Profile: models.Profile{
				URN:       mp.EntityURN,
				ProfileID: mp.PublicID,
				FirstName: mp.FirstName,
				LastName:  mp.LastName,
				Headline:  mp.Occupation,
			},
			ViewedAt:  msToTime(el.ViewedAt),
			ViewCount: el.ViewCount,
		})
	}
	return result, nil
}

// --- Profile Photo ---

// UploadProfilePhoto uploads a new profile photo from a local file path.
// It follows LinkedIn's three-step media upload process:
// 1. Register the upload and get an upload URL + asset URN
// 2. PUT the image binary to the upload URL
// 3. Associate the asset URN with the profile
func (s *ProfileService) UploadProfilePhoto(profileID, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading image file: %w", err)
	}

	contentType := "image/jpeg"
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".gif" {
		contentType = "image/gif"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}

	// Step 1: register upload
	registerPayload := map[string]interface{}{
		"registerUploadRequest": map[string]interface{}{
			"owner": fmt.Sprintf("urn:li:person:%s", profileID),
			"recipes": []string{"urn:li:digitalmediaRecipe:profile-photo"},
			"serviceRelationships": []map[string]interface{}{
				{
					"identifier":       "urn:li:userGeneratedContent",
					"relationshipType": "OWNER",
				},
			},
		},
	}

	var registerResp struct {
		Value struct {
			UploadMechanism struct {
				HttpUpload struct {
					UploadURL string `json:"uploadUrl"`
				} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
			} `json:"uploadMechanism"`
			Asset string `json:"asset"`
		} `json:"value"`
	}

	if err := s.c.Post(client.EndpointProfilePhotoRegister, registerPayload, &registerResp); err != nil {
		return fmt.Errorf("register photo upload: %w", err)
	}

	uploadURL := registerResp.Value.UploadMechanism.HttpUpload.UploadURL
	assetURN := registerResp.Value.Asset
	if uploadURL == "" || assetURN == "" {
		return fmt.Errorf("invalid upload registration response")
	}

	// Step 2: upload binary
	if err := s.c.PutBinary(uploadURL, data, contentType); err != nil {
		return fmt.Errorf("upload photo binary: %w", err)
	}

	// Step 3: associate with profile
	photoPath := fmt.Sprintf(client.EndpointProfilePhoto, profileID)
	return s.c.Post(photoPath, map[string]interface{}{
		"displayImageUrn": assetURN,
	}, nil)
}

// --- helpers ---

// extractProfileFromEnvelope tries to find a profile in a normalized Voyager envelope.
func extractProfileFromEnvelope(env voyagerNormalizedEnvelope) (*models.Profile, error) {
	for _, raw := range env.Included {
		var vp voyagerFullProfile
		if err := json.Unmarshal(raw, &vp); err != nil {
			continue
		}
		if vp.MiniProfile.EntityURN != "" {
			return mapVoyagerProfile(vp), nil
		}
	}
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
