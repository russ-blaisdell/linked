package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/output"
)

func newJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage LinkedIn jobs",
	}
	cmd.AddCommand(
		newJobsGetCmd(),
		newJobsSavedCmd(),
		newJobsSaveCmd(),
		newJobsUnsaveCmd(),
		newJobsAppliedCmd(),
		newJobsRecommendedCmd(),
		newJobsCompanyCmd(),
	)
	return cmd
}

func newJobsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <job-id>",
		Short: "Get details for a job posting",
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

			job, err := li.Jobs.GetJob(args[0])
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(job)
			}

			p.Header(job.Title)
			p.Field("Company", job.Company.Name)
			p.Field("Location", job.Location)
			if job.Remote {
				p.Field("Remote", "Yes")
			}
			p.Field("Employment", job.EmploymentType)
			p.Field("Experience", job.ExperienceLevel)
			p.Field("Posted", job.PostedAt)
			p.Field("Expires", job.ExpiresAt)
			p.Field("Job ID", job.ID)
			if job.Description != "" {
				p.Println()
				p.Header("Description")
				p.Printf("  %s\n", wordWrap(job.Description, 80))
			}
			return nil
		},
	}
}

func newJobsSavedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "saved",
		Short: "List your saved jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Jobs.ListSavedJobs(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No saved jobs")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, j := range result.Items {
					rows = append(rows, []string{j.Title, j.Company.Name, j.Location, j.ID})
				}
				p.Table([]string{"Title", "Company", "Location", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Saved Jobs (%d)", result.Pagination.Total))
			for _, j := range result.Items {
				p.Printf("  %s at %s  —  %s  (ID: %s)\n", j.Title, j.Company.Name, j.Location, j.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newJobsSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save <job-id>",
		Short: "Save a job posting",
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
			if err := li.Jobs.SaveJob(args[0]); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Job %s saved", args[0]))
			return nil
		},
	}
}

func newJobsUnsaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unsave <job-id>",
		Short: "Remove a job from saved jobs",
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
			if err := li.Jobs.UnsaveJob(args[0]); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Job %s removed from saved", args[0]))
			return nil
		},
	}
}

func newJobsRecommendedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "recommended",
		Short: "List LinkedIn's recommended jobs for you",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}
			result, err := li.Jobs.GetRecommendedJobs(start, count)
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}
			if len(result.Items) == 0 {
				p.Warn("No recommended jobs found")
				return nil
			}
			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, j := range result.Items {
					rows = append(rows, []string{j.Title, j.Company.Name, j.Location, j.ID})
				}
				p.Table([]string{"Title", "Company", "Location", "ID"}, rows)
				return nil
			}
			p.Header(fmt.Sprintf("Recommended Jobs (%d)", result.Pagination.Total))
			for _, j := range result.Items {
				p.Printf("  %s at %s  —  %s  (ID: %s)\n", j.Title, j.Company.Name, j.Location, j.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newJobsCompanyCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "company <company-urn>",
		Short: "Search jobs posted by a specific company",
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
			result, err := li.Jobs.SearchJobsByCompany(args[0], start, count)
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}
			if len(result.Items) == 0 {
				p.Warn("No jobs found for this company")
				return nil
			}
			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, j := range result.Items {
					remote := ""
					if j.Remote {
						remote = "Remote"
					}
					rows = append(rows, []string{j.Title, j.Location, remote, j.ID})
				}
				p.Table([]string{"Title", "Location", "Remote", "ID"}, rows)
				return nil
			}
			p.Header(fmt.Sprintf("Company Jobs (%d)", result.Pagination.Total))
			for _, j := range result.Items {
				loc := j.Location
				if j.Remote {
					loc += " (Remote)"
				}
				p.Printf("  %s  —  %s  (ID: %s)\n", j.Title, loc, j.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newJobsAppliedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "applied",
		Short: "List jobs you have applied to",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Jobs.ListAppliedJobs(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No applied jobs found")
				return nil
			}

			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, j := range result.Items {
					rows = append(rows, []string{j.Title, j.Company.Name, j.PostedAt, j.ID})
				}
				p.Table([]string{"Title", "Company", "Posted", "ID"}, rows)
				return nil
			}

			p.Header(fmt.Sprintf("Applied Jobs (%d)", result.Pagination.Total))
			for _, j := range result.Items {
				p.Printf("  %s at %s  (ID: %s)\n", j.Title, j.Company.Name, j.ID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}
