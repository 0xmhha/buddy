# Buddy — Session Handoff

> 다른 세션에서 이 프로젝트를 이어 받는 사람(또는 미래의 자기 자신)이 *처음 5분 안에* 어디까지 와있는지 파악하고, *다음 한 시간 안에* 일을 재개할 수 있도록 만든 문서.

**Last updated:** 2026-05-20 (**v0.12.0** Wave 7 진행) — v0.8.0 W7-1 + v0.9.0 W7-2 + v0.10.0 W7-3a + v0.11.0 W7-3b + **v0.12.0 W7-5 F2.E Notification (ADR-016 — closes C-5)**. **cli-buddy-spec §9 W4/W5/W6/W8 Done.** v1.0.0 entry 9 조건 중 **7/9 closed** (B-1/B-3/B-4 + C-1/C-2/C-3/C-5); 남은: B-2 + C-4 (Drift) — *진척률 ~83%*. **세션 인계 진입점: 본 문서 §1 + `BACKLOG.md` + `CHANGELOG.md [0.12.0]`**.)

## 트랙 상태

> 이 repo는 두 트랙이 공존한다. **현재 능동 트랙은 plugin buddy.** 트랙 정체성 / 책임 경계의 SSoT는 [`docs/two-tracks-charter.md`](./two-tracks-charter.md).

| 트랙 | 상태 | 위치 | Entry doc |
|------|------|------|----------|
| **plugin buddy** — Claude Code plugin (skill / MCP / agent / hook 카탈로그). 9-phase orchestrator, **148 procedures, 99 commands**, single-router dispatch | 🟢 ACTIVE — Skill Completion Cycle 100% (44/44 신규 + 통합 2). charter scope 12 stage 100% cover. 다음 후보: dogfood 검증 + Korea cluster / analytics-mcp 의 trigger 발화 시 deferred 작성 | `plugin/`, `docs/superpowers/` | [`docs/two-tracks-charter.md`](./two-tracks-charter.md) §2 + [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) |
| **cli buddy** — TUI 자동화 agent 관리 툴 (plugin buddy 내재화 + hook reliability monitor 통합). 진짜 목적은 *agent 생성 / 실행 / 종료 / 설정 관리*. | 🟢 **cli-buddy-spec §9 W3-1~W3-6 모두 Done (2026-05-19)** — W3-1 spec (ADR-005) + W3-2 TUI 6/6 (list/detail/scheduler-preview/in-app delete/log tail/in-app edit/create form) + W3-3 agent runtime + scheduler (v0.4.0/v0.6.5) + W3-4 parser + conditional branches + continue_on_fail + auto_cascade (v0.5/v0.6/v0.7.1) + W3-5 main.go split (v0.6.3) + hook monitor TUI 통합 (v0.7.2 H key → ModeHookStats pane) + W3-6 reference webtoon agent example (v0.6.6). `internal/agent/` (types/spec/store /executor/runtime/scheduler/parser/PreviewSchedule) + migration v4 + `buddy agent {create,list,show,run,log,delete,scheduler,purge}` + `buddy tui` (7 modes: list/detail/scheduler/delete-confirm/log-tail/hook-stats + edit/create shell-out). Remaining cascade work = v1.0.0 entry B-2 (production dogfood, user-paced) 만. | `cmd/`, `internal/`, `archive/ts-poc/` | [`docs/two-tracks-charter.md`](./two-tracks-charter.md) §3 + [`docs/cli-buddy-spec.md`](./cli-buddy-spec.md) (Accepted) |

**plugin buddy 진행 상태 (2026-05-11):**

> **Version note**: 아래 Phase 컬럼의 `v1.0.x` ~ `v1.0.8` 는 *역사적 작업 단위* — 어느 publish 도 안 됐고 현재 marketplace.json 은 **v0.4.0** (ADR-004 §1.4 + analytics-mcp + cli buddy W3-3 minor bumps). 각 Phase 의 commit 은 git history 에 있으나 *release tag 0건* (v1.x 시점 기준).

