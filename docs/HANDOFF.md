# Buddy — Session Handoff

> 다른 세션에서 이 프로젝트를 이어 받는 사람(또는 미래의 자기 자신)이 *처음 5분 안에* 어디까지 와있는지 파악하고, *다음 한 시간 안에* 일을 재개할 수 있도록 만든 문서.

**Baseline**: v0.13.0 (2026-05-20). **남은 작업**: v1.0.0 entry condition B-2 (production dogfood, user-paced) + cli buddy W4 follow-on 일부 + trigger-bound Wave 5 / W3-4.

## 트랙 상태

> 트랙 정체성 / 책임 경계의 SSoT는 [`docs/two-tracks-charter.md`](./two-tracks-charter.md).

| 트랙 | 잔여 작업 |
|------|----------|
| **plugin buddy** — Claude Code plugin (skill / MCP / agent / hook 카탈로그). 9-phase orchestrator, single-router dispatch | **B-2 production dogfood** (cycle-3 active) + Korea cluster (Wave 5, trigger-bound) |
| **cli buddy** — TUI 자동화 agent 관리 + AI-usage coaching (ADR-009/010) | **W4 follow-on**: W4-1 branch-aware skill selection / W4-2 self-check fail semantics / W4-4 scheduler live indicator / W4-5 log-tail scrollback. W3-4 macOS notarization (Apple Dev ID trigger-bound). |

**Dogfood guide (B-2 실행 절차)**: [`docs/dogfood-guide.md`](./dogfood-guide.md) — 3 surface (plugin / cli buddy / hook monitor) 통합 실행. **현 active cycle**: [`docs/notes/2026-05-20-dogfood-result-cycle-3.md`](./notes/2026-05-20-dogfood-result-cycle-3.md).

---

## 0. 한 줄 정체성

**Buddy** = Claude Code 위에서 *hook 신뢰성·세션 관제·오케스트레이션·AI-usage coaching* 을 한 자리에서 다루는 Go CLI. 페르소나는 "친구" — 침묵 default, 차분한 한국어, 이모지 X.

- **Origin:** `github.com/0xmhha/buddy.git`
- **License:** Apache 2.0
- **Stack:** Go 1.25+ (SSoT: `go.mod` go directive — sync `.github/workflows/release.yml` + `README.md` when bumping; `make verify-go-version` checks this), `modernc.org/sqlite` (pure Go), `spf13/cobra`, `stretchr/testify`
- **Binary:** `bin/buddy` (`make build`로 빌드, ~9.5MB static)
- **Module path:** `github.com/0xmhha/buddy`

---

## 1. 잔여 작업 (one-glance)

> **잔여 작업의 단일 SSoT 는 [`BACKLOG.md`](./BACKLOG.md).** 본 §1 은 *세션 인계 시 한눈 요약* 만.

| 영역 | 상태 |
|------|------|
| **Plugin v1.0.0 entry (ADR-010, 9 조건)** | **8/9 closed.** 잔여: B-2 production dogfood (user-paced — cycle-3 active) |
| cli buddy W4 follow-on (TUI / runtime UX) | dogfood signal 대기 — W4-1 / W4-2 / W4-4 / W4-5 |
| Wave 5 Korea cluster (consult-korea-legal-context / draft-korea-patent-application / audit-korea-cii-vulnerability) | trigger-bound (target market = Korea) |
| macOS notarization (W3-4) | trigger-bound (Apple Dev ID 발급 필요) |
| Post-v1.0: W7-3c F2.C skill autogen | ADR-014 Phase 3, deferred |

**Latest release:** v0.13.0 (2026-05-20). milestone-driven per ADR-011.
**테스트:** `go test -race -count=1 ./...` 전체 race-clean.

**다음 액션** — v1.0.0 ship gating 은 단 한 조건:
- **B-2 production dogfood**: 사용자 페이스 — [`docs/dogfood-guide.md`](./dogfood-guide.md). 완료 시 v1.0.0 release 가능.

---

## 2. 30초 안에 재개하기

```bash
cd <repo-root>

# 1. 최신화 + 빌드 + 테스트
git fetch origin && git status
make build
go test -race -count=2 ./...

# 2. 핵심 문서 3개 읽기
cat README.md             # 한 페이지로 프로젝트 파악
cat docs/HANDOFF.md       # 이 문서
cat docs/BACKLOG.md       # 잔여 작업 SSoT
```

---

## 3. 다음 작업 — 시나리오별

### A — dogfood 후 feedback 가져옴

**Trigger 발화 예시:** "dogfood 결과 정리했어", "feedback 반영해줘", "며칠 써보니 X가 불편하더라"

1. `docs/dogfood-feedback-template.md` 채운 버전(또는 자유 형식) 받기
2. feedback 항목을 분류:
   - **버그/회귀** → 시나리오 C 처리. patch release 가 필요하면 별도 branch.
   - **새 명령/플래그/UX** → `BACKLOG.md` W4 또는 신규 Wave 후보로 매핑
   - **ADR 필요한 결정** → `docs/superpowers/decisions/<date>-<topic>.md` 작성

