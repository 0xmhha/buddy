---
name: sync-release-docs
description: code change diff를 기준으로 affected docs를 audit하고 auto-update 또는 ask를 결정한다. feature ship 후, release tag 전, docs drift가 의심될 때 사용한다.
type: skill
---

# Document Release — Diff 기반 문서 Sync


Ship 후 문서 sync. 프로젝트의 모든 문서 파일을 읽고, 브랜치 diff와 cross-reference하며, README / ARCHITECTURE / CONTRIBUTING / 프로젝트 instruction 문서를 실제로 ship된 것과 일치하게 업데이트. CHANGELOG voice를 재작성 없이 polish, 선택적으로 VERSION bump, 파일별 health summary 생성.

스킬은 **대부분 자동화**: diff에서 명백히 따르는 factual 업데이트는 직접 적용. 위험하거나 주관적이거나 narrative 결정만 사용자에게 묻는다.

## 이 스킬을 사용하는 경우

- Feature 브랜치가 merge 준비됐고 문서가 아직 안 건드려짐
- 릴리스 태깅 중이고 CHANGELOG / VERSION을 diff와 reconcile 원함
- 문서가 코드에서 drift (경로, 개수, 명령 테이블, 프로젝트 구조)
- PR 리뷰가 "README가 이걸 언급 안 함" 또는 "ARCHITECTURE 다이어그램이 stale"을 surface
- `/ship` 스타일 flow 직후, PR merge 전

신규 문서 저작(손으로)이나 content 재작성(voice 보존 — 재발명 아님)엔 사용 **안 함**.

## 핵심 원칙

1. **Diff가 source of truth.** 주장된 모든 업데이트는 `git diff <base>...HEAD`의 구체적 변경까지 trace back 필수.
2. **Voice 보존 > factual freshness.** 기존 prose 안의 fact 업데이트. Tone, intro paragraph, philosophy 섹션을 재작성하지 말 것.
3. **Factual은 auto-update, narrative는 ask.** 변경이 "경로 rename"이면 → 그냥 한다. 변경이 "보안 모델 rephrase"면 → 묻는다.
4. **CHANGELOG clobber 금지.** 기존 엔트리 내 wording polish. 삭제, reorder, regenerate, CHANGELOG.md에 `Write` 금지.
5. **VERSION bump 조용히 금지.** 명백해 보여도 항상 질문으로 confirm.
6. **뭐가 바뀌었는지 명시적.** 모든 edit는 사용자가 스캔할 수 있는 한 줄 summary 획득.

## 9단계 워크플로우

### Step 0: 플랫폼과 베이스 브랜치 감지

Git 호스팅 플랫폼을 remote URL에서 감지 (`git remote get-url origin`):

- URL에 `github.com` → **GitHub** (`gh` CLI 사용)
- URL에 `gitlab` → **GitLab** (`glab` CLI 사용)
- 그 외엔 `gh auth status` / `glab auth status`로 self-hosted 인스턴스 체크
- 둘 다 없음 → **unknown** (git 네이티브 명령만 사용)

이 PR/MR이 타겟팅하는 베이스 브랜치 결정, repo 기본값으로 fall back:

1. `gh pr view --json baseRefName -q .baseRefName` (GitHub) 또는 `glab mr view -F json` → `target_branch` (GitLab)
2. `gh repo view --json defaultBranchRef -q .defaultBranchRef.name` (GitHub) 또는 `glab repo view -F json` → `default_branch` (GitLab)
3. Git 네이티브 fallback: `git symbolic-ref refs/remotes/origin/HEAD | sed 's|refs/remotes/origin/||'`
4. 최후 수단: `origin/main`, 그다음 `origin/master`, 그다음 리터럴 `main` 시도.

감지된 베이스 브랜치 출력. 이후 모든 단계에서 `<base>`로 사용.

### Step 1: Diff 스코프

현재 브랜치가 베이스 브랜치이면 abort ("Feature 브랜치에서 실행.").

Diff 컨텍스트 수집:

```bash
git diff <base>...HEAD --stat
git log <base>..HEAD --oneline
git diff <base>...HEAD --name-only
```

변경을 doc 관련 카테고리로 분류:

- **New features** — 새 파일, 새 명령, 새 public surface
- **Changed behavior** — 수정된 API, config 변경, rename된 개념
- **Removed functionality** — 삭제된 파일, 제거된 명령
- **Infrastructure** — 빌드 시스템, 테스트 setup, CI 변경