| Phase | 상태 | 작업 단위 (release X) | Skill count |
|-------|------|---------------------|-------------|
| Phase 0 (foundation: routing infra + CI) | ✅ Done | (was tagged v1.0.0, never published) | 78 baseline |
| Phase 1 (§3 Technical Design 4) | ✅ Done | (was v1.0.2) | +4 |
| Phase 2 (§4 Implementation Plan 6) | ✅ Done | (was v1.0.3) | +6 |
| Phase 3 (§7 Release Safety Nets 7) | ✅ Done | (was v1.0.4) | +7 |
| Phase 4 (§6 Use-case Test 2) | ✅ Done | (was v1.0.5) | +2 |
| Phase 5 ext Cluster C (§3 cascade bridge 2) | ✅ Done | (was v1.0.6) | +2 |
| Phase 5 ext Cluster A (§6 launch readiness 3) | ✅ Done | (was v1.0.7) | +3 |
| Phase 5 ext Cluster B (§3 SaaS pattern 3) | ✅ Done | (was v1.0.8) | +3 |
| **N-1 closure** (ADR-001: `disable-model-invocation: true` × 57 commands) | ✅ Done | — | — |
| **Skill Completion Cycle Batch 1~7** (44 신규 + 2 통합) | ✅ Done | (was v1.1.0 candidate) | +44 (+2 갱신) |
| **Cycle 1 dogfood fix** (B1-B5) + Cycle 2 structural | ✅ Done | (was v1.1.1 patch) | — |
| **v0.2.0 release prep** (doc cleanup 12 + B6 lint/template + ADR-004) | ✅ Done | v0.2.0 (2026-05-11, tag pushed) | — |
| **v0.3.0 release** (analytics-mcp W4-2.1 ~ W4-2.6 — 7 MCP tools + SQLite adapter + 10 skill MCP integration) | ✅ Done | v0.3.0 (2026-05-11) | — |
| **v0.4.0 release** (cli buddy W3-3 agent runtime + background scheduler + cli binary version namespace align) | ✅ Done | v0.4.0 (2026-05-11) | — |
| **v0.4.1 patch** (release workflow: include buddy-mcp_* in publish pattern) | ✅ Done | v0.4.1 | — |
| **v0.5.0 release** (cli buddy W3-4 partial: PROCEDURE output parser — verdict + next-phase metadata) | ✅ Done | v0.5.0 (2026-05-12) | — |
| **v0.6.0 release** (cli buddy W3-4 follow-on: conditional next-phase branches — `- cond → \`skill\`` cascade rule extraction) | ✅ Done | v0.6.0 (2026-05-12) | — |
| **v0.6.1 patch** (release.yml: bump actions/checkout@v6, setup-go@v6, action-gh-release@v3 — Node 24 compatibility ahead of 2026-06-02 default switch) | ✅ Done | v0.6.1 (2026-05-12) | — |
| **v0.6.2 patch** (release pipeline hardening: `make verify-go-version` + `make verify-versions` + push/PR `ci.yml` test gate + `buddy agent log <agent-id>` subcommand + `Store.LatestRun`) | ✅ Done | v0.6.2 (2026-05-12) | — |
| **v0.6.3 patch** (Tier 1.4 exponential backoff: new `backoff_strategy` + `backoff_max` YAML fields, backward-compat. W3-5 retrofit: `cmd/buddy/main.go` 688→147 lines split into 6 sibling files + `agent.go` rename) | ✅ Done | v0.6.3 (2026-05-12) | — |
| **v0.6.4 patch** (Tier 1.5 streaming log capture: `Executor.Run` signature + `LogSink` + per-line `agent_logs` entries; `buddy agent log <id>` mid-progress visibility. Plus verify-quality F1–F4 cleanup: pipe-close hygiene, error-path FD leak fix, `translateExitCode` signature trim, dead-code prune) | ✅ Done | v0.6.4 (2026-05-12) | — |
| **v0.6.5 patch** (Tier 1.6 scheduler live refresh: `RefreshInterval` polling + `tracked` map diff (add/remove/update) + CLI `--refresh` / `--no-refresh`. Closes the last cli-buddy-spec §9 W3-3 follow-on item) | ✅ Done | v0.6.5 (2026-05-12) | — |
| **v0.6.6 patch** (Tier 1.8 webhook output target: `output.type: webhook` + URL/method/headers/timeout. Closes the last W3-6 reference webtoon agent dependency) | ✅ Done | v0.6.6 (2026-05-12) | — |
| **v0.7.0 release** (Bundle: W3-2 minimum-viable TUI + 3 follow-on (detail / scheduler-preview / in-app delete with confirm) + F5 log retention (`buddy agent purge --before <dur> [--apply]`) + W3-6 reference webtoon agent example. New: `agent.PreviewSchedule` helper, `internal/tui` package 51 race-clean tests) | ✅ Done | v0.7.0 (2026-05-17) | — |
| **v0.7.1 patch** (W3-2 follow-on finish: live log tail (t in detail → `Store.LogsSince` ~1s tea.Tick polling) + in-app spec edit (e in detail → $EDITOR shell-out → `Store.UpdateSpec`) + create form (c in list → $EDITOR on `CreateStarterYAML`) — W3-2 6/6 Done. W3-4 chain control: per-step `continue_on_fail` + chain-level `auto_cascade` with `max_depth` (default 5) — runtime queue-based loop, `StepResult.CascadeDepth`. cli-buddy-spec §9 W3-2 + W3-4 모두 Done.) | ✅ Done | v0.7.1 (2026-05-18) | — |
| **v0.7.2 patch** (TUI hook-stats pane, A-3.2 W3-5 follow-on: `H` in list view → ModeHookStats pane via `internal/queries.Run` (count/failures/p50/p95 per hook, default window 1h, same column ordering as `buddy stats` CLI). cli-buddy-spec §9 W3-5 fully closed — main.go split (v0.6.3) + hook monitor 통합 (v0.7.2). 95 race-clean TUI tests.) | ✅ Done — **본 publish 의 baseline** | **v0.7.2** (2026-05-18) | — |
| **ADR-006** (B6 PROCEDURE form bulk-allowlist + --strict CI) | ✅ Done | (post-v0.7.2, 2026-05-19) — plugin v1.0.0 entry B-3 closed | — |
| **ADR-007** (router does not maintain cross-invocation conversational state — supersedes B7) | ✅ Done | (post-v0.7.2, 2026-05-19) — plugin v1.0.0 entry B-4 closed | — |
| **Total** | **148 procedures / 99 commands / 17 ship-release stages + 2 ADRs / 7 MCP analytics tools / 8 cli agent subcommands / `buddy tui` (7 modes) / PROCEDURE parser + conditional branches + drift gates + exponential backoff + streaming logs + scheduler live refresh + webhook output + log retention + scheduler preview + in-app delete + log tail + in-app edit + create form + auto-cascade + continue_on_fail + hook-stats pane + B6 --strict CI gate** | **v0.7.2 + ADR-006/007** (2026-05-19) | 148 |
| **Plugin v1.0.0 entry status** | B-1 W3 cascade ✅ Done / B-2 production dogfood ❌ user-paced / B-3 B6 --strict ✅ ADR-006 / B-4 router smart-skip ✅ ADR-007 supersede | — | — |
| Korea cluster 3 (consult-korea-legal-context / draft-korea-patent-application / audit-korea-cii-vulnerability) | ⏳ deferred (D-F F1) | trigger: target market = Korea | — |
| analytics-mcp | ⏳ deferred (D-C C2) | trigger: §8 일부 구현 후 — *현재 trigger 가능* | — |
| ~~feature-management-mcp~~ | ✅ **Closed by ADR-008 (2026-05-19)** — cli buddy feature registry (5 MCP tools + features table) locked in at v0.3.0 shape; naming split from external `feature-management-saas-mcp` reference design. | — | — |
| Cycle 2 live dispatch | ⏳ deferred | trigger: 별 세션 / 새 cycle | — |

