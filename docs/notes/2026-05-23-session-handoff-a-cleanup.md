# 세션 핸드오프 — 2026-05-23 A 카테고리 cleanup 완료

> **다음 세션의 진입 지점**. 이 문서 + `git status` + `git log -5 --oneline` + `docs/BACKLOG.md` 만 보면 이 세션의 모든 결정·작업·잔여 사항을 완벽히 이어받을 수 있다.
>
> **표기 약속**: 이 문서 안의 `<REPO_ROOT>` 는 buddy 레포의 working tree 루트, `<HOME>` 은 OS 사용자의 홈 디렉토리. 본 머신에서는 각각 `/Users/wm-it-22-00661/Work/github/study/ai/buddy` 와 `/Users/wm-it-22-00661` 이지만, *다른 머신의 reader 는 자신의 경로로 substitution* 하면 된다.

---

## 0. 다른 머신 / 새 환경에서 시작하는 경우 (전제 조건)

같은 머신·같은 세션이면 §1 로 직진해도 된다. *다른 머신* 이거나 *clone 된 신선한 working tree* 라면 아래 항목을 먼저 확인.

| 도구 / 자원 | 필요성 | 부재 시 대응 |
|------------|--------|--------------|
| `git` + remote 접근 | 필수 — 본 세션의 모든 commit 이 `origin/main` 에 push 됨 (HEAD `1758d7a` 이후). `git pull --ff-only origin main` 으로 동기 | 다른 방법 없음 |
| Go toolchain (1.25) | 필수 — §10 의 무결성 검증 (`go vet` / `go build`) + 모든 후속 작업 | brew/asdf/공식 설치 |
| `ripgrep` (`rg`) | 권장 — §8 의 ADR ref 잔여 검증, B 카테고리 grep | 부재 시 `grep -rn` 로 대체 |
| `gh` (GitHub CLI) | 선택 — E2 release / PR 작업 시 | 부재 시 web UI |
| GPG key (commit signing) | 조건부 — 본 레포가 GPG signed commit 강제하면 필수. memory `gpg-signing-setup.md` 참고 | key 동기 또는 임시 `git -c commit.gpgsign=false commit ...` |
| buddy plugin | 조건부 — `/buddy:*` slash command 사용 시 필수. 설치 위치 예: `<HOME>/.claude/plugins/cache/buddy/buddy/<version>/` | 부재 시 `/buddy:*` 사용 옵션 모두 *수동 절차* 로 대체 |
| `~/.claude/CLAUDE.md` 글로벌 룰 | 권장 — 한국어 응답 / commit attribution 금지 / 개발단계 용어 금지 등. 본 문서 §7.1 에 핵심 재명시 | dotfiles 동기 또는 §7.1 만 신뢰 |
| 본 머신 한정 자원 (다른 머신엔 *없음*, 참조 안 됨) | — | jsonl transcript (`<HOME>/.claude/projects/.../*.jsonl`), 본 세션 added memory file (§ 부록 B). 다른 머신에서는 §5 Step 3 의 *jsonl 일괄 리뷰* 옵션 불가 — *재실행* 또는 *단일 axis* 만 사용 |

### 0.1 다른 머신에서 첫 5분 권장 sequence

```bash
# 1. clone 또는 pull
git clone <remote-url> <REPO_ROOT>     # 처음이면
cd <REPO_ROOT>
git fetch --all
git pull --ff-only origin main         # 본 세션 5 commit 포함된 상태로 동기

# 2. 핸드오프 + SSoT 읽기
cat docs/notes/2026-05-23-session-handoff-a-cleanup.md
cat docs/BACKLOG.md | head -120

# 3. 무결성 검증 (이전 세션의 cleanup 이 깨지지 않았는지)
go vet ./...
go build ./...

# 4. 본 세션의 cleanup 결과 spot check
rg --type go -g '!*_test.go' 'ADR-[0-9]' | wc -l   # = 0 이어야 함
```

### 0.2 Memory 동기 (선택)

본 세션이 *이 머신의 user-global memory* 에 한 항목 추가 (`feedback-background-task-emoji.md` + `MEMORY.md` index 갱신). **이 memory 는 git 으로 sync 되지 않음** — 다른 머신은 *별도 복사* 또는 *해당 규칙 모르고 진행*.

- (a) 복사 원할 때: 부록 B 의 두 파일 내용을 다른 머신의 `<HOME>/.claude/projects/-Users-XXX-...-ai-buddy/memory/` 에 동일 구조로 복사
- (b) Skip: 다른 머신에서는 🔄 emoji 규칙 없이 진행해도 무방 (cosmetic 표기 규칙 한 건)

---

## 1. 한 줄 요약

이전 세션 (W3-W5 sequential refactor, 9 commits push) 의 **잔여 "A 카테고리"** (A1/A2/A3) 를 본 세션에서 완전히 마무리. 단, **commit 은 본 세션에서 하지 않음** — 사용자가 "여기까지 진행하고 나머지는 다른 세션에서" 명시했으므로 *uncommitted 상태로 working tree 보존*. 다음 세션의 첫 step 은 *현 uncommitted 작업물의 commit 분리* 다.

