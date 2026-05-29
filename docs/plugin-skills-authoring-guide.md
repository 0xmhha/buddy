# Plugin Skill 작성 가이드 — Authoring & Improvement Standard

> **목적**: buddy 프로젝트의 ~150 skill 작성·평가·개선 작업의 단일 기준 문서(SSoT).
> 본 문서는 Anthropic 공식 docs와 prompt engineering 가이드를 기반으로, buddy의 PROCEDURE.md 형식에 맞춰 정리한 작성 표준이다.
>
> **사용 시점**:
> - 신규 skill 작성 시 — 본 가이드의 권장 구조를 그대로 따라 작성
> - 기존 skill 개선 시 — §6의 평가 체크리스트로 점검 후 부족한 항목 보강
> - skill quality 평가 — §6의 체크리스트가 정량 평가 지표
>
> **출처**:
> - https://code.claude.com/docs/en/skills.md (Claude Code Skills 공식)
> - https://code.claude.com/docs/en/plugins.md (Plugin 시스템)
> - https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering (Prompt Engineering)
> - Anthropic Cookbook의 system prompt / role-playing 예제
>
> **상위 관계 문서**:
> - [`plugin/skills/router/references/engineering-phases.md`](../plugin/skills/router/references/engineering-phases.md) — Phase 정의 + I/O Contract 표준
> - [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) — 전체 skill 카탈로그

---

## §0. buddy의 skill 구조 특이점

본 가이드를 읽기 전에 알아야 할 사항.

### 0.1 PROCEDURE.md vs SKILL.md

공식 Claude Code는 `SKILL.md` 파일명을 사용하지만, buddy는 의도적으로 **`PROCEDURE.md`**를 사용한다 (router 1개만 auto-discoverable SKILL.md, 나머지 ~154개는 PROCEDURE.md).

| 항목 | 공식 | buddy |
|------|------|-------|
| 파일명 | `SKILL.md` | `PROCEDURE.md` |
| Frontmatter | 필수 | **PROCEDURE.md에는 없음** (router 경유 lazy-load) |
| Description-based dispatch | frontmatter | `skill-catalog.md` 중앙 |
| 토큰 비용 | 모든 skill description 상시 로드 | router 1개만 (99.44% 절감) |

### 0.2 본 가이드 적용 범위

| 영역 | 적용 대상 |
|------|---------|
| **§1 Frontmatter** | `router/SKILL.md`, `plugin/commands/*.md` (PROCEDURE.md는 해당 X) |
| **§2 본문 구조** | 모든 PROCEDURE.md |
| **§3 Persona** | 모든 PROCEDURE.md (선택적이지만 권장) |

PROCEDURE.md는 frontmatter가 없지만 본문 구조와 persona는 핵심 적용 대상이다.

---

## §1. Skill 호출 시점 결정 — Frontmatter

> 본 섹션은 `router/SKILL.md`, `plugin/commands/*.md` 같이 frontmatter를 사용하는 파일에 적용. PROCEDURE.md는 frontmatter가 없으므로 §2부터.

### 1.1 Frontmatter의 역할

YAML frontmatter는 Claude가 skill을 "**언제 호출할 것인가**"를 결정하는 메타데이터다. Claude는 skill 본문을 보지 않고 frontmatter만 보고 자동 호출 여부를 판단한다 (description-based dispatch). 따라서 frontmatter 품질이 skill의 발견성을 좌우한다.

### 1.2 핵심 키 상세

#### 1.2.1 `description` (가장 중요)

**역할**: Claude가 사용자 발화/맥락을 분석하여 이 skill을 호출할지 결정하는 기준.

**규칙**:
- 최대 1,536자 (cap)
- 첫 문장에 **"use when"** 키워드 배치 (가장 강한 trigger signal)
- 사용자가 표현할 법한 자연어 패턴을 포함
- 추상적 표현보다 구체적 키워드

**Good 예시**:
```yaml
description: |
  Use when user wants to review code changes for security vulnerabilities,
  OWASP Top 10 patterns, or pre-commit security audit. Trigger keywords:
  "security review", "OWASP check", "audit security", "보안 검토".
  Performs CSO-mode audit with severity classification and remediation steps.
```

**Bad 예시**:
```yaml
description: |
  이 스킬은 다양한 보안 검토를 수행하는 종합적인 도구로서, 사용자의 코드를 분석하여
  잠재적인 취약점을 발견하고 적절한 조치 방안을 제시하는 역할을 합니다.
  (왜 bad: "use when" 부재, 추상적 narrative, 키워드 후순위)
```

