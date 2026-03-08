package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage LinkedIn credentials",
	}
	cmd.AddCommand(
		newAuthSetupCmd(),
		newAuthWhoamiCmd(),
		newAuthRemoveCmd(),
		newAuthListCmd(),
	)
	return cmd
}

func newAuthSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure LinkedIn session credentials",
		Long: `Store your LinkedIn session cookies for linked to use.

You need four cookies from your browser (Chrome / Firefox / Safari):

  1. li_at       — your main LinkedIn session token
  2. JSESSIONID  — used for CSRF validation (e.g. "ajax:1234567890abcdef")
  3. bcookie     — browser fingerprint (e.g. "v=2&xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
  4. bscookie    — secure browser fingerprint (e.g. "v=1&timestamp&uuid&token")

How to get them:
  1. Open LinkedIn in your browser and log in.
  2. Open DevTools → Application → Cookies → https://www.linkedin.com
  3. Copy the values for 'li_at', 'JSESSIONID', 'bcookie', and 'bscookie'.

These are stored at:
  ~/.openclaw/credentials/linkedin/<profile>/creds.json  (mode 0600)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}

			reader := bufio.NewReader(os.Stdin)

			p.Header("LinkedIn Credentials Setup")
			fmt.Println()

			fmt.Print("  li_at value: ")
			liAt, _ := reader.ReadString('\n')
			liAt = strings.TrimSpace(liAt)
			if liAt == "" {
				return fmt.Errorf("li_at is required")
			}

			fmt.Print("  JSESSIONID value: ")
			jsessionid, _ := reader.ReadString('\n')
			jsessionid = strings.TrimSpace(jsessionid)
			if jsessionid == "" {
				return fmt.Errorf("JSESSIONID is required")
			}

			fmt.Print("  bcookie value: ")
			bcookie, _ := reader.ReadString('\n')
			bcookie = strings.TrimSpace(bcookie)
			if bcookie == "" {
				return fmt.Errorf("bcookie is required")
			}

			fmt.Print("  bscookie value: ")
			bscookie, _ := reader.ReadString('\n')
			bscookie = strings.TrimSpace(bscookie)
			if bscookie == "" {
				return fmt.Errorf("bscookie is required")
			}

			creds := &models.Credentials{
				LiAt:       liAt,
				JSESSIONID: jsessionid,
				Bcookie:    bcookie,
				Bscookie:   bscookie,
			}

			// Verify before saving.
			fmt.Println()
			p.Println("  Verifying credentials…")

			c, err := client.New(creds)
			if err != nil {
				return fmt.Errorf("creating client: %w", err)
			}

			li := newLinkedInFromClient(c)
			profile, err := li.Profile.GetMe()
			if err != nil {
				return fmt.Errorf("credentials verification failed: %w", err)
			}

			creds.ProfileID = profile.ProfileID
			if err := config.SaveCredentials(globalProfile, creds); err != nil {
				return fmt.Errorf("saving credentials: %w", err)
			}

			fmt.Println()
			p.Success(fmt.Sprintf("Authenticated as %s %s (%s)", profile.FirstName, profile.LastName, profile.ProfileID))
			return nil
		},
	}
}

func newAuthWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the currently authenticated LinkedIn account",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			profile, err := li.Profile.GetMe()
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(profile)
			}

			p.Header(fmt.Sprintf("%s %s", profile.FirstName, profile.LastName))
			p.Field("Profile ID", profile.ProfileID)
			p.Field("Headline", profile.Headline)
			p.Field("Location", profile.Location)
			p.Field("URN", profile.URN)
			return nil
		},
	}
}

func newAuthRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove stored credentials for a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			if err := config.DeleteCredentials(globalProfile); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Credentials removed for profile %q", globalProfile))
			return nil
		},
	}
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured credential profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			profiles, err := config.ListProfiles()
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(profiles)
			}
			if len(profiles) == 0 {
				p.Warn("No profiles configured — run 'linked auth setup'")
				return nil
			}
			p.Header("Configured profiles:")
			for _, pr := range profiles {
				p.Printf("  %s\n", pr)
			}
			return nil
		},
	}
}
