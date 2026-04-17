# Phase 1.2 — Authentication, Society Management & API Gateway

**Duration:** Week 3-4
**Depends On:** Phase 1.1 (complete)
**Goal:** Working auth flow (OTP → JWT), society/flat CRUD, role-based middleware, and API gateway in front of services.

---

## Deliverables

### 1. Proper JWT Authentication
- [ ] Generate RSA key pair (`keys/private.pem`, `keys/public.pem`)
- [ ] Implement RS256 JWT token signing in auth-service
- [ ] Access token: 15-minute expiry, contains `user_id`, `society_id`, `role`
- [ ] Refresh token: 30-day expiry, stored in Redis, single-use rotation
- [ ] Token refresh endpoint (replace placeholder)
- [ ] Logout endpoint (blacklist access token, delete refresh token from Redis)

### 2. Auth Middleware (`packages/go-common/middleware/`)
- [ ] `JWTMiddleware` — Validates JWT on every request, injects `user_id` and `role` into Fiber context
- [ ] `RequireRole(roles ...string)` — Checks if user has required role for the endpoint
- [ ] `RequireSociety()` — Ensures user belongs to the society being accessed (prevents cross-society data access)
- [ ] Apply middleware to all routes except `/health`, `/auth/otp/*`, `/auth/refresh`

### 3. SMS OTP Integration
- [ ] MSG91 API integration for sending OTP in production
- [ ] DLT template registration (TRAI compliance)
- [ ] Environment-based: dev = log OTP + accept `000000`, prod = real SMS
- [ ] OTP rate limiting hardened: 3 attempts/phone/15min, CAPTCHA after 2 fails (future)

### 4. Society Management (Complete)
- [ ] Create society → auto-generate flat grid based on blocks × floors × flats-per-floor
- [ ] Update society settings (maintenance amount, due day, late fee, language)
- [ ] Invite members: generate shareable invite link with society code
- [ ] Admin can change member roles (resident → admin, etc.)
- [ ] Admin can remove members
- [ ] Society stats endpoint: total flats, occupied, complaints open, payments pending

### 5. User Profile (Complete)
- [ ] Update profile: name, email, language, senior citizen mode
- [ ] Update FCM token (for push notifications)
- [ ] Get profile with society memberships
- [ ] Account deletion request (DPDPA compliance, mark for 30-day erasure)

### 6. API Gateway (Kong or Traefik)
- [ ] Set up Kong OSS (or Traefik) in docker-compose
- [ ] Route `/api/v1/auth/*` → auth-service:8081
- [ ] Route `/api/v1/complaints/*` → complaint-service:8082 (placeholder)
- [ ] Route `/api/v1/visitors/*` → visitor-service:8083 (placeholder)
- [ ] Rate limiting plugin: 100 req/min per user
- [ ] CORS configuration
- [ ] Request logging
- [ ] JWT validation at gateway level (optional, or per-service)

### 7. Complaint Service — Scaffold
- [ ] Scaffold complaint-service with same structure as auth-service
- [ ] Implement: Create complaint (text), List complaints, Get complaint detail
- [ ] Complaint status updates (open → in_progress → resolved → closed)
- [ ] Connect to NATS: publish `complaint.created` event
- [ ] Voice complaint endpoint (placeholder — calls voice-service in Phase 1.3)

### 8. Notice Service — Scaffold
- [ ] Scaffold notice-service
- [ ] Implement: Create notice, List notices (pinned first), Get notice with read count
- [ ] Mark as read endpoint
- [ ] Connect to NATS: publish `notice.posted` event

### 9. CI/CD Foundation
- [ ] `.github/workflows/ci.yml` — On push: lint + test for all Go services
- [ ] `.github/workflows/mobile-build.yml` — Placeholder for Expo EAS build

### 10. React Native App — Login Flow
- [ ] Initialize Expo project at `apps/mobile/`
- [ ] Set up Expo Router with tab navigation
- [ ] Login screen: phone input → OTP input → home
- [ ] Store JWT in react-native-mmkv
- [ ] Axios instance with auth token injection
- [ ] Home screen: placeholder with society name + member info

---

## Architecture After Phase 1.2

```
[React Native App]
        |
        v
[API Gateway (Kong)] :8080
   |         |          |
   v         v          v
[auth]    [complaint] [notice]
:8081     :8082       :8085
   |         |          |
   +---------+----------+--------> [PostgreSQL] :5432
   +---------+----------+--------> [Redis] :6379
             |          |
             +----------+--------> [NATS] :4222
```

---

## Success Criteria

At the end of Phase 1.2:
1. User can login via OTP on the React Native app
2. JWT tokens are properly signed and validated
3. Society CRUD works with role-based access control
4. Complaint can be raised (text + photo) and tracked
5. Notices can be posted and read
6. API gateway routes traffic to correct services
7. CI pipeline runs lint + test on every push

---

## Estimated Effort

| Task | Days |
|------|------|
| JWT + Auth middleware | 2 |
| Society management complete | 1 |
| API Gateway setup | 1 |
| Complaint service scaffold | 2 |
| Notice service scaffold | 1 |
| React Native login flow | 3 |
| CI/CD pipeline | 1 |
| Testing + integration | 2 |
| **Total** | **~13 days (2 weeks)** |
