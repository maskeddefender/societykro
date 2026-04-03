# Phase 1.1 — Project Scaffolding & Foundation

**Duration:** Week 1-2
**Goal:** Set up the complete monorepo structure, all configuration files, Docker compose for local dev, database migrations, and a working "hello world" across all layers.

**At the end of Phase 1.1, you should be able to:**
1. Run `make dev` and have all services start
2. Hit `http://localhost:8080/health` on the API gateway
3. Run the React Native app on a phone via Expo Go
4. Connect to PostgreSQL and see empty tables with correct schema
5. Run `make test` and see green tests

---

## Phase 1.1 Tasks

### 1. Repository Setup
- [x] Create monorepo at `/home/walnut/societykro`
- [ ] Initialize git
- [ ] Create `.gitignore`
- [ ] Create `README.md`
- [ ] Create `Makefile` with common commands
- [ ] Create `.editorconfig`

### 2. Backend Foundation (Go)
- [ ] Initialize Go workspace (`go.work`)
- [ ] Create `services/auth-service/` scaffold
- [ ] Create `services/complaint-service/` scaffold
- [ ] Create `services/visitor-service/` scaffold
- [ ] Create `services/notification-service/` scaffold
- [ ] Create `services/message-router/` scaffold
- [ ] Create `packages/go-common/` (shared Go utilities)
- [ ] Create `packages/proto/` (gRPC protobuf definitions)

### 3. Mobile App Foundation (React Native)
- [ ] Initialize Expo project at `apps/mobile/`
- [ ] Configure Expo Router (file-based routing)
- [ ] Set up Zustand stores skeleton
- [ ] Set up i18next with Hindi + English
- [ ] Set up React Query provider
- [ ] Create basic tab navigation (Home, Complaints, Visitors, Payments, More)

### 4. Web Admin Foundation
- [ ] Initialize Next.js at `apps/web-admin/`
- [ ] Set up Tailwind + Shadcn/ui
- [ ] Create login page skeleton
- [ ] Create dashboard layout skeleton

### 5. Database
- [ ] Create `db/migrations/` with all SQL from PRD
- [ ] Create `db/seeds/` with test data
- [ ] Create `db/queries/` for sqlc

### 6. Infrastructure (Local Dev)
- [ ] Create `docker-compose.yml` (PostgreSQL, Redis, NATS, MinIO)
- [ ] Create Dockerfiles for each Go service
- [ ] Create `.env.example`

### 7. CI/CD Foundation
- [ ] Create `.github/workflows/ci.yml`
- [ ] Create `.github/workflows/mobile-build.yml`

---

## Phase 1.1 Deliverables

| Deliverable | Status |
|-------------|--------|
| Monorepo with all folders created | Pending |
| All Go services scaffolded with health endpoint | Pending |
| React Native app running with tab navigation | Pending |
| Next.js admin with login page | Pending |
| Docker compose up = all DBs running | Pending |
| Database migrations applied | Pending |
| CI pipeline running on push | Pending |
| Makefile with dev/test/build/migrate commands | Pending |

---

## What Phase 1.2 Covers (next)

- Auth service: OTP send/verify, JWT tokens, user registration
- Society/flat CRUD
- API gateway (Kong) configuration
- Mobile app: Login screen, OTP verification flow
