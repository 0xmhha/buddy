# finish-development-branch — 개발 브랜치 종료 sub-orchestrator (safe-only)

## 1. 목적

§5 `build-feature` 의 "완료 기준" 7 항목 통과 후 *개발 브랜치를 PR 로 종료* 하는 4-stage chain (§7-1 Release Preparation) + Pre-flight remote sync 를 *하나의 sub-orchestrator* 로 묶음. *force-류 명령 절대 금지* — 모든 git 작업은 [`router/references/git-safety-rules.md`](../router/references/git-safety-rules.md) 준수.

진입 조건: §5 `build-feature` 완료 기준 통과 (PR-ready branch).
산출물: PR URL + updated CHANGELOG entry + synced docs + quality-gate report + mergeable=CLEAN 검증.
다음 phase: §7 `ship-release` 의 §7-2 (Pre-Launch Safety Nets) 또는 PR review 대기.

> *Anti-rationalization 정합*: 본 sub-orchestrator 는 **결정 (A2) / 리뷰 (HIGH) / 완료 발화 (MID-1) / git 실행 (MID-4)** 4-layer 의 *git 실행 layer* 진입점. *Iron Law* ([`router/references/verification-discipline.md`](../router/references/verification-discipline.md)) + *Safety Rules* 동시 활성.

## 2. 사용 시점

- §5 `build-feature` 완료 기준 7 항목 통과 직후 (권장 진입점)
- §7 `ship-release` 의 §7-1 단계 일괄 호출 (4 stage 개별 호출 대체)
- 긴 PR 대기 후 *base drift* 발생 시 재호출 (sync + re-push)

### 사용 부적합

- 개발 브랜치가 *PR-ready 아닌 상태* — build-feature 완료 기준 미통과
- 사용자가 *PR 생성 전 추가 작업* 의도 — sub-orchestrator 가 PR 만들어버리면 의도 충돌
- *push 된 commit 의 history rewrite 의도* (squash / amend) — 본 sub-orchestrator 범위 밖, 사용자 직접 수행

## 3. 입력

### 필수

- 현재 작업 branch (git current branch)
- base branch (default: `main`, env `BUDDY_DEFAULT_BASE` override 가능)
- 변경 요약 1 줄 (PR title 후보)

### 선택

- `--dry-run`: 실제 push / PR 생성 없이 *각 stage 의 산출물* 만 출력
- `--skip-sync`: Stage 0 (pre-flight sync) 건너뜀 — 사용자가 *이미 sync 함* 을 명시한 경우만
- `--skip-changelog`: 사용자 가시 변경 없을 때 (refactor / docs only)
- `--draft`: PR 을 draft 상태로 생성

## 4. 핵심 원칙

1. **Force-류 명령 자동 실행 절대 금지** — `--force` / `--force-with-lease` / `rebase pushed branch` / `reset --hard` / `clean -fd` 등. [`git-safety-rules.md`](../router/references/git-safety-rules.md) §2 전체 준수.
2. **STOP 우선 — 자동 복구 금지** — sync conflict / push reject / merge conflict 발생 시 STOP + 사용자 정보 제공만. 추측 resolve 금지.
3. **Read-first** — Stage 0 시작 시 `git fetch` + `git status` + `git rev-list --count` 같은 read-only 명령으로 *현재 상태 파악* 후 write 명령 실행.
4. **dry-run 우선 권장** — 첫 호출 시 `--dry-run` 권장. PR / push 는 rollback 비용 있음.
5. **Iron Law 정합 발화** — "PR 생성 완료" 발화 직전 `gh pr view <url> --json mergeable,mergeStateStatus` 응답 확인. `mergeable: MERGEABLE` + `mergeStateStatus: CLEAN` 일 때만 발화.
6. **사용자 명시 승인 시에만 force 류 사용** — 사용자가 squash 후 force-with-lease 같은 명시 의도를 표명하면 *그 명령만 수동 실행 안내*. sub-orchestrator 가 직접 force 명령 실행 금지.

## 5. Stage 흐름

```
finish-development-branch (sub-orchestrator)
├── Stage 0: pre-flight-sync         (신규) fetch + diverged check + safe merge only
├── Stage 1: setup-quality-gates     [기존] typecheck / lint / test / secret scan
├── Stage 2: write-changelog         [기존] semver + CHANGELOG release-summary
├── Stage 3: sync-release-docs       [기존] code change 대비 docs drift audit
├── Stage 4: auto-create-pr          [기존] commit → branch push → PR 생성
└── Stage 5: verify-mergeable        (신규) gh pr view mergeable=CLEAN 확인 (Iron Law)
```

## 6. 실행 절차

### Stage 0: Pre-flight Sync

> [`git-safety-rules.md`](../router/references/git-safety-rules.md) §4.1 *안전 절차* 준수.

```bash
# 0a. fetch (read-only)
git fetch origin <base>

# 0b. diverged 측정 (read-only)
counts=$(git rev-list --left-right --count HEAD...origin/<base>)
# 결과: "ahead\tbehind" (탭 분리)
```

