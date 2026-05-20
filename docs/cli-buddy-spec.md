# cli buddy — Spec

> **목적**: charter §6.2 의 *cli buddy 진화 1 순위* 작업 — *진짜 목적 (자동화 agent 관리) 의 spec 작성*. 본 문서는 cli buddy 의 *책임 / 구성 / runtime / 인터페이스 / lifecycle* 의 lock-in spec.
>
> **Status**: **Accepted (2026-05-11 — ADR-005)**. Lock-in 항목: §1.3 4-책임 scope / §4.1 option (a) Claude Code subprocess / §5.1 v0.1.0 자산 재배치 / §7 ADR-002 actual rewrite trigger / §9 6-phase 분할 (3-6 month estimate). §10 Open questions Q-1~Q-6 deferred — trigger 시점 별 ADR. W3-2 (TUI 설계) → W3-3 (agent runtime) → W3-4 (plugin buddy 내재화) → W3-5 (v0.1.0 재배치) → W3-6 (reference impl) cascade 가 *trigger 해제 상태* — 사용자 페이스 따라 진입.
>
> **Predecessor**: [`docs/two-tracks-charter.md`](./two-tracks-charter.md) §3 / §6.2
> **Lock-in ADR**: [`docs/superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md`](./superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md) (ADR-005)

---

## 0. 한 줄 정의 (charter §3 인용)

> "cli buddy 는 어떤 용도냐면, 'plugin buddy' 를 활용하여 자동화 agent 로 동작할수 있도록 지원하는 툴이다. 자동화된 agent 를 관리하고, 실행 및 종료 시키고, 설정을 변경하는등을 지원하는 툴이다. cli buddy 는 tui 로 화면을 지원하면서, 여러 자동화된 agent 를 설정하고 관리하는 툴"

핵심 4 책임: **agent 생성 / 실행 / 종료 / 설정 변경**.

---

## 1. 정체성 + scope

### 1.1 무엇인가

| 차원 | 정의 |
|------|------|
| 형태 | Go binary CLI (`bin/buddy` — 기존 v0.1.0 자산 재사용) |
| 인터페이스 | TUI (Terminal User Interface) — `charmbracelet/bubbletea` + `lipgloss` |
| 의존 | plugin buddy 를 *내재화* (charter §3.1, §4.1) |
| 분배 | release binary (linux / darwin × amd64 / arm64) — 기존 GitHub Actions release workflow 재사용 (M6 자산) |

### 1.2 핵심 가치

`plugin buddy 의 skill / MCP / agent / hook` 4 자산을 *자동화 agent 안에서 활용* → 사용자 *manual trigger 없이* 작업 진행.

### 1.3 Scope (9 책임 — ADR-009 2026-05-19 확장)

> **확장 컨텍스트**: ADR-005 lock-in 시점 (2026-05-11) 의 4 책임은 *자동화 agent 관리* 단독. ADR-009 (2026-05-19) 가 *AI-usage coaching* 5 영역을 추가하여 cli buddy 의 정체성을 "automation agent + AI 사용 코치" 로 확장.

#### A. 자동화 agent 관리 (v0.1.0 ~ v0.7.5 ship 완료)

| 책임 | 정의 |
|------|------|
| **agent 생성** | spec 입력 → agent 정의 + 등록. 설정 (스케줄 / 입력 / plugin buddy 호출 패턴) 포함 |
| **agent 실행** | 등록된 agent 를 백그라운드 / 스케줄 / on-demand 실행. plugin buddy 내재화 layer 통과 |
| **agent 종료** | 실행 중 agent 정지 / 폐기 / 자원 회수 |
| **agent 설정 변경** | 스케줄 / 입력 / plugin buddy 호출 패턴 / 출력 destination 등 |

#### B. AI-usage coaching (ADR-009, 미구현 — 5 영역)

