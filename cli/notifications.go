package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/output"
)

func newNotificationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "View LinkedIn notifications",
	}
	cmd.AddCommand(
		newNotificationsListCmd(),
		newNotificationsMarkReadCmd(),
		newNotificationsMarkAllReadCmd(),
		newNotificationsCountCmd(),
	)
	return cmd
}

func newNotificationsListCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Notifications.List(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No notifications")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, n := range result.Items {
					read := "✓"
					if !n.Read {
						read = "●"
					}
					rows = append(rows, []string{read, n.Type, truncate(n.Body, 80), n.CreatedAt, n.ID})
				}
				p.Table([]string{"", "Type", "Body", "Time", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Notifications (%d)", result.Pagination.Total))
			for _, n := range result.Items {
				read := "  "
				if !n.Read {
					read = "● "
				}
				p.Printf("%s[%s] %s  —  %s\n    ID: %s\n\n", read, n.Type, n.Body, n.CreatedAt, n.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of notifications")
	return cmd
}

func newNotificationsMarkReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mark-read <notification-urn>",
		Short: "Mark a notification as read",
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
			if err := li.Notifications.MarkRead(args[0]); err != nil {
				return err
			}
			p.Success("Notification marked as read")
			return nil
		},
	}
}

func newNotificationsMarkAllReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mark-all-read",
		Short: "Mark all notifications as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Notifications.MarkAllRead(); err != nil {
				return err
			}
			p.Success("All notifications marked as read")
			return nil
		},
	}
}

func newNotificationsCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count",
		Short: "Get the number of unread notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			badge, err := li.Notifications.GetBadgeCount()
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(badge)
			}
			p.Printf("Unread notifications: %d\n", badge.UnreadCount)
			return nil
		},
	}
}
