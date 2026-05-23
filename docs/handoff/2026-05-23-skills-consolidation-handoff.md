# Skills Consolidation Handoff — 2026-05-23

> **이 문서의 목적**: 다른 머신 / 다른 세션에서 *컨텍스트 0 으로 시작* 하여 이 작업을 이어받을 수 있도록 self-contained 핸드오프. 이 문서만 읽으면 다음 작업이 시작 가능해야 함.

---

## 0. TL;DR — 다음 세션이 즉시 해야 할 일

1. 이 문서 전체 읽기 (10-15 분)
2. `git pull` (buddy repo) — 최신 commit `54efc0f` 까지 확인
3. SKILLS_ANALYSIS.md `/Users/wm-it-22-00661/Work/github/study/ai/skill/SKILLS_ANALYSIS.md` § A.1 잔여 항목 확인
4. 사용자에게 "어느 항목부터 진행할까요" 확인 또는 LOW-1 부터 시작
5. **§9 작업 진행 패턴 (4-블록 설명 → 승인 → 실행)** 반드시 준수

---

## 1. 프로젝트 개요

### 1.1 위치

| 디렉토리 | 역할 | git tracked |
|----------|------|-------------|
| `/Users/wm-it-22-00661/Work/github/study/ai/skill/` | 17개 source repo + 분석 자료 | NO (분석 작업용) |
| `/Users/wm-it-22-00661/Work/github/study/ai/buddy/` | 통합 대상 plugin + Go CLI | YES (origin: 사용자 GitHub) |

### 1.2 목적

17개 source repo (총 167 skills) 의 best parts 를 `buddy` 프로젝트에 통합. 한 번의 plugin install 로 Claude Code 워크플로우 전체 개선.

핵심 제약:
- **contradictions 방지** — 같은 시점에 여러 스킬이 모순된 조언 → 토큰 낭비 + worse output
- **duplications 방지** — 동일 책임 스킬 중복 → 라우팅 혼란

사용자 명시 목표:
> "한번 잘 셋업하면, 그 뒤로는 잘사용할수 있기 때문이야"
> = 셋업 비용을 한 번 치르고, 이후 매 세션에서 reliable 하게 사용

### 1.3 buddy 아키텍처 (필수 이해)

- **9-phase 라이프사이클 orchestrator** (Priority 1):
  - §1 `concretize-idea` → §2 `define-features` → §3 `design-system` → §4 `plan-build` → §5 `build-feature` → §6 `verify-quality` → §7 `ship-release` → §8 `iterate-product` → §9 `manage-lifecycle`
- **Cross-phase sub-orchestrator** (Priority 2): `autoplan`, `finish-development-branch` (MID-4 신규)
- **Stage skill** (Priority 3): 각 phase 안의 specific skill (예: `decompose-feature-to-actor-tracks`)
- **Domain skill** (Priority 4): cross-phase 호출 가능 (예: `write-adr`, `decompose-blocker`)
- **Pattern library** (Priority 5): dispatch-only reference 구현 (예: `monitor-regressions`, `benchmark-llm-models`)
- **Cross-skill reference** (router/references/): 모든 스킬이 lazy-load 가능한 SSoT (예: `skill-catalog.md`, `routing-rules.md`, `verification-discipline.md`, `git-safety-rules.md`)
- **단일 SKILL.md** (router/SKILL.md) + 모든 다른 스킬은 PROCEDURE.md — auto-discovery 차단을 위함
- **Command 파일**: `plugin/commands/<name>.md` — `/buddy:<name>` 호출 시 router 통해 dispatch

검증: `make test-routing` 10/10 통과 필수. PROCEDURE.md 개수 변경 시 `scripts/test-router-wireup.sh` 의 count 도 lockstep 갱신.

---

## 2. 사용자 명시 규칙 (반드시 준수)

### 2.1 사용자 CLAUDE.md / rules 발췌 (전체는 `~/.claude/CLAUDE.md` 참조)

