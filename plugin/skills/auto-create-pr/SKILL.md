---
name: auto-create-pr
description: "feature/task 완료 후 commit → branch push → PR 생성을 자동화. PR title/body 표준 (summary, test plan, screenshots, breaking change notice), reviewer 지정, label 자동 부여, 관련 issue 링크. 트리거: 'PR 만들어줘' / 'pull request 생성' / 'PR 자동화' / 'feature ship 준비' / '리뷰 요청' / 'PR description 작성' / 'auto PR'. 입력: feature_id + branch + 변경 file + acceptance criteria + 관련 issue. 출력: PR URL + status check 요약 + reviewer 지정. 흐름: build-with-tdd/dispatch-parallel-agents → auto-create-pr → automate-release-tagging."
type: skill
---

# Auto Create PR — PR 자동 생성 + 표준 description

## 1. 목적

feature/task 구현 완료 후 **commit → branch push → PR 생성**을 자동화. PR description을 표준 format으로 채우고, reviewer / label / linked issue를 자동 지정한다.

핵심 가치: **PR 작성 시간 5-10분 → 30초**, 모든 PR이 일관된 품질 (summary / test plan / screenshots / breaking change).

## 2. 사용 시점 (When to invoke)

- `build-with-tdd` 또는 `dispatch-parallel-agents` 완료 후 ship 단계
- bug fix 완료 후 PR 생성
- documentation 변경 후 PR
- dependency upgrade 후 PR
- feature flag 활성화 PR
- regression test 추가 PR (postmortem 후)
- design review 결과 반영 PR

## 3. 입력 (Inputs)

### 필수
- feature_id 또는 task_id
- branch name
- base branch (main / develop / release/x.y)
- 변경 file 목록
- acceptance criteria (PR description에 들어감)

### 선택
- 관련 issue 번호 (`Closes #123`)
- screenshots (UI 변경 시)
- breaking change 여부
- reviewer 명단
- label
- draft PR 여부

### 입력 부족 시 forcing question
- "어느 base branch에 머지해? main? develop? release/x.y?"
- "breaking change야? PR title에 `BREAKING:` prefix + body에 migration 가이드."
- "UI 변경 있어? screenshot 첨부 강제."
- "관련 issue 있어? `Closes #N` 자동 링크."

## 4. 핵심 원칙 (Principles)

1. **Conventional commit prefix** — `feat:` / `fix:` / `refactor:` / `docs:` / `test:` / `chore:` / `perf:` / `ci:`. PR title도 동일.
2. **PR description은 stakeholder용** — code 자체보다 "왜 / 무엇 / 검증 어떻게" 중심. 비기술자도 이해 가능.
3. **Test plan 강제** — 어떻게 검증했는가 명시. CI 자동 + manual QA 단계 포함.
4. **Breaking change 명시** — title `BREAKING:` prefix + body에 migration guide. semver MAJOR 결정에 사용.
5. **Screenshots는 UI 변경 시 강제** — before/after 쌍이 이상적.
6. **Linked issue 자동** — `Closes #123` / `Fixes #456` 자동 추가. issue tracker와 동기.
7. **Reviewer 코드 영역 기반** — file path → owner mapping (CODEOWNERS 활용).
8. **Draft PR 우선** — incomplete work도 PR로 visibility. ready 상태 시점에 mark ready.

## 5. 단계 (Phases)

### Phase 1. Pre-Push 검증
1. branch가 base와 diverged 했는지 (`git diff base...HEAD --stat`)
2. 미커밋 변경 사항 (`git status --porcelain`) 0건 확인
3. local test 통과 (`npm test` / `pytest`)
4. lint / format 통과
5. branch up-to-date with base (rebase 권장)

### Phase 2. Commit 정리
- atomic commit (한 commit = 한 logical change)
- conventional commit format
- 메시지에 scope 포함 (예: `feat(auth): add email/password login`)
- 필요시 commit squash / interactive rebase

### Phase 3. Branch Push
```bash
git push -u origin <branch>
```
- 첫 push면 `-u` (upstream 설정)
- existing branch면 `--force-with-lease` (다른 사람 변경 보호)

### Phase 4. PR Title 작성
형식: `<type>(<scope>): <description>`

