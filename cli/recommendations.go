package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newRecommendationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommendations",
		Short: "View and manage LinkedIn recommendations",
	}
	cmd.AddCommand(
		newRecommendationsReceivedCmd(),
		newRecommendationsGivenCmd(),
		newRecommendationsRequestCmd(),
		newRecommendationsHideCmd(),
		newRecommendationsShowCmd(),
	)
	return cmd
}

func newRecommendationsReceivedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "received",
		Short: "List recommendations you have received",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Recommendations.ListReceived(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No received recommendations")
				return nil
			}

			p.Header(fmt.Sprintf("Received Recommendations (%d)", result.Pagination.Total))
			for _, rec := range result.Items {
				p.Printf("  From: %s %s  (%s)\n", rec.RecommenderProfile.FirstName, rec.RecommenderProfile.LastName, rec.CreatedAt)
				p.Printf("  Status: %s  |  Relationship: %s\n", rec.Status, rec.Relationship)
				p.Printf("    \"%s\"\n\n", wordWrap(rec.Body, 100))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newRecommendationsGivenCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "given",
		Short: "List recommendations you have written",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Recommendations.ListGiven(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No given recommendations")
				return nil
			}

			p.Header(fmt.Sprintf("Given Recommendations (%d)", result.Pagination.Total))
			for _, rec := range result.Items {
				p.Printf("  For: %s %s  (%s)\n", rec.RecommendeeProfile.FirstName, rec.RecommendeeProfile.LastName, rec.CreatedAt)
				p.Printf("    \"%s\"\n\n", truncate(rec.Body, 200))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newRecommendationsRequestCmd() *cobra.Command {
	var message, relationship string
	cmd := &cobra.Command{
		Use:   "request <profile-urn>",
		Short: "Request a recommendation from a connection",
		Example: `  linked recommendations request urn:li:member:12345678
  linked recommendations request urn:li:member:12345678 --relationship COLLEAGUE --message "Hi, would you write me a recommendation?"`,
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

			input := models.RecommendationRequestInput{
				RecipientProfileURN: args[0],
				Message:             message,
				Relationship:        relationship,
			}

			if err := li.Recommendations.RequestRecommendation(input); err != nil {
				return err
			}

			p.Success("Recommendation request sent")
			return nil
		},
	}
	cmd.Flags().StringVar(&message, "message", "", "Personal message to include with the request")
	cmd.Flags().StringVar(&relationship, "relationship", "COLLEAGUE", "Relationship type: COLLEAGUE, MANAGER, REPORT, CLASSMATE, etc.")
	return cmd
}

func newRecommendationsHideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hide <recommendation-urn>",
		Short: "Hide a received recommendation from your profile",
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
			if err := li.Recommendations.HideRecommendation(args[0]); err != nil {
				return err
			}
			p.Success("Recommendation hidden from profile")
			return nil
		},
	}
}

func newRecommendationsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <recommendation-urn>",
		Short: "Show a previously hidden recommendation on your profile",
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
			if err := li.Recommendations.ShowRecommendation(args[0]); err != nil {
				return err
			}
			p.Success("Recommendation now visible on profile")
			return nil
		},
	}
}
