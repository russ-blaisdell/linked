# LinkedIn CLI Skill

## Overview

This skill enables OpenClaw to interact with LinkedIn using the `linked` command-line tool. Use it when the user wants to search for jobs, manage messages, update their profile, handle connections, manage recommendations, engage with posts, or perform any other LinkedIn action.

## Prerequisites

- `linked` must be installed and in PATH
- Credentials must be configured: run `linked auth setup` once to store session cookies

## When to Use

Trigger this skill when the user asks about or wants to:
- Search for jobs, people, or companies on LinkedIn
- Read, send, star, archive, or delete LinkedIn messages
- View or update their LinkedIn profile (headline, summary, experience, education, skills, certifications, languages, volunteer work, projects, publications, honors, courses)
- Set or clear Open to Work status
- See who viewed their profile
- Upload a profile photo
- Manage connection requests (send, accept, ignore, withdraw, remove)
- Follow or unfollow people or companies
- View mutual connections with someone
- View, request, hide, show, decline, or delete recommendations
- Create, edit, delete, or like posts; post with an image; comment, share
- View their LinkedIn feed or a member's recent activity
- View or dismiss LinkedIn notifications
- Get information about a LinkedIn company or its employees

## Commands

### Authentication
```
linked auth setup          # Configure credentials (one-time setup)
linked auth whoami         # Verify auth and show current account
linked auth list           # List configured profiles
```

### Profile — View and Update
```
linked profile get                          # Get own profile
linked profile get --urn <profile-id>       # Get another member's profile
linked profile update --headline "..."      # Update headline
linked profile update --summary "..."       # Update about/summary
linked profile update --location "..."      # Update location
linked profile skills                       # List profile skills (top-level shortcut)
linked profile contact                      # Get contact info
linked profile who-viewed                   # See who viewed your profile recently
linked profile photo <path>                 # Upload a new profile photo
```

### Profile — Experience
```
linked profile experience list
linked profile experience add    --title "..." --company "..." [--current] [--start-year 2022] [--description "..."]
linked profile experience update <id> [flags]
linked profile experience remove <id>
```

### Profile — Education
```
linked profile education list
linked profile education add    --school "..." --degree "..." --field "..." [--start-year 2018] [--end-year 2022]
linked profile education update <id> [flags]
linked profile education remove <id>
```

### Profile — Skills
```
linked profile skills list
linked profile skills add <skill-name>
linked profile skills remove <id>
```

### Profile — Certifications
```
linked profile certifications list
linked profile certifications add    --name "..." --authority "..." [--license "..."] [--url "..."]
linked profile certifications update <id> [flags]
linked profile certifications remove <id>
```

### Profile — Languages
```
linked profile languages list
linked profile languages add    --name "..." --proficiency PROFESSIONAL_WORKING
linked profile languages remove <id>
```

### Profile — Volunteer
```
linked profile volunteer list
linked profile volunteer add    --role "..." --org "..." [--cause "..."] [--current]
linked profile volunteer update <id> [flags]
linked profile volunteer remove <id>
```

### Profile — Projects
```
linked profile projects list
linked profile projects add    --title "..." [--description "..."] [--url "..."] [--current]
linked profile projects update <id> [flags]
linked profile projects remove <id>
```

### Profile — Publications
```
linked profile publications list
linked profile publications add    --name "..." --publisher "..." [--url "..."]
linked profile publications update <id> [flags]
linked profile publications remove <id>
```

### Profile — Honors
```
linked profile honors list
linked profile honors add    --title "..." --issuer "..." [--year 2023]
linked profile honors remove <id>
```

### Profile — Courses
```
linked profile courses list
linked profile courses add    --name "..." --number "..." [--occupation "..."]
linked profile courses remove <id>
```

### Profile — Open to Work
```
linked profile open-to-work   [--title "..."] [--job-types FULL_TIME,CONTRACT]
linked profile close-to-work
```

### Search
```
linked search people "<keywords>"           # Search LinkedIn members
linked search people "<keywords>" --company <name> --network FIRST,SECOND
linked search people "<keywords>" --title "Engineer" --location "NYC"
linked search jobs "<keywords>"             # Search job postings
linked search jobs "<keywords>" --remote --location "New York"
linked search jobs "<keywords>" --experience-level MID_SENIOR_LEVEL
linked search companies "<keywords>"        # Search companies
linked search posts "<keywords>"            # Search posts and content
```

### Messages
```
linked messages list                              # All conversations
linked messages unread                            # Unread conversations only
linked messages read <conversation-id>            # Read a thread
linked messages send "<text>" --conversation <id> # Reply to existing
linked messages send "<text>" --to <urn>          # Start new conversation
linked messages mark-read <conversation-id>       # Mark as read
linked messages star <conversation-id>            # Star (bookmark)
linked messages unstar <conversation-id>          # Remove star
linked messages archive <conversation-id>         # Archive
linked messages unarchive <conversation-id>       # Restore from archive
linked messages delete <conversation-id> <message-urn>  # Delete a message
linked messages delete-conversation <conversation-id>   # Delete entire conversation
```