| 규칙 | 출처 | 적용 |
|------|------|------|
| **Git commit 메시지에 `Co-Authored-By` 또는 "Generated with [Claude Code]" attribution 금지** | 사용자 명시 선호 | conventional commit body 만 영어로 작성, attribution 라인 추가 금지 |
| **사용자가 한국어로 쓰면 한국어로 응답** | 사용자 명시 | 한국어 prose + 영어 technical identifiers + 영어 commit 메시지 |
| **uncommitted 변경 종료 시 사용자에게 commit 여부 확인** | 사용자 명시 | 자율 commit / 폐기 금지 |
| **Read 전 Edit/Write 금지** | Core Principles | edit/write 전 무조건 read |
| **불변성 우선, fail fast / fail explicitly** | Core Principles | 에러 silent 삼킴 금지 |
| **3회 연속 동일 오류 시 사용자에게 보고 후 지시 대기** | Execution Protocol | 무한 재시도 금지 |

### 2.2 작업 중 학습된 추가 규칙 (이번 세션)

| 규칙 | 발견 시점 | 이유 |
|------|----------|------|
| **자동화 시스템 git force-류 절대 금지** | MID-4 진행 중 사용자 강한 push-back | `--force-with-lease` 도 race window 내 silent 손실 가능. AI 자동화에선 reflog 의존 복구 사실상 불가능 |
| **"phase" vs "Step" 용어 분리** | write-a-skill 작성 시 | "phase" = buddy 9-phase 라이프사이클. 스킬 내부 절차는 "Step" |
| **ADR-003 §2.4 verbatim 0건 유지** | 초기 정책 | 모든 흡수는 inspired-by / adopt-with-edits / reference-only — 원문 그대로 복사 금지 |
| **각 항목 진행 전 4-블록 설명 + 승인 패턴** | A 잔여 작업 시작 시 사용자 요청 | 무엇을 / 왜 / 미설정 시 문제 / 흡수 분류 |
| **race condition 대응 시 사용자 컨텍스트 우선** | A4 / MID-1 / MID-3 / MID-4 commit | 다른 세션의 commit 에 내 파일이 번들되어도 *사용자가 알고 있는 작업이면* 분할 비용 회피 |
| **simulation test 기반 iterative** | MID-1 / MID-2 / MID-4 진행 시 | draft → 위치 검토 → 시뮬레이션 2 시나리오 → 수정 → final |

---

## 3. 완료 작업 (commit chronology)

### 3.1 사전 작업 (이전 세션)

- `write-a-skill` 메타 스킬 신설 (5-라운드 self-validation)
- `explore-design-variants` → `verify-best-alternative` 리네임 + 엔지니어링 한정 scope-narrow
- 7 design 스킬에 anti-bias gate (3+ orthogonal 후보 강제) 추가
- `decompose-blocker` 신설 (language-independent, 3-attempt auto-trigger)
- ADR-018 (rename + forced gate + decompose-blocker)

### 3.2 이번 세션 commits (시간순)

| Commit | 항목 | 변경 요약 |
|--------|------|---------|
| `df4b64a` | A1-A3 | verify-best-alternative engineering-first rewrite + 7 design 스킬 anti-bias gate + decompose-blocker 3회 한도 |
| `830eb04` | **HIGH** review-engineering | Eng Manager Persona 직후 신규 섹션 5 anti-rationalization 규칙 (금지 발화 / no-issues 증명 / 7/7 STOP / 작성자 verify / per-dimension challenge 자증명). Output 7번째 산출물 (self-attestation MET/NOT MET) |
| `1523e65` | **MID-1** verification-discipline (병행 작업과 통합) | `router/references/verification-discipline.md` 신규 (182줄, 10 섹션 SSoT). 4 skill cross-link (build-with-tdd / iterate-fix-verify / verify-quality / build-feature). skill-catalog §5 참조 등록 |
| `4754e84` | **MID-2** loop-methods | `diagnose-bug/references/loop-methods.md` 신규 (234줄, 8 섹션: disproportionate-effort + 10 methods + iterate-on-loop + non-deterministic + cannot-build-a-loop STOP + cascade + anti-pattern). Phase 1 제목 "Reproduce → Build a feedback loop" 격상, Phase 1/1.5 cross-link 4곳 |
| `5c9fba5` | **MID-3** publish-to-tracker | 신규 스킬 (280줄, 2 모드 `prd`/`issues`, multi-tracker abstraction). command 파일. define-features + plan-build cross-link. catalog 등록. test count 149→150 |
| `54efc0f` | **MID-4** finish-development-branch + git-safety-rules | 신규 sub-orchestrator (275줄, 5 stage safe-only). 신규 SSoT `git-safety-rules.md` (188줄, force-류 자동 금지). guard-destructive-commands 의 force-with-lease 안전 분류 제거. 5 cross-link. test count 150→151 |

