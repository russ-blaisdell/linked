# linked — LinkedIn CLI for OpenClaw

A full-featured LinkedIn command-line tool written in Go. Uses LinkedIn's internal Voyager API — cookie-based authentication, no developer account or API key required.

Designed to integrate with [OpenClaw](https://www.getopenclaw.ai/) as an installable skill, but works as a standalone CLI.

---

## Features

| Domain | Commands |
|--------|----------|
| **Auth** | Setup, verify, remove, and list credential profiles |
| **Profile** | View own or others' profiles, update headline/summary/location, view skills, experience, education, and contact info |
| **Search** | Search people, jobs, companies, and posts with filters |
| **Messages** | List conversations, read threads, send messages, mark as read |
| **Connections** | List connections, send/accept/ignore/withdraw invitations, follow/unfollow |
| **Jobs** | Search, view, save/unsave, view saved and applied jobs |
| **Companies** | Company info, follow/unfollow, view company posts |
| **Posts & Feed** | Home feed, create posts, like/unlike, comment, share, view comments |
| **Recommendations** | View received/given, request recommendations, hide/show |
| **Notifications** | List notifications, mark as read |

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

# View a job posting
linked jobs get 987654321

# Check unread messages
linked messages unread

# Reply to a conversation
linked messages send "Thanks for reaching out!" --conversation 2-abc123

# Search for people
linked search people "product manager" --company google --network FIRST,SECOND

# View your profile
linked profile get

# Update your headline
linked profile update --headline "Principal Engineer at Acme"

# List your connections
linked connections list

# Send a connection request
linked connections request urn:li:member:12345678 --note "Hi, we met at GopherCon!"

# View recommendations on your profile
linked recommendations received

# Request a recommendation
linked recommendations request urn:li:member:12345678 \
  --relationship COLLEAGUE \
  --message "Hi, would you be willing to write me a recommendation?"

# View notifications
linked notifications list
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
linked profile get      [--urn <member-urn>]          # view a profile (own if --urn omitted)
linked profile update   [--headline "..."]             # update profile fields
                       [--summary "..."]
                       [--location "..."]
linked profile skills   [--urn <member-urn>]          # list profile skills
linked profile contact  [--urn <member-urn>]          # view contact info
```

### Search

```
linked search people "<keywords>"
  --company <name>                  # filter by company
  --title <title>                   # filter by job title
  --school <school>                 # filter by school
  --network FIRST,SECOND,THIRD      # filter by network distance
  --location "<location>"           # filter by location
  --count <n>                       # number of results (default: 10)
  --start <n>                       # pagination offset

linked search jobs "<keywords>"
  --location "<location>"
  --remote                          # remote jobs only
  --company <name>
  --experience-level <level>        # ENTRY, MID, SENIOR, etc.
  --employment-type <type>          # FULL_TIME, CONTRACT, etc.
  --count <n>
  --start <n>

linked search companies "<keywords>"   [--count <n>]
linked search posts "<keywords>"       [--count <n>]
```

### Messages

```
linked messages list   [--count <n>] [--start <n>]    # all conversations
linked messages unread [--count <n>]                  # unread only
linked messages read <conversation-id>                # read a thread
linked messages send "<text>"
  --conversation <id>    # reply to existing conversation
  --to <member-urn>      # start a new conversation
```

### Connections

```
linked connections list    [--count <n>]              # 1st-degree connections
linked connections pending [--count <n>]              # received invitations
linked connections sent    [--count <n>]              # sent invitations

linked connections request <member-urn> [--note "..."] # send connection request
linked connections accept  <invitation-id>             # accept received invitation
linked connections ignore  <invitation-id>             # ignore received invitation
linked connections withdraw <invitation-urn>           # withdraw sent invitation

linked connections follow   <member-urn>               # follow a member
linked connections unfollow <member-urn>               # unfollow a member
```

### Jobs

```
linked jobs get    <job-id>              # view job posting details
linked jobs saved  [--count <n>]        # saved jobs
linked jobs save   <job-id>             # save a job
linked jobs unsave <job-id>             # remove from saved
linked jobs applied [--count <n>]       # applied jobs
```

### Companies

```
linked companies get      <universal-name>    # company info (e.g. "anthropic")
linked companies follow   <company-urn>       # follow a company
linked companies unfollow <company-urn>       # unfollow a company
linked companies posts    <company-urn>       # recent posts from a company
```

### Posts & Feed

```
linked posts feed                                # home feed
linked posts create "<text>" [--visibility PUBLIC|CONNECTIONS]
linked posts like    <post-urn>                  # like a post
linked posts unlike  <post-urn>                  # remove like
linked posts comment <post-urn> "<text>"         # comment on a post
linked posts share   <post-urn>                  # reshare a post
linked posts comments <post-urn>                 # view comments on a post
```

### Recommendations

```
linked recommendations received                  # recommendations on your profile
linked recommendations given                     # recommendations you've written
linked recommendations request <member-urn>
  --relationship COLLEAGUE|MANAGER|REPORT|...
  --message "..."
linked recommendations hide <recommendation-urn> # hide from your profile
linked recommendations show <recommendation-urn> # show on your profile
```

### Notifications

```
linked notifications list                        # recent notifications
linked notifications mark-read <notification-urn>
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
tests/integration/  56-test integration suite (no network)
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
make test       # run 56 integration tests (verbose)
make lint       # go vet ./...
make release    # cross-platform: darwin/arm64, darwin/amd64, linux/amd64
```

### Running tests

Tests run entirely against an in-process mock server — no network, no credentials required:

```bash
make test
# ok   github.com/russ-blaisdell/linked/tests/integration   0.4s (56 tests)
```

The mock server:
- Validates the `csrf-token` header on every request (returns 401 if missing/wrong)
- Serves realistic JSON fixtures from `mock/fixtures/`
- Tracks stateful operations (sent messages, created posts, saved jobs, likes, follows) for assertion in tests

### Adding a feature

1. Add API method(s) to `internal/api/`
2. Add structs to `internal/models/` if needed
3. Add mock handler(s) in `mock/server.go` and a fixture if needed
4. Add the Cobra command in `cli/` and register it in `cli/root.go`
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
