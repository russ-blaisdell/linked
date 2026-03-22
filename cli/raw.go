package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
)

func newRawCmd() *cobra.Command {
	var method, data string

	cmd := &cobra.Command{
		Use:   "raw <path>",
		Short: "Make a raw authenticated request to LinkedIn",
		Long: `Fetch any LinkedIn API path with full authentication and return the raw response.
Useful for exploring and debugging API endpoints.

The path is appended to https://www.linkedin.com — pass everything after that.
For GraphQL endpoints the Accept header is automatically set to application/json.`,
		Example: `  linked raw /voyager/api/me
  linked raw "/voyager/api/graphql?queryId=...&variables=(...)"
  linked raw --method POST --data '{"body":{"text":"hello"}}' /voyager/api/messengerMessages?action=createMessage`,
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

			var body []byte
			var status int

			if method == "POST" && data != "" {
				body, status, err = c.RawPost(args[0], []byte(data))
			} else {
				body, status, err = c.RawGet(args[0])
			}
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

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method (GET or POST)")
	cmd.Flags().StringVarP(&data, "data", "d", "", "JSON body for POST requests")
	return cmd
}
