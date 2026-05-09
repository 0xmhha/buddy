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

> 출처 walkthrough: [`plugin/skills/define-features/PROCEDURE.md`](../../../plugin/skills/define-features/PROCEDURE.md) (10 stage), condensed 적용.
> External reference: `0xmhha/ai-m` (Go TUI multi-session manager) — `STATUS.md` "Trigger-driven backlog" 패턴이 buddy 의 D-1~D-7 과 동일 철학.

### 2.1 Actors (stage 1: identify-actors)

| Actor | 분류 | 역할 |
|-------|------|------|
| CLI user | user (single) | `buddy` CLI 호출, dashboard 보기, config tune. v0.2 의 유일한 사람-actor. |
| Active Claude Code session(s) | system (multi-instance, **N≥1**) | 각 세션이 transcript JSONL + hook event 의 source. cardinality = 동시 떠있는 세션 수. |
| buddy daemon | system | aggregator + outbox drain (v0.1 에서 이미 존재). v0.2 는 daemon 에 *session enumerator* + *transcript reader* 책임 추가. |
| Anthropic transcript JSONL | 3rd-party (data) | `~/.claude/sessions/*.json` + `~/.claude/projects/**/*.jsonl` 가 contract. 스키마 변경 = breakage 위험 (F-7 참조). |
| Pricing table | 3rd-party (data) | model 단가. 코드 embed 으로 시작 (B-4.5), 외부 fetch 는 v1.0+ deferred. |

**핵심 cardinality**: v0.1 = 1 user × 1 daemon × 1 DB. v0.2 = 1 user × 1 daemon × **N sessions × N transcripts**. 이 N 이 §3 의 모든 design decision 입력.

### 2.2 Use cases (stage 2: per-actor)

| ID | Actor | Use case | Trigger |
|----|-------|----------|---------|
| UC-1 | CLI user | "지금 떠있는 모든 Claude Code 세션을 한 화면에 본다" | `buddy dashboard` (또는 `sessions list`) 호출 |
| UC-2 | CLI user | "각 세션의 누적 token / 추정 cost 를 본다" | UC-1 dashboard 안에서 column 표시 |
| UC-3 | CLI user | "특정 세션의 hook health drill-down" | UC-1 dashboard 에서 session 선택 |
| UC-4 | buddy daemon | "transcript JSONL 변경을 감지해 token usage aggregate 갱신" | fsnotify (ai-m 도 동일 의존) 또는 polling |
| UC-5 | buddy daemon | "활성 세션을 PID + transcript path 로 enumerate" | 주기적 (e.g. 5s) |

### 2.3 Features (stage 4-5: compose + spec)

| Feature | UC composition | 산출 (interfaces) | 의존 (depends_on) |
|---------|----------------|-------------------|-------------------|
| F1 Session discovery | UC-5 | `internal/sessions/` package: `Lister.List(ctx) []Session` interface + 1 impl (`fsLister` scans `~/.claude/`). Session struct = {id, pid, transcript_path, started_at, last_active}. | — (v0.1 인프라 무관) |
| F2 Transcript reader | UC-4 | `internal/sessions/transcript.go`: `Reader.Tail(path, lastOffset) ([]TokenUsage, newOffset)`. 증분 read + offset 보존. | F1 |
| F3 Multi-session stats | UC-2 | `internal/queries/sessions.go`: SQLite query — sessions × hook_events × token_usage join. 기존 `stats` subcommand 확장 (`--all-sessions` 또는 `sessions show`). | F1, F2 |
| F4 Cost estimate | UC-2 | `internal/pricing/`: model name → $/M token table (embed). `Estimate(usage) Cost` 함수. | F2 |
| F5 Dashboard UI | UC-1, UC-3 | `internal/ui/dashboard/` (TUI, **bubbletea + lipgloss** — D-1 결정 결과 §3.2). 단일 view: session list + per-session cost/token + drill-down. | F3, F4 |
| F6 i18n full split | (cross-cutting) | `internal/persona/en.go` 채우기 + subcommand `--config` locale 해석 (M5 deferred forward-pointer). | — |

