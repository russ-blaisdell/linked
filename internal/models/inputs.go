package models

// ProfileUpdate holds mutable fields for updating a profile.
type ProfileUpdate struct {
	Headline string `json:"headline,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Location string `json:"location,omitempty"`
}

// SendMessageInput holds parameters for sending a message.
type SendMessageInput struct {
	ConversationURN string // existing conversation URN (for replies)
	RecipientURNs   []string // profile URNs (for new conversations)
	Body            string
}

// ConnectionRequestInput holds parameters for a connection request.
type ConnectionRequestInput struct {
	ProfileURN string
	Note       string
}

// CreatePostInput holds parameters for creating a post.
type CreatePostInput struct {
	Body       string
	Visibility string // PUBLIC, CONNECTIONS
	ImagePath  string // optional local file path
}

// RecommendationRequestInput holds parameters for requesting a recommendation.
type RecommendationRequestInput struct {
	RecipientProfileURN string
	Message             string
	Relationship        string // COLLEAGUE, MANAGER, REPORT, etc.
}

// SearchPeopleInput holds search filters for people search.
type SearchPeopleInput struct {
	Keywords    string
	Company     string
	Title       string
	School      string
	Location    string
	Network     []string // FIRST, SECOND, THIRD
	Start       int
	Count       int
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