**5단계 비전:** 1) Plugin install ✅ → 2) TUI 상위 레이어(`ai-m` 류) → 3) 설정/세션 관리 툴(`claude-code-organizer` 류) → 4) Dashboard + 칸반 → 5) 4단계에 1~3단계가 모두 녹아듦.

**Dogfood guide (v0.7.x — Plugin v1.0.0 entry condition B-2 실행 절차):** [`docs/dogfood-guide.md`](./dogfood-guide.md) — 3 surface (plugin / cli buddy / hook monitor) 통합 실행 + finding 수집 format + improvement workflow + ADR escalation. v0.1 hook-only legacy 는 [`DOGFOOD.md`](../DOGFOOD.md) (deprecated entry, hook 부분만 유효).

---

---

## 0. 한 줄 정체성

**Buddy** = Claude Code 위에서 *hook 신뢰성·세션 관제·오케스트레이션*을 한 자리에서 다루는 Go CLI. 페르소나는 "친구" — 침묵 default, 차분한 한국어, 이모지 X.

- **Origin:** `github.com/0xmhha/buddy.git`
- **License:** Apache 2.0
- **Stack:** Go 1.25+ (SSoT: `go.mod` go directive — sync `.github/workflows/release.yml` + `README.md` when bumping; `make verify-go-version` checks this), `modernc.org/sqlite` (pure Go), `spf13/cobra`, `stretchr/testify`
- **Binary:** `bin/buddy` (`make build`로 빌드, ~9.5MB static)
- **Module path:** `github.com/wm-it-22-00661/buddy` (이전 머신 잔재 — v0.2 cleanup 후보)

