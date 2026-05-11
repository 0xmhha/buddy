# Cycle Handoff — v0.2.0 reset → v0.3.0 ship → W3-3 partial (2026-05-11)

> **목적**: 본 cycle 의 *완료된 작업 + 미구현 / 남은 작업* 을 한 문서에 영속화. 다른 세션이 *처음 5분 안에* 본 cycle 의 종착점을 파악하고 *다음 한 시간 안에* 자연스러운 다음 step 에 진입할 수 있도록.
> **선행 핸드오프**: [`docs/notes/2026-05-11-session-handoff.md`](./2026-05-11-session-handoff.md) (본 cycle 시작 시점 — plugin v1.1.1 patch + Cycle 1 close baseline).
> **본 cycle 범위**: 사용자 요청 "남은 작업을 모두 진행하고, 작업 결과물인 buddy 프로젝트의 기능을 검증하고 나서 release" 부터 시작 → v0.2.0 / v0.3.0 publish + cli buddy spec lock-in + W3-3 minimum-viable subset 까지.

---

## §1. 본 cycle 한 줄 요약

plugin v1.1.1 의 *premature major* 를 v0.2.0 으로 reset (ADR-004) → analytics-mcp full body (W4-2.1 ~ W4-2.6) ship 하여 v0.3.0 release → cli buddy spec lock-in (ADR-005) → cli buddy W3-3 agent runtime 의 minimum-viable subset 까지 도달. **11 commits + 2 git tags + 2 ADRs + 9 source files 신규 + 20 packages race-clean**.

---

## §2. 본 cycle commits (11개, 모두 origin/main push 완료)

| # | commit | 요지 | 영향 영역 |
|---|--------|------|----------|
| 1 | `9ab21dd` | docs: remove 12 superseded notes/plans + cross-refs | `docs/notes/` (-7), `docs/notes/audits/` (-1+dir), `docs/superpowers/plans/` (-4+dir) |
| 2 | `080934b` | release: v1.1.1 → v0.2.0 (ADR-004) | manifest 2 + CHANGELOG + ADR-004 + HANDOFF/tasks/README |
| 3 | `45cefaf` | feat(quality): PROCEDURE skeleton template + lint script (B6) | `plugin/skills/.template/` + `scripts/lint-skill-procedure.sh` + Makefile |
| 4 | `bdf957e` | fix(router): 42 command body backfill + wireup count | `plugin/commands/*.md` × 42 + `scripts/test-router-wireup.sh` |
| 5 | `02c8737` | docs: rename tag plugin-v0.2.0 → v0.2.0 + ADR-004 §2.4 revise | HANDOFF + tasks + ADR-004 |
| 6 | `e20f986` | feat(mcp): register 7 analytics_query_* tool stubs (W4-2.1) | `internal/mcp/analytics_tool.go` (신규) + server.go |
| 7 | `e3b59de` | docs(skills): 10 PROCEDURE MCP integration sections (W4-2.4) | 10 PROCEDURE.md + spec |
| 8 | `139dd2f` | feat(analytics): SQLite reference adapter + 7 handlers (W4-2.2 + W4-2.3) | `internal/analytics/` 5 files (신규) + analytics_tool wiring + cmd/buddy-mcp/main.go |
| 9 | `166486e` | release: v0.2.0 → v0.3.0 (analytics-mcp ship) | manifest 2 + CHANGELOG + SSoT |
| 10 | `4647c32` | docs(adr): cli buddy spec lock-in (ADR-005, Draft → Accepted) | cli-buddy-spec + ADR-005 (신규) + ADR Index + HANDOFF |
| 11 | `846c2cd` | feat(agent): cli buddy W3-3 agent runtime (minimum-viable subset) | migration v4 + `internal/agent/` 7 files (신규) + cmd/buddy/agent.go + json.go |

`origin/main`: `8755e8b → 846c2cd`. Push 완료, 작업 디렉토리 clean.

---

## §3. 완료된 작업 (categorized)

### §3.1 Release artifacts

| Artifact | 결과 |
|----------|------|
| Git tag `v0.2.0` | published 2026-05-11 (commit `bdf957e`) — version reset baseline |
| Git tag `v0.3.0` | published 2026-05-11 (commit `166486e`) — analytics-mcp ship |
| GitHub Release `v0.2.0` | English notes — version reset rationale + namespace 정책 + commits |
| GitHub Release `v0.3.0` | English notes — analytics-mcp 7 tools + SQLite adapter + migration notes |
| marketplace.json + plugin.json | both at v0.3.0 |
| `internal/mcp/server.go` Version constant | `0.3.0` (MCP server self-report) |