### 3.3 SKILLS_ANALYSIS.md 진행 상황

위치: `/Users/wm-it-22-00661/Work/github/study/ai/skill/SKILLS_ANALYSIS.md` (non-git, 분석 마스터 문서)

§3 A.1 항목 상태:
- A1 ✅ 완료 (df4b64a)
- A2 ✅ 완료 (df4b64a)
- A3 ✅ 완료 (df4b64a)
- A4 ⏭️ 건너뛰기 결정 (실용 영향 0)
- HIGH ✅ 완료 (830eb04)
- MID-1 ✅ 완료 (1523e65)
- MID-2 ✅ 완료 (4754e84)
- MID-3 ✅ 완료 (5c9fba5)
- MID-4 ✅ 완료 (54efc0f)
- **LOW 1-3 ⬜ 미진행** (다음 세션 대상)
- **신규 follow-up: parallel agent 충돌 방지 ⬜ 미진행**

§3 B-I 카테고리: 모두 unchecked, 아직 시작 안 함.

---

## 4. A-카테고리 anti-rationalization 4-layer 시스템 (완성됨)

이 시스템은 *AI 가 sycophancy / 자기 합리화로 빠지는 경로를 차단* 하기 위한 4 layer 게이트. 새 스킬 작성 / 기존 스킬 보강 시 *어느 layer 에 속하는지* 의식 필요.

| Layer | 시점 | 게이트 | 위치 |
|-------|------|--------|------|
| **결정 (Decision)** | §3 design 단계의 결정 commit 직전 | 3+ orthogonal 후보 강제 + rubric 비교 | 7 design 스킬 본문 + `docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md` (A2) |
| **리뷰 (Review)** | plan/spec 리뷰 시점 | 5 anti-rationalization 규칙 (금지 발화 / no-issues 증명 / 7/7 STOP / 작성자 verify / per-dim challenge) + Output self-attestation | `review-engineering/PROCEDURE.md` (HIGH) |
| **완료 발화 (Claim)** | "test pass" / "GREEN" / "fix verified" 등 발화 직전 | Iron Law (no fresh evidence = no claim) + Gate Function 5-step | `router/references/verification-discipline.md` (MID-1) |
| **git 실행 (Action)** | git 명령 자동 실행 직전 | force-류 금지 + 안전 대체 + STOP 절차 | `router/references/git-safety-rules.md` + runtime hook `guard-destructive-commands` (MID-4) |

각 layer 의 누락은 *다른 layer 의 보호 우회 경로*. 모두 활성 시에만 시스템 완전.

---

## 5. 핵심 파일 빠른 참조

### 5.1 분석 마스터

- `/Users/wm-it-22-00661/Work/github/study/ai/skill/SKILLS_ANALYSIS.md` — 전체 분석 + 진행 상태 (non-git)

### 5.2 Cross-skill SSoT (router/references/)

| 파일 | 역할 | 신설 |
|------|------|------|
| `skill-catalog.md` | 전체 스킬 카탈로그 + 트리거 | 기존 |
| `routing-rules.md` | 라우팅 충돌 결정 | 기존 |
| `verification-discipline.md` | 완료 발화 Iron Law SSoT | MID-1 신설 |
| `git-safety-rules.md` | git 안전 원칙 SSoT | MID-4 신설 |

### 5.3 ADR / Spec

- `docs/superpowers/decisions/` — ADR 디렉토리
  - ADR-003: 4분류 정책 (verbatim / adopt-with-edits / reference-only / inspired-by)
  - ADR-018: explore-design-variants → verify-best-alternative 리네임 + forced gate + decompose-blocker (2026-05-21)
