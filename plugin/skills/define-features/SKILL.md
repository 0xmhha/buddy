---
name: define-features
description: This skill should be used when the user wants to "define features", "create feature backlog", "break down a PRD into features", "identify what to build", "map use cases", "define actors", or has a PRD ready and needs to plan implementation. Orchestrates §2 Feature Definition & Backlog phase — runs actor identification, use case mapping, system boundary analysis, and feature composition.
---

# define-features — §2 Feature Definition & Backlog Orchestrator

§2 라이프사이클 단계의 진입점. PRD → feature backlog (actor / use case / system boundary 포함 feature spec 목록) 생성.

**진입 조건**: PRD 확정 (§1 산출물).
**산출물**: Feature backlog — 각 feature는 actor list + per-actor use cases + system boundary + acceptance criteria + test_plan 포함.
**다음 phase**: Feature backlog 확정 후 → `design-system` (§3).

---

## 핵심 원칙 (Q8=(a))

feature는 **여러 actor의 use case 합성**으로 정의된다. actor / use case / system boundary 매핑이 §2의 첫 작업이며, feature는 합성의 *결과*이지 입력이 아니다.

use case 분해 없이 feature만 정의하면:
- ❌ §3 infra 설계 시 어느 시스템이 무엇을 담당하는지 불명확
- ❌ §4 actor별 병렬 implementation track 분배 불가
- ❌ §6 actor별 통합 테스트 누락
- ❌ §8 actor별 funnel metric 측정 불가

---

## Stage 흐름

```
define-features (§2 phase orchestrator)
├── stage 1: identify-actors            (actor 열거 — user/admin/system/3rd-party 분류)
├── stage 2: map-actor-use-cases        (actor별 use case 식별)
├── stage 3: map-use-case-to-system-boundary  (use case → 시스템 경계 매핑)
├── stage 4: compose-feature-from-use-cases   (cross-actor use case → feature)
├── stage 5: define-feature-spec        (feature spec 작성 — 확장 포맷)
├── stage 6: query-feature-registry     (유사 feature 검색 — reuse/fork/new 결정)
├── stage 7: score-feature-priority     (RICE/ICE/MoSCoW 우선순위)
├── stage 8: map-feature-dependencies   (feature 간 선후 그래프)
├── stage 9: split-work-into-features   (큰 feature 분리)
└── stage 10: triage-work-items         (백로그 상태 머신 진입)
```

---

## 실행 절차

### Stage 1: Actor 식별

PRD에서 시스템에 참여하는 모든 actor를 열거한다.

Actor 분류 기준:
- **user**: 최종 사용자 (role별로 세분화: anonymous user, authenticated user, admin 등)
- **system**: 내부 서비스/백엔드 (auth-service, notification-service 등)
- **3rd-party**: 외부 SaaS/API (SendGrid, Stripe, GitHub OAuth 등)
- **external-tool**: 관리자 도구, 모니터링, CI/CD 시스템

`identify-actors` skill을 invoke한다. skill이 미존재하면 orchestrator가 직접 수행.

### Stage 2: Actor별 Use Case 매핑

각 actor의 시점에서 use case를 식별한다.

형식:
```
Actor: {actor_id}
Use cases:
  1. {동사 + 목적어} (예: "이메일/비밀번호로 회원가입한다")
  2. ...
```

`map-actor-use-cases` skill을 invoke한다.

### Stage 3: System Boundary 매핑

각 actor의 use case가 어느 시스템 경계에서 실행되는지 매핑한다.

예시:
```
Actor: user → use case "회원가입" → system_boundary: frontend-spa
Actor: auth-system → use case "credentials 검증" → system_boundary: backend-auth-service
Actor: email-verifier → use case "verification 이메일 발송" → system_boundary: external-saas (SendGrid)
```

`map-use-case-to-system-boundary` skill을 invoke한다.

### Stage 4: Feature 합성

Cross-actor use case를 묶어 feature를 정의한다.

예시: `signup-email-password` feature = user(frontend) + auth-system(backend) + email-verifier(external SaaS) use case 합성

`compose-feature-from-use-cases` skill을 invoke한다.

### Stage 5: Feature Spec 작성

각 feature에 대해 확장 포맷으로 spec을 작성한다.

Feature spec 필수 포맷:
```yaml
feature_id: {kebab-case-id}
name: {feature 이름}
summary: {한 문장}
problem: {해결하는 문제}
actors:
  - id: {actor_id}
    use_cases:
      - {use case 1}
      - {use case 2}
    system_boundary: {frontend-spa / backend-{service} / external-saas / ...}
scope: {포함 범위}
out_of_scope: {제외 범위}
acceptance_criteria:
  {actor_id}: {actor 관점 완료 기준}
test_plan:
  per_actor:
    {actor_id}: {unit / integration / E2E 기준}
  integration: {cross-actor 흐름 테스트}
implementation_notes: {설계 결정 / trade-off}
code_links: []
status: draft
owners: []
updated_at: {ISO 8601}
```

`define-feature-spec` skill을 invoke한다.

### Stage 6: Feature Registry 조회

`query-feature-registry` skill을 invoke해 유사 feature를 검색한다.
- reuse 가능하면: 기존 spec에서 fork
- 변형 필요하면: 기존 spec을 base로 adapt
- 신규라면: stage 5 spec 확정

### Stage 7: 우선순위 결정

`score-feature-priority` skill을 invoke한다 (미존재 시 orchestrator가 직접 수행).

RICE 기준:
- Reach: 영향받는 actor 수 / 예상 사용자 비율
- Impact: 문제 해결 강도 (1-10)
- Confidence: 추정 신뢰도 (%)
- Effort: 개발 공수 (person-weeks)

MoSCoW 분류: Must/Should/Could/Won't

### Stage 8: Dependency 매핑

`map-feature-dependencies` skill을 invoke해 feature 간 선후 그래프를 작성한다 (미존재 시 orchestrator가 직접 수행).

```
feature-A → feature-B (B는 A의 구현 후 가능)
feature-C || feature-D (병렬 진행 가능)
```

### Stage 9-10: Backlog 정리

`split-work-into-features`와 `triage-work-items`로 최종 backlog를 생성하고 상태 머신에 진입한다.

---

## Cross-phase 파급 효과

§2에서 생성된 actor / use case / system boundary는 다음 phase의 입력이 된다:

| Phase | 활용 방식 |
|-------|---------|
| §3 Technical Design | actor system boundary → infra component 매핑 |
| §4 Implementation Plan | actor별 implementation track = 병렬 worker 분배 단위 |
| §6 Quality | actor별 통합 테스트 설계 (user E2E / system unit / 3rd-party contract) |
| §8 Iterate | actor별 funnel metric (conversion / success rate / deliverability) |

---

## 다음 phase

- `/buddy:design-system` — §3 Technical Design (권장)
- `/buddy:plan-build` — feature backlog가 크고 명확하면 §3를 건너뛰고 §4로 (소규모 프로젝트)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §§2, §3.4
- Feature spec 포맷 (확장): 동 문서 §3.4
