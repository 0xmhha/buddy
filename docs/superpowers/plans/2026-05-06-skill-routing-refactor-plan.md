# Skill Routing Refactor — 단일 Router 패턴

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 78개 skill 의 frontmatter 자동 로드(현재 ≈8K 토큰/턴) 를 단일 `router` skill 만 자동 로드되도록 통합하여 ≈99% 감축. 기존 27 개 slash command 동작은 그대로 보존하고, 다중 skill 합성용 `/buddy:run` · `/buddy:chain` · `/buddy:parallel` 추가.

## Why

- `plugin/skills/*/SKILL.md` × 78 개의 frontmatter (평균 360 자 description) 가 매 턴 컨텍스트에 상주.
- 기존 `SKILLS.md` / `SKILL_ROUTER.md` 의 "lazy-load" 의도는 SKILL.md auto-discovery 때문에 실효성 없음.
- Claude Code 의 skill discovery 는 파일명이 `SKILL.md` 인 것만 스캔 — `references/*.md` 는 자동 로드되지 않음(Buddy 자체가 이미 사용하는 패턴으로 검증됨).

## Solution Architecture

```
plugin/
├── .claude-plugin/plugin.json          # 27 commands → all "skill": "router" + 3 new
├── commands/buddy/
│   ├── <existing 26 cmd md files>      # body 만 router 호출로 전환
│   ├── run.md           (NEW)          # 단일 skill 동적 호출
│   ├── chain.md         (NEW)          # 순차 실행
│   ├── parallel.md      (NEW)          # 병렬 실행 (Agent dispatch)
│   └── status.md        (NEW)          # 누락된 status command 추가
└── skills/
    ├── router/SKILL.md                 # 유일한 자동 로드 skill
    └── <each-skill>/PROCEDURE.md       # 기존 SKILL.md → rename (×78)
```

**라우팅 흐름:**
1. 사용자가 `/buddy:<name> "args"` 입력
2. `commands/buddy/<name>.md` 가 fire → "Skill 도구로 router 호출, target=<name>, args=$ARGUMENTS" 지시
3. Router skill body (자동 로드된 frontmatter 가 트리거됨) 가 target → `plugin/skills/<target>/PROCEDURE.md` 매핑 → Read → 실행

## Out of Scope

- `plugin/agents/`, `plugin/hooks/`, `plugin/mcp/`, `plugin/rules/` 변경
- `_archive/*` skill (이미 격리됨)
- 신규 functional skill 추가 (라우팅 인프라만)
- README · CHANGELOG 외 사용자 문서 신규 작성

## Prerequisites

- Buddy repo at `/Users/wm-it-22-00661/Work/github/study/ai/buddy`, branch `main`, working tree clean
- `git mv` 사용 가능 (history 보존)
- bats-core (기존 테스트 인프라가 있다면 회귀 확인용)

---

## Phase 1: Smoke Verification (1 skill)

검증을 1 개 skill 에 먼저 적용해 패턴이 실제로 작동하는지 확인. 실패 시 mass migration 회피.

### Task 1.1: Pilot 1 개 skill 마이그레이션

**Files:**
- Modify: `plugin/skills/consult-codex/SKILL.md` → rename to `PROCEDURE.md`
- Create: `plugin/skills/router/SKILL.md` (skeletal, single-target only)
- Modify: `plugin/commands/buddy/consult-codex.md` (router 호출로 전환)

**Steps:**
- [ ] **Step 1**: `plugin/skills/router/SKILL.md` 생성 — frontmatter 약 80 자 description, body 에 단일 skill 호출 로직만:
    - 입력: target skill 이름 + user args
    - 동작: `Read plugin/skills/<target>/PROCEDURE.md` → 본문 절차에 따라 실행
    - 인덱스 섹션은 이 단계에서 비워둠 (Task 4.1 에서 채움)