- `docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md` — A2 게이트 매핑 spec
- `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` — 9-phase 아키텍처 spec

### 5.4 신설된 스킬 / reference (이번 세션 + 이전 세션)

```
plugin/skills/
├── write-a-skill/                        (이전 세션, 신규)
│   ├── PROCEDURE.md
│   └── references/
│       ├── subagent-pressure-test.md
│       ├── attribution-classification.md
│       └── yaml-output-spec.md
├── verify-best-alternative/              (이전 세션, rename + rewrite)
├── decompose-blocker/                    (이전 세션, 신규)
├── publish-to-tracker/                   (MID-3 신설)
│   └── PROCEDURE.md
├── finish-development-branch/            (MID-4 신설)
│   └── PROCEDURE.md
├── diagnose-bug/
│   ├── PROCEDURE.md                      (MID-2 Phase 1/1.5 보강)
│   └── references/
│       └── loop-methods.md               (MID-2 신설)
├── router/
│   └── references/
│       ├── verification-discipline.md    (MID-1 신설)
│       └── git-safety-rules.md           (MID-4 신설)
└── (기존 약 145개 스킬 — skill-catalog.md 참조)

plugin/commands/
├── publish-to-tracker.md                  (MID-3 신설)
├── finish-development-branch.md           (MID-4 신설)
└── (기존 약 101개 command)
```

### 5.5 검증

```bash
cd /Users/wm-it-22-00661/Work/github/study/ai/buddy
make test-routing                          # 10/10 통과 필수
```

PROCEDURE 개수: 현재 **151** (+ 1 router SKILL.md = 152 skills total). `scripts/test-router-wireup.sh` 의 threshold 가 항상 lockstep.

---

## 6. 남은 작업 리스트

### 6.1 A 카테고리 잔여 3 영역 + 1 follow-up

#### LOW 1: mattpocock 자동화 검토 (setup-pre-commit, git-guardrails)

**위치**: `/Users/wm-it-22-00661/Work/github/study/ai/skill/mattpocock-skill/skills/engineering/` (또는 `productivity/`)
**현재 상태**: buddy 의 `compose-safety-mode` + `guard-destructive-commands` 와 비교 필요
**기대 변경**: pre-commit hook / git guardrail 패턴 흡수 — git-safety-rules.md (MID-4) 와 자연 연결 (MID-4 가 design-time policy, LOW-1 이 runtime automation 보완 가능)
**난이도**: 중간 (기존 buddy 스킬과 정합 검토 필요)

#### LOW 2: ubiquitous-language 검토

**위치**: `/Users/wm-it-22-00661/Work/github/study/ai/skill/mattpocock-skill/skills/engineering/ubiquitous-language/SKILL.md`
**현재 상태**: buddy 의 actor 모델 (define-features 의 actor identification) 와 보완 가능
**기대 변경**: DDD 도메인 어휘 일관성 검사 패턴 흡수 — 신규 스킬 또는 define-features 보강
**난이도**: 낮음 (단일 스킬 추가)

#### LOW 3: caveman 검토

**위치**: `/Users/wm-it-22-00661/Work/github/study/ai/skill/mattpocock-skill/skills/productivity/caveman/SKILL.md`
**현재 상태**: 75% 토큰 압축 패턴 — buddy 의 cli-buddy token monitor 와 시너지
**기대 변경**: 신규 스킬 또는 router 본문에 lazy-load 가이드 추가
**난이도**: 낮음

#### 신규 Follow-up: parallel agent 충돌 방지 (MID-4 작업 중 사용자 제기)

**배경**: MID-4 작업 중 사용자가 *sub-agent 병렬 작업 시 충돌* 시나리오 제기. finish-development-branch 의 Stage 0 pre-flight sync 는 *PR 직전 시점만* cover. plan 단계의 *근본 보강* 별도 필요.

