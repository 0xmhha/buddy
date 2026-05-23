# Git Safety Rules — 자동화 시스템 금지 / 안전 대체 / STOP 절차

> Cross-skill shared reference. *모든 buddy 스킬이 git 명령을 자동 실행할 때* 본 reference 의 금지/안전 정책 준수 의무. 사용자 명시 승인 없이 destructive 명령 자동 실행 금지.
>
> 보완 관계:
> - **본 reference** = *design-time* 지침 (스킬 본문 작성 시 금지/안전 대체/STOP 절차)
> - **`guard-destructive-commands`** = *runtime* hook 기반 명령 직전 차단 (PreToolUse hook)
>
> 두 layer 모두 *force / rewrite-pushed-history / 자동 복구 시도* 차단. 한 layer 가 누락되면 다른 layer 가 catch.

---

## 1. 핵심 원칙

1. **이미 push 된 commit 의 SHA rewrite 절대 자동 금지** — `rebase` / `amend` / `squash` 가 push 된 branch 에서 발생하면 *force push 필요* → destructive cascade.
2. **모든 force-류 push 자동 금지** — `--force` / `--force-with-lease` / `+refs/...` 모두 동일하게 *금지*. force-with-lease 의 expected-sha 검사는 *덜 위험* 일 뿐 *안전* 아님 (sub-second race window).
3. **자동 복구 시도 금지** — destructive 발생 / push reject / merge conflict 발생 시 *STOP* + 사용자 정보 제공만. `git reflog` 의존 복구는 사용자 책임.
4. **명시적 사용자 승인 시 예외** — 위 금지 명령도 사용자가 *명시적 의도* + *결과 이해* 를 표명한 경우 사용 가능. 단 "그냥 진행해" 같은 모호한 답은 승인 아님.
5. **STOP 시 정보 우선** — 사용자에게 (a) 현재 상태 (b) 발생 가능 시나리오 3-4 (c) 각 시나리오의 안전 복구 절차를 제공. *판단* 은 사용자 몫.

---

## 2. 자동 실행 금지 명령

| 명령 | 위험 | 안전 대체 |
|------|------|----------|
| `git push --force` / `-f` | remote history overwrite, 다른 작업자 commit 손실 | push reject 시 STOP, 사용자 처리 |
| `git push --force-with-lease[=...]` | expected-sha 검사 *후* push — race window 내 silent 손실 가능. AI 자동화에선 reflog 복구 사실상 불가 | 동일하게 STOP |
| `git push +refs/heads/...` | force push 의 다른 형태 | 동일 |
| `git rebase origin/<base>` *push 된 branch 에서* | local SHA rewrite → 이후 push 가 force 필요 → destructive cascade | `git merge origin/<base>` (merge commit, push 시 force 불요) |
| `git rebase -i HEAD~N` *push 된 commit 포함 시* | 동일 cascade | 동일 — push 된 commit 은 절대 자동 rewrite 안 함 |
| `git commit --amend` *push 된 commit 에서* | 동일 cascade | 새 commit 으로 보정 (`fix:` / `revert:` ) |
| `git reset --hard` | uncommitted + commit-pointer 동시 폐기. reflog 없으면 복구 0 | STOP + 사용자 처리. 의도가 명확하면 사용자가 직접 실행 |
| `git checkout .` / `git restore .` | uncommitted 폐기 | 동일 — 자동 사용 금지 |
| `git clean -fd` / `-fdx` | untracked 파일 + 디렉토리 폐기. `-x` 는 gitignore 도 포함 | STOP + 사용자 처리 |
| `git branch -D <name>` | merged 검사 없는 강제 삭제 | `git branch -d` (안전) — fully-merged 만 삭제. 실패 시 STOP |
| `git submodule update --remote --merge --force` | submodule force | STOP + 사용자 처리 |
| `git filter-branch` / `git filter-repo` | 전체 history rewrite | 사용자 명시 의도 필수. 자동 실행 절대 금지 |
| `git gc --prune=now --aggressive` 자동 | reflog object 조기 삭제 → 복구 불가 | 자동 금지 |

---

## 3. 안전 자동 실행 가능 명령

| 명령 | 안전 사유 |
|------|---------|
| `git fetch <remote>` | read-only. local working tree 영향 없음 |
| `git status` / `git status -sb` | read-only |
| `git rev-list --left-right --count A...B` | diverged 측정 — read-only |
| `git log` / `git show` / `git diff` | read-only |
| `git merge <ref>` (no `--squash`, no `--no-ff` 가 정책이면) | merge commit 생성, force 불요. conflict 시 working tree 에 남음 → STOP 가능 |
| `git pull --no-rebase --ff-only` 또는 `--no-ff` | merge mode. force 불요 |
| `git push origin <branch>` (force 없이) | reject 시 STOP — local 영향 없음 |
| `git add <specific-paths>` | 명시 경로만. `git add .` / `-A` 는 *주의* — `.env` / 시크릿 끌고 갈 위험 |
| `git commit -m "..."` | 새 commit 생성 — 안전 |
| `git revert <sha>` | 새 commit 으로 변경 취소 — history rewrite 아님 |
| `git branch <name>` / `git switch -c <name>` | 신규 branch 생성 |
| `git stash push -m "..."` / `git stash list` | stash 보존 — pop 은 conflict 가능하므로 *주의* |
| `git worktree add` / `git worktree list` | 신규 worktree |
| `git worktree remove <path>` *clean worktree* | clean 검사 후 안전 — dirty 면 STOP |

---

## 4. 상황별 안전 절차

### 4.1 PR 생성 직전 base 와 sync

