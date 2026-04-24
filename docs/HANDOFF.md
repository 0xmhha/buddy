# Buddy — Session Handoff

> 다른 세션에서 이 프로젝트를 이어 받는 사람(또는 미래의 자기 자신)이 *처음 5분 안에* 어디까지 와있는지 파악하고, *다음 한 시간 안에* 일을 재개할 수 있도록 만든 문서.

**Last updated:** 2026-04-24 (M4 + dogfood 준비 + roadmap 마무리 직후)

---

## 0. 한 줄 정체성

**Buddy** = Claude Code 위에서 *hook 신뢰성·세션 관제·오케스트레이션*을 한 자리에서 다루는 Go CLI. 페르소나는 "친구" — 침묵 default, 차분한 한국어, 이모지 X.

- **위치:** `/Users/wm-it-22-00661/Work/github/study/ai/buddy/`
- **Origin:** `github.com/0xmhha/buddy.git`
- **License:** Apache 2.0
- **Stack:** Go 1.22+, `modernc.org/sqlite` (pure Go), `spf13/cobra`, `stretchr/testify`
- **Binary:** `bin/buddy` (`make build`로 빌드, ~9.5MB static)

---

## 1. 현재 위치 (one-glance)

| 항목 | 상태 |
|------|------|
| M1 schema/SQLite/outbox | ✅ |
| M2 hook-wrap CLI | ✅ |
| M3 daemon + aggregator + cliwrapcfg | ✅ |
| M4 install/uninstall/doctor/stats/events | ✅ |
| **DOGFOOD.md + feedback template** | ✅ |
| **docs/roadmap.md (M5+)** | ✅ |
| 사용자의 실제 dogfood 사용 | ⏳ 대기 (사용자 본인 행동) |
| Dogfood feedback 회수 | ⏳ 대기 |
| M5 (config + 4 friction fix) | 📋 계획 완료, 시작 안 함 |
| M6 (release prep) | 📋 계획 완료 |
| v0.2 / v0.3 / v1.0 | 📋 outline only |

**테스트:** 11개 패키지, 100+ tests pass (race-clean).
**Sync 상태:** origin/main과 0/1 — polish commit 1개가 로컬에 남아있을 수 있음 (사용자가 직접 push하는 패턴).

---

## 2. 30초 안에 재개하기

```bash
cd /Users/wm-it-22-00661/Work/github/study/ai/buddy

# 1. 최신화 + 빌드 + 테스트
git fetch origin && git status
make build
make test

# 2. 핵심 문서 3개 읽기
cat README.md          # 한 페이지로 프로젝트 파악
cat docs/HANDOFF.md    # 이 문서 (지금 읽고 있음)
cat docs/roadmap.md    # 다음 작업 SSoT

# 3. 메모리도 가능하면 일별
cat ~/.claude/projects/-Users-wm-it-22-00661-Work-github-study-ai/memory/MEMORY.md
cat ~/.claude/projects/-Users-wm-it-22-00661-Work-github-study-ai/memory/project_harness_engineering_analysis.md
```

---

## 3. 다음 작업 — 시나리오별

사용자의 입력이 어떤 종류냐에 따라 분기.

### 시나리오 A — 사용자가 dogfood 후 feedback 가져옴

**Trigger 발화 예시:** "dogfood 결과 정리했어", "feedback 반영해줘", "며칠 써보니 X가 불편하더라"

**다음 행동:**
1. `docs/dogfood-feedback-template.md`를 사용자가 채운 버전을 받는다 (또는 사용자가 자유 형식으로 말함)
2. feedback 항목을 `docs/roadmap.md` §M5 task 우선순위에 매핑
3. 가장 큰 마찰 1~2개를 *M5 첫 task*로 끌어올림
4. `subagent-driven-development` skill로 M5 첫 task 시작

**M5에서 *반드시* 다뤄야 하는 4개 friction (DOGFOOD dry-run에서 발견):**
1. `buddy install` 후 `daemon start` 안 하고 `doctor` 실행 시 cryptic SQL error → roadmap §M5 T6
2. `buddy daemon start` 직후 `pid -1` 출력 (cmd.Process.Release race) → roadmap §M5 T7
3. `--db` parent dir 없으면 `out of memory (14)` cryptic error → roadmap §M5 T8
4. `uninstall`이 daemon 안 끄면 orphan → roadmap §M5 T9

각각 fix 후보 (a)/(b)와 "결정 시점: M5 implementation kickoff. 기본 후보 (a)" 명시되어 있음.

### 시나리오 B — 사용자가 "M5 진행" 류 발화

**Trigger 발화:** "M5 진행", "config 명령 만들자", "다음 마일스톤"