**4 skill 보강 작업**:
1. `decompose-track-to-tasks/PROCEDURE.md` — task 산출 yaml 에 `expected_touched_files: [path1, path2]` 필드 추가
2. `map-task-dependencies/PROCEDURE.md` — file-overlap edge 자동 추가 (두 task 가 같은 파일 → sequential edge)
3. `plan-parallel-execution/PROCEDURE.md` — same-wave 내 file-overlap=0 검증 + *high-conflict zone* (schema / migration / shared config) single-track only 강제
4. `dispatch-parallel-agents/PROCEDURE.md` — worktree base periodic refresh (긴 작업 중 base drift 대비)

**난이도**: 중-상 (4 skill 동시 보강 + 정합성)

### 6.2 B-I 카테고리 (모두 시작 안 함)

SKILLS_ANALYSIS.md §3 의 B-I 항목 unchecked:
- **B**: designer-skills 73개 — 전부 유지 vs 핵심 ~30개 압축?
- **C**: 테마/브랜드 5종 — 어느 하나 *프라이머리* 결정? (ui-ux-pro-max 권장)
- **D**: 마케팅 SEO 4종 — 통합 vs 유지?
- **E**: obsidian — 개인 vault 와 일반 skills 둘 다 유지?
- **F-I**: SKILLS_ANALYSIS.md 직접 확인

---

## 7. 진행 시 작업 패턴 (반드시 준수)

### 7.1 4-블록 설명 + 승인 패턴

각 항목 진행 전 사용자에게 다음 4 블록 제시:

1. **무엇을 수정?** — 신규 파일 / 보강 위치 / cross-link 위치를 정확히 명시
2. **배경 / 이유** — 왜 이 변경이 필요한지. buddy 현 상태 + 통합 source 의 통찰
3. **미설정 시 문제** — 5-7개 구체 시나리오. 무방비 시 어떤 sycophancy / 합리화 / 누락 발생
4. **흡수 분류 (ADR-003)** — verbatim 0건 / adopt-with-edits / inspired-by / reference-only 중 어느 것

승인 받은 후 실행. *모호한 답 (e.g. "그냥 해")* 에는 *재확인* 요청.

### 7.2 작업 단계 (반복 가능 cycle)

1. **TaskCreate** (3-5 tasks)
2. **자료 조사** — Read 위주, 필요 시 Grep / Glob (대체로 Bash 보다 dedicated tool)
3. **Draft** — Write tool 로 신규 파일 또는 Edit tool 로 기존 보강
4. **Simulation test 2 시나리오** — "이 변경을 적용한 상태에서 X 상황이 오면 어떻게 동작?" 으로 *내부 추적*. 발견 점 수정
5. **`make test-routing` 통과** — 10/10 필수. PROCEDURE 추가 시 `scripts/test-router-wireup.sh` count 갱신
6. **Lock 체크** — `ls -la .git/index.lock` + `ps -ef | grep git` 로 활성 buddy git 프로세스 확인. stale lock 만 제거 (live 면 대기)
7. **단독 stage** — `git add <specific-files>` (절대 `git add .` 금지)
8. **Conventional commit** — 영어 detailed body, scope 명시, ADR/spec 참조 인용. Co-Authored-By 절대 금지
9. **SKILLS_ANALYSIS.md 업데이트** — 해당 항목 strikethrough + commit hash + 변경 요약 1-2 줄 (한국어)
10. **TaskList 정리** — 완료 상태로 마킹

### 7.3 Race condition 대응

**상황**: 다른 세션이 buddy 에서 동시 작업 중 (특히 daemon 자동 commit 가능성). git index.lock 자주 발생.

**확인**:
```bash
ls -la /Users/wm-it-22-00661/Work/github/study/ai/buddy/.git/index.lock
ps -ef | grep "git" | grep -v grep | grep -i "study/ai/buddy"
```

**처리**:
- 활성 buddy git 프로세스 0건이면 stale lock → `rm -f .git/index.lock`
- 활성 git 프로세스 있으면 대기

**Commit 후 다른 세션이 내 파일을 자기 commit 에 번들한 경우**:
- 이번 세션에서 사용자는 분할보다 **현 상태 유지** 선호 (사용자 컨텍스트로 이미 알고 있는 작업)
- SKILLS_ANALYSIS.md 에 commit hash 메모 (e.g. "병행 작업과 통합")
- audit 시 file path 로 검색 가능

