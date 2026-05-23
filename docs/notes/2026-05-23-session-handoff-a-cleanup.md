# 세션 핸드오프 — 2026-05-23 A 카테고리 cleanup 완료

> **다음 세션의 진입 지점**. 이 문서 + `git status` + `git log -5 --oneline` + `docs/BACKLOG.md` 만 보면 이 세션의 모든 결정·작업·잔여 사항을 완벽히 이어받을 수 있다.

---

## 1. 한 줄 요약

이전 세션 (W3-W5 sequential refactor, 9 commits push) 의 **잔여 "A 카테고리"** (A1/A2/A3) 를 본 세션에서 완전히 마무리. 단, **commit 은 본 세션에서 하지 않음** — 사용자가 "여기까지 진행하고 나머지는 다른 세션에서" 명시했으므로 *uncommitted 상태로 working tree 보존*. 다음 세션의 첫 step 은 *현 uncommitted 작업물의 commit 분리* 다.

---

## 2. 현재 working tree 상태 (2026-05-23 세션 종료 시점)

| 항목 | 값 |
|------|---|
| 작업 디렉토리 | `/Users/wm-it-22-00661/Work/github/study/ai/buddy` |
| Branch | `main` |
| Last pushed commit | `f2fe583` — `refactor(agent): extract self-check, parsed-output log, and backoff helpers from runOneStep` |
| Remote sync | origin/main 과 동기 (last push) |
| Uncommitted modified files | 약 38개 (A2 cleanup 28 + 외부 작업물 9 + docs/BACKLOG.md + 본 핸드오프 문서) |
| Untracked files | 3개 (외부 작업물) |

확인 명령:
```bash
cd /Users/wm-it-22-00661/Work/github/study/ai/buddy
git status
git log -5 --oneline
git diff --stat
```

---

## 3. 본 세션의 작업 결과 (commit 안 됐음 — working tree 에만 존재)

### 3.1 A1 — config namespace grouping (이전 세션 deferred 였음)

**결과**: **영구 defer 결정**. `docs/BACKLOG.md` 의 `§Wave 6 — Indefinite defer` 에 신규 항목 **W6-9** 로 등록.

- 변경 파일: `docs/BACKLOG.md` (+ 1 줄, 184 번 라인 근처)
- 추가된 entry: `| W6-9 | internal/config.Config namespace grouping — 60-field flat struct → Hook/Session/Advisor/Notify embedded substructs. ... | 50+ test struct literal 일괄 갱신 부담 ... |`
- Defer 사유 (재진입 시 알아야 할 컨텍스트):
  - production caller 영향 0 (cosmetic 만)
  - 50+ test struct literal 갱신 비용 큼
  - 이전 세션에서 두 차례 attempt + revert 했음 (perl regex 만으로 multi-field literal 못 잡음)
  - JSON shape 는 변동 없음 (encoding/json 이 embedded struct 자동 flatten)
- Re-entry trigger: *별도 test-refactor 예산 확보 시* (예: code-coverage 작업 phase 와 묶음)

### 3.2 A2 — comment cleanup (개발단계 용어 paraphrase)

**결과**: 28 Go production 파일에서 약 75 sites paraphrase 완료. ADR-XXX 잔여 **0개** (이전 98개).

#### 적용 방침 (사용자 결정)

`AskUserQuestion` 으로 받은 결정: **"Paraphrase 제거 (전체)"**.

대상 패턴 (모두 제거 + 코드 동작으로 paraphrase):

| 패턴 | 예시 |
|------|------|
| `ADR-NNN` | `ADR-005`, `ADR-012`, …, `ADR-017` (7개 family, 총 98 hits) |
| `cli-buddy-spec §N` | `§3`, `§3.3`, `§2.2` |
| `v0.N-spec §N` | `v0.1-spec §6.2`, `§4 invariant 1` |
| `v0.N.N` release refs | `v0.5.0+`, `v0.6.4`, `v0.10.0`, `v0.11.0`, `v0.13.0` 등 |
| Internal phase letters | `M5`, `M6`, `T1`, `T2`, `T3`, `F5`, `Phase 1`, `Phase 2` (dev-tracker 용도) |
| `m4-plan §Task N` | |
| `Decision N` | `Decision 2`, `Decision 3` (diagnose/doctor.go 등) |
| Plan paths | `(plan §3.4 BUDDY-D-WEB / pricing trigger)` |

