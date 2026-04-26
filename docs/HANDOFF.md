# Buddy — Session Handoff

> 다른 세션에서 이 프로젝트를 이어 받는 사람(또는 미래의 자기 자신)이 *처음 5분 안에* 어디까지 와있는지 파악하고, *다음 한 시간 안에* 일을 재개할 수 있도록 만든 문서.

**Last updated:** 2026-04-26 (M5 완료, PR #1 open)

---

## 0. 한 줄 정체성

**Buddy** = Claude Code 위에서 *hook 신뢰성·세션 관제·오케스트레이션*을 한 자리에서 다루는 Go CLI. 페르소나는 "친구" — 침묵 default, 차분한 한국어, 이모지 X.

- **Origin:** `github.com/0xmhha/buddy.git`
- **License:** Apache 2.0
- **Stack:** Go 1.22+, `modernc.org/sqlite` (pure Go), `spf13/cobra`, `stretchr/testify`
- **Binary:** `bin/buddy` (`make build`로 빌드, ~9.5MB static)
- **Module path:** `github.com/wm-it-22-00661/buddy` (이전 머신 잔재 — v0.2 cleanup 후보)

---

## 1. 현재 위치 (one-glance)

| 항목 | 상태 |
|------|------|
| M1 schema/SQLite/outbox | ✅ |
| M2 hook-wrap CLI | ✅ |
| M3 daemon + aggregator + cliwrapcfg | ✅ |
| M4 install/uninstall/doctor/stats/events | ✅ |
| **M5** config CLI + 4 friction fix + purge + 페르소나 catalog | ✅ (PR #1) |
| DOGFOOD.md + feedback template | ✅ |
| 사용자의 실제 dogfood 사용 | ⏳ 대기 (M5 friction fix 반영판으로 재시작 권장) |
| Dogfood feedback 회수 | ⏳ 대기 |
| M6 (release prep) | 📋 계획 완료, 시작 안 함 |
| v0.2 / v0.3 / v1.0 | 📋 outline only |

**테스트:** 15개 패키지, ~150+ tests pass (`-race -count=2` clean).
**Sync 상태:** `m5` 브랜치 `origin/m5`에 push 완료. PR #1 (`m5 → main`) review/merge 대기.
**Active PR:** https://github.com/0xmhha/buddy/pull/1

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

### 시나리오 A — PR #1 머지 후 후속 작업

**Trigger 발화 예시:** "PR 머지했어", "M5 머지됐다", "다음 마일스톤"

**다음 행동:**
1. `git fetch origin && git checkout main && git pull` 으로 로컬 main 동기화
2. M6 진행 의향이면 `docs/roadmap.md` §3 (M6 — Release prep) 읽고 plan 만들기
3. 새 feature brunch (`m6` 등) 생성 후 `subagent-driven-development` skill로 task별 진행
4. 시나리오 D(dogfood feedback) 가 들어오면 v0.2/v0.3 우선순위 입력으로 활용

### 시나리오 B — PR #1 review 피드백 반영

**Trigger 발화 예시:** "리뷰 코멘트 받았어", "PR 수정해줘"

**다음 행동:**
1. `gh pr view 1 --comments` 로 코멘트 fetch
2. `m5` 브랜치 체크아웃 후 코멘트별 fix
3. `git push origin m5` (force-with-lease 필요한 경우만)
4. PR 자동 갱신

### 시나리오 C — 사용자가 dogfood 후 feedback 가져옴

**Trigger 발화 예시:** "dogfood 결과 정리했어", "feedback 반영해줘", "며칠 써보니 X가 불편하더라"

**M5 이전엔 M5 우선순위 입력**이었지만, 이제 **v0.2/v0.3 우선순위 입력**으로 활용.

**다음 행동:**
1. `docs/dogfood-feedback-template.md` 채운 버전(또는 자유 형식) 받기
2. feedback 항목을 분류:
   - **버그/회귀** → 시나리오 E 처리
   - **새 명령/플래그** → roadmap §M6 Open question 또는 v0.2/v0.3 task로 매핑
   - **dashboard UX 형식 (TUI vs web)** → roadmap §4 v0.2 open question에 답
   - **task tracker 통합 여부** → roadmap §5 v0.3 open question에 답

### 시나리오 D — 사용자가 "M6 release", "v0.2 dashboard", "v0.3 task" 류 발화

`docs/roadmap.md` 해당 §읽고 plan 만들어 `subagent-driven-development`. M6는 의존성상 v0.2/v0.3보다 먼저(roadmap §8 그래프).

### 시나리오 E — "buddy doctor가 X를 보여줘", "stats가 Y가 안 돼" 같은 버그 리포트

1. **재현 먼저** (사용자 환경에서 정확한 명령 + output 받기)
2. 영향 패키지 추정:
   - 메시지 wording 이상 → `internal/persona/ko.go`
   - DB 관련 → `internal/db/`, `internal/diagnose/`, `internal/queries/`
   - daemon 동작 → `internal/daemon/`, `internal/aggregator/`
   - install/uninstall → `internal/install/`
3. `superpowers:systematic-debugging` skill 적용
4. fix 후 regression test 추가

### 시나리오 F — "buddy를 어떻게 써?" 같은 사용자 본인용 질문

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

**M5 종료 시점에 정리됨:**
- ~~M5 config hot-reload vs restart~~ → restart 요구 (1차 결정 적용)
- M5 outbox 수동 cleanup 경로 — *현재 미발생, 데이터 들어오면 결정* (deferred)
- **M5 i18n carry-over (v0.2 i18n sweep targets):**
  - `config.ValidationError.Reason` 카탈로그 wiring (persona keys 이미 declared)
  - `queries.ErrInvalidLimit` / `ErrInvalidWindow` 카탈로그 이전
  - English locale map 채우기
  - 서브커맨드 `--config` 인지 locale 해석 (현재 root PersistentPreRunE는 default path만)

**남아있는 것 (M6 이후):**
1. **M6 darwin notarization** — v0.1.0 미적용 결정 (사용자가 `xattr -d com.apple.quarantine` 안내)
2. **v0.2 TUI vs web** — dogfood 결과가 신호
3. **v0.2 멀티-머신 통합** — v1.0+ 검토
4. **v0.3 buddy = task tracker vs 외부 tracker 실행 엔진** — 후자가 scope 작음
5. **v0.3 DAG 시각화 형식** — graphviz / mermaid / TUI
6. **v1.0 plugin 권한 경계** — DB 직접 vs IPC. v1.0 T2가 "MCP transport 재활용" 권장으로 좁혀짐.

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
| `docs/m4-plan.md` | M4 구현 시 사용한 plan (이미 완료, 역사) |
| `docs/dogfood-and-roadmap-plan.md` | dogfood + post-v0.1 roadmap 작성 plan (역사) |
| `docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md` | v1.0 plugin 설계 사전 작업 |
| `docs/superpowers/plans/2026-04-24-buddy-plugin-scaffold-plan.md` | v1.0 plugin 구현 plan |
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

- ~~Daemon test flake~~ ✅ M5 prereq commit (`a82272f`)에서 fix됨 — `db.Open` busy_timeout + signal handler order.
- ~~`docs/roadmap.md` 8 open question 인덱스 추가~~ ✅ 이 HANDOFF.md §5 에 통합.
- **`cmd/buddy/main.go` 분할 (650+ lines):** M5 후 651 lines. M6에서 명령이 더 늘기 전에 sibling files (`config_cmd.go`, `loadconfig.go`, `purge_cmd.go`, `spawn.go` 는 이미 분리됨)에 추가로 install/daemon/doctor/stats/events/hookwrap 도 옮기면 좋음. **권고: M6 시작 전.**
- **모듈 path:** `github.com/wm-it-22-00661/buddy` — 이전 머신 잔재. 현재 origin인 `github.com/0xmhha/buddy` 와 일치시키려면 모든 import 경로 일괄 변경 필요. v0.2 cleanup 후보.
- **gofmt drift 한 번 정리:** `gofmt -l .`이 가끔 비어있지 않으면 한 commit으로 정리 (현재는 clean).

---

## 12. 마지막 commit으로 무엇이 들어갔나 (sanity check)

```bash
git log --oneline -6 m5
```

예상 (M5 완료 직후 기준 — rebased 이후 SHA):
```
a625694 fix(persona): address T5 review (1 Important + 3 Minor)
914defa feat(persona): user-facing message catalog (M5 T5)
ac4fa79 feat(purge): buddy purge --before <date> --apply (M5 T4)
d3c9e54 feat(config): wire config into doctor/daemon (M5 T3)
21b5366 style(config): drop unneeded loop-var copy (Go 1.22+)
d64e75a fix(config): address T2 review (2 Important + 3 Minor)
```

만약 위와 다르면 누군가 직접 작업했다는 뜻. `git log --since='1 month ago' m5` 로 확인. PR #1 이 머지되어 main에 squash/merge 형태로 들어가면 SHA는 또 달라짐.

---

이 문서가 더 이상 정확하지 않으면, **이 문서부터 업데이트하자.** 다른 세션이 의지하는 SSoT.
