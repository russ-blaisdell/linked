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

// participantNames returns a comma-joined list of participant names.
func participantNames(participants []models.Profile) string {
	names := make([]string, 0, len(participants))
	for _, pr := range participants {
		names = append(names, pr.FirstName+" "+pr.LastName)
	}
	return strings.Join(names, ", ")
}