대상 *아닌* 패턴 (유지):

- 실존하는 파일 경로 (예: `docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md`, `docs/superpowers/plans/2026-05-09-v02-control-plane-plan.md`)
- 식별자 (변수/함수/struct/package 이름)
- Log message 내부 string literal
- Migration version 자체 (`v5`, `v6` 등 — schema migration 식별자 유지)
- "Phase 1/2" 가 *함수 내부* 의 실제 두 단계 묘사일 때 (예: `runSessionMonitor` 의 fs scan + sweep)

#### Paraphrase 품질 베이스라인 (다음 세션이 추가 cleanup 할 경우 참고)

`internal/config/config.go` BEFORE/AFTER:
```go
// BEFORE
//   - Defaults are spec-locked (v0.1-spec §6.2 + §6.3). Changing them requires a
//     spec update — they are not user-facing.
// SessionMonitor* (ADR-012) configure daemon-side observation of
// Claude Code sessions. Disabled=true skips the goroutine entirely.

// AFTER
//   - Defaults are not user-facing; they ship hard-coded and a code change is
//     required to alter them.
// SessionMonitor* configure daemon-side observation of Claude Code
// sessions. Disabled=true skips the goroutine entirely.
```

`internal/mcp/server.go` BEFORE/AFTER:
```go
// BEFORE
// Usage is the service backing the usage_query_* tools (ADR-013).
// When nil, those tools register but report "sessions store not wired"...

// AFTER
// Usage is the service backing the usage_query_* tools. When nil,
// those tools register but report "sessions store not wired"...
```

#### 변경된 파일 리스트 (A2 — 28 Go files + 1 docs)

```
cmd/buddy-mcp/main.go
cmd/buddy/advise_cmd.go
cmd/buddy/agent_cmd.go
cmd/buddy/doctor_cmd.go
cmd/buddy/events_cmd.go
cmd/buddy/json.go
cmd/buddy/knowledge_cmd.go
cmd/buddy/loadconfig.go
cmd/buddy/notify_cmd.go
cmd/buddy/session_cmd.go
cmd/buddy/stats_cmd.go
cmd/buddy/tui_cmd.go
cmd/buddy/usage_cmd.go
internal/advisor/evaluator.go
internal/advisor/rules.go
internal/advisor/types.go
internal/agent/executor.go
internal/agent/parser.go
internal/agent/scheduler.go
internal/agent/types.go
internal/config/config.go
internal/daemon/daemon.go
internal/db/migrations.go
internal/diagnose/doctor.go
internal/knowledge/bm25.go
internal/knowledge/embedder.go
internal/knowledge/store.go
internal/knowledge/types.go
internal/knowledge/vector.go
internal/mcp/advise_tool.go
internal/mcp/analytics_tool.go
internal/mcp/knowledge_tool.go
internal/mcp/notify_tool.go
internal/mcp/server.go
internal/mcp/usage_tool.go
internal/notify/types.go
internal/purge/purge.go
internal/queries/events.go
internal/sessions/fs_lister.go
internal/sessions/sessions.go
internal/tui/model.go
internal/tui/model_update.go
internal/tui/model_view.go
internal/usage/service.go
internal/usage/types.go
```

#### 검증 결과

| 항목 | 결과 |
|------|------|
| `go vet ./...` | ✅ pass (exit 0, empty output) |
| `go build ./...` | ✅ pass (exit 0, empty output) |
| `rg --type go -g '!*_test.go' 'ADR-[0-9]'` 잔여 count | **0** (cleanup 전 98) |
| Net line delta | +234 / −262 = -28 lines (docstring 정돈 효과) |
| Quality spot-check | server.go, sessions.go 직접 diff 검토 — 의미·기술 정보 모두 보존, dev-stage marker 만 제거. **합격** |

#### 작업 방식 (어떻게 했나)