**작성 체크**:
- [ ] 첫 문장에 "Use when" 또는 "사용 시점" 명시
- [ ] 사용자의 자연어 발화 패턴 3-5개 포함
- [ ] 추상적 형용사("종합적", "효과적") 제거
- [ ] 1,536자 이내 (실측 권장 500-800자)

#### 1.2.2 `disable-model-invocation`

**역할**: `true`로 설정 시 Claude가 자동 호출 불가, 사용자만 수동 호출 가능.

**언제 `true`로 설정해야 하는가**:
- **Side-effect 작업** (deploy, commit, push, DB migration, 외부 API 호출)
- **비가역적 작업** (파일 삭제, 데이터베이스 drop)
- **비용 발생 작업** (외부 SaaS 호출, paid API)
- **민감 작업** (인증 정보 수정, 권한 변경)

**예시**:
```yaml
description: |
  Use when user explicitly requests to push commits to remote.
disable-model-invocation: true   # 사용자가 의도적으로 호출해야 함
```

**buddy 적용 점검 대상**: `auto-create-pr`, `ship-release`, `setup-canary-deploy`, `setup-rollback-runbook` 등 release/deploy 관련 명령.

#### 1.2.3 `user-invocable`

**역할**: `false`로 설정 시 사용자 명령으로 호출 불가, Claude만 자동 호출.

**언제 `false`로 설정해야 하는가**:
- **배경 지식/패턴 라이브러리** (다른 skill이 내부적으로만 호출)
- **메타 skill** (다른 skill의 작동 방식을 정의)

**예시**:
```yaml
description: |
  Pattern library for QA tier classification used by verify-quality.
user-invocable: false
```

**buddy 적용 점검 대상**: `classify-qa-tiers`, `classify-review-risks`, `apply-builder-ethos` 같은 pattern library 스킬들.

#### 1.2.4 `allowed-tools`

**역할**: Skill 활성화 중 특정 도구를 pre-approve. 매번 사용자에게 권한 묻지 않음.

**언제 사용**:
- Skill이 특정 도구만 사용하면 명시
- 보안 관점에서 사용 도구를 좁힘

**예시**:
```yaml
allowed-tools:
  - Read
  - Grep
  - Glob
  # Bash, Edit, Write는 사용 안 함
```

#### 1.2.5 기타 옵션

| 키 | 의미 | 사용 시점 |
|----|------|---------|
| `arguments` | 인자 schema 정의 | 인자가 있는 command |
| `paths` | skill이 작동하는 경로 패턴 | 특정 파일/디렉토리에만 동작 |
| `shell` | 셸 명령 사전 실행 | dynamic context 수집 |
| `model` | 특정 모델 강제 | 특정 모델에 최적화된 skill |
| `effort` | thinking budget | 깊은 사고가 필요한 skill |
| `argument-hint` | 사용자에게 보이는 인자 hint | UX 향상 |
| `context: fork` | 격리된 subagent 실행 | 큰 context 작업 격리 |
| `agent: Explore` | 특정 agent로 실행 | 탐색 전용 skill |

### 1.3 buddy 적용 가이드

#### 1.3.1 router/SKILL.md (유일한 auto-discoverable)

```yaml
---
name: router
description: |
  Use when a buddy command requests dispatch to a target PROCEDURE.
  Reads ${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md and executes
  its instructions.
---
```

#### 1.3.2 plugin/commands/*.md (모든 슬래시 커맨드)

```yaml
---
description: <간결한 1-2 문장 — 사용자에게 보임>
argument-hint: "<인자 형식>"
disable-model-invocation: true   # buddy 모든 command 표준
---
```

**왜 buddy의 모든 command가 `disable-model-invocation: true`인가**: ADR-2026-05-09 결정 — Claude의 자동 호출보다 router skill을 통한 명시적 routing을 우선. command는 사용자만 호출 가능, router가 description matching으로 dispatch 담당.

---

## §2. Skill 본문 구조 — LLM 처리 최적화

> 본 섹션은 모든 PROCEDURE.md (및 SKILL.md)의 본문 작성 표준.

### 2.1 권장 섹션 순서

LLM이 가장 잘 따르는 본문 구조는 다음 순서다:

```markdown
# <skill-name>

<1-3 줄: skill의 정체성 + 진입/종료 조건>

## 1. 페르소나 (선택)
당신은 <역할>이다. 당신의 책임은 <범위>이다.

## 2. Input Requirements
| Input | Required | Type | Source | 미제공 시 |
| ... | ... | ... | ... | ... |

## 3. Output Contract
| Output | Type | Format | Consumers |
| ... | ... | ... | ... |

## 4. 이 스킬을 사용하는 경우 (When to invoke)
- 상황 1
- 상황 2

## 5. 이 스킬을 사용하지 않는 경우 (When NOT to invoke)
- 상황 X — 대안: 다른 스킬

## 6. 핵심 원칙 (Principles)
1. 원칙 1 — 이유
2. 원칙 2 — 이유

## 7. 실행 절차 (Steps)
### Step 1. <imperative 동사>
<구체적 동작>

### Step 2. <imperative 동사>
...

## 8. 출력 형식 (Output Format)
응답은 다음 구조로:
- 요약 (50단어 이내)
- 근거 (bullet 3-5)
- ...

## 9. 검증 체크리스트 (Verification)
- [ ] 항목 1
- [ ] 항목 2

## 10. Anti-patterns
| ❌ | ✅ | 이유 |

## 11. 다음 단계 (Next Steps)
→ <다음 skill> — <연결 이유>

## 12. 참조 (References)
- [상세 내용](reference.md)
```

**섹션 순서가 중요한 이유**: LLM은 본문을 위에서 아래로 처리한다. 페르소나/Input을 먼저 명시하면 후속 절차 해석이 더 정확해진다. 검증 체크리스트와 anti-pattern을 뒤에 두면 "마지막에 본 정보"가 출력 직전 강하게 영향을 미친다.

### 2.2 핵심 원칙

#### 2.2.1 본문 길이 — 500줄 이하

**근거**: Skill 본문은 호출 시점에 컨텍스트에 고정되어 세션 전체에 남는다. 길수록 토큰 비용 누적.

**가이드**:
- 500줄 초과 → 분리 검토 필수
- 1000줄 초과 → 거의 무조건 분리
- 분리 방법: `reference.md`, `examples.md`, `templates.md` 등 보조 파일로 lazy-load

**예외**: 한 번에 다 읽혀야 하는 절차(예: critical security audit checklist)는 길어도 OK.

#### 2.2.2 동적 상태는 inline 주입

**Bad**:
```markdown
## 실행 절차
1. 현재 git status를 확인하고
2. 변경된 파일이 있으면...
```
(LLM이 git status를 어떻게 알 것인가? 추측해야 함)

**Good**:
```markdown
## 실행 절차

### Step 1. 현재 상태 확인

다음은 현재 git status다:

!`git status --short`

위 결과를 보고...
```
(`` !`...` `` 문법으로 실제 명령 결과가 본문에 주입됨)

**주입 가능한 dynamic context 예시**:
- `` !`git diff HEAD` `` — 마지막 커밋 대비 변경
- `` !`git log --oneline -10` `` — 최근 커밋 10개
- `` !`gh pr view --json title,body,state` `` — 현재 PR 상태
- `` !`jq . package.json` `` — 패키지 정보
- `` !`ls -la docs/` `` — 디렉토리 내용

#### 2.2.3 절차는 imperative

**Bad**:
```markdown
## 실행 절차

이 단계에서는 사용자가 입력한 데이터를 검증할 수도 있고,
검증을 건너뛸 수도 있습니다. 상황에 따라 판단해 주세요.
```
(모호한 조건문, 책임 회피)

**Good**:
```markdown
## 실행 절차

### Step 1. 입력 검증

`$ARGUMENTS`가 비어있으면 사용자에게 다음 질문을 하라:
> "어떤 작업을 하시나요?"

`$ARGUMENTS`가 있으면 Step 2로 진행한다. (검증 skip 조건: --skip-validation 플래그 명시)
```
(명령형, 조건 명확, 분기 explicit)

#### 2.2.4 Few-shot 예시 — 짧으면 본문, 길면 분리

**본문 inline 권장**:
- 1-2개 예시
- 각 예시 10줄 이내
- 입력 → 출력 형식

**보조 파일 분리 권장**:
- 5개 이상의 예시
- 각 예시가 20줄 초과
- 예시 자체가 길이 늘리는 주범

