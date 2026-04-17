# Phase 2.1 — Next.js Web Admin Dashboard

**Duration:** 3-4 days
**Depends On:** Phase 1.2 (complete)
**Goal:** A fully functional admin dashboard for society committee members to manage complaints, payments, notices, members, and vendors.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | Next.js 14+ (App Router) |
| UI Components | Shadcn/ui (copy-paste, no dependency) |
| Styling | Tailwind CSS v3 |
| Charts | Recharts |
| Tables | TanStack Table v8 |
| State | Zustand (same as mobile) |
| Auth | JWT (same tokens as mobile, cookie-based) |
| HTTP | Axios with JWT interceptor |

---

## Screens

| # | Screen | Route | Description |
|---|--------|-------|-------------|
| 1 | Login | `/login` | Phone + OTP (same auth-service) |
| 2 | Dashboard | `/` | Stats cards, charts, recent activity |
| 3 | Complaints | `/complaints` | Table with filters, assign vendor, update status |
| 4 | Complaint Detail | `/complaints/[id]` | Full detail, comments, status history |
| 5 | Payments | `/payments` | Collection report, defaulters, generate bills |
| 6 | Notices | `/notices` | Create, list, read receipts |
| 7 | Visitors | `/visitors` | Today's log, history, blacklist |
| 8 | Members | `/members` | List, roles, add/remove |
| 9 | Vendors | `/vendors` | CRUD, ratings, job history |
| 10 | Settings | `/settings` | Society config, maintenance amount, late fees |

---

## Deliverables

- [ ] Initialize Next.js project at `apps/web-admin/`
- [ ] Tailwind + Shadcn/ui setup
- [ ] Auth: login page, JWT cookie, auth middleware
- [ ] Sidebar layout with navigation
- [ ] Dashboard with stats + charts
- [ ] Complaints management (table + detail + actions)
- [ ] Payment management (bills + defaulters + reports)
- [ ] Notice management (create + list + read stats)
- [ ] Member management (list + role change)
- [ ] Vendor management (CRUD)
- [ ] Society settings page