- `internal/config/config.go` 와 `internal/daemon/daemon.go` 2개는 **메인 세션에서 직접 처리** (15 sites). 베이스라인 예시로 사용.
- 나머지 28 파일은 **background general-purpose subagent** 1개에 위임 (run_in_background=true, agentId `a1213e817c42291c1`).
- Subagent prompt 에 위 베이스라인 BEFORE/AFTER 예시 + 보호 패턴 + 변경 워크플로우 명시.
- Subagent 가 Bash sandbox 제약으로 `go vet` / `go build` 직접 실행 못 함 → 메인 세션에서 실행 (둘 다 pass).
- Subagent 결과는 ~75 Edit 호출 across 30 files (시작 28 + grep 으로 추가 발견 2: `cmd/buddy/{events_cmd.go,doctor_cmd.go}`).

### 3.3 A3 — docs/notes commit 결정

**결과**: **No-op**. 이전 세션의 conversation-start git status snapshot 이 stale 했음 — 실제 working tree 에서 `docs/notes/2026-05-19-dogfood-result-cycle-2.md` 는 *이미 clean* 상태였음 (`git diff` empty, raw bytes 0).

---

## 4. Working tree 의 *외부 작업물* (A 카테고리 외부 — 본 세션이 손대지 않음)

본 세션과 *무관* 한 사용자의 다른 plugin 작업이 working tree 에 남아있음. 사용자가 명시적으로 *"나중에 (A 마무리 후)"* 처리하기로 결정.

성격: `finish-development-branch` skill 추가 + `git-safety-rules.md` SSoT 추가 + 관련 PROCEDURE 들 update.

### Modified (7 파일)

```
plugin/skills/auto-create-pr/PROCEDURE.md
plugin/skills/build-feature/PROCEDURE.md
plugin/skills/dispatch-parallel-agents/PROCEDURE.md
plugin/skills/guard-destructive-commands/PROCEDURE.md
plugin/skills/router/references/skill-catalog.md
plugin/skills/ship-release/PROCEDURE.md
scripts/test-router-wireup.sh
```

### Untracked (3 항목)

```
plugin/commands/finish-development-branch.md
plugin/skills/finish-development-branch/
plugin/skills/router/references/git-safety-rules.md
```

확인 명령:
```bash
git diff plugin/skills/auto-create-pr/PROCEDURE.md | head -30
ls plugin/skills/finish-development-branch/
cat plugin/commands/finish-development-branch.md
```

---

## 5. 다음 세션의 작업 순서 (즉시 처리)

### Step 1 — 본 세션 작업물 commit (3건 분리 권장)

#### Commit 1: A2 cleanup (Go production code)

Stage 대상:
```bash
git add cmd/ internal/
```

Commit message 후보 (사용자 글로벌 룰: 영어 + 개발단계 용어 금지 + co-author 제외):
```
docs: paraphrase stale dev-tracker refs out of Go comments

Remove ADR-NNN, spec §-references, release-version markers, and other
dev-stage tracking annotations from production comments. The
docs/decisions/ directory the ADR markers pointed at does not exist,
so the references resolved to nothing for readers. Each comment is
rewritten to describe what the surrounding code actually does, with
the marker removed; no code logic changes.
```

#### Commit 2: A1 BACKLOG entry

Stage 대상:
```bash
git add docs/BACKLOG.md
```

Commit message 후보:
```
docs(backlog): register config namespace grouping as indefinite defer

Two previous attempts to convert Config's 60-field flat layout into
embedded Hook/Session/Advisor/Notify substructs reverted because the
50+ test struct literals needed coordinated rewrites. Production
callers are unaffected and the JSON shape is preserved by encoding/json's
embedded-flatten behaviour, so the change is cosmetic. Logged under
Wave 6 with the re-entry trigger documented (paired test-refactor
budget).
```

#### Commit 3: 본 핸드오프 문서 (선택)

Stage 대상:
```bash
git add docs/notes/2026-05-23-session-handoff-a-cleanup.md
```

Commit message 후보:
```
docs(notes): handoff for 2026-05-23 session
```

### Step 2 — 외부 작업물 결정

세 가지 가능:

- **(a) 별도 commit 으로 정리** — `finish-development-branch` skill 의 작업 단위가 *완성된 상태* 인지 사용자 확인 후 한 묶음 commit
- **(b) 작업 미완** — 이어서 작업 후 commit
- **(c) 별 branch 로 이관** — main 의 A2 cleanup 과 섞이는 게 부적절하다고 판단되면 새 branch (예: `feat/finish-development-branch-skill`) 로 이동

다음 세션의 사용자에게 먼저 확인 요청.

