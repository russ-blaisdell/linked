# linked — LinkedIn CLI for OpenClaw

A full-featured LinkedIn command-line tool written in Go. Uses LinkedIn's internal Voyager API — cookie-based authentication, no developer account or API key required.

Designed to integrate with [OpenClaw](https://www.getopenclaw.ai/) as an installable skill, but works as a standalone CLI.

---

## Features

| Domain | Commands |
|--------|----------|
| **Auth** | Setup, verify, remove, and list credential profiles |
| **Profile** | View/update profile; full CRUD for experience, education, skills, certifications, languages, volunteer, projects, publications, honors, courses; open-to-work toggle; who viewed; upload photo |
| **Search** | Search people, jobs, companies, and posts with filters |
| **Messages** | List, read, send; star/unstar, archive/unarchive, delete messages and conversations; mark as read |
| **Connections** | List, send/accept/ignore/withdraw/remove invitations; follow/unfollow; mutual connections |
| **Jobs** | Search, view, save/unsave, saved/applied/recommended jobs; search by company |
| **Companies** | Company info, follow/unfollow, company posts, employee search |
| **Posts & Feed** | Home feed; create text or image posts; get, edit, delete; like/unlike; comment, share, delete comments, like comments; member activity |
| **Recommendations** | View received/given; request, hide, show, decline, delete |
| **Notifications** | List, mark as read (single or all), unread badge count |

---

## Installation

### Prerequisites

- Go 1.24+
- A LinkedIn account (logged in via browser)

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

`linked` authenticates using your LinkedIn session cookies — the same cookies your browser uses. No LinkedIn developer account is needed.

### Setup

```bash
linked auth setup
```

You will be prompted for two values from your browser session on linkedin.com:

| Cookie | Description |
|--------|-------------|
| `li_at` | Your main session token (`AQEDARxxxxxxx...`) |
| `JSESSIONID` | CSRF token (starts with `ajax:`) |

**How to get your cookies:**

1. Log in to [linkedin.com](https://www.linkedin.com) in your browser
2. Open DevTools (`⌘ Opt I` on Mac / `F12` on Windows)
3. Go to **Application → Cookies → https://www.linkedin.com** (Chrome/Edge) or **Storage → Cookies** (Firefox/Safari)
4. Copy the values for `li_at` and `JSESSIONID`

See [docs/auth.md](docs/auth.md) for detailed browser-specific instructions.

Credentials are stored at `~/.openclaw/credentials/linkedin/default/creds.json` with `0600` permissions.

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

`li_at` cookies typically last one year. If you get a 401 error, log out and back in to LinkedIn, then re-run `linked auth setup`.

---

## Quick Start

```bash
# Verify authentication
linked auth whoami

# Search for remote Go jobs
linked search jobs "golang engineer" --remote

# View recommended jobs
linked jobs recommended

# Check unread messages
linked messages unread

# Reply to a conversation
linked messages send "Thanks for reaching out!" --conversation 2-abc123

# Search for people at a company
linked search people "product manager" --company google --network FIRST,SECOND

# View your profile
linked profile get

# Update your headline
linked profile update --headline "Principal Engineer at Acme"

# Add a new position to your profile
linked profile experience add --title "Staff Engineer" --company "Acme Corp" --current

# Find who viewed your profile recently
linked profile who-viewed

# Upload a new profile photo
linked profile photo ./headshot.jpg

# List your connections
linked connections list

# See mutual connections with someone
linked connections mutual urn:li:member:12345678

# Send a connection request
linked connections request urn:li:member:12345678 --note "Hi, we met at GopherCon!"

# Create a post with an image
linked posts create-with-image "Excited to share this!" --image ./photo.jpg

# See recent notifications
linked notifications list

# Mark all notifications as read
linked notifications mark-all-read

# Get unread notification count
linked notifications count

# View recommendations on your profile
linked recommendations received

# Request a recommendation
linked recommendations request urn:li:member:12345678 \
  --relationship COLLEAGUE \
  --message "Hi, would you be willing to write me a recommendation?"
```

---

## Output Formats

All commands support `--output` / `-o`:

```bash
linked search jobs "engineer" -o json     # machine-readable JSON
linked connections list -o table          # tabular output
linked profile get -o pretty              # human-readable with colour (default)
```

Use `-o json` when piping to other tools or in agent workflows.

---

## Full Command Reference

### Auth

```
linked auth setup   [--profile <name>]   # configure credentials
linked auth whoami  [--profile <name>]   # verify auth and show current account
linked auth remove  [--profile <name>]   # delete stored credentials
linked auth list                         # list all configured profiles
```

### Profile

```
linked profile get      [--urn <member-urn>]         # view a profile (own if --urn omitted)
linked profile update   [--headline "..."]            # update top-level profile fields
                        [--summary "..."]
                        [--location "..."]
linked profile skills   [--urn <member-urn>]         # list profile skills
linked profile contact  [--urn <member-urn>]         # view contact info
linked profile who-viewed                            # see who viewed your profile recently
linked profile photo <path>                          # upload a new profile photo

linked profile experience list                       # list your positions
linked profile experience add    [flags]             # add a new position
linked profile experience update <id> [flags]        # update a position
linked profile experience remove <id>                # delete a position

linked profile education list                        # list your education
linked profile education add    [flags]              # add an education entry
linked profile education update <id> [flags]         # update an education entry
linked profile education remove <id>                 # delete an education entry

linked profile skills list                           # list skills
linked profile skills add <name>                     # add a skill
linked profile skills remove <id>                    # remove a skill

linked profile certifications list                   # list certifications
linked profile certifications add    [flags]         # add a certification
linked profile certifications update <id> [flags]    # update a certification
linked profile certifications remove <id>            # remove a certification

linked profile languages list                        # list languages
linked profile languages add    [flags]              # add a language
linked profile languages remove <id>                 # remove a language

linked profile volunteer list                        # list volunteer experiences
linked profile volunteer add    [flags]              # add a volunteer experience
linked profile volunteer update <id> [flags]         # update a volunteer experience
linked profile volunteer remove <id>                 # remove a volunteer experience

linked profile projects list                         # list projects
linked profile projects add    [flags]               # add a project
linked profile projects update <id> [flags]          # update a project
linked profile projects remove <id>                  # remove a project

linked profile publications list                     # list publications
linked profile publications add    [flags]           # add a publication
linked profile publications update <id> [flags]      # update a publication
linked profile publications remove <id>              # remove a publication

linked profile honors list                           # list honors and awards
linked profile honors add    [flags]                 # add an honor
linked profile honors remove <id>                    # remove an honor

linked profile courses list                          # list courses
linked profile courses add    [flags]                # add a course
linked profile courses remove <id>                   # remove a course

linked profile open-to-work   [flags]                # set Open to Work status
linked profile close-to-work                         # clear Open to Work status
```

### Search

```
linked search people "<keywords>"
  --company <name>                  # filter by company
  --title <title>                   # filter by job title
  --school <school>                 # filter by school
  --network FIRST,SECOND,THIRD      # filter by network distance
  --location "<location>"           # filter by location
  --count <n>                       # number of results
  --start <n>                       # pagination offset

linked search jobs "<keywords>"
  --location "<location>"
  --remote                          # remote jobs only
  --company <name>
  --experience-level <level>        # ENTRY_LEVEL, MID_SENIOR_LEVEL, DIRECTOR, etc.
  --employment-type <type>          # FULL_TIME, PART_TIME, CONTRACT, etc.
  --count <n>
  --start <n>

linked search companies "<keywords>"   [--count <n>]
linked search posts "<keywords>"       [--count <n>]
```

### Messages

```
linked messages list          [--count <n>] [--start <n>]  # all conversations
linked messages unread        [--count <n>]                # unread only
linked messages read <conversation-id>                     # read a thread
linked messages send "<text>"
  --conversation <id>    # reply to existing conversation
  --to <member-urn>      # start a new conversation (repeatable)
linked messages mark-read <conversation-id>                # mark conversation as read
linked messages star              <conversation-id>        # star a conversation
linked messages unstar            <conversation-id>        # remove star
linked messages archive           <conversation-id>        # archive a conversation
linked messages unarchive         <conversation-id>        # restore from archive
linked messages delete            <conversation-id> <message-urn>  # delete a message
linked messages delete-conversation <conversation-id>      # delete entire conversation
```

### Connections

```
linked connections list     [--count <n>]              # 1st-degree connections
linked connections pending  [--count <n>]              # received invitations
linked connections sent     [--count <n>]              # sent invitations

linked connections request  <member-urn> [--note "..."] # send connection request
linked connections accept   <invitation-id>             # accept received invitation
linked connections ignore   <invitation-id>             # ignore received invitation
linked connections withdraw <invitation-urn>            # withdraw sent invitation
linked connections remove   <member-urn>                # remove a connection

linked connections follow   <member-urn>                # follow a member
linked connections unfollow <member-urn>                # unfollow a member
linked connections mutual   <member-urn>  [--count <n>] # mutual connections
```

### Jobs

```
linked jobs get         <job-id>         # view job posting details
linked jobs saved       [--count <n>]    # saved jobs
linked jobs save        <job-id>         # save a job
linked jobs unsave      <job-id>         # remove from saved
linked jobs applied     [--count <n>]    # applied jobs
linked jobs recommended [--count <n>]    # LinkedIn-recommended jobs for you
linked jobs company     <company-urn>    # jobs posted by a specific company
```

### Companies

```
linked companies get       <universal-name>    # company info (e.g. "anthropic")
linked companies follow    <company-urn>       # follow a company
linked companies unfollow  <company-urn>       # unfollow a company
linked companies posts     <company-urn>       # recent posts from a company
linked companies employees <company-id>        # search members who work there
```

### Posts & Feed

```
linked posts feed                                              # home feed
linked posts create "<text>"  [--visibility PUBLIC|CONNECTIONS]
linked posts create-with-image "<text>" --image <path>        # post with image
                              [--visibility PUBLIC|CONNECTIONS]
linked posts get     <post-urn>                               # get a post
linked posts edit    <post-urn> "<new-text>"                  # edit a post
linked posts delete  <post-urn>                               # delete a post
linked posts like    <post-urn>                               # like a post
linked posts unlike  <post-urn>                               # remove like
linked posts comment <post-urn> "<text>"                      # comment on a post
linked posts share   <post-urn> [--commentary "..."]          # reshare a post
linked posts comments       <post-urn>  [--count <n>]         # view comments
linked posts delete-comment <post-urn> <comment-urn>          # delete a comment
linked posts like-comment   <post-urn> <comment-urn>          # like a comment
linked posts activity       <profile-id> [--count <n>]        # member's recent posts
```

### Recommendations

```
linked recommendations received                              # recommendations on your profile
linked recommendations given                                 # recommendations you've written
linked recommendations request <member-urn>
  --relationship COLLEAGUE|MANAGER|REPORT|CLASSMATE|...
  --message "..."
linked recommendations hide    <recommendation-urn>          # hide from your profile
linked recommendations show    <recommendation-urn>          # show on your profile
linked recommendations decline <recommendation-urn>          # decline a pending request
linked recommendations delete  <recommendation-urn>          # delete a recommendation you gave
```

### Notifications

```
linked notifications list                           # recent notifications
linked notifications mark-read <notification-urn>  # mark one as read
linked notifications mark-all-read                 # mark all as read
linked notifications count                         # unread badge count
```

---

## OpenClaw Integration

Install the skill so OpenClaw can use `linked` on your behalf:

```bash
make skill
openclaw gateway   # restart to pick up the new skill
```

Then talk to your agent naturally:

```
@openclaw find remote golang jobs on LinkedIn
@openclaw show my unread LinkedIn messages
@openclaw update my LinkedIn headline to "Senior Engineer at Acme"
@openclaw who sent me connection requests recently?
@openclaw request a recommendation from my colleague John
@openclaw who viewed my LinkedIn profile this week?
@openclaw post an update to LinkedIn: "Excited to announce..."
@openclaw how many unread notifications do I have on LinkedIn?
```

See [docs/openclaw-setup.md](docs/openclaw-setup.md) for full setup instructions.

---

## Architecture

```
cmd/linked/          Entry point — calls cli.Execute()
cli/                Cobra commands, one file per domain
internal/
  api/              LinkedIn Voyager API service layer
  client/           HTTP client (auth, CSRF headers, error handling)
  config/           Credential storage (~/.openclaw/credentials/linkedin/)
  models/           Go structs for all LinkedIn data types
  output/           pretty / JSON / table renderers
mock/               In-process mock Voyager server for testing
  fixtures/         Realistic JSON response fixtures
tests/integration/  Integration test suite (no network)
skill/linkedin/     OpenClaw skill definition
docs/               Auth guide, OpenClaw setup guide
```

### HTTP Client

The `client.Client` wraps `net/http` with a cookie jar and automatically injects all headers required by the Voyager API:

- `csrf-token` — derived from `JSESSIONID`
- `x-restli-protocol-version: 2.0.0`
- `x-li-lang: en_US`
- `x-li-track` — client version and device info
- `Accept: application/vnd.linkedin.normalized+json+2.1`
- `User-Agent` — spoofs a standard browser UA

---

## Development

```bash
make build      # compile binary → dist/linked
make test       # run integration tests (verbose)
make lint       # go vet ./...
make release    # cross-platform: darwin/arm64, darwin/amd64, linux/amd64
```

### Running tests

Tests run entirely against an in-process mock server — no network, no credentials required:

```bash
make test
```

The mock server:
- Validates the `csrf-token` header on every request (returns 401 if missing/wrong)
- Serves realistic JSON fixtures from `mock/fixtures/`
- Tracks stateful operations (sent messages, created posts, saved jobs, likes, follows) for assertion in tests

### Adding a feature

1. Add API method(s) to `internal/api/`
2. Add structs to `internal/models/` if needed
3. Add mock handler(s) in `mock/server.go` and a fixture if needed
4. Add the Cobra command in `cli/` and register it in the subcommand list
5. Add integration tests in `tests/integration/`
6. Update `skill/linkedin/skill.md` with the new command

---

## Important Notes

- **Terms of Service**: This tool uses LinkedIn's internal Voyager API — the same API the LinkedIn website itself uses. It is not an officially supported developer integration. Use for personal automation only.
- **Rate limits**: LinkedIn rate-limits API requests. Avoid bulk or automated operations.
- **Cookie security**: Your `li_at` cookie grants full access to your LinkedIn account. Never share it or commit it to version control.

---

## License

MIT