---

## 2. 현재 working tree 상태 (2026-05-23 세션 종료 시점)

| 항목 | 값 |
|------|---|
| 작업 디렉토리 | `<REPO_ROOT>` (본 머신 예시: `/Users/wm-it-22-00661/Work/github/study/ai/buddy`) |
| Branch | `main` |
| Last pushed commit | `f2fe583` — `refactor(agent): extract self-check, parsed-output log, and backoff helpers from runOneStep` |
| Remote sync | origin/main 과 동기 (last push) |
| Uncommitted modified files | 약 38개 (A2 cleanup 28 + 외부 작업물 9 + docs/BACKLOG.md + 본 핸드오프 문서) |
| Untracked files | 3개 (외부 작업물) |

> *이 §2 의 "Uncommitted modified files" / "Untracked files" 카운트는 handoff 작성 시점의 snapshot. 이후 본 세션의 commit + 외부 사용자 commit 으로 모두 처리됨 — §14 Addendum 참고.*

확인 명령 (`<REPO_ROOT>` 는 자신의 buddy 레포 root 로 치환):
```bash
cd <REPO_ROOT>
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

- **B 일괄 리뷰** *(본 머신 한정)*: 이전 5-skill review 결과 (이전 세션의 jsonl 에 있음, 경로 예: `<HOME>/.claude/projects/-Users-<HOME-USERNAME>-Work-github-study-ai-buddy/a5ed3598-c0b5-404d-a62c-80a2c36f2de1.jsonl`) 를 다시 읽어 정리. **다른 머신에서는 jsonl 이 없으므로 이 옵션 불가** — 아래 두 옵션 중 택일.
- **B 재실행**: `/buddy:parallel measure-code-health,review-architecture,review-engineering,audit-test-coverage-meaningful,classify-review-risks -- "..."` — W3-W5 + A2 cleanup 후의 새 baseline 으로 delta 측정. Token 비용 큼. *모든 머신에서 가능*.
- **B 단일 axis**: 사용자가 가장 신경 쓰는 한 axis 만 (예: `/buddy:run review-engineering`). *모든 머신에서 가능*.

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

> *모든 명령에서 `<REPO_ROOT>` 는 자신의 buddy 레포 root 로 치환. 본 머신 예시: `/Users/wm-it-22-00661/Work/github/study/ai/buddy`. 다른 머신은 자신의 경로.*

```bash
# 1. 작업 디렉토리 진입 (다른 머신이면 §0.1 의 clone/pull 선행)
cd <REPO_ROOT>

# 2. 현재 상태 확인
git status
git log -5 --oneline
git diff --stat | head -40

# 3. 이 핸드오프 문서 읽기 (가장 먼저)
cat docs/notes/2026-05-23-session-handoff-a-cleanup.md

# 4. SSoT 확인
cat docs/BACKLOG.md | head -120

# 5. 변경된 핵심 파일 sample diff (본 세션 commit 이후엔 git show 로 봄)
git show 96c989f -- internal/mcp/server.go | head -50
git show 96c989f -- internal/sessions/sessions.go | head -50
git show 85c2745 -- docs/BACKLOG.md

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

본 세션에서 추가된 memory 파일 (본 머신 한정):

- 경로: `<HOME>/.claude/projects/-Users-<HOME-USERNAME>-Work-github-study-ai-buddy/memory/feedback-background-task-emoji.md` (본 머신 예시 `<HOME>` = `/Users/wm-it-22-00661`)
- index 항목 (같은 디렉토리의 `MEMORY.md` 끝줄):
  ```
  - [Background task emoji](feedback-background-task-emoji.md) — 진행 중 background 작업은 status 표에서 🔄 이모지 prefix/suffix 로 표기
  ```

### B.1 적용 범위

| 시나리오 | 자동 적용 여부 | 조치 |
|---------|--------------|------|
| 같은 머신, 다음 세션 | ✅ 자동 — Claude Code 가 user-global memory 를 세션 시작 시 자동 로드 | 없음 |
| 다른 머신, 새 세션 | ❌ **자동 안 됨** — memory 파일이 git sync 대상이 아니라 *각 머신 한정*. 해당 머신에는 파일 자체가 없음 | (a) 위 두 파일 내용을 다른 머신의 동일 위치에 복사하거나, (b) 규칙 자체를 skip — cosmetic 표기 한 건이므로 진행에 본질적 영향 없음 |

### B.2 Memory 파일의 본문 (다른 머신에서 수동 복사할 때 그대로 붙여넣기 가능)

`feedback-background-task-emoji.md`:

````markdown
---
name: feedback-background-task-emoji
description: Background 작업(주로 Agent run_in_background=true) 이 진행 중인 task 는 status 표/리스트 표기 시 task 내용 옆에 🔄 이모지를 prefix 또는 suffix 로 붙인다. Foreground in_progress 와 시각적으로 구분 가능하게 한다.
metadata:
  type: feedback
