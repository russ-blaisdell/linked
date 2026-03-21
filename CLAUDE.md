# CLAUDE.md — linked

## Project Overview

`linked` is a LinkedIn command-line interface written in Go, targeting LinkedIn's internal Voyager API (cookie-based auth, no developer account required). It is designed to integrate with OpenClaw as an installable skill.

Module: `github.com/russ-blaisdell/linked`
Binary: `linked`
Go version: 1.25+

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
  raw.go                linked raw <path> — authenticated GET for API exploration

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
                          GET/POST/PUT/DELETE/PutBinary/GetGraphQL/RawGet wrappers
    tls.go              Chrome TLS fingerprint via utls + HTTP/2 transport
    endpoints.go        BaseURL, Voyager API path constants, GraphQL queryId constants
    errors.go           apiError — maps HTTP status codes to structured errors (401, 404, 429)
  harparser/
    harparser.go        Parse HAR files for browser fingerprint + cookies
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
- `client.Client` wraps `net/http` with a cookie jar + custom `cookieTransport`; exposes `Get`, `Post`, `Put`, `Delete`, `PutBinary`, `GetGraphQL`, `RawGet`
- `cookieTransport` injects `li_at`, `JSESSIONID`, `bcookie`, and `bscookie` at the transport level so they are present on every request including redirect hops, with exact values (bypassing Go's cookie sanitizer which strips `"` from cookie values)
- Go's `net/http` cookie jar sanitizes cookie values and strips `"` characters — this breaks LinkedIn's quoted `JSESSIONID` and `bcookie` values, hence the transport-level injection
- **TLS fingerprinting**: Uses `utls` (github.com/refraction-networking/utls) to present a Chrome TLS fingerprint, bypassing Cloudflare bot detection. Without this, Go's default TLS stack is flagged and sessions are revoked.
- **HTTP/2**: Uses `golang.org/x/net/http2` transport for Chrome-fingerprinted connections, since LinkedIn requires HTTP/2 (visible as `:authority`/`:method` pseudo-headers in browser traffic).
- `client.NewWithBaseURL` accepts a custom base URL so tests can point at the mock server (uses default transport for localhost, Chrome TLS for real LinkedIn)
- Errors are typed: 401 → auth error, 429 → rate limit error, 404 → not found
- `RawGet` returns raw response bytes for endpoint exploration (`linked raw` command)

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
  "bscookie": "\"v=1&timestamp&uuid&token\"",
  "profileId": "your-profile-id",
  "createdAt": "2025-01-01T00:00:00Z",
  "fingerprint": {
    "userAgent": "Mozilla/5.0 (X11; Linux x86_64) ...",
    "secChUa": "\"Chromium\";v=\"146\", ...",
    "secChUaPlatform": "\"Linux\"",
    "xLiTrack": "{\"clientVersion\":\"1.13.42665\", ...}"
  }
}
```

The `fingerprint` field is optional — populated by `linked auth setup --har <file>`. When present, these values override the hardcoded defaults in the HTTP client. When absent, defaults matching Linux + Chrome 146 are used.

Stored at: `~/.openclaw/credentials/linkedin/<profile>/creds.json`

---

## Live API Test Results (2026-03-21)

LinkedIn has deprecated most of the old Voyager REST API (`/voyager/api/*`) and migrated to GraphQL (`/voyager/api/graphql?queryId=...`) and Dash endpoints (`/voyager/api/voyager*Dash*`, `/voyager/api/*/dash/*`). Session revocation was caused by Go's default TLS fingerprint being flagged by Cloudflare — fixed with `utls` Chrome fingerprint impersonation and HTTP/2 support.

**When LinkedIn starts rejecting sessions:** capture a fresh HAR from Chrome DevTools and run `linked auth setup --har <file>` to update the browser fingerprint stored in credentials.

### Command status

| Command | Status | Endpoint |
|---------|--------|----------|
| `auth setup` | ✅ | `--har` flag extracts browser fingerprint |
| `auth whoami` | ✅ | `/voyager/api/me` |
| `auth list` | ✅ | local only |
| `auth remove` | ✅ | local only |
| `profile get` | ✅ | self: `/me`; others: `/identity/dash/profiles?q=memberIdentity` |
| `profile skills` | ✅ | returns empty if none set |
| `profile experience list` | ✅ | returns empty if none set |
| `profile education list` | ✅ | returns empty if none set |
| `profile who-viewed` | ✅ | GraphQL `voyagerFeedDashIdentityModule` + `voyagerPremiumDashAnalyticsObject` |
| `profile contact` | ❌ 410 | deprecated, no replacement found |
| `connections list` | ✅ | `/relationships/dash/connections` + `/identity/dash/profiles` for name resolution |
| `connections pending` | ✅ | GraphQL `voyagerRelationshipsDashInvitationViews` + `InvitationsSummary` for total count |
| `connections sent` | ❌ 400 | old REST endpoint dead, needs GraphQL migration |
| `connections mutual` | ❌ 404 | old REST endpoint dead |
| `messages list` | ✅ | GraphQL `voyagerMessagingDashMessengerConversations` (find-conversations-by-category) |
| `messages unread` | ✅ | same endpoint, filtered client-side |
| `messages read` | ✅ | GraphQL `voyagerMessagingDashMessengerMessages` (get-messages-by-conversation) |
| `messages send` | ⏸️ stubbed | old REST dead, new write protocol unknown |
| `messages mark-read` | ⏸️ stubbed | new write protocol unknown |
| `messages star/unstar` | ⏸️ stubbed | new write protocol unknown |
| `messages archive/unarchive` | ⏸️ stubbed | new write protocol unknown |
| `messages delete*` | ⏸️ stubbed | new write protocol unknown |
| `posts feed` | ✅ | `/feed/updatesV2?q=feed` (normalized format, actor resolution from header + actor fields) |
| `posts get/create/edit/delete` | ❌ 404 | `/ugcPosts` dead |
| `posts like/unlike/comment` | ❌ 404 | `/socialActions` dead |
| `posts activity` | ❌ 400 | needs new query param |
| `jobs get` | ✅ | GraphQL `voyagerJobsDashJobPostings` (fetch-full-job-posting) |
| `jobs recommended` | ✅ | GraphQL `voyagerJobsDashJobsFeed` (full-jobs-feed-get-all) |
| `jobs saved` | ✅ | GraphQL `voyagerJobsDashJobCards` (job-cards-by-job-search-v2, slug=saved) |
| `jobs applied` | ✅ | GraphQL `voyagerJobsDashJobCards` (slug=applied) |
| `jobs company` | ❌ 404 | old search endpoint dead |
| `companies get` | ✅ | `/organization/companies?q=universalName` (normalized response) |
| `companies posts` | ❌ 400 | needs new endpoint |
| `companies employees` | ❌ 404 | uses dead search endpoint |
| `search people` | ❌ 404 | `/search/hits` deprecated, needs GraphQL migration |
| `search jobs` | ❌ 404 | `/jobs/search` dead |
| `search companies` | ❌ 404 | `/search/hits` deprecated |
| `search posts` | ❌ 404 | `/search/blended` dead |
| `notifications list` | ✅ | `/voyagerIdentityDashNotificationCards` (normalized format) |
| `notifications count` | ✅ | `/voyagerNotificationsDashBadgingItemCounts` |
| `notifications mark-read` | ⏸️ stubbed | new endpoint unknown |
| `recommendations *` | ❌ 404 | `/identity/recommendations` dead |
| `raw` | ✅ | debug command for endpoint exploration |

### GraphQL queryId reference

| queryId | Purpose |
|---------|---------|
| `voyagerFeedDashIdentityModule.803fe19f843a4d461478049f70d7babd` | Who-viewed count (feed widget) |
| `voyagerPremiumDashAnalyticsObject.faf9c8e3233e83980f323f07c637b3c3` | Who-viewed details (Premium) |
| `voyagerRelationshipsDashInvitationViews.57e1286f887065b96393b947e09ef04c` | Pending invitations |
| `voyagerRelationshipsDashInvitationsSummary.26002c38d857d2d5cd4503df1a43a0ab` | Invitation count summary |
| `voyagerRelationshipsDashSentInvitationViews.1901307baa315a33bf17bb743daf1250` | Sent invitations |
| `voyagerMessagingDashMessengerConversations.ccc086e11ebcecef63b31ac465ccfebd` | Conversation list (by category) |
| `voyagerMessagingDashMessengerMessages.073958b6fdfe5f5ceeb4d0416523317e` | Messages in conversation |
| `voyagerMessagingDashMessengerMailboxCounts.15769ef365ec721fc539d76dbef5f813` | Mailbox unread counts |
| `voyagerJobsDashJobsFeed.40bc6ea7c5b88757481d40f6e4527f17` | Recommended jobs feed |
| `voyagerJobsDashJobCards.7fb7b035d6233f835789e4088cdbf44b` | Job cards (saved/applied collections) |
| `voyagerJobsDashJobPostings.891aed7916d7453a37e4bbf5f1f60de4` | Single job posting detail |
