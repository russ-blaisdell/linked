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
		newProfileExperienceCmd(),
		newProfileEducationCmd(),
		newProfileCertificationsCmd(),
		newProfileLanguagesCmd(),
		newProfileVolunteerCmd(),
		newProfileProjectsCmd(),
		newProfilePublicationsCmd(),
		newProfileHonorsCmd(),
		newProfileCoursesCmd(),
		newProfileOpenToWorkCmd(),
		newProfileCloseToWorkCmd(),
		newProfileWhoViewedCmd(),
		newProfilePhotoCmd(),
	)
	return cmd
}

// ---- Get ----

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

// ---- Update ----

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

// ---- Skills (list) ----

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
					id := sk.ID
					if id == "" {
						id = "-"
					}
					rows = append(rows, []string{sk.Name, fmt.Sprintf("%d", sk.Endorsements), id})
				}
				p.Table([]string{"Skill", "Endorsements", "ID"}, rows)
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

// ---- Contact ----

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

// ---- Experience ----

func newProfileExperienceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experience",
		Short: "Manage work experience on your profile",
	}
	cmd.AddCommand(
		newProfileExperienceListCmd(),
		newProfileExperienceAddCmd(),
		newProfileExperienceUpdateCmd(),
		newProfileExperienceRemoveCmd(),
	)
	return cmd
}

func newProfileExperienceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List work experience on your profile",
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
				return p.JSON(profile.Experience)
			}
			if len(profile.Experience) == 0 {
				p.Warn("No experience entries")
				return nil
			}
			p.Header("Work Experience")
			for _, exp := range profile.Experience {
				dates := exp.StartDate
				if exp.Current {
					dates += " – Present"
				} else if exp.EndDate != "" {
					dates += " – " + exp.EndDate
				}
				p.Printf("  %s at %s  (%s)\n", exp.Title, exp.CompanyName, dates)
				if exp.ID != "" {
					p.Printf("    ID: %s\n", exp.ID)
				}
				p.Println()
			}
			return nil
		},
	}
}

func newProfileExperienceAddCmd() *cobra.Command {
	var title, company, companyURN, location, description string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a work experience entry",
		Example: `  linked profile experience add --title "Engineer" --company "Acme" --start-year 2022 --current
  linked profile experience add --title "Intern" --company "Corp" --start-year 2021 --end-year 2022`,
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
			id, err := li.Profile.AddExperience(me.ProfileID, models.ExperienceInput{
				Title: title, CompanyName: company, CompanyURN: companyURN,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth,
				Current: current, Description: description, Location: location,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Experience added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Job title (required)")
	cmd.Flags().StringVar(&company, "company", "", "Company name (required)")
	cmd.Flags().StringVar(&companyURN, "company-urn", "", "Company URN (optional)")
	cmd.Flags().StringVar(&location, "location", "", "Location")
	cmd.Flags().StringVar(&description, "description", "", "Role description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month (1-12)")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month (1-12)")
	cmd.Flags().BoolVar(&current, "current", false, "Currently in this role")
	return cmd
}

func newProfileExperienceUpdateCmd() *cobra.Command {
	var title, company, companyURN, location, description string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:   "update <position-id>",
		Short: "Update a work experience entry",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdateExperience(me.ProfileID, args[0], models.ExperienceInput{
				Title: title, CompanyName: company, CompanyURN: companyURN,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth,
				Current: current, Description: description, Location: location,
			}); err != nil {
				return err
			}
			p.Success("Experience updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Job title")
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	cmd.Flags().StringVar(&companyURN, "company-urn", "", "Company URN")
	cmd.Flags().StringVar(&location, "location", "", "Location")
	cmd.Flags().StringVar(&description, "description", "", "Role description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month (1-12)")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month (1-12)")
	cmd.Flags().BoolVar(&current, "current", false, "Currently in this role")
	return cmd
}

func newProfileExperienceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <position-id>",
		Short: "Remove a work experience entry",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteExperience(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Experience entry removed")
			return nil
		},
	}
}

// ---- Education ----

func newProfileEducationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "education",
		Short: "Manage education on your profile",
	}
	cmd.AddCommand(
		newProfileEducationListCmd(),
		newProfileEducationAddCmd(),
		newProfileEducationUpdateCmd(),
		newProfileEducationRemoveCmd(),
	)
	return cmd
}

func newProfileEducationListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List education entries on your profile",
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
				return p.JSON(profile.Education)
			}
			if len(profile.Education) == 0 {
				p.Warn("No education entries")
				return nil
			}
			p.Header("Education")
			for _, edu := range profile.Education {
				p.Printf("  %s", edu.SchoolName)
				if edu.Degree != "" {
					p.Printf(" — %s", edu.Degree)
				}
				if edu.FieldOfStudy != "" {
					p.Printf(" in %s", edu.FieldOfStudy)
				}
				if edu.StartDate != "" || edu.EndDate != "" {
					p.Printf("  (%s – %s)", edu.StartDate, edu.EndDate)
				}
				p.Println()
				if edu.ID != "" {
					p.Printf("    ID: %s\n", edu.ID)
				}
				p.Println()
			}
			return nil
		},
	}
}

func newProfileEducationAddCmd() *cobra.Command {
	var school, schoolURN, degree, field, description string
	var startYear, endYear int
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add an education entry",
		Example: `  linked profile education add --school "MIT" --degree "BS" --field "Computer Science" --start-year 2018 --end-year 2022`,
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
			id, err := li.Profile.AddEducation(me.ProfileID, models.EducationInput{
				SchoolName: school, SchoolURN: schoolURN, Degree: degree,
				FieldOfStudy: field, StartYear: startYear, EndYear: endYear,
				Description: description,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Education added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&school, "school", "", "School name (required)")
	cmd.Flags().StringVar(&schoolURN, "school-urn", "", "School URN (optional)")
	cmd.Flags().StringVar(&degree, "degree", "", "Degree name")
	cmd.Flags().StringVar(&field, "field", "", "Field of study")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	return cmd
}

func newProfileEducationUpdateCmd() *cobra.Command {
	var school, schoolURN, degree, field, description string
	var startYear, endYear int
	cmd := &cobra.Command{
		Use:   "update <education-id>",
		Short: "Update an education entry",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdateEducation(me.ProfileID, args[0], models.EducationInput{
				SchoolName: school, SchoolURN: schoolURN, Degree: degree,
				FieldOfStudy: field, StartYear: startYear, EndYear: endYear,
				Description: description,
			}); err != nil {
				return err
			}
			p.Success("Education updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&school, "school", "", "School name")
	cmd.Flags().StringVar(&schoolURN, "school-urn", "", "School URN")
	cmd.Flags().StringVar(&degree, "degree", "", "Degree name")
	cmd.Flags().StringVar(&field, "field", "", "Field of study")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	return cmd
}

func newProfileEducationRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <education-id>",
		Short: "Remove an education entry",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteEducation(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Education entry removed")
			return nil
		},
	}
}

// ---- Certifications ----

func newProfileCertificationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certifications",
		Short: "Manage certifications on your profile",
	}
	cmd.AddCommand(
		newProfileCertificationsAddCmd(),
		newProfileCertificationsUpdateCmd(),
		newProfileCertificationsRemoveCmd(),
	)
	return cmd
}

func newProfileCertificationsAddCmd() *cobra.Command {
	var name, authority, licenseNum, url string
	var startYear, startMonth, endYear, endMonth int
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a certification",
		Example: `  linked profile certifications add --name "AWS Solutions Architect" --authority "Amazon" --start-year 2023`,
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
			id, err := li.Profile.AddCertification(me.ProfileID, models.CertificationInput{
				Name: name, Authority: authority, LicenseNum: licenseNum, URL: url,
				StartYear: startYear, StartMonth: startMonth, EndYear: endYear, EndMonth: endMonth,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Certification added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Certification name (required)")
	cmd.Flags().StringVar(&authority, "authority", "", "Issuing organization")
	cmd.Flags().StringVar(&licenseNum, "license", "", "License/credential ID")
	cmd.Flags().StringVar(&url, "url", "", "Credential URL")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Issue year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Issue month (1-12)")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "Expiry year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "Expiry month (1-12)")
	return cmd
}

func newProfileCertificationsUpdateCmd() *cobra.Command {
	var name, authority, licenseNum, url string
	var startYear, startMonth, endYear, endMonth int
	cmd := &cobra.Command{
		Use:   "update <cert-id>",
		Short: "Update a certification",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdateCertification(me.ProfileID, args[0], models.CertificationInput{
				Name: name, Authority: authority, LicenseNum: licenseNum, URL: url,
				StartYear: startYear, StartMonth: startMonth, EndYear: endYear, EndMonth: endMonth,
			}); err != nil {
				return err
			}
			p.Success("Certification updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Certification name")
	cmd.Flags().StringVar(&authority, "authority", "", "Issuing organization")
	cmd.Flags().StringVar(&licenseNum, "license", "", "License/credential ID")
	cmd.Flags().StringVar(&url, "url", "", "Credential URL")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Issue year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Issue month (1-12)")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "Expiry year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "Expiry month (1-12)")
	return cmd
}

func newProfileCertificationsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <cert-id>",
		Short: "Remove a certification",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteCertification(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Certification removed")
			return nil
		},
	}
}

