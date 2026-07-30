<div align="center">

# 📡 tiny-smtp-client

**A raw-socket SMTP client, written from scratch in Go.**
No `net/smtp`. No mail libraries. Just TCP, TLS, and the protocol — implemented by hand.

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white">
  <img alt="Protocol" src="https://img.shields.io/badge/Protocol-SMTP-blue">
  <img alt="Transport" src="https://img.shields.io/badge/Transport-TCP-success">
  <img alt="Security" src="https://img.shields.io/badge/Security-STARTTLS-green">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-yellow">
</p>

</div>

---

## 🎬 See it in action

https://github.com/user-attachments/assets/75da3634-9df0-480f-8f81-c9bb841c4fed

*A live run: TCP connect → STARTTLS upgrade → AUTH LOGIN → MIME message with attachment → delivered.*

---

## Why This Project Exists

Most applications send email through high-level libraries or third-party services. While convenient, those abstractions hide the protocol conversation that actually delivers an email.

This project takes the opposite approach.

Instead of relying on Go's SMTP libraries, it opens a raw TCP connection to a real SMTP server and implements the SMTP conversation manually. It negotiates `STARTTLS`, authenticates with `AUTH LOGIN`, constructs a MIME message, and sends an email with a binary attachment.

The goal is not to build a production-ready mail client, but to understand SMTP by implementing the protocol directly instead of treating it as a black box.

---

## Getting started

**1. Clone the repository**

```bash
git clone git@github.com:AhmadHussainRandhawa/tiny-smtp-client.git
cd tiny-smtp-client
```

**2. Get a Gmail App Password** (not your account password) — Google Account → Security → App Passwords.

**3. Set environment variables:**

```bash
export SMTP_EMAIL="your.email@gmail.com"
export SMTP_PASSWORD="your-16-character-app-password"
export SMTP_RCPT="recipient@example.com"
```

**4. Place an attachment** at `attachments/image.png` (or edit `attachmentPath` in `main.go`).

**5. Run it:**

```bash
go run main.go
```

Every line of the live SMTP conversation prints to stdout as it happens — you're watching the diagram above execute in real time.

---

## How it works

```
┌──────────────┐         plaintext            ┌──────────────────┐
│  This client │ ───────────────────────────▶ │  smtp.gmail.com  │
│              │ ◀─────────────────────────── │     :587         │
└──────────────┘         EHLO / STARTTLS      └──────────────────┘
       │
       │  TLS handshake (tls.Client)
       ▼
┌──────────────┐         encrypted            ┌──────────────────┐
│  This client │ ───────────────────────────▶ │  smtp.gmail.com  │
│              │ ◀─────────────────────────── │  (same socket)   │
└──────────────┘   EHLO / AUTH LOGIN / DATA   └──────────────────┘
```

1. **Plaintext handshake** — connect on port 587 and send `EHLO`.
2. **STARTTLS** — confirm the server advertises `STARTTLS`, request it, then upgrade the *existing* TCP connection to TLS in place (`tls.Client` wraps the same socket — no new connection is opened).
3. **Re-negotiate EHLO** — required after the TLS upgrade, since capabilities are re-advertised over the encrypted channel.
4. **AUTH LOGIN** — authenticate using a base64-encoded username, then password, exchanged as separate challenge/response steps.
5. **Envelope** — `MAIL FROM` and `RCPT TO` establish the SMTP envelope (distinct from the `From:`/`To:` *headers* inside the message body).
6. **DATA** — stream a hand-built MIME multipart message: a `text/plain` part and a base64-encoded `image/png` attachment, separated by a boundary string, terminated with the `<CRLF>.<CRLF>` sequence.
7. **QUIT** — close out the session cleanly.

Every response from the server is read and printed live, so the full protocol conversation is visible while the program runs — this is also what makes the demo video worth watching.

---  

## 📡 SMTP Session

The client establishes a TCP connection, upgrades it to TLS, authenticates, submits a MIME email, and gracefully closes the SMTP session.

