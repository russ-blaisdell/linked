package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/internal/output"
)

func newPostsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Create and interact with LinkedIn posts",
	}
	cmd.AddCommand(
		newPostsFeedCmd(),
		newPostsCreateCmd(),
		newPostsLikeCmd(),
		newPostsUnlikeCmd(),
		newPostsCommentCmd(),
		newPostsShareCmd(),
		newPostsCommentsCmd(),
	)
	return cmd
}

func newPostsFeedCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Get your LinkedIn home feed",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := newPrinter()
			if err != nil {
				return err
			}
			li, err := newLinkedIn()
			if err != nil {
				return err
			}

			result, err := li.Posts.GetFeed(start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(result)
			}

			if len(result.Items) == 0 {
				p.Warn("No posts in feed")
				return nil
			}

			p.Header("Home Feed")
			for _, post := range result.Items {
				author := post.AuthorProfile.FirstName + " " + post.AuthorProfile.LastName
				if author == " " {
					author = "Unknown"
				}
				p.Printf("  %s  [%s]\n", author, post.PostedAt)
				p.Printf("  %s\n", truncate(post.Body, 200))
				p.Printf("  👍 %d  💬 %d  🔁 %d  URN: %s\n\n",
					post.LikeCount, post.CommentCount, post.ShareCount, post.URN)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of posts")
	return cmd
}

func newPostsCreateCmd() *cobra.Command {
	var visibility string
	cmd := &cobra.Command{
		Use:   "create <text>",
		Short: "Create a new LinkedIn post",
		Example: `  linked posts create "Excited to share my latest project!"
  linked posts create "Announcing our new product" --visibility CONNECTIONS`,
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

			me, err := li.Profile.GetMe()
			if err != nil {
				return err
			}

			input := models.CreatePostInput{
				Body:       strings.Join(args, " "),
				Visibility: visibility,
			}

			if err := li.Posts.CreatePost(me.URN, input); err != nil {
				return err
			}

			p.Success("Post created successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&visibility, "visibility", "PUBLIC", "Visibility: PUBLIC or CONNECTIONS")
	return cmd
}

func newPostsLikeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "like <post-urn>",
		Short: "Like a post",
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
			if err := li.Posts.LikePost(args[0]); err != nil {
				return err
			}
			p.Success("Post liked")
			return nil
		},
	}
}

func newPostsUnlikeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlike <post-urn>",
		Short: "Remove your like from a post",
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
			if err := li.Posts.UnlikePost(args[0]); err != nil {
				return err
			}
			p.Success("Like removed")
			return nil
		},
	}
}

func newPostsCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comment <post-urn> <text>",
		Short: "Comment on a post",
		Example: `  linked posts comment urn:li:activity:12345 "Great insights!"`,
		Args: cobra.ExactArgs(2),
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

			if err := li.Posts.CommentOnPost(args[0], me.URN, args[1]); err != nil {
				return err
			}
			p.Success("Comment posted")
			return nil
		},
	}
}

func newPostsShareCmd() *cobra.Command {
	var commentary string
	cmd := &cobra.Command{
		Use:   "share <post-urn>",
		Short: "Reshare a post",
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

			if err := li.Posts.SharePost(args[0], me.URN, commentary); err != nil {
				return err
			}
			p.Success(fmt.Sprintf("Post %s reshared", args[0]))
			return nil
		},
	}
	cmd.Flags().StringVar(&commentary, "commentary", "", "Optional text to add when resharing")
	return cmd
}

func newPostsCommentsCmd() *cobra.Command {
	var start, count int
	cmd := &cobra.Command{
		Use:   "comments <post-urn>",
		Short: "Get comments on a post",
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

			comments, err := li.Posts.GetComments(args[0], start, count)
			if err != nil {
				return err
			}

			if p.Format() == output.FormatJSON {
				return p.JSON(comments)
			}

			if len(comments) == 0 {
				p.Warn("No comments")
				return nil
			}

			p.Header(fmt.Sprintf("Comments (%d)", len(comments)))
			for _, c := range comments {
				author := c.AuthorProfile.FirstName + " " + c.AuthorProfile.LastName
				p.Printf("  %s  [%s]  👍 %d\n    %s\n\n", author, c.PostedAt, c.LikeCount, c.Body)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&start, "start", 0, "Pagination offset")
	cmd.Flags().IntVar(&count, "count", 20, "Number of comments")
	return cmd
}