---

## 1. 현재 위치 (one-glance)

> **잔여 작업 + 진행률 + 코드 gap 의 단일 SSoT 는 [`BACKLOG.md`](./BACKLOG.md).** 본 §1 은 *세션 인계 시 한눈 요약* 만.

| 영역 | 상태 |
|------|------|
| Plugin orchestrator — 9-phase × 148 skill | ✅ 100% (Skill Completion Cycle) |
| Plugin commands — 99 | ✅ 100% (N-1 closure ADR-001 + B6 --strict ADR-006) |
| MCP tools — 12 (analytics 7 + feature 5) | ✅ 100% (ADR-008 lock-in) |
| cli buddy W3-1~W3-6 (automation agent 관리) | ✅ 100% (TUI 7 modes + agent runtime + scheduler + cascade chain + 자산 재배치 + reference agent) |
| Hook reliability monitor — M1~M6 + v0.1.0 release | ✅ 100% (4 binaries published 2026-04-26) |
| **cli buddy W4~W8 (AI-usage coaching, ADR-009)** | **80%** (4/5 ship: F2.A v0.8.0 + F2.B v0.9.0 + F2.C v0.10/v0.11 + F2.E v0.12.0. 남은: F2.D Drift) |
| **Plugin v1.0.0 entry conditions B-1~B-4** | 75% (3/4 closed, B-2 production dogfood 잔여) |
| **Whole-product v1.0.0 entry (ADR-010, B-1~B-4 + C-1~C-5 = 9 조건)** | **78%** (7/9 closed — B-1/B-3/B-4 + C-1/C-2/C-3/C-5). 남은: B-2 + C-4 (마지막 C-x). |
| Dogfood **cycle-2 production** (B-2 trigger) | ⚠ ~5% (pre-flight Done, 3 paths user-paced) |
| i18n sweep (M5 deferred) | 25% (W2-3 en/ko parity ✅, 3 잔여) |
| Release polish (M6 deferred) | 25% (ci.yml ✅, SHA pin / notarize / VERSION SSoT 잔여) |
| TUI / runtime UX follow-on | 0% (dogfood signal 대기) |
| **종합 진행률 (whole-product 기준)** | **약 83%** (자동화 agent 100%, AI-usage coaching 80%, 정책 + polish 부분 진척) |

**테스트:** 27 packages race-clean (`go test -race -count=1 ./...`), `make test-skill-form --strict` 148/62 allowlist/86 pass/0 deviate.
**Latest release:** v0.12.0 (2026-05-20) — W7-5 F2.E Notification ship (closes C-5). milestone-driven per ADR-011.

**다음 액션** — [`BACKLOG.md`](./BACKLOG.md) 두 wave 병행:
- **Wave 1 (B-2)**: 사용자 페이스, [`docs/notes/2026-05-19-dogfood-result-cycle-2.md`](./notes/2026-05-19-dogfood-result-cycle-2.md) §B
- **Wave 7 잔여 (AI 단독)**: **W7-4 F2.D Drift Detection (마지막 C-x — closes C-4)** → W7-3c F2.C skill-gen (post-v1.0).

---

## 2. 30초 안에 재개하기

```bash
cd <repo-root>      # 예: /Users/0xtopaz/work/github/0xmhha/buddy

# 1. 최신화 + 빌드 + 테스트
git fetch origin && git status
make build
go test -race -count=2 ./...

# 2. 핵심 문서 3개 읽기
cat README.md          # 한 페이지로 프로젝트 파악
cat docs/HANDOFF.md    # 이 문서 (지금 읽고 있음)
cat docs/roadmap.md    # 다음 작업 SSoT (M5 후 = M6/v0.2/v0.3/v1.0)

# 3. PR 상태 확인
gh pr view 1 --json state,reviews,mergeable
```