**파일명 컨벤션**: `examples.md`, `examples/case-{n}.md`

#### 2.2.5 Forcing question — 본문에 명시

사용자에게 결정 받을 지점은 본문에 명시한다 (외부 `AskUserQuestion` 같은 도구 의존 X — LLM이 명시된 forcing question을 더 잘 따름).

**패턴**:
```markdown
### Step 3. 사용자 확인 gate

다음 중 하나를 선택하라:

A) 진행 — 위 분석 결과를 신뢰하고 다음 단계로
B) 수정 — 어느 부분이 잘못됐는지 알려주세요
C) 중단 — 작업을 멈추고 다른 접근 시도

사용자 응답을 받은 후에만 다음 step으로 진행한다.
```

**Anti-pattern**:
```markdown
필요하면 사용자에게 물어보세요.
```
(언제, 무엇을, 어떻게 물어보는지 불명확)

#### 2.2.6 출력 형식 강제 — 명시적 schema

**Bad**:
```markdown
결과를 정리해서 보여줘.
```

**Good**:
```markdown
## 출력 형식

응답은 다음 구조로:

```yaml
analysis:
  summary: "<50단어 이내 요약>"
  findings:
    - severity: <critical|high|medium|low>
      description: "<무엇이 문제인가>"
      remediation: "<어떻게 고치는가>"
  next_steps: [<list of recommended next actions>]
```

이 형식을 벗어나는 응답은 검증 실패로 처리한다.
```

### 2.3 Anti-patterns 종합

| ❌ Bad | ✅ Good | 이유 |
|--------|--------|------|
| "현재 상황은 X일 것입니다" | `` !`get-status-command` `` | LLM이 추측하지 않게 |
| "A 또는 B를 선택할 수 있습니다" | "A로 진행한다. (B는 X 조건 시만)" | 분기 명확 |
| "나머지는 당신 판단에" | 명시적 체크리스트 | 책임 회피 X |
| "효율적으로 처리해 주세요" | 구체적 절차 + 출력 형식 | 추상적 형용사 X |
| "필요하면 사용자에게..." | "다음 시점에 사용자에게 X를 물어라" | 조건/내용 명시 |
| 500+ 줄 본문 | 분리 후 `reference.md` 링크 | 토큰 비용 |
| 동적 정보 묘사 | inline 주입 | 정확성 |

---

## §3. Persona / Role Assignment

### 3.1 효과 — 공식 입장

**Anthropic의 prompt engineering 공식 가이드 입장**:

> "역할 부여(role assignment)는 Claude의 출력을 task에 맞춰 정렬하는 효과적인 기법이다. 단 모호한 페르소나보다 **구체적**인 페르소나가 효과적이며, persona는 **능력 부여가 아닌 task 프레이밍 도구**임을 이해해야 한다."

**효과의 두 측면**:
1. **Style/Tone 정렬**: 어휘 선택, 강조점, 표현 방식이 일관됨
2. **평가 기준 활성화**: "당신은 보안 전문가다"라고 하면 보안 관점 평가 기준이 응답에 자연스럽게 반영됨

### 3.2 효과적인 경우

✅ **사용 권장**:

| Task 유형 | 효과 | 예시 |
|----------|------|------|
| **구조화된 분석/리뷰** | 평가 기준이 명확해짐 | `review-engineering`, `audit-security` |
| **도메인 전문성 요구** | 전문 어휘가 정확해짐 | `design-data-model` (DBA persona), `review-privacy-data-risk` (compliance persona) |
| **스타일/톤 일관성** | 출력 형식·강조점 정렬 | `critique-plan` (founder persona), `review-design` (designer persona) |
| **여러 관점 대비** | 각 관점이 명확히 분리됨 | autoplan의 4 review (scope/eng/design/devex) |
| **discipline 강제** | 권위 있는 톤이 절차 준수 유도 | `build-with-tdd` (TDD 전문가), `iterate-fix-verify` (refactor 전문가) |

### 3.3 비효과적/역효과인 경우

❌ **사용 비권장**:

| Task 유형 | 이유 | 대안 |
|----------|------|------|
| **창의적 브레인스토밍** | persona가 사고 범위 제한 | persona 없이 open-ended |
| **다각 분석 필요** | 단일 persona가 blind spot 생성 | 여러 persona 순차 적용 (autoplan 패턴) |
| **순수 사실 정리** | 불필요한 주관성 주입 | persona 생략 |
| **데이터 변환/포맷팅** | persona가 의미 없음 | 절차만 명시 |
| **API/CLI 설명** | 중립적 정보 전달이 더 정확 | persona 생략 |