- [ ] **Step 2**: `git mv plugin/skills/consult-codex/SKILL.md plugin/skills/consult-codex/PROCEDURE.md`
- [ ] **Step 3**: `plugin/commands/buddy/consult-codex.md` body 의 "Skill 도구로 `consult-codex` skill 호출" 부분을 "Skill 도구로 `router` skill 호출. target: `consult-codex`, args: `$ARGUMENTS`" 로 변경
- [ ] **Step 4**: 다음을 수동 검증:
    - `find plugin/skills/consult-codex -name SKILL.md` → 결과 없음
    - `/buddy:consult-codex` 호출 가능한지 (실제 실행은 별도 — 호출 path 만 확인)
- [ ] **Step 5**: 작동 확인 후 commit (메시지: `feat(plugin): pilot router pattern with consult-codex skill`)

**Acceptance:**
- consult-codex 디렉토리에 SKILL.md 없고 PROCEDURE.md 만 존재
- router skill 이 자동 로드되는 유일한 skill 후보로 추가됨
- `/buddy:consult-codex` command md 가 router 호출로 전환됨

**Decision gate:** 이 단계에서 router skill 이 기대대로 fire 되지 않으면 **STOP** — Plan 재검토. (예상되는 회피 시나리오: command md 가 그대로 작동하지 않으면 plugin.json 도 함께 수정해야 함을 발견)

---

## Phase 2: Router Skill 본격 구현

### Task 2.1: Router skill body 완성

**Files:**
- Modify: `plugin/skills/router/SKILL.md`

**Steps:**
- [ ] **Step 1**: Body 에 다음 섹션 추가:
    - `## How to dispatch` — target 기반 단일 호출 절차 (Read PROCEDURE.md → 실행)
    - `## Chain mode` — 콤마 분리 target 리스트 → 순차 실행, 단계별 산출물을 다음 단계 입력으로
    - `## Parallel mode` — 콤마 분리 target 리스트 → Agent tool 로 sub-agent 디스패치, 각자 PROCEDURE 읽고 실행, 결과 집계
    - `## Skill index` — 78 개 PROCEDURE 의 이름 + 1 줄 용도 (SKILLS.md 내용을 옮기되 압축)
- [ ] **Step 2**: Body 길이 검증 — 500 줄 이내 유지 (skill-creator 권장)
- [ ] **Step 3**: 인덱스 섹션이 너무 길면 `references/skill-index.md` 로 분리하고 SKILL.md 본문에는 링크만

**Acceptance:**
- router/SKILL.md 자체가 78 개 procedure 를 매핑·dispatch 가능
- 기존 SKILLS.md / SKILL_ROUTER.md 의 라우팅 정보가 router/SKILL.md 또는 references 로 이전됨

---

## Phase 3: Mass Migration

### Task 3.1: 77 개 SKILL.md 일괄 rename

**Files:**
- Modify: `plugin/skills/<each>/SKILL.md` × 77 (consult-codex 제외, router 제외, _archive 제외)

**Steps:**
- [ ] **Step 1**: 다음 셸로 대상 목록 생성:
    ```bash
    find plugin/skills -mindepth 2 -name SKILL.md \
      | grep -v "/router/" \
      | grep -v "/_archive/" \
      > /tmp/skill-migration-list.txt
    wc -l /tmp/skill-migration-list.txt   # 기대값: 77
    ```
- [ ] **Step 2**: 일괄 git mv:
    ```bash
    while read f; do
      git mv "$f" "${f%/SKILL.md}/PROCEDURE.md"
    done < /tmp/skill-migration-list.txt
    ```
- [ ] **Step 3**: 검증:
    - `find plugin/skills -name SKILL.md | wc -l` → 1 (router 만)
    - `find plugin/skills -name PROCEDURE.md | wc -l` → 78 (77 신규 + Phase 1 의 consult-codex)
- [ ] **Step 4**: commit (`refactor(plugin): rename 77 SKILL.md to PROCEDURE.md to deactivate auto-discovery`)

**Acceptance:**
- repo 전체에 SKILL.md 가 router 1 개만 남음
- git history 가 rename 으로 보존됨

---

