package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search LinkedIn for people, jobs, companies, and posts",
	}
	cmd.AddCommand(
		newSearchPeopleCmd(),
		newSearchJobsCmd(),
		newSearchCompaniesCmd(),
		newSearchPostsCmd(),
	)
	return cmd
}

func newSearchPeopleCmd() *cobra.Command {
	var company, title, school, location string
	var network []string
	var start, count int

	cmd := &cobra.Command{
		Use:   "people <keywords>",
		Short: "Search LinkedIn members",
		Example: `  linked search people "software engineer"
  linked search people "product manager" --company google --network FIRST,SECOND
  linked search people "designer" --title "Senior" -o table`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			input := models.SearchPeopleInput{
				Keywords: strings.Join(args, " "),
				Company:  company,
				Title:    title,
				School:   school,
				Location: location,
				Network:  network,
				Start:    start,
				Count:    count,
			}

			results, err := li.Search.SearchPeople(input)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(results)
			}

			if len(results.Items) == 0 {
				p.Warn("No results found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(results.Items))
				for _, r := range results.Items {
					rows = append(rows, []string{
						r.Profile.FirstName + " " + r.Profile.LastName,
						r.Profile.Headline,
						r.Profile.ProfileID,
						r.Distance,
					})
				}
				p.Table([]string{"Name", "Headline", "Profile ID", "Distance"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("People (%d results)", results.Pagination.Total))
			for _, r := range results.Items {
				p.Printf("  %s %s  •  %s\n", r.Profile.FirstName, r.Profile.LastName, r.Profile.Headline)
				p.Printf("    %s  (distance: %s)\n\n", r.Profile.ProfileID, r.Distance)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "Filter by current company")
	cmd.Flags().StringVar(&title, "title", "", "Filter by job title")
	cmd.Flags().StringVar(&school, "school", "", "Filter by school")
	cmd.Flags().StringVar(&location, "location", "", "Filter by location")
	cmd.Flags().StringSliceVar(&network, "network", nil, "Filter by network distance: FIRST, SECOND, THIRD")
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newSearchJobsCmd() *cobra.Command {
	var location, company, experienceLevel, employmentType string
	var remote bool
	var start, count int

	cmd := &cobra.Command{
		Use:   "jobs <keywords>",
		Short: "Search LinkedIn job postings",
		Example: `  linked search jobs "golang engineer"
  linked search jobs "product manager" --location "New York" --remote
  linked search jobs "designer" --experience-level MID_SENIOR_LEVEL -o table`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			input := models.SearchJobsInput{
				Keywords:        strings.Join(args, " "),
				Location:        location,
				Remote:          remote,
				Company:         company,
				ExperienceLevel: experienceLevel,
				EmploymentType:  employmentType,
				Start:           start,
				Count:           count,
			}

			results, err := li.Search.SearchJobs(input)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(results)
			}

			if len(results.Items) == 0 {
				p.Warn("No jobs found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(results.Items))
				for _, j := range results.Items {
					remote := ""
					if j.Remote {
						remote = "Remote"
					}
					rows = append(rows, []string{j.Title, j.Company.Name, j.Location, remote, j.ID})
				}
				p.Table([]string{"Title", "Company", "Location", "Remote", "Job ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Jobs (%d results)", results.Pagination.Total))
			for _, j := range results.Items {
				loc := j.Location
				if j.Remote {
					loc += " (Remote)"
				}
				p.Printf("  %s at %s\n", j.Title, j.Company.Name)
				p.Printf("    %s  —  ID: %s\n\n", loc, j.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location filter")
	cmd.Flags().BoolVar(&remote, "remote", false, "Remote jobs only")
	cmd.Flags().StringVar(&company, "company", "", "Filter by company")
	cmd.Flags().StringVar(&experienceLevel, "experience-level", "", "Experience level (ENTRY_LEVEL, MID_SENIOR_LEVEL, DIRECTOR, etc.)")
	cmd.Flags().StringVar(&employmentType, "employment-type", "", "Employment type (FULL_TIME, PART_TIME, CONTRACT, etc.)")
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newSearchCompaniesCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "companies <keywords>",
		Short: "Search LinkedIn companies",
		Example: `  linked search companies "Anthropic"
  linked search companies "fintech startup" -o table`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			results, err := li.Search.SearchCompanies(strings.Join(args, " "), start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(results)
			}
			if len(results.Items) == 0 {
				p.Warn("No companies found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(results.Items))
				for _, co := range results.Items {
					rows = append(rows, []string{co.Name, co.Industry, co.ID})
				}
				p.Table([]string{"Name", "Industry", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Companies (%d results)", results.Pagination.Total))
			for _, co := range results.Items {
				p.Printf("  %s  (%s)  —  ID: %s\n", co.Name, co.Industry, co.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newSearchPostsCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "posts <keywords>",
		Short: "Search LinkedIn posts and content",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			results, err := li.Search.SearchPosts(strings.Join(args, " "), start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(results)
			}
			p.Header(fmt.Sprintf("Posts (%d results)", results.Pagination.Total))
			for _, post := range results.Items {
				p.Printf("  [%s] %s\n  URN: %s\n\n", post.PostedAt, truncate(post.Body, 120), post.URN)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
