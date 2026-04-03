# SocietyKro

Ultra-lite society management platform for Indian housing societies. Target: 10M+ users.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Mobile | React Native (Expo) + Hermes |
| Web Admin | Next.js + Shadcn/ui + Tailwind |
| Backend | Go (Golang) microservices |
| Database | PostgreSQL + Redis + ScyllaDB (Phase 2) |
| Queue | NATS (Phase 1) -> Kafka (Phase 3) |
| Voice/Language | Bhashini API (22 Indian languages) |
| Chatbot | Rule-based + Gemma 2 9B |
| WhatsApp | Meta Cloud API |
| Telegram | Telegram Bot API |

## Quick Start

```bash
# 1. Clone and setup
git clone <repo-url>
cd societykro
make setup

# 2. Start databases
make docker-up

# 3. Run migrations
make migrate

# 4. Seed test data
make seed

# 5. Start development
make dev
```

## Project Structure

```
societykro/
  apps/mobile/          React Native (Expo) mobile app
  apps/web-admin/       Next.js admin dashboard
  apps/guard-app/       Simplified guard interface
  services/             Go microservices
  packages/             Shared code (proto, go-common, types)
  infra/                Terraform, K8s, Docker
  db/                   Migrations, seeds, SQL queries
  docs/                 PRDs and documentation
  scripts/              Build and utility scripts
```

## Available Commands

Run `make help` to see all commands.