---

## 3. 다음 작업 — 시나리오별

사용자의 입력이 어떤 종류냐에 따라 분기.

### 시나리오 A — 사용자가 dogfood 후 feedback 가져옴

**Trigger 발화 예시:** "dogfood 결과 정리했어", "feedback 반영해줘", "며칠 써보니 X가 불편하더라"

v0.1.0 release 가 끝났으므로 dogfood feedback 은 **v0.2 / v0.3 우선순위 + v0.2 dashboard UX (TUI vs web) 결정** 입력.

**다음 행동:**
1. `docs/dogfood-feedback-template.md` 채운 버전(또는 자유 형식) 받기
2. feedback 항목을 분류:
   - **버그/회귀** → 시나리오 D 처리. patch release (v0.1.1) 가 필요하면 별도 brunch.
   - **새 명령/플래그** → v0.2 / v0.3 task 로 매핑
   - **dashboard UX 형식 (TUI vs web)** → roadmap §4 v0.2 open question 에 답
   - **task tracker 통합 여부** → roadmap §5 v0.3 open question 에 답

### 시나리오 B — 사용자가 "v0.2 dashboard", "v0.3 task", "i18n sweep" 류 발화

`docs/roadmap.md` 해당 §(§4/§5/§6) 읽고 plan 만들어 `subagent-driven-development`. dogfood feedback 없이 plan대로 진행할지 한 번 확인.

### 시나리오 C — patch release (v0.1.1) 필요한 버그 발견

**다음 행동:**
1. `git checkout main && git pull origin main`
2. `git checkout -b fix/<issue-name>`
3. `superpowers:systematic-debugging` 으로 fix + regression test
4. PR → main 머지
5. main 에서 `cmd/buddy/main.go`의 `var version = "0.1.1"`, `Makefile`의 `RELEASE_VERSION = 0.1.1`, `CHANGELOG.md` 신규 entry, README의 `VERSION=0.1.1` 동시 업데이트 (5곳 동기화 필수 — workflow의 tag↔Makefile sanity check 가 보호)
6. `git tag v0.1.1 && git push origin v0.1.1` → tag-triggered workflow가 release publish

### 시나리오 D — "buddy doctor가 X를 보여줘", "stats가 Y가 안 돼" 같은 버그 리포트

1. **재현 먼저** (사용자 환경에서 정확한 명령 + output 받기)
2. 영향 패키지 추정:
   - 메시지 wording 이상 → `internal/persona/ko.go`
   - DB 관련 → `internal/db/`, `internal/diagnose/`, `internal/queries/`
   - daemon 동작 → `internal/daemon/`, `internal/aggregator/`
   - install/uninstall → `internal/install/`
3. `superpowers:systematic-debugging` skill 적용
4. fix 후 regression test 추가

### 시나리오 E — "buddy를 어떻게 써?" 같은 사용자 본인용 질문

`DOGFOOD.md` §1~6 안내. 사용자가 본인 머신에 install 시작하려는 시점.

---

## 4. Lock-in 결정 사항 (다시 결정하지 말 것)

다음 세션에서 *왜 이렇게 했지?* 의문이 드는 항목들은 모두 결정 완료된 것. 사용자가 명시적으로 뒤집지 않는 한 유지.

