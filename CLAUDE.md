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
make test         # run integration tests (verbose)
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
  profile.go            linked profile get|update|skills|contact|who-viewed|photo|
                          experience|education|certifications|languages|
                          volunteer|projects|publications|honors|courses|
                          open-to-work|close-to-work
  search.go             linked search people|jobs|companies|posts
  messages.go           linked messages list|unread|read|send|mark-read|
                          star|unstar|archive|unarchive|delete|delete-conversation
  connections.go        linked connections list|pending|sent|request|accept|ignore|
                          withdraw|remove|follow|unfollow|mutual
  jobs.go               linked jobs get|saved|applied|save|unsave|recommended|company
  companies.go          linked companies get|follow|unfollow|posts|employees
  posts.go              linked posts feed|create|create-with-image|get|edit|delete|
                          like|unlike|comment|share|comments|delete-comment|
                          like-comment|activity
  recommendations.go    linked recommendations received|given|request|hide|show|decline|delete
  notifications.go      linked notifications list|mark-read|mark-all-read|count

internal/
  api/                  LinkedIn Voyager API service layer
    linkedin.go         api.LinkedIn struct — holds all service pointers
    profile.go          GetMe, GetProfile, GetContactInfo, UpdateProfile,
                          AddExperience, UpdateExperience, DeleteExperience,
                          AddEducation, UpdateEducation, DeleteEducation,
                          AddSkill, DeleteSkill,
                          AddCertification, UpdateCertification, DeleteCertification,
                          AddLanguage, DeleteLanguage,
                          AddVolunteer, UpdateVolunteer, DeleteVolunteer,
                          AddProject, UpdateProject, DeleteProject,
                          AddPublication, UpdatePublication, DeletePublication,
                          AddHonor, DeleteHonor,
                          AddCourse, DeleteCourse,
                          SetOpenToWork, ClearOpenToWork,
                          GetWhoViewed, UploadProfilePhoto
    search.go           SearchPeople, SearchJobs, SearchCompanies, SearchPosts
    messaging.go        ListConversations, GetConversation, ListUnread, SendMessage,
                          MarkRead, StarConversation, UnstarConversation,
                          ArchiveConversation, UnarchiveConversation,
                          DeleteMessage, DeleteConversation
    connections.go      ListConnections, ListPendingInvitations, ListSentInvitations,
                          SendConnectionRequest, AcceptInvitation, IgnoreInvitation,
                          WithdrawInvitation, RemoveConnection,
                          FollowProfile, UnfollowProfile, GetMutualConnections
    jobs.go             GetJob, ListSavedJobs, SaveJob, UnsaveJob, ListAppliedJobs,
                          GetRecommendedJobs, SearchJobsByCompany
    companies.go        GetCompany, FollowCompany, UnfollowCompany,
                          GetCompanyPosts, GetCompanyEmployees
    posts.go            GetFeed, GetMemberPosts, GetPost, CreatePost,
                          CreatePostWithImage, EditPost, DeletePost,
                          LikePost, UnlikePost, CommentOnPost, SharePost,
                          GetComments, DeleteComment, LikeComment
    recommendations.go  ListReceived, ListGiven, RequestRecommendation,
                          HideRecommendation, ShowRecommendation,
                          DeclineRecommendation, DeleteRecommendation
    notifications.go    List, MarkRead, MarkAllRead, GetBadgeCount
    util.go             Shared helpers (urnToID, msToTime)
  client/
    client.go           HTTP client: cookie jar, CSRF header injection,
                          GET/POST/PUT/DELETE/PutBinary wrappers
    endpoints.go        BaseURL and all Voyager API path constants
    errors.go           apiError — maps HTTP status codes to structured errors (401, 404, 429)
  config/
    credentials.go      Load/Save/Delete/List credentials at
                          ~/.openclaw/credentials/linkedin/<profile>/creds.json
  models/
    models.go           All Go structs for LinkedIn data types:
                          Profile, Experience, Education, Skill, Language,
                          Certification, VolunteerExperience, Project, Publication,
                          Honor, Course, OpenToWork, ProfileViewer,
                          Connection, Invitation, Conversation, Message,
                          Job, Company, Post, Comment, Recommendation,
                          Notification, NotificationBadge, MutualConnection,
                          and all Paged* wrappers
    inputs.go           Input structs for API write operations
  output/
    output.go           Printer: pretty (color), json, table renderers; output.Error()

