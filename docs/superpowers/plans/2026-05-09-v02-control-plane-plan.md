# Buddy v0.2 Control Plane — 9-phase plan (meta-dogfood)

> **Purpose:** Buddy v0.2 Control Plane (multi-session dashboard) plan을 9-phase orchestrator로 구체화. **Meta-dogfood — buddy가 자기 자신의 v0.2 cycle 입력**.
> **Predecessor outline:** [`docs/roadmap.md §4`](../../roadmap.md#4-v02--control-plane-멀티-세션-dashboard) — informal PRD, 본 plan 의 §1 입력.
> **Dogfood ledger:** [`docs/notes/2026-05-09-dogfood-meta-buddy-v02.md`](../../notes/2026-05-09-dogfood-meta-buddy-v02.md) — phase 진행 + 마찰 기록.
> **Status (2026-05-09):** §1 → §3 진입. §4~§7 다음 세션. §8~§9 n/a (production traffic / EOL 시점).

---

## §1 Idea & Business Validation

> 출처 walkthrough: [`plugin/skills/concretize-idea/PROCEDURE.md`](../../../plugin/skills/concretize-idea/PROCEDURE.md) 8-stage 파이프라인 — informal artifact 보유로 *condensed* 적용. Stage 1 gate / Stage 3 gate 통과로 가정 (v0.1 dogfood 신호 입력).

### 1.1 Core hypothesis

Claude Code 사용자는 *동시에 여러 세션*을 띄우는 패턴이 흔하다. v0.1 의 `buddy doctor` / `buddy stats` 는 단일 머신·단일 세션 관점 — 다중 세션의 token / cost / hook health 를 한 화면에 못 본다. v0.2 = 이 gap 메우는 control plane.

### 1.2 Validated assumptions

| 가정 | Validation 출처 | 확신도 |
|------|---------------|--------|
| 사용자가 multi-session으로 작업 | recon 패턴이 PID→JSONL 매핑 검증함 ([harness analysis 메타-인사이트 #1](../../../harness-engineering-analysis.md)) | High |
| transcript JSONL 에 token usage 정보 존재 | v0.1 spec §6.1 옵션 A `tokenUsage` 스키마 검증, decision-1 §1-D | High |
| Cost estimate 수요 (token-monitor 같은 외부 도구가 같은 일을 함) | decision-1 §1-D "token-monitor와 중복" 분석 | High |

### 1.3 Open questions (이번 cycle 입력)

- D-1: TUI vs web (B-4.4) — §3 design 의 핵심 결정.
- D-2: 멀티-머신 통합 — v1.0+ 로 deferred.

### 1.4 Business viability (condensed)

| 차원 | 평가 |
|------|------|
| Market | Claude Code 사용자 (정확한 모수 미공개, 추정 만 단위 ~ 수십만). |
| Customer | Primary = Claude Code 파워유저 (multi-session). Buyer = 동일 (개인 도구). |
| WTP | OSS 가정 — 직접 monetize X. value = "buddy 의 신뢰 wedge 위에 dashboard 가 있어야 v1.0 통합 가능". |
| GTM | GitHub releases + README + 입소문. 외부 채널 0. |
| Risks | (1) v0.2 dogfood feedback 부재 시 TUI/web 결정 불가, (2) token-monitor 와 책임 중복. |
| Regulatory | n/a (로컬 도구). |

> **Stage 3 gate result:** 통과. WTP 수치 부재는 OSS 특성, regulatory clear, 치명 결함 없음.

### 1.5 Customer segmentation

- **Primary user:** Claude Code 사용자 중 multi-session + token cost 추적 의도. Early adopter = `harness/token-monitor` 사용자 + `recon` 사용자 (이미 patchwork 사용 중).
- **Secondary:** team / org 사용자 (멀티-머신 — v1.0+).

### 1.6 Risk register

- **Technical:** transcript JSONL 스키마 변경 (Claude Code 버전 dependency). 완화: schema validation, 변경 시 fail-soft.
- **Business:** dogfood feedback 회수 지연으로 TUI/web 결정 미확정 (B-1 의존).
- **Legal:** 0.

### 1.7 PRD draft pointer

이 plan 의 §1.1~§1.6 + roadmap.md §4 가 PRD 입력. 별도 PRD 파일 미생성 (informal PRD 보존 정책).

### 1.8 autoplan review (skipped — 다음 phase 진입 직전 재검토)

> autoplan 4-mode review (review-scope / engineering / design / devex) 는 §3 design 산출물 확정 후 별도 cycle 로 호출 권장. 현재 informal PRD 단계는 검토 비용 대비 신호가 약함.

---

## §2 Feature Definition & Backlog

> 출처 walkthrough: [`plugin/skills/define-features/PROCEDURE.md`](../../../plugin/skills/define-features/PROCEDURE.md) — *(다음 세션)*

(다음 세션에서 채울 항목)
- T1~T6 features 의 actor / use case 매핑
- `identify-actors` skill 호출 결과 형식 적용
- feature 명세 표준 양식 (skill-map.md §4 참조)

---

## §3 Technical Design

> 출처 walkthrough: [`plugin/skills/design-system/PROCEDURE.md`](../../../plugin/skills/design-system/PROCEDURE.md) — *(이번 세션 진입까지)*

### 3.1 핵심 결정 영역 (§3 cascade chain)

§3 cascade 5 stage 모두 v1.0.6~v1.0.8 에서 사용 가능:
1. `define-tech-stack` — 언어/프레임워크/DB/runtime
2. `map-use-cases-to-infra` — actor × infra matrix
3. `derive-system-topology` — system boundary
4. `design-data-model` — schema / migration
5. `design-api-contract` — REST/GraphQL/RPC
+ ADR (`write-adr`) — 결정 영속화

SaaS 패턴 3 stage (v1.0.8): `design-event-schema` / `design-auth-model` / `design-tenant-model` — v0.2 가 OSS 단일 머신 도구라 **모두 미적용** (auth/tenant 무관, async event 없음).

### 3.2 v0.2 의 §3 결정 — 시작점만 (다음 세션 진행)

(다음 세션에서 채울 항목)
- D-1 TUI vs web 결정 — design-system stage 안에서 trade-off 평가
- T1 session discovery — derive-system-topology 적용
- T2 transcript parser — design-api-contract (internal API) 적용
- T5 cost estimate — design-data-model (pricing table 스키마) 적용

---

## §4~§7 — 다음 세션 이후

(skip)

---

## §8~§9 — n/a

- §8 iterate-product: production traffic 의존 (v0.2 release 후).
- §9 manage-lifecycle: deprecation/EOL 시점 (1년+).

---

## 다음 액션

1. dogfood ledger F-2 / F-3 추가 (이번 세션 §1 walkthrough 마찰 정리).
2. 다음 세션: §2 진입 — `define-features` PROCEDURE 적용.
3. §3 design 본격화 — TUI vs web trade-off 평가 + ADR 작성.