| 책임 | 정의 | 현 fragment | impl trigger |
|------|------|-------------|--------------|
| **F2.A Session Monitor** | Claude Code 세션(들)의 lifecycle / token / message / 시간 추적 | `internal/sessions/` + migration v3 (passive registry) | Claude Code session-lifecycle hook 안정 |
| **F2.B Usage Analysis** | 활동 ledger → 사용 패턴 / 마찰 지점 / 시간 분포 분석 (로컬, privacy-preserving) | `internal/analytics/` 7 MCP tools (v0.3.0, 현재 generic 분석) | F2.A 가 ~30일 데이터 누적 |
| **F2.C Advisory** | 분석 → actionable 한국어 prose 권고 ("이 hook 12% 실패, 설정 X 권장") | none | F2.B 의 analytic primitive 안정 |
| **F2.D Drift Detection** | 단일 conversation 안 *원래 목적* vs *현재 turn* 의 semantic drift 감지 | none | F2.A turn-level 데이터 + LLM-driven 비교 path 안정 |
| **F2.E Notification** | F2.C / F2.D 의 메시지를 TUI banner / desktop / shell prompt / webhook 으로 전달 | `output.type: webhook` (machine-to-machine, 다른 개념) | F2.C 가 out-of-band 전달 가치 있는 advisory 생성 |

→ B 영역의 자세한 design 은 영역별 후속 ADR (F2.A ~ F2.E) 에서. 본 spec 은 책임 *명시* 까지.

### 1.4 Non-goals

- plugin buddy 의 skill / MCP / agent / hook 자산 *추가 / 변경* — plugin buddy 의 책임
- 외부 SaaS 의 *직접 통합* (예: Twitter API, OpenAI API) — agent 의 책임 (cli buddy 는 *agent runtime + 관리*)
- 사용자 데이터 *영속 저장* — agent 가 자체 storage 에 (cli buddy 는 *metadata + 실행 log* 만)
- 다중 사용자 협업 — single-user 도구 (multi-user 는 v2.0+ 검토)

---

## 2. 사용 예시 — charter §3.5 인용

### 2.1 웹툰 그리기 agent

> "웹툰 그리기 agent 를 생성한다고 할때, 'plugin buddy' 기능을 활용하여, buddy cli 가 agent 를 생성하고, agent 를 자동으로 실행해두고 관리한다. agent 는 웹툰의 세계관등의 작업을 정리하고, 매일 연재할 스토리를 정리하고, 그림을 생성하여 webtoon-by-ai 서비스에 특정 시간에 배포하여 작품을 노출하고, 유저들은 이것을 감상할수 있도록 할 수 있다."

### 2.2 흐름 도식

```
[사용자] —/—> [cli buddy TUI]
                    │
                    ▼
            "웹툰 그리기 agent 생성"
                    │
                    ▼
       [agent runtime 등록 + spec lock-in]
       agent.spec:
         - identity: webtoon-creator
         - schedule: daily 03:00 KST
         - plugin buddy skill chain:
             concretize-idea → write-prd
             → design-system → build-feature
         - output: webtoon-by-ai service publish API
                    │
                    ▼ (특정 시간 trigger)
       [agent 실행 — plugin buddy 내재화 layer 호출]
            │
            │  내부에서 plugin buddy 호출:
            │    /buddy:concretize-idea (세계관 정리)
            │    /buddy:define-product-spec (오늘의 스토리)
            │    [외부 image gen MCP] (그림 생성)
            │    [webtoon-by-ai publish API] (배포)
            │
            ▼
       [webtoon-by-ai service 작품 노출]
            │
            ▼
       [유저 감상]
```

### 2.3 다른 agent 예시 (참고)

| agent | 스케줄 | plugin buddy 활용 |
|-------|------|----------------|
| 일일 retro 작성 | 매일 18:00 | `/buddy:summarize-retro` + `/buddy:generate-improvement-tasks` |
| 마케팅 콘텐츠 자동 발행 | 매일 09:00 | `/buddy:draft-marketing-copy` + `/buddy:automate-marketing-content` |
| 경쟁사 모니터링 | 매주 월요일 | `/buddy:analyze-competition-and-substitutes` |
| 이슈 triage | 1시간 마다 | `/buddy:triage-customer-support-ticket` + `/buddy:generate-improvement-tasks` |
| 비용 anomaly 감시 | 1일 1회 | `/buddy:analyze-cost-anomaly` |

---

## 3. 아키텍처

### 3.1 4 sub-component