분기:

| ahead | behind | 처리 |
|-------|--------|------|
| 0 | 0 | 완전 동기. Stage 1 진행 |
| N | 0 | 우리만 앞섰음. Stage 1 진행 |
| 0 | N | 우리만 뒤. `git merge origin/<base>` (fast-forward) |
| N | M | 양쪽 diverged. `git merge origin/<base>` (merge commit) |

```bash
# 0c. merge 필요 시
git merge origin/<base>

# 0d. merge conflict 발생
#     → STOP. 사용자에게 정보 제공만:
#       - git diff --name-only --diff-filter=U  # 충돌 파일
#       - 각 파일의 ours / theirs / base hunk 표시
#       - 자동 resolve 시도 금지
```

merge conflict 시 출력 예시 (사용자 처리용):

```
[STOP] Merge conflict 발생 — 자동 진행 불가

충돌 파일:
- src/auth/login.ts (3 hunks)
- docs/api/auth.md (1 hunk)

가능 시나리오:
1) 다른 작업자가 같은 영역 수정 후 base 에 merge → 양쪽 의도 확인 후 manual resolve
2) 자동 generation (codegen) 결과가 양쪽 다름 → re-generation 후 resolve
3) 의도 충돌 (다른 actor track 가 같은 schema 수정) → §4 plan-build 재검토

복구 절차:
- 사용자가 직접 conflict resolve → git add → git merge --continue
- 또는 git merge --abort (merge 전 상태로 안전 복귀)
```

### Stage 1: Quality Gate

```bash
# Stage 0 의 merge 가 있었으면 base 변경이 test 깰 수 있음 — quality gate 재실행
```

`setup-quality-gates` skill 호출 — 이미 5단계에서 설정한 husky / lint-staged 활성 여부 확인:

- [ ] typecheck 통과
- [ ] lint 통과
- [ ] all tests 통과
- [ ] security scan (secret leak 없음)

실패 시 STOP. 사용자에게 실패 항목 + 출력 로그 제공.

### Stage 2: Changelog

`write-changelog` skill 호출 — semver 결정 + CHANGELOG release 섹션 + user-facing change summary.

`--skip-changelog` 옵션 시 건너뜀 (refactor / docs-only PR).

### Stage 3: Doc Sync

`sync-release-docs` skill 호출 — code change diff 기반 affected docs 식별 + auto-update 또는 ask 결정.

> docs 변경이 발생하면 *추가 commit* 으로 분리. PR 본문에 명시.

### Stage 4: PR 생성

`auto-create-pr` skill 호출. *force 옵션 없는 push*:

```bash
git push origin <branch>     # force / force-with-lease 없음
```

push reject 시 [`git-safety-rules.md`](../router/references/git-safety-rules.md) §4.2 절차:

```
[STOP] push reject — 자동 진행 불가

git output:
<git push 의 stderr 인용>

가능 시나리오:
1) Remote 가 local 보다 앞섰음 (다른 작업자 push)
   → 안전 복구: Stage 0 재실행 (sync + 재push)
2) Branch 가 protected (force / direct push 차단)
   → 안전 복구: PR 만 가능, push 방식 변경
3) History 가 rewrite 됨 (다른 누군가 force push)
   → 위험 상황. 즉시 사용자 조사 필요

자동 진행 금지. 사용자 처리 필요.
```

PR 생성 (gh CLI):

```bash
gh pr create \
  --title "<title>" \
  --body-file <body.md> \
  --base <base> \
  $([ "$DRAFT" = "1" ] && echo "--draft")
```

### Stage 5: Mergeable 검증 (Iron Law)

> [`verification-discipline.md`](../router/references/verification-discipline.md) Iron Law 적용. "PR 생성 완료" 발화 직전 fresh verify.

```bash
gh pr view <pr-url> --json mergeable,mergeStateStatus
```

분기:

| mergeable | mergeStateStatus | 의미 | 행동 |
|-----------|------------------|------|------|
| MERGEABLE | CLEAN | 충돌 없음, CI 통과 | "PR 생성 완료" 발화 OK |
| MERGEABLE | BLOCKED | 충돌 없음 but reviewer 필요 / branch protection | "PR 생성 완료, review 대기" 발화 |
| MERGEABLE | BEHIND | base 가 앞섰음 (PR 만든 사이 base 변경) | Stage 0 재실행 또는 사용자에게 "base 가 또 움직임" 안내 |
| CONFLICTING | DIRTY | merge conflict 발생 | STOP. 사용자에게 PR 페이지 conflict 영역 안내 |
| UNKNOWN | UNKNOWN | GitHub 가 아직 계산 중 | 30초 대기 후 재시도 (최대 3회), 그래도 UNKNOWN 면 STOP |

## 7. 산출물 (호출자 반환)