// ---- Languages ----

func newProfileLanguagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "languages",
		Short: "Manage languages on your profile",
	}
	cmd.AddCommand(
		newProfileLanguagesAddCmd(),
		newProfileLanguagesRemoveCmd(),
	)
	return cmd
}

func newProfileLanguagesAddCmd() *cobra.Command {
	var name, proficiency string
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a language",
		Example: `  linked profile languages add --name "Spanish" --proficiency PROFESSIONAL_WORKING`,
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
			id, err := li.Profile.AddLanguage(me.ProfileID, models.LanguageInput{Name: name, Proficiency: proficiency})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Language added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Language name (required)")
	cmd.Flags().StringVar(&proficiency, "proficiency", "PROFESSIONAL_WORKING", "Proficiency: ELEMENTARY, LIMITED_WORKING, PROFESSIONAL_WORKING, FULL_PROFESSIONAL, NATIVE_OR_BILINGUAL")
	return cmd
}

func newProfileLanguagesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <language-id>",
		Short: "Remove a language",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteLanguage(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Language removed")
			return nil
		},
	}
}

// ---- Volunteer ----

func newProfileVolunteerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volunteer",
		Short: "Manage volunteer experience on your profile",
	}
	cmd.AddCommand(
		newProfileVolunteerListCmd(),
		newProfileVolunteerAddCmd(),
		newProfileVolunteerUpdateCmd(),
		newProfileVolunteerRemoveCmd(),
	)
	return cmd
}

func newProfileVolunteerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List volunteer experience",
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
			p.Header(fmt.Sprintf("Volunteer Experience for %s %s", me.FirstName, me.LastName))
			p.Warn("Volunteer experience is included in your full profile — use 'linked profile get' to view all sections.")
			return nil
		},
	}
}

func newProfileVolunteerAddCmd() *cobra.Command {
	var role, org, cause, description string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a volunteer experience",
		Example: `  linked profile volunteer add --role "Mentor" --org "Code.org" --cause "Education" --start-year 2022 --current`,
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
			id, err := li.Profile.AddVolunteer(me.ProfileID, models.VolunteerInput{
				Role: role, Organization: org, Cause: cause, Description: description,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth, Current: current,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Volunteer experience added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "Your role (required)")
	cmd.Flags().StringVar(&org, "org", "", "Organization name (required)")
	cmd.Flags().StringVar(&cause, "cause", "", "Cause")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month")
	cmd.Flags().BoolVar(&current, "current", false, "Currently volunteering")
	return cmd
}

func newProfileVolunteerUpdateCmd() *cobra.Command {
	var role, org, cause, description string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:   "update <volunteer-id>",
		Short: "Update a volunteer experience",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdateVolunteer(me.ProfileID, args[0], models.VolunteerInput{
				Role: role, Organization: org, Cause: cause, Description: description,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth, Current: current,
			}); err != nil {
				return err
			}
			p.Success("Volunteer experience updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "Your role")
	cmd.Flags().StringVar(&org, "org", "", "Organization name")
	cmd.Flags().StringVar(&cause, "cause", "", "Cause")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month")
	cmd.Flags().BoolVar(&current, "current", false, "Currently volunteering")
	return cmd
}

func newProfileVolunteerRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <volunteer-id>",
		Short: "Remove a volunteer experience",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteVolunteer(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Volunteer experience removed")
			return nil
		},
	}
}

// ---- Projects ----

func newProfileProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects on your profile",
	}
	cmd.AddCommand(
		newProfileProjectsAddCmd(),
		newProfileProjectsUpdateCmd(),
		newProfileProjectsRemoveCmd(),
	)
	return cmd
}