```text
┌───────────────┐                                      ┌────────────────┐
│ SMTP Client   │                                      │ SMTP Server    │
└───────────────┘                                      └────────────────┘

      │                                                           │
      │ TCP Connect (587)                                         │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │                   220 Service Ready                       │
      │                                                           │
      │ EHLO localhost                                            │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │              250 Server Capabilities                      │
      │                                                           │
      │ STARTTLS                                                  │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │                 220 Ready to Start TLS                    │
      │                                                           │
      │═══════════════════ TLS Handshake ═════════════════════════│
      │                                                           │
      │ EHLO localhost (again)                                    │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │              250 Server Capabilities                      │
      │                                                           │
      │ AUTH LOGIN                                                │
      ├─────────────────────────────────────────────────────────► │
      │  ◄────────────────────────────────────────────────────────┤
      │                334 Username Challenge                     │
      │ Base64(Email)                                             │
      ├─────────────────────────────────────────────────────────► │
      │  ◄────────────────────────────────────────────────────────┤
      │                334 Password Challenge                     │
      │ Base64(App Password)                                      │
      ├─────────────────────────────────────────────────────────► │
      │  ◄────────────────────────────────────────────────────────┤
      │               235 Authentication Successful               │
      │                                                           │
      │ MAIL FROM                                                 │
      ├─────────────────────────────────────────────────────────► │
      │ RCPT TO                                                   │
      ├─────────────────────────────────────────────────────────► │
      │ DATA                                                      │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │                  354 Start Mail Input                     │
      │                                                           │
      │ MIME Message + Attachment                                 │
      ├─────────────────────────────────────────────────────────► │
      │ <CRLF>.<CRLF>                                             │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │                 250 Message Accepted                      │
      │                                                           │
      │ QUIT                                                      │
      ├─────────────────────────────────────────────────────────► │
      │ ◄─────────────────────────────────────────────────────────┤
      │                      221 Goodbye                          │
      │                                                           │

```

---

## What's implemented

| Capability | Status | Notes |
|---|:---:|---|
| Raw TCP connection to SMTP server | ✅ | No standard library SMTP client used |
| Multiline response parsing | ✅ | Distinguishes `250-` continuation from `250 ` terminator |
| `STARTTLS` detection + upgrade | ✅ | Upgrades an existing plaintext connection mid-session |
| Post-TLS re-handshake | ✅ | Re-issues `EHLO` after upgrade — capabilities pre-TLS can't be trusted |
| `AUTH LOGIN` | ✅ | Base64-encoded credential exchange |
| MIME `multipart/mixed` construction | ✅ | Hand-built boundaries, headers, and encoding |
| Base64 file attachment | ✅ | Binary → base64 → embedded in message body |
| Proper `DATA` termination | ✅ | `<CRLF>.<CRLF>` end-of-message sequence |

## What's intentionally out of scope

| Not implemented | Why |
|---|---|
| Full RFC 5321 compliance (pipelining, `BDAT`, `DSN`) | Diminishing returns for a learning project |
| Multiple auth mechanisms (`PLAIN`, `XOAUTH2`) | `AUTH LOGIN` alone proves the concept |
| Retry / connection pooling / queuing | That's a mail *relay*, a different project |
| Support for arbitrary SMTP providers | Handshake is written and tested against Gmail specifically |
| Error recovery mid-transaction | Server errors are printed, not handled — a deliberate line, not a bug |

> These are boundaries, not gaps. Knowing where to stop is part of the exercise.

---

## Security

- Credentials are read from environment variables only — never hardcoded, never committed.
- Use a Gmail **App Password**, scoped and independently revocable — not your primary password.
- `.gitignore` excludes any local `.env` or credential file by default.

---

## Part of a series

This is the first of three projects built to lock in a deep dive on email systems — from sending, to routing, to receiving:

| Project | What it proves | Status |
|---|---|:---:|
| **[tiny-smtp-client](https://github.com/AhmadHussainRandhawa/tiny-smtp-client)** *(this repo)* | Sending mail, from a raw socket up | ✅ Complete |
| **[mx-lookup-go](https://github.com/AhmadHussainRandhawa/mx-lookup-tool)** | DNS resolution of mail routing (MX + A records) | ✅ Complete |
| **[smtp-server-go](https://github.com/AhmadHussainRandhawa/tiny-smtp-server)** | Receiving mail — including malformed input | ✅ Complete |

---

<div align="center">

**License:** [MIT](./LICENSE)

</div>