예시:
- `feat(auth): add email/password login flow`
- `fix(billing): prevent double charge on retry`
- `BREAKING(api): rename /v1/users to /v2/users`
- `refactor(core): extract retry logic to utility`

### Phase 5. PR Body 작성 (표준 template)

```markdown
## Summary
<2-3 문장. 무엇을 / 왜 / 어떤 영향>

## What changed
- <파일/모듈별 변경>
- <...>

## Why this approach
<선택한 접근의 이유 + 거부한 대안>

## Test plan
- [x] Unit tests added/updated (`<file>`)
- [x] Integration tests pass
- [ ] Manual QA: <시나리오>
- [ ] E2E test: <테스트 명>

## Screenshots
<UI 변경 시>
| Before | After |
|---|---|
| ... | ... |

## Breaking changes
<있으면 명시 + migration guide>

## Related issues
- Closes #123
- Related: #456

## Checklist
- [x] Tests pass locally
- [x] Lint / typecheck pass
- [x] Documentation updated (if applicable)
- [x] Changelog entry added (if user-facing)
- [ ] Migration guide written (if breaking)
```

### Phase 6. PR 생성 (gh CLI)
```bash
gh pr create \
  --title "<title>" \
  --body-file <body.md> \
  --base <base> \
  --reviewer <reviewers> \
  --label <labels> \
  --draft  # 옵션
```

### Phase 7. Reviewer 자동 지정
1. CODEOWNERS 파일 파싱
2. 변경 file path → owner 매핑
3. owner 중 active member 식별
4. team review 우선 (개인보다 빠른 응답)

### Phase 8. Label 자동 부여
- conventional commit type → label 매핑
  - feat → `enhancement`
  - fix → `bug`
  - docs → `documentation`
  - refactor → `refactor`
- size label (XS / S / M / L / XL by LOC changed)
- domain label (auth / billing / ui)

### Phase 9. CI / Status Check 모니터
- PR 생성 후 CI workflow 자동 실행
- pass / fail 상태 보고
- 실패 시 PR comment로 surface
- 자동 retry는 의도된 경우만 (flaky test ≠ retry)

## 6. 출력 템플릿 (Output Format)

```yaml
pr_creation:
  pr_url: "https://github.com/org/repo/pull/142"
  pr_number: 142
  branch: feature/auth-login
  base: main

title:
  format: "feat(auth): add email/password login flow"
  conventional_commit_type: feat
  scope: auth
  breaking_change: no

body:
  template_version: v1.0
  sections_present:
    summary: yes
    what_changed: yes
    why_approach: yes
    test_plan: yes
    screenshots: yes
    breaking_changes: no
    related_issues: yes
    checklist: yes

linked_issues:
  closes: ["#123"]
  related: ["#456"]

reviewers:
  individual: ["@alice", "@bob"]
  team: ["@org/auth-team"]
  source: CODEOWNERS

labels:
  - enhancement
  - auth
  - size/M

status_checks:
  configured:
    - name: ci/lint
      status: pending
    - name: ci/typecheck
      status: pending
    - name: ci/test
      status: pending
    - name: ci/e2e
      status: pending

draft: false
mergeable: pending
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `build-with-tdd` 또는 `dispatch-parallel-agents` — `Skill` tool로 invoke
- 페어: `setup-quality-gates` (CI 게이트가 PR status check)
- 페어: `write-changelog` (PR description ↔ changelog 정합)
- 다음 단계: `automate-release-tagging` (PR merge 후 tag/release)
- 후속: `monitor-regressions` (배포 후 회귀 감시)

## 8. Anti-patterns

1. **PR title이 unclear** — "fix bug" / "update code" — type / scope / 구체 description 강제.
2. **Body에 test plan 없음** — reviewer가 검증 못함. test plan 강제.
3. **Breaking change 묵시** — 사용자 혼란 + migration 누락. 명시 + guide.
4. **Reviewer 수동 지정** — file 영역 모르고 잘못 지정. CODEOWNERS 자동.
5. **Force push merged branch** — review 흔적 손실. squash/merge는 GitHub에서.
6. **Draft PR 끝까지** — visibility 부족. ready 시점에 mark ready 강제.
7. **CI 실패 무시 PR** — merge 막히지만 알림 없으면 잠수. PR comment로 surface.
8. **Label 없이 PR** — triage 어려움. type / size / domain 자동 부여.