한 줄 출력: `Analyzing N files changed across M commits.`

### Step 2: Doc 인벤토리

Repo의 모든 문서 파일 발견 (생성, hidden, dependency dir skip):

```bash
find . -maxdepth 2 -name "*.md" \
  -not -path "./.git/*" \
  -not -path "./node_modules/*" \
  -not -path "./.venv/*" \
  -not -path "./dist/*" \
  -not -path "./build/*" \
  | sort
```

`VERSION`과 모든 `CHANGELOG*` 파일을 명시적으로 추가. 리스트 출력: `Found K documentation files to review.`

### Step 3: 파일별 영향 분석

각 문서 파일 전체 읽기. Diff와 cross-reference. 이 generic heuristic 사용 — 어떤 프로젝트에도 적용:

**README.md**
- Diff에 보이는 모든 feature와 capability를 묘사하나?
- Install / setup / quickstart instruction이 여전히 정확?
- 예시 명령, 스크린샷, 데모 설명이 여전히 유효?
- Troubleshooting 단계가 여전히 맞나?

**ARCHITECTURE.md** (또는 design / overview 문서)
- 컴포넌트 다이어그램과 모듈 설명이 현재 코드와 일치?
- 설계 근거("왜 X를 선택했는지") 섹션이 여전히 참?
- 보수적으로 — 아키텍처 문서는 PR마다 flip할 가능성 낮은 것을 묘사. Diff가 명확히 모순하는 것만 업데이트.

**CONTRIBUTING.md** — 새 기여자 smoke test
- Setup instruction을 이 repo를 본 적 없다는 듯 walk. 각 단계가 성공할까?
- 테스트/빌드 명령이 프로젝트에 실제 있는 것(package.json, Makefile 등)과 일치?
- Workflow 설명이 현재?
- 첫 기여자를 실패시키거나 혼란시킬 모든 것 플래그.

**프로젝트 instruction 파일** (예: `CLAUDE.md`, `AGENTS.md`, `.cursorrules`)
- 프로젝트 구조 섹션이 실제 파일 트리와 일치?
- 리스트된 스크립트와 명령이 정확?
- 참조된 경로가 여전히 존재?

**기타 .md 파일**
- 각각 읽고, 목적과 audience 식별.
- Diff가 모순하는 모든 것 플래그.

각 파일에 필요한 업데이트 분류:

- **Auto-update** — Diff에 직접 traceable한 factual 수정 (테이블에 row 추가, 경로 fix, count 업데이트, 프로젝트 트리 새로고침)
- **Ask user** — Narrative 변경, 섹션 제거, security/threat-model 변경, 대규모 재작성(한 섹션 >~10줄), 모호한 관련성, 완전 새 섹션

### Step 4: Auto-update vs Ask 분리

**Auto-update는 `Edit` 도구로 직접 적용.** 수정된 각 파일에 파일이 아니라 **구체적 변경**을 명명하는 한 줄 summary 출력:

> Bad: `Updated README.md`
> Good: `README.md: added /new-command to commands table, updated command count from 9 to 10`

**Auto-update OK:**
- 기존 테이블/리스트에 row 추가
- 파일 경로, 함수 이름, 파일 개수 업데이트
- 프로젝트 구조 트리 새로고침
- 문서 전체의 버전 번호 fix
- rename된 파일로 cross-reference 링크 업데이트
- 삭제된 파일 참조 제거

**Auto-update NOT OK:**
- README 인트로 / 프로젝트 포지셔닝 / "이게 뭐지" 섹션
- ARCHITECTURE 철학이나 설계 근거
- 보안 모델 설명
- Threat model, trust boundary, data flow narrative
- 모든 문서에서 전체 섹션 제거
- Fact가 아니라 의미를 바꾸는 모든 것

각 "Ask user" 항목에 AskUserQuestion 사용. 각 prompt는 포함:
- 프로젝트 이름 + 브랜치 + 리뷰 중 파일 (한 문장 grounding)
- 평문으로 구체적 결정 (jargon 아님)
- 한 줄 reasoning 있는 추천
- `Skip — leave as-is` 선택 포함하는 letter 옵션

각 답변 직후 승인된 변경 즉시 적용.

### Step 5: Voice 보존 체크 (CHANGELOG)