### Connections
```
linked connections list                      # 1st-degree connections
linked connections pending                   # Pending received invitations
linked connections sent                      # Sent invitations
linked connections request <urn>             # Send connection request
linked connections request <urn> --note "..." # With personal note
linked connections accept <invitation-id>    # Accept invitation
linked connections ignore <invitation-id>    # Ignore invitation
linked connections withdraw <invitation-urn> # Withdraw sent request
linked connections remove <member-urn>       # Remove a connection
linked connections follow <urn>              # Follow a member
linked connections unfollow <urn>            # Unfollow a member
linked connections mutual <urn>              # Mutual connections with a member
```

### Jobs
```
linked jobs get <job-id>                     # Job posting details
linked jobs saved                            # Saved jobs
linked jobs save <job-id>                    # Save a job
linked jobs unsave <job-id>                  # Remove from saved
linked jobs applied                          # Applied jobs
linked jobs recommended                      # LinkedIn-recommended jobs for you
linked jobs company <company-urn>            # Jobs at a specific company
```

### Companies
```
linked companies get <company-id>            # Company info (use universal name e.g. "anthropic")
linked companies follow <company-urn>        # Follow company
linked companies unfollow <company-urn>      # Unfollow company
linked companies posts <company-urn>         # Recent company posts
linked companies employees <company-id>      # Members who work there
```

### Posts & Feed
```
linked posts feed                            # Home feed
linked posts create "<text>"                 # Create a text post
linked posts create-with-image "<text>" --image <path>  # Post with image
linked posts get <post-urn>                  # Get a post
linked posts edit <post-urn> "<new-text>"    # Edit a post
linked posts delete <post-urn>               # Delete a post
linked posts like <post-urn>                 # Like a post
linked posts unlike <post-urn>               # Remove like
linked posts comment <post-urn> "<text>"     # Comment on post
linked posts share <post-urn>                # Reshare
linked posts comments <post-urn>             # Get comments on a post
linked posts delete-comment <post-urn> <comment-urn>   # Delete a comment
linked posts like-comment <post-urn> <comment-urn>     # Like a comment
linked posts activity <profile-id>           # Member's recent posts
```

### Recommendations
```
linked recommendations received              # Recommendations on your profile
linked recommendations given                 # Recommendations you wrote
linked recommendations request <urn>         # Request a recommendation
linked recommendations request <urn> --relationship COLLEAGUE --message "..."
linked recommendations hide <urn>            # Hide from profile
linked recommendations show <urn>            # Show on profile
linked recommendations decline <urn>         # Decline a pending request
linked recommendations delete <urn>          # Delete a recommendation you gave
```

### Notifications
```
linked notifications list                    # Recent notifications
linked notifications mark-read <urn>         # Mark one as read
linked notifications mark-all-read           # Mark all as read
linked notifications count                   # Unread badge count
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

**Find and evaluate job openings:**
```
linked jobs recommended -o json
linked search jobs "senior golang engineer" --remote -o json
linked jobs get <job-id> -o json
```

**Respond to messages:**
```
linked messages unread -o json
linked messages read <conversation-id> -o json
linked messages send "Thanks for reaching out!" --conversation <id>
```

**Manage your profile:**
```
linked profile get -o json
linked profile update --headline "Staff Engineer at Acme"
linked profile experience add --title "Staff Engineer" --company "Acme" --current
linked profile who-viewed -o json
```

**Handle connection requests:**
```
linked connections pending -o json
linked connections accept <invitation-id>
linked connections mutual <member-urn> -o json
```

**Engage with content:**
```
linked posts feed -o json
linked posts like <post-urn>
linked posts comment <post-urn> "Great insights!"
linked posts activity <profile-id> -o json
```

**Request a recommendation:**
```
linked connections list -o json   # find the connection's URN
linked recommendations request <urn> --relationship COLLEAGUE \
  --message "Hi, would you be willing to write me a recommendation based on our work together at Acme?"
