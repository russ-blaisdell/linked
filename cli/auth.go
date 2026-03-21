package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/config"
	"github.com/russ-blaisdell/linked/internal/harparser"
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
	var harPath string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure LinkedIn session credentials",
		Long: `Store your LinkedIn session cookies for linked to use.

Option 1 — Import from a HAR file (recommended):
  linked auth setup --har ~/Downloads/www.linkedin.com.har

  The HAR file provides browser fingerprint data (User-Agent, etc.) that
  prevents session revocation. If the HAR doesn't contain cookies, you
  will be prompted for them.

  How to capture a HAR file:
    1. Open LinkedIn in Chrome and log in.
    2. Open DevTools → Network tab.
    3. Browse a few pages on LinkedIn.
    4. Right-click in the Network tab → "Save all as HAR with content".

Option 2 — Enter cookies manually:
  linked auth setup

  You need four cookies from DevTools → Application → Cookies:
    1. li_at       — your main LinkedIn session token
    2. JSESSIONID  — used for CSRF validation (e.g. "ajax:1234567890abcdef")
    3. bcookie     — browser fingerprint (e.g. "v=2&xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
    4. bscookie    — secure browser fingerprint (e.g. "v=1&timestamp&uuid&token")

Credentials are stored at:
  ~/.openclaw/credentials/linkedin/<profile>/creds.json  (mode 0600)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}

			var creds *models.Credentials

			if harPath != "" {
				result, err := harparser.Parse(harPath)
				if err != nil {
					return fmt.Errorf("parsing HAR file: %w", err)
				}

				if result.HasCookies {
					creds = &models.Credentials{
						LiAt:       result.LiAt,
						JSESSIONID: result.JSESSIONID,
						Bcookie:    result.Bcookie,
						Bscookie:   result.Bscookie,
					}
					p.Success("Extracted cookies from HAR file")
				} else {
					p.Warn("HAR file does not contain cookies — prompting manually")
					fmt.Println()
					creds, err = promptForCookies()
					if err != nil {
						return err
					}
				}
				creds.Fingerprint = result.Fingerprint
				p.Success(fmt.Sprintf("Extracted browser fingerprint: %s", result.Fingerprint.UserAgent))
			} else {
				p.Header("LinkedIn Credentials Setup")
				fmt.Println()
				creds, err = promptForCookies()
				if err != nil {
					return err
				}
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

	cmd.Flags().StringVar(&harPath, "har", "", "path to a HAR file captured from linkedin.com")
	return cmd
}

func promptForCookies() (*models.Credentials, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("  li_at value: ")
	liAt, _ := reader.ReadString('\n')
	liAt = strings.TrimSpace(liAt)
	if liAt == "" {
		return nil, fmt.Errorf("li_at is required")
	}

	fmt.Print("  JSESSIONID value: ")
	jsessionid, _ := reader.ReadString('\n')
	jsessionid = strings.TrimSpace(jsessionid)
	if jsessionid == "" {
		return nil, fmt.Errorf("JSESSIONID is required")
	}

	fmt.Print("  bcookie value: ")
	bcookie, _ := reader.ReadString('\n')
	bcookie = strings.TrimSpace(bcookie)
	if bcookie == "" {
		return nil, fmt.Errorf("bcookie is required")
	}

	fmt.Print("  bscookie value: ")
	bscookie, _ := reader.ReadString('\n')
	bscookie = strings.TrimSpace(bscookie)
	if bscookie == "" {
		return nil, fmt.Errorf("bscookie is required")
	}

	return &models.Credentials{
		LiAt:       liAt,
		JSESSIONID: jsessionid,
		Bcookie:    bcookie,
		Bscookie:   bscookie,
	}, nil
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
