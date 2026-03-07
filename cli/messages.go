package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Manage LinkedIn messages and conversations",
	}
	cmd.AddCommand(
		newMessagesListCmd(),
		newMessagesReadCmd(),
		newMessagesSendCmd(),
		newMessagesUnreadCmd(),
		newMessagesMarkReadCmd(),
		newMessagesStarCmd(),
		newMessagesUnstarCmd(),
		newMessagesArchiveCmd(),
		newMessagesUnarchiveCmd(),
		newMessagesDeleteCmd(),
		newMessagesDeleteConversationCmd(),
	)
	return cmd
}

func newMessagesListCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all conversations",
		Example: `  linked messages list
  linked messages list --count 50 -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			convs, err := li.Messaging.ListConversations(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(convs)
			}

			if len(convs.Items) == 0 {
				p.Warn("No conversations found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(convs.Items))
				for _, c := range convs.Items {
					participants := participantNames(c.Participants)
					lastMsg := ""
					if c.LastMessage != nil {
						lastMsg = truncate(c.LastMessage.Body, 60)
					}
					unread := ""
					if c.Unread {
						unread = "●"
					}
					rows = append(rows, []string{unread, participants, lastMsg, c.UpdatedAt, c.ID})
				}
				p.Table([]string{"", "Participants", "Last Message", "Updated", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Conversations (%d)", convs.Pagination.Total))
			for _, c := range convs.Items {
				unread := "  "
				if c.Unread {
					unread = "● "
				}
				p.Printf("%s%s  [%s]\n", unread, participantNames(c.Participants), c.UpdatedAt)
				if c.LastMessage != nil {
					p.Printf("    %s\n", truncate(c.LastMessage.Body, 100))
				}
				p.Printf("    ID: %s\n\n", c.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of conversations")
	return cmd
}

func newMessagesReadCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "read <conversation-id>",
		Short: "Read messages in a conversation thread",
		Example: `  linked messages read 2-abc123def456
  linked messages read 2-abc123def456 --count 50 -o json`,
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

			msgs, err := li.Messaging.GetConversation(args[0], start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(msgs)
			}

			p.Header(fmt.Sprintf("Conversation %s (%d messages)", args[0], msgs.Pagination.Total))
			for _, m := range msgs.Items {
				p.Printf("  [%s] %s %s:\n", m.SentAt, m.SenderProfile.FirstName, m.SenderProfile.LastName)
				p.Printf("    %s\n\n", m.Body)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of messages")
	return cmd
}

func newMessagesSendCmd() *cobra.Command {
	var conversationURN string
	var recipientURNs []string

	cmd := &cobra.Command{
		Use:   "send <message>",
		Short: "Send a message",
		Example: `  linked messages send "Hello!" --conversation 2-abc123def456
  linked messages send "Hi there" --to urn:li:member:12345678`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}

			if conversationURN == "" && len(recipientURNs) == 0 {
				return fmt.Errorf("provide --conversation <id> to reply or --to <urn> to start a new conversation")
			}

			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			input := models.SendMessageInput{
				ConversationURN: conversationURN,
				RecipientURNs:   recipientURNs,
				Body:            strings.Join(args, " "),
			}

			if err := li.Messaging.SendMessage(input); err != nil {
				return err
			}

			p.Success("Message sent")
			return nil
		},
	}
	cmd.Flags().StringVar(&conversationURN, "conversation", "", "Conversation ID to reply to")
	cmd.Flags().StringSliceVar(&recipientURNs, "to", nil, "Recipient profile URN(s) for a new conversation")
	return cmd
}

func newMessagesUnreadCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "unread",
		Short: "List unread conversations",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			convs, err := li.Messaging.ListUnread(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(convs)
			}

			if len(convs.Items) == 0 {
				p.Success("No unread messages")
				return nil
			}

			p.Header(fmt.Sprintf("Unread (%d)", convs.Pagination.Total))
			for _, c := range convs.Items {
				p.Printf("  %s  [%s]\n", participantNames(c.Participants), c.UpdatedAt)
				if c.LastMessage != nil {
					p.Printf("    %s\n", truncate(c.LastMessage.Body, 100))
				}
				p.Printf("    ID: %s\n\n", c.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of conversations")
	return cmd
}

func newMessagesMarkReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mark-read <conversation-id>",
		Short: "Mark a conversation as read",
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
			if err := li.Messaging.MarkRead(args[0]); err != nil {
				return err
			}
			p.Success("Conversation marked as read")
			return nil
		},
	}
}

func newMessagesStarCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "star <conversation-id>",
		Short: "Star (bookmark) a conversation",
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
			if err := li.Messaging.StarConversation(args[0]); err != nil {
				return err
			}
			p.Success("Conversation starred")
			return nil
		},
	}
}

func newMessagesUnstarCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unstar <conversation-id>",
		Short: "Remove star from a conversation",
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
			if err := li.Messaging.UnstarConversation(args[0]); err != nil {
				return err
			}
			p.Success("Conversation unstarred")
			return nil
		},
	}
}

func newMessagesArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <conversation-id>",
		Short: "Archive a conversation",
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
			if err := li.Messaging.ArchiveConversation(args[0]); err != nil {
				return err
			}
			p.Success("Conversation archived")
			return nil
		},
	}
}

func newMessagesUnarchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <conversation-id>",
		Short: "Restore an archived conversation",
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
			if err := li.Messaging.UnarchiveConversation(args[0]); err != nil {
				return err
			}
			p.Success("Conversation unarchived")
			return nil
		},
	}
}

func newMessagesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <conversation-id> <message-urn>",
		Short: "Delete a specific message from a conversation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			if err := li.Messaging.DeleteMessage(args[0], args[1]); err != nil {
				return err
			}
			p.Success("Message deleted")
			return nil
		},
	}
}

func newMessagesDeleteConversationCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-conversation <conversation-id>",
		Short: "Delete an entire conversation",
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
			if err := li.Messaging.DeleteConversation(args[0]); err != nil {
				return err
			}
			p.Success("Conversation deleted")
			return nil
		},
	}
}

// participantNames returns a comma-joined list of participant names.
func participantNames(participants []models.Profile) string {
	names := make([]string, 0, len(participants))
	for _, pr := range participants {
		names = append(names, pr.FirstName+" "+pr.LastName)
	}
	return strings.Join(names, ", ")
}
