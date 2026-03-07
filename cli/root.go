// Package cli contains all Cobra command definitions for linked.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/api"
	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
	"github.com/russ-blaisdell/linked/internal/output"
)

var (
	globalOutputFormat string
	globalProfile      string
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "linked",
	Short: "LinkedIn CLI for OpenClaw",
	Long: `linked — LinkedIn command-line interface for OpenClaw.

Interact with LinkedIn from the terminal: search jobs, manage messages,
update your profile, handle connections, and more.

Authentication:
  Run 'linked auth setup' on first use to configure your LinkedIn credentials.

Output formats:
  --output pretty   Human-readable output with colour (default)
  --output json     Machine-readable JSON (used by OpenClaw)
  --output table    Tabular output
`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalOutputFormat, "output", "o", "pretty", "Output format: pretty, json, table")
	rootCmd.PersistentFlags().StringVar(&globalProfile, "profile", "default", "LinkedIn credential profile to use")

	rootCmd.AddCommand(
		newAuthCmd(),
		newProfileCmd(),
		newSearchCmd(),
		newMessagesCmd(),
		newConnectionsCmd(),
		newJobsCmd(),
		newCompaniesCmd(),
		newPostsCmd(),
		newRecommendationsCmd(),
		newNotificationsCmd(),
	)
}

// newPrinter creates a Printer from the global output flag.
func newPrinter() (*output.Printer, error) {
	f, err := output.ParseFormat(globalOutputFormat)
	if err != nil {
		return nil, err
	}
	return output.New(f), nil
}

// newLinkedIn builds an authenticated LinkedIn API client from stored credentials.
func newLinkedIn() (*api.LinkedIn, error) {
	creds, err := config.LoadCredentials(globalProfile)
	if err != nil {
		return nil, err
	}
	c, err := client.New(creds)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	return api.New(c), nil
}

// newLinkedInFromClient wraps an existing client in the LinkedIn service.
func newLinkedInFromClient(c *client.Client) *api.LinkedIn {
	return api.New(c)
}
