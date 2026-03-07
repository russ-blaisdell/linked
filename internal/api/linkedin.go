// Package api provides high-level LinkedIn Voyager API services.
// Each service corresponds to a domain of LinkedIn functionality.
package api

import (
	"github.com/russ-blaisdell/linked/internal/client"
)

// LinkedIn is the top-level service container giving access to all
// LinkedIn API domains.
type LinkedIn struct {
	Profile         *ProfileService
	Search          *SearchService
	Messaging       *MessagingService
	Connections     *ConnectionsService
	Jobs            *JobsService
	Companies       *CompaniesService
	Posts           *PostsService
	Recommendations *RecommendationsService
	Notifications   *NotificationsService
}

// New wires all services around the given client.
func New(c *client.Client) *LinkedIn {
	return &LinkedIn{
		Profile:         NewProfileService(c),
		Search:          NewSearchService(c),
		Messaging:       NewMessagingService(c),
		Connections:     NewConnectionsService(c),
		Jobs:            NewJobsService(c),
		Companies:       NewCompaniesService(c),
		Posts:           NewPostsService(c),
		Recommendations: NewRecommendationsService(c),
		Notifications:   NewNotificationsService(c),
	}
}
