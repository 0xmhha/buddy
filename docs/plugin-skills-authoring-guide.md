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

| 영역 | 적용 파일 | 비적용 파일 |
|------|---------|-----------|
| **§1 Frontmatter (전체)** | `plugin/skills/router/SKILL.md`, `plugin/commands/*.md` | `plugin/skills/<name>/PROCEDURE.md` (frontmatter 없음) |
| **§2 본문 구조** | 모든 `PROCEDURE.md` + `router/SKILL.md` 본문 | command 파일 본문은 짧으므로 일부만 |
| **§3 Persona** | 모든 `PROCEDURE.md` (선택적이지만 권장) | router/command는 페르소나 무관 |

**중요**: buddy의 PROCEDURE.md는 frontmatter가 없기 때문에 §1의 `description`, `disable-model-invocation`, `user-invocable`, `allowed-tools` 등은 PROCEDURE.md에 직접 적용할 수 없다. 대신:

- PROCEDURE.md의 dispatch trigger 텍스트는 `plugin/skills/router/references/skill-catalog.md`에 중앙 등록 (description 역할)
- 사용자 호출 가능 여부는 `plugin/commands/<name>.md` 파일의 **존재 여부**로 제어 (있으면 user-invocable, 없으면 router 경유 자동 호출만)
- 자동 호출 차단은 `plugin/commands/<name>.md`의 `disable-model-invocation: true`로 제어 (§1.3.2 표준)

---

## §1. Skill 호출 시점 결정 — Frontmatter

> **적용 대상**: `plugin/skills/router/SKILL.md`, `plugin/commands/*.md`만. PROCEDURE.md는 frontmatter가 없으므로 본 섹션 모든 키 직접 적용 불가 — buddy 등가 매핑은 §0.2 참조.

### 1.1 Frontmatter의 역할

YAML frontmatter는 Claude가 skill을 "**언제 호출할 것인가**"를 결정하는 메타데이터다. Claude는 progressive disclosure 방식으로 skill을 처리한다:

1. **세션 시작 시점**: 모든 skill의 frontmatter(특히 `description` + `when_to_use`)가 listing context에 로드된다. 본문은 아직 로드되지 않음.
2. **dispatch 결정 시점**: Claude는 사용자 발화/맥락을 frontmatter 텍스트와 매칭하여 skill 호출 여부를 결정한다 (description-based dispatch).
3. **호출 후**: 해당 skill의 본문이 컨텍스트에 주입되어 실행 절차에 활용된다.

따라서 frontmatter 품질이 skill의 **발견성**(2단계)을 좌우하고, 본문 품질이 **실행 정확성**(3단계)을 좌우한다.

### 1.2 핵심 키 상세

#### 1.2.1 `description` (가장 중요)

**역할**: Claude가 사용자 발화/맥락을 분석하여 이 skill을 호출할지 결정하는 기준.

**규칙**:
- **Truncation cap: `description` + `when_to_use` 합산 1,536자** (skill listing에서 잘림 — frontmatter 작성 자체의 한계는 아니나, 초과분은 Claude가 dispatch 결정 시 보지 못함)
- Cap은 `maxSkillDescriptionChars` 설정으로 변경 가능 (전역 설정이므로 일반적으론 default 유지)
- 핵심 use case를 **앞쪽**에 배치 (cap 초과 시 뒷부분이 잘림)
- 첫 문장에 **"use when"** 키워드 배치 (가장 강한 trigger signal)
- 사용자가 표현할 법한 자연어 패턴을 포함
- 추상적 표현보다 구체적 키워드

**출처**: https://code.claude.com/docs/en/skills.md — "the combined `description` and `when_to_use` text is truncated at 1,536 characters in the skill listing to reduce context usage"

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
- [ ] `description` + `when_to_use` 합산 1,536자 이내 (실측 권장 500-800자, 핵심 use case는 앞쪽 배치)

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

**buddy 적용 위치**: `plugin/commands/<name>.md` 파일에만 적용 가능 (PROCEDURE.md는 frontmatter 없음).

**buddy 표준**: §1.3.2에 따라 **모든 `plugin/commands/*.md`**는 `disable-model-invocation: true`로 설정 (Claude 자동 호출보다 router 경유 명시적 dispatch 우선). 따라서 점검은 "side-effect 작업인지 여부와 무관하게 표준 준수했는가"가 된다.

**점검 우선순위 (side-effect 큰 command부터)**: `plugin/commands/auto-create-pr.md`, `plugin/commands/ship-release.md`, `plugin/commands/setup-canary-deploy.md`, `plugin/commands/setup-rollback-runbook.md` — release/deploy 관련 command. 누락 시 즉시 추가.

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

**buddy 적용 — 구조적 차이 주의**:

buddy의 PROCEDURE.md는 frontmatter가 없으므로 `user-invocable: false`를 **PROCEDURE.md에 직접 적용할 수 없다**. 대신 buddy는 다음 메커니즘으로 동등 효과를 얻는다:

| 정책 | buddy 구현 방식 |
|------|---------------|
| 사용자 호출 가능 (`user-invocable: true` 효과) | `plugin/commands/<name>.md` 파일 **생성** |
| 사용자 호출 불가 (`user-invocable: false` 효과) | `plugin/commands/<name>.md` 파일 **부재** (router 경유 자동/내부 호출만 가능) |

**buddy 적용 점검 대상**: `classify-qa-tiers`, `classify-review-risks`, `apply-builder-ethos` 등 **pattern library 스킬**들은 다른 스킬에서 내부적으로만 호출되어야 한다 → `plugin/commands/<name>.md` 파일이 **없어야 함**. 현재 상태를 확인하려면:

