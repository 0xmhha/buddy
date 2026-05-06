---
name: automate-release-tagging
description: "merged PR set으로부터 semver auto-decision (breaking change 감지 → MAJOR, feat → MINOR, fix → PATCH), git tag 생성, GitHub Release publish, release branch 전략 관리, 이전 tag와의 비교 changelog 자동. 트리거: 'release 만들자' / '버전 태그 찍어' / 'GitHub Release 게시' / 'semver 결정' / 'release branch' / 'tag v1.2.0' / 'release publish'. 입력: base branch, 이전 tag, merged PR set, release notes 위치. 출력: semver verdict + git tag + GitHub Release + release branch (필요 시). 흐름: auto-create-pr/write-changelog → automate-release-tagging → design-deploy-strategy."
type: skill
---

# Automate Release Tagging — Semver 결정 + Tag + Release Publish

## 1. 목적

merged PR set을 분석해 **semver 자동 결정**(breaking → MAJOR / feat → MINOR / fix → PATCH), **git tag 생성**, **GitHub Release publish**, **release branch 전략**을 자동화한다.

`write-changelog`가 CHANGELOG 본문을 작성한다면, 이 스킬은 **release artifact 자체** (tag + GitHub Release + branch)를 만든다.

핵심 가치: **release 작업 시간 30분 → 2분**, 일관된 semver decision, 모든 release가 audit 가능 artifact.

## 2. 사용 시점 (When to invoke)

- sprint / milestone 종료 후 release 준비
- hotfix 후 patch release
- breaking change merge 후 major version bump
- monorepo의 package별 독립 release
- feature flag → general availability 전환
- security patch 긴급 release
- pre-release (alpha / beta / rc) 발행

## 3. 입력 (Inputs)

### 필수
- base branch (main / develop / release/x.y)
- 이전 release tag (`git describe --tags --abbrev=0`)
- merged PR set (since last tag)
- release notes 위치 (CHANGELOG.md / GitHub Release body)

### 선택
- release branch 전략 (trunk-based / git-flow / GitHub flow)
- pre-release 여부 (alpha / beta / rc)
- monorepo package scope
- signing key (signed tag)

### 입력 부족 시 forcing question
- "이전 tag 뭐야? 첫 release면 v0.1.0 / v1.0.0 시작 결정 필요."
- "trunk-based야 git-flow야? release branch 만들지 결정."
- "monorepo야? package별 independent versioning?"
- "pre-release야 stable release야? alpha / beta / rc / stable?"

## 4. 핵심 원칙 (Principles)

1. **Semver 자동 결정** — conventional commit type + breaking marker → MAJOR/MINOR/PATCH 매핑.
2. **Breaking change 명시 우선** — `BREAKING:` prefix or `BREAKING CHANGE:` footer 검출 → MAJOR 강제.
3. **Pre-release는 별도 tag** — `v1.2.0-alpha.1`, `v1.2.0-beta.1`, `v1.2.0-rc.1`. 본 release와 분리.
4. **Tag는 immutable** — 한 번 push된 tag는 재생성 금지. 잘못되면 새 tag.
5. **Signed tag 권장** — `git tag -s` GPG 서명 (release 진위 검증).
6. **GitHub Release body는 CHANGELOG와 동기** — drift 방지.
7. **Release branch는 long-lived 시 유지** — `release/1.x` 같은 maintenance branch는 보존.
8. **Monorepo는 prefix tag** — `frontend@v1.2.0`, `backend@v2.0.0`.

## 5. 단계 (Phases)

### Phase 1. Previous Release Detection
```bash
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
# 첫 release면 NULL → v0.1.0 / v1.0.0 결정
```

### Phase 2. Commit Range Analysis
이전 tag → HEAD 범위 commit 분석:
```bash
git log $PREV_TAG..HEAD --pretty=format:"%H %s"
```

각 commit:
- conventional commit type 추출 (feat / fix / docs / refactor / etc.)
- breaking change marker 검출:
  - title `BREAKING:` prefix
  - body `BREAKING CHANGE:` footer
- merged PR 번호 추출
- author / co-authors

### Phase 3. Semver Decision
규칙 (우선순위 순):
1. **MAJOR**: breaking change 1건+
2. **MINOR**: feat / perf 1건+ (breaking 없음)
3. **PATCH**: fix / refactor / docs / chore만

자동 결정 후 user confirmation 옵션.

### Phase 4. Pre-release 결정
- `alpha`: 활발한 개발 중, breaking 가능
- `beta`: feature complete, bug 가능
- `rc`: 거의 stable, 마지막 검증
- `stable`: production ready