```
1. git fetch origin <base>                                    # read-only
2. counts=$(git rev-list --left-right --count HEAD...origin/<base>)
3. ahead/behind 측정 후 분기:
   - "ahead 0 behind 0"  → 동기됨, 다음 단계
   - "ahead N behind 0"  → 우리만 앞섰음, 다음 단계
   - "ahead 0 behind N"  → 우리만 뒤, git merge origin/<base> (fast-forward)
   - "ahead N behind M"  → 양쪽 diverged, git merge origin/<base> (merge commit)
4. merge 시 conflict 발생:
   - STOP. 사용자에게 conflict 파일 목록 + 각 파일의 양쪽 hunk 정보 제공
   - 자동 resolve 시도 금지 (의도 모르는 conflict resolution = silent corruption)
5. clean merge 후 quality gate 재실행 (base 변경이 test 깰 수 있음)
```

### 4.2 Push reject 시

```
git push origin <branch>     # force 옵션 없음

if reject (non-fast-forward / branch-protection / etc):
  → STOP
  → 사용자에게 정보 제공:
    a) reject 사유 (git output 그대로 인용)
    b) 발생 가능 시나리오:
       - Remote 가 local 보다 앞섰음 (다른 작업자 push) — 안전 절차: §4.1 sync
       - Branch 가 protected (force / direct push 차단) — 안전 절차: PR 만 가능
       - History 가 rewrite 됨 (예: 다른 누군가 force push) — 위험 상황, 조사 필요
    c) 각 시나리오의 안전 복구 절차
  → "자동 진행 금지. 사용자 처리 필요." 명시
```

### 4.3 Merge conflict 발생 시

```
1. git merge / git pull 후 conflict
2. STOP — 자동 resolve 시도 금지
3. 사용자에게:
   - 충돌 파일 목록 (git diff --name-only --diff-filter=U)
   - 각 충돌 hunk 의 (a) base 버전 (b) ours (c) theirs
   - 의도 추측 없이 *사실만* 제공
4. 사용자 명시 지시 후에만 conflict 영역 수정. 추측 resolve 금지
5. 사용자가 "abort" 요청 시: git merge --abort (안전 — merge 시작 전 상태로 복귀)
```

### 4.4 Stash conflict (pop / apply 시)

```
1. git stash pop 후 conflict
2. STOP — stash 자동 drop 금지 (drop 시 손실 가능)
3. git stash apply (drop 안 함) 가 더 안전 — 사용자가 검토 후 명시 drop
```

### 4.5 Dirty working tree 발견 시

```
1. git status 에 uncommitted 변경 발견
2. 자동 commit / 자동 stash / 자동 discard 금지
3. 사용자에게 정보 제공:
   - 변경 파일 목록
   - 각 파일의 변경 line 수
4. 사용자가 처리 방향 결정 (commit / stash / discard / ignore)
```

---

## 5. 사용자 명시 승인 예외 처리

위 §2 금지 명령도 사용자가 *명시적 의도 + 결과 이해* 를 표명하면 사용 가능:

**충분한 승인 예시**:
- "내 solo feature 브랜치이고 squash 후 force-with-lease 로 push 해줘"
- "이 commit history 를 rewrite 해서 PR 깔끔하게 해줘 — 다른 사람 작업 없음을 확인했음"

**불충분한 승인 예시** (그대로 진행 금지):
- "그냥 진행해"
- "알아서 해줘"
- "빨리 끝내자"

→ 모호한 답은 *추가 확인* 필요. "이 명령은 X / Y / Z 위험이 있는데 그래도 진행할까요?" 형식으로 명시 재확인.

---

## 6. Cross-link — 어느 스킬이 본 reference 를 참조해야 하나

| 스킬 | 적용 시점 |
|------|----------|
| `finish-development-branch` | Stage 0 (pre-flight sync) / push reject 시 |
| `auto-create-pr` | Phase 1 (Pre-Push 검증) / Phase 3 (Branch Push) |
| `dispatch-parallel-agents` | Phase 7 (Branch Reconcile) — worker branch merge 시 |
| `guard-destructive-commands` | hook patterns 와 본 reference 정합 (runtime ↔ design-time) |
| `compose-safety-mode` | max safety mode 합성 시 git-safety-rules 도 활성 |
| 모든 git 명령 자동 실행 스킬 | 명령 실행 전 §2 금지 목록 매칭 검사 |

각 스킬 본문에 1줄 cross-reference 추가:

```markdown
> 본 stage / phase 의 git 명령은 [`router/references/git-safety-rules.md`](../router/references/git-safety-rules.md) 의 §2 금지 / §3 안전 / §4 절차 준수. force-류 / rewrite-pushed / 자동 복구 시도 금지.
```

---

## 7. A-카테고리 anti-rationalization 시스템 정합

| 시점 | 게이트 | 위치 |
|------|--------|------|
| **결정** (design) | 3+ orthogonal 후보 강제 | 7 design 스킬 (A2) |
| **리뷰** (plan) | 5 anti-rationalization 규칙 | review-engineering (HIGH) |
| **완료 발화** (claim) | Iron Law + Gate Function | verification-discipline (MID-1) |
| **git 실행** (action) | 금지/안전/STOP 절차 | *본 reference* (MID-4 보강) |

git destructive 명령은 *완료 발화 게이트* 보다 더 강한 게이트가 필요 — 발화는 "되돌릴 수 있는 거짓말" 이지만 destructive git 은 *되돌릴 수 없는 손실*. 따라서 *별도 layer* 로 격리.

---

## 8. Bottom Line

**자동화 시스템 git 원칙**: read-first / write-cautiously / force-never / recover-by-user.

명령 실행 전 §2 매칭 검사. STOP 발생 시 정보 제공만, 추측 진행 금지.

이건 협상 대상 아님.