**다음 행동:**
1. `docs/roadmap.md` §M5 전체 읽기 (T1~T9)
2. **dogfood feedback 없으면 사용자에게 한 번 확인:** "dogfood 결과 먼저 정리할래, 아니면 plan 그대로 진행할까?"
3. `subagent-driven-development` skill로 M5 task 1개씩
4. M5 끝나면 `finishing-a-development-branch`

### 시나리오 C — 사용자가 "M6 release", "v0.2 dashboard", "v0.3 task" 류 발화

`docs/roadmap.md` 해당 §읽고 plan 만들어 `subagent-driven-development`. M5 안 끝났으면 *왜 M5 먼저인지* 한 번 짚음 (의존성 그래프 — roadmap §8).

### 시나리오 D — "buddy를 어떻게 써?" 같은 사용자 본인용 질문

`DOGFOOD.md` §1~6 안내. 사용자가 본인 머신에 install 시작하려는 시점.

### 시나리오 E — "buddy doctor가 X를 보여줘", "stats가 Y가 안 돼" 같은 버그 리포트

1. **재현 먼저** (사용자 환경에서 정확한 명령 + output 받기)
2. `internal/diagnose/` 또는 `internal/queries/` 또는 `cmd/buddy/main.go` 해당 부분 읽기
3. `superpowers:systematic-debugging` skill 적용
4. fix 후 regression test 추가

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
| Threshold defaults | 30s/5s/20%/1000/stderr | spec §6.2. M5에서 config로 노출. |
| 페르소나 | **친구 톤 한국어** | spec §6.3. **침묵 default**, critical도 절제, **이모지 X**. 모든 user-facing 메시지가 이걸 따름. |
| 통계 윈도우 | 5min/60min/1440min (24h) | aggregator hardcoded. |
| 통계 percentile | SELECT-based MAX (cross-tool) | 정확하지만 비용 있음. >10k events/window면 streaming quantile로 교체. |
| Cross-harness | 의도적 비범위 (v1.0+에서만 검토) | |
| Windows | 의도적 비범위 (v1.0+) | |
| 명령 페르소나 분기 | install/uninstall/doctor = 친구 톤. events = 구조적 debug 한 줄 (단 follow start/end는 친구 톤). | events는 grep/awk 대상이라 친구 X. |

---

## 5. Open Questions (보존 — 결정 X)

`docs/roadmap.md` 전반에 산재한 8개. 결정은 *데이터 들어왔을 때*.

1. **M5 config hot-reload vs restart 요구** — 1차안 restart, 재검토 여지
2. **M5 outbox cleanup 수동 경로 필요 여부** — purge가 events/stats만 다룸
3. **M6 darwin notarization** — v0.1.0 미적용 결정
4. **v0.2 TUI vs web** — dogfood 결과가 신호 (터미널-only vs 브라우저 분리 사용자)
5. **v0.2 멀티-머신 통합** — v1.0+ 검토
6. **v0.3 buddy = task tracker vs 외부 tracker 실행 엔진** — 후자가 scope 작음
7. **v0.3 DAG 시각화 형식** — graphviz / mermaid / TUI
8. **v1.0 plugin 권한 경계** — DB 직접 vs IPC. v1.0 T2가 "MCP transport 재활용" 권장으로 좁혀짐.