```bash
# pattern library 스킬에 commands 파일이 존재하는지 점검
for skill in classify-qa-tiers classify-review-risks apply-builder-ethos; do
  test -f "plugin/commands/$skill.md" && echo "WARN: $skill has command (should not)" || echo "OK: $skill no command"
done
```

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

**buddy 적용 — 현 정책**: buddy의 PROCEDURE.md는 frontmatter가 없어 `allowed-tools`를 직접 선언할 수 없다. 대신 `plugin/commands/<name>.md`에 선언하면 해당 command로 진입한 router → PROCEDURE 실행 전체가 pre-approve된 도구 범위로 제한된다. 현재 buddy commands는 `allowed-tools`를 명시하지 않음 (전체 도구 권한 가정) — 보안 민감 command(예: `auto-create-pr`, `ship-release`)에 도입 검토 필요 (TODO 항목).

#### 1.2.5 기타 옵션

| 키 | 의미 | 사용 시점 | buddy 적용 |
|----|------|---------|---------|
| `when_to_use` | dispatch trigger 보조 텍스트 — `description` 뒤에 append되어 1,536자 cap을 함께 공유 | description이 짧고 trigger 문구를 분리하고 싶을 때 | description 통합 작성 권장 (분리 사용 X) |
| `arguments` | 인자 schema 정의 | 인자가 있는 command | `plugin/commands/*.md`에 명시. PROCEDURE.md는 적용 불가 |
| `paths` | skill이 작동하는 경로 패턴 (글로브) | 특정 파일/디렉토리에만 동작 | buddy는 router 경유 dispatch이므로 일반적으로 사용 X |
| `shell` | 셸 명령 사전 실행 | dynamic context 수집 | command/router에서 사용 가능. PROCEDURE.md 내부의 `` !`cmd` `` 인라인 주입과 구분 |
| `model` | 특정 모델 강제 (model ID 예: `claude-opus-4-7`) | 특정 모델에 최적화된 skill | buddy는 사용자 선택 모델 존중이 기본. 사용 권장 X (예외: 모델별 specialization이 명확한 경우만) |
| `effort` | thinking budget (low/medium/high/xhigh/max) | 깊은 사고가 필요한 skill. 사용 가능 level은 모델별로 다름. 미설정 시 세션 effort 상속 | 복잡한 분석/리뷰 command(예: `ultrareview`, `consult-codex`)에 `high`/`xhigh` 검토 |
| `argument-hint` | 사용자에게 보이는 인자 hint | UX 향상 | 인자 받는 command 전부 권장 |
| `context: fork` | 격리된 subagent 실행 | 큰 context 작업 격리 | buddy는 `run` composition command가 처리. 개별 skill에서 사용 X |
| `agent: Explore` | 특정 agent로 실행 | 탐색 전용 skill | buddy는 Explore subagent를 router 내부에서 호출. command frontmatter 직접 사용 X |

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

### 1.4 Claude 모델별 가이드 — `model` / `effort` 키 적용

> **출처**: https://code.claude.com/docs/en/model-config (2026-06-01 fetch). 모델 라인업·effort 매트릭스는 Anthropic 업데이트에 따라 변경되므로 본 §1.4는 정기 갱신 대상.

#### 1.4.1 Model alias / ID 형식

`model` 키에는 alias 또는 full model name을 사용한다:

| 종류 | 예시 | 비고 |
|------|------|------|
| **Alias** (권장) | `sonnet`, `opus`, `haiku`, `sonnet[1m]`, `opus[1m]`, `opusplan`, `best`, `default` | 모델 업데이트 시 자동으로 최신 버전으로 추적 |
| **Full model name** | `claude-opus-4-7`, `claude-opus-4-8`, `claude-sonnet-4-6`, `claude-haiku-4-5` | 특정 버전 핀 고정 — 재현성 보장 |
| **Suffix `[1m]`** | `claude-opus-4-8[1m]` | 1M token context window (Opus 4.6+ / Sonnet 4.6 지원) |

**buddy 정책**: 사용자가 `/model`로 선택한 모델 존중이 기본. PROCEDURE.md/command에 `model` 키를 명시적으로 지정하는 것은 권장하지 않음 — **모델별 specialization이 절대적으로 필요한 경우만** 한정 사용.

#### 1.4.2 Effort level 모델별 지원

`effort` 키는 모델별로 지원되는 level이 다르다. 미지원 level 설정 시 fallback이 발생한다 (가장 가까운 하위 level로):

| 모델 | 지원 effort level | 기본값 |
|------|----------------|-------|
| **Opus 4.8 / Opus 4.7** | `low`, `medium`, `high`, `xhigh`, `max` | Opus 4.8 = `high`, Opus 4.7 = `xhigh` |
| **Opus 4.6 / Sonnet 4.6** | `low`, `medium`, `high`, `max` (xhigh 미지원 → `high` fallback) | `high` |
| **Haiku 4.5 (및 기타 Haiku)** | **effort 미지원** | (해당 없음) |

**Effort level 선택 가이드**:

| Level | 사용 시점 |
|-------|---------|
| `low` | 짧고 latency-sensitive한 작업, 지능 요구 낮음 (예: 단순 데이터 변환) |
| `medium` | cost-sensitive하며 지능 일부 trade-off 가능 |
| `high` | 균형 — 일반 코딩 작업 기본 |
| `xhigh` | 깊은 추론, token spend 증가 (Opus 4.7/4.8만) |
| `max` | 가장 깊은 추론, 토큰 제약 없음 — overthinking 위험. 채택 전 테스트 필수. session-only |

