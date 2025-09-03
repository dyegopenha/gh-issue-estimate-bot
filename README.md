# GitHub Issue Estimate Bot (Go)

A minimal GitHub App that listens for newly created issues and reminds the author to include an estimate in the form:

```
Estimate: X days
```

If the estimate is **missing**, the app posts a concise comment asking the author to add one.

## Features
- GitHub App (not a PAT): authenticates with an **installation access token**.
- Listens to `issues` webhooks, action `opened`.
- Detects estimate string with a strict but forgiving regex (case-insensitive; accepts pluralization and decimals).
- Idempotent comments: won’t duplicate reminders thanks to a hidden marker in the comment.
- Verifies `X-Hub-Signature-256` with your webhook secret.
- Container-ready (Dockerfile).
- Works locally with **smee.io** for webhook tunneling.

---

## 1) Create the GitHub App

1. Go to **GitHub › Settings › Developer settings › GitHub Apps › New GitHub App**.
2. **GitHub App name**: `issue-estimate-bot` (or anything).
3. **Webhook**:
   - **Webhook URL**: use your **smee.io** channel URL for now (you can change later).
   - **Webhook secret**: set a strong secret; you’ll also put this in `.env`.
4. **Permissions** (Repository):
   - **Issues**: **Read & write** (needed to post comments).
   - **Metadata**: **Read-only**.
5. **Subscribe to events**:
   - ✅ `Issues`
6. Click **Create GitHub App**.
7. On the app page:
   - **Generate a private key** → download `private-key.pem` and place it at `./configs/private-key.pem` (or update the path in `.env`).
   - Click **Install App** → install it on the **target repository** where you want this to run.

> Tip: You can set the app’s **Webhook URL** to your smee channel (e.g., `https://smee.io/your-random-id`) during development, and later point it to your public server in production.

---

## 2) Local development with smee.io

- Visit <https://smee.io> and **Start a new channel**. Copy the channel URL.
- In a separate terminal, run the smee client to forward webhooks to your local server:
  ```bash
  npx smee-client --url https://smee.io/YOUR-CHANNEL --target http://localhost:3000/webhook
  ```

---

## 3) Configure the app

Copy the example env and fill in values:

```bash
cp .env.example .env
# edit .env with your APP_ID, WEBHOOK_SECRET, PRIVATE_KEY_PATH
```

- `APP_ID`: Found on your GitHub App page (integer).
- `WEBHOOK_SECRET`: The exact secret you set in the app.
- `PRIVATE_KEY_PATH`: Path to the downloaded `.pem` private key (default `./configs/private-key.pem`).
- `PORT`: Defaults to `3000`.

Place your private key at the path in `.env` (don’t commit it).

---

## 4) Run

### With Go
```bash
go mod tidy
go run ./cmd/server
```

### With Docker
```bash
docker build -t issue-estimate-bot:local .
docker run --rm -p 3000:3000 --env-file .env -v $PWD/configs:/app/configs issue-estimate-bot:local
```

Now create a new issue in a repo where the app is installed. If the body does **not** contain an estimate like `Estimate: 2 days`, the bot will add a reminder comment.

---

## 5) Production notes

- Host behind HTTPS and a firewall (e.g., on Fly.io, Render, Heroku, EC2 + ALB).
- Set the GitHub App’s **Webhook URL** to your public `/webhook` endpoint.
- Rotate the **private key** periodically and keep it off logs, images, and repos.
- Keep the **webhook secret** in a secure store (e.g., AWS Secrets Manager).
- Enforce TLS 1.2+; prefer minimal surface area.
- Use health checks on `/healthz` and structured logs.

---

## 6) How the detection works

We check for the pattern (case-insensitive):

```
\bEstimate:\s*\d+(\.\d+)?\s*days?\b
```

Examples that **match**:
- `Estimate: 2 days`
- `Estimate: 1 day`
- `Estimate: 2.5 days`

Examples that **do not match**:
- `ETA 2d`
- `Estimate ~2 days` (no colon; use exact `Estimate:`)

---

## 7) Repo structure

```
.
├── cmd
│   └── server
│       └── main.go
├── internal
│   ├── estimate
│   │   ├── estimate.go
│   │   └── estimate_test.go
│   ├── githubapp
│   │   └── client.go
│   └── webhook
│       └── handler.go
├── configs
│   └── private-key.pem        # (you provide; do not commit)
├── .env.example
├── Dockerfile
└── README.md
```

---

## Troubleshooting

- **401/403** when posting comments: Ensure the app is installed on the target repo and has `Issues: Read & write`.
- **Signature verification failed**: Confirm `WEBHOOK_SECRET` matches exactly and smee is forwarding to `/webhook`.
- **No comment posted**: Check logs; confirm the issue body truly lacks `Estimate: X days`.