### 3.4 작성 패턴

#### 3.4.1 구체성 우선

**Bad**:
```markdown
당신은 전문가입니다.
```

**Good**:
```markdown
당신은 **10년 이상 fintech 산업에서 OWASP Top 10 + AWS IAM + 카드 데이터 처리 보안을 다뤄온 senior CISO**다.
당신의 역할은 **PR 직전의 코드 변경**을 보고 **결제 데이터 흐름과 인증 경계**의 잠재 취약점을 찾는 것이다.
당신은 **vague한 보안 조언이 아니라 구체적 코드 위치 + 공격 시나리오 + 패치 방향**을 출력한다.
```

#### 3.4.2 페르소나 구성 요소

효과적인 페르소나는 다음 4-5요소를 포함:

| 요소 | 예시 |
|------|------|
| **경력/배경** | "10년 이상 fintech 산업" |
| **전문 영역** | "OWASP Top 10 + AWS IAM" |
| **현재 역할 (이 skill에서)** | "PR 직전 코드 변경 리뷰" |
| **무엇을 보는가** | "결제 데이터 흐름 + 인증 경계" |
| **무엇을 출력하는가 (자가 출력 규칙)** | "vague한 조언 아닌 구체 위치 + 시나리오 + 패치" |

#### 3.4.3 페르소나 위치 — 본문 최상단

`# <skill-name>` 제목 직후, 다른 섹션보다 먼저 배치한다. 이유: LLM이 본문을 위에서 아래로 해석하므로, persona가 후속 모든 섹션의 해석 lens가 된다.

```markdown
# audit-security

당신은 ... (페르소나, 3-5줄)

## Input Requirements
...
```

#### 3.4.4 길이 — 3-7줄

너무 짧으면 (1줄) 추상적, 너무 길면 (10줄+) 본문 토큰 낭비.

### 3.5 buddy 적용 가이드

#### 3.5.1 페르소나 추가가 권장되는 skill 유형

| Phase | 권장 skill 예시 | 페르소나 |
|-------|--------------|---------|
| §1 | `validate-idea`, `assess-business-viability` | YC partner / VC analyst |
| §2 | `identify-actors`, `compose-feature-from-use-cases` | senior product manager |
| §3 | `design-data-model`, `design-api-contract` | senior backend architect |
| §3 review | `review-architecture`, `review-engineering`, `review-design`, `review-devex`, `review-scope` | 각자의 senior 역할 (이미 적용됨) |
| §4 | `plan-build`, `estimate-build-timeline` | engineering manager |
| §5 | `build-with-tdd`, `iterate-fix-verify`, `diagnose-bug` | senior engineer with TDD discipline |
| §6 | `audit-security`, `audit-test-coverage-meaningful`, `audit-accessibility` | 각 도메인 전문가 (CISO/QA lead/a11y advocate) |
| §7 | `prepare-launch-checklist`, `setup-incident-paging` | SRE / launch coordinator |
| §8 | `handle-incident`, `conduct-postmortem`, `analyze-cost-anomaly` | SRE on-call lead |

#### 3.5.2 페르소나 추가가 비권장되는 skill 유형

| 유형 | skill 예시 | 이유 |
|------|----------|------|
| **데이터 변환** | `update-docs-with-code`, `sync-release-docs` | 절차 자체가 명확, persona 불필요 |
| **메타 도구** | `status`, `save-context`, `restore-context` | 객관적 정보 출력 |
| **단순 dispatcher** | `start`, `router` | 라우팅 결정만, persona 무관 |
| **패턴 라이브러리** | `freeze-edit-scope`, `compose-safety-mode` | 기계적 적용 |

#### 3.5.3 페르소나 일관성 정책

여러 skill이 같은 도메인이면 페르소나 톤을 일관화한다. 예:
- 모든 `review-*` 스킬: "senior X manager" 형식 통일
- 모든 `audit-*` 스킬: "전문가 + 평가 기준" 형식 통일
- 모든 `analyze-*` 스킬: "데이터 분석가" 형식 통일

이렇게 하면 사용자가 skill 호출 시 일관된 경험을 얻는다.

### 3.6 페르소나 미사용 시 발생하는 문제

