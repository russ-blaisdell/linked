package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/output"
)

func newCompaniesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "View and follow LinkedIn companies",
	}
	cmd.AddCommand(
		newCompaniesGetCmd(),
		newCompaniesFollowCmd(),
		newCompaniesUnfollowCmd(),
		newCompaniesPostsCmd(),
		newCompaniesEmployeesCmd(),
	)
	return cmd
}

func newCompaniesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <company-id>",
		Short: "Get company information by universal name or ID",
		Example: `  linked companies get anthropic
  linked companies get google -o json`,
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

			co, err := li.Companies.GetCompany(args[0])
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(co)
			}

			p.Header(co.Name)
			p.Field("Industry", co.Industry)
			p.Field("Headquarters", co.Headquarters)
			p.Field("Employees", co.EmployeeCount)
			p.Field("Website", co.Website)
			p.Field("ID", co.ID)
			p.Field("URN", co.URN)
			if co.Headline != "" {
				p.Println()
				p.Printf("  %s\n", co.Headline)
			}
			return nil
		},
	}
}

func newCompaniesFollowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "follow <company-urn>",
		Short: "Follow a company",
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
			if err := li.Companies.FollowCompany(args[0]); err != nil {
				return err
			}
			p.Success("Now following company")
			return nil
		},
	}
}

func newCompaniesUnfollowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unfollow <company-urn>",
		Short: "Unfollow a company",
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
			if err := li.Companies.UnfollowCompany(args[0]); err != nil {
				return err
			}
			p.Success("Unfollowed company")
			return nil
		},
	}
}

func newCompaniesEmployeesCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "employees <company-id>",
		Short: "Search LinkedIn members who work at a company",
		Example: `  linked companies employees google
  linked companies employees anthropic --count 50 -o table`,
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
			result, err := li.Companies.GetCompanyEmployees(args[0], start, count)
			if err != nil {
				return err
			}
			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}
			if len(result.Items) == 0 {
				p.Warn("No employees found")
				return nil
			}
			if p.Format() == output.FormatTable {
				rows := make([][]string, 0, len(result.Items))
				for _, e := range result.Items {
					rows = append(rows, []string{e.Profile.FirstName + " " + e.Profile.LastName, e.Profile.Headline, e.Profile.ProfileID})
				}
				p.Table([]string{"Name", "Headline", "Profile ID"}, rows)
				return nil
			}
			p.Header(fmt.Sprintf("Employees at %s (%d)", args[0], result.Pagination.Total))
			for _, e := range result.Items {
				p.Printf("  %s %s  —  %s\n    %s\n\n", e.Profile.FirstName, e.Profile.LastName, e.Profile.Headline, e.Profile.ProfileID)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of results")
	return cmd
}

func newCompaniesPostsCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "posts <company-urn>",
		Short: "Get recent posts from a company",
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

			result, err := li.Companies.GetCompanyPosts(args[0], start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No posts found")
				return nil
			}

			p.Header(fmt.Sprintf("Company Posts (%d)", result.Pagination.Total))
			for _, post := range result.Items {
				p.Printf("  [%s] 👍 %d  💬 %d  🔁 %d\n", post.PostedAt, post.LikeCount, post.CommentCount, post.ShareCount)
				p.Printf("    %s\n\n", truncate(post.Body, 200))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of posts")
	return cmd
}
