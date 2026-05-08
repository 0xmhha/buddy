# design-auth-model — OAuth2 / JWT / SAML / SSO / RBAC 다층 결정

§3 Technical Design phase 의 stage. **define-tech-stack 의 generic auth choice 보다 deeper layer**. authentication (sign-in mechanism) + session (token lifetime + storage) + authorization (RBAC / ABAC / ReBAC) + federation (SSO / SAML / OIDC) + MFA (TOTP / WebAuthn / SMS) 5 axis 의 design decision. 산출물은 auth matrix + token lifetime policy + role hierarchy + MFA enforcement plan + ADR handoff.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Authentication 만 design** — sign-in 만 보고 authorization (어느 user 가 어느 resource 접근) 결정 안 함.
- **Token lifetime 0 또는 무한** — 너무 짧으면 UX 마찰, 너무 길면 revocation gap.
- **MFA 무결정** — "필요할 때 추가" → security incident 발생 후 회고적 도입 비용 큼.
- **Federation 결정 deferred** — enterprise 고객 sign-up 시점에 SAML 강제 도입 → roadmap 지연.
- **Session storage 결정 모호** — JWT stateless 의 trade-off 와 session-cookie 비교 안 함.
- **Password policy 결정 부재** — minimum length / breach check / rotation policy.
- **Account recovery anti-pattern** — security questions / SMS-only 등 weak path.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

auth 영역의 5 axis decision 통합. design-api-contract 가 endpoint level 결정 후 본 skill 이 cross-endpoint auth model 결정. 5 axis 모두 명시 결정 = security incident risk 사전 회피.

## 2. 사용 시점 (When to invoke)

- §3 design 단계, design-api-contract 후 (chain 권장)
- 신규 auth 요구 (enterprise SSO, B2B federation)
- security incident 후 model 재평가
- MFA 도입 결정 시
- compliance audit 진입 시 (SOC 2 / HIPAA — auth model 명시 필수)

## 3. 입력 (Inputs)

### 필수
- §3 design-api-contract 산출물 (endpoint + auth header 형식)
- §3 define-tech-stack ADR (auth library / framework — Auth0 / Cognito / Supabase Auth / 자체)
- §2 feature spec actor list (User / Tenant Admin / Auditor / Ops)
- pricing tier (Free / Pro / Enterprise — federation 차이)

### 선택
- 산업 규제 (HIPAA / FINRA / SOC 2 — MFA 의무, password policy)
- enterprise customer 요구 (SAML / SCIM / Just-in-time provisioning)
- existing user base migration (legacy auth → new model)

### 입력이 부족할 때 forcing question
- "MFA 의무 vs 권장? 산업 규제 / 산업 표준 / 사용자 기대 결정 필요."
- "Enterprise tier 에서 SAML 의무? B2B 고객 onboarding 시 매번 도입 비용보다 v1 부터 design 비용이 작음."
- "session storage 가 JWT stateless 인가, server session 인가? revocation latency requirement 결정."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — token lifetime / MFA / federation 모두 단호 결정.
- **사용자 입력을 challenge** — "OAuth2 쓰면 됨" 발화에 "어느 grant? authorization code? refresh token rotation? PKCE?" push.
- **Specificity 강제** — vague "secure auth" 거부, "JWT RS256 access 15min + refresh HttpOnly 30day with rotation, argon2id mem=64MB t=3 p=4, MFA TOTP optional Pro, MFA WebAuthn enforced Enterprise".
- **Layered defense** — 단일 token 보다 access+refresh+session, MFA layer 추가.

도메인 원칙:

1. **5 axis 의무** — Authentication / Session / Authorization / Federation / MFA.
2. **Token lifetime trade-off 정량** — UX (마찰) vs security (gap).
3. **Authorization model 명시** — RBAC (role-based) / ABAC (attribute-based) / ReBAC (relationship-based).
4. **Password policy NIST 800-63B** — length ≥ 12 + breach check + no forced rotation.
5. **Account recovery 명시** — email-only weak / TOTP backup / recovery codes / human escalation.

## 5. 단계 (Phases)

