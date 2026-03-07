package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestGetFeed(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Posts.GetFeed(0, 20)
	if err != nil {
		t.Fatalf("GetFeed() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one post in feed")
	}

	post := result.Items[0]
	if post.URN == "" {
		t.Error("post URN should not be empty")
	}
	if post.Body == "" {
		t.Error("post Body should not be empty")
	}
	if post.PostedAt == "" {
		t.Error("post PostedAt should not be empty")
	}
}

func TestCreatePost(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.CreatePostInput{
		Body:       "Excited to share my latest project!",
		Visibility: "PUBLIC",
	}

	if err := li.Posts.CreatePost("urn:li:member:123456789", input); err != nil {
		t.Fatalf("CreatePost() error: %v", err)
	}

	posts := s.CreatedPosts()
	if len(posts) != 1 {
		t.Fatalf("expected 1 created post, got %d", len(posts))
	}
}

func TestCreatePostDefaultVisibility(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.CreatePostInput{
		Body: "Another post without explicit visibility",
	}

	if err := li.Posts.CreatePost("urn:li:member:123456789", input); err != nil {
		t.Fatalf("CreatePost() error: %v", err)
	}
}

func TestLikeUnlikePost(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	urn := "urn:li:activity:aaa001"
	if err := li.Posts.LikePost(urn); err != nil {
		t.Fatalf("LikePost() error: %v", err)
	}
	if err := li.Posts.UnlikePost(urn); err != nil {
		t.Fatalf("UnlikePost() error: %v", err)
	}
}

func TestCommentOnPost(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Posts.CommentOnPost(
		"urn:li:activity:aaa001",
		"urn:li:member:123456789",
		"Great insights, thanks for sharing!",
	); err != nil {
		t.Fatalf("CommentOnPost() error: %v", err)
	}
}

func TestSharePost(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Posts.SharePost(
		"urn:li:activity:aaa001",
		"urn:li:member:123456789",
		"This is worth sharing!",
	); err != nil {
		t.Fatalf("SharePost() error: %v", err)
	}
}

func TestGetComments(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	comments, err := li.Posts.GetComments("urn:li:activity:aaa001", 0, 20)
	if err != nil {
		t.Fatalf("GetComments() error: %v", err)
	}

	if len(comments) == 0 {
		t.Fatal("expected at least one comment")
	}

	c := comments[0]
	if c.Body == "" {
		t.Error("comment body should not be empty")
	}
}