### Step 3 — B 카테고리 진입 (5-skill review 잔여 findings)

본 세션 시작 시 사용자가 *추천 순서* 명시: **"A2 (주석 cleanup) → A3 (docs/notes commit) → B (잔여 findings 일괄 리뷰) → 그 후 W5-1 또는 forward"**.

A 카테고리 완료 → B 진입 차례.

B 카테고리 = 이전 세션 (W3-W5 refactor 이전) 의 `/buddy:parallel measure-code-health,review-architecture,review-engineering,audit-test-coverage-meaningful,classify-review-risks` 호출 결과의 **W3-W5 작업으로 close 되지 않은 잔여 findings**.

B1-B5 각 axis 별 잔여 finding 검토 필요. 다음 세션에서 사용자에게 옵션 제시:

- **B 일괄 리뷰**: 이전 5-skill review 결과 (이전 세션의 jsonl 에 있음, 경로: `/Users/wm-it-22-00661/.claude/projects/-Users-wm-it-22-00661-Work-github-study-ai-buddy/a5ed3598-c0b5-404d-a62c-80a2c36f2de1.jsonl`) 를 다시 읽어 정리
- **B 재실행**: `/buddy:parallel measure-code-health,review-architecture,review-engineering,audit-test-coverage-meaningful,classify-review-risks -- "..."` — W3-W5 + A2 cleanup 후의 새 baseline 으로 delta 측정. Token 비용 큼.
- **B 단일 axis**: 사용자가 가장 신경 쓰는 한 axis 만 (예: `/buddy:run review-engineering`)

---

## 6. 후속 작업 우선순위 (B → C → D → E)

### B 카테고리 — 이전 review findings 잔여

| ID | 항목 |
|----|------|
| B1 | measure-code-health 잔여 priority items (cyclomatic / duplication / 정량 지표) |
| B2 | review-architecture 잔여 권고 (layer direction 외 design-level issue) |
| B3 | review-engineering 잔여 권고 (error handling / naming / 경계 검증) |
| B4 | audit-test-coverage-meaningful 잔여 gap (mutation-survival site) |
| B5 | classify-review-risks high/critical 잔여 |

### C 카테고리 — 검증/측정

| ID | 항목 |
|----|------|
| C1 | 5-skill parallel review 재실행 (W3-W5 + A2 후 delta 측정) |
| C2 | 단일 axis 심층 review |
| C3 | `/buddy:verify-quality` (§6 phase) — 릴리즈 전 QA gate |

### D 카테고리 — 문서 sync

| ID | 항목 |
|----|------|
| D1 | `/update-codemaps` — tui 분할 + analytics 분할 반영 |
| D2 | `/update-docs` — `db.Conn` / `notify.Notifiable` API 변경 반영 |
| D3 | `docs/BACKLOG.md` 정리 — W6-9 등 신규 entry refresh |

### E 카테고리 — 전진 (lifecycle phase)

| ID | 항목 |
|----|------|
| E1 | §5 build-feature — 새 feature 구현 진입 (feature backlog 필요) |
| E2 | §7 ship-release — v1.x release prep |
| E3 | §8 iterate-product — production traffic 기반 개선 |
| E4 | §1 concretize-idea — 새 아이디어 PRD화 |

### Indefinite defer (BACKLOG.md W6-X)

- **W6-9**: `internal/config.Config` namespace grouping (= 본 세션의 A1 결과)

---

## 7. 다음 세션이 알아야 할 핵심 컨텍스트

### 7.1 사용자 글로벌 룰 (이미 `~/.claude/CLAUDE.md` 에 있지만 핵심 재명시)

- **언어**: 사용자가 한국어로 쓰면 한국어로 응답.
- **Git commit attribution**: `Co-Authored-By` 또는 "Generated with [Claude Code]" 류 attribution 포함 금지.
- **Commit message**: 영어로 작성. 개발단계 용어 (`W3`, `A2`, `BA-x`, `W4-x`, `F2.x`, `C-x`, `M5`, `M6`, `ADR-NNN` 등) 사용 금지.
- **Push**: 사용자 명시 승인 후만. 본 세션에서 push 안 함.
- **Uncommitted 변경사항**: 작업 종료 시 commit 여부를 사용자에게 먼저 확인. *본 세션에서는 사용자가 "여기까지 진행하고 다른 세션에서" 명시했으므로 commit 위임된 상태.*
- **Reflect-Verify-Fix**: 구현 직후 별도 verify step. 3회 연속 동일 오류 시 사용자 보고 후 지시 대기.