| 결정 | 선택 | 근거 |
|------|------|------|
| 도메인 | Claude Code only | cross-harness parity는 "환상" (harness-engineering-analysis §3 갭 C). 한 harness 깊이 > 여러 harness 얕게. |
| Language | Go (was TS) | Hook wrapper self-overhead. Go 5-10ms vs Node 50-100ms. *친구가 방해꾼이 되면 안 됨.* TS PoC는 `archive/ts-poc/`에 보존. |
| SQLite | `modernc.org/sqlite` | pure Go, cgo 없음, 사내 빌드 환경 안전. |
| Process supervision | hybrid `0xmhha/cli-wrapper` | buddy 단독 동작이 default. cli-wrapper는 옵션. |
| Schema 필드 | 옵션 A (toolName/toolArgs(off)/modelName/tokenUsage/customTags) | spec §6.1 + decision-1-schema-fields.md. v0.1에서 확정. |
| Threshold defaults | 30s/5s/20%/1000/stderr | spec §6.2. M5에서 config로 노출 ✅. |
| 페르소나 | **친구 톤 한국어** | spec §6.3. **침묵 default**, critical도 절제, **이모지 X**. 모든 user-facing 메시지가 이걸 따름. M5에서 `internal/persona/` 카탈로그로 통합 ✅. |
| 통계 윈도우 | 5min/60min/1440min (24h) | aggregator hardcoded. |
| 통계 percentile | SELECT-based MAX (cross-tool) | 정확하지만 비용 있음. >10k events/window면 streaming quantile로 교체. |
| Cross-harness | 의도적 비범위 (v1.0+에서만 검토) | |
| Windows | 의도적 비범위 (v1.0+) | |
| 명령 페르소나 분기 | install/uninstall/doctor/config/purge = 친구 톤. events = 구조적 debug 한 줄 (단 follow start/end는 친구 톤). | events는 grep/awk 대상이라 친구 X. |
| **(M5 추가) Config 우선순위** | 명시적 CLI flag > config 파일 > spec-locked default | doctor, daemon run/start, config CLI 모두 동일. |
| **(M5 추가) Config 변경 적용** | restart 요구 (hot-reload 없음) | 단순함 우선. 사용 패턴 보고 v0.2에서 재검토. |
| **(M5 추가) `purge` 와 outbox** | 절대 건드리지 않음 (구조적 — 패키지에 outbox SQL 0개) | spec §4 invariant 1: outbox는 sync write WAL. |
| **(M5 추가) `uninstall` 기본 동작** | daemon 자동 stop, `--keep-daemon` 으로 opt-out | orphan daemon 방지. |
| **(M5 추가) DB lock race fix** | `db.Open` DSN에 `_pragma=busy_timeout(5000)` | concurrent open(daemon writer + doctor reader) WAL header lock 경합 회피. |
| **(M5 추가) Daemon SIGTERM race fix** | `signal.NotifyContext` 를 PID 파일 publish 전에 설치 | 외부 caller가 PID 파일 보고 SIGTERM 보냈을 때 default handler로 떨어지지 않게. |
| **(M5 추가) 페르소나 i18n 토대** | typed `Key` 상수 + en→ko fallback. en map은 v0.2 sweep까지 비어있음. | 본격 ko/en 분리는 v0.2. |

---

## 5. Open Questions (보존 — 결정 X)

`docs/roadmap.md` 전반에 산재. 결정은 *데이터 들어왔을 때*.

**M5 종료 시점에 정리됨 (모두 v0.2 sweep 후보):**
- ~~M5 config hot-reload vs restart~~ → restart 요구 (1차 결정 적용)
- M5 outbox 수동 cleanup 경로 — *현재 미발생, 데이터 들어오면 결정* (deferred)
- **i18n carry-over (v0.2 i18n sweep targets):**
  - `config.ValidationError.Reason` 카탈로그 wiring (persona keys 이미 declared)
  - `queries.ErrInvalidLimit` / `ErrInvalidWindow` 카탈로그 이전
  - English locale map 채우기
  - 서브커맨드 `--config` 인지 locale 해석 (현재 root PersistentPreRunE는 default path만)

**M6 종료 시점에 정리됨 (모두 v0.2 polish 후보):**
- darwin notarization — v0.1.0 미적용 (사용자가 `xattr -d com.apple.quarantine` 안내)
- 3rd-party Actions SHA pinning — supply-chain 강화
- 별도 `ci.yml` — PR/push 시 test 강제
- `VERSION` 파일 / build-time embed — release cadence 잦아지면

**남아있는 것 (v0.2+):**
1. **v0.2 TUI vs web** — dogfood 결과가 신호
2. **v0.2 멀티-머신 통합** — v1.0+ 검토
3. **v0.3 buddy = task tracker vs 외부 tracker 실행 엔진** — 후자가 scope 작음
4. **v0.3 DAG 시각화 형식** — graphviz / mermaid / TUI
5. **v1.0 plugin 권한 경계** — DB 직접 vs IPC. v1.0 T2가 "MCP transport 재활용" 권장으로 좁혀짐.

---

## 6. 워크플로우 — 어떤 skill을 언제