### B — `BACKLOG.md` 우선순위 작업

`BACKLOG.md` 의 in-progress / next 항목 진행. 큰 작업은 `superpowers:subagent-driven-development`, 작은 작업은 직접.

### C — 버그 리포트

1. **재현 먼저** (사용자 환경에서 정확한 명령 + output 받기)
2. 영향 패키지 추정:
   - 메시지 wording → `internal/persona/`
   - DB 관련 → `internal/db/`, `internal/diagnose/`, `internal/queries/`
   - daemon 동작 → `internal/daemon/`, `internal/aggregator/`
   - agent runtime → `internal/agent/`
   - session / usage / knowledge / advise / notify → `internal/{sessions,usage,knowledge,advisor,notify}/`
   - TUI → `internal/tui/`
3. `superpowers:systematic-debugging` skill 적용
4. fix 후 regression test 추가

### D — 사용자 본인용 install / usage 질문

`DOGFOOD.md` 와 `docs/dogfood-guide.md` 안내.

---

## 4. Lock-in 결정 사항 (다시 결정하지 말 것)

다음 세션에서 *왜 이렇게 했지?* 의문이 드는 항목들은 모두 결정 완료. 사용자가 명시적으로 뒤집지 않는 한 유지.

| 결정 | 선택 | 근거 |
|------|------|------|
| 도메인 | Claude Code only | cross-harness parity는 "환상". 한 harness 깊이 > 여러 harness 얕게. |
| Language | Go (was TS) | Hook wrapper self-overhead. Go 5-10ms vs Node 50-100ms. |
| SQLite | `modernc.org/sqlite` | pure Go, cgo 없음. |
| Process supervision | hybrid `0xmhha/cli-wrapper` | buddy 단독 동작이 default. cli-wrapper는 옵션. |
| Schema 필드 | Option A (toolName / toolArgs(off) / modelName / tokenUsage / customTags) | [`docs/archive/decision-1-schema-fields.md`](./archive/decision-1-schema-fields.md). |
| Threshold defaults | 30s / 5s / 20% / 1000 / stderr | v0.1 spec §6.2. config 노출됨. |
| 페르소나 | **친구 톤 한국어** | **침묵 default**, critical도 절제, **이모지 X**. 모든 user-facing 메시지가 이걸 따름. `internal/persona/` 카탈로그 통합. |
| 통계 윈도우 | 5min / 60min / 1440min (24h) | aggregator hardcoded. |
| 통계 percentile | SELECT-based MAX (cross-tool) | 정확하지만 비용 있음. >10k events/window면 streaming quantile로 교체. |
| Cross-harness | 의도적 비범위 (v1.0+ 검토만) | |
| Windows | 의도적 비범위 (v1.0+) | |
| 명령 페르소나 분기 | install / uninstall / doctor / config / purge = 친구 톤. events = 구조적 debug 한 줄 (follow start/end는 친구 톤). | events는 grep / awk 대상. |
| Config 우선순위 | 명시적 CLI flag > config 파일 > spec-locked default | doctor, daemon run/start, config CLI 모두 동일. |
| Config 변경 적용 | restart 요구 (hot-reload 없음) | 단순함 우선. 사용 패턴 보고 재검토. |
| `purge` 와 outbox | 절대 건드리지 않음 (구조적 — 패키지에 outbox SQL 0개) | spec §4 invariant 1: outbox는 sync write WAL. |
| `uninstall` 기본 동작 | daemon 자동 stop, `--keep-daemon` 으로 opt-out | orphan daemon 방지. |
| DB lock race fix | `db.Open` DSN에 `_pragma=busy_timeout(5000)` | concurrent open WAL header lock 경합 회피. |
| Daemon SIGTERM race fix | `signal.NotifyContext` 를 PID 파일 publish 전에 설치 | 외부 caller SIGTERM 보호. |
| 페르소나 i18n | typed `Key` 상수 + en↔ko parity (57/57) | sweep 완료. |
| Release cadence | milestone-driven (Wave 완료 단위로 minor bump) | ADR-011. |
| Plugin / cli buddy 책임 분리 | plugin = skill / MCP / agent / hook 카탈로그. cli = 자동화 agent 관리 + AI-usage coaching. | [`docs/two-tracks-charter.md`](./two-tracks-charter.md). |
| router cross-invocation state | router 는 maintain 안 함 (각 invocation 독립) | ADR-007. |
| PROCEDURE form allowlist | bulk allowlist + `--strict` CI | ADR-006. |

---

## 5. Open Questions (보존 — 결정 X)

데이터 / dogfood 신호 들어왔을 때 결정.