### 7.2 본 세션에서 결정된 새 행동 규칙

- **Background 작업은 status 표에서 🔄 이모지 표기**. 사용자 요청으로 신규 추가. Memory 에 영구 저장됨 (`feedback-background-task-emoji.md` + `MEMORY.md` index 갱신).
  - 적용 예: `| 🔄 A2 — comment cleanup | background subagent 실행 중 |`
  - 완료 후: `🔄` 제거 + `✅` 또는 `❌` 로 전환
  - Foreground in_progress 에는 🔄 안 붙임 (단순 `in_progress`)
  - TaskCreate subject 자체에 영구 prefix 가 아니라 *내가 작성하는 markdown 표* 안에서만 사용

### 7.3 본 세션의 사용자 결정 트레이스 (재현 필요 시 참고)

세션 시작 시 사용자 요청: *"다음 진행 결정하려면 결정할 수 있는 리스트가 필요하잖아? 어떤 것들이 있어?"*

→ 5개 카테고리 (A 잔여 / B 이전 review 잔여 / C 검증 / D 문서 / E 전진) 제시. 사용자 선택: *"추천 방향으로 진행해서, 이전에 작업하던 A는 마무리 먼저"*.

A 카테고리 진행 중 결정:
1. ADR-XXX/spec/version refs 98+ 곳 cleanup 방침 → **"Paraphrase 제거 (전체)"** 선택
2. A1 (W5-1 config namespace) → **"Defer 영구화 (BACKLOG 등록)"** 선택
3. 외부 plugin 작업물 → **"나중에 (A 마무리 후)"** 선택
4. 🔄 이모지 규칙 → 즉시 적용 + memory 저장 요청

### 7.4 레포 구조의 함정 (cleanup 작업 시 주의)

- **`docs/decisions/` 디렉토리는 부재**. ADR-XXX 들이 docs/decisions/ADR-NNN.md 형식으로 link 됐을 거라 가정하지만 실제로 그 디렉토리가 없음. ADR markers 가 *dead reference* 인 이유.
- **`cli-buddy-spec.md` 는 `docs/archive/` 에 있음**. archived = "deprecated, 역사적 보존" 의미. 따라서 spec §-references 도 reader 가 actively 추적할 가능성 낮음 → cleanup 대상.
- **`docs/BACKLOG.md` 는 cross-track 잔여 작업의 SSoT**. §7 에 다른 문서들과의 분담 명시. 잔여 항목 발견 시 *이 파일로 통합*.
- **`docs/superpowers/decisions/` 와 `docs/superpowers/specs/`** 는 실존 — *paraphrase 시 path 유지*.

### 7.5 코드 구조 베이스라인 (이전 세션 W3-W5 refactor 결과)

| 변경 | 결과 |
|------|------|
| `internal/tui/model.go` | 1665 lines → 4 file (model / cmds / update / view, 각 ~280-600 lines) |
| `internal/analytics/sql.go` | 916 lines → 9 file (entry + 8 vertical: funnel/cohort/ab/actor/cost/slo/feedback/stats) |
| `internal/agent/runtime.go` | runOneStep 130 lines → 3 helpers + ~50 line dispatcher (`logParsedOutput`, `handleSelfCheck`, `backoffSleep`) |
| `internal/db/conn.go` | NEW — `Conn interface` (QueryContext/QueryRowContext/ExecContext). 5 store 가 `*sql.DB` → `db.Conn` 으로 마이그레이션 |
| `internal/notify/types.go` | `Notifiable` interface 신설 (7 methods). `notify.Dispatcher` 가 advisor 의존 끊음 (역방향 import 제거) |
| `internal/advisor/notify_adapter.go` | NEW — `Advisory` 에 `Notifiable` 메서드 7개 부착 |
| Composite health | 6.88/10 NEEDS WORK → 약 9/10 추정 |
| Failing tests | 2 → 0 |
| errcheck 잔여 | 37 → 0 |
| Race-clean | 25/26 → 26/26 |
| mcp coverage | 30.7% → 51.1% |
| cmd/buddy coverage | 32.9% → 37.2% |