```
┌────────────────────────────────────────────────────────┐
│                  cli buddy (binary)                    │
│                                                        │
│  ┌──────────────┐   ┌──────────────┐   ┌─────────────┐│
│  │  TUI         │   │ agent runtime│   │ plugin buddy││
│  │ (bubbletea)  │   │ (scheduler   │   │ embedding   ││
│  │              │   │  + executor) │   │ layer       ││
│  └──────┬───────┘   └──────┬───────┘   └──────┬──────┘│
│         │                  │                   │       │
│         └──────────────────┴───────────────────┘       │
│                            │                           │
│                  ┌─────────┴────────┐                  │
│                  │  state store     │                  │
│                  │  (SQLite, 기존)  │                  │
│                  └──────────────────┘                  │
└────────────────────────────────────────────────────────┘
                              │
                              │ Claude Code session 호출
                              ▼
┌────────────────────────────────────────────────────────┐
│                Claude Code session                     │
│  with plugin buddy installed                           │
│                                                        │
│  ┌──────────────┐   ┌──────────────┐                   │
│  │  router      │ → │  PROCEDURE   │                   │
│  │  skill       │   │  load + exec │                   │
│  └──────────────┘   └──────────────┘                   │
└────────────────────────────────────────────────────────┘
```

### 3.2 Sub-component 책임

| component | 책임 |
|----------|------|
| **TUI** | agent list 표시 / 생성 form / 실행 status / log tail / 설정 편집 |
| **agent runtime** | scheduler (cron / on-demand) + executor (plugin buddy 호출 + 결과 capture + retry) |
| **plugin buddy embedding layer** | Claude Code session spawn + plugin buddy command 호출 + 결과 parse |
| **state store** | agent metadata + 실행 log + 결과 cache (기존 v0.1.0 SQLite 재사용) |

### 3.3 데이터 모델

| 테이블 | 책임 |
|------|------|
| `agents` | agent 정의 (id, name, spec_yaml, schedule, status, created_at, last_run) |
| `agent_runs` | 각 실행 instance (agent_id, started_at, ended_at, exit_code, log_path, plugin_buddy_calls) |
| `agent_logs` | 실행 log (run_id, timestamp, level, message) — partition by run_id, retention TBD |
| `hook_events` | 기존 v0.1.0 — agent 안의 plugin buddy 호출 시 hook 도 capture |

→ 기존 v0.1 schema 재사용 + 3 테이블 추가 (migration v4 / v5 / v6).

---

## 4. plugin buddy embedding layer

### 4.1 내재화 패턴

cli buddy 가 plugin buddy 를 호출하는 4 옵션:

| 옵션 | 방법 | 장점 | 단점 |
|------|------|------|------|
| **(a) Claude Code subprocess** | `claude` CLI 호출 + stdin/stdout 통신 | 단순 / claude code 정합 | claude 인증 / API key 의존 |
| (b) MCP server stdio | plugin buddy 의 MCP server (deferred) 직접 호출 | 인증 minimize | MCP server 구현 의존 |
| (c) 직접 PROCEDURE.md parse + 자체 LLM 호출 | Anthropic API direct | claude code 의존 0 | LLM 호출 비용 자체 부담 |
| (d) Hybrid | basic = (a), advanced = (b) | 유연 | 두 path 유지 부담 |

**권장 (Spec lock-in)**: **(a) Claude Code subprocess** — 가장 단순 + claude code 사용자 가정 + 기존 인증 활용.

### 4.2 실행 흐름

```
agent.run():
  1. agent spec 의 plugin buddy command chain load
  2. for each command in chain:
     a. spawn `claude` subprocess (혹은 reuse session)
     b. send `/buddy:<command> "<args>"` via stdin
     c. capture output + structured result
     d. parse PROCEDURE 산출물 (12 section format)
     e. validate self-check (§6 of PROCEDURE) pass / fail
     f. if fail → retry (exponential backoff) or abort
  3. aggregate result → output destination
  4. 결과 + log + metric → state store
```

### 4.3 결과 parse + 검증

PROCEDURE.md 의 §5 산출물 형식 + §6 self-check 가 *machine-parseable* 영역이라 cli buddy 가:
- 산출물 markdown → structured (yaml / json) 변환
- self-check 의 boolean checklist → pass/fail 판정
- next phase suggestion 자동 cascade

---

## 5. 구성 요소 — 현재 자산 재배치

### 5.1 v0.1.0 sub-feature 재배치 결정

charter §3.6 가 명시:
> "v0.1.0 = hook reliability monitor (한 sub-feature 만 구현)"

→ v0.1.0 의 hook reliability monitor 를 *cli buddy 의 sub-feature* 로 재배치:

| v0.1.0 자산 | cli buddy 안의 위치 |
|-----------|-------------------|
| `cmd/buddy/main.go` (685 lines) | TUI entry + subcommand dispatcher (분할 — W6-1) |
| `internal/db/`, `internal/schema/` | state store (재사용 + 3 신규 테이블) |
| `internal/daemon/`, `internal/aggregator/` | hook reliability monitor sub-feature (별도 mode 로 진입) |
| `internal/install/`, `internal/diagnose/` | install + diagnose subcommand (cli buddy 자체 install / diagnose) |
| `internal/persona/` | TUI message catalog (en/ko) |
| `internal/sessions/`, `internal/pricing/` | (이전 v0.2 시도 — cli buddy agent 의 *Claude Code session 모니터링* 영역에 통합 가능) |
| `cmd/buddy-mcp/` | (별도 — plugin buddy 의 MCP server, cli buddy 와 책임 다름) |

### 5.2 신규 추가 component

| component | 위치 | 비용 |
|----------|------|------|
| TUI views (agent list / create / status / log) | `internal/ui/` | HIGH (bubbletea 학습 + UX 설계) |
| agent runtime (scheduler + executor) | `internal/agent/` | HIGH (cron 기반 + retry + error handling) |
| plugin buddy embedding layer | `internal/embedding/` 또는 `internal/claudecode/` | MED-HIGH |
| 신규 SQLite migration (agents / agent_runs / agent_logs) | `internal/db/migrations.go` | MED |
| agent spec format (YAML) + parser | `internal/spec/` | LOW |
| reference agent — 웹툰 (W3-6) | `examples/webtoon-agent/` | HIGH (별도 cycle) |

---

## 6. CLI subcommand surface (예상)

기존 v0.1.0 의 7 subcommand + cli buddy 신규 7 subcommand:

### 6.1 v0.1.0 subcommand (재사용)

```
buddy install        # cli buddy 자체 install + plugin buddy 자동 설치 안내
buddy uninstall      # cli buddy + (옵션) plugin buddy 제거
buddy daemon start   # hook reliability monitor sub-feature
buddy daemon stop
buddy doctor         # cli buddy + plugin buddy 상태 진단
buddy stats          # hook events stats (기존)
buddy events         # hook events stream
```

### 6.2 cli buddy 신규 subcommand

```
buddy tui            # TUI 진입 (default 또는 -t)
buddy agent create   # agent spec 작성 + 등록
buddy agent list     # 등록 agent 목록
buddy agent show <id>  # detail
buddy agent run <id>   # 즉시 실행
buddy agent stop <id>  # 실행 중 정지
buddy agent edit <id>  # spec 편집
buddy agent delete <id>
buddy agent log <id>   # 실행 log
buddy schedule list  # 등록 cron schedule 목록
```

→ 기존 v0.1.0 명령 호환 유지 + cli buddy 신규 명령 추가. semver minor bump (v0.2.0).

---

## 7. ADR-002 의 actual rewrite (W5)

cli buddy spec 작성 자체가 *roadmap.md §4/§5/§6 outline 의 actual rewrite trigger* (ADR-002 §3.2).

### 7.1 v0.2 outline (P1 + P2 항목 cli buddy 안에 흡수)

| roadmap §4 v0.2 항목 | cli buddy 안에서 |
|------------------|-------------|
| T1 활성 세션 발견 (`internal/sessions/`) | agent_runs 의 *Claude Code session id* 캡처 |
| T2 token usage parser | agent_runs 의 token / cost 측정 |
| T3 multi-session stats | agent 단위 stats (`buddy stats --by-agent`) |
| T4 dashboard UI (TUI vs web) | TUI 채택 (charter §3.1) |
| T5 cost estimate | agent 별 cost (pricing table 재사용) |
| T6 i18n full split | persona en/ko sweep (W7-2 와 묶음) |

→ 모두 *cli buddy 의 sub-feature* 로 흡수.

### 7.2 v0.3 outline (P1 항목 흡수)

