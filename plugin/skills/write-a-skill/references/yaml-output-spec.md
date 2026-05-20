# YAML Output Spec — write-a-skill 출력 schema 상세

본 문서는 PROCEDURE.md §6에 명시된 YAML 출력 schema의 *필드별 required/optional 분류*와 *두 종류의 §6 yaml 경계*를 명확화한다.

---

## 1. 두 종류의 §6 yaml — 경계 명시

PROCEDURE.md §6에 등장하는 yaml은 **두 가지 다른 대상**을 가리킨다. 작업자가 자주 혼동.

### 1.1 yaml-A: *write-a-skill 자체의* 출력 템플릿

PROCEDURE.md §6 본문 "다음 yaml 구조로 결과를 호출자에게 반환한다" 직후의 큰 yaml 블록 (line ~237 이하). 이것은 *write-a-skill 호출 결과*를 호출자(상위 LLM 또는 다음 스킬)에게 돌려주는 schema.

필드: `skill_name`, `skill_type`, `lifecycle_phase`, `artifacts_created`, `description_meta`, `attribution`, `sister_skills`, `subagent_test_log`, `anti_patterns_count`, `checklist_items`, `routing_conflicts_detected`, `make_test_routing`, `next_steps`.

### 1.2 yaml-B: *작성 중인 새 스킬의* §6 출력 템플릿

PROCEDURE.md §5 Step 3의 buddy 표준 골격 안 "## 6. 출력 템플릿 (structured YAML/JSON 권장)" 항목. 이것은 *새로 작성되는 스킬의* §6에 어떤 모양의 yaml을 넣을지의 *가이드*. 새 스킬마다 schema가 다름.

가이드 원칙(yaml-B에 적용):
- structured form (free-form 산문 금지)
- 호출자가 그대로 소비 가능
- YAML 권장 (JSON·structured markdown도 OK)

### 1.3 혼동 방지 규칙

- *write-a-skill을 실행 중*이면 yaml-A schema에 결과를 채워 반환
- *새 스킬을 작성 중*이면 yaml-B 가이드를 따라 *그 스킬의* §6 yaml schema를 설계 (yaml-A schema와 무관)
- 두 yaml은 *완전 별개의 schema*. yaml-B가 yaml-A를 닮을 필요 없음.

---

## 2. yaml-A 필드별 required / optional 분류

| 필드 | 분류 | 비고 |
|------|------|------|
| `skill_name` | **required** | 모든 호출 |
| `skill_type` | **required** | 4분류 중 하나 |
| `lifecycle_phase` | **required** | §1~§9 또는 Cross-cutting |
| `artifacts_created` | **required** | 최소 1항목 (PROCEDURE.md) |
| `artifacts_created[].path` | **required** | 항목마다 |
| `artifacts_created[].role` | **required** | 항목마다 |
| `artifacts_created[].lines` | optional | PROCEDURE.md 항목에만 권장 |
| `artifacts_created[].change` | optional | modify 액션 항목에만 |
| `description_meta.text` | **required** | catalog 등재값과 동일 |
| `description_meta.length_chars` | **required** | ≤1024 검증용 |
| `description_meta.triggers_explicit` | optional | true 권장 |
| `attribution.classification` | **required** | 4분류 또는 `none` |
| `attribution.source_assets` | conditional | `classification != none`이면 required, `none`이면 omit |
| `sister_skills.predecessors` | **required** | 빈 배열 허용 (진입점) |
| `sister_skills.pairs` | optional | 없으면 빈 배열 |
| `sister_skills.successors` | **required** | 빈 배열 허용 (말단) |
| `subagent_test_log.red.scenario` | **required** | Step 2 결과 |
| `subagent_test_log.red.failure_mode` | **required** | Step 2 결과 |
| `subagent_test_log.green.iterations` | **required** | Step 3 결과 |
| `subagent_test_log.green.final_result` | **required** | pass / fail |
| `subagent_test_log.refactor_count` | **required** | 0~3 |
| `anti_patterns_count` | **required** | 5~10 (discipline-enforcing은 ≥7 권장) |
| `checklist_items` | conditional | discipline-enforcing: ≥6 필수 / 기타: 0+ (선택) |
| `routing_conflicts_detected` | **required** | true/false |
| `make_test_routing` | **required** | pass / fail / not_run |
| `next_steps` | optional | 비어도 OK |

