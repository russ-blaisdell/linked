# linked

[![CI](https://github.com/russ-blaisdell/linked/actions/workflows/ci.yml/badge.svg)](https://github.com/russ-blaisdell/linked/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/russ-blaisdell/linked)](https://goreportcard.com/report/github.com/russ-blaisdell/linked)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue)](https://go.dev/doc/go1.24)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

> A full-featured LinkedIn CLI written in Go — and an AI-powered agent skill for [OpenClaw](https://www.getopenclaw.ai/).

`linked` gives you complete programmatic control over LinkedIn from the terminal: manage your profile, messages, connections, jobs, posts, and more. Because it uses LinkedIn's internal Voyager API (cookie-based auth), **no developer account or API key is required** — just your existing LinkedIn session.

When paired with OpenClaw, `linked` becomes an intelligent agent skill that can conduct full conversations with you: reviewing your profile, interviewing you about your experience, and rewriting everything to be stronger.

---

## Table of Contents

- [What it can do](#what-it-can-do)
- [Agent Skills](#agent-skills)
  - [Profile Coach](#profile-coach)
  - [Job Search Assistant](#job-search-assistant)
  - [Message Manager](#message-manager)
- [Installation](#installation)
- [Authentication](#authentication)
- [Quick Start](#quick-start)
- [Output Formats](#output-formats)
- [Full Command Reference](#full-command-reference)
  - [Auth](#auth)
  - [Profile](#profile)
  - [Search](#search)
  - [Messages](#messages)
  - [Connections](#connections)
  - [Jobs](#jobs)
  - [Companies](#companies)
  - [Posts & Feed](#posts--feed)
  - [Recommendations](#recommendations)
  - [Notifications](#notifications)
- [OpenClaw Integration](#openclaw-integration)
- [Architecture](#architecture)
- [Development](#development)
- [Testing](#testing)
- [Important Notes](#important-notes)

---

## What it can do

| Domain | Capabilities |
|--------|-------------|
| **Profile** | View any profile; update headline, summary, location; full CRUD for experience, education, skills, certifications, languages, volunteer work, projects, publications, honors, and courses; open-to-work toggle; see who viewed your profile; upload a photo |
| **Search** | Search people with company/title/school/network/location filters; search jobs with remote/location/experience/employment-type filters; search companies and posts |
| **Messages** | List and read conversations; send new messages or reply; star, archive, mark-read; delete individual messages or entire conversations |
| **Connections** | List connections; send, accept, ignore, withdraw, and remove invitations; follow/unfollow members; see mutual connections |
| **Jobs** | View job details; save/unsave; list saved, applied, and recommended jobs; search jobs at a specific company |
| **Companies** | Company info; follow/unfollow; recent company posts; employee search |
| **Posts & Feed** | Home feed; create text or image posts; edit and delete posts; like, comment, share; view, delete, and like comments; see a member's recent activity |
| **Recommendations** | View received and given; request from connections; hide, show, decline, and delete |
| **Notifications** | List notifications; mark one or all as read; unread badge count |

---

## Agent Skills

When installed as an OpenClaw skill, `linked` enables your AI agent to handle complex LinkedIn workflows through natural conversation. See [skill/linkedin/skill.md](skill/linkedin/skill.md) for the full skill definition.

### Profile Coach

The most powerful workflow in the skill. Your agent will:

1. **Fetch your entire profile** — all sections in parallel: experience, education, skills, certifications, languages, volunteer work, projects, publications, honors, courses, and received recommendations
2. **Silently audit it** — identifying gaps and weaknesses before asking a single question (missing summary, headline that's just a job title, experience descriptions that list duties instead of accomplishments, misaligned skills, etc.)
3. **Interview you** — ask targeted, section-by-section questions to understand your accomplishments, impact, and goals; skip questions it can already answer from the profile data
4. **Write the improvements** — rewrite your headline, summary, and experience descriptions using professional standards: action verbs, quantified impact, the right length and format for each section
5. **Confirm before applying** — show you each rewritten section for approval before committing any changes
6. **Suggest next steps** — recommend requesting recommendations, adding a photo, setting Open to Work, or aligning skills with your updated summary

**Trigger it with:**
```
@openclaw review my LinkedIn profile and help me improve it
@openclaw my LinkedIn is outdated — can you rewrite it?
@openclaw I'm job hunting, make my LinkedIn profile stronger
@openclaw do a full LinkedIn profile audit
@openclaw rewrite my experience section on LinkedIn
```

### Job Search Assistant

Your agent can search for jobs, pull full descriptions, compare against your profile, and help you decide where to apply.

```
@openclaw find remote senior Go engineer jobs
@openclaw what jobs does Anthropic have open on LinkedIn?
@openclaw show me my LinkedIn job recommendations
@openclaw find full-time product manager roles in New York
```

### Message Manager

Stay on top of your inbox without opening LinkedIn.

```
@openclaw show my unread LinkedIn messages
@openclaw read my conversation with Jane Smith
@openclaw send a reply to the recruiter from Acme: "Thanks, I'm interested!"
@openclaw archive all the old LinkedIn conversations
@openclaw who has sent me connection requests this week?
```

---

## Installation

### Prerequisites

- Go 1.24+
- A LinkedIn account with an active browser session

### Build from source

```bash
git clone https://github.com/russ-blaisdell/linked
cd linked
make install   # builds and installs to /usr/local/bin/linked
```

Or build without installing:

```bash
make build     # produces dist/linked
```

---

## Authentication

`linked` authenticates using your LinkedIn session cookies — the same cookies your browser sends on every request. No LinkedIn developer account, OAuth app, or API key needed.

You need two cookies from your logged-in browser session:

| Cookie | What it looks like | Purpose |
|--------|--------------------|---------|
| `li_at` | `AQEDARxxxxxxx...` | Your session token |
| `JSESSIONID` | `ajax:1234567890abcdef` | CSRF validation |

### Setup

```bash
linked auth setup
```

You will be prompted for each value. See **[docs/auth.md](docs/auth.md)** for step-by-step cookie extraction instructions for Chrome, Firefox, and Safari.

### Verify

```bash
linked auth whoami
```

### Multiple accounts

```bash
linked auth setup --profile work
linked auth setup --profile personal

linked --profile work messages list
linked --profile personal profile get
```

### Cookie expiry

`li_at` cookies typically last one year. If you get a `401 Unauthorized` error, your session has expired — log out and back in to LinkedIn, then re-run `linked auth setup`.

---

## Quick Start

```bash
# Confirm auth is working
linked auth whoami

# View your profile
linked profile get

# Search for remote Go jobs
linked search jobs "golang engineer" --remote -o table

# See LinkedIn's recommended jobs for you
linked jobs recommended

# Check unread messages
linked messages unread

# Reply to a conversation
linked messages send "Thanks for reaching out!" --conversation 2-abc123

# Find who viewed your profile
linked profile who-viewed

# Update your headline
linked profile update --headline "Staff Engineer | Building reliable distributed systems"

# Add a new position
linked profile experience add --title "Staff Engineer" --company "Acme Corp" --current

# Search for people at a company
linked search people "product manager" --company google --network FIRST,SECOND -o table

# Send a connection request with a note
linked connections request urn:li:member:12345678 --note "Hi, we met at GopherCon!"

# Create a post
linked posts create "Excited to share my latest project!"

# Create a post with an image
linked posts create-with-image "Check this out" --image ./screenshot.png

# Request a recommendation
linked recommendations request urn:li:member:12345678 \
  --relationship COLLEAGUE \
  --message "Would you be willing to write me a recommendation?"

# Check unread notification count
linked notifications count

# Mark all notifications read
linked notifications mark-all-read
```

---

## Output Formats

Every command supports `--output` / `-o` with three formats:

| Format | Flag | Best for |
|--------|------|----------|
| Pretty (default) | `-o pretty` | Reading in terminal, colored output |
| JSON | `-o json` | Piping to other tools, agent workflows, scripting |
| Table | `-o table` | Scanning lists of results at a glance |

```bash
linked search jobs "engineer" --remote -o json | jq '.[].title'
linked connections list -o table
linked profile get -o json
```

When used as an OpenClaw agent skill, all commands run with `-o json` so the agent can parse and reason about the results.

---

## Full Command Reference

### Auth

```
linked auth setup   [--profile <name>]   # configure credentials interactively
linked auth whoami  [--profile <name>]   # verify auth and show current account
linked auth remove  [--profile <name>]   # delete stored credentials
linked auth list                         # list all configured profiles
```

> See [docs/auth.md](docs/auth.md) for detailed setup instructions.

---

### Profile

**View and update top-level fields:**
```
linked profile get      [--urn <member-urn>]   # your profile, or anyone else's
linked profile update   --headline "..."        # update headline
                        --summary "..."         # update about/summary
                        --location "..."        # update location
linked profile contact  [--urn <member-urn>]   # contact info (email, phone, etc.)
linked profile who-viewed                       # recent profile viewers
linked profile photo    <path>                  # upload a new profile photo
```

**Experience:**
```
linked profile experience list
linked profile experience add    --title "..." --company "..." [--current] [--description "..."]
linked profile experience update <id> [flags]
linked profile experience remove <id>
```

**Education:**
```
linked profile education list
linked profile education add    --school "..." --degree "..." [--field "..."] [--end-year 2022]
linked profile education update <id> [flags]
linked profile education remove <id>
```

**Skills:**
```
linked profile skills list
linked profile skills add    <name>
linked profile skills remove <id>
```

**Certifications:**
```
linked profile certifications list
linked profile certifications add    --name "..." --authority "..." [--license "..."] [--url "..."]
linked profile certifications update <id> [flags]
linked profile certifications remove <id>
```

**Languages:**
```
linked profile languages list
linked profile languages add    --name "..." --proficiency PROFESSIONAL_WORKING
linked profile languages remove <id>
```

**Volunteer experience:**
```
linked profile volunteer list
linked profile volunteer add    --role "..." --org "..." [--cause "..."] [--current]
linked profile volunteer update <id> [flags]
linked profile volunteer remove <id>
```

**Projects:**
```
linked profile projects list
linked profile projects add    --title "..." [--description "..."] [--url "..."] [--current]
linked profile projects update <id> [flags]
linked profile projects remove <id>
```

**Publications:**
```
linked profile publications list
linked profile publications add    --name "..." --publisher "..." [--url "..."]
linked profile publications update <id> [flags]
linked profile publications remove <id>
```

**Honors and awards:**
```
linked profile honors list
linked profile honors add    --title "..." --issuer "..." [--year 2023]
linked profile honors remove <id>
```

**Courses:**
```
linked profile courses list
linked profile courses add    --name "..." [--number "..."]
linked profile courses remove <id>
```

**Open to Work:**
```
linked profile open-to-work  [--title "..."] [--job-types FULL_TIME,CONTRACT]
linked profile close-to-work
```

---

### Search

```
linked search people "<keywords>"
  --company <name>                   # filter by current company
  --title   <title>                  # filter by job title
  --school  <school>                 # filter by school
  --network FIRST,SECOND,THIRD       # filter by connection distance
  --location "<location>"
  --count <n>  --start <n>

linked search jobs "<keywords>"
  --location "<location>"
  --remote                           # remote jobs only
  --company <name>
  --experience-level <level>         # ENTRY_LEVEL, MID_SENIOR_LEVEL, DIRECTOR, EXECUTIVE
  --employment-type  <type>          # FULL_TIME, PART_TIME, CONTRACT, INTERNSHIP
  --count <n>  --start <n>

linked search companies "<keywords>"   [--count <n>]
linked search posts    "<keywords>"    [--count <n>]
```

---

### Messages

```
linked messages list                [--count <n>]   # all conversations
linked messages unread              [--count <n>]   # unread only
linked messages read   <conv-id>                    # read a thread
linked messages send   "<text>"
  --conversation <id>       # reply to existing conversation
  --to <member-urn>         # start a new conversation

linked messages mark-read          <conv-id>        # mark as read
linked messages star               <conv-id>        # bookmark a conversation
linked messages unstar             <conv-id>
linked messages archive            <conv-id>
linked messages unarchive          <conv-id>
linked messages delete             <conv-id> <message-urn>
linked messages delete-conversation <conv-id>
```

---

### Connections

```
linked connections list     [--count <n>]             # 1st-degree connections
linked connections pending  [--count <n>]             # received invitations
linked connections sent     [--count <n>]             # sent invitations

linked connections request  <member-urn> [--note "..."]
linked connections accept   <invitation-id>
linked connections ignore   <invitation-id>
linked connections withdraw <invitation-urn>
linked connections remove   <member-urn>              # disconnect from someone

linked connections follow   <member-urn>
linked connections unfollow <member-urn>
linked connections mutual   <member-urn> [--count <n>]   # shared connections
```

---

### Jobs

```
linked jobs get         <job-id>          # full job posting details
linked jobs recommended [--count <n>]     # LinkedIn's recommendations for you
linked jobs saved       [--count <n>]     # your saved jobs
linked jobs save        <job-id>
linked jobs unsave      <job-id>
linked jobs applied     [--count <n>]     # jobs you've applied to
linked jobs company     <company-urn>     # all open roles at a company
```

---

### Companies

```
linked companies get       <universal-name>   # info by slug (e.g. "anthropic", "google")
linked companies follow    <company-urn>
linked companies unfollow  <company-urn>
linked companies posts     <company-urn>      # recent company posts
linked companies employees <company-id>       # members who work there
```

---

### Posts & Feed

```
linked posts feed                                        # your home feed
linked posts create "<text>"  [--visibility PUBLIC|CONNECTIONS]
linked posts create-with-image "<text>" --image <path>  # post with attached image
linked posts get    <post-urn>
linked posts edit   <post-urn> "<new-text>"
linked posts delete <post-urn>

linked posts like    <post-urn>
linked posts unlike  <post-urn>
linked posts comment <post-urn> "<text>"
linked posts share   <post-urn> [--commentary "..."]

linked posts comments       <post-urn> [--count <n>]    # view all comments
linked posts delete-comment <post-urn> <comment-urn>
linked posts like-comment   <post-urn> <comment-urn>
linked posts activity       <profile-id> [--count <n>]  # a member's recent posts
```

---

### Recommendations

```
linked recommendations received                 # on your profile
linked recommendations given                    # written by you
linked recommendations request <member-urn>
  --relationship COLLEAGUE|MANAGER|REPORT|CLASSMATE
  --message "..."
linked recommendations hide    <urn>            # hide from your profile
linked recommendations show    <urn>            # make visible again
linked recommendations decline <urn>            # decline a pending request
linked recommendations delete  <urn>            # delete one you wrote
```

---

### Notifications

```
linked notifications list                          # recent notifications
linked notifications mark-read <notification-urn>  # mark one as read
linked notifications mark-all-read                 # clear the entire badge
linked notifications count                         # unread count only
```

---

## OpenClaw Integration

`linked` ships with a complete OpenClaw skill definition at [`skill/linkedin/skill.md`](skill/linkedin/skill.md). Once installed, your OpenClaw agent understands the full capability of the CLI and can drive it through natural conversation — including the multi-step [Profile Coach](#profile-coach) workflow.

### Install

```bash
make skill            # copies skill.md to ~/.openclaw/workspace/skills/linkedin/
openclaw gateway      # restart to load the new skill
```

### Verify

```bash
openclaw skills list | grep linkedin
```

### Try it

```
@openclaw help me improve my LinkedIn profile
@openclaw find remote Go engineer jobs on LinkedIn
@openclaw show my unread LinkedIn messages
@openclaw who sent me connection requests this week?
@openclaw what are my recommended jobs on LinkedIn?
@openclaw update my LinkedIn headline to "Staff Engineer | Distributed Systems"
@openclaw who viewed my LinkedIn profile recently?
@openclaw how many unread LinkedIn notifications do I have?
@openclaw request a recommendation from my colleague Sarah at Acme
```

> See **[docs/openclaw-setup.md](docs/openclaw-setup.md)** for full setup instructions, troubleshooting, and multi-profile configuration.

---

## Architecture

```
cmd/linked/          Entry point — calls cli.Execute()
cli/                 Cobra commands, one file per domain
internal/
  api/               LinkedIn Voyager API service layer
  client/            HTTP client (cookie jar, CSRF headers, error handling)
  config/            Credential storage (~/.openclaw/credentials/linkedin/)
  models/            Go structs for all LinkedIn data and input types
  output/            pretty / JSON / table renderers
mock/                In-process mock Voyager server for testing
  fixtures/          Realistic JSON response fixtures
tests/integration/   Integration test suite (zero network calls)
skill/linkedin/      OpenClaw skill definition
docs/                Auth and setup guides
```

The HTTP client automatically injects all headers required by the Voyager API on every request: `csrf-token`, `x-restli-protocol-version`, `x-li-lang`, `x-li-track`, `Accept`, and `User-Agent`.

Profile photo and image post uploads use LinkedIn's three-step media flow: register upload → PUT binary to the upload URL → associate the asset URN in the create request.

---

## Development

```bash
make build      # compile binary → dist/linked
make test       # run integration tests (verbose)
make lint       # go vet ./...
make release    # cross-compile: darwin/arm64, darwin/amd64, linux/amd64
make skill      # install skill to ~/.openclaw/workspace/skills/linkedin/
make clean      # remove dist/
```

### Adding a feature

1. Add the API method(s) to `internal/api/`
2. Add structs to `internal/models/` if needed
3. Add a mock handler in `mock/server.go` and a fixture if needed
4. Add the Cobra command in `cli/` and register it in the subcommand list
5. Add integration tests in `tests/integration/`
6. Update [`skill/linkedin/skill.md`](skill/linkedin/skill.md) with the new command

> See **[CLAUDE.md](CLAUDE.md)** for detailed architecture notes, design decisions, and the complete file-by-file breakdown.

---

## Testing

[![CI](https://github.com/russ-blaisdell/linked/actions/workflows/ci.yml/badge.svg)](https://github.com/russ-blaisdell/linked/actions/workflows/ci.yml)

All tests run entirely against an in-process mock Voyager server — no network access and no real LinkedIn credentials required.

```bash
make test        # run integration tests (verbose)
make test-short  # run tests without verbose output
```

Expected output:

```
ok   github.com/russ-blaisdell/linked/tests/integration
```

### How the mock server works

`mock.New()` starts an `httptest.Server` that replicates the Voyager API surface:

- **CSRF validation** — every request must include a matching `csrf-token` header; requests without it fail with `403`, exactly as LinkedIn does
- **JSON fixtures** — responses are served from `mock/fixtures/`, so tests assert against realistic data shapes
- **Stateful mutations** — the server tracks side effects in memory: messages sent, posts created, jobs saved, likes applied, follows, and more; tests can assert that operations actually changed server state

### Test coverage

Integration tests cover all 10 command domains:

| File | Domain |
|------|--------|
| `auth_test.go` | Credential loading and `whoami` |
| `profile_test.go` | Get/update profile; experience, education, skills, certifications, languages, volunteer, projects, publications, honors, courses; open-to-work; who viewed; photo upload |
| `search_test.go` | People, jobs, companies, posts |
| `messaging_test.go` | List, read, send, star, archive, mark-read, delete |
| `connections_test.go` | List, pending/sent invitations, request, accept, ignore, withdraw, remove, follow, unfollow, mutual |
| `jobs_test.go` | Get, save, unsave, saved, applied, recommended, company jobs |
| `companies_test.go` | Get, follow, unfollow, posts, employees |
| `posts_test.go` | Feed, create, create-with-image, edit, delete, like, unlike, comment, share, comments, delete-comment, like-comment, activity |
| `recommendations_test.go` | Received, given, request, hide, show, decline, delete |
| `notifications_test.go` | List, mark-read, mark-all-read, count |

### CI

The [CI workflow](.github/workflows/ci.yml) runs on every push and pull request to `main`:

1. `go vet ./...` — static analysis
2. `go test ./tests/integration/... -v -count=1` — full integration suite
3. Cross-platform build check: `darwin/arm64`, `darwin/amd64`, `linux/amd64`

---

## Important Notes

- **Terms of Service** — This tool uses LinkedIn's internal Voyager API, the same API the LinkedIn website uses. It is not an officially supported developer integration. Use it for personal automation only.
- **Rate limits** — LinkedIn rate-limits API requests. Avoid bulk or automated operations that could trigger throttling.
- **Cookie security** — Your `li_at` cookie grants full access to your LinkedIn account. Never share it or commit it to version control. Credential files are stored with `0600` permissions.

---

## License

MIT