이 중 몇 개는 dogfood feedback이 답을 줄 가능성 (#1, #4 특히).

---

## 6. 워크플로우 — 어떤 skill을 언제

| 상황 | Skill |
|------|-------|
| 큰 새 기능 (예: M5, M6, v0.2의 한 마일스톤 전체) | `subagent-driven-development` |
| 사용자가 "권장 방향" 류 빠른 진행 원함 | direct (subagent 없이) |
| 사용자가 "분석", "탐색" 요청 | parallel `Explore` agents |
| 버그 리포트 들어옴 | `superpowers:systematic-debugging` |
| 새 design 결정 필요 (예: TUI vs web) | `superpowers:brainstorming` |
| 큰 작업 종료 | `superpowers:finishing-a-development-branch` |

**중요:** 사용자는 "결정 피로"에 민감. 합리적 default 잡고 *한 번만 confirm* 받는 패턴 선호 (지금까지 내내). 8개 옵션 늘어놓는 거 비추.

---

## 7. 메모리 시스템과의 관계

`~/.claude/projects/-Users-wm-it-22-00661-Work-github-study-ai/memory/` 에 이 프로젝트 entry 있음:
- `MEMORY.md` — 한 줄 인덱스 (Buddy Project 항목)
- `project_harness_engineering_analysis.md` — 분석부터 M4까지의 진척 요약 (이 문서가 더 상세)

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
| `docs/roadmap.md` | M5 이후 무엇을 할지 결정할 때 |
| `docs/v0.1-spec.md` | M1~M4 구현 의도/invariant 확인 |
| `docs/decision-1-schema-fields.md` | hook event schema 왜 이렇게 결정됐는지 |
| `docs/m4-plan.md` | M4 구현 시 사용한 plan (이미 완료) |
| `docs/dogfood-and-roadmap-plan.md` | 본 문서 직전 작업의 plan (이미 완료) |
| `DOGFOOD.md` | 사용자가 본인 머신에 install할 때 안내 |
| `docs/dogfood-feedback-template.md` | 며칠 사용 후 회고 템플릿 |
| `../harness-engineering-analysis.md` | 33개 harness 도구 분석 (Buddy 설계 근거) |
| `archive/ts-poc/` | TS PoC 자산 (참조용, *사용 X*) |

---

## 9. 자주 쓰는 명령

```bash
# 빌드 + 테스트
make build
make test          # race detector 포함
go vet ./...
gofmt -l .

# 로컬에서 buddy 사용해보기 (테스트용 격리 환경)
mkdir -p /tmp/buddy-sandbox/{claude,data}
echo '{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo pre"}]}]}}' > /tmp/buddy-sandbox/claude/settings.json
./bin/buddy install \
  --claude-dir /tmp/buddy-sandbox/claude \
  --buddy-dir /tmp/buddy-sandbox/data \
  --buddy-binary "$PWD/bin/buddy"
./bin/buddy daemon start --db /tmp/buddy-sandbox/data/buddy.db
./bin/buddy doctor --db /tmp/buddy-sandbox/data/buddy.db --pid /tmp/buddy-sandbox/data/daemon.pid
./bin/buddy uninstall --claude-dir /tmp/buddy-sandbox/claude --buddy-binary "$PWD/bin/buddy"
rm -rf /tmp/buddy-sandbox
```

---

## 10. 사용자 컨텍스트

(다음 세션 Claude가 사용자 의도를 더 잘 읽도록.)

- **언어:** 한국어 우선. 영어 발화하면 영어로 응답.
- **결정 피로 회피:** 8 옵션 펼치지 말 것. 합리적 default + 한 번 confirm.
- **Git commit attribution:** `Co-Authored-By` / "Generated with Claude" 류 *절대* 추가 X.
- **종료 시 commit 확인:** uncommitted 남으면 사용자에게 commit 여부 *먼저* 물어봐. 자율 commit/폐기 X.
- **Push:** 사용자가 직접 push 패턴 선호. AI는 commit까지만.
- **Output format (사용자 글로벌 룰):** 답변 끝에 Fact-based Answer 섹션 (`Fact:` `Your Opinion:` 라벨 분리, 확신도 High/Mid/Low/None 부여). 단, *이 룰은 일반 응답용이지 모든 산출물에 적용은 아님* — 문서 작성 시는 적용 X.
- **Skill 호출:** 사용자가 `/skill-name` 명시적으로 부르면 그 skill 따름. 안 부르면 내 판단.

---

## 11. 즉시 처리할 수 있는 작은 것

다음 세션이 빠른 워밍업으로 처리 가능한 항목 (우선순위 낮음, *원할 때*):

- **gofmt drift 한 번 정리:** `gofmt -l .`이 가끔 비어있지 않으면 한 commit으로 정리
- **Daemon test flake:** `internal/daemon/TestRun_DrainsOutboxThenStopsOnContextCancel`이 race detector 부하 시 가끔 timing flake. CI 도입 전 한 번 안정화 필요.
- **`cmd/buddy/main.go` 분할:** 600+ lines. M5 명령 추가 전에 `cmd/buddy/cmd/{install,daemon,doctor,stats,events,hookwrap}.go`로 쪼개면 좋음 (final review에서 권고).
- **`docs/roadmap.md` 8 open question 인덱스 추가** (final review M5 minor): 한 곳에 모아 보면 편함.

---

## 12. 마지막 commit으로 무엇이 들어갔나 (sanity check)

```bash
git log --oneline -10
```

예상 (이 문서 commit 시점 기준):
```
<sha> docs: HANDOFF.md — session resume guide
a461de3 docs: roadmap polish (typo + M5 decision-deadlines + plugin IPC + i18n forward)
b721998 docs: add post-v0.1 roadmap (M5/M6/v0.2/v0.3/v1.0)
e206ea0 docs: fix DOGFOOD review issues (2 Critical + 4 Important)
925579d docs: add DOGFOOD guide + feedback template
ebf5ecd docs: add dogfood-and-roadmap plan
3607fa7 M4 docs/polish: README quickstart, spec §7 update, gofmt drift, --db help
68ab12a M4-T4 fix: address review (3 Important + Minor lint/boundary/test)
7378590 M4-T4: buddy events
90799e7 M4-T3 fix: extract internal/format, fix case-sensitivity + by-tool downgrade
```

만약 위와 다르면 누군가 직접 작업했다는 뜻. `git log --since='1 month ago'`로 확인.

---

이 문서가 더 이상 정확하지 않으면, **이 문서부터 업데이트하자.** 다른 세션이 의지하는 SSoT.