### §3.2 ADRs (총 5건, 본 cycle 에서 2 신규)

| ID | Date | Title | Status |
|----|------|-------|--------|
| ADR-001 | 2026-05-09 | Disable model invocation for all `plugin/commands/*.md` | Accepted |
| ADR-002 | 2026-05-10 | roadmap × charter cli buddy gap | Accepted |
| ADR-003 | 2026-05-10 | `superpowers` external attribution policy | Accepted |
| **ADR-004** | **2026-05-11** | **Plugin track version reset v1.x → v0.x** | **Accepted** |
| **ADR-005** | **2026-05-11** | **cli buddy spec scope lock-in (Draft → Accepted)** | **Accepted** |

### §3.3 Code — 신규 / 확장 packages

| Package | 상태 | 책임 |
|---------|------|------|
| **`internal/analytics/`** (신규) | full impl — types / schema / adapter interface / SQLAdapter / 9 integration tests | 7-table SQLite schema + Adapter interface + reference SQLite impl backing 7 MCP tools (funnel / cohort / ab / actor_failure / cost / slo_burn / feedback_corpus) |
| **`internal/agent/`** (신규) | minimum-viable — types / spec / store / executor / runtime / 9 tests | agent CRUD + Subprocess/Mock executor + Runtime.Run(agent) with retry + status transitions + file/stdout output target |
| `internal/mcp/` (확장) | adapter wiring + 4 new tests | `Options.Analytics` field + 7 `analytics_query_*` handlers with stub fallback + adapter-wired integration test |
| `internal/db/` (migration v4) | 3 새 table | `agents` + `agent_runs` + `agent_logs` with FK cascade |
| `cmd/buddy/` (확장) | agent.go + json.go | `buddy agent {create,list,show,run,delete}` 5 subcommands |
| `cmd/buddy-mcp/` (확장) | env-driven adapter | `BUDDY_ANALYTICS_BACKEND` + `BUDDY_ANALYTICS_DSN` env vars → SQLite adapter wire-up |
| `plugin/skills/.template/PROCEDURE.md` (신규) | B6 skeleton | canonical 8-section reference template |
| `scripts/lint-skill-procedure.sh` (신규) | B6 lint | 148 skill 의 form 검증 (Forms A/B/C accepted, 43 deviation 보고) |

### §3.4 Doc cleanup + SSoT updates

| 영역 | 결과 |
|------|------|
| 12 superseded docs 제거 | notes 7 + audits 1 (+ dir) + plans 4 (+ dir) — git history 보존 |
| Cross-ref 정리 | HANDOFF / tasks / 2026-05-11-session-handoff / ADR-001 / ADR-003 / analytics-mcp-spec — orphan ref 0 |
| `docs/cli-buddy-spec.md` | Draft → **Accepted** (ADR-005), §9 W3-3 row partial Done 명시 + 7 deferred follow-on 항목 |
| `docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` | Draft → **Accepted (v0.2.0 — W4-2.1)**, §8 phase table 모두 갱신 (W4-2.1~W4-2.6 Done, W4-2.7 deferred) |
| `docs/HANDOFF.md` | track table v0.2.0 → v0.3.0 + cli buddy 트랙 🟡 부분 → 🟢 본격 진입 |
| `docs/tasks.md` | track summary release column + 작성일 갱신 |
| `docs/superpowers/decisions/README.md` (ADR Index) | ADR-004 + ADR-005 rows + 향후 ADR 후보 갱신 |
| CHANGELOG.md | `[0.2.0]` + `[0.3.0]` entries 본문 + `[Unreleased]` W3-3 entry |

### §3.5 Verification gate (본 cycle 전체)

| 항목 | 결과 |
|------|------|
| `go build ./...` | clean |
| `go vet ./...` | clean |
| `go test -race -count=1 -timeout=120s ./...` | **20 packages PASS** (internal/agent 신규 1 추가) |
| `make test-routing` | 10/10 PASS |
| `make test-skill-form` | 86 pass / 43 deviate / 19 allowlist (불변) |
| End-to-end CLI smoke (`buddy agent create/list/show/delete`) | PASS |
| End-to-end MCP integration (`TestAnalyticsTools_AdapterWiredReturnsJSON`) | PASS |
| Plugin install verification (`/plugin` + `/reload-plugins`) | "buddy 0.3.0 enabled" |

