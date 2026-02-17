# PicoHub

Secure skill marketplace for [PicoClaw](https://github.com/yu01/picoclaw) - AI agents on $10 RISC-V boards.

Built with security-first design, learning from OpenClaw/ClawHub's malware issues. Every skill upload is validated and scanned.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go + chi router + SQLite (FTS5) |
| Frontend | Next.js 14 (App Router) + Tailwind CSS |
| Skills | Python (SKILL.md + manifest.json) |
| Auth | JWT (HMAC-SHA256) + bcrypt |
| Scanning | ClamAV interface (Noop scanner in dev) |

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Install & Run

```bash
# Install dependencies
make install

# Start development servers (backend:8080 + frontend:3000)
make dev
```

Or run separately:

```bash
# Backend (port 8080)
cd backend && CGO_ENABLED=1 go run .

# Frontend (port 3000)
cd frontend && npm run dev
```

Then open http://localhost:3000

### Demo Accounts

| Email | Password | Role |
|-------|----------|------|
| admin@picohub.dev | admin123 | Admin |
| tanaka@example.com | password123 | User |
| suzuki@example.com | password123 | User |

## API Endpoints

```
POST   /api/v1/auth/register     - Register
POST   /api/v1/auth/login        - Login (returns JWT)
GET    /api/v1/auth/me           - Profile [auth]

GET    /api/v1/skills            - List skills (?q=&category=&sort=)
GET    /api/v1/skills/featured   - Featured skills
GET    /api/v1/skills/categories - Categories
GET    /api/v1/skills/{slug}     - Skill detail
POST   /api/v1/skills            - Upload skill [auth]
GET    /api/v1/skills/{slug}/download - Download

GET    /api/v1/skills/{slug}/reviews  - List reviews
POST   /api/v1/skills/{slug}/reviews  - Create review [auth]

GET    /api/v1/health            - Health check
```

## Security

- **Passwords**: bcrypt (cost 12)
- **JWT**: HMAC-SHA256, 24h expiry
- **Uploads**: 10MB limit, ZIP validation, manifest.json required, symlink detection
- **Scanning**: ClamAV interface (pluggable)
- **Rate limiting**: Login 5/min, Register 3/hour, API 100/min
- **SQL**: Parameterized queries only

## Project Structure

```
PicoHub/
├── backend/             # Go API server (port 8080)
│   ├── main.go          # Entry point + routing
│   └── internal/
│       ├── config/      # Environment config
│       ├── database/    # SQLite + migrations + seed
│       ├── handler/     # HTTP handlers
│       ├── middleware/   # Auth, CORS, rate limit, logger
│       ├── model/       # Data models
│       ├── repository/  # Database operations
│       ├── scanner/     # ClamAV interface
│       └── service/     # Business logic
├── frontend/            # Next.js web app (port 3000)
│   └── src/
│       ├── app/         # Pages (/, /skills, /upload, /auth/*)
│       ├── components/  # UI components
│       ├── hooks/       # React hooks
│       ├── lib/         # API client
│       └── types/       # TypeScript types
├── skills/              # 5 sample skills
│   ├── line-messenger/
│   ├── rakuten-shopping/
│   ├── weather-reminder/
│   ├── mercari-lister/
│   └── notion-lite/
├── Makefile
└── LICENSE              # MIT
```

## Sample Skills

| Skill | Category | Description |
|-------|----------|-------------|
| LINE Messenger | messaging | LINE message send/receive via Messaging API |
| Rakuten Shopping | shopping | Product search & price comparison |
| Weather Reminder | utility | Forecast + umbrella/laundry/heatstroke alerts |
| Mercari Lister | commerce | Auto-generate listing text for Mercari |
| Notion Lite | productivity | Lightweight Notion API integration |

## License

MIT