| roadmap §5 v0.3 항목 | cli buddy 안에서 |
|------------------|-------------|
| T1 task DAG schema | agent spec 의 *task chain* 으로 변형 (task = agent 안의 step) |
| T2 `buddy task` CLI | rename → `buddy agent run` (P3 단어 충돌 해소) |
| T3 wave 그룹화 | agent 의 parallel step 실행 |
| T4 retry policy | agent runtime 의 retry (exponential backoff) |
| T5 외부 task tracker | (deferred) Linear / Jira / GitHub integration 별도 cycle |

### 7.3 v1.0 outline (P3 단어 충돌 해소)

| roadmap §6 v1.0 항목 | cli buddy 안에서 |
|------------------|-------------|
| T1 AGENTS.md auto-sync | *AGENTS.md = capability metadata* (Anthropic 표준) — cli buddy 의 *agent (자동화 실행체)* 와 별개. cli buddy 가 사용자 프로젝트의 AGENTS.md 갱신은 *plugin buddy 영역* (별도 skill 후보) |
| T2 plugin model | cli buddy 의 *plugin buddy 내재화 layer* (§4) — 다른 의미. 외부 plugin 로드는 deferred |
| T3 MCP server | cli buddy 가 *MCP server 로 노출* — agent runtime 을 외부에서 호출 가능 (advanced 영역, deferred) |
| T4 cross-harness | 별도 spec — cli buddy v2.0+ |

---

## 8. Acceptance criteria — cli buddy v0.2.0 release gate

| 항목 | gate |
|------|-----|
| TUI 진입 | `buddy tui` 명령으로 즉시 entry, agent list 표시 |
| agent CRUD | create / list / show / edit / delete 모두 동작 |
| agent 실행 | scheduler (cron) + on-demand 실행 + 결과 capture |
| plugin buddy 호출 | Claude Code subprocess 정상 spawn + PROCEDURE 산출 parse |
| reference agent | 웹툰 agent (W3-6) 동작 — 매일 03:00 trigger + plugin buddy chain 호출 + log 기록 |
| state store | 3 신규 테이블 migration + race-clean test |
| Backward compat | 기존 v0.1.0 7 subcommand 모두 동작 |
| Quality gates | `go test -race -count=2 ./...` clean / `go build ./...` clean / `go vet` clean |
| Documentation | README 갱신 (TUI screenshot + cli buddy quickstart) + CHANGELOG entry |

---

## 9. Phase 분할 — implementation sequence

| Phase | skill 작성 단위 | 비용 추정 | 의존 |
|-------|------------|--------|------|
| W3-1 spec | 본 문서 | LOW (Done) | — |
| W3-2 TUI | bubbletea 학습 + agent list / create form | **Done 2026-05-17** — minimum-viable + 6/6 named follow-on items shipped: list (j/k nav + AltScreen lifecycle), detail (Enter/l → LatestRun summary), scheduler-preview (`s` → `agent.PreviewSchedule` next-fire per scheduled agent, decoupled from running scheduler), in-app delete (`d` → y/N confirm → `Store.Delete` + FK cascade), live log-tail (`t` from detail → `Store.LogsSince` ~1s `tea.Tick` polling, self-cancels on mode change), in-app spec edit (`e` from detail → `tea.ExecProcess` shell-out to `$EDITOR`/`$VISUAL`/`vi` → `ParseSpec` + rename guard + `Store.UpdateSpec`), create form (`c` from list → same shell-out pattern on `CreateStarterYAML` → `Store.Create`). Deferred follow-on-of-follow-on items: scheduler pane live "currently running" indicator (requires Scheduler instance coupling); log-tail scrollback + auto-stop on run-end. | W3-1 |
| W3-3 agent runtime | scheduler (cron) + executor + retry | HIGH (partial Done 2026-05-11 — minimum-viable subset) | W3-1 |
| W3-4 plugin buddy embedding | Claude Code subprocess + PROCEDURE parse | **Done 2026-05-17** — parser ship in v0.5.0 + conditional branches ship in v0.6.0 + retry/fail semantics shipped 2026-05-17 (`chain[].continue_on_fail` per-step flag, default false preserves v0.6.x fail-fast) + auto-cascade shipped 2026-05-17 (`auto_cascade: {max_depth:N}` chain-level config; runtime queue-based loop appends `NextPhase.Skills[0]` after every successful step, capped at `DefaultCascadeMaxDepth=5`; skip-on-failure; `StepResult.CascadeDepth` tracks original-vs-cascaded). Branch-aware selection (vs Skills[0]) is the remaining sub-item, deferred pending dogfood signal. | W3-1 (Subprocess executor 의 spawn 부분 W3-3 안에서 ship) |
| W3-5 v0.1.0 재배치 | main.go 분할 + sub-feature 재배치 | **Done 2026-05-18** — main.go 685→147 lines split shipped v0.6.3 (6 sibling files in `cmd/buddy/`); hook reliability monitor cli buddy 통합 shipped 2026-05-18 as `buddy tui` `H` key → ModeHookStats pane (`internal/queries.Run` wrapped via `tui.HookStatsFetcher` injection, renders count / failures / p50 / p95 per hook, same column order as `buddy stats` CLI). 기존 `buddy daemon/stats/events` CLI surface 그대로 — integration 은 consumer (TUI) layer 에서만 추가 (additive). | W3-2 / W3-3 / W3-4 |
| W3-6 reference agent | 웹툰 agent example | HIGH (Done 2026-05-12 — `examples/webtoon-agent/spec.yaml` + README; ParseSpec regression test gates the example. Exercises Tier 1.4 backoff / 1.5 streaming / 1.6 scheduler refresh / 1.8 webhook + W3-3 chain) | W3-2 ~ W3-5 |