---

## §4. 미구현 / 남은 작업

### §4.1 cli buddy W3 cascade (highest-leverage 다음 작업)

ADR-005 lock-in 이후 *trigger 해제 상태*. spec §9 의 phase 순서:

| Phase | 상태 | 비용 | 다음 자연스러운 step? |
|-------|------|------|-------|
| W3-1 spec | ✅ Accepted | LOW | — |
| **W3-3 follow-on** — background cron scheduler / exponential backoff / streaming log capture / `buddy agent log` / `edit` 등 | 🟡 partial | MED | **Y** (이미 schedule 필드 정의됨, executor 검증됨, 자연 확장) |
| **W3-2 TUI** — `bubbletea` + `lipgloss` 학습 + agent list view + create form | 0% | HIGH | (W3-3 follow-on 과 병행 가능) |
| **W3-4 plugin buddy embedding 본격** — Claude Code subprocess 의 PROCEDURE §6 self-check parse + cascade chain 자동 라우팅 | partial (spawn 부분만) | MED-HIGH | (W3-2 또는 W3-3 따라 trigger) |
| **W3-5 v0.1.0 자산 재배치** — `cmd/buddy/main.go` 685줄 분할 + hook reliability monitor 를 cli buddy sub-feature 로 통합 | 0% | MED | (cli buddy 신규 명령 추가 직전) |
| **W3-6 reference agent (웹툰)** — spec §2.1 의 사용 예시 구현 + W3-2~5 모두 dogfood | 0% | HIGH | (W3-2 ~ W3-5 의존) |

Spec §9 추정: 6 phase × 평균 1~3 week = **3~6 month** single-dev cadence.

### §4.2 Plugin v1.0.0 entry conditions (ADR-004 §2.2)

| # | 조건 | 현 상태 |
|---|------|--------|
| 1 | cli buddy W3-2~W3-6 cascade 완료 | 🟡 W3-1 Done / W3-3 partial / 나머지 0% |
| 2 | plugin+cli 통합 dogfood production-proven (외부 SaaS 1건 이상 end-to-end) | ❌ 미시작 |
| 3 | PROCEDURE 양식 (B6) 통일 + `--strict` lint enforcement | ❌ 43 deviation 잔존, lint 는 report-only |
| 4 | router smart-skip session state (B7) 해결 또는 supersede | ❌ 미해결 |

→ 4 조건 모두 만족 시 v0.x.x → **v1.0.0**. 현재 1/4 부분 진척.

### §4.3 B6 follow-up (PROCEDURE 양식 통일)

- 148 skill 중 **43 deviation** 잔존 (`make test-skill-form` 으로 확인 가능)
- Forms A (Stage skill 8-section, Batch 1~7 신규) / B (Principles skill, methodology) / C (Phase 1-3 era 12-section with Cross-phase cascade) — 3 valid form 인정
- 외부 자산 차용 (gstack CSO, monitoring snippets 등) 의 free-form 일부는 *의도된 보존* — 단순 양식 통일 X
- 별 ADR 후보 (ADR Index 의 *향후 ADR 후보* §B6 row)
- trigger: `--strict` lint 활성 결정 / contributor PR 발생 시

### §4.4 analytics-mcp W4 follow-on

| 항목 | 상태 | trigger |
|------|------|--------|
| W4-2.2 Custom SQL adapter | ✅ SQLite 채택 | PostgreSQL / MySQL 확장은 사용자 production scale 발생 시 |
| W4-2.3 7 handler 본격 | ✅ Done | — |
| W4-2.4 PROCEDURE example | ✅ Done (10 skill) | — |
| W4-2.7 standalone `analytics-mcp-v0.1.0` tag | ⏳ deferred | packaging 결정 시 (ADR-004 milestone 묶음) |
| Real production dogfood | ❌ stub 만 검증 | 사용자 실 events 적재 후 |
| PostgreSQL / MySQL adapter | ❌ | 같은 Adapter interface, driver 만 교체 |

### §4.5 Trigger-bound deferred items (사용자 발화 대기)

| 항목 | trigger |
|------|--------|
| **Korea cluster 3 skill** (`consult-korea-legal-context` / `draft-korea-patent-application` / `audit-korea-cii-vulnerability`) | target market = Korea 결정 시 |
| **USA / EU cluster** | target market = USA 또는 EU 결정 시 |
| **feature-management-mcp** | cli buddy W3-3 agent runtime 완전체 진입 후 (현재 minimum-viable 만 ship) |
| **Go CLI v0.2 dogfood feedback** | 사용자 3~7일 사용 후 dashboard UX (TUI vs web) 결정 |
| **Cycle 2 live dispatch** (5 single-skill + 4 stage cascade) | 별 세션, token 6-8x 권장 |