```

**Check notification status:**
```
linked notifications count -o json
linked notifications list -o json
linked notifications mark-all-read
```

---

## Profile Coach Workflow

Trigger this workflow when the user asks to improve, refresh, rewrite, or review their LinkedIn profile. Examples:

- "Help me improve my LinkedIn profile"
- "My LinkedIn is outdated, can you fix it?"
- "Rewrite my experience section"
- "I'm job hunting — make my profile stronger"
- "Do a LinkedIn profile review"

### Step 1 — Gather the full profile

Fetch all sections in parallel. Do not skip any — gaps in the data lead to weak questions.

```
linked profile get -o json
linked profile experience list -o json
linked profile education list -o json
linked profile skills list -o json
linked profile certifications list -o json
linked profile languages list -o json
linked profile volunteer list -o json
linked profile projects list -o json
linked profile publications list -o json
linked profile honors list -o json
linked profile courses list -o json
linked recommendations received -o json
```

### Step 2 — Analyze before asking anything

Before speaking to the user, silently assess the profile for:

**Headline**
- Is it a job title only, or does it communicate value? ("Software Engineer" vs "Software Engineer helping fintech teams ship faster")
- Does it reflect their current goals (job seeker, open to work, consultant, etc.)?
- Is it under 220 characters?

**About / Summary**
- Does one exist? (Many profiles have none — this is a major gap)
- Does it open with a hook, or does it start with "I am a..."?
- Does it explain what they do, who they help, and what makes them distinct?
- Does it include a call to action?

**Experience**
- Are descriptions present, or just job titles with no bullet points?
- Are descriptions written with accomplishments ("Led migration that reduced latency by 40%") or just duties ("Responsible for backend services")?
- Are dates correct and complete? Are there gaps that might need explanation?
- Is the most recent role clearly the most detailed?

**Skills**
- Are the top skills actually the most relevant to their goals?
- Are there obvious skills missing given their experience?

**Education**
- Is it present and complete?

**Other sections**
- Are any sections empty that could strengthen the profile (certifications, volunteer, publications, projects)?
- Are recommendations present? How many?

### Step 3 — Conduct the interview

Present your findings as a brief assessment first ("Your experience descriptions are strong, but your headline reads like a job title and your summary is missing — those are the two highest-impact things to fix"), then ask targeted questions. Do not ask everything at once — group questions by section and wait for answers before moving on.

**Headline questions (ask if weak or missing):**
- What do you want people to do after reading your profile — hire you, partner with you, follow you?
- Are you currently job searching, or building visibility for your current role?
- What are the 2-3 things you're most known for professionally?
- What industries or types of companies do you most want to attract?

**Summary / About questions (ask if missing or generic):**
- How would you describe what you do to someone outside your industry?
- What problem do you solve for the people or companies you work with?
- What's something about your background or approach that makes you different?
- What do you want someone to do after reading your profile? (reach out, hire you, follow your content?)

**Experience questions (ask per role that has weak or missing descriptions):**
- What were the 2-3 most important things you accomplished in this role?
- Did you lead or grow a team? Build something from scratch? Fix something broken?
- Are there numbers you can attach to your impact — team size, revenue, cost savings, performance improvements, users served?
- What would your manager or teammates say was your biggest contribution there?

**Skills / positioning questions (ask if skills seem misaligned or sparse):**
- What skills do you most want to be known for right now?
- Are there tools, languages, or methodologies you use daily that aren't listed?
- What are you trying to move toward — a new role, new industry, or a more senior level?

**Goals question (ask once, early, if not clear from context):**
- What's the goal for updating your profile? Passive job search, active job search, thought leadership, business development, something else?

### Step 4 — Write the improvements

Use the answers to write improved content. Apply these standards:

**Headline format:**
`[What you do] + [who you do it for or what outcome you create]`
- Good: "Staff Engineer | Helping distributed teams build reliable backend systems"
- Good: "Product Manager | 0→1 products in fintech and healthtech"
- Avoid: "Software Engineer at Acme Corp" (just a title)
- Max 220 characters. No buzzwords like "passionate" or "results-driven".

**Summary format (3–4 short paragraphs):**
1. Opening hook — one punchy sentence about what you do and for whom
2. What you bring — 2-3 specific strengths, approaches, or areas of expertise
3. Background highlights — a sentence or two on notable experience or accomplishments
4. Call to action — "Open to senior IC roles in infrastructure" or "Feel free to reach out if..."
- Write in first person. Keep it conversational. Avoid corporate jargon.
- Target 150–300 words.

**Experience descriptions:**
- Lead with the scope or context if helpful ("In a team of 8 engineers, I...")
- Use action verbs: Built, Led, Designed, Reduced, Increased, Launched, Migrated, etc.
- Follow the format: Action + What + Result/Impact whenever possible
  - "Redesigned the data ingestion pipeline, reducing processing time from 4 hours to 12 minutes"
  - "Led a team of 5 engineers to migrate a monolith to microservices, enabling 3x faster feature delivery"
- Include 3–5 bullet points or a short paragraph per role. Most recent role should be the richest.
- Quantify wherever possible — numbers make accomplishments concrete and searchable.

### Step 5 — Confirm and apply

Show the user each rewritten section before applying it. Give them the option to tweak wording before committing. Once confirmed, apply updates using the CLI:

```
# Update headline and summary
linked profile update --headline "<new headline>" --summary "<new summary>"

# Update an experience description (use the ID from experience list)
linked profile experience update <id> --description "<rewritten bullets>"

# Add a missing skill
linked profile skills add "<skill name>"
```

Apply one section at a time and confirm success before moving to the next. After all updates are applied, run `linked profile get -o json` to verify the final state.

### Step 6 — Recommend next steps

After the profile is updated, suggest what would make it even stronger:

- If recommendations are sparse: offer to draft recommendation request messages to 2-3 relevant connections
- If no profile photo exists: remind them that profiles with photos get significantly more views
- If Open to Work is relevant: offer to set it with appropriate job types and locations
- If the summary mentions specific skills: check that those skills are listed in the Skills section
- Suggest they share a post announcing their updated profile or new direction, if appropriate
