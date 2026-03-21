package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
)

func newRawCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "raw <path>",
		Short: "Make a raw authenticated GET request to LinkedIn",
		Long: `Fetch any LinkedIn API path with full authentication and return the raw response.
Useful for exploring and debugging API endpoints.

The path is appended to https://www.linkedin.com — pass everything after that.
For GraphQL endpoints the Accept header is automatically set to application/json.`,
		Example: `  linked raw /voyager/api/me
  linked raw "/voyager/api/graphql?queryId=voyagerRelationshipsDashInvitationViews.57e1286f887065b96393b947e09ef04c&variables=(q:receivedInvitation,start:0,count:5)"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.LoadCredentials(globalProfile)
			if err != nil {
				return err
			}
			c, err := client.New(creds)
			if err != nil {
				return fmt.Errorf("creating client: %w", err)
			}

			body, status, err := c.RawGet(args[0])
			if err != nil {
				return err
			}

			// Pretty-print JSON if possible, otherwise dump raw.
			var pretty json.RawMessage
			if json.Unmarshal(body, &pretty) == nil {
				formatted, _ := json.MarshalIndent(pretty, "", "  ")
				fmt.Printf("HTTP %d\n%s\n", status, formatted)
			} else {
				fmt.Printf("HTTP %d\n%s\n", status, body)
			}
			return nil
		},
	}
}