```yaml
finish_branch_result:
  base: main
  branch: <current-branch>
  pre_flight_sync:
    diverged: ahead=N behind=M
    action: merged | fast-forward | no-op | STOP-conflict
    conflict_files: []        # STOP 시 파일 목록
  quality_gate:
    typecheck: pass | fail
    lint: pass | fail
    tests: <X>/<Y> pass
    secret_scan: clean | leaked
  changelog:
    semver: PATCH | MINOR | MAJOR
    entries_added: N
    skipped: false
  docs_sync:
    drift_detected: <N> files
    auto_updated: <N>
    requires_manual: <N>
  pr:
    url: "https://github.com/owner/repo/pull/123"
    number: 123
    title: "<title>"
    draft: true | false
  mergeable:
    state: MERGEABLE | CONFLICTING | UNKNOWN
    status: CLEAN | BLOCKED | BEHIND | DIRTY | UNKNOWN
    verified_at: <timestamp>     # Iron Law evidence
  total_duration_sec: <N>
  dry_run: false
```

> `mergeable.state` / `mergeable.status` / `mergeable.verified_at` 은 *실제 `gh pr view` 응답 확인 후* 에만 채움 (verification-discipline Iron Law).

## 8. 검증 (Pre-Completion Checklist)

- [ ] Stage 0 fetch 실행 + diverged 측정 완료
- [ ] 필요 시 merge 만 수행 (rebase / force 사용 안 함)
- [ ] merge conflict 시 STOP 후 사용자에게 정보 제공 (자동 resolve 안 함)
- [ ] Stage 1 quality gate 통과 (Stage 0 merge 후 재실행 포함)
- [ ] Stage 4 push 시 force / force-with-lease 미사용
- [ ] push reject 시 STOP + git-safety-rules §4.2 절차 따름
- [ ] Stage 5 mergeable 응답 확인 후에만 "PR 생성 완료" 발화 (Iron Law)
- [ ] 시크릿 / PII commit 없음 (write-changelog + auto-create-pr 가 secret-scan stage 포함)
- [ ] dry-run 결과 사용자 confirm 후 실제 실행 (대규모 변경 시)

## 9. 다음 phase

- `/buddy:ship-release` — §7-2 Pre-Launch Safety Nets 진입 (release 흐름 계속)
- PR review 대기 — reviewer assign 후 *passive* 상태
- merge 후 → `automate-release-tagging` (§7-4) 호출 (별도 stage)

## 10. Anti-patterns

1. **force-with-lease 사용** — "더 안전한 force" 라는 잘못된 인식. 자동화 시스템에선 STOP 으로 처리
2. **rebase pushed branch 자동 시도** — SHA rewrite cascade → force push 필요 → destructive
3. **자동 conflict resolve** — 의도 모르는 resolve = silent corruption
4. **push reject 후 자동 force 전환** — 가장 흔한 사고. 절대 금지
5. **`git pull --rebase` 자동 사용** — pushed commit 있으면 rewrite 발생. `git pull --no-rebase` (merge mode) 만 안전
6. **Stage 0 skip 후 push** — base 와 diverged 인 상태에서 push 시도 → reject 또는 PR conflict
7. **Iron Law 위반 발화** — `gh pr view` 응답 확인 없이 "PR 생성 완료" 발화
8. **dry-run 생략 + 대규모 변경 push** — 미리 검토 없이 PR 생성 → 잘못된 base 등 발견 시 PR 닫기 / 재생성 필요
9. **`--skip-sync` 남용** — Stage 0 의도가 *모든 PR 의 conflict 사전 예방*. skip 은 사용자가 *이미 sync 함* 을 명시한 경우만
10. **CHANGELOG 누락 채로 release-bound PR** — write-changelog 실패 시 STOP, skip 안 함

## 11. 참조

- [`router/references/git-safety-rules.md`](../router/references/git-safety-rules.md) — 자동화 git 안전 원칙 SSoT (force 금지 / safe 대체 / STOP 절차)
- [`router/references/verification-discipline.md`](../router/references/verification-discipline.md) — Iron Law (Stage 5 mergeable 검증)
- [`auto-create-pr/PROCEDURE.md`](../auto-create-pr/PROCEDURE.md) — Stage 4 위임 대상
- [`write-changelog/PROCEDURE.md`](../write-changelog/PROCEDURE.md) — Stage 2 위임 대상
- [`sync-release-docs/PROCEDURE.md`](../sync-release-docs/PROCEDURE.md) — Stage 3 위임 대상
- [`setup-quality-gates/PROCEDURE.md`](../setup-quality-gates/PROCEDURE.md) — Stage 1 위임 대상
- [`ship-release/PROCEDURE.md`](../ship-release/PROCEDURE.md) — §7-1 단계의 *권장 호출 패턴*. 본 sub-orchestrator 가 §7-1 의 chain 옵션을 *명시화*
- [`guard-destructive-commands/PROCEDURE.md`](../guard-destructive-commands/PROCEDURE.md) — runtime hook 기반 force / destructive 명령 차단 (본 reference 와 보완)