func newProfileProjectsAddCmd() *cobra.Command {
	var title, description, url string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a project",
		Example: `  linked profile projects add --title "My App" --description "A mobile app" --url "https://myapp.com" --start-year 2023 --current`,
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
			id, err := li.Profile.AddProject(me.ProfileID, models.ProjectInput{
				Title: title, Description: description, URL: url,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth, Current: current,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Project added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Project title (required)")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().StringVar(&url, "url", "", "Project URL")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month")
	cmd.Flags().BoolVar(&current, "current", false, "Currently working on this project")
	return cmd
}

func newProfileProjectsUpdateCmd() *cobra.Command {
	var title, description, url string
	var startYear, startMonth, endYear, endMonth int
	var current bool
	cmd := &cobra.Command{
		Use:   "update <project-id>",
		Short: "Update a project",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdateProject(me.ProfileID, args[0], models.ProjectInput{
				Title: title, Description: description, URL: url,
				StartYear: startYear, StartMonth: startMonth,
				EndYear: endYear, EndMonth: endMonth, Current: current,
			}); err != nil {
				return err
			}
			p.Success("Project updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Project title")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVar(&url, "url", "", "Project URL")
	cmd.Flags().IntVar(&startYear, "start-year", 0, "Start year")
	cmd.Flags().IntVar(&startMonth, "start-month", 0, "Start month")
	cmd.Flags().IntVar(&endYear, "end-year", 0, "End year")
	cmd.Flags().IntVar(&endMonth, "end-month", 0, "End month")
	cmd.Flags().BoolVar(&current, "current", false, "Currently working on this project")
	return cmd
}

func newProfileProjectsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <project-id>",
		Short: "Remove a project",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteProject(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Project removed")
			return nil
		},
	}
}

// ---- Publications ----

func newProfilePublicationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publications",
		Short: "Manage publications on your profile",
	}
	cmd.AddCommand(
		newProfilePublicationsAddCmd(),
		newProfilePublicationsUpdateCmd(),
		newProfilePublicationsRemoveCmd(),
	)
	return cmd
}

func newProfilePublicationsAddCmd() *cobra.Command {
	var name, publisher, url, description string
	var year, month int
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a publication",
		Example: `  linked profile publications add --name "My Paper" --publisher "ACM" --year 2023 --url "https://doi.org/..."`,
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
			id, err := li.Profile.AddPublication(me.ProfileID, models.PublicationInput{
				Name: name, Publisher: publisher, URL: url,
				Description: description, Year: year, Month: month,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Publication added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Publication title (required)")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Publisher")
	cmd.Flags().StringVar(&url, "url", "", "Publication URL")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&year, "year", 0, "Publication year")
	cmd.Flags().IntVar(&month, "month", 0, "Publication month (1-12)")
	return cmd
}

func newProfilePublicationsUpdateCmd() *cobra.Command {
	var name, publisher, url, description string
	var year, month int
	cmd := &cobra.Command{
		Use:   "update <publication-id>",
		Short: "Update a publication",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.UpdatePublication(me.ProfileID, args[0], models.PublicationInput{
				Name: name, Publisher: publisher, URL: url,
				Description: description, Year: year, Month: month,
			}); err != nil {
				return err
			}
			p.Success("Publication updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Publication title")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Publisher")
	cmd.Flags().StringVar(&url, "url", "", "Publication URL")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&year, "year", 0, "Publication year")
	cmd.Flags().IntVar(&month, "month", 0, "Publication month (1-12)")
	return cmd
}

func newProfilePublicationsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <publication-id>",
		Short: "Remove a publication",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeletePublication(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Publication removed")
			return nil
		},
	}
}

// ---- Honors ----

func newProfileHonorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "honors",
		Short: "Manage honors and awards on your profile",
	}
	cmd.AddCommand(
		newProfileHonorsAddCmd(),
		newProfileHonorsRemoveCmd(),
	)
	return cmd
}

