package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newConnectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connections",
		Short: "Manage LinkedIn connections and invitations",
	}
	cmd.AddCommand(
		newConnectionsListCmd(),
		newConnectionsPendingCmd(),
		newConnectionsSentCmd(),
		newConnectionsRequestCmd(),
		newConnectionsAcceptCmd(),
		newConnectionsIgnoreCmd(),
		newConnectionsWithdrawCmd(),
		newConnectionsRemoveCmd(),
		newConnectionsFollowCmd(),
		newConnectionsUnfollowCmd(),
		newConnectionsMutualCmd(),
	)
	return cmd
}

func newConnectionsListCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your 1st-degree connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Connections.ListConnections(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No connections found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, c := range result.Items {
					rows = append(rows, []string{c.FirstName + " " + c.LastName, c.Headline, c.ProfileID})
				}
				p.Table([]string{"Name", "Headline", "Profile ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Connections (%d total)", result.Pagination.Total))
			for _, c := range result.Items {
				p.Printf("  %s %s  —  %s\n    %s\n\n", c.FirstName, c.LastName, c.Headline, c.ProfileID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of connections")
	return cmd
}

func newConnectionsPendingCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "pending",
		Short: "List pending received connection invitations",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Connections.ListPendingInvitations(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No pending invitations")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, inv := range result.Items {
					rows = append(rows, []string{
						inv.FromProfile.FirstName + " " + inv.FromProfile.LastName,
						inv.FromProfile.Headline,
						inv.SentAt,
						inv.ID,
					})
				}
				p.Table([]string{"From", "Headline", "Sent", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Pending Invitations (%d)", result.Pagination.Total))
			for _, inv := range result.Items {
				p.Printf("  From: %s %s  (%s)\n", inv.FromProfile.FirstName, inv.FromProfile.LastName, inv.SentAt)
				if inv.FromProfile.Headline != "" {
					p.Printf("    %s\n", inv.FromProfile.Headline)
				}
				if inv.Insight != "" {
					p.Printf("    %s\n", inv.Insight)
				}
				if inv.Message != "" {
					p.Printf("    \"%s\"\n", inv.Message)
				}
				p.Printf("    ID: %s\n\n", inv.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newConnectionsSentCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "sent",
		Short: "List sent connection invitations",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Connections.ListSentInvitations(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No sent invitations")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, inv := range result.Items {
					rows = append(rows, []string{
						inv.ToProfile.FirstName + " " + inv.ToProfile.LastName,
						inv.ToProfile.Headline,
						inv.SentAt,
						inv.ID,
					})
				}
				p.Table([]string{"To", "Headline", "Sent", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Sent Invitations (%d)", result.Pagination.Total))
			for _, inv := range result.Items {
				p.Printf("  To: %s %s  (%s)  ID: %s\n", inv.ToProfile.FirstName, inv.ToProfile.LastName, inv.SentAt, inv.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newConnectionsRequestCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use:   "request <profile-urn>",
		Short: "Send a connection request",
		Example: `  linked connections request urn:li:member:12345678
  linked connections request urn:li:member:12345678 --note "Hi, I'd love to connect!"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			input := models.ConnectionRequestInput{
				ProfileURN: args[0],
				Note:       note,
			}
			if err := li.Connections.SendConnectionRequest(input); err != nil {
				return err
			}
			p.Success("Connection request sent")
			return nil
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "Personal note to include with the request (max 300 chars)")
	return cmd
}

func newConnectionsAcceptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept <invitation-id>",
		Short: "Accept a received connection invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.AcceptInvitation(args[0], ""); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Invitation %s accepted", args[0]))
			return nil
		},
	}
}

func newConnectionsIgnoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ignore <invitation-id>",
		Short: "Ignore a received connection invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.IgnoreInvitation(args[0]); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Invitation %s ignored", args[0]))
			return nil
		},
	}
}

func newConnectionsWithdrawCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw <invitation-urn>",
		Short: "Withdraw a sent connection invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.WithdrawInvitation(args[0]); err != nil {
				return err
			}
			p.Success("Invitation withdrawn")
			return nil
		},
	}
}

func newConnectionsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <profile-urn>",
		Short: "Remove a 1st-degree connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.RemoveConnection(args[0]); err != nil {
				return err
			}
			p.Success("Connection removed")
			return nil
		},
	}
}

func newConnectionsMutualCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "mutual <profile-urn>",
		Short: "List mutual connections with a LinkedIn member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			result, err := li.Connections.GetMutualConnections(args[0], start, count)
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}
			if len(result.Items) == 0 {
				p.Warn("No mutual connections found")
				return nil
			}
			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, m := range result.Items {
					rows = append(rows, []string{m.Profile.FirstName + " " + m.Profile.LastName, m.Profile.Headline, m.Profile.ProfileID})
				}
				p.Table([]string{"Name", "Headline", "Profile ID"}, rows)
				return nil
			}
			p.Header(fmt.Sprintf("Mutual Connections (%d)", result.Pagination.Total))
			for _, m := range result.Items {
				p.Printf("  %s %s  —  %s\n    %s\n\n", m.Profile.FirstName, m.Profile.LastName, m.Profile.Headline, m.Profile.ProfileID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newConnectionsFollowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "follow <profile-urn>",
		Short: "Follow a LinkedIn member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.FollowProfile(args[0]); err != nil {
				return err
			}
			p.Success("Now following " + args[0])
			return nil
		},
	}
}

func newConnectionsUnfollowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unfollow <profile-urn>",
		Short: "Unfollow a LinkedIn member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Connections.UnfollowProfile(args[0]); err != nil {
				return err
			}
			p.Success("Unfollowed " + args[0])
			return nil
		},
	}
}