### 7.4 Force git 금지 (사용자 강한 push-back, MID-4 학습)

자동화 시스템에서 **절대 금지**:
- `git push --force` / `-f`
- `git push --force-with-lease[=...]`
- `git rebase origin/<base>` (push 된 branch 에서)
- `git rebase -i HEAD~N` (push 된 commit 포함 시)
- `git commit --amend` (push 된 commit)
- `git reset --hard`
- `git checkout .` / `git restore .`
- `git clean -fd` / `-fdx`
- `git branch -D` (force delete)
- `git filter-branch` / `filter-repo`
- `git gc --prune=now`

이유: `--force-with-lease` 도 race window 내 silent 손실 가능. AI 자동화에선 reflog 의존 복구 사실상 불가능.

전체 정책은 `plugin/skills/router/references/git-safety-rules.md` (§2 표) 참조.

사용자 명시 승인 시에만 *수동 실행 안내* (sub-orchestrator 가 직접 실행 금지).

### 7.5 ADR-003 4분류 정책

| 분류 | 의미 | 예시 |
|------|------|------|
| **verbatim** | 원문 그대로 복사 | **0건 유지 — 절대 금지** |
| **adopt-with-edits** | 형식 채택 + buddy 도메인 어휘 재진술 | PRD template (publish-to-tracker, MID-3) |
| **inspired-by** | 개념 차용, 자체 발명 포함 | Iron Law (verification-discipline, MID-1) / 10 methods (loop-methods, MID-2) |
| **reference-only** | 출처 명시만, 본문 자체 발명 | finish-development-branch (MID-4, buddy 자체 발명) |

매 commit 메시지 + 신규 스킬 본문 §11 참조 섹션에 분류 명시.

### 7.6 Phase / Step 용어 분리

- **"phase"** = buddy 의 9-phase 라이프사이클 단계 (concretize-idea / define-features / ... / manage-lifecycle)
- 스킬 내부 절차는 **"Step"** 또는 **"Stage"** 사용 (write-a-skill 패턴)
- diagnose-bug 의 Phase 1-9 는 *이미 작성된 용어* 유지 (변경 비용 큼 — 예외)

---

## 8. Source repo 목록 (17개)

`/Users/wm-it-22-00661/Work/github/study/ai/skill/` 안:

| 디렉토리 | 핵심 | 흡수 진행 |
|----------|------|---------|
| superpowers/ | brainstorming, writing-skills, receiving-code-review, verification-before-completion 등 | A1-A3, HIGH, MID-1 완료 |
| mattpocock-skill/ | diagnose, to-prd, to-issues, ubiquitous-language, caveman, setup-pre-commit, git-guardrails 등 | MID-2, MID-3 완료 / LOW 1-3 미진행 |
| designer-skills/ | UI/UX 디자인 70+ 스킬 | B 카테고리 미진행 |
| vercel-agent-skills/ | react-best-practices 등 | 미진행 |
| (기타 14개 — SKILLS_ANALYSIS.md §0 참조) | | 대부분 미진행 |

---

## 9. 다음 세션 시작 시 권장 절차

### 9.1 즉시 (5 분)

1. `git pull` (buddy repo, branch: `main`)
2. `git log --oneline | head -10` — 최신 commit `54efc0f` 확인
3. `cd /Users/wm-it-22-00661/Work/github/study/ai/buddy && make test-routing` — 10/10 통과 확인
4. 이 문서 (`docs/handoff/2026-05-23-skills-consolidation-handoff.md`) + SKILLS_ANALYSIS.md (non-git 위치 — 다른 머신에 sync 필요할 수 있음) 읽기

### 9.2 사용자 의도 확인 (5 분)

사용자에게 다음 중 선택 요청:
- **LOW 1** (mattpocock 자동화 — git-safety-rules 와 자연 연결, *권장 시작점*)
- **LOW 2** (ubiquitous-language)
- **LOW 3** (caveman)
- **신규 follow-up** (parallel agent 충돌 방지 — 4 skill 보강, 중-상 난이도)
- **B-I 카테고리 진입** (designer-skills / 테마 / 마케팅 등)
- 다른 우선순위