- **TUI vs web dashboard** — dogfood 결과가 신호 (v0.2+ 검토)
- **멀티-머신 통합** — v1.0+ 검토
- **buddy = task tracker vs 외부 tracker 실행 엔진** — 후자가 scope 작음 (v0.3+)
- **DAG 시각화 형식** — graphviz / mermaid / TUI (v0.3+)
- **plugin 권한 경계** — DB 직접 vs IPC. "MCP transport 재활용" 권장으로 좁혀짐 (v1.0+)
- **macOS notarization** — Apple Dev ID 발급 trigger
- **Korea cluster (Wave 5)** — target market = Korea trigger
- **outbox 수동 cleanup 경로** — 데이터 들어오면 결정

---

## 6. 워크플로우 — 어떤 skill을 언제

| 상황 | Skill |
|------|-------|
| 큰 새 기능 (한 Wave 전체 등) | `subagent-driven-development` |
| 사용자가 "권장 방향" 빠른 진행 원함 | direct (subagent 없이) |
| 사용자가 "분석", "탐색" 요청 | parallel `Explore` agents |
| 버그 리포트 | `superpowers:systematic-debugging` |
| 새 design 결정 필요 | `superpowers:brainstorming` |
| 큰 작업 종료 | `superpowers:finishing-a-development-branch` |

**중요:** 사용자는 "결정 피로"에 민감. 합리적 default 잡고 *한 번만 confirm* 받는 패턴 선호. 8개 옵션 늘어놓는 거 비추.

---

## 7. 메모리 시스템과의 관계

세션 메모리 시스템에 이 프로젝트 entry가 있을 수 있음. 머신 / 유저 의존이라 경로는 환경마다 다름.

**역할 분리:**
- 메모리 = 세션 시작 시 자동 로드, 짧고 인덱스
- HANDOFF.md = 깊이 있는 운영 문서, 필요 시 직접 읽기

이 문서가 변하면 메모리도 한 줄 업데이트.

---

## 8. 파일 인덱스 — 어떤 문서를 언제 읽나

| 문서 | 언제 읽나 |
|------|----------|
| `README.md` | 프로젝트 한 페이지 소개 |
| `docs/HANDOFF.md` | **현재 문서.** 다른 세션 인계 |
| `docs/BACKLOG.md` | **잔여 작업 + 진행률 SSoT** |
| `docs/two-tracks-charter.md` | plugin vs cli buddy 책임 경계 SSoT |
| `docs/dogfood-guide.md` | B-2 dogfood 실행 절차 |
| `docs/dogfood-feedback-template.md` | 사용 후 회고 템플릿 |
| `docs/response-format-guide.md` | 응답 포맷 스타일 가이드 |
| `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` | plugin 9-phase 아키텍처 현행 SSoT |
| `docs/superpowers/decisions/` | ADR Index — 결정의 "왜" |
| `docs/notes/2026-05-20-dogfood-result-cycle-3.md` | **현 active dogfood cycle** |
| `docs/notes/` (older) | 이전 cycle handoffs (historical) |
| `docs/archive/` | 완료된 spec / 결정 문서 (역사적 보존) |
| `DOGFOOD.md` | 사용자가 본인 머신에 install할 때 안내 |

---

## 9. 자주 쓰는 명령

```bash
# 빌드 + 테스트
make build
go test -race -count=2 ./...
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

- **언어:** 한국어 우선. 영어 발화하면 영어로 응답.
- **결정 피로 회피:** 8 옵션 펼치지 말 것. 합리적 default + 한 번 confirm.
- **Git commit attribution:** `Co-Authored-By` / "Generated with Claude" 류 *절대* 추가 X.
- **Author / Committer:** `mhha <mhha@wemade.com>` 로 통일.
- **종료 시 commit 확인:** uncommitted 남으면 사용자에게 commit 여부 *먼저* 물어봐. 자율 commit / 폐기 X.
- **Push:** 사용자가 직접 push 패턴 선호. AI는 commit까지만. (단, PR 생성 / 업데이트 등 사용자가 명시적으로 지시하면 push 가능.)
- **main branch force push:** 시스템 룰상 절대 금지. main에 영향이 있는 history rewrite는 별도 branch에서 rebase 후 PR로.
- **Output format:** 일반 응답 끝에 Fact-based Answer 섹션 (`Fact:` `Your Opinion:` 라벨 분리, 확신도 High / Mid / Low / None 부여). 문서 작성 산출물에는 적용 X.
- **Skill 호출:** 사용자가 `/skill-name` 명시적으로 부르면 그 skill 따름. 안 부르면 내 판단.

---

## 11. 즉시 처리할 수 있는 작은 것

- **gofmt drift 한 번 정리:** `gofmt -l .`이 가끔 비어있지 않으면 한 commit으로 정리 (현재는 clean).

---

## 12. 마지막 commit으로 무엇이 들어갔나 (sanity check)

```bash
git log --oneline -10 main
```

위로 hotfix / 새 작업이 쌓이면 SHA는 달라짐.

---

이 문서가 더 이상 정확하지 않으면, **이 문서부터 업데이트하자.** 다른 세션이 의지하는 SSoT.
