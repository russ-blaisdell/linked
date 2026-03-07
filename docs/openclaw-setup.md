# OpenClaw Integration Guide

## Prerequisites

- OpenClaw installed and running (`openclaw doctor`)
- `linked` built and installed (`make install`)
- LinkedIn credentials configured (`linked auth setup`)

## Install the skill

```bash
make skill
```

This copies `skill/linkedin/skill.md` to `~/.openclaw/workspace/skills/linkedin/skill.md`.

Verify it appears:

```bash
openclaw skills list | grep linkedin
```

## Restart the gateway

```bash
openclaw gateway
```

## Test it

Send a message to your OpenClaw agent (via Slack or any connected channel):

```
@openclaw search LinkedIn for senior go engineer remote jobs
```

```
@openclaw show my LinkedIn unread messages
```

```
@openclaw look up the Anthropic company on LinkedIn
```

```
@openclaw who viewed my LinkedIn profile this week?
```

```
@openclaw what are my LinkedIn job recommendations?
```

```
@openclaw update my LinkedIn headline to "Staff Engineer at Acme"
```

```
@openclaw how many unread LinkedIn notifications do I have?
```

## How it works

When you ask OpenClaw about LinkedIn, the agent reads the skill definition at `~/.openclaw/workspace/skills/linkedin/skill.md` to understand what commands are available and when to use them. It then calls `linked` as a subprocess, capturing JSON output to build its response.

The agent uses `linked --output json` for machine-readable data. All commands support this flag.

## Multiple profiles

If you have multiple LinkedIn accounts, configure them with:

```bash
linked auth setup --profile work
linked auth setup --profile personal
```

Tell the agent which profile to use in your message:

```
@openclaw using my work LinkedIn profile, check my messages
```

Or set a default in the skill file by editing `skill.md` to add `--profile work` to the default command patterns.

## Troubleshooting

**`linked: command not found`**
Run `make install` to install the binary to `/usr/local/bin`.

**`401 Unauthorized`**
Your LinkedIn session cookies have expired. Run `linked auth setup` again with fresh cookies from your browser.

**`429 Rate Limited`**
LinkedIn rate limits requests. Wait a few minutes and try again. Avoid running many commands in quick succession.

**Skill not recognized by OpenClaw**
Check that the skill file is in the right location:
```bash
ls ~/.openclaw/workspace/skills/linkedin/skill.md
```
Then restart the gateway: `openclaw gateway`