**buddy 정책**: 일반 PROCEDURE.md는 session effort 상속(기본값). `consult-codex`, `critique-plan`, `ultrareview`처럼 **복잡한 분석/리뷰 command**에 한해 `effort: high` 또는 `xhigh` 명시 검토.

#### 1.4.3 모델별 행동 차이 (PROCEDURE.md 작성 시 고려)

| 모델 | 일반 특성 | PROCEDURE.md 작성 시 고려 |
|------|----------|-------------------------|
| **Opus** (4.7/4.8) | 깊은 추론, 다단계 분석, cross-reference에 강함 | 복잡한 워크플로우·아키텍처·리뷰 skill에 자연 적합. 모호한 절차도 능동적으로 보완하는 경향 |
| **Sonnet** (4.6) | 처리량·비용 효율, 패턴 매칭 작업에 적합 | 명령형 절차·체크리스트·체계화된 분기는 안전. 복잡한 추론이 필요한 step은 `effort: high` 보강 |
| **Haiku** (4.5) | 빠르고 저렴, 배경 작업에 사용. effort 미지원 | PROCEDURE.md 본문 작성 시 **추론 부담이 적은 절차**로 작성. 분기·체크리스트·schema 강제는 OK. 깊은 분석을 요구하는 step은 부적합 |

**buddy 정책**: PROCEDURE.md는 **모델 중립적으로 작성**한다. 즉 어느 모델이 호출하더라도 정확히 동작하도록 절차·schema·분기를 명시화. 모델별 가정에 의존하는 본문(예: "Opus라면 알아서 처리할 것")은 피한다.

### 1.5 Skill Catalog Entry 평가 기준 — Frontmatter의 4번째 위치

> **배경**: buddy의 PROCEDURE.md는 frontmatter가 없는 대신, `plugin/skills/router/references/skill-catalog.md`의 **각 entry가 사실상 frontmatter의 description 역할**을 한다. router가 dispatch 결정 시 이 entry를 본다 (가이드 §0.1 표 참조).
>
> **누락 발견 2026-06-01**: 본 §1.5는 evaluate-skill이 PROCEDURE.md / command.md / router/SKILL.md만 평가하고 **catalog entry 품질을 평가하지 못하는 결함**을 메우기 위해 신설.

#### 1.5.1 Catalog entry 표 형식

```
| `<skill-name>` | <호출 방법> | <1줄 description — 트리거 키워드 + 목적> |
```

| 컬럼 | 의미 | 예시 |
|------|------|------|
| `<skill-name>` | PROCEDURE.md 디렉토리 이름과 동일 | `concretize-idea` |
| `<호출 방법>` | 3종: `command + dispatch` / `dispatch only (via /buddy:start)` / `(직접 호출 불가, ...)` | `command + dispatch` |
| `<1줄 description>` | router의 dispatch 결정용 트리거 텍스트. 가이드 §1.2.1 description 규칙과 동일 원칙 적용 | "idea/concept → PRD + 사업성 검증. greenfield 진입점" |

#### 1.5.2 Catalog Entry 평가 항목 (CE1-CE5)

PROCEDURE.md 평가 시 **항상 같이 평가**한다 (paired evaluation — §1.6 참조).

- **CE1**: entry가 catalog의 phase별 표 중 정확한 phase에 위치 — `engineering-phases.md` §2 phase 정의와 일치
- **CE2**: 호출 방법 컬럼이 3종 중 하나로 명시 + 실제 `plugin/commands/<name>.md` 존재 여부와 일치
  - `command + dispatch` → command 파일 존재해야 함
  - `dispatch only` → command 파일 부재해야 함 (가이드 §1.2.3 buddy 등가 매핑)
  - `(직접 호출 불가, X 경유)` → command 파일 부재 + 경유 경로 명시
- **CE3**: description 1줄에 **트리거 키워드 + 목적**을 모두 담음. 추상 형용사("종합적", "효과적", "다양한") 0개
- **CE4**: 같은 phase 내 다른 entry와 **차별점 명확** — 다른 entry와 80% 이상 키워드 중복 시 fail
- **CE5**: description 길이 50-300자 권장 (너무 짧으면 trigger 부족, 너무 길면 표 가독성 저하)

#### 1.5.3 Catalog entry 평가의 가중치

- **카테고리 가중치**: 1.0 (Frontmatter F1-F5와 동등 — dispatch 신호이지만 단일 라인이라 본문 가중치 2.0보다 낮음)
- **항목 수**: 5
- **카테고리 최대 점수**: 5 × 1.0 = 5

평가 케이스 분모 갱신은 §4.4 paired evaluation 통합 케이스(PC1-PC5)로 처리. 상세 §4.4 참조.

### 1.6 Paired Evaluation 정책 — 평가 누락 방지 (Lost-in-the-middle 회피)

