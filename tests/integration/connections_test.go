package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/models"
)

func TestListConnections(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Connections.ListConnections(0, 20)
	if err != nil {
		t.Fatalf("ListConnections() error: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(result.Items))
	}

	first := result.Items[0]
	if first.FirstName == "" {
		t.Error("connection FirstName should not be empty")
	}
	if first.ProfileID == "" {
		t.Error("connection ProfileID should not be empty")
	}
}

func TestListPendingInvitations(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Connections.ListPendingInvitations(0, 20)
	if err != nil {
		t.Fatalf("ListPendingInvitations() error: %v", err)
	}

	if len(result.Items) == 0 {
		t.Fatal("expected at least one pending invitation")
	}

	inv := result.Items[0]
	if inv.ID == "" {
		t.Error("invitation ID should not be empty")
	}
	if inv.FromProfile.FirstName == "" {
		t.Error("invitation FromProfile.FirstName should not be empty")
	}
	if inv.Direction != "INBOUND" {
		t.Errorf("Direction = %q, want INBOUND", inv.Direction)
	}
}

func TestListSentInvitations(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	result, err := li.Connections.ListSentInvitations(0, 20)
	if err != nil {
		t.Fatalf("ListSentInvitations() error: %v", err)
	}

	for _, inv := range result.Items {
		if inv.Direction != "OUTBOUND" {
			t.Errorf("sent invitation Direction = %q, want OUTBOUND", inv.Direction)
		}
	}
}

func TestSendConnectionRequest(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.ConnectionRequestInput{
		ProfileURN: "urn:li:member:987654321",
		Note:       "I'd love to connect and discuss opportunities.",
	}

	if err := li.Connections.SendConnectionRequest(input); err != nil {
		t.Fatalf("SendConnectionRequest() error: %v", err)
	}
}

func TestSendConnectionRequestNoNote(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	input := models.ConnectionRequestInput{
		ProfileURN: "urn:li:member:987654321",
	}

	if err := li.Connections.SendConnectionRequest(input); err != nil {
		t.Fatalf("SendConnectionRequest() without note error: %v", err)
	}
}

func TestAcceptInvitation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Connections.AcceptInvitation("inv001", "secret"); err != nil {
		t.Fatalf("AcceptInvitation() error: %v", err)
	}
}

func TestIgnoreInvitation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Connections.IgnoreInvitation("inv001"); err != nil {
		t.Fatalf("IgnoreInvitation() error: %v", err)
	}
}

func TestWithdrawInvitation(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	if err := li.Connections.WithdrawInvitation("urn:li:invitation:inv001"); err != nil {
		t.Fatalf("WithdrawInvitation() error: %v", err)
	}
}

func TestFollowUnfollow(t *testing.T) {
	s := startServer(t)
	li := newTestLinkedIn(t, s)

	urn := "urn:li:member:987654321"

	if err := li.Connections.FollowProfile(urn); err != nil {
		t.Fatalf("FollowProfile() error: %v", err)
	}
	if err := li.Connections.UnfollowProfile(urn); err != nil {
		t.Fatalf("UnfollowProfile() error: %v", err)
	}
}