> **Q8=(a) cascade**: feature 가 actor × use case 합성의 *결과*. F1~F4 = system actor (daemon) 의 use case 직접 매핑. F5 = user actor 가 system actor 산출을 소비. F6 = cross-cutting (i18n).

### 2.4 Backlog state (stage 7-10: priority + DAG + triage)

| Feature | Priority | 분해 가치 | DAG depth |
|---------|----------|----------|-----------|
| F1 | Must | 1 (atomic) | 0 |
| F2 | Must | 1 | 1 (after F1) |
| F3 | Must | 1 | 2 |
| F4 | Should | 1 | 1 (after F2, parallel with F3) |
| F5 | Must | 가능 (TUI view 별 split — list view → drill view 순) | 3 |
| F6 | Could | 1 | 0 (independent) |

**Critical path**: F1 → F2 → F3 → F5. 4 단계. F4 / F6 병렬.

> stage 6 (`query-feature-registry`) skip — buddy 자체에 feature registry 미존재 (A-3 작업, 본 plan 의 §3 이후 영역).

---

## §3 Technical Design

> 출처 walkthrough: [`plugin/skills/design-system/PROCEDURE.md`](../../../plugin/skills/design-system/PROCEDURE.md). §3 cascade 5 stage 모두 사용 가능 (v1.0.6~v1.0.8). SaaS pattern 3 stage (event/auth/tenant) **모두 N/A** (OSS 단일 머신).
> External reference: `0xmhha/ai-m` (`internal/llm/backend/` interface + tmux/cliwrap 두 impl 패턴, `internal/server/server.go` websocket 서버, `apps/desktop/` Electron) — buddy v0.2 design 의 directly applicable precedent.

### 3.1 §3 cascade chain — v0.2 적용

| Stage | Skill | v0.2 적용 결과 |
|-------|-------|---------------|
| 3.1.1 | `define-tech-stack` | Language **Go** (v0.1 재사용), DB **SQLite + WAL** (v0.1 재사용), TUI **`charmbracelet/bubbletea` + `lipgloss`** (ai-m precedent), File watch **`fsnotify`** (ai-m + buddy 공통 의존 후보), Process supervision = v0.1 cli-wrapper 재사용. **Lock-in**: Go monolith. Electron / web 은 v1.0+ trigger-driven backlog (3.4 참조). |
| 3.1.2 | `map-use-cases-to-infra` | UC-1/3/5 → buddy daemon process; UC-2 → in-memory aggregate + SQLite query; UC-4 → daemon 의 fsnotify watcher; transcript = 외부 fs path (`~/.claude/`). 모든 component 단일 머신 / 단일 process. |
| 3.1.3 | `derive-system-topology` | **단일 process, 4 sub-component**: (a) daemon (existing) + session enumerator + transcript reader, (b) SQLite store (existing schema + `sessions` 테이블 추가), (c) TUI dashboard (`internal/ui/dashboard/`), (d) pricing table (embed). Sub-component 간 의존 = compile-time linkage, no IPC. |
| 3.1.4 | `design-data-model` | 신규 테이블 1개: `sessions(id PK, pid, transcript_path, started_at, last_active, total_input_tokens, total_output_tokens, total_cache_read, total_cache_create, last_offset)`. transcript reader 가 `last_offset` 기반 증분 read. 기존 `hook_events` 와 `session_id` foreign key (이미 존재). Migration: `00X_sessions.sql`. |
| 3.1.5 | `design-api-contract` | **모든 API internal**. Public API = `buddy` CLI subcommand 만 (`buddy dashboard`, `buddy sessions list/show`). Internal interface 3개: `sessions.Lister`, `sessions.Reader`, `pricing.Estimator`. v1.0+ 에서 websocket protocol 노출 시 별도 spec — ai-m `internal/server/server.go` 의 JSON-RPC over websocket 패턴 참조 가능. |

### 3.2 D-1 결정 — TUI first, web/desktop deferred

**Decision: TUI 채택 (`bubbletea` + `lipgloss`).**

**Rationale (ai-m 차용 가능 패턴 평가):**