### 9.3 항목 진행 시 (수 시간)

§7.1 의 4-블록 설명 + 승인 패턴 따라 진행. §7.2 의 10-단계 cycle 반복.

---

## 10. 알려진 위험 / 함정

### 10.1 SKILLS_ANALYSIS.md 가 non-git

`/Users/wm-it-22-00661/Work/github/study/ai/skill/SKILLS_ANALYSIS.md` 는 git tracked 아님. 다른 머신에서 작업하려면:
- 해당 파일을 별도 sync 메커니즘 (iCloud / Dropbox / scp / rsync) 필요
- 또는 *이 핸드오프 문서의 §3.3 진행 상태* 만 보고 작업 후, 완료 시 새 머신에서 같은 파일 갱신
- 또는 SKILLS_ANALYSIS.md 자체를 git tracked 위치로 이동 (사용자 결정 사항)

### 10.2 buddy daemon background 실행

`bin/buddy daemon` 이 항상 background 실행 (5:32AM 부터). git index.lock 의 빈번한 등장 원인. 다만 daemon 자체는 git 명령 실행 안 함 — stale lock 안전 제거 가능.

### 10.3 다른 세션 commit 과의 attribution race

이번 세션에서 3회 발생 (1523e65 / 5c9fba5 직후 / 54efc0f 직후). 패턴:
- 내가 stage 한 직후 다른 세션이 commit → 내 파일이 다른 commit 에 번들
- 또는 내 commit 직후 다른 commit 이 위에 쌓임 (이건 정상)
- 사용자 선호: **분할 비용 회피, 현 상태 유지**. SKILLS_ANALYSIS.md 에 메모만 추가.

### 10.4 자동 force git 위험

§7.4 절대 금지. 사용자가 *명시적 의도 + 결과 이해* 표명 시에만 *수동 안내*. "그냥 해" / "알아서 해" 같은 모호한 답은 승인 아님 — 재확인.

### 10.5 PROCEDURE.md count drift

신규 스킬 추가 시 `scripts/test-router-wireup.sh` 의 count 갱신 lockstep 필수. 현재 **151**. `make test-routing` 실패 시 가장 흔한 원인.

---

## 11. 통신 / 응답 스타일

### 11.1 한국어 응답

사용자 한국어 → 한국어 응답. 영어 technical identifier (commit hash / skill name / file path) 는 영어 그대로.

### 11.2 Fact-based output 형식 (사용자 CLAUDE.md 명시)

```
Fact: 확신도 "None" 인 항목 (= 추론 아닌 사실) 만 기재
Opinion:
  - High prediction
  - Mid prediction
  - Low prediction
  - None
```

이번 세션에서는 *간결한 표 + bullet list + 4-블록 설명* 으로 응답. 매 응답마다 Fact/Opinion 라벨 의무는 아님 (사용자가 지적 없으면).

### 11.3 Commit message 형식

```
<scope>(<area>): <subject>

<empty line>

<body — 여러 단락, 영어, why 위주, file path / line number 인용 가능>

<empty line>

Refs: <SKILLS_ANALYSIS 항목 + sibling commit hash + spec 경로>
```

Co-Authored-By 절대 금지. "Generated with [Claude Code]" 금지.

---

## 12. 부록 — Commit body 예시 (5c9fba5 publish-to-tracker)

상세 conventional commit body 예시는 `git show 5c9fba5` 로 확인 가능. 구조:
- 첫 단락: *문제 정의* (buddy 의 어느 gap)
- 둘째 단락: *해결책 개요*
- 셋째-N 단락: *구체 변경* (file / line / 의도)
- 마지막 단락: *Refs* (SKILLS_ANALYSIS 항목 / sibling commit / spec)

다음 세션도 같은 패턴 유지 권장. 한국어 응답이지만 commit body 는 영어 (국제 협업 / 외부 reviewer 가독성).

---

**문서 끝.** 다음 세션 시작 시 §9.1 부터 진행.

질문 / 모호한 부분 발견 시 사용자에게 즉시 확인. 추측 진행 금지 (§2.1 의 "3회 연속 동일 오류 시 보고" 원칙 적용).