## Phase 4: Command Migration

### Task 4.1: 기존 26 개 command md 를 router 호출로 일괄 전환

**Files:**
- Modify: `plugin/commands/buddy/*.md` × 26 (consult-codex 제외 — Phase 1 에서 완료)

**Steps:**
- [ ] **Step 1**: 각 command md 의 "Skill 도구로 `<name>` skill 호출" 패턴을 식별. 표준 패턴 확인:
    ```bash
    grep -l "Skill.*skill을 invoke\|Skill.*skill을 호출" plugin/commands/buddy/*.md
    ```
- [ ] **Step 2**: 표준 본문 템플릿 작성 (router 호출용):
    ```
    `Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

    - target PROCEDURE: `<COMMAND_NAME>`
    - 사용자 인자:
      ```
      $ARGUMENTS
      ```
    ```
- [ ] **Step 3**: 26 개 파일에 적용. 각 파일의 `<COMMAND_NAME>` 자리에 자기 이름 치환 (셸 스크립트로 자동화 가능)
- [ ] **Step 4**: 검증:
    - 모든 command md 가 `router` 를 참조하는지: `grep -L "router" plugin/commands/buddy/*.md` → consult-codex.md 외 없어야 함 (consult-codex 는 Phase 1 에서 이미 변경됨)
    - `Skill: ` (구 패턴) 잔존 여부 확인
- [ ] **Step 5**: commit (`refactor(commands): route all 27 buddy commands through router skill`)

**Acceptance:**
- 26 개 command md 가 router 호출 본문으로 전환됨
- frontmatter (description, argument-hint) 변경 없음 — 사용자 노출 동작 유지

### Task 4.2: 신규 command md 4 개 추가

**Files:**
- Create: `plugin/commands/buddy/run.md`
- Create: `plugin/commands/buddy/chain.md`
- Create: `plugin/commands/buddy/parallel.md`
- Create: `plugin/commands/buddy/status.md` (현재 plugin.json 에 등록되어 있으나 md 가 없음 — 누락 보완)

**Steps:**
- [ ] **Step 1**: `run.md` — 인자 첫 토큰을 target 으로 해석, 나머지를 args 로 router 에 전달
- [ ] **Step 2**: `chain.md` — `argument-hint`: `<skill,skill,... [-- args]>`, body 에서 콤마 분해 후 router 의 chain mode 호출
- [ ] **Step 3**: `parallel.md` — chain.md 와 같지만 router 의 parallel mode 호출
- [ ] **Step 4**: `status.md` — 기존 plugin.json 의 status command 와 동일 동작, target=`status` 로 router 호출
- [ ] **Step 5**: commit (`feat(commands): add run/chain/parallel/status command entrypoints`)

**Acceptance:**
- 4 개 신규 command 파일 존재, 각자 router 호출
- `/buddy:status` 가 작동 (기존 누락 보완)

### Task 4.3: plugin.json commands 배열 정렬

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`

**Steps:**
- [ ] **Step 1**: 현재 27 개 command 의 `"skill"` 필드를 일괄 `"router"` 로 변경. `"name"` 과 `"description"` 은 유지.
- [ ] **Step 2**: 신규 4 개 command (run, chain, parallel, status) 항목 추가 (이미 status 가 있다면 중복 제거)
- [ ] **Step 3**: JSON 유효성 검증: `jq . plugin/.claude-plugin/plugin.json > /dev/null`
- [ ] **Step 4**: commit (`refactor(plugin): point all commands at router skill`)

**Acceptance:**
- plugin.json 의 모든 command 가 `"skill": "router"`
- run / chain / parallel / status command 등록됨

---

## Phase 5: Cross-reference Cleanup

### Task 5.1: PROCEDURE.md 들의 내부 link 수정

**Files:**
- Modify: PROCEDURE.md 들 중 다른 skill 의 SKILL.md 를 직접 참조하는 파일

**Steps:**
- [ ] **Step 1**: 검색:
    ```bash
    grep -rln "skills/[^/]*/SKILL\.md" plugin/skills/
    ```
- [ ] **Step 2**: 각 매칭에서 `SKILL.md` → `PROCEDURE.md` 로 sed 치환 (단, router 자기 자신 참조는 제외)
- [ ] **Step 3**: 검증: 위 grep 의 잔존 매칭이 router/SKILL.md 만 가리키는지
- [ ] **Step 4**: commit (`fix(plugin): update intra-skill links to PROCEDURE.md after rename`)

**Acceptance:**
- 깨진 link 없음

### Task 5.2: SKILLS.md / SKILL_ROUTER.md 정리

**Files:**
- Modify or Delete: `plugin/SKILLS.md`, `plugin/SKILL_ROUTER.md`

**Steps:**
- [ ] **Step 1**: router/SKILL.md 가 두 파일의 정보를 흡수했는지 확인
- [ ] **Step 2**: 결정:
    - (A) 두 파일을 router 의 references 로 이동 (`plugin/skills/router/references/skill-catalog.md`, `routing-rules.md`)
    - (B) 두 파일 삭제, 모든 정보가 router 에 통합됨
- [ ] **Step 3**: 선택한 옵션대로 적용
- [ ] **Step 4**: commit (`docs(plugin): consolidate skill catalog into router skill`)

**Acceptance:**
- 라우팅 정보 source-of-truth 가 router skill 1 곳

---

## Phase 6: Documentation & Smoke Test

### Task 6.1: README · CHANGELOG 업데이트

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

**Steps:**
- [ ] **Step 1**: README 의 "skills" 섹션 갱신 — 단일 router 패턴 설명, run/chain/parallel 사용법 1 줄 예시
- [ ] **Step 2**: CHANGELOG 에 entry 추가 — breaking change 가 *없음* 명시 (사용자 노출 command 동일), 내부 구조 변경만
- [ ] **Step 3**: commit (`docs: update README and CHANGELOG for router refactor`)

**Acceptance:**
- 사용자가 README 만 봐도 새 구조 이해 가능

### Task 6.2: Token 절감 측정 + smoke test

**Files:**
- (no source change)

**Steps:**
- [ ] **Step 1**: 자동 로드 description 토큰 측정:
    ```bash
    awk '/^---$/{c++;next} c==1 && /^description:/{sub(/^description: */, ""); print}' \
      plugin/skills/router/SKILL.md | wc -c
    ```
    기대값: 200 자 미만 (≈50–60 토큰)
- [ ] **Step 2**: 5 개 representative command smoke test (수동 호출):
    - `/buddy:status`
    - `/buddy:concretize-idea "test idea"`
    - `/buddy:audit-security`
    - `/buddy:chain validate-idea, assess-business-viability "test idea"`
    - `/buddy:parallel review-engineering, review-design "this branch"`
- [ ] **Step 3**: 각 호출이 router 를 거쳐 올바른 PROCEDURE.md 를 읽어 실행하는지 확인
- [ ] **Step 4**: 결과를 `docs/superpowers/plans/2026-05-06-skill-routing-refactor-plan.md` 의 "Verification log" 섹션에 기록
- [ ] **Step 5**: 모든 검증 통과 시 최종 commit (`chore: verify router refactor smoke tests`)

**Acceptance:**
- 자동 로드 description ≈ 100–200 자 수준 (vs. 28K 자 baseline)
- 5 개 smoke test 통과
- chain / parallel 모드가 의도대로 동작

---

## Verification Log

(Phase 6.2 에서 채움)

- Baseline auto-loaded description chars: **28,125** (78 skills)
- Post-refactor auto-loaded description chars: **<TBD>**
- Reduction: **<TBD>%**
- Smoke test results: **<TBD>**

---

## Rollback Plan

각 Phase 의 commit 이 독립적이므로 phase 단위 `git revert` 로 롤백 가능. Phase 1 이 가장 먼저 검증 게이트이므로, 여기서 실패 시 그 commit 만 revert 하면 main 이 깨끗하게 복원됨.