**크리티컬: CHANGELOG 엔트리 clobber 금지.** 여기서 일은 voice polish이지 regeneration 아님.

CHANGELOG가 이 브랜치에서 수정 안 됐으면 이 단계 전체 skip.

CHANGELOG가 수정됐으면 예외 없이 이 규칙 따름:

1. **먼저 전체 CHANGELOG 읽기.** 뭐든 건드리기 전 이미 있는 것 이해.
2. **기존 엔트리 내 wording만 수정.** 엔트리 삭제, reorder, 교체 금지.
3. **엔트리를 처음부터 regenerate 금지.** 엔트리가 ship된 것의 source of truth — 역사를 재작성이 아니라 prose polish.
4. **엔트리가 틀리거나 불완전해 보이면 질문.** 구조적 이슈를 조용히 fix 금지.
5. **정확한 `old_string` 매치로 `Edit` 사용.** CHANGELOG.md에 `Write` 금지 — `Write`는 전체 파일을 덮어쓰고 조용히 content drop.

Voice polish 자체에 **sell test** 적용: 사용자가 각 bullet을 읽으며 "오 좋다, 써보고 싶다"를 생각할까? 아니면 wording 재작성(content 아님):

- 구현 세부가 아니라 사용자가 이제 **할 수 있는** 것 lead
- "You can now..."가 "Refactored the..."를 이김
- Commit 메시지처럼 읽히는 엔트리 플래그
- 내부 / 기여자 변경은 별도 `### For contributors` 서브섹션에
- 사소한 wording 조정은 auto-fix; 의미를 바꾸는 재작성 전에 질문

### Step 6: Cross-doc 일관성 체크

파일별 pass 후 모든 문서에 걸친 sweep:

1. README의 feature 리스트가 프로젝트-instruction 파일(CLAUDE.md / AGENTS.md / 동등물)이 묘사하는 것과 일치?
2. ARCHITECTURE의 컴포넌트 리스트가 CONTRIBUTING의 프로젝트 구조와 일치?
3. CHANGELOG의 최신 버전이 VERSION 파일(둘 다 존재 시)과 일치?
4. **Discoverability:** 모든 문서가 README나 주 프로젝트-instruction 파일에서 도달 가능? `ARCHITECTURE.md`가 존재하지만 어느 entry-point 파일도 링크 안 하면 플래그. 모든 문서는 front door에서 discoverable해야.
5. 명확한 factual 모순(예: 버전 mismatch)은 auto-fix. Narrative 모순은 질문.

### Step 7: TODO / open-work cleanup

`TODOS.md`(또는 동등한 open-tasks 파일) 존재 시 두 번째 pass. 없으면 skip.

1. **아직 표시 안 된 완료 항목.** Diff를 open TODO와 cross-reference. TODO가 이 브랜치 변경으로 명확히 해결되면 버전 + 날짜 있는 `Completed` 섹션으로 이동. 보수적으로 — 명백한 증거 있는 항목만 mark.
2. **설명 업데이트 필요 항목.** TODO가 상당히 변경된 파일이나 컴포넌트를 참조하면 설명이 stale일 수 있음. 업데이트, 완료, 그대로 둘지 질문.
3. **새 연기 작업.** Diff에서 `TODO`, `FIXME`, `HACK`, `XXX` 주석 스캔. 의미 있는 연기 작업 나타내는 각 항목(trivial inline 노트 아님)에 대해 TODOS.md에 캡처할지 질문.

### Step 8: VERSION bump 질문

**크리티컬: 물어보지 않고 VERSION bump 금지.** `VERSION`(또는 `package.json`의 `version` 같은 동등 버전 파일) 없으면 이 단계 조용히 skip.

VERSION이 이미 이 브랜치에서 수정됐는지 체크:

```bash
git diff <base>...HEAD -- VERSION
```

**VERSION이 bump 안 됐으면**, 질문:

- 추천: 문서만 변경이면 `Skip`; 코드 변경과 함께 문서가 ship되면 `Bump PATCH`
- A) Bump PATCH — 코드 변경과 함께 문서 ship
- B) Bump MINOR — 유의미 standalone 릴리스
- C) Skip — 버전 bump 불필요

**VERSION이 이미 bump됐으면** 조용히 skip 금지. Bump가 여전히 전체 변경 스코프를 커버하는지 체크:

1. 현재 VERSION의 CHANGELOG 엔트리 읽기. 어떤 feature를 묘사?
2. Diff 읽기. 현재 엔트리에 언급 안 된 유의미 변경(새 feature, 새 모듈, 주요 refactor)이 있나?
3. **엔트리가 모두 커버하면:** `VERSION: Already bumped to vX.Y.Z, covers all changes.` 출력하고 계속.
4. **유의미 uncovered 변경 존재하면:** 다시 bump할지 질문 — "feature A"에 설정된 VERSION이 B가 자기 엔트리 자격 있을 때 "feature B"를 조용히 흡수하면 안 됨.
   - A) 다음 patch로 bump — 새 변경에 자기 버전 부여
   - B) 현재 버전 유지 — 새 변경을 기존 엔트리로 fold
   - C) Skip — 나중에 처리

### Step 9: Commit, push, 리포트

**먼저 empty 체크.** `git status` 실행. 이전 단계가 문서 파일 수정 안 했으면 `All documentation is up to date.` 출력하고 commit 없이 exit.

**Commit:**

1. 수정된 문서 파일을 **이름으로** stage (`git add -A`나 `git add .` 금지).
2. 하나의 commit 생성:

   ```bash
   git commit -m "docs: update project documentation for vX.Y.Z"
   ```

3. 현재 브랜치로 push: `git push`

**PR/MR body 업데이트 (idempotent):**

1. 기존 body를 tempfile로 읽기 (Step 0의 플랫폼 사용):
   - GitHub: `gh pr view --json body -q .body > /tmp/doc-pr-body-$$.md`
   - GitLab: `glab mr view -F json | python3 -c "import sys,json; print(json.load(sys.stdin).get('description',''))" > /tmp/doc-pr-body-$$.md`
2. Body에 `## Documentation` 섹션 있으면 교체. 아니면 append.
3. Documentation 섹션은 각 수정 파일을 Step 4의 같은 "구체적으로 뭐가 바뀌었는지" 한 줄 summary와 함께 리스트.
4. 업데이트된 body 다시 쓰기:
   - GitHub: `gh pr edit --body-file /tmp/doc-pr-body-$$.md`
   - GitLab: `glab mr update -d "$(cat /tmp/doc-pr-body-$$.md)"`
5. Cleanup: `rm -f /tmp/doc-pr-body-$$.md`
6. PR/MR 없으면 메시지와 함께 skip. Body 업데이트 실패하면 warn하되 계속 — 변경은 어쨌든 commit에 있음.

## 출력

스킬은 최종 출력으로 **구조화된 문서 health summary** 생성. 모든 관련 파일이 한 row:

```
Documentation health:
  README.md       [status] ([details])
  ARCHITECTURE.md [status] ([details])
  CONTRIBUTING.md [status] ([details])
  CHANGELOG.md    [status] ([details])
  TODOS.md        [status] ([details])
  VERSION         [status] ([details])
```

상태 값:

- **Updated** — 뭐가 바뀌었는지 설명과 함께
- **Current** — 변경 불필요
- **Voice polished** — wording 조정, factual 변경 없음
- **Not bumped** — 사용자가 skip 선택
- **Already bumped** — VERSION이 이 브랜치에서 앞서 설정됨
- **Skipped** — 이 repo에 파일 없음

Plus push된 commit SHA와 (해당 시) 업데이트된 PR/MR body의 URL.

## 중요 규칙 (recap)

- **수정 전 읽기.** 수정 전 항상 전체 파일 content 읽기.
- **CHANGELOG clobber 금지.** Wording polish만; 삭제, 교체, regenerate, `Write` 금지.
- **VERSION 조용히 bump 금지.** 명백해 보여도 항상 질문. 이미 bump됐어도 coverage 체크.
- **뭐가 바뀌었는지 명시적.** 모든 edit는 한 줄, 파일별 summary 획득.
- **프로젝트별 가정이 아니라 generic heuristic.** Audit 체크는 수정 없이 어떤 repo에서도 동작해야.
- **Discoverability 중요.** 모든 문서 파일은 프로젝트 front door (README 또는 주 instruction 파일)에서 도달 가능해야.
- **Voice: 친근, 사용자 지향, 구체.** 코드를 본 적 없는 똑똑한 사람에게 설명하듯 쓰기. refactor된 것이 아니라 사용자가 할 수 있는 것 lead.