---

Background 작업(주로 `Agent(run_in_background=true)` 또는 `Bash(run_in_background=true)`) 이 진행 중인 task 항목은 사용자에게 status 보고 시 **task 내용 옆에 🔄 이모지** 를 붙인다 (prefix 또는 suffix 둘 다 가능, table 의 경우 status 열 또는 subject 열 옆).

**Why:** 사용자가 "특정 이모지를 지정해서, 작업 리스트에서 작업 내용 옆에 이모지를 붙여주도록해" 라고 명시. Foreground in_progress (사용자 turn 안에서 동기 진행) 와 background (parent agent 가 다른 작업 동시 가능) 가 시각적으로 구분 안 되면 사용자가 *지금 어떤 작업이 동시에 도는지* 파악 불가.

**How to apply:**
- Background 시작 직후 status 표를 그릴 때, 그 task 행의 subject 또는 status 컬럼에 `🔄` 표시.
- 예: `| A2 — comment cleanup | 🔄 in_progress (background) |`
- Background 완료(또는 사용자가 결과 review 후 종료) 시점에 🔄 제거 + 일반 표기 (✅ / ❌) 로 전환.
- Foreground in_progress 에는 🔄 안 붙임 — 그건 단순 `in_progress` 로 충분.
- TaskCreate 의 subject 자체에 영구로 넣는 게 아니라, *내가 사용자에게 보여주는 markdown 표* 안에서만 사용. Task tool 데이터에는 cleaner subject 유지.
````

---

## 14. Addendum (handoff 작성 직후 발생한 사실)

본 문서가 commit 된 직후, 본 세션 외부에서 일어난 두 변화를 기록한다. **§2 / §4 / §5 / §10 / §11 의 일부 기술은 *handoff 작성 시점의 snapshot* 이며 아래 사실이 이를 무효화 / 보완한다**.

### 14.1 외부 작업물이 사용자 직접 commit 으로 흡수됨

- Commit: `54efc0f` — `feat(finish-development-branch): add safe-only PR-flow orchestrator + git-safety-rules SSoT`
- Author: `mhha <mhha@wemade.com>` (= 사용자 본인, 본 세션 외부)
- Timestamp: 2026-05-23 20:03:06 KST
- 흡수된 파일: §4 의 modified 7 + untracked 3 (총 10 파일 + `plugin/skills/finish-development-branch/PROCEDURE.md` 등 신규)

**영향**:
- §4 ("Working tree 의 *외부 작업물*") — 더 이상 working tree 에 *남아있지 않음*. 다음 세션이 *추가 처리할 필요 없음*.
- §5 **Step 2 ("외부 작업물 결정")** — *완료된 상태*. skip 가능. Step 1 (본 세션 commit 분리) → Step 3 (B 카테고리 진입) 로 직진.
- §10 의 "외부 작업물 unchanged" 보장은 *handoff 작성 시점* 까지만 유효 — 이후 사용자가 *직접* commit 함.
- §11 의 commit 명단에 **`54efc0f` 추가**. 다음 세션이 `git log f2fe583..HEAD` 또는 `git log -5 --oneline` 으로 직접 확인 가능.

### 14.2 본 세션의 3 commit + 외부 1 commit 모두 local 상태 (push 안 됨)

본 세션 commit 3건이 추가로 만들어진 상태:

```
d2fb08f docs(notes): session handoff for 2026-05-23 cleanup work
85c2745 docs(backlog): register config namespace grouping as indefinite defer
96c989f docs: paraphrase stale dev-tracker references out of Go comments
54efc0f feat(finish-development-branch): add safe-only PR-flow orchestrator + git-safety-rules SSoT  ← 사용자 직접
f2fe583 (origin/main) refactor(agent): extract self-check, parsed-output log, and backoff helpers from runOneStep
```

→ `main` 은 `origin/main` 보다 **4 commit ahead**. **Push 는 본 세션에서 진행하지 않음** (사용자 결정). 다음 세션이 `git push` 결정.

### 14.3 다음 세션이 실제로 봐야 할 §5 Step 순서 (Addendum 반영)

- ~~Step 1: 본 세션 작업물 commit 분리~~ ✅ **이미 완료** (`96c989f` / `85c2745` / `d2fb08f`).
- ~~Step 2: 외부 작업물 결정~~ ✅ **사용자 직접 commit 으로 처리됨** (`54efc0f`).
- **Step 3 (실질 첫 행동)**: 사용자에게 *push 여부* 확인 → push 진행 또는 skip → 그 후 **B 카테고리 진입** (5-skill review 잔여 findings).

### 14.4 Working tree 현 상태 (handoff 작성 후 + 본 addendum 작성 전 시점)

```
git status            → 작업 폴더 깨끗함 (untracked 0건, modified 0건)
git log origin/main..HEAD  → 4 commit (위 14.2 의 4건)
```

본 addendum 까지 commit 하면 5 commit ahead 가 된다.

