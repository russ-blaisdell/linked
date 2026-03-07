package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
)

// PostsService handles LinkedIn post and feed operations.
type PostsService struct {
	c *client.Client
}

// NewPostsService returns a new PostsService.
func NewPostsService(c *client.Client) *PostsService {
	return &PostsService{c: c}
}

// GetFeed returns the authenticated user's home feed.
func (s *PostsService) GetFeed(start, count int) (*models.PagedPosts, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":     "chronologicalFeed",
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
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
					LikeCount    int `json:"likeCount"`
					CommentCount int `json:"commentCount"`
					ShareCount   int `json:"shareCount"`
				} `json:"socialDetail,omitempty"`
				CreatedAt int64 `json:"createdAt"`
				Actor     struct {
					Urn  string `json:"urn"`
					Name struct {
						Text string `json:"text"`
					} `json:"name,omitempty"`
				} `json:"actor,omitempty"`
			} `json:"com.linkedin.voyager.feed.render.UpdateV2"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointFeed, params, &raw); err != nil {
		return nil, fmt.Errorf("get feed: %w", err)
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

// GetMemberPosts returns posts authored by a specific member.
func (s *PostsService) GetMemberPosts(memberURN string, start, count int) (*models.PagedPosts, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	params := map[string]string{
		"q":       "memberShareFeed",
		"authors": fmt.Sprintf("List(%s)", memberURN),
		"start":   fmt.Sprintf("%d", start),
		"count":   fmt.Sprintf("%d", count),
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
					LikeCount    int `json:"likeCount"`
					CommentCount int `json:"commentCount"`
					ShareCount   int `json:"shareCount"`
				} `json:"socialDetail,omitempty"`
				CreatedAt int64 `json:"createdAt"`
			} `json:"com.linkedin.voyager.feed.render.UpdateV2"`
		} `json:"elements"`
		Paging struct {
			Start int `json:"start"`
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"paging"`
	}

	if err := s.c.Get(client.EndpointFeed, params, &raw); err != nil {
		return nil, fmt.Errorf("get member posts: %w", err)
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

// GetPost returns a single post by URN.
func (s *PostsService) GetPost(postURN string) (*models.Post, error) {
	path := fmt.Sprintf(client.EndpointUGCPost, postURN)
	var raw struct {
		EntityURN       string `json:"entityUrn"`
		LifecycleState  string `json:"lifecycleState"`
		Author          string `json:"author"`
		SpecificContent struct {
			Share struct {
				Commentary struct {
					Text string `json:"text"`
				} `json:"shareCommentary"`
			} `json:"com.linkedin.ugc.ShareContent"`
		} `json:"specificContent"`
		CreatedAt int64 `json:"created,omitempty"`
	}
	if err := s.c.Get(path, nil, &raw); err != nil {
		return nil, fmt.Errorf("get post %q: %w", postURN, err)
	}
	return &models.Post{
		URN:      raw.EntityURN,
		Body:     raw.SpecificContent.Share.Commentary.Text,
		PostedAt: msToTime(raw.CreatedAt),
	}, nil
}

// DeletePost deletes a post by URN.
func (s *PostsService) DeletePost(postURN string) error {
	path := fmt.Sprintf(client.EndpointUGCPost, postURN)
	return s.c.Delete(path)
}

// EditPost updates the text of an existing post.
func (s *PostsService) EditPost(postURN string, input models.EditPostInput) error {
	path := fmt.Sprintf(client.EndpointUGCPost, postURN)
	visibility := input.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}
	payload := map[string]interface{}{
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": input.Body,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": visibility,
		},
	}
	return s.c.Put(path, payload, nil)
}

// CreatePost creates a new text post.
func (s *PostsService) CreatePost(authorURN string, input models.CreatePostInput) error {
	visibility := input.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}
	payload := map[string]interface{}{
		"author":         authorURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": input.Body,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": visibility,
		},
	}
	return s.c.Post(client.EndpointPostCreate, payload, nil)
}

