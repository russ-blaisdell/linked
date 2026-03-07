# Authentication

`linked` authenticates using your LinkedIn session cookies. This is the same mechanism the LinkedIn website uses — no developer account or API key is required.

## What you need

Two cookies from your browser session on linkedin.com:

| Cookie | Example value | Purpose |
|--------|--------------|---------|
| `li_at` | `AQEDARxxxxxxx...` | Your main session token |
| `JSESSIONID` | `ajax:1234567890abcdef` | CSRF validation |

## Getting your cookies

### Chrome / Edge
1. Open [linkedin.com](https://www.linkedin.com) and log in
2. Open DevTools (`⌘ Opt I` on Mac, `F12` on Windows)
3. Go to **Application** → **Cookies** → `https://www.linkedin.com`
4. Find `li_at` — copy its **Value**
5. Find `JSESSIONID` — copy its **Value** (it starts with `ajax:`)

### Firefox
1. Open [linkedin.com](https://www.linkedin.com) and log in
2. Open DevTools (`⌘ Opt I` / `F12`)
3. Go to **Storage** → **Cookies** → `https://www.linkedin.com`
4. Copy the values for `li_at` and `JSESSIONID`

### Safari
1. Enable the Develop menu: Safari → Preferences → Advanced → Show Develop menu
2. Open [linkedin.com](https://www.linkedin.com) and log in
3. Develop → Show Web Inspector → Storage → Cookies
4. Copy the values for `li_at` and `JSESSIONID`

## Storing credentials

Run the setup wizard:

```
linked auth setup
```

You will be prompted for each value. Credentials are stored at:

```
~/.openclaw/credentials/linkedin/default/creds.json
```

The file is created with mode `0600` (readable only by you).

## Verifying

```
linked auth whoami
```

This confirms your credentials work and shows which account is authenticated.

## Multiple accounts

Use `--profile` to manage separate credentials:

```
linked auth setup --profile work
linked auth setup --profile personal

linked --profile work messages list
linked --profile personal profile get
```

## Cookie expiry

`li_at` cookies typically last 1 year. If you see a 401 error, your session may have expired — log out and back in to LinkedIn, then re-run `linked auth setup`.

## Security

- Cookies are stored in `~/.openclaw/credentials/linkedin/` with `0600` permissions
- Never share your `li_at` value — it grants full access to your LinkedIn account
- The credential file is excluded from git via `.gitignore`

## Terms of Service note

This tool uses LinkedIn's internal Voyager API (the same API the LinkedIn website uses). This is not an officially supported developer integration. Use it for personal automation only, and avoid bulk or automated operations that could be considered spam.