mock/
  server.go             httptest-based mock Voyager server with CSRF middleware
  fixtures/             Realistic JSON response fixtures

tests/integration/      Integration test suite (uses mock server, no network)
  helpers_test.go       newTestClient() — spins up mock.New() and wires client
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
- Cookie-based: `li_at` (session token) + `JSESSIONID` (CSRF token) + `bcookie` (browser fingerprint)
- LinkedIn binds `li_at` to the `bcookie` it was issued with — sending a mismatched `bcookie` causes LinkedIn to immediately revoke the session (302 + `Set-Cookie: li_at=delete me`)
- All three cookies must come from the same browser session
- CSRF token = `JSESSIONID` with surrounding quotes stripped
- All required headers set per request: `csrf-token`, `x-restli-protocol-version`, `x-li-lang`, `x-li-track`, `Accept`
- Credentials stored at `~/.openclaw/credentials/linkedin/<profile>/creds.json` (mode 0600)

### Client Layer
- `client.Client` wraps `net/http` with a cookie jar + custom `cookieTransport`; exposes `Get`, `Post`, `Put`, `Delete`, `PutBinary`
- `cookieTransport` injects `li_at`, `JSESSIONID`, and `bcookie` at the transport level so they are present on every request including redirect hops, with exact values (bypassing Go's cookie sanitizer which strips `"` from cookie values)
- Go's `net/http` cookie jar sanitizes cookie values and strips `"` characters — this breaks LinkedIn's quoted `JSESSIONID` and `bcookie` values, hence the transport-level injection
- `client.NewWithBaseURL` accepts a custom base URL so tests can point at the mock server
- Errors are typed: 401 → auth error, 429 → rate limit error, 404 → not found

### Media Upload Flow
LinkedIn requires a three-step process for profile photos and image posts:
1. POST to register the upload → get `uploadUrl` + `assetUrn`
2. PUT binary data to `uploadUrl` via `PutBinary`
3. POST/PUT the asset URN in the actual create/update request

### CLI Layer
- Global flags: `--output` (pretty|json|table, default: pretty), `--profile` (default: "default")
- Each command file constructs a `*api.LinkedIn` via `newLinkedIn()` (loads credentials) and a `*output.Printer` via `newPrinter()`
- `SilenceUsage: true` and `SilenceErrors: true` on root — errors printed via `output.Error()` then `os.Exit(1)`

### Testing
- `mock.New()` starts an `httptest.Server` with CSRF validation middleware
- Tests use `mock.TestCredentials()` to get valid test cookies
- The mock server is stateful for mutating operations: tracks sent messages, created posts, saved jobs, likes, follows
- All tests pass with zero network calls

### Output
- Three formats: `pretty` (colored, human-readable), `json` (raw JSON for piping/agents), `table` (tabular)
- OpenClaw always uses `-o json`

---

## Adding a New Command

1. Add the API method(s) to the appropriate file in `internal/api/`
2. Add any new request/response structs to `internal/models/`
3. Add mock handler(s) to `mock/server.go` and a fixture if needed
4. Add a Cobra command in the relevant `cli/*.go` file and register it in the subcommand list
5. Add integration tests in `tests/integration/`
6. Update `skill/linkedin/skill.md` so OpenClaw knows about the new command

---

## Credential File Format

```json
{
  "li_at": "AQEDARxxxxxxx...",
  "jsessionid": "\"ajax:1234567890abcdef\"",
  "bcookie": "\"v=2&xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\"",
  "profileId": "your-profile-id",
  "createdAt": "2025-01-01T00:00:00Z"
}
```

Stored at: `~/.openclaw/credentials/linkedin/<profile>/creds.json`

---

## Live API Test Results (2026-03-08)

Auth was failing (session revoked after 2 calls). **Root cause: stale client fingerprint headers.**
Fixed by updating `internal/client/client.go` to match real Chrome 145 values:
- `userAgent` → Chrome/145
- `clientVersion` → `1.13.42665`
- `x-li-track` → added `mpVersion`, `displayDensity`, `displayWidth`, `displayHeight`; timezone `America/New_York`
- Added `sec-ch-ua`, `sec-ch-ua-mobile`, `sec-ch-ua-platform` headers

**When LinkedIn starts rejecting again:** capture a fresh HAR from Chrome DevTools (Network tab → right-click → Save all as HAR) and update these values to match the real browser.

### Read-only command status

| Command | Status | Notes |
|---------|--------|-------|
| `auth whoami` | ✅ | |
| `auth list` | ✅ | |
| `profile get` | ✅ | |
| `profile skills` | ✅ | returns empty if none set |
| `profile contact` | ❌ 410 | LinkedIn deprecated this endpoint |
| `profile experience list` | ✅ | returns empty if none set |
| `profile education list` | ✅ | returns empty if none set |
| `profile who-viewed` | ✅ | total count returned for free accounts; individual names require Premium |
| `search people` | ❌ 404 | `/voyager/api/search/hits` is deprecated — needs new endpoint |
| `search jobs` | ❌ 404 | needs investigation |
| `search companies` | ❌ 404 | `/voyager/api/search/hits` is deprecated — needs new endpoint |
| `search posts` | ❌ 404 | needs investigation |
| `messages list` | 🔲 not tested | |
| `messages unread` | 🔲 not tested | |
| `messages read` | 🔲 not tested | |
| `connections list` | 🔲 not tested | |
| `connections pending` | 🔲 not tested | |
| `connections sent` | 🔲 not tested | |
| `connections mutual` | 🔲 not tested | |
| `jobs recommended` | 🔲 not tested | |
| `jobs saved` | 🔲 not tested | |
| `jobs applied` | 🔲 not tested | |
| `jobs get` | 🔲 not tested | |
| `jobs company` | 🔲 not tested | |
| `companies get` | 🔲 not tested | |
| `companies employees` | 🔲 not tested | |
| `companies posts` | 🔲 not tested | |
| `posts feed` | 🔲 not tested | |
| `posts get` | 🔲 not tested | |
| `posts comments` | 🔲 not tested | |
| `posts activity` | 🔲 not tested | |
| `recommendations received` | 🔲 not tested | |
| `recommendations given` | 🔲 not tested | |
| `notifications list` | 🔲 not tested | |
| `notifications count` | 🔲 not tested | |

### Known broken endpoints (need fixes)

1. **`profile contact`** — `GET /voyager/api/identity/profiles/:id/profileContactInfo` returns 410. LinkedIn deprecated this. Need to find replacement or remove command.
2. **`search people` / `search companies`** — `GET /voyager/api/search/hits` returns 404. Endpoint is deprecated. HAR shows LinkedIn now uses `/voyager/api/graphql` with `queryId` params for search. Needs research to find correct replacement endpoint.
3. **`search jobs`** — `GET /voyager/api/jobs/search` returns 404. Needs replacement endpoint.
4. **`search posts`** — `GET /voyager/api/search/blended` returns 404. Needs replacement endpoint.

### Fixed endpoints

- **`profile who-viewed`** — Was returning 400 from the deprecated `/identity/wvmpCards` endpoint. Fixed (2026-03-08) to use two GraphQL calls:
  1. `voyagerFeedDashIdentityModule.803fe19f843a4d461478049f70d7babd` — returns the 90-day viewer count in `feedDashIdentityModuleByModuleType.elements[].widgets[]` where `widgetType == "WHO_VIEWED_MY_PROFILE"` and `statistic.text` is the count string. Available to free accounts.
  2. `voyagerPremiumDashAnalyticsObject.faf9c8e3233e83980f323f07c637b3c3` — returns individual viewer details. Requires LinkedIn Premium; returns empty elements for free accounts.