### W4 ~ W8 — AI-usage coaching (ADR-009)

| Phase | 영역 | 비용 추정 | 의존 |
|-------|------|---------|------|
| **W4 F2.A Session Monitor** | ✅ v0.8.0 ship (2026-05-19) — ADR-012 Hybrid hook+fsLister + schema v5 + CLI list/show/purge + daemon poll | — | Done |
| **W5 F2.B Usage Analysis** | ✅ v0.9.0 ship (2026-05-19) — ADR-013 sessions-only live aggregation + 7 metric + CLI `buddy usage` + 5 MCP `usage_query_*` + TUI Usage pane | — | Done |
| **W6 F2.C Advisory** | 3-phase split per ADR-014/015. **Phase 1 v0.10.0 ✅** (knowledge retrieval foundation). **Phase 2 v0.11.0 ✅** (advisor — 5 rules + CLI `buddy advise` + TUI Usage advisory section + MCP `usage_advise` + daemon advisorMonitor goroutine + advisories table v7; closes C-3). Phase 3 v0.12.0 post-v1.0 (skill autogen). | — | Done (Phase 1+2) |
| **W7 F2.D Drift Detection** | 단일 conversation 안 *원래 목적* vs *현재 turn* semantic similarity 평가 + drift alert 생성 | HIGH | W4 turn-level 데이터 + LLM-driven 비교 path |
| **W8 F2.E Notification** | ✅ v0.12.0 ship (2026-05-20) — ADR-016 4-channel (desktop osascript/notify-send + webhook + TUI banner + shell prompt) + daemon auto-dispatch + per-channel severity floor + dedup + notification_log v8 | — | Done |

→ W4 ~ W8 은 ADR-009 vision 의 *5 영역 = 5 phase*. W4/W5 ship 완료, W6~W8 잔여.

### W3-3 partial Done — 2026-05-11 ship summary

**v0.3.x 안에 ship 한 minimum-viable subset**:
- `internal/db/migrations.go` v4 — agents / agent_runs / agent_logs 테이블 + FK cascade
- `internal/agent/` package — types / spec(YAML) / store(CRUD + log) / executor(Subprocess+Mock) / runtime(retry + output target)
- `cmd/buddy/agent.go` — `buddy agent create | list | show | run | delete` 5 subcommands
- 9 race-clean tests (`internal/agent/runtime_test.go`)

**W3-3 follow-on 추가 ship (2026-05-11, 동일 cycle)**:
- ✅ background cron scheduler — `internal/agent/scheduler.go` (`Scheduler` + per-agent in-flight guard + sequential dispatch) + `buddy agent scheduler {start,status}` CLI. Spec.schedule 필드가 cron expression 일 때 활성. robfig/cron/v3 dependency (whole-second precision — sub-second `@every` 비활성).