| 옵션 | Pro | Con | v0.2 채택 |
|------|-----|-----|----------|
| TUI (bubbletea) | Single binary, no extra deps, terminal-native UX, ai-m precedent + cli-wrapper 의존 일관 | 비-개발자 / GUI 선호 사용자 alienation | **✅ 채택** |
| Web (HTTP server + browser) | 친숙한 UI, multi-platform | HTTP server lifecycle 추가, port conflict 가능, browser 분리가 사용자 mental load | ❌ deferred (v1.0+ trigger-driven) |
| Desktop app (Electron — ai-m `apps/desktop/`) | Rich UI, native window | Electron 의존 ~100MB, Node + Go cross-build, single-binary 정책 위배 | ❌ N/A (v0.2 non-goal) |
| Hybrid (TUI + websocket server, ai-m 패턴) | 둘 다 만족 | 복잡도 두 배, dogfood feedback 부재로 어느 쪽 신호가 강한지 불명 | ❌ deferred |

**Trigger to revisit**: B-1 dogfood feedback 에서 *명시적 web 요청* 또는 *team / multi-user* 시나리오 등장 시. ai-m 의 STATUS.md "trigger-driven backlog" 철학 그대로 차용.

### 3.3 차용 가능한 ai-m 패턴 (적용 / deferred 분류)

| 패턴 | 위치 (ai-m) | v0.2 적용 |
|------|-------------|----------|
| Backend abstraction (interface + 2 impls) | `internal/llm/backend/` (tmux / cliwrap) | **부분 적용** — `sessions.Lister` interface + 1 impl (`fsLister`). 두 번째 impl 은 trigger-driven (예: `procLister` PID 기반 — fs scan 으로 부족할 때). |
| TUI core + websocket server 분리 | `internal/server/server.go` | **deferred (v1.0+)** — 3.2 결정대로 TUI only. |
| `STATUS.md` "Resolved + Trigger-driven backlog" 단일 문서 | `docs/superpowers/STATUS.md` | **차용 가치 있음** — buddy `tasks.md` 의 D-1~D-7 가 동일 철학, 형식만 STATUS.md 더 단순. **A-1 후속 SSoT polish 후보** (Wave 1 끝났으니 deferred). |
| `PersistentPicker` reattach UI (`AIM-P2-2`) | `internal/ui/views/persistent_picker.go` (추정) | **N/A v0.2** — buddy v0.2 는 dashboard read-only. session reattach 는 v0.3 task DAG executor 의 retry 영역. |
| `--restart` policy + `default_restart` config (`AIM-P2-5`) | engine + cliwrap backend | **N/A v0.2** — buddy daemon 은 자기 자신만 supervise. session-level restart 는 v0.3+. |

### 3.4 Trigger-driven backlog (v0.2 → v1.0+ deferred)

| ID | Title | Trigger condition |
|----|-------|------------------|
| BUDDY-D-WEB | websocket server 노출 | dogfood feedback 에서 web UI 명시적 요청 |
| BUDDY-D-DESKTOP | Electron / Tauri desktop app | team / org 사용자 등장 (멀티-머신 D-2 와 묶임) |
| BUDDY-D-MULTI-IMPL | `sessions.Lister` 두 번째 impl (procLister) | fs scan 으로 active session 누락 케이스 발견 |
| BUDDY-D-RESTART | session-level restart policy | v0.3 task DAG executor 가 session 재시작 책임 가질 때 |

> **AIM 차용 정책**: code copy 하지 않음. 패턴 / 구조 / decision rationale 만 차용. ai-m 은 cli-wrapper 위에서 *세션 자체* 를 관리 (full TUI multiplexer), buddy v0.2 는 *세션 관찰* (read-only dashboard). 책임 경계 명확.

### 3.5 §3 ADR 작성 (write-adr stage 호출)

본 plan §3.1~§3.4 = ADR 입력. 별도 ADR-002 파일은 §3 본격화 후 작성 — 결정 후보:
- **ADR-002 v0.2 dashboard UI = TUI**
  - Context: 3.2 표 + ai-m precedent + dogfood feedback 부재.
  - Decision: bubbletea + lipgloss TUI, single binary.
  - Consequences: web/desktop deferred (3.4 trigger-driven). single-binary 정책 유지.
  - Status: Proposed (코드 진입 시 Accepted 로).

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
