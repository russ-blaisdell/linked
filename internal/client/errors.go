package client

import "fmt"

// LinkedInError represents an error returned by the LinkedIn API.
type LinkedInError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *LinkedInError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("linkedin api error %d: %s (%s)", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("linkedin api error %d: %s", e.StatusCode, e.Message)
}

// ErrUnauthorized is returned when credentials are missing or expired.
var ErrUnauthorized = &LinkedInError{StatusCode: 401, Message: "unauthorized — run 'linked auth setup' to configure credentials"}

// ErrNotFound is returned for 404 responses.
var ErrNotFound = &LinkedInError{StatusCode: 404, Message: "resource not found"}

// ErrRateLimit is returned on 429 responses.
var ErrRateLimit = &LinkedInError{StatusCode: 429, Message: "rate limited by LinkedIn — please wait before retrying"}

// ErrForbidden is returned on 403 responses.
var ErrForbidden = &LinkedInError{StatusCode: 403, Message: "forbidden — you do not have permission to access this resource"}

// apiError constructs a LinkedInError from a status code and body.
func apiError(statusCode int, body []byte) *LinkedInError {
	switch statusCode {
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 429:
		return ErrRateLimit
	default:
		return &LinkedInError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("unexpected status %d", statusCode),
			Details:    string(body),
		}
	}
}