### 2.1 conditional 필드 처리 규칙

- `attribution.source_assets`: `classification == "none"`이면 *필드 자체를 생략* (null/empty 배열 X)
- `checklist_items`: discipline-enforcing이 아니면 *값 0 또는 필드 생략* 모두 허용

### 2.2 null vs missing vs empty

- *모름·해당 없음*은 `null` (예: `make_test_routing: null` = 실행 안 함)
- *해당 안 됨*(conditional false)은 *필드 생략*
- *해당 있지만 결과 없음*은 빈 배열 `[]` (예: `sister_skills.pairs: []`)

---

## 3. yaml-A 출력 예시 (전체 채움)

```yaml
skill_name: "consult-codex-mock-mode"
skill_type: technique
lifecycle_phase: "Cross-cutting"

artifacts_created:
  - path: "plugin/skills/consult-codex-mock-mode/PROCEDURE.md"
    role: primary
    lines: 312
  - path: "plugin/skills/router/references/skill-catalog.md"
    role: catalog
    change: "added 1 line in Cross-cutting Utilities table"
  - path: "plugin/skills/router/references/routing-rules.md"
    role: routing
    change: "no change (no conflict)"
  - path: "NOTICE"
    role: attribution
    change: "no change (no external asset)"

description_meta:
  text: "consult-codex를 외부 API 호출 없이 fixture 응답으로 시뮬레이션. 테스트 환경·CI에서 codex CLI 의존성 없이 second opinion mock."
  length_chars: 124
  triggers_explicit: true

attribution:
  classification: none
  # source_assets 필드 자체 생략 (classification: none)

sister_skills:
  predecessors:
    - consult-codex
  pairs: []
  successors: []

subagent_test_log:
  red:
    scenario: "테스트 환경에서 mock 응답을 어떻게 결정하는가"
    failure_mode: "subagent가 외부 API를 시도하여 mock 의도 위반"
  green:
    iterations: 1
    final_result: pass
  refactor_count: 1

anti_patterns_count: 6
# checklist_items 필드 생략 (technique type, 선택)
routing_conflicts_detected: false
make_test_routing: pass

next_steps:
  - "consult-codex 자체에 mock-mode 옵션 흡수 검토"
```

---

## 4. yaml-B 가이드 (새 스킬 §6 설계)

새 스킬의 §6에 들어갈 yaml은 *그 스킬 고유 schema*. 다음 *형식 규칙*만 buddy 표준:

1. **호출자 소비 가능**: 다음 스킬이 *grep·jq·yaml parser*로 한 번에 추출 가능한 구조
2. **모든 필드 명시적**: 추정 가능한 default도 명시 (`null` 또는 빈 배열로)
3. **자유 서술 금지**: notes·comments·remarks 같은 free-form 필드는 *최소화*
4. **타입 일관**: 같은 의미 필드는 같은 타입 (예: timestamp는 ISO8601, count는 integer)

### 4.1 yaml-B 골격 예시 (brainstorm 스킬 가정)

```yaml
skill_name: "brainstorm"
status: completed | partial | aborted

# brainstorm 고유 출력
forcing_questions_answered:
  - question: "<...>"
    answer: "<...>"
    confidence: 0-1

candidate_actions:
  - action: "<concrete next step>"
    rationale: "<why>"
    cost_estimate: low | mid | high

ranked_recommendation:
  primary: "<action id>"
  fallback: "<action id>"

aborted_paths:
  - reason: "<...>"
    path_attempted: "<...>"

next_skill_dispatch:
  suggested: "<skill-name>" | null
```

이 schema는 write-a-skill의 yaml-A와 완전 다름 — 그게 정상.

---

## 5. yaml 작성 검증

작성된 yaml이 spec을 따르는지 *기계적 검증* 권장:

```bash
# yamllint으로 syntax 검증
yamllint plugin/skills/<name>/sample-output.yaml

# jq로 필수 필드 존재 확인 (yaml → json 변환 후)
yq -o json sample.yaml | jq -e '.skill_name and .skill_type and .lifecycle_phase'
```

CI 통합 시 `make test-routing`과 별개로 `make test-yaml-schema` 추가 권장.