// CreatePostWithImage creates a post with an attached image.
// It follows LinkedIn's three-step upload process.
func (s *PostsService) CreatePostWithImage(authorURN string, input models.CreatePostWithImageInput) error {
	data, err := os.ReadFile(input.ImagePath)
	if err != nil {
		return fmt.Errorf("reading image: %w", err)
	}

	contentType := "image/jpeg"
	ext := strings.ToLower(filepath.Ext(input.ImagePath))
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
			"owner":   authorURN,
			"recipes": []string{"urn:li:digitalmediaRecipe:feedshare-image"},
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

	if err := s.c.Post(client.EndpointMediaUpload+"?action=registerUpload", registerPayload, &registerResp); err != nil {
		return fmt.Errorf("register image upload: %w", err)
	}

	uploadURL := registerResp.Value.UploadMechanism.HttpUpload.UploadURL
	assetURN := registerResp.Value.Asset
	if uploadURL == "" || assetURN == "" {
		return fmt.Errorf("invalid upload registration response")
	}

	// Step 2: upload binary
	if err := s.c.PutBinary(uploadURL, data, contentType); err != nil {
		return fmt.Errorf("upload image binary: %w", err)
	}

	// Step 3: create post with media
	visibility := input.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}
	altText := input.ImageAltText
	if altText == "" {
		altText = "Image"
	}
	payload := map[string]interface{}{
		"author":         authorURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": input.Body,
				},
				"shareMediaCategory": "IMAGE",
				"media": []map[string]interface{}{
					{
						"status":      "READY",
						"description": map[string]interface{}{"text": altText},
						"media":       assetURN,
					},
				},
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": visibility,
		},
	}
	return s.c.Post(client.EndpointPostCreate, payload, nil)
}

// LikePost likes a post by URN.
func (s *PostsService) LikePost(postURN string) error {
	path := fmt.Sprintf(client.EndpointLike, postURN)
	payload := map[string]interface{}{
		"actor": postURN,
	}
	return s.c.Post(path, payload, nil)
}

// UnlikePost removes a like from a post.
func (s *PostsService) UnlikePost(postURN string) error {
	path := fmt.Sprintf(client.EndpointLike, postURN)
	return s.c.Delete(path)
}

// CommentOnPost adds a comment to a post.
func (s *PostsService) CommentOnPost(postURN, authorURN, body string) error {
	path := fmt.Sprintf(client.EndpointComments, postURN)
	payload := map[string]interface{}{
		"actor": authorURN,
		"message": map[string]interface{}{
			"attributes": []interface{}{},
			"text":       body,
		},
	}
	return s.c.Post(path, payload, nil)
}

// DeleteComment deletes a comment from a post.
func (s *PostsService) DeleteComment(postURN, commentURN string) error {
	path := fmt.Sprintf(client.EndpointCommentByID, postURN, commentURN)
	return s.c.Delete(path)
}

// LikeComment likes a specific comment on a post.
func (s *PostsService) LikeComment(postURN, commentURN, actorURN string) error {
	path := fmt.Sprintf(client.EndpointCommentLike, postURN, commentURN)
	return s.c.Post(path, map[string]interface{}{"actor": actorURN}, nil)
}

// SharePost reshares a post.
func (s *PostsService) SharePost(postURN, authorURN, commentary string) error {
	payload := map[string]interface{}{
		"author":         authorURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": commentary,
				},
				"shareMediaCategory": "RESHARE",
				"media": []map[string]interface{}{
					{
						"status":      "READY",
						"originalUrn": postURN,
					},
				},
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}
	return s.c.Post(client.EndpointPostCreate, payload, nil)
}

// GetComments returns comments on a post.
func (s *PostsService) GetComments(postURN string, start, count int) ([]models.Comment, error) {
	if count == 0 {
		count = client.DefaultCount
	}
	path := fmt.Sprintf(client.EndpointComments, postURN)
	params := map[string]string{
		"start": fmt.Sprintf("%d", start),
		"count": fmt.Sprintf("%d", count),
	}

	var raw struct {
		Elements []struct {
			EntityURN string `json:"entityUrn"`
			Actor     string `json:"actor"`
			Message   struct {
				Text string `json:"text"`
			} `json:"message,omitempty"`
			LikeCount int   `json:"likeCount,omitempty"`
			CreatedAt int64 `json:"createdAt,omitempty"`
		} `json:"elements"`
	}

	if err := s.c.Get(path, params, &raw); err != nil {
		return nil, fmt.Errorf("get comments for %q: %w", postURN, err)
	}

	comments := make([]models.Comment, 0, len(raw.Elements))
	for _, el := range raw.Elements {
		comments = append(comments, models.Comment{
			URN:       el.EntityURN,
			Body:      el.Message.Text,
			LikeCount: el.LikeCount,
			PostedAt:  msToTime(el.CreatedAt),
		})
	}
	return comments, nil
}