페르소나 없이 작성된 skill의 전형적 문제:
1. **출력 톤이 매번 다름** — 호출 시점의 사용자 발화 톤에 휘둘림
2. **평가 기준이 모호** — "좋다/나쁘다" 판단이 일관되지 않음
3. **도메인 용어 불일치** — 같은 개념을 매번 다른 단어로 표현
4. **깊이 부족** — 일반론적 응답에 머무름 (전문가 관점 결여)

---

## §4. 통합 평가 체크리스트

skill 한 개를 평가할 때 사용하는 점검표. 각 항목 통과 시 1점, 통과율 80% 이상이 합격.

### 4.1 Frontmatter (router/SKILL.md, plugin/commands/*.md만 해당)

- [ ] **F1**: `description` 첫 문장에 "Use when" 또는 "사용 시점" 명시
- [ ] **F2**: `description`에 사용자 자연어 발화 패턴 3개 이상 포함
- [ ] **F3**: `description`이 1,536자 이내, 추상적 형용사 없음
- [ ] **F4**: side-effect 작업이면 `disable-model-invocation: true` 명시
- [ ] **F5**: pattern library / 메타 skill이면 `user-invocable: false` 명시

### 4.2 본문 구조 (모든 PROCEDURE.md)

- [ ] **B1**: 제목 직후 1-3줄 정체성 + 진입/종료 조건 명시
- [ ] **B2**: 본문 500줄 이하 (초과 시 보조 파일 분리)
- [ ] **B3**: Input Requirements 표 존재 (이미 적용됨 — engineering-phases.md §4 형식)
- [ ] **B4**: Output Contract 표 존재 (이미 적용됨)
- [ ] **B5**: "이 스킬을 사용하는 경우" 섹션 명시
- [ ] **B6**: "이 스킬을 사용하지 않는 경우" 섹션 명시 (대안 skill 안내)
- [ ] **B7**: 실행 절차가 imperative (Step 1, Step 2, ... 명령형)
- [ ] **B8**: 분기 조건이 명시적 (모호한 "필요 시" 없음)
- [ ] **B9**: 동적 상태는 `` !`...` `` 문법으로 inline 주입 (있는 경우)
- [ ] **B10**: 출력 형식 명시 (YAML/markdown 구조 강제)
- [ ] **B11**: 검증 체크리스트 또는 anti-pattern 섹션 존재
- [ ] **B12**: 다음 단계 (Next Steps) 명시

### 4.3 Persona (페르소나 권장 skill 한정)

- [ ] **P1**: 본문 최상단 (제목 직후)에 페르소나 정의
- [ ] **P2**: 페르소나 구성 요소 4-5개 포함 (경력/전문영역/현재역할/관점/출력규칙)
- [ ] **P3**: 페르소나 3-7줄 길이
- [ ] **P4**: 같은 도메인 skill과 페르소나 톤 일관

### 4.4 점수 계산

| 카테고리 | 항목 수 | 가중치 |
|---------|--------|-------|
| Frontmatter (해당 파일만) | 5 | 1.0 |
| 본문 구조 | 12 | 2.0 |
| Persona (해당 skill만) | 4 | 1.5 |

**최종 점수 = (통과 항목 가중치 합) / (전체 항목 가중치 합) × 100**

| 점수 | 평가 |
|------|------|
| 90+ | 우수 — 다른 skill의 참고 모델 |
| 80-89 | 합격 — 운영 가능 |
| 70-79 | 보강 필요 — 부족 항목 즉시 개선 |
| 70 미만 | 재작성 권장 — 구조적 결함 |

---

## §5. 적용 가이드

### 5.1 신규 skill 작성 시 — 표준 템플릿

새 PROCEDURE.md를 작성할 때 다음 템플릿으로 시작:

```markdown
# <skill-name>

<1-3줄: skill의 정체성 + 진입/종료 조건 + 다음 phase>

## Persona (선택, 권장)

당신은 <경력 + 전문영역>이다.
당신의 역할은 <현재 skill에서의 책임 범위>이다.
당신은 <무엇을 보고 무엇을 출력하는가>.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| ... | ✅ | knowledge/artifact/decision | ... | "질의문" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| ... | artifact/decision | ... | ... |

---

## 1. 이 스킬을 사용하는 경우

- 상황 1
- 상황 2

## 2. 이 스킬을 사용하지 않는 경우

- 상황 X — 대안: `<other-skill>`

## 3. 핵심 원칙

1. **원칙 1** — 이유
2. **원칙 2** — 이유

## 4. 실행 절차

### Step 1. <imperative 동사>

<구체적 동작>

(필요시 동적 context 주입: !`command`)

### Step 2. <imperative 동사>

...

## 5. 출력 형식

응답은 다음 구조로:

```yaml
result:
  summary: "<50단어>"
  ...
