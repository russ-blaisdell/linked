# LinkedIn CLI Skill

## Overview

This skill enables OpenClaw to interact with LinkedIn using the `linked` command-line tool. Use it when the user wants to search for jobs, manage messages, update their profile, handle connections, manage recommendations, engage with posts, or perform any other LinkedIn action.

## Prerequisites

- `linked` must be installed and in PATH
- Credentials must be configured: run `linked auth setup` once to store session cookies

## When to Use

Trigger this skill when the user asks about or wants to:
- Search for jobs, people, or companies on LinkedIn
- Read, send, or manage LinkedIn messages
- View or update their LinkedIn profile
- Manage connection requests (send, accept, ignore, withdraw)
- Follow or unfollow people or companies
- View, request, hide, or show recommendations
- Create posts, like, comment, or share content
- View their LinkedIn feed or notifications
- Get information about a LinkedIn company

## Commands

### Authentication
```
linked auth setup          # Configure credentials (one-time setup)
linked auth whoami         # Verify auth and show current account
linked auth list           # List configured profiles
```

### Profile
```
linked profile get                          # Get own profile
linked profile get --urn <profile-id>       # Get another member's profile
linked profile update --headline "..."      # Update headline
linked profile update --summary "..."       # Update about/summary
linked profile update --location "..."      # Update location
linked profile skills                       # List profile skills
linked profile contact                      # Get contact info
```

### Search
```
linked search people "<keywords>"           # Search LinkedIn members
linked search people "<keywords>" --company <name> --network FIRST,SECOND
linked search jobs "<keywords>"             # Search job postings
linked search jobs "<keywords>" --remote --location "New York"
linked search companies "<keywords>"        # Search companies
linked search posts "<keywords>"            # Search posts
```

### Messages
```
linked messages list                        # All conversations
linked messages unread                      # Unread conversations only
linked messages read <conversation-id>      # Read a thread
linked messages send "<text>" --conversation <id>   # Reply
linked messages send "<text>" --to <urn>            # New conversation
```

### Connections
```
linked connections list                     # 1st-degree connections
linked connections pending                  # Pending received invitations
linked connections sent                     # Sent invitations
linked connections request <urn>            # Send connection request
linked connections request <urn> --note "..." # With personal note
linked connections accept <invitation-id>   # Accept invitation
linked connections ignore <invitation-id>   # Ignore invitation
linked connections withdraw <invitation-urn> # Withdraw sent request
linked connections follow <urn>             # Follow a member
linked connections unfollow <urn>           # Unfollow a member
```

### Jobs
```
linked jobs get <job-id>                    # Job posting details
linked jobs saved                           # Saved jobs
linked jobs save <job-id>                   # Save a job
linked jobs unsave <job-id>                 # Remove from saved
linked jobs applied                         # Applied jobs
```

### Companies
```
linked companies get <company-id>           # Company info (use universal name e.g. "anthropic")
linked companies follow <company-urn>       # Follow company
linked companies unfollow <company-urn>     # Unfollow company
linked companies posts <company-urn>        # Recent company posts
```

### Posts & Feed
```
linked posts feed                           # Home feed
linked posts create "<text>"                # Create a post
linked posts like <post-urn>                # Like a post
linked posts unlike <post-urn>              # Remove like
linked posts comment <post-urn> "<text>"    # Comment on post
linked posts share <post-urn>               # Reshare
linked posts comments <post-urn>            # Get comments
```

### Recommendations
```
linked recommendations received             # Recommendations on your profile
linked recommendations given                # Recommendations you wrote
linked recommendations request <urn>        # Request a recommendation
linked recommendations request <urn> --relationship COLLEAGUE --message "..."
linked recommendations hide <urn>           # Hide from profile
linked recommendations show <urn>           # Show on profile
```

### Notifications
```
linked notifications list                   # Recent notifications
linked notifications mark-read <urn>        # Mark as read
```

## Output Formats

All commands support `--output` / `-o`:
- `pretty` (default) — human-readable with colour
- `json` — machine-readable JSON (best for parsing results)
- `table` — tabular format

For structured data in agent workflows, always use `-o json`.

## Multiple Profiles

Use `--profile <name>` to switch between LinkedIn accounts:
```
linked --profile work messages list
linked --profile personal profile get
```

## Example Agent Workflows

**Find relevant job openings:**
```
linked search jobs "senior golang engineer" --remote -o json
linked jobs get <job-id> -o json
```

**Respond to messages:**
```
linked messages unread -o json
linked messages read <conversation-id> -o json
linked messages send "Thanks for reaching out!" --conversation <id>
```

**Request a recommendation:**
```
linked connections list -o json   # find the connection's URN
linked recommendations request <urn> --relationship COLLEAGUE --message "Hi, would you be willing to write me a recommendation based on our work together at Acme?"
```