format: `v1.2.0-<channel>.<number>` (예: `v1.2.0-rc.2`)

### Phase 5. Release Branch 결정
- **trunk-based**: tag만, branch 없음
- **GitHub flow**: long-lived 없음, hotfix는 cherry-pick
- **git-flow**: `release/x.y` 일시 branch + main merge 후 삭제
- **maintenance**: `release/1.x` long-lived (보안 패치용)

### Phase 6. Tag 생성
```bash
NEW_TAG="v$NEW_VERSION"
git tag -s "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"
```

monorepo:
```bash
git tag -s "frontend@v1.2.0" -m "frontend release v1.2.0"
git tag -s "backend@v2.0.0" -m "backend release v2.0.0"
```

### Phase 7. GitHub Release Publish
```bash
gh release create "$NEW_TAG" \
  --title "Release $NEW_TAG" \
  --notes-file CHANGELOG_$NEW_TAG.md \
  --target $BASE_BRANCH \
  $([ "$PRE_RELEASE" = yes ] && echo "--prerelease")
```

artifact 첨부 (binary / archive):
```bash
gh release upload "$NEW_TAG" dist/*.tar.gz dist/*.zip
```

### Phase 8. Release Branch 생성 (필요 시)
```bash
# git-flow style: release branch
git checkout -b release/$VERSION
git push -u origin release/$VERSION

# maintenance branch
git checkout -b release/$MAJOR.x
git push -u origin release/$MAJOR.x
```

### Phase 9. Notification
- Slack / Discord webhook
- PR/issue auto-comment ("included in $NEW_TAG")
- email to stakeholder list

## 6. 출력 템플릿 (Output Format)

```yaml
release:
  previous_tag: v1.1.5
  new_tag: v1.2.0
  base_branch: main
  release_branch: null  # trunk-based
  pre_release: false

semver_decision:
  verdict: MINOR
  reasoning: "3 feat commits, 0 breaking, 5 fixes"
  breaking_changes: []
  feat_commits:
    - sha: abc123
      pr: "#142"
      message: "feat(auth): email/password login"
    - sha: def456
      pr: "#145"
      message: "feat(billing): prorated upgrade"
  fix_commits:
    - sha: 789abc
      pr: "#143"
      message: "fix(checkout): prevent double charge"
  user_confirmed: yes

tag:
  name: v1.2.0
  signed: yes
  signed_by: developer@example.com
  message: "Release v1.2.0"
  pushed: yes

github_release:
  url: "https://github.com/org/repo/releases/tag/v1.2.0"
  notes_source: CHANGELOG.md
  is_prerelease: false
  artifacts:
    - dist/app-v1.2.0.tar.gz
    - dist/app-v1.2.0.zip

release_branch:
  strategy: trunk-based
  branch_created: null

monorepo:
  enabled: false
  packages: []

notifications:
  - channel: slack
    target: "#release"
    sent: yes
  - channel: github_pr_comment
    prs: ["#142", "#143", "#145"]
    sent: yes

next_steps:
  - skill: design-deploy-strategy
    action: "Trigger production deploy with v1.2.0"
  - skill: monitor-regressions
    action: "Watch console err / perf delta vs v1.1.5"
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `auto-create-pr` (모든 PR merged 후) — `Skill` tool로 invoke
- 페어: `write-changelog` (CHANGELOG 본문)
- 페어: `sync-release-docs` (release docs drift 검증)
- 다음 단계: `design-deploy-strategy` (tag → 실제 배포)
- 후속: `monitor-regressions` (배포 후 회귀 감시)
- 후속: `summarize-retro` (release 회고)

## 8. Anti-patterns

1. **Manual semver decision** — 일관성 약. conventional commit 기반 자동.
2. **Tag re-push** — `-f` force push. immutable 원칙 위반. 잘못되면 새 tag.
3. **Unsigned tag** — 진위 검증 불가. signed tag 권장.
4. **GitHub Release body 빈 채로** — release notes 0. CHANGELOG 동기 강제.
5. **Pre-release를 stable로 표시** — semver 의미 깨짐. alpha/beta/rc 명시.
6. **Release branch 무한 long-lived** — `release/1.x` 5년 유지 → 유지비용. EOL 정책.
7. **Monorepo single tag** — 모든 package 같이 bump. independent versioning 권장.
8. **Notification 없이 release** — stakeholder 모름. Slack / PR comment 자동.