### Phase 1. Authentication (sign-in mechanism)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Primary mechanism | **email + password** | broadest reach, fallback before SSO |
| Password hash | **argon2id** (mem=64MB, t=3, p=4) | NIST recommended, GPU-resistant |
| Password policy | **min 12 char + HaveIBeenPwned breach check + no forced rotation (NIST 800-63B)** | UX-friendly + breach-aware |
| Account lockout | 5 failed → 15 min lockout + audit log | brute-force defense |
| Account recovery | email link (24h TTL) + TOTP backup codes 10ea | layered |

post-launch enhancement candidates: passkey (WebAuthn) primary, magic link.

### Phase 2. Session (token lifetime + storage)

| Aspect | Decision | Trade-off |
|--------|----------|-----------|
| Access token | **JWT RS256** (asymmetric — easier rotation + multi-service verify) | symmetric (HS256) 보다 키 관리 안전, sign 비용 약간 증가 |
| Access TTL | **15 min** | 30 min UX better but revocation gap → 15 min compromise |
| Refresh token | **opaque random + HttpOnly Secure SameSite=Lax cookie** | XSS 노출 방지, CSRF 는 SameSite + Idempotency-Key |
| Refresh TTL | **30 day** with **rotation on use** | 사용 시 매번 rotate → 탈취 시 detect 가능 |
| Refresh rotation detect | replay 발견 → user 의 모든 session revoke + alert | session theft pattern detect |
| Storage | access = memory only / refresh = HttpOnly cookie | XSS 안전 |
| Revocation | denylist in Redis (TTL = remaining access lifetime) | 즉시 revoke 가능 |

### Phase 3. Authorization (RBAC / ABAC / ReBAC)

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Model | **RBAC** with tenant scope | SaaS auth 의 4 actor (User / Admin / Auditor / Ops) → role 4 개로 충분 |
| Role hierarchy | User < TenantAdmin < Owner. Auditor (read-only) cross-cutting. Ops (platform staff) global. | layered + cross-cutting |
| Scope | per-tenant 모든 resource. cross-tenant 는 Ops 만. | RLS in DB layer (per design-data-model) |
| Permission grain | endpoint-level (예: `users:read` / `users:write` / `audit:read`) | function-level granular |
| Custom policy | ABAC enhancement Pro+ tier (예: time-of-day / IP whitelist) | enterprise 특수 요구 |
| Policy storage | code (compile-time) — runtime DB policy 는 v2+ | simpler v1, v2 에서 dynamic policy |

role-permission matrix (compile-time):

| Role | users:read | users:write | sessions:write | audit:read | admin:* | system:* |
|------|------------|-------------|----------------|------------|---------|----------|
| User (self only) | own | own | own | none | none | none |
| TenantAdmin | tenant | tenant | tenant | tenant | tenant | none |
| Owner | tenant + billing | tenant | tenant | tenant | tenant | none |
| Auditor | read-only | none | none | tenant | none | none |
| Ops | global | global | global | global | global | global |

### Phase 4. Federation (SSO / SAML / OIDC)

| Tier | Federation support |
|------|---------------------|
| Free | email + password only |
| Pro | + Google OAuth + GitHub OAuth (OIDC) |
| Enterprise | + SAML 2.0 + custom OIDC + SCIM (auto-provisioning) |
| (post-launch) | + Just-in-time provisioning + HRIS sync (Workday, BambooHR) |

federation library:
- v1: built-in (TS Fastify plugin) for Google + GitHub OAuth
- enterprise-only SAML: **WorkOS** ($/SSO connection) or **JumpCloud** — vendor cost vs build
- decision: **WorkOS** Enterprise tier ($X/connection/mo) — build 비용 (~6 PM) > 1y vendor cost

### Phase 5. MFA enforcement

| Tier | MFA policy |
|------|------------|
| Free | optional TOTP (RFC 6238) |
| Pro | optional TOTP + recovery codes 10ea |
| Enterprise | **enforced** WebAuthn (passkey) primary + TOTP fallback |
| Compliance (HIPAA / SOC 2) | enforced regardless of tier — admin-flag override |

MFA library:
- TOTP: speakeasy (Node) + QR generation
- WebAuthn: SimpleWebAuthn (TS-native)
- enrollment UX: 가입 후 24h grace period, 그 후 강제 (Enterprise)

