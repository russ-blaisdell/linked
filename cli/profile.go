package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "View and update LinkedIn profiles",
	}
	cmd.AddCommand(
		newProfileGetCmd(),
		newProfileUpdateCmd(),
		newProfileSkillsCmd(),
		newProfileContactCmd(),
	)
	return cmd
}

func newProfileGetCmd() *cobra.Command {
	var profileID string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a LinkedIn profile (own or another member's)",
		Example: `  linked profile get
  linked profile get --urn john-doe
  linked profile get -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			var profile *models.Profile
			if profileID == "" {
				profile, err = li.Profile.GetMe()
			} else {
				profile, err = li.Profile.GetProfile(profileID)
			}
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(profile)
			}
			printProfile(p, profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileID, "urn", "", "Public profile ID of the member to fetch (omit for your own)")
	return cmd
}

func newProfileUpdateCmd() *cobra.Command {
	var headline, summary, location string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update your LinkedIn profile",
		Example: `  linked profile update --headline "Senior Engineer at Acme"
  linked profile update --summary "Passionate about building great products"
  linked profile update --location "San Francisco, CA"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}

			update := models.ProfileUpdate{
				Headline: headline,
				Summary:  summary,
				Location: location,
			}
			if err := li.Profile.UpdateProfile(me.ProfileID, update); err != nil {
				return err
			}

			p.Success("Profile updated successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&headline, "headline", "", "New headline")
	cmd.Flags().StringVar(&summary, "summary", "", "New summary / about text")
	cmd.Flags().StringVar(&location, "location", "", "New location")
	return cmd
}

func newProfileSkillsCmd() *cobra.Command {
	var profileID string
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List skills on a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			var profile *models.Profile
			var fetchErr error
			if profileID == "" {
				profile, fetchErr = li.Profile.GetMe()
			} else {
				profile, fetchErr = li.Profile.GetProfile(profileID)
			}
			if fetchErr != nil {
				return fetchErr
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(profile.Skills)
			}

			if len(profile.Skills) == 0 {
				p.Warn("No skills found on this profile")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(profile.Skills))
				for _, sk := range profile.Skills {
					rows = append(rows, []string{sk.Name, fmt.Sprintf("%d", sk.Endorsements)})
				}
				p.Table([]string{"Skill", "Endorsements"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Skills for %s %s", profile.FirstName, profile.LastName))
			for _, sk := range profile.Skills {
				p.Printf("  %-40s %d endorsements\n", sk.Name, sk.Endorsements)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileID, "urn", "", "Public profile ID (omit for your own)")
	return cmd
}

func newProfileContactCmd() *cobra.Command {
	var profileID string
	cmd := &cobra.Command{
		Use:   "contact",
		Short: "Get contact information for a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			if profileID == "" {
				me, err := li.Profile.GetMe()
				if err != nil {
					return err
				}
				profileID = me.ProfileID
			}

			info, err := li.Profile.GetContactInfo(profileID)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(info)
			}

			p.Header("Contact Information")
			p.Field("Emails", strings.Join(info.Emails, ", "))
			p.Field("Phone", strings.Join(info.PhoneNumbers, ", "))
			p.Field("Twitter", strings.Join(info.TwitterHandles, ", "))
			p.Field("Websites", strings.Join(info.Websites, ", "))
			p.Field("Address", info.Address)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileID, "urn", "", "Public profile ID (omit for your own)")
	return cmd
}

// printProfile renders a profile in pretty format.
func printProfile(p *output.Printer, profile *models.Profile) {
	p.Header(fmt.Sprintf("%s %s", profile.FirstName, profile.LastName))
	p.Field("Profile ID", profile.ProfileID)
	p.Field("Headline", profile.Headline)
	p.Field("Location", profile.Location)
	p.Field("Industry", profile.Industry)
	p.Field("URN", profile.URN)

	if profile.Summary != "" {
		p.Println()
		p.Header("About")
		p.Printf("  %s\n", wordWrap(profile.Summary, 80))
	}

	if len(profile.Experience) > 0 {
		p.Println()
		p.Header("Experience")
		for _, exp := range profile.Experience {
			dates := exp.StartDate
			if exp.Current {
				dates += " – Present"
			} else if exp.EndDate != "" {
				dates += " – " + exp.EndDate
			}
			p.Printf("  %s at %s", exp.Title, exp.CompanyName)
			if dates != "" {
				p.Printf("  (%s)", dates)
			}
			p.Println()
		}
	}

	if len(profile.Education) > 0 {
		p.Println()
		p.Header("Education")
		for _, edu := range profile.Education {
			p.Printf("  %s", edu.SchoolName)
			if edu.Degree != "" {
				p.Printf(" — %s", edu.Degree)
			}
			if edu.FieldOfStudy != "" {
				p.Printf(" in %s", edu.FieldOfStudy)
			}
			p.Println()
		}
	}
}

// wordWrap inserts newlines to wrap text at maxWidth characters.
func wordWrap(text string, maxWidth int) string {
	words := strings.Fields(text)
	var lines []string
	var current strings.Builder
	for _, word := range words {
		if current.Len()+len(word)+1 > maxWidth && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString("  ") // indent continuation lines
		}
		if current.Len() > 2 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}