```

## 6. 검증 체크리스트

- [ ] 항목 1
- [ ] 항목 2

## 7. Anti-patterns

| ❌ | ✅ | 이유 |
|----|----|------|
| ... | ... | ... |

## 8. 다음 단계

→ `<next-skill>` — <연결 이유>

## 9. 참조 (긴 내용 분리 시)

- [상세 spec](./reference.md)
- [예시 모음](./examples.md)
```

### 5.2 기존 skill 개선 시 — 작업 순서

1. **§4 체크리스트로 점수 산정**
2. **70 미만**: 재작성 후보 — 우선순위 결정 후 표준 템플릿으로 재작성
3. **70-89**: 부족 항목 명시 후 점진 개선
4. **90+**: 다른 skill의 참고 모델로 등록

### 5.3 우선순위 — 어떤 skill부터 개선할 것인가

| Tier | 기준 | 작업 강도 |
|------|------|---------|
| 1순위 | Phase orchestrator (concretize-idea, design-system 등 9개) | 사용자 진입점 — 영향 큰 |
| 2순위 | Cross-cutting (status, save-context, decompose-blocker 등) | 빈번한 호출 |
| 3순위 | Phase별 stage skill | Phase 1부터 순차 |
| 4순위 | Pattern library (classify-qa-tiers 등) | 다른 skill이 호출하는 인프라 |

---

## §6. 본 가이드의 한계 / 가정

### 6.1 명시 가정

| # | 가정 | 잠재 반론 |
|---|------|----------|
| G1 | LLM은 페르소나/구조화된 절차를 더 정확히 따른다 | 모델별로 효과 차이 있음 (Sonnet vs Opus) |
| G2 | 500줄 임계는 토큰 비용 vs 정확성의 trade-off | 정확성 요구가 매우 높으면 길어도 OK |
| G3 | 페르소나 구체성이 더 효과적 | 일부 task는 중립이 더 정확 (§3.3 참조) |
| G4 | Anthropic 공식 가이드가 buddy에도 그대로 적용 | buddy의 PROCEDURE.md 형식 차이는 §0에서 보정 |

### 6.2 정기 갱신 trigger

| trigger | 갱신 부분 |
|---------|---------|
| Anthropic이 새 frontmatter 키 발표 | §1 |
| 새 prompt engineering best practice 공개 | §2 / §3 |
| buddy에서 실험적으로 발견된 패턴 | 해당 섹션 |
| Skill 평가 후 공통 결함 발견 | §4 체크리스트 보강 |

---

## §7. 참조

### 7.1 Anthropic 공식 docs

- **Claude Code Skills**: https://code.claude.com/docs/en/skills.md
- **Plugin 시스템**: https://code.claude.com/docs/en/plugins.md
- **Prompt Engineering 개요**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering
- **System Prompts**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/system-prompts
- **Be Clear and Direct**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/be-clear-and-direct
- **Use XML Tags**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags
- **Chain of Thought**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/chain-of-thought
- **Use Examples**: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-examples

### 7.2 buddy 자체 문서

- `plugin/skills/router/references/engineering-phases.md` — Phase 정의 + I/O Contract 표준 형식
- `plugin/skills/router/references/skill-catalog.md` — 전체 skill 카탈로그
- `plugin/skills/router/references/routing-rules.md` — 라우팅 충돌 결정
- `docs/plugin-skills-classification-matrix.md` — 152 skill 분류
- `docs/plugin-skills-flow-graph.md` — I/O Contract 전이 그래프

### 7.3 참고 buddy skill (페르소나 적용 모범 사례)

- `review-engineering` — senior engineering manager
- `review-design` — senior product designer
- `review-devex` — developer advocate
- `audit-security` — CSO mode
- `critique-plan` — CEO/founder

이 5개는 페르소나 작성의 참고 모델로 등록.

---

**문서 끝**. 본 가이드는 buddy skill 개선의 단일 평가 기준. 신규/갱신 시 §5의 절차를 따른다.