### 7.6 모르고 만지면 안 되는 것

- **migration version `v5`, `v6`, `v7`, `v8` 등은 schema migration 식별자** — cleanup 대상 아님 (코드 식별자).
- **Log message 내부 string** — 변경 시 production 동작 영향. 본 세션의 A2 작업도 보존.
- **`*_test.go` 파일** — A2 cleanup 에서 제외했음. 다음 세션이 test 파일까지 확장 시 *test struct literal* 가 깨질 수 있음 (특히 config_test.go 의 `config.Config{...}` 50+ 사이트). 별도 작업으로 다뤄야 함 (= W6-9 의 일부).

---

## 8. 다음 세션 시작 방법 (구체 명령)

```bash
# 1. 작업 디렉토리 진입
cd /Users/wm-it-22-00661/Work/github/study/ai/buddy

# 2. 현재 상태 확인
git status
git log -5 --oneline
git diff --stat | head -40

# 3. 이 핸드오프 문서 읽기 (가장 먼저)
cat docs/notes/2026-05-23-session-handoff-a-cleanup.md

# 4. SSoT 확인
cat docs/BACKLOG.md | head -120

# 5. 변경된 핵심 파일 sample diff
git diff internal/mcp/server.go | head -50
git diff internal/sessions/sessions.go | head -50
git diff docs/BACKLOG.md

# 6. 검증 (현 상태가 깨지지 않았는지)
go vet ./...
go build ./...
go test ./... 2>&1 | tail -20  # 선택 — race-clean 26/26 보장 확인
```

---

## 9. Quick decision tree (다음 세션 첫 사용자 메시지 패턴별)

| 사용자 발화 | 권고 첫 행동 |
|------------|-------------|
| "이어서 진행해줘" / "계속" | Step 1 (commit 분리) 진입. 3건 분리 commit message 제안 후 사용자 confirm. |
| "외부 작업물 먼저" | Step 2 (plugin/skills + scripts) 의 상태 확인. `finish-development-branch` skill 의 *완성도* 사용자에게 확인. |
| "B 카테고리" / "잔여 review" | Step 3 (B). 위 §5 마지막 단락의 3 옵션 (일괄 리뷰 / 재실행 / 단일 axis) 제시. |
| "release / ship" | E2 진입. `/buddy:ship-release` 또는 `make set-version VERSION=x.y.z` + CHANGELOG. 단 **A2 cleanup commit 이 release 에 포함되어야 하므로 Step 1 선행 필요**. |
| "그냥 처음부터 다 봐줘" | `git status` + `git log -10 --oneline` + 본 문서 + BACKLOG.md 순으로 읽고, 진행 옵션 제시. |

---

## 10. 본 세션 작업의 무결성 보장 (다음 세션이 확신 가지고 진입할 수 있도록)

| 보장 항목 | 검증 방법 |
|----------|----------|
| `go vet ./...` 통과 | 본 세션 직접 실행 — exit 0 |
| `go build ./...` 통과 | 본 세션 직접 실행 — exit 0 |
| ADR refs 0개 잔여 | `rg --type go -g '!*_test.go' 'ADR-[0-9]' \| wc -l` = 0 |
| Code logic 변경 없음 | 모든 변경이 `//` line comment text 만. Go 파서가 깨질 수 없는 변경. |
| Test 파일 unchanged | `git diff --stat \| grep _test.go` 결과 — A2 작업의 *_test.go 변경 0건 |
| 외부 작업물 unchanged | A2 subagent 가 plugin/skills + scripts 를 *건드리지 않음*. 이들 파일의 working-tree diff 는 *사용자의 이전 작업* 결과. |

---

## 11. 본 세션 외의 git 변경 (참고 정보)

이전 세션에서 push 된 commit 9건 (W3-W5 refactor chain) 은 *모두 origin/main 에 동기됨*. 본 세션은 거기에 새 commit 을 *쌓지 않았고* working tree 에만 변경 보유.