### §4.6 Always-deferred (현재 trigger 부재)

- ADR-001 의 N-1 후속 quick wins (A-5 — Quick Win C body slim / CONTRIBUTING.md lint)
- gofmt drift 정리 (현재 clean)
- D-2 / D-4 / D-5 open questions (roadmap.md §1 — 멀티-머신 / DAG 시각화 / plugin 권한 경계)

---

## §5. 다음 세션 진입점 — 우선순위

### (a) **W3-3 background scheduler** (Recommended)
- 이미 `agents.schedule` 컬럼 + `AgentSpec.Schedule` 필드 정의됨, 단지 사용 안 됨
- cron 라이브러리 + tick goroutine + Runtime.Run 호출 정도 — 1 cycle 안에 ship 가능
- W3-3 의 자연 확장 + cli buddy 의 "자동화" identity 정합

### (b) **W3-4 PROCEDURE self-check parse 본격**
- SubprocessExecutor 는 stdout/stderr buffer 만 capture — §6 self-check pass/fail 자동 판정 X
- 148 skill 의 self-check 양식이 inconsistent (B6 deviation 43건과 동일 root cause) — B6 follow-on 과 묶을 가치 있음
- 비용: MED-HIGH (양식 분석 + parser + 테스트)

### (c) **W3-2 TUI**
- `bubbletea` + `lipgloss` 학습 비용 HIGH
- agent list view + create form 의 *visible 가치* 큼 — dogfood signal 양산
- 단독으로 3-4 week 추정 — *별 세션 시리즈* 권장

### (d) **B6 follow-up — 43 deviation 통일**
- 외부 자산 보존 vs 통일의 trade-off 결정 필요
- contributor PR 가 발생하기 전까진 *report-only* 로 충분
- 별 ADR 후보

### (e) **plugin v1.0.0 entry condition #2 — production dogfood**
- 사용자 실 프로젝트에 buddy plugin install 후 9-phase cycle 1회 완주
- *사용자 페이스 의존* — AI 가 단독 진행 불가
- production-proven evidence 가 가장 큰 v1.0.0 trigger

→ **(a) W3-3 background scheduler** 가 *AI 단독 진행 + 자연스러운 확장 + 작은 비용* 의 교집합. (e) 가 가장 큰 가치지만 사용자 페이스 의존.

---

## §6. Gotcha — 알아둬야 할 것

### §6.1 두 트랙의 정체성
- **plugin buddy** = Claude Code plugin (148 skills + 99 commands + 7 MCP tools + analytics adapter). v0.3.0 published.
- **cli buddy** = Go binary CLI (`bin/buddy`). v0.1.0 binary published 2026-04-26. cli buddy 의 진짜 목적 (자동화 agent 관리) 은 W3-3 minimum-viable 만 진입. *binary release tag bump 안 됨* — `buddy agent ...` 명령은 unreleased 상태 (`[Unreleased]` CHANGELOG entry).

→ `~/.buddy/buddy.db` 가 두 트랙 *공유* state store. plugin buddy 의 hook events / sessions / features + cli buddy 의 agents / agent_runs / agent_logs 가 같은 SQLite 에 동거. migration v4 가 그 위에 add-on.

### §6.2 Tag namespace 충돌 가능성
- `v0.1.0` = Go CLI binary release (2026-04-26)
- `v0.2.0` = plugin buddy (2026-05-11, version reset baseline)
- `v0.3.0` = plugin buddy (2026-05-11, analytics-mcp)
- 다음 plugin minor bump 는 `v0.4.0`. 다음 cli buddy binary release 도 같은 namespace 공유 (ADR-004 §2.4 revised) — release title 로 artifact 명시 (`v0.4.0 — plugin buddy XXX` vs `v0.2.0 — Go CLI XXX`).

### §6.3 analytics-mcp 의 *opt-in* 정책
- 7 `analytics_query_*` tool 은 *unconditionally register* — `BUDDY_ANALYTICS_BACKEND` 없으면 friend-tone Korean text 응답 (transport error 아님)
- `BUDDY_ANALYTICS_BACKEND=sql` + `BUDDY_ANALYTICS_DSN=<path>` 설정 시 SQLite adapter activate
- 다른 값 (`mixpanel`/`amplitude`/`datadog`/`stripe`/`elasticsearch`) recognised but stubbed — 사용자가 *예상치 못한 fallback* 받지 않도록 명시