**W3-4 partial Done (2026-05-11, v0.5.0)**:
- ✅ PROCEDURE output parser — `internal/agent/parser.go` (`ParseClaudeOutput`) extracts §self-check verdict (pass / fail / pending / unknown) + per-item checklist detail + §next-phase skill candidates. Pattern-matches Form A (`## 6. 검증`) / Form C (`## 11. Verification gate`) / English aliases without requiring 148 PROCEDURE rewrites.
- ✅ `StepResult.Parsed` 필드 추가 — 각 step 의 self-check + next-phase metadata 가 `agent_runs.result_json` 에 포함.
- ✅ Runtime log line — self-check verdict + count + next-phase candidates 를 `agent_logs` 에 기록.

**W3-4 follow-on Done (2026-05-12, v0.6.0)**:
- ✅ Conditional next-phase branches — `NextPhase.Branches []NextPhaseBranch` captures `- <cond> → \`skill\`` style cascade rules (Hangul / English / mixed). Detection guards: backtick-leading bullets rejected, LHS ≤30 runes (char-count, not bytes), ASCII `->` + Unicode `→` both match.
- ✅ Per-branch runtime log line — `next-phase branch: "<cond>" → <skills>` (or `→ (no skill)` when RHS has no backtick), preserves the v0.5.0 union `next-phase candidates: <list>` line.
- ⏳ deferred (W3-4 follow-on remaining): self-check fail 시 step retry / abort 의미 변경, next-phase auto-cascade with branch-selection policy (env-var / CLI flag / interactive prompt — design pending).

**남은 W3-3 follow-on**:
- exponential backoff (현재 fixed `backoff_delay`)
- streaming log capture (현재 stdout/stderr buffer 전체만)
- PROCEDURE §6 self-check parse (W3-4 본격 — pass/fail 자동 판정)
- production-proven dogfood (사용자 실 agent 등록 시)
- webhook / API output target (현재 stdout / file 만)
- `buddy agent log <id>` / `buddy agent edit <id>` 등 추가 subcommand
- ~~scheduler live-refresh (현재 startup 시 load only — agent 추가/삭제는 restart 필요)~~ ✅ Done (Tier 1.6, 2026-05-12 — `RefreshInterval` polling + `RefreshDisabled` opt-out; CLI: `--refresh <duration>` / `--no-refresh`)

→ 6 phase × 평균 1~3 week = **3~6 month** estimate (single-dev cadence).

---

## 10. Open questions

| ID | 질문 | 결정 시점 |
|----|------|--------|
| Q-1 | TUI 외 *web dashboard* 추가? | 첫 dogfood 후 |
| Q-2 | agent 의 *AI 모델 선택* (사용자가 결정 vs 자동) | W3-3 시 |
| Q-3 | agent log retention 정책 | W3-3 시 |
| Q-4 | multi-machine agent 분산 | v2.0+ |
| Q-5 | agent 간 *통신 / data 공유* (한 agent 산출 → 다른 agent 입력) | W3-6 reference 작성 시 |
| Q-6 | sandbox / security 정책 (agent 가 *임의 명령 실행* 위험) | W3-3 / W3-5 시 |

---

## 11. 다음 액션

1. 본 spec lock-in (사용자 confirm 후 Status = Accepted)
2. ADR 작성 — *cli buddy spec scope decision* (예상 ADR-{N+1})
3. Roadmap 의 v0.2/v0.3/v1.0 outline 의 *재평가 마킹* 갱신 (ADR-002 trigger 발화) — §7 본 문서 안 매핑 따라 actual rewrite
4. W3-2 TUI 설계 진입 — bubbletea 학습 + 첫 view (agent list)
5. plugin buddy v1.1.0 release 와 별개 — cli buddy 는 v0.1.0 → v0.2.0 path

---

## 12. References

- [`docs/two-tracks-charter.md`](./two-tracks-charter.md) §3 / §4.1 / §6.2
- [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./superpowers/decisions/2026-05-10-roadmap-charter-gap.md) (ADR-002)
- [`docs/roadmap.md`](./roadmap.md) §4 / §5 / §6 (재평가 필요 마킹)
- [`docs/v0.1-spec.md`](./v0.1-spec.md) (v0.1.0 자산)
- charmbracelet/bubbletea / lipgloss / x/term — TUI stack (ai-m precedent)
- 0xmhha/cli-wrapper — supervision 패턴 (기존 v0.1 의존)

---

> 본 spec 은 **Accepted** (2026-05-11). ADR-005 (`docs/superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md`) 가 lock-in 근거. W3-2 ~ W3-6 cascade 의 trigger 해제 — 사용자 페이스 따라 진입.
