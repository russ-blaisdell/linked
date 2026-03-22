#!/usr/bin/env python3
"""Send a LinkedIn message via the messenger API.

Usage: messenger_send.py <creds_json_path> <conversation_urn> <body_text>

Exits 0 on success, 1 on failure. Prints the response JSON to stdout.
Uses Python's http.client which has been verified to work with LinkedIn's
messenger endpoint without triggering session revocation.
"""
import http.client
import json
import os
import sys
import uuid


def main():
    if len(sys.argv) < 4:
        print("Usage: messenger_send.py <creds_path> <conversation_urn> <body_text>", file=sys.stderr)
        sys.exit(1)

    creds_path = sys.argv[1]
    conversation_urn = sys.argv[2]
    body_text = sys.argv[3]

    with open(creds_path) as f:
        creds = json.load(f)

    csrf = creds["jsessionid"].strip('"')

    # Find the mailbox URN (fsd_profile) - we need it from /me
    # but to avoid an extra call, derive it from the conversation URN
    # which contains: urn:li:msg_conversation:(urn:li:fsd_profile:XXX,thread-id)
    mailbox_urn = ""
    if "fsd_profile:" in conversation_urn:
        start = conversation_urn.index("urn:li:fsd_profile:")
        end = conversation_urn.index(",", start)
        mailbox_urn = conversation_urn[start:end]

    if not mailbox_urn:
        print("Could not extract mailbox URN from conversation URN", file=sys.stderr)
        sys.exit(1)

    random_bytes = os.urandom(16)
    origin_token = str(uuid.uuid4())

    payload = {
        "message": {
            "body": {"attributes": [], "text": body_text},
            "renderContentUnions": [],
            "conversationUrn": conversation_urn,
            "originToken": origin_token,
        },
        "mailboxUrn": mailbox_urn,
        "trackingId": random_bytes.decode("latin-1"),
        "dedupeByClientGeneratedToken": False,
    }

    body = json.dumps(payload, ensure_ascii=False)

    conn = http.client.HTTPSConnection("www.linkedin.com")
    headers = {
        "accept": "application/json",
        "content-type": "text/plain;charset=UTF-8",
        "csrf-token": csrf,
        "x-restli-protocol-version": "2.0.0",
        "cookie": (
            f'li_at={creds["li_at"]}; '
            f'JSESSIONID={creds["jsessionid"]}; '
            f'bcookie={creds["bcookie"]}; '
            f'bscookie={creds.get("bscookie", "")}'
        ),
    }

    conn.request(
        "POST",
        "/voyager/api/voyagerMessagingDashMessengerMessages?action=createMessage",
        body.encode("utf-8"),
        headers,
    )
    resp = conn.getresponse()
    data = resp.read().decode()

    if resp.status == 200:
        print(data)
        sys.exit(0)
    else:
        print(f"HTTP {resp.status}: {data}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