### §6.4 SubprocessExecutor 의 dogfood 0
- `Runtime.Run` 의 SubprocessExecutor 가 spawn 하는 `claude` CLI 는 *MockExecutor 로만 검증됨*
- 실제 claude CLI 의 stdin/stdout buffering 동작 / exit code 의미 / 비동기 input 처리 등은 *사용자 실 dogfood 시점에 처음 surface*
- 첫 실사용에서 *payload format* 또는 *prompt 전달 방식* 변경이 필요할 가능성 — 그 경우 `internal/agent/executor.go` 의 `SubprocessExecutor.Run` 만 수정 (인터페이스 안정)

### §6.5 marketplace cache 가능성
- 본 cycle 에서 *3번* version reset / bump (1.1.1 → 0.2.0 → 0.3.0). 사용자가 *오래된 marketplace cache* 가지고 있을 가능성
- 재install 가이드는 본 cycle 중 작성된 instruction (이전 응답 참조). 핵심: `claude plugin marketplace remove buddy && claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy`

---

## §7. References

### 본 cycle 신규 ADR
- [`docs/superpowers/decisions/2026-05-11-plugin-version-reset.md`](../superpowers/decisions/2026-05-11-plugin-version-reset.md) (ADR-004) — version reset rationale + v1.0.0 entry conditions
- [`docs/superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md`](../superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md) (ADR-005) — cli buddy spec scope lock-in

### Spec 산출물
- [`docs/cli-buddy-spec.md`](../cli-buddy-spec.md) — **Accepted** (Draft → Accepted by ADR-005). §9 W3-3 partial Done summary 포함
- [`docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md`](../superpowers/specs/2026-05-10-analytics-mcp-spec.md) — **Accepted (v0.2.0 — W4-2.1)**. §8 phase table 갱신

### Code packages (본 cycle 신규)
- `internal/analytics/` — Adapter interface + SQLite reference impl (5 files + 9 tests)
- `internal/agent/` — Agent runtime engine (6 prod files + 1 helper + 9 tests)
- `internal/db/migrations.go` v4 — agents / agent_runs / agent_logs tables
- `cmd/buddy/agent.go` — 5 subcommands
- `cmd/buddy-mcp/main.go` — env-driven adapter wire-up
- `plugin/skills/.template/PROCEDURE.md` + `scripts/lint-skill-procedure.sh` — B6 infrastructure

### 트랙 SSoT
- [`docs/two-tracks-charter.md`](../two-tracks-charter.md) — plugin / cli buddy 정체성 (변경 없음)
- [`docs/HANDOFF.md`](../HANDOFF.md) — track status table (본 cycle 갱신 — Last updated 2026-05-11 v0.3.0)
- [`docs/tasks.md`](../tasks.md) — cross-track 작업 인벤토리 (본 cycle 갱신)

### Release tags
- https://github.com/0xmhha/buddy/releases/tag/v0.2.0 — plugin buddy version reset baseline
- https://github.com/0xmhha/buddy/releases/tag/v0.3.0 — plugin buddy analytics-mcp ship

---

## §8. 다음 세션 entry quick start

```bash
# 1. 최신 main + tags
git fetch origin --tags && git log --oneline -5 main
# 기대: 846c2cd 가 최신

# 2. 빌드 + 테스트
make build
go test -race -count=1 -timeout=120s ./...
# 기대: 20 packages PASS

# 3. quality gates
make test-routing      # 10/10
make test-skill-form   # 86 pass / 43 deviate

# 4. 본 핸드오프 + 최신 ADR 읽기
cat docs/notes/2026-05-11-cycle-handoff.md  # 이 문서
cat docs/superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md  # ADR-005
cat docs/cli-buddy-spec.md  # cli buddy roadmap (Accepted)

# 5. 권장 다음 step — W3-3 background scheduler
ls internal/agent/  # 현재 state — runtime.go 의 Run() 은 on-demand 만
# 스케줄러 추가 시:
#   - internal/agent/scheduler.go (cron parsing + tick loop)
#   - daemon-style 또는 `buddy agent schedule start` 명령
#   - go test 추가
#   - schedule 필드 활성화 (현재 informational)
```

---

> 본 문서가 정확하지 않으면 *이 문서부터 업데이트*. 다음 세션이 의지하는 본 cycle 의 SSoT.