| 상황 | Skill |
|------|-------|
| 큰 새 기능 (예: M6, v0.2의 한 마일스톤 전체) | `subagent-driven-development` |
| 사용자가 "권장 방향" 류 빠른 진행 원함 | direct (subagent 없이) |
| 사용자가 "분석", "탐색" 요청 | parallel `Explore` agents |
| 버그 리포트 들어옴 | `superpowers:systematic-debugging` |
| 새 design 결정 필요 (예: TUI vs web) | `superpowers:brainstorming` |
| 큰 작업 종료 | `superpowers:finishing-a-development-branch` |

**중요:** 사용자는 "결정 피로"에 민감. 합리적 default 잡고 *한 번만 confirm* 받는 패턴 선호. 8개 옵션 늘어놓는 거 비추.

---

## 7. 메모리 시스템과의 관계

세션 메모리 시스템에 이 프로젝트 entry가 있을 수 있음. 머신/유저 의존이라 경로는 환경마다 다름.

**역할 분리:**
- 메모리 = 세션 시작 시 자동 로드, 짧고 인덱스
- HANDOFF.md = 깊이 있는 운영 문서, 필요 시 직접 읽기

이 문서가 변하면 메모리도 한 줄 업데이트.

---

## 8. 파일 인덱스 — 어떤 문서를 언제 읽나

| 문서 | 언제 읽나 |
|------|----------|
| `README.md` | 프로젝트 한 페이지 소개 (always) |
| `docs/HANDOFF.md` | **현재 문서.** 다른 세션에서 이어 받을 때 가장 먼저 |
| `docs/roadmap.md` | M6 이후 무엇을 할지 결정할 때 |
| `docs/v0.1-spec.md` | M1~M5 구현 의도/invariant 확인 |
| `docs/decision-1-schema-fields.md` | hook event schema 왜 이렇게 결정됐는지 |
| `docs/tasks.md` | cross-track 작업 인벤토리 (Wave 우선순위 + 잔여 29 skill cluster) |
| `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` | plugin 9-phase 아키텍처 현행 SSoT |
| `docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md` | ADR-001 — N-1 closure (commands에 `disable-model-invocation: true` 강제) |
| `docs/superpowers/decisions/2026-05-11-plugin-version-reset.md` | ADR-004 — plugin track version reset v1.x → v0.x + v1.0.0 entry conditions |
| `docs/superpowers/decisions/2026-05-11-cli-buddy-spec-lock-in.md` | ADR-005 — cli buddy spec Draft → Accepted lock-in |
| `docs/notes/2026-05-11-cycle-handoff.md` | **본 cycle close 핸드오프** (v0.2.0 reset → v0.3.0 ship → W3-3 partial). 가장 최신 |
| `docs/notes/2026-05-11-session-handoff.md` | 본 cycle 시작 시점 baseline (Cycle 1 close + v1.1.1 patch) |
| `DOGFOOD.md` | 사용자가 본인 머신에 install할 때 안내 |
| `docs/dogfood-feedback-template.md` | 며칠 사용 후 회고 템플릿 |
| `archive/ts-poc/` | TS PoC 자산 (참조용, *사용 X*) |

---

## 9. 자주 쓰는 명령

```bash
# 빌드 + 테스트
make build
go test -race -count=2 ./...    # 15 packages, ~150+ tests
go vet ./...
gofmt -l .

# 로컬에서 buddy 사용해보기 (테스트용 격리 환경)
SANDBOX=$(mktemp -d)
mkdir -p $SANDBOX/claude
echo '{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo pre"}]}]}}' > $SANDBOX/claude/settings.json
./bin/buddy install \
  --claude-dir $SANDBOX/claude \
  --buddy-dir $SANDBOX \
  --buddy-binary "$PWD/bin/buddy"
./bin/buddy daemon start --db $SANDBOX/buddy.db
./bin/buddy doctor --db $SANDBOX/buddy.db --pid $SANDBOX/daemon.pid
./bin/buddy config show --config $SANDBOX/config.json
./bin/buddy purge --db $SANDBOX/buddy.db --before 30d        # dry-run
./bin/buddy uninstall --claude-dir $SANDBOX/claude --buddy-dir $SANDBOX --buddy-binary "$PWD/bin/buddy"
rm -rf $SANDBOX
```

---

## 10. 사용자 컨텍스트

(다음 세션 Claude가 사용자 의도를 더 잘 읽도록.)