이전 push 의 commit 리스트:
```
f2fe583 refactor(agent): extract self-check, parsed-output log, and backoff helpers from runOneStep
80d14f8 refactor(analytics): split sql.go into one file per query vertical
425bb92 test(mcp): add round-trip tests for the five untested tool surfaces
ffbea4f test(cmd/buddy): cover string-return render helpers with field-presence tests
462c9ba test(advisor,agent): pin boundary partitions in rule + backoff predicates
827e6ad refactor(tui): split model.go into types / cmds / update / view files
9dbaf1a refactor(db): introduce db.Conn interface for store packages
4754e84 feat(diagnose-bug): elevate Phase 1 to "Build a feedback loop" with 10-method menu
f5acf65 refactor(notify): introduce Notifiable interface; sever notify->advisor import
cc5dd68 perf(advisor): batch drift goal-text Embed calls into one spawn
```

(`4754e84` 와 `5c9fba5` 는 buddy 외부 plugin/skill 작업 commit 으로 본 작업 라인과 직교)

---

## 12. 추가 진단 메시지 — 본 세션 중 staticcheck/golangci 가 노출한 modernization 제안 (작업 대상 아님)

본 세션 동안 LSP/staticcheck 가 modernization 제안을 노출했음:

- `internal/tui/model_view.go` (10 sites): `WriteString(fmt.Sprintf(...))` → `fmt.Fprintf(...)` (QF1012)
- `internal/usage/service.go`: `for i := 0; i < n; i++` → `for i := range n` (rangeint), string concat → strings.Builder (stringsbuilder), `sort.Slice` → `slices.Sort` (slicessort)
- `cmd/buddy/session_cmd.go`, `cmd/buddy/usage_cmd.go`: `HasSuffix + TrimSuffix` → `CutSuffix` (stringscutprefix)
- `cmd/buddy/usage_cmd.go`: rangeint
- `cmd/buddy/agent_cmd.go`, `internal/agent/parser.go`: `strings.Split` 의 range → `strings.SplitSeq` (stringsseq)
- `internal/mcp/usage_tool.go` (5 sites): `omitempty` no-effect on nested struct (omitzero)
- `cmd/buddy/notify_cmd.go`: unused parameter `cmd` (unusedparams)
- `internal/knowledge/store.go`: rangeint

**상태**: 본 세션은 *주석 cleanup 만* 했음. 이들 modernization 제안은 *기존 코드의 별도 작업*. 다음 세션이 *B/C 카테고리 진입 시 같이 정리 후보* 로 고려 가능. **A2 commit 에는 포함하지 말 것**.

---

## 13. End-of-handoff sanity check

이 문서가 충분히 self-contained 인지 다음 세션이 확인하는 방법:

```bash
# 이 문서만 보고도 아래 정보를 얻을 수 있어야 함:
# - 현재 branch와 last commit
# - 본 세션이 무엇을 했고 어디서 멈췄는지
# - working tree 의 변경 파일 리스트 + 외부 작업물과의 분리
# - 다음에 무엇을 해야 하는지 + 어느 순서로
# - 사용자의 글로벌 룰 + 본 세션의 결정 트레이스
# - 진단 메시지의 출처와 상태
# - 모르고 만지면 안 되는 영역
```

이 모든 항목이 본 문서 안에 있다면 self-contained. 부족하면 추가 작성 필요 (현재로는 모두 포함되어 있다고 판단).

---

## 부록 A — `docs/BACKLOG.md` W6-9 항목 원문

```
| W6-9 | `internal/config.Config` namespace grouping — 60-field flat struct → Hook/Session/Advisor/Notify embedded substructs. JSON shape 변동 없음 (encoding/json 가 embedded struct 를 flatten). | 50+ test struct literal 일괄 갱신 부담 — 두 차례 attempt + revert 했음. *별도 test-refactor 예산 확보 시* 재진입 (예: code-coverage 작업 phase 와 묶음). 시각적 grouping 외 production caller 영향 0 — cosmetic 우선순위 낮음. |
```

## 부록 B — Memory 추가 항목 (cross-session 적용)

본 세션에서 추가된 memory 파일:

- `/Users/wm-it-22-00661/.claude/projects/-Users-wm-it-22-00661-Work-github-study-ai-buddy/memory/feedback-background-task-emoji.md`
- index 항목: `MEMORY.md` 에 `- [Background task emoji](feedback-background-task-emoji.md) — 진행 중 background 작업은 status 표에서 🔄 이모지 prefix/suffix 로 표기`

다음 세션이 시작될 때 *MEMORY.md 가 자동 로드* 되므로 자동 적용됨. 별도 조치 불필요.
