# Buddy — 작업 인벤토리

> 두 트랙(Plugin + Go CLI) + 코드 housekeeping의 *pending* 작업을 한 자리에서 본다.
> 우선순위는 §끝의 [우선순위](#우선순위-권장)에서.
>
> **SSoT 분담:**
> - 이 문서: cross-track 작업 인벤토리 (실행 단위)
> - [`roadmap.md`](./roadmap.md): Go CLI 트랙의 마일스톤 SSoT (M5/M6/v0.2/v0.3/v1.0)
> - [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md): plugin 9-phase 아키텍처 SSoT
> - [`HANDOFF.md`](./HANDOFF.md): 세션 인계 가이드
> - [`notes/2026-05-09-handoff-N1-closure.md`](./notes/2026-05-09-handoff-N1-closure.md): N-1 closure (ADR-001 적용 결과)
>
> 작성일: 2026-05-05 / 최종 갱신: 2026-05-09 (v1.0.8 + N-1 closure + Phase 5 ext A+B+C 8 skill Done 반영) / 상태: WORKING

---

## 트랙 상태 요약 (한 줄)

| 트랙 | 상태 | 마지막 release |
|------|------|---------------|
| **plugin buddy** (9-phase orchestrator + **148 procedures + 99 commands**) | ACTIVE — Skill Completion Cycle 100% (44/44). charter scope 12 stage cover 100%. unreleased (v1.1.0 후보) | v1.0.8 (2026-05-08, 신규 batch 1~7 unreleased) |
| **cli buddy** (TUI 자동화 agent 관리 — charter §3) | PAUSED — spec 미작성. v0.1.0 의 hook reliability monitor 가 sub-feature | v0.1.0 (2026-04-26) |
| **Housekeeping** | ad-hoc | — |

---

## A. Plugin 트랙

### A-1. SSoT drift 갱신 (Wave 1)

| ID | 작업 | 위치 | 완료 조건 |
|----|------|------|----------|
| A-1.1 | Plugin 진행 상태 표 v1.0.5 → v1.0.8 갱신 | `HANDOFF.md` 트랙 상태 표 | Phase 5 ext Cluster A+B+C 8 skill Done 반영, "Total" 행 105 procedures / 57 commands |
| A-1.2 | A-2 잔여 표 갱신 (37 → 29) | `tasks.md` §A-2 | Phase 5 ext 8 skill 완료 반영, 잔여 cluster별 재집계 |
| A-1.3 | README counts 갱신 | `README.md` | 49 commands → 57, 97 skills → 105, stage skills 표에 신규 8 skill 추가 |
| A-1.4 | N-1 closure 반영 | `HANDOFF.md` §0 / §3 / §11 | ADR-001 적용 사실 + `disable-model-invocation: true` 컨벤션 명시 |

### A-2. 잔여 stage skill — ✅ 모두 Done (Skill Completion Cycle 100%, 2026-05-10)

> 출처: [`spec §4 Stage Skill Gap`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md#4-단계별-skill-군집화--gap-분석) + [`Phase 7 deferred re-evaluation`](./superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md) + [`skill-completion-plan`](./superpowers/plans/2026-05-10-skill-completion-plan.md).
>
> 본 cycle 산출 commits: `18a79b6` (Batch 1) → `b23a995` (2) → `1c5c528` (3) → `6f0677d` (4) → `3544342` (5) → `9cb6ca3` (6a) → `f5b8afb` (6b) → `900944f` (7).

| Phase | 잔여 (이전) | Skill Completion Cycle 후 | 상태 |
|-------|-----------|--------------------|-----|
| §1 concretize-idea (Cluster E 5) | 5 | 0 | ✅ Done (Batch 1+2) |
| §2 define-features | 1 | 0 | ✅ Done (Batch 2) |
| §3 design-system 부가 (Cluster B residual 4) | 4 | 0 | ✅ Done (Batch 3) |
| §5 build-feature (Cluster D 5) | 5 | 0 | ✅ Done (Batch 4) |
| §6 verify-quality 부가 (Cluster A residual 3) | 3 | 0 | ✅ Done (Batch 5) |
| §8 iterate-product (Cluster F 7) | 7 | 0 | ✅ Done (Batch 6a) |
| §9 manage-lifecycle (Cluster G 4) | 4 | 0 | ✅ Done (Batch 7) |
| 그룹 4 (12) — form-factor / design 적용 4 / 그로스+마케팅 6 / target-market 결정 | (신규 발견) | 0 | ✅ Done (Batch 1, 3, 6b) |
| 그룹 2 신규 (review-legal-regulatory) | (신규 발견) | 0 | ✅ Done (Batch 2) |
| 그룹 2 통합 PROCEDURE 갱신 (define-product-spec / review-engineering) | (신규 발견) | 0 | ✅ Done (Batch 7) |
| **합계** | **29 + 12 + 1 + 2** | **0 신규 잔여** | **44/44 (100%)** |

**Deferred 보존 (trigger 발화 시 활성):**

| 항목 | trigger | 결정 |
|------|--------|------|
| Korea cluster 3 (`consult-korea-legal-context` / `draft-korea-patent-application` / `audit-korea-cii-vulnerability`) | target market = Korea 결정 시 | D-F F1 |
| `analytics-mcp` | §8 cluster F 일부 구현 후 — *현재 trigger 가능* | D-C C2 |
| `feature-management-mcp` | cli buddy 트랙 spec 작성 시점 | D-C C2 (cli buddy 트랙 분리) |

> 아래 §A-2.1 ~ §A-2.7 sub-section 은 *historical 보존* — 각 cluster 의 정의 + 작성 시점 reference. 신규 잔여 inventory 는 본 표 위쪽 갱신 대상.

#### A-2.1 §1 concretize-idea — Cluster E (5)

| Skill | 1줄 용도 |
|-------|---------|
| `analyze-competition-and-substitutes` | 경쟁/대체재 분석 (positioning matrix + moat) |
| `map-customer-segments` | 사용자 vs 구매자 세그먼트 분리 + persona |
| `map-jobs-to-be-done` | JTBD 프레임워크 적용 |
| `analyze-market-size` | TAM/SAM/SOM 산출 + bottom-up/top-down cross-check |
| `conduct-customer-interview` | 인터뷰 스크립트 + 결과 코딩 |

#### A-2.2 §2 define-features (1)

| Skill | 1줄 용도 |
|-------|---------|
| `estimate-feature-effort` | feature 단위 effort estimation (T-shirt + ideal-h) |

#### A-2.3 §3 design-system 부가 — Cluster B residual (4)

| Skill | 1줄 용도 |
|-------|---------|
| `design-observability` | logs / metrics / traces 3-pillar 설계 + SLO/SLI |
| `design-secret-management` | secret store + rotation + audit + leak detection |
| `design-i18n-strategy` | locale 분기 + ICU MessageFormat + RTL + 번역 워크플로우 |
| `design-accessibility-baseline` | WCAG 2.2 AA baseline + a11y annotation 컨벤션 |

#### A-2.4 §5 build-feature — Cluster D (5, IDE 대체 가능)

| Skill | 1줄 용도 |
|-------|---------|
| `generate-from-api-contract` | OpenAPI/GraphQL → 클라이언트/서버 stub 자동생성 |
| `generate-tests-from-spec` | spec → unit/integration test skeleton |
| `pair-program-loop` | red-green-refactor pair driving 컨텍스트 |
| `refactor-with-rename-trace` | LSP rename + 호출 그래프 추적 |
| `update-docs-with-code` | 코드 변경 → README/ADR/changelog 동기화 |

#### A-2.5 §6 verify-quality 부가 — Cluster A residual (3)

| Skill | 1줄 용도 |
|-------|---------|
| `audit-i18n-coverage` | locale별 번역 누락 / fallback 누수 검출 |
| `chaos-test` | failure injection (network / pod kill / latency) |
| `audit-test-coverage-meaningful` | line coverage가 아니라 mutation/behavior 검증 |

#### A-2.6 §8 iterate-product — Cluster F (7, production traffic 의존)

| Skill | 1줄 용도 |
|-------|---------|
| `analyze-feature-adoption` | 신규 기능 활성화율 + power user 코호트 |
| `analyze-user-cohort` | acquisition cohort retention curve |
| `analyze-actor-failure-rate` | actor별 실패 패턴 + 회복 전략 |
| `analyze-cost-anomaly` | 비용 spike 탐지 + root cause |
| `triage-customer-support-ticket` | 티켓 분류 + recurring issue 패턴화 |
| `analyze-customer-feedback-corpus` | NPS/리뷰 텍스트 토픽 모델링 |
| `audit-error-budget` | SLO 잔여 burn rate + release 통제 |

#### A-2.7 §9 manage-lifecycle — Cluster G (4, 1년+ deferred)

| Skill | 1줄 용도 |
|-------|---------|
| `deprecate-feature` | deprecation timeline + sunset notice + telemetry |
| `migrate-customers` | major change 시 고객 마이그레이션 plan |
| `archive-product` | EOL 체크리스트 + data export + tombstone |
| `spin-off-feature` | 기능 분리 후 별도 product/repo로 전환 |

### A-3. buddy MCP server 확장

| ID | 작업 | 위치 | 의존 |
|----|------|------|------|
| A-3.1 | `feature.query` 구현 (PRD/도메인 태그/API shape 기반 유사 검색) | `cmd/buddy-mcp/`, `internal/feature/` | embedding 백엔드 결정 |
| A-3.2 | `feature.store` 구현 (신규 feature 명세 저장) | 동일 | 스키마 설계 |
| A-3.3 | `feature.update` 구현 (status / acceptance / code_links 갱신) | 동일 | A-3.2 |
| A-3.4 | `feature.link_code` 구현 (commit/PR/patch 연결) | 동일 | A-3.2 |
| A-3.5 | `feature.export_patch` 구현 (재사용 가능한 patch artifact 생성) | 동일 | A-3.4 |
| A-3.6 | `feature.similarity` 구현 (재사용성 점수화) | 동일 | A-3.1 |
| A-3.7 | analytics-mcp (신규 — funnel/AB/cohort 노출) | `cmd/buddy-mcp/` 또는 신규 binary | §8 cluster 일부 선결 |

> 기존 baseline: `cmd/buddy-mcp/` doctor / stats / feature tools, `claude mcp add/remove` 통합 모두 DONE.

### A-4. Plugin dogfood

| ID | 작업 | 비고 |
|----|------|------|
| A-4.1 | 실 SaaS 프로젝트에 `claude plugin install buddy@buddy` | 모든 후속 결정의 입력 |
| A-4.2 | 9-phase orchestrator 단일 cycle 검증 (idea → ship-release) | A-4.1 의존 |
| A-4.3 | dogfood 결과로 A-2 잔여 29 skill 재정렬 | A-4.2 의존 |
| A-4.4 | dogfood 중 발견된 router/orchestrator 마찰 fix | ad-hoc, A-4.2 의존 |

### A-5. N-1 후속 (deferred, low priority)

| ID | 작업 | 비고 |
|----|------|------|
| A-5.1 | Quick Win C — commands/*.md body slim (~600 → ~200 byte/file) | 누적 invocation cost 감소, 멀티-invoke 세션 trace로 측정 가능 시 |
| A-5.2 | CONTRIBUTING.md + lint enforcement of `disable-model-invocation: true` | 미래 신규 command 방지, 첫 contributor PR 직전까지 deferred |

---

## B. Go CLI 트랙 (PAUSED)

### B-1. Dogfood feedback 회수

| ID | 작업 | 비고 |
|----|------|------|
| B-1.1 | 사용자 3~7일 사용 후 [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) 작성 | 사용자 페이스 |
| B-1.2 | 회수된 피드백 → v0.2 dashboard UX (TUI vs web) 결정 | B-4.4 입력 |
| B-1.3 | 회수된 피드백 → v0.3 task tracker 통합 여부 결정 | B-5.5 입력 |

### B-2. v0.2 i18n sweep (M5 deferred)

| ID | 작업 | 위치 |
|----|------|------|
| B-2.1 | `config.ValidationError.Reason` persona catalog wiring | `internal/persona/`, `translateConfigError` |
| B-2.2 | `queries.ErrInvalidLimit` / `ErrInvalidWindow` 카탈로그 이전 | `internal/queries/` |
| B-2.3 | English locale 카탈로그 채우기 | `internal/persona/en.go` (현재 빈 map) |
| B-2.4 | Subcommand `--config` 인지 locale 해석 | root `PersistentPreRunE` |

### B-3. v0.2 release polish (M6 deferred)

| ID | 작업 | 비고 |
|----|------|------|
| B-3.1 | macOS notarization (현재 `xattr -d com.apple.quarantine` 안내) | Apple Developer ID 필요 |
| B-3.2 | Third-party Actions SHA pinning + Dependabot | supply-chain 강화 |
| B-3.3 | 별도 `ci.yml` (PR/push 시 test/vet/fmt 강제) | 현재 `release.yml`만 존재 |
| B-3.4 | `VERSION` 파일 / build-time embed | 현재 `0.1.0` 5곳 분산 |

### B-4. v0.2 Control Plane (multi-session dashboard)

> 출처: [`roadmap.md §4`](./roadmap.md#4-v02--control-plane-멀티-세션-dashboard).

| ID | 작업 | 위치 |
|----|------|------|
| B-4.1 | 활성 세션 발견 (recon 패턴 차용 — `~/.claude/sessions/*.json` + `~/.claude/projects/**/*.jsonl`) | 신규 `internal/sessions/` |
| B-4.2 | Token usage parser (transcript JSONL → `tokenUsage` 스키마 재사용) | `internal/sessions/transcript.go` |
| B-4.3 | Multi-session stats 확장 (`--all-sessions` 또는 `buddy sessions list/show`) | 기존 stats 확장 또는 신규 subcommand |
| B-4.4 | Dashboard UI — TUI (`bubbletea`+`lipgloss`) vs web — **open question (D-1)** | B-1.2 결정 후 |
| B-4.5 | Cost estimate (model 단가 테이블 × tokens) | 신규 `internal/pricing/` |
| B-4.6 | i18n full split (en/ko) — M5 T5 forward-pointer 회수 | B-2 통합 |

### B-5. v0.3 Orchestration (task DAG executor)

> 출처: [`roadmap.md §5`](./roadmap.md#5-v03--orchestration-task-dag-executor).

| ID | 작업 | 위치 |
|----|------|------|
| B-5.1 | task DAG schema (`tasks`, `task_deps`) | 신규 migration `internal/store/migrations/00X_tasks.sql` |
| B-5.2 | `buddy task add/list/run/status` CLI | `internal/cli/task.go` |
| B-5.3 | Wave 그룹화 (errgroup, `maxParallelTasks` config) | 신규 `internal/orchestrator/` |
| B-5.4 | Retry policy (exponential backoff + max attempts + quality gate) | 동일 |
| B-5.5 | 외부 task tracker 통합 — Linear/Jira/GitHub Issues vs 실행 엔진 only — **open question (D-3)** | B-1.3 결정 후 |

### B-6. v1.0 통합

> 출처: [`roadmap.md §6`](./roadmap.md#6-v10--통합-agentsmd-auto-sync-plugin-model-mcp-server).

| ID | 작업 | 위치 |
|----|------|------|
| B-6.1 | AGENTS.md auto-sync (capability 자동 기록/갱신, install/uninstall hook) | 신규 `internal/agentsmd/` |
| B-6.2 | Plugin model (외부 binary + MCP transport 재활용, IPC = stdio JSON-RPC) | 신규 `pkg/plugin/` |
| B-6.3 | MCP server 본격 노출 (A-3과 통합) | `cmd/buddy-mcp/` 또는 `buddy mcp serve` |
| B-6.4 | Cross-harness 시도 (Codex/OpenCode) — **시도 가치 검토만** | 별도 spec 필요 |
| B-6.5 | Plugin model 권한 경계 — DB 직접 vs IPC — **open question (D-5)** | B-6.2 의존 |

---

## C. 코드 housekeeping

| ID | 작업 | 위치 | 비고 |
|----|------|------|------|
| C-1 | `cmd/buddy/main.go` 685 lines 분할 (install/daemon/doctor/stats/events/hookwrap sibling 분리) | `cmd/buddy/` | v0.2 새 명령 추가 전 권장 |
| ~~C-2~~ | ~~Module path drift 정리 (`wm-it-22-00661/buddy` → `0xmhha/buddy`)~~ | ✅ Done (`4ce3ccb`) | 43 파일 / 91 import 줄 / go.mod 1 줄 일괄 변경 |
| C-3 | gofmt drift 정리 (필요 시) | repo 전체 | 한 commit으로 처리, 현재는 clean |

---

## D. 결정 대기 (Open Questions)

| ID | 질문 | 결정 trigger |
|----|------|-------------|
| D-1 | v0.2 dashboard UX (TUI vs web) | B-1 dogfood feedback |
| D-2 | v0.2 멀티-머신 통합 여부 | v1.0+ 검토 |
| D-3 | v0.3 buddy = task tracker vs 실행 엔진 only | B-1 dogfood feedback |
| D-4 | v0.3 DAG 시각화 형식 (graphviz / mermaid / TUI) | B-5 본격화 시 |
| D-5 | v1.0 plugin 권한 경계 (DB 직접 vs IPC) | v0.3 사용 패턴 |
| D-6 | M5 outbox 수동 cleanup 경로 | outbox 미드레인 케이스 발생 시 |
| D-7 | Q4 MCP 우선순위 (현재 보류, in-house 진행 중) | A-3 본격화 시 |

---

## 우선순위 권장 (Wave별)

**Wave 1 — 즉시 (다음 세션 컨텍스트 손실 방지):**

1. **A-1.1 ~ A-1.4** — SSoT drift 갱신 (`HANDOFF.md` / `tasks.md` / `README.md`)

**Wave 2 — 검증 (피드백이 잔여 우선순위 결정):**

2. **A-4.1 ~ A-4.4** — Plugin dogfood (실 SaaS 프로젝트 install + 9-phase orchestrator 검증)

**Wave 3 — 조건부 (dogfood 신호에 따라):**

3. **A-2.5** — Cluster A residual (§6 audit 3): `audit-i18n-coverage`, `chaos-test`, `audit-test-coverage-meaningful`
4. **A-2.3** — Cluster B residual (§3 design 4): `design-observability`, `design-secret-management`, `design-i18n-strategy`, `design-accessibility-baseline`
5. **A-3.1 ~ A-3.6** — buddy MCP feature.* tools (Q4 trigger 시)

**Wave 4 — 사용자 페이스:**

6. **B-1.1 ~ B-1.3** — Go CLI dogfood feedback (3~7일 사용 후)
7. **C-1** — 코드 housekeeping `cmd/buddy/main.go` 분할 (v0.2 새 명령 추가 직전). C-2 는 `4ce3ccb` 으로 완료됨.

**Wave 5 — Go CLI 본격 재개:**

8. **B-2** → **B-3** → **B-4** → **B-5**

**Wave 6 — 1년+ deferred / production traffic 의존:**

9. **A-2.4** (Cluster D 5) → **A-2.1** (Cluster E 5) → **A-2.6** (Cluster F 7) → **A-2.7** (Cluster G 4)
10. **B-6** — v1.0 통합

**Always-deferred:**

- **A-5** — N-1 후속 (Quick Win C, CONTRIBUTING.md lint)
- **C-3** — gofmt drift (현재 clean)
- **D-2 / D-4 / D-5** — 결정 trigger 미발생

---

## 참조

- [`HANDOFF.md`](./HANDOFF.md) — 세션 인계 + 워크플로우 skill 분기
- [`roadmap.md`](./roadmap.md) — Go CLI 마일스톤 SSoT
- [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) — plugin 9-phase 아키텍처 SSoT
- [`superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md`](./superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md) — Cluster A~H 분석 (잔여 29 skill 재평가 입력)
- [`superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](./superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md) — ADR-001 (N-1 closure)
- [`v0.1-spec.md`](./v0.1-spec.md) — Go CLI v0.1 spec (LOCKED)
- [`skill-map.md`](./skill-map.md) — 11-stage → 9-phase 매핑 (참조용)
- [`DOGFOOD.md`](../DOGFOOD.md) + [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — Go CLI dogfood
- [`notes/2026-05-09-handoff-N1-closure.md`](./notes/2026-05-09-handoff-N1-closure.md) — N-1 closure handoff
