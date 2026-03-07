# CLAUDE.md — linked

## Project Overview

`linked` is a LinkedIn command-line interface written in Go, targeting LinkedIn's internal Voyager API (cookie-based auth, no developer account required). It is designed to integrate with OpenClaw as an installable skill.

Module: `github.com/russ-blaisdell/linked`
Binary: `linked`
Go version: 1.24+

---

## Development Commands

```bash
make build        # compile → dist/linked
make test         # run 56 integration tests (verbose)
make test-short   # run tests (no verbose output)
make lint         # go vet ./...
make install      # build + install to /usr/local/bin/linked
make release      # cross-compile darwin/arm64, darwin/amd64, linux/amd64
make clean        # remove dist/
make skill        # install OpenClaw skill to ~/.openclaw/workspace/skills/linkedin/
```

Tests run with no network access — all requests go to the mock server.

---

## Repository Structure

```
cmd/linked/              Entry point (main.go calls cli.Execute())
cli/                    Cobra command definitions, one file per domain
  root.go               Root command, global flags (--output, --profile), subcommand wiring
  auth.go               linked auth setup|whoami|remove|list
  profile.go            linked profile get|update|skills|contact
  search.go             linked search people|jobs|companies|posts
  messages.go           linked messages list|unread|read|send
  connections.go        linked connections list|pending|sent|request|accept|ignore|withdraw|follow|unfollow
  jobs.go               linked jobs get|saved|applied|save|unsave
  companies.go          linked companies get|follow|unfollow|posts
  posts.go              linked posts feed|create|like|unlike|comment|share|comments
  recommendations.go    linked recommendations received|given|request|hide|show
  notifications.go      linked notifications list|mark-read

internal/
  api/                  LinkedIn Voyager API service layer
    linkedin.go         api.LinkedIn struct (holds *client.Client), entry point for all API calls
    profile.go          GetMe, GetProfile, GetContactInfo, UpdateProfile, GetSkills, GetExperience, GetEducation
    search.go           SearchPeople, SearchJobs, SearchCompanies, SearchPosts
    messaging.go        ListConversations, GetConversation, ListUnread, SendMessage, MarkRead
    connections.go      ListConnections, ListPendingInvitations, ListSentInvitations, SendConnectionRequest, AcceptInvitation, IgnoreInvitation, WithdrawInvitation, Follow, Unfollow
    jobs.go             GetJob, ListSavedJobs, SaveJob, UnsaveJob, ListAppliedJobs
    companies.go        GetCompany, FollowCompany, UnfollowCompany, GetCompanyPosts
    posts.go            GetFeed, CreatePost, LikePost, UnlikePost, CommentOnPost, SharePost, GetComments
    recommendations.go  ListReceivedRecommendations, ListGivenRecommendations, RequestRecommendation, HideRecommendation, ShowRecommendation
    notifications.go    ListNotifications, MarkNotificationRead
    util.go             Shared helpers (pagination helpers, URN helpers, etc.)
  client/
    client.go           HTTP client: cookie jar setup, CSRF header injection, GET/POST/PUT/DELETE wrappers
    endpoints.go        BaseURL and all Voyager API path constants
    errors.go           apiError — maps HTTP status codes to structured error types (401, 404, 429)
  config/
    credentials.go      Load/Save/Delete/List credentials at ~/.openclaw/credentials/linkedin/<profile>/creds.json
  models/
    models.go           All Go structs for LinkedIn data types
    inputs.go           Input structs for API calls (request bodies)
  output/
    output.go           Printer: pretty (color), json, table renderers; output.Error()

mock/
  server.go             httptest-based mock Voyager server with auth middleware and stateful handlers
  fixtures/             Realistic JSON response fixtures
    profile.json
    conversations.json
    jobs.json
    connections.json
    notifications.json
    recommendations.json

tests/integration/      Full integration test suite (56 tests, uses mock server)
  helpers_test.go       newTestClient() helper — spins up mock.New() and wires client
  *_test.go             One file per domain

skill/linkedin/
  skill.md              OpenClaw skill definition (installed by `make skill`)

docs/
  auth.md               Step-by-step cookie extraction for Chrome/Firefox/Safari
  openclaw-setup.md     OpenClaw integration setup guide
```

---

## Key Design Decisions

### Authentication
- Cookie-based: `li_at` (session token) + `JSESSIONID` (CSRF token)
- CSRF token = `JSESSIONID` with surrounding quotes stripped
- All required headers set per request: `csrf-token`, `x-restli-protocol-version`, `x-li-lang`, `x-li-track`, `Accept`
- Credentials stored at `~/.openclaw/credentials/linkedin/<profile>/creds.json` (mode 0600)

### Client Layer
- `client.Client` wraps `net/http` with a cookie jar; exposes `Get`, `Post`, `Put`, `Delete`
- `client.NewWithBaseURL` accepts a custom base URL so tests can point at the mock server
- Errors are typed: 401 → auth error, 429 → rate limit error, 404 → not found

### CLI Layer
- Global flags: `--output` (pretty|json|table, default: pretty), `--profile` (default: "default")
- Each command file constructs a `*api.LinkedIn` via `newLinkedIn()` (loads credentials) and a `*output.Printer` via `newPrinter()`
- `SilenceUsage: true` and `SilenceErrors: true` on root — errors printed via `output.Error()` then `os.Exit(1)`

### Testing
- `mock.New()` starts an `httptest.Server` with CSRF validation middleware
- Tests use `mock.TestCredentials()` to get valid test cookies
- The mock server is stateful for mutating operations: tracks sent messages, created posts, saved jobs, likes, follows
- All 56 tests pass with zero network calls

### Output
- Three formats: `pretty` (colored, human-readable), `json` (raw JSON for piping/agents), `table` (tabular)
- OpenClaw always uses `-o json`

---

## Adding a New Command

1. Add the API method(s) to the appropriate file in `internal/api/`
2. Add any new request/response structs to `internal/models/`
3. Add mock handler(s) to `mock/server.go` and a fixture if needed
4. Add a Cobra command in the relevant `cli/*.go` file and register it in `root.go init()`
5. Add integration tests in `tests/integration/`
6. Update `skill/linkedin/skill.md` so OpenClaw knows about the new command

---

## Credential File Format

```json
{
  "li_at": "AQEDARxxxxxxx...",
  "jsessionid": "ajax:1234567890abcdef",
  "profile": "default",
  "created_at": "2025-01-01T00:00:00Z"
}
```

Stored at: `~/.openclaw/credentials/linkedin/<profile>/creds.json`