> **배경**: buddy 한 skill의 dispatch 신호가 4 위치(router/SKILL.md frontmatter / commands/*.md frontmatter / PROCEDURE.md 본문 / catalog entry)에 분산되어 있다. evaluate-skill이 단일 위치만 평가하면 다른 위치의 결함을 놓친다 (2026-06-01 실제 발생).

#### 1.6.1 정책

사용자가 `evaluate-skill <name>`을 호출하면 evaluate-skill은 **다음 4 위치를 자동 동반 평가**한다 (paired evaluation):

| # | 위치 | 평가 조건 | 미존재 시 |
|---|------|---------|---------|
| 1 | `plugin/skills/<name>/PROCEDURE.md` | 항상 (이게 없으면 skill 자체 부재 → 에러) | 에러: skill 부재 |
| 2 | `plugin/commands/<name>.md` | 파일이 **존재할 때만** 평가. 부재는 pattern library skill의 정상 상태 | F1-F5 카테고리 분모에서 제외 (PC3/PC4 케이스) |
| 3 | `plugin/skills/router/references/skill-catalog.md` 의 `<name>` entry | grep으로 entry 발견 시 평가 | CE1-CE5 fail (등재 누락은 router가 dispatch 못 함) |
| 4 | `plugin/skills/router/SKILL.md` | `<name>` == `router`일 때만 평가 (router 자신 평가) | 일반 평가에서는 #1-#3만 평가 |

#### 1.6.2 출력 통합

paired evaluation 결과는 **위치별 점수 + 통합 점수** 둘 다 출력한다:

```yaml
paired_evaluation:
  procedure:      { passed: 11, total: 12, score: 22 }    # B1-B12
  command:        { passed: 4,  total: 5,  score: 4 }     # F1-F5 (없으면 N/A)
  catalog_entry:  { passed: 5,  total: 5,  score: 5 }     # CE1-CE5
  persona:        { passed: 3,  total: 4,  score: 4.5 }   # P1-P4 (적용 시)
  total:          { passed_weighted_sum: 35.5, denominator: 40, score: 88.75, grade: "합격" }
  case: PC1
```

#### 1.6.3 단일 위치 평가의 정당한 사용처

paired가 기본이지만 다음 경우 단일 위치 평가(C1-C4) 정당화:

- 새 catalog entry 추가 직후 entry만 점검 → C-CE (catalog entry 단독)
- command.md frontmatter만 수정 후 점검 → C1 또는 C2 (command 단독)
- PROCEDURE.md 본문만 수정 후 점검 → C3 또는 C4 (PROCEDURE 단독)

명시적으로 단일 위치만 평가하라고 사용자가 요청한 경우 외에는 항상 paired.

#### 1.6.4 evaluate-skill 동기 요구사항

본 §1.6 정책은 `plugin/skills/evaluate-skill/PROCEDURE.md` Step 1 (입력 파싱) + Step 2-7 (평가 절차) + Step 9 (리포트 출력)와 **1:1 동기화**되어야 한다. 동기화 실패 시 paired 누락 결함 재발.

---

## §2. Skill 본문 구조 — LLM 처리 최적화

> 본 섹션은 모든 PROCEDURE.md (및 SKILL.md)의 본문 작성 표준.

### 2.0 PROCEDURE.md의 프롬프트 역할

PROCEDURE.md 본문은 router가 호출한 직후 Claude의 **세션 컨텍스트에 시스템 프롬프트 영역으로 주입**된다 (사용자 메시지가 아님). 따라서 작성 관점은:

| 구분 | PROCEDURE.md 본문 | 사용자 메시지 (런타임) |
|------|----------------|---------------------|
| 역할 | 시스템 프롬프트 (instruction set) | 작업 입력 (data + ask) |
| 시점 | 호출 시 1회 주입, 세션 내 지속 | 매 턴 추가 |
| 작성자 | skill 저자 | 사용자 |
| 톤 | 명령형 ("...하라", "...한다") | 자연어 발화 |
| 영속성 | 세션 끝까지 컨텍스트에 남음 | 압축 대상이 될 수 있음 |

이 차이가 §2.1~§2.2의 작성 규칙(imperative 톤, 명시적 분기, schema 강제 등)의 근거다. PROCEDURE.md는 LLM에게 "지금부터 이 절차를 따르라"고 지시하는 메타 명령 문서로 작성한다.

### 2.1 권장 섹션 순서 — 필수 / 권장 / 선택 3-Tier 분류

**(2026-06-01 재분류 A-M1)**: 기존 12 섹션을 그대로 모두 작성하면 본문이 비대해져 lost-in-the-middle 위험이 커진다. **필수 4 + 권장 4 + 선택 4** 3-Tier로 재분류하여, 모든 PROCEDURE.md가 12 섹션을 채울 필요 없게 한다.

#### 섹션 분류

| Tier | # | 섹션 | 의미 |
|------|---|------|------|
| 🔴 **필수 (4)** | 1 | `# <skill-name>` + 1-3줄 정체성/진입·종료 조건 | skill 식별 |
| 🔴 필수 | 2 | `## Input Requirements` 표 | I/O Contract (`engineering-phases.md` §4 표준) |
| 🔴 필수 | 3 | `## Output Contract` 표 | I/O Contract |
| 🔴 필수 | 4 | `## 실행 절차` (Step 1, Step 2, ... imperative) | skill의 동작 본체 |
| 🟡 **권장 (4)** | 5 | `## Persona` (페르소나 권장 skill만 — §3.5.1 참조) | task 프레이밍 |
| 🟡 권장 | 6 | `## 이 스킬을 사용하는 경우` | dispatch 보조 |
| 🟡 권장 | 7 | `## 이 스킬을 사용하지 않는 경우` (+ 대안 skill) | dispatch 보조 |
| 🟡 권장 | 8 | `## 출력 형식` (YAML/markdown schema) | 결과 일관성 |
| 🟢 **선택 (4)** | 9 | `## 핵심 원칙` | 절차의 근거 설명 (절차가 짧으면 생략) |
| 🟢 선택 | 10 | `## 검증 체크리스트` | self-check (긴 절차에서만) |
| 🟢 선택 | 11 | `## Anti-patterns` | 흔한 실수 방지 (반복 호출되는 skill에서 가치) |
| 🟢 선택 | 12 | `## 다음 단계` (→ next skill) + `## 참조` | 흐름·외부 자료 link |

#### 섹션 배치 순서 (포함하는 경우만)

```
1 (정체성) → 5 (Persona) → 2 (Input) → 3 (Output)
→ 6 (사용 case) → 7 (비사용 case) → 9 (핵심 원칙) → 4 (실행 절차)
→ 8 (출력 형식) → 10 (체크리스트) → 11 (Anti-patterns) → 12 (다음 단계)
```

#### 작성 가이드

- **PROCEDURE.md 본문 ≤ 300줄**: 필수 4개 + 권장 일부만으로 충분. 선택은 거의 생략.
- **PROCEDURE.md 본문 300-500줄**: 필수 + 권장 모두 + 선택 일부.
- **본문 500줄 초과**: §2.2.1 long-context 전략 적용 — 선택 섹션을 `references/<topic>.md`로 분리.

**섹션 순서가 중요한 이유**: LLM은 본문을 위에서 아래로 처리한다. 페르소나/Input을 먼저 명시하면 후속 절차 해석이 더 정확해진다. 검증 체크리스트와 anti-pattern을 뒤에 두면 "마지막에 본 정보"가 출력 직전 강하게 영향을 미친다.

#### 기존 PROCEDURE.md 트랜지션 정책 (회귀 방지)

**기존 PROCEDURE.md가 12 섹션 일부만 가지고 있어도 evaluate-skill 점수가 갑자기 떨어지지 않도록**, B5~B12 평가 항목은 다음과 같이 조정된다:

- B1-B4 (필수 4): 미충족 시 즉시 fail
- B5-B8 (권장 4): "권장이지만 미충족이 명확한 이유 있으면 pass" — 절차가 30줄 미만으로 짧거나, dispatch가 단순한 경우 등
- B9-B12 (선택 4): "있으면 보너스, 없어도 base score 유지" — fail 처리 안 함

이는 본 가이드 §4.2 체크리스트의 보강 사항으로, evaluate-skill PROCEDURE.md Step 4가 본 트랜지션 정책을 적용한다.

### 2.2 핵심 원칙

#### 2.2.1 본문 길이 — 500줄 이하

**근거**: Skill 본문은 호출 시점에 컨텍스트에 고정되어 세션 전체에 남는다. 길수록:
- 토큰 비용 누적 (세션 전체 곱하기 컨텍스트 크기)
- **lost-in-the-middle** 현상 (중간 위치 정보의 활용도 저하 — long-context LLM의 알려진 약점)
- 다른 skill·문서와의 cross-reference 정확도 하락

**가이드**:
- 500줄 초과 → 분리 검토 필수
- 1000줄 초과 → 거의 무조건 분리

**Long-context 처리 전략** (300줄 이상이 예상되는 skill 작성 시):

| 전략 | 사용 시점 | buddy 예시 |
|------|---------|---------|
| **보조 파일 lazy-load** (가장 흔함) | 대량 예시·레퍼런스·템플릿 | `references/<topic>.md` (Skill 본문에서 `Read this if needed: [...]` 형식으로 link) |
| **본문 정보 압축** | 핵심 절차 + 표 위주 | 산문 narrative 제거, 표·체크리스트로 변환 |
| **핵심을 위/아래에 배치** | 중간 정보를 줄일 수 없을 때 | 페르소나·핵심 절차는 상단, 검증 체크리스트·anti-pattern은 하단. **중간은 LLM이 가장 약하게 처리하는 zone** |
| **단원 분리(skill 자체 분할)** | 한 skill이 여러 책임 보유 | 별도 PROCEDURE.md로 분리 후 router catalog에 등록. 본 가이드 §5.3 6-cluster 제안 참조 |
| **forking subagent** (`context: fork`) | 한 호출 결과가 매우 크나, 핵심 결론만 필요 | router/command frontmatter에 명시. PROCEDURE.md 내부엔 직접 사용 불가 |

**예외**: 한 번에 다 읽혀야 하는 절차(예: critical security audit checklist)는 길어도 OK — 단 그 경우에도 단원별 anchor와 명시적 인덱스로 LLM이 부분 참조 가능하게 작성한다.

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

판단 기준은 **총 라인 수**다 (예시 개수 × 예시별 라인). 단일 임계로 통일.

| 총 라인 합 | 위치 | 비고 |
|----------|------|------|
| ≤ 30 줄 | 본문 inline | 1-2개 예시, 각 10줄 이내가 전형 |
| 30~60 줄 | 본문 inline 가능, 분리도 OK | skill 본문 전체가 300줄 미만이면 inline 유지 권장 |
| > 60 줄 | 보조 파일 분리 | `examples.md` 또는 `examples/case-{n}.md` |

**파일명 컨벤션**: `examples.md` (단일 파일) 또는 `examples/case-{n}.md` (케이스별 분리).

**경계 영역(30~60줄) 판단 추가 기준**:
- skill 본문이 이미 300줄 초과 → 분리
- 예시 간 차이가 단순한 입력값 차이만 → 본문 (장황한 반복)
- 예시별 설명·context가 풍부함 → 분리 검토

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

### 2.4 XML 태그 사용 (선택, bonus) — A-N1

> **상태**: 본 §2.4는 **optional**. 기존 PROCEDURE.md가 XML 태그를 사용하지 않아도 evaluate-skill에서 "권장 미달"로 판정하지 않음 (보너스 점수 0, 미사용 시 감점 없음). 새 skill 작성 시 적합한 경우에만 사용.

Anthropic prompt engineering 가이드는 다음 경우 XML 태그를 권장한다:

| XML 태그 | 사용 시점 | PROCEDURE.md 적용 예 |
|---------|---------|-----------------|
| `<context>...</context>` | 배경 정보를 절차와 분리 | 호출 시점의 환경 정보를 명시적으로 묶을 때 |
| `<example>...</example>` | few-shot 예시를 구획화 | §2.2.4 few-shot 예시를 본문에 inline할 때 |
| `<output>...</output>` | LLM의 출력 영역 지정 | §2.2.6 출력 형식 schema에서 |
| `<instructions>...</instructions>` | 절차 명령을 명확히 분리 | 본문 다른 부분(설명·예시)과 절차를 시각적으로 구분 |

**buddy 정책**: 마크다운 헤더(`## 실행 절차`, `## 출력 형식`)만으로도 충분한 경우 XML 태그는 불필요. **다음 조건 만족 시에만 XML 사용 고려**:

- 본문에 같은 종류 정보가 여러 번 반복 등장 (예: 다양한 예시 → `<example>` * N)
- 마크다운 헤더 nesting이 4단계를 초과 (가독성 저하)
- 출력 schema가 복잡해서 시작·끝 경계가 모호

XML 태그는 마크다운과 자유롭게 혼용 가능. 출처: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags

### 2.5 Chain-of-Thought 패턴 (선택, bonus) — A-N2

> **상태**: 본 §2.5도 **optional**. 미사용 시 감점 없음. 새 skill 작성 시 적합한 경우에만 사용.

Anthropic CoT 가이드는 복잡한 추론이 필요한 단계에서 명시적 thinking 단계를 권장한다. PROCEDURE.md에서는 **절차 내 특정 step에서 명시적 추론을 요구할 때** 다음 패턴이 도움된다:

```markdown
### Step 3. 영향 평가

다음 정보를 **먼저 추론한 뒤** 평가를 시작하라:

<thinking>
1. 변경 대상 파일 목록을 나열한다
2. 각 파일이 다른 모듈에서 import되는지 확인한다
3. import 그래프에서 cascade 영향 범위를 추정한다
</thinking>

위 thinking 결과를 토대로 다음 표를 채워라:
| 영향 차원 | 범위 | 신뢰도 |
| ... | ... | ... |
```

**buddy 정책**: `<thinking>` 태그는 **사실상 Claude의 응답 형식**(extended thinking)에 가깝다. PROCEDURE.md 본문에 강제로 명시할 필요는 없으며, 다음 좁은 경우에만 사용:

- step에서 LLM이 **여러 가능성을 비교·검토한 뒤** 결과를 선택해야 할 때
- 출력 직전 self-verification이 필요한 step (e.g. 보안 audit)

대다수의 절차 step은 imperative 명령(§2.2.3)만으로 충분. CoT를 남발하면 본문이 길어져 lost-in-the-middle 위험 증가.

출처: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/chain-of-thought

### 2.6 한국어/영어 혼용 정책 — A-N7

> **상태**: 본 정책은 **신규 작성 skill에만 적용**. 기존 156 PROCEDURE.md는 현 상태 동결 — 평가/재작성 압력 발생 X. 점진적 정렬은 자연스러운 수정 사이클에서 발생.

#### 권장 언어 매트릭스

| 섹션·요소 | 권장 언어 | 근거 |
|---------|---------|------|
| `description` (frontmatter) | **영어 + 한국어 trigger 키워드 동반** | Anthropic dispatch가 영어 위주로 학습되었으나, 사용자가 한국어로 발화하므로 키워드 포함 (예시: 가이드 §1.2.1 Good 예시) |
| 페르소나 (`Persona` 단락) | **한국어** | buddy 사용자가 한국어 화자, 페르소나 톤이 응답 톤에 전이됨 |
| 절차 명령 (`### Step 1.`) | **한국어 명령형** | 사용자가 절차를 직접 읽을 수도 있고, LLM은 양 언어 모두 처리. 한국어 명령이 톤 일관성에 유리 |
| Input/Output Contract 표 | **한국어** (Type/Required 같은 키만 영어 fixed) | 한국어 사용자가 빠르게 파싱 |
| 출력 schema (`### 출력 형식`) | **YAML 키는 영어, 설명은 한국어** | 다른 도구·skill과 interop |
| 코드 예시·명령어 | **영어** (변경 불가) | shell 명령은 영어 |
| Anti-pattern / 검증 체크리스트 | **한국어** | 사용자 가독성 |

#### 신규 작성 체크리스트

- [ ] frontmatter `description`: 영어 본문 + 한국어 키워드 동반
- [ ] 본문 prose: 한국어 (명령형 톤)
- [ ] 표 컬럼 헤더 fixed key(Required, Type, Source)는 영어, 나머지 한국어
- [ ] YAML/JSON schema key는 영어, value 설명은 한국어
- [ ] 외부 도구 인터페이스(shell, gh, git, jq 등)는 영어 그대로

#### 기존 PROCEDURE.md는 어떻게 하나

- **점진적 정렬**: 어떤 PROCEDURE.md를 수정할 일이 생기면 그때 위 매트릭스로 정렬. 일제 변환 안 함.
- **evaluate-skill 영향 없음**: 본 §2.6은 §4 평가 체크리스트에 항목 추가 안 함 — 언어 불일치는 fail 사유 X.
- **회귀 점검**: 신규 skill만 본 매트릭스 적용 점검.

---

## §3. Persona / Role Assignment

### 3.1 효과 — 공식 입장 (요약)

아래는 Anthropic 공식 prompt engineering 가이드의 role assignment 관련 내용을 buddy 작성 표준 관점에서 **요약·재구성**한 것이다. 직접 인용이 아닌 paraphrase이며, 원문 확인은 출처 링크 참조:

> 역할 부여(role assignment)는 Claude의 출력을 task에 맞춰 정렬하는 효과적인 기법이다. 단 모호한 페르소나보다 **구체적**인 페르소나가 효과적이며, persona는 **능력 부여가 아닌 task 프레이밍 도구**임을 이해해야 한다.

**출처**:
- https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/system-prompts (System prompts와 role 지정)
- https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/be-clear-and-direct (구체성 원칙)

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
| §9 | `deprecate-feature`, `archive-product`, `migrate-customers`, `spin-off-feature` | Senior Product Manager (sunset 전문) — 사용자 영향·소통·timeline 중심으로 작성. 톤: "고객의 마이그레이션을 부드럽게 안내한다", "남은 사용자에게 무엇을 보장하는가" |

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
- [ ] **F3**: `description` + `when_to_use` 합산 1,536자 이내, 추상적 형용사 없음
- [ ] **F4**: side-effect 작업이면 `disable-model-invocation: true` 명시 (buddy 표준은 모든 command에 적용)
- [ ] **F5**: pattern library / 메타 skill은 `plugin/commands/<name>.md` **부재** 확인 (buddy 등가 매핑 — §1.2.3 참조)

### 4.1b Catalog Entry (모든 PROCEDURE.md의 paired 평가 대상 — §1.5)

PROCEDURE.md 평가 시 `plugin/skills/router/references/skill-catalog.md`의 해당 entry도 같이 평가한다.

- [ ] **CE1**: entry가 정확한 phase 표에 위치 (`engineering-phases.md` §2 정의와 일치)
- [ ] **CE2**: 호출 방법 컬럼이 3종 중 하나 + 실제 command 파일 존재/부재와 일치
- [ ] **CE3**: description에 트리거 키워드 + 목적 모두 포함, 추상 형용사 0개
- [ ] **CE4**: 같은 phase 내 다른 entry와 차별점 명확 (키워드 80% 이상 중복 X)
- [ ] **CE5**: description 길이 50-300자 권장

### 4.2 본문 구조 (모든 PROCEDURE.md)

**§2.1 3-Tier 분류 동기화** — 각 항목의 Tier에 따라 평가 정책이 다르다:

#### 🔴 필수 (4) — 미충족 시 즉시 fail

- [ ] **B1** (필수): 제목 직후 1-3줄 정체성 + 진입/종료 조건 명시
- [ ] **B2** (필수): 본문 500줄 이하 (초과 시 보조 파일 분리)
- [ ] **B3** (필수): Input Requirements 표 존재 (`engineering-phases.md` §4 형식 준수)
- [ ] **B4** (필수): Output Contract 표 존재 (`engineering-phases.md` §4 형식 준수)

#### 🟡 권장 (4) — "명확한 사유 있으면 pass" (점수 미차감)

- [ ] **B5** (권장): "이 스킬을 사용하는 경우" 섹션 명시 *(예외 사유 예: 단일 dispatch skill로 사용 case가 자명한 경우)*
- [ ] **B6** (권장): "이 스킬을 사용하지 않는 경우" 섹션 명시 (대안 skill 안내) *(예외 사유 예: 대체 skill이 없는 도메인 단독 skill)*
- [ ] **B7** (권장): 실행 절차가 imperative (Step 1, Step 2, ... 명령형) *(예외 사유 예: 30줄 미만의 단순 dispatch skill)*
- [ ] **B8** (권장): 분기 조건이 명시적 (모호한 "필요 시" 없음) *(예외 사유 예: 분기가 없는 일자 절차)*

#### 🟢 선택 (4) — 있으면 bonus, 없어도 base score 유지 (fail 처리 X)

- [ ] **B9** (선택): 동적 상태는 `` !`...` `` 문법으로 inline 주입 *(동적 상태 자체가 없는 skill은 자동 N/A)*
- [ ] **B10** (선택): 출력 형식 명시 (YAML/markdown 구조 강제)
- [ ] **B11** (선택): 검증 체크리스트 또는 anti-pattern 섹션 존재
- [ ] **B12** (선택): 다음 단계 (Next Steps) 명시

> **evaluate-skill 동기 정책**: B1-B4는 binary pass/fail. B5-B8은 미충족 시 "예외 사유 검토" 단계를 거쳐 fail/pass 결정 (LLM 판단). B9-B12는 미충족도 통과 → 분모에 포함되지만 통과 가중치 합에 영향 없음 (자동 pass). 본 정책은 evaluate-skill PROCEDURE.md Step 4-5와 1:1로 동기화되어야 한다.

### 4.3 Persona (페르소나 권장 skill 한정)

- [ ] **P1**: 본문 최상단 (제목 직후)에 페르소나 정의
- [ ] **P2**: 페르소나 구성 요소 4-5개 포함 (경력/전문영역/현재역할/관점/출력규칙)
- [ ] **P3**: 페르소나 3-7줄 길이
- [ ] **P4**: 같은 도메인 skill과 페르소나 톤 일관

### 4.4 점수 계산

#### 카테고리 가중치

| 카테고리 | 항목 수 | 가중치 |
|---------|--------|-------|
| Frontmatter | 5 (F1-F5) | 1.0 |
| Catalog Entry | 5 (CE1-CE5) | 1.0 |
| 본문 구조 | 12 (B1-B12) | 2.0 |
| Persona | 4 (P1-P4) | 1.5 |

#### 적용성 판정 (분모 결정 규칙)

각 카테고리는 **평가 대상 파일·skill의 성격에 따라 분모에 포함 여부가 달라진다**. 분모를 동적으로 계산해야 공정 비교가 가능하다.

| 카테고리 | 분모 포함 조건 | 미포함 시 |
|---------|--------------|---------|
| Frontmatter (F1-F5) | 평가 파일이 `router/SKILL.md` 또는 `plugin/commands/*.md` | PROCEDURE.md 단독 평가 시 F1-F5 분모에서 제외 |
| Catalog Entry (CE1-CE5) | 평가 대상이 `skill-catalog.md` entry를 가진 skill (router 자신은 제외) | router/SKILL.md / command.md 단독 평가 시 제외 |
| 본문 구조 (B1-B12) | **항상 포함** (PROCEDURE.md 평가 시) | command.md 단독 평가 시 일부만 (B1, B12) |
| Persona (P1-P4) | 가이드 §3.5.1 권장 매트릭스에 해당 skill이 "권장" 분류 | "비권장" 또는 "메타/dispatcher" skill은 P1-P4 분모에서 제외 |

#### 평가 케이스별 분모 (paired evaluation 통합)

**paired evaluation**(§1.6): 사용자가 `start` 같은 skill 이름을 입력하면 evaluate-skill은 다음 3 위치를 **자동 동반 평가**한다:
- `plugin/skills/start/PROCEDURE.md` (본문)
- `plugin/commands/start.md` (있으면, frontmatter)
- `skill-catalog.md`의 `start` entry (있으면)

→ 점수 계산은 **통합 분모**로 한다. 단일 위치만 평가하던 기존 C1-C4는 폐기되고 **paired 통합 케이스 PC1-PC4**로 대체.

| 케이스 | 평가 대상 조합 | Persona | 분모 합 | 계산 |
|-------|--------------|---------|--------|------|
| **PC1** | PROCEDURE.md + command.md + catalog entry 모두 존재, persona 권장 | 권장 | **40** | 5×1.0 (F) + 5×1.0 (CE) + 12×2.0 (B) + 4×1.5 (P) = 5+5+24+6 |
| **PC2** | PROCEDURE.md + command.md + catalog entry 모두 존재, persona 비권장 | 비권장 | **34** | 5×1.0 + 5×1.0 + 12×2.0 + 0 = 5+5+24 |
| **PC3** | PROCEDURE.md + catalog entry (command 없음 — pattern library), persona 권장 | 권장 | **35** | 0 + 5×1.0 + 12×2.0 + 4×1.5 = 5+24+6 |
| **PC4** | PROCEDURE.md + catalog entry (command 없음), persona 비권장 | 비권장 | **29** | 0 + 5×1.0 + 12×2.0 + 0 = 5+24 |
| **PC5** | router/SKILL.md 단독 (catalog entry 없음 — router는 카탈로그 자체) | N/A | **29** | 5×1.0 + 0 + 12×2.0 + 0 = 5+24 |

**legacy C1-C4** (단일 위치 평가, 디버깅·부분 평가용으로만 유지):

| 케이스 | 평가 대상 | Persona | 분모 합 |
|-------|---------|---------|--------|
| **C1** | router/SKILL.md or command.md 단독 | 권장 | **35** (5+24+6) |
| **C2** | router/SKILL.md or command.md 단독 | 비권장 | **29** (5+24) |
| **C3** | PROCEDURE.md 단독 | 권장 | **30** (24+6) |
| **C4** | PROCEDURE.md 단독 | 비권장 | **24** (24) |

기본은 **paired (PC1-PC5)**. 단일 위치 평가가 필요한 경우(예: 새 catalog entry만 추가 후 점검)에만 C1-C4 사용.

#### 최종 점수 공식

```
최종 점수 = (해당 케이스 통과 항목 가중치 합) / (해당 케이스 분모 합) × 100
```

**예시** (PROCEDURE.md + persona 권장, 즉 C3):
- B1-B12 중 10개 통과, P1-P4 중 3개 통과
- 통과 가중치 합 = 10×2.0 + 3×1.5 = 20 + 4.5 = 24.5
- 분모 = 30
- 최종 점수 = 24.5 / 30 × 100 = **81.67점 → 합격(80-89)**

#### Grade 기준 (공통)

| 점수 | 평가 |
|------|------|
| 90+ | 우수 — 다른 skill의 참고 모델 |
| 80-89 | 합격 — 운영 가능 |
| 70-79 | 보강 필요 — 부족 항목 즉시 개선 |
| 70 미만 | 재작성 권장 — 구조적 결함 |

> **evaluate-skill 동기**: 본 §4.4의 케이스별 분모 정의는 `plugin/skills/evaluate-skill/PROCEDURE.md` Step 7과 1:1로 동기화되어야 한다. 한쪽만 변경 시 점수 산정 결과가 어긋난다.

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
| G1 | LLM은 페르소나/구조화된 절차를 더 정확히 따른다 | 모델별로 효과 차이 있음. Opus(4.7/4.8)는 추상 절차도 능동적으로 보완, Sonnet 4.6은 명령형/체크리스트에 최적, Haiku 4.5는 깊은 추론 부담이 적은 절차에만 적합 — 모델별 행동 차이는 §1.4.3 참조. PROCEDURE.md는 모델 중립으로 작성 |
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