- **언어:** 한국어 우선. 영어 발화하면 영어로 응답.
- **결정 피로 회피:** 8 옵션 펼치지 말 것. 합리적 default + 한 번 confirm.
- **Git commit attribution:** `Co-Authored-By` / "Generated with Claude" 류 *절대* 추가 X. 이번 PR 본문 작성 때도 이 룰 따랐음.
- **Author/Committer:** `mhha <mhha@wemade.com>` 로 통일 (M5 작업 후 history rewrite).
- **종료 시 commit 확인:** uncommitted 남으면 사용자에게 commit 여부 *먼저* 물어봐. 자율 commit/폐기 X.
- **Push:** 사용자가 직접 push 패턴 선호. AI는 commit까지만. (단, PR 생성/업데이트 등 사용자가 명시적으로 지시하면 push 가능.)
- **main brunch force push:** 시스템 룰상 절대 금지. main에 영향이 있는 history rewrite는 별도 brunch에서 rebase 후 PR로.
- **Output format (사용자 글로벌 룰):** 일반 응답 끝에 Fact-based Answer 섹션 (`Fact:` `Your Opinion:` 라벨 분리, 확신도 High/Mid/Low/None 부여). *문서 작성 산출물에는 적용 X*.
- **Skill 호출:** 사용자가 `/skill-name` 명시적으로 부르면 그 skill 따름. 안 부르면 내 판단.

---

## 11. 즉시 처리할 수 있는 작은 것

다음 세션이 빠른 워밍업으로 처리 가능한 항목 (우선순위 낮음, *원할 때*):

- ~~Daemon test flake~~ ✅ M5 prereq commit 에서 fix됨 — `db.Open` busy_timeout + signal handler order.
- ~~`docs/roadmap.md` 8 open question 인덱스 추가~~ ✅ 이 HANDOFF.md §5 에 통합.
- ~~Versioned binary~~ ✅ M6 T3.
- ~~Cross-compile + release workflow~~ ✅ M6 T1+T2.
- ~~Plugin command context bloat (N-1)~~ ✅ ADR-001 (`disable-model-invocation: true` × 57 commands, commit `ca99762`). **신규 command 추가 시 이 컨벤션 강제** — `plugin/commands/*.md` 모든 신규 파일은 frontmatter에 `disable-model-invocation: true`를 포함해야 한다.
- ~~**`cmd/buddy/main.go` 분할 (685 lines)**~~ ✅ Done (2026-05-12, post-v0.6.2). 688 → 147 lines. 8 sub-feature 가 sibling `<feature>_cmd.go` 로 분리됨 (events / stats / doctor / install / daemon / hookwrap) + `agent.go` → `agent_cmd.go` rename. [Unreleased] 에 entry 보존.
- ~~**모듈 path:** `github.com/wm-it-22-00661/buddy` — 이전 머신 잔재.~~ ✅ commit `4ce3ccb` 으로 `github.com/0xmhha/buddy` 일괄 변경 완료 (43 파일 / 91 import / go.mod 1 줄).
- **gofmt drift 한 번 정리:** `gofmt -l .`이 가끔 비어있지 않으면 한 commit으로 정리 (현재는 clean). ([`tasks.md`](./tasks.md) C-3)

---

## 12. 마지막 commit으로 무엇이 들어갔나 (sanity check)

```bash
git log --oneline -6 main
```

예상 (v0.1.0 release 직후 기준):
```
03a4fa9 Add release prep: cross-compile matrix, tag-triggered workflow, CHANGELOG (#2)
514f667 docs: sync HANDOFF, roadmap, v0.1-spec with M5 outcome
bd97352 Add config, retention purge, message catalog, and reliability fixes (#1)
d3b5b0b docs: HANDOFF.md — session resume guide
a461de3 docs: roadmap polish (typo + M5 decision-deadlines + plugin IPC + i18n forward)
b721998 docs: add post-v0.1 roadmap (M5/M6/v0.2/v0.3/v1.0)
```

만약 위와 다르면 누군가 직접 작업했다는 뜻. `git log --since='1 month ago' main` 로 확인. v0.1.0 tag 위로 hotfix가 쌓이면 SHA는 또 달라짐.

---

이 문서가 더 이상 정확하지 않으면, **이 문서부터 업데이트하자.** 다른 세션이 의지하는 SSoT.