backup mechanisms:
- recovery codes (10 single-use, generated at TOTP enrollment)
- support escalation (human-verified, audit-logged, rate-limited)
- NEVER security questions (anti-pattern, NIST advised against)

## 6. 산출물 형식 (Output format)

```markdown
## design-auth-model Output — <project name>

### Summary
<3 줄: auth mechanism / session strategy / authorization model / federation tier / MFA enforcement>

### Authentication
| Decision | Choice | Rationale |
|----------|--------|-----------|
| ... | ... | ... |

### Session
| Aspect | Decision | Trade-off |
|--------|----------|-----------|
| Access token | JWT RS256 / 15min | ... |
| Refresh token | opaque / 30d / rotation | ... |
| Revocation | Redis denylist | ... |

### Authorization
| Aspect | Decision |
|--------|----------|
| Model | RBAC + per-tenant scope |
| Role hierarchy | User < TenantAdmin < Owner. Auditor cross-cutting. Ops global. |
| Permission grain | endpoint-level |

| Role | users | sessions | audit | admin | system |
|------|-------|----------|-------|-------|--------|
| User | own | own | none | none | none |
| TenantAdmin | tenant | tenant | tenant | tenant | none |
| ... | ... | ... | ... | ... | ... |

### Federation
| Tier | Support |
|------|---------|
| Free | email + password |
| Pro | + Google + GitHub OAuth |
| Enterprise | + SAML (WorkOS) + SCIM |

### MFA
| Tier | Policy |
|------|--------|
| Free | optional TOTP |
| Pro | optional TOTP + recovery codes |
| Enterprise | enforced WebAuthn + TOTP fallback |
| HIPAA / SOC 2 | enforced regardless |

### ADR Handoff
다음 단계: `/buddy:write-adr "Auth model — JWT + RBAC + WorkOS Enterprise SAML + WebAuthn MFA"`

### Cascade
- **§3 design-api-contract**: auth header 형식 + endpoint role 정합
- **§3 design-data-model**: users.role / sessions.refresh_token_hash schema
- **§7 setup-canary-deploy**: auth-related metric (login p99, failed auth rate) 정합
- **§6 audit-security**: auth model 검증

### Next Step
<구체 action — 1줄: 예 "WorkOS account 생성 + WebAuthn library 통합 task 생성">
```

## 7. Cross-phase cascade

- **§3 design-api-contract**: auth header / endpoint role 정합
- **§3 design-data-model**: schema 정합
- **§6 audit-security**: model 검증
- **§7 setup-incident-paging**: auth alert routing (failed login rate spike)

## 8. 다음 skill (next in stage flow)

- `design-tenant-model` (Cluster B) — multi-tenant 결정
- `write-adr` — auth model 영속화

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack`** (§3) — 그것은 auth library / framework 선택 (Auth0 vs Cognito vs 자체), 본 skill 은 그 위의 5 axis design.
- **vs `design-api-contract`** (§3) — 그것은 endpoint-level auth header 형식, 본 skill 은 cross-endpoint model.
- **vs `audit-security`** (§6) — 그것은 implementation 후 침투 테스트, 본 skill 은 design 단계 model.

## 10. 중요 규칙

- **5 axis 의무** — single axis 거부.
- **Token lifetime 정량** — minute / day 단위.
- **Authorization model 명시** — RBAC / ABAC / ReBAC 중 어느 것.
- **Password policy NIST 준수** — forced rotation 거부.
- **Account recovery 명시** — security questions 거부.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 8 출력 섹션 모두 채워짐
- [ ] Authentication 5 decision 명시 (mechanism / hash / policy / lockout / recovery)
- [ ] Session 6 aspect 명시 (access type / TTL / refresh / TTL / rotation / storage / revocation)
- [ ] Authorization model + role hierarchy + permission matrix
- [ ] Federation 3+ tier
- [ ] MFA 4 tier (Free / Pro / Enterprise / Compliance)
- [ ] ADR handoff 명시
- [ ] §4 posture
- [ ] §0 anti-pattern 부재 — auth only / lifetime 부재 / MFA 미정 / federation deferred / storage 모호 / password policy 부재 / recovery anti-pattern 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
