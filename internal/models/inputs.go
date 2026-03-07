package models

// ProfileUpdate holds mutable top-level fields for updating a profile.
type ProfileUpdate struct {
	Headline string `json:"headline,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Location string `json:"location,omitempty"`
}

// ExperienceInput holds parameters for adding or updating a position.
type ExperienceInput struct {
	Title       string
	CompanyName string
	CompanyURN  string
	StartYear   int
	StartMonth  int
	EndYear     int
	EndMonth    int
	Current     bool
	Description string
	Location    string
}

// EducationInput holds parameters for adding or updating an education entry.
type EducationInput struct {
	SchoolName   string
	SchoolURN    string
	Degree       string
	FieldOfStudy string
	StartYear    int
	EndYear      int
	Description  string
}

// SkillInput holds parameters for adding a skill.
type SkillInput struct {
	Name string
}

// CertificationInput holds parameters for adding or updating a certification.
type CertificationInput struct {
	Name        string
	Authority   string
	LicenseNum  string
	URL         string
	StartYear   int
	StartMonth  int
	EndYear     int
	EndMonth    int
}

// LanguageInput holds parameters for adding a language.
type LanguageInput struct {
	Name        string
	Proficiency string // ELEMENTARY, LIMITED_WORKING, PROFESSIONAL_WORKING, FULL_PROFESSIONAL, NATIVE_OR_BILINGUAL
}

// VolunteerInput holds parameters for adding or updating a volunteer experience.
type VolunteerInput struct {
	Role         string
	Organization string
	Cause        string
	StartYear    int
	StartMonth   int
	EndYear      int
	EndMonth     int
	Current      bool
	Description  string
}

// ProjectInput holds parameters for adding or updating a project.
type ProjectInput struct {
	Title       string
	Description string
	URL         string
	StartYear   int
	StartMonth  int
	EndYear     int
	EndMonth    int
	Current     bool
}

// PublicationInput holds parameters for adding or updating a publication.
type PublicationInput struct {
	Name        string
	Publisher   string
	Year        int
	Month       int
	URL         string
	Description string
}

// HonorInput holds parameters for adding a honor or award.
type HonorInput struct {
	Title       string
	Issuer      string
	Year        int
	Month       int
	Description string
}

// CourseInput holds parameters for adding a course.
type CourseInput struct {
	Name        string
	Number      string
	Occupation  string
}

// OpenToWorkInput holds parameters for setting Open to Work status.
type OpenToWorkInput struct {
	JobTypes        []string // FULL_TIME, PART_TIME, CONTRACT, INTERNSHIP, etc.
	Locations       []string
	Title           string
	PreferenceTypes []string // OPEN_TO_WORK, HIRING
}

// SendMessageInput holds parameters for sending a message.
type SendMessageInput struct {
	ConversationURN string   // existing conversation URN (for replies)
	RecipientURNs   []string // profile URNs (for new conversations)
	Body            string
}

// ConnectionRequestInput holds parameters for a connection request.
type ConnectionRequestInput struct {
	ProfileURN string
	Note       string
}

// CreatePostInput holds parameters for creating a text post.
type CreatePostInput struct {
	Body       string
	Visibility string // PUBLIC, CONNECTIONS
}

// CreatePostWithImageInput holds parameters for creating a post with an image.
type CreatePostWithImageInput struct {
	Body        string
	ImagePath   string
	ImageAltText string
	Visibility  string
}

// EditPostInput holds parameters for editing an existing post.
type EditPostInput struct {
	Body       string
	Visibility string
}

// RecommendationRequestInput holds parameters for requesting a recommendation.
type RecommendationRequestInput struct {
	RecipientProfileURN string
	Message             string
	Relationship        string // COLLEAGUE, MANAGER, REPORT, etc.
}

// SearchPeopleInput holds search filters for people search.
type SearchPeopleInput struct {
	Keywords string
	Company  string
	Title    string
	School   string
	Location string
	Network  []string // FIRST, SECOND, THIRD
	Start    int
	Count    int
}

// SearchJobsInput holds search filters for job search.
type SearchJobsInput struct {
	Keywords        string
	Location        string
	Remote          bool
	Company         string
	ExperienceLevel string
	EmploymentType  string
	Start           int
	Count           int
}