func newProfileHonorsAddCmd() *cobra.Command {
	var title, issuer, description string
	var year, month int
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add an honor or award",
		Example: `  linked profile honors add --title "Dean's List" --issuer "MIT" --year 2022`,
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
			id, err := li.Profile.AddHonor(me.ProfileID, models.HonorInput{
				Title: title, Issuer: issuer, Description: description,
				Year: year, Month: month,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Honor added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Honor/award title (required)")
	cmd.Flags().StringVar(&issuer, "issuer", "", "Issuing organization")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&year, "year", 0, "Year issued")
	cmd.Flags().IntVar(&month, "month", 0, "Month issued (1-12)")
	return cmd
}

func newProfileHonorsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <honor-id>",
		Short: "Remove an honor or award",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteHonor(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Honor removed")
			return nil
		},
	}
}

// ---- Courses ----

func newProfileCoursesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "courses",
		Short: "Manage courses on your profile",
	}
	cmd.AddCommand(
		newProfileCoursesAddCmd(),
		newProfileCoursesRemoveCmd(),
	)
	return cmd
}

func newProfileCoursesAddCmd() *cobra.Command {
	var name, number, occupation string
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a course",
		Example: `  linked profile courses add --name "Machine Learning" --number "CS229" --occupation "Student"`,
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
			id, err := li.Profile.AddCourse(me.ProfileID, models.CourseInput{
				Name: name, Number: number, Occupation: occupation,
			})
			if err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Course added (ID: %s)", id))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Course name (required)")
	cmd.Flags().StringVar(&number, "number", "", "Course number")
	cmd.Flags().StringVar(&occupation, "occupation", "", "Associated occupation/company")
	return cmd
}

func newProfileCoursesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <course-id>",
		Short: "Remove a course",
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
			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}
			if err := li.Profile.DeleteCourse(me.ProfileID, args[0]); err != nil {
				return err
			}
			p.Success("Course removed")
			return nil
		},
	}
}

// ---- Open to Work ----

func newProfileOpenToWorkCmd() *cobra.Command {
	var title string
	var jobTypes, locations []string
	cmd := &cobra.Command{
		Use:     "open-to-work",
		Short:   "Set your Open to Work status",
		Example: `  linked profile open-to-work --job-types FULL_TIME,CONTRACT --locations "Remote" --title "Software Engineer"`,
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
			if err := li.Profile.SetOpenToWork(me.URN, models.OpenToWorkInput{
				JobTypes:        jobTypes,
				Locations:       locations,
				Title:           title,
				PreferenceTypes: []string{"OPEN_TO_WORK"},
			}); err != nil {
				return err
			}
			p.Success("Open to Work status set")
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Desired job title")
	cmd.Flags().StringSliceVar(&jobTypes, "job-types", []string{"FULL_TIME"}, "Job types: FULL_TIME, PART_TIME, CONTRACT, INTERNSHIP, TEMPORARY")
	cmd.Flags().StringSliceVar(&locations, "locations", nil, "Preferred locations (e.g. \"Remote\", \"San Francisco, CA\")")
	return cmd
}

func newProfileCloseToWorkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close-to-work",
		Short: "Remove your Open to Work status",
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
			if err := li.Profile.ClearOpenToWork(me.ProfileID); err != nil {
				return err
			}
			p.Success("Open to Work status removed")
			return nil
		},
	}
}

// ---- Who Viewed ----

func newProfileWhoViewedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "who-viewed",
		Short: "See who viewed your profile recently",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Profile.GetWhoViewed(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No profile viewers found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, v := range result.Items {
					rows = append(rows, []string{
						v.Profile.FirstName + " " + v.Profile.LastName,
						v.Profile.Headline,
						v.ViewedAt,
						v.Profile.ProfileID,
					})
				}
				p.Table([]string{"Name", "Headline", "Viewed At", "Profile ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Who Viewed Your Profile (%d total)", result.Pagination.Total))
			for _, v := range result.Items {
				p.Printf("  %s %s  —  %s\n", v.Profile.FirstName, v.Profile.LastName, v.Profile.Headline)
				p.Printf("    Viewed: %s\n\n", v.ViewedAt)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

// ---- Profile Photo ----

func newProfilePhotoCmd() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:     "photo",
		Short:   "Upload a new profile photo",
		Example: `  linked profile photo --file ~/Pictures/headshot.jpg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
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
			if err := li.Profile.UploadProfilePhoto(me.ProfileID, filePath); err != nil {
				return err
			}
			p.Success("Profile photo updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to image file (jpg, png, gif, webp)")
	return cmd
}

// ---- print helpers ----

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
			current.WriteString("  ")
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
