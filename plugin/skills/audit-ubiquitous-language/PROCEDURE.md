# Audit Ubiquitous Language — 코드 식별자 / PRD 어휘 / 도메인 비즈니스 어휘 3자 일관성 검사

당신은 **buddy 의 도메인 어휘 일관성 감독관** 이다. 코드의 식별자 (변수 / 함수 / 클래스 / 모듈), PRD 의 명시 어휘, 도메인 비즈니스 용어 — 이 *3 source* 가 동일 개념을 *동일 어휘* 로 부르고 있는지 audit 하고, mismatch 를 *제안* 으로 출력한다. 본 스킬은 *분석 + 제안* — rename 실행은 위임.

핵심 차별점:
- `refactor-with-rename-trace` 가 *rename 의 실행 + 추적* 이라면, 본 스킬은 *rename 결정 전 단계의 어휘 일관성 검사*. 본 스킬의 출력 → `refactor-with-rename-trace` 입력.
- `review-architecture` 가 *구조적 무결성* 검토라면, 본 스킬은 *어휘적 무결성* 검토. 직교 차원.
- `measure-code-health` 가 *style / lint / coverage 같은 정량 지표* 라면, 본 스킬은 *의미 (어휘) 일관성 — 정성 + 정량 혼합*.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 코드베이스 + PRD/도메인 어휘 | ✅ | artifact | 현재 코드베이스 + 도메인 문서 | (자동 감지) |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Vocabulary drift report (mismatch pairs + severity + remediation) | artifact | structured YAML | `refactor-with-rename-trace` |

---

## 1. 목적

코드베이스의 *ubiquitous language* (Evans 2003, Vernon 2013 의 DDD 개념) 일관성을 *3 source 기준* 으로 검사하고, drift 발견 시 *remediation 제안* 을 출력한다.

3 source 정의:
- **코드 식별자** — 변수 / 함수 / 클래스 / 모듈 / 파일 / 디렉토리 이름 (production code + test fixture + mock + 주석)
- **PRD 어휘** — `docs/prd.md`, `docs/feature-spec/*`, `docs/features.yaml`, ADR 본문, design doc 등 *문서화된 비즈니스 의도*
- **도메인 비즈니스 어휘** — 사용자 / 도메인 전문가가 *실제로 사용하는* 용어 (인터뷰 / Slack 채널 / 고객 support ticket 등에 등장)

3 source 가 *동일 개념* 을 *동일 어휘* 로 부르면 ✅ — 그렇지 않으면 drift. 본 스킬은 drift 를 *제안* 으로 출력하되 *실행은 안 함*.

본 스킬 산출물:
- `drift_findings` 리스트 (mismatch pair + 파일:라인 + severity + remediation 제안)
- `health_score` (drift rate + consistency score + confidence)
- `glossary_updates` (필요 시 `docs/glossary.md` 생성 / 갱신)
- 후속 스킬 dispatch 제안 (`refactor-with-rename-trace` / `update-docs-with-code` / `write-adr`)

---

## 2. 사용 시점

다음 상황에서 *반드시* 호출:

- **refactor 직전** — `refactor-with-rename-trace` 호출 전 *어휘 결정* 단계. rename 대상이 *진짜 drift 인지 / 의도된 별칭인지* 결정.
- **PR review 안** — *신규 식별자* 추가된 PR 에서 PRD 어휘 정합 확인. `review-engineering` cascade.
- **신규 feature 정의 후** — `define-features` → `define-feature-spec` 종료 직후 *PRD 어휘 ↔ 기존 코드 어휘* drift 확인.
- **주기적** (monthly / quarterly / release 직전) — drift 는 시간 경과로 누적. 단일 호출만으로 부족.

호출하지 마라:

- **style / lint / convention 검사** (camelCase vs snake_case 같은) — `measure-code-health` 위임.
- **rename 실행** — 본 스킬은 *제안* 만. 실행은 `refactor-with-rename-trace`.
- **단일 변수 1건 rename 확인** — overkill. 단순 grep 으로 충분.
- **i18n 번역 검사** — `audit-i18n-coverage` 위임 (단, *기술 식별자의 다국어 정합* 은 본 스킬 scope).

---

## 3. 입력

### 필수 결정 (Step 1 종료 전까지 모두 확정)

| 결정 | 수집 경로 | 기본값 |
|------|----------|--------|
| **audit_scope.code_paths** | command arg 또는 forcing question | repository root |
| **audit_scope.prd_sources** | command arg 또는 forcing question | `docs/prd.md`, `docs/feature-spec/`, `docs/features.yaml` (존재 시) |
| **audit_scope.bounded_contexts** | forcing question | 단일 context — multi-context 시 명시 |
| **audit_scope.language** | 자동 감지 | primary code language (`go.mod` / `package.json` / `pyproject.toml` 등) |
| **audit_scope.excluded** | 선택 | `vendor/`, `node_modules/`, generated code, `.draft` 파일 |
| **bounded_context 어휘 source** | forcing question | (없으면) PRD + 코드 만 audit. 도메인 인터뷰 / Slack archive 등 *3 번째 source* 가 있으면 명시 |

### 선택 입력

- **기준 Glossary** — `docs/glossary.md` 존재 시 *canonical 어휘* baseline. 없으면 audit 결과로 *신규 생성 제안*.
- **이전 audit 결과** — drift 변화 추이 측정 시.
- **무시 어휘 리스트** — 의도된 별칭 (예: legacy API 호환성 위한 oldName) — false positive 차단.

### 입력 부족 시 forcing question

다음 중 하나라도 답할 수 없으면 호출자에게 되묻는다:

- "audit 할 *bounded context* 는 단일이야 multi 야? multi 면 각 context 별 어휘 분리 검사 필요."
- "PRD 어휘 source 는 어디? `docs/prd.md` 만? 아니면 feature-spec / ADR / design doc 까지?"
- "*의도된 별칭* 이 있어? (legacy 호환 / 외부 API 어휘) — 없으면 모든 mismatch 가 drift 후보."
- "이전 audit 결과 있어? 있으면 *변화 추이* 측정 가능. 없으면 단일 시점 snapshot."

---

## 4. 핵심 원칙

1. **3 source 모두 cross-reference 의무** — *코드만* / *PRD 만* 검사 금지. 3 source (code / PRD / 도메인) 가 *모두* 입력. 도메인 source 부재 시 *2 source* 명시 + low_confidence true.
2. **분석만, 실행 금지** — 본 스킬은 *제안 출력* 만. 자동 rename / 자동 PRD 수정 *0건* 금지. 실행은 `refactor-with-rename-trace` / `update-docs-with-code` 위임.
3. **Test / fixture / mock / 주석 포함** — production code 만 audit 금지. test fixture 의 oldName 도 drift. 주석의 stale term 도 drift.
4. **언어 / i18n 인지** — 한국어 / 일본어 / 영어 mixed 코드베이스에서 *번역 layer* 식별. `userId` (영어 식별자) ↔ `사용자_식별자` (PRD 한국어) 의 *번역 일관성* 검사 — *단 본 스킬은 어휘 매핑* 만, *번역 누락 검사* 는 `audit-i18n-coverage` 위임.
5. **Bounded context 분리** — multi-context (예: billing / auth / catalog) 시 *각 context 안* 의 일관성 만 검사. context 간 동의어 (auth 의 `user` vs catalog 의 `customer`) 는 *의도된 경계* 일 수 있음 — drift 아님.
6. **Confidence 측정** — drift 식별 정확도가 낮으면 (예: NLP 의미 매핑 < 80% confidence) *low_confidence true* + 사용자 확인 요청. AI 자율 판정 금지.
7. **Glossary SSoT** — `docs/glossary.md` 존재 시 *canonical*. 없으면 audit 결과로 *생성 제안*. Glossary 갱신은 본 스킬의 *3차 산출물*.
8. **AI 합리화 차단** — "이 mismatch 는 의도된 것 같음" 같은 *self-rationalization* 금지. 모든 mismatch 를 *출력*, 사용자가 *명시 무시* 결정.
9. **Severity 분류 강제** — drift 마다 high / medium / low 강제 분류. ambiguous 미사용. high = production code path + 외부 API surface, medium = internal API, low = test / 주석.
10. **Iron Law (verification-discipline)** — audit 종료 발화 직전 *fresh evidence* 확인. `drift_findings: []` 출력 시 *실제 0 mismatch* vs *측정 누락* 구분 강제.

---

## 5. 실행 단계 (Steps)

### Step 1. Scope Lock-in

- audit_scope 6 필드 (`code_paths` / `prd_sources` / `bounded_contexts` / `language` / `excluded` / 도메인 source) 모두 확정
- *bounded context 가 multi* 면 각 context 별 audit 분리 (이하 Step 2-5 를 context 별 반복)
- 도메인 source 부재 시 *2 source mode* 명시 + low_confidence 기본값 true

### Step 2. Code Identifier Extraction

- `code_paths` 안 *모든* 식별자 추출 (변수 / 함수 / 클래스 / 모듈 / 파일명 / 디렉토리명 / 주석 안 명사구)
- production + test + fixture + mock + 주석 모두 포함
- `excluded` paths 제외
- LSP 또는 정적 분석 도구 사용 권장 (Go: `gopls symbols`, Python: `jedi`, JS/TS: `tsserver`)
- 결과: 식별자 frequency 표 + 위치 매핑

### Step 3. PRD / Domain Vocabulary Extraction

- `prd_sources` 안 *모든 명사구* 추출
- 도메인 source (인터뷰 transcript / Slack archive / support ticket 등) 가 있으면 *동일 추출*
- 결과: canonical 어휘 후보 표

### Step 4. Cross-Reference Matching

- *동일 개념* 의 어휘 pair 매핑 — 코드 식별자 ↔ PRD 어휘 ↔ 도메인 어휘
- 매핑 알고리즘:
  - 정확 일치 (`userId` ↔ `user id` ↔ `사용자 ID`) → ✅
  - 동의어 (`user` ↔ `customer` ↔ `고객`) → 의도된 별칭 검토
  - 오타 / 약어 (`usr` ↔ `user`) → drift high
  - 번역 mismatch (`memberId` ↔ `사용자 ID`) → drift medium
  - stale legacy (`oldUserId` 코드 잔존 + PRD 는 `userId`) → drift high
- 매칭 confidence < 80% 시 *low_confidence true* + 해당 pair 사용자 확인 요청

### Step 5. Drift Classification + Remediation 제안

각 drift 별로:
- `mismatch_type` 강제 분류 — `synonym` / `typo` / `abbreviation` / `translation` / `stale_legacy` 중 하나
- `severity` 강제 — `high` (production + external API) / `medium` (internal API) / `low` (test / 주석)
- `remediation.suggested_action` 강제 — `rename` (코드 변경 권장) / `update_prd` (PRD 변경 권장) / `add_glossary` (어휘 정의 명시 권장) / `accept` (의도된 별칭 인정)
- `remediation.delegated_to` 강제 — `refactor-with-rename-trace` (rename 실행) / `manual` (사용자) / 다른 skill

### Step 6. Health Score 계산

- `drift_rate` = mismatch / total identifier (0.0-1.0)
- `consistency_score` = `10 * (1 - drift_rate)` 또는 weighted by severity (high=3, medium=2, low=1 weight)
- `confidence` = Step 4 의 평균 매칭 confidence
- `low_confidence` = `confidence < 0.8`

### Step 7. Glossary 갱신 또는 생성

- `docs/glossary.md` 존재 시 *canonical 어휘 추가 / 갱신 / deprecation 마크 제안*
- 부재 시 *신규 생성 제안* — audit 결과의 canonical 어휘 표를 baseline 으로
- *실 작성은 `update-docs-with-code` 위임*

### Step 8. 출력 + 후속 skill dispatch 제안

- §6 yaml output 작성
- `next_steps` 에 후속 skill 명시:
  - drift high count > 0 → `refactor-with-rename-trace`
  - PRD update 권장 → `update-docs-with-code`
  - 어휘 결정 ADR 필요 (예: 동의어 선택) → `write-adr`
  - 주기적 audit 권장 → `setup-quality-gates` (pre-commit hook 추가)

---

## 6. 출력 템플릿

```yaml
skill_name: audit-ubiquitous-language
status: completed | partial | aborted

audit_scope:
  language: "<primary code language>"
  bounded_contexts: ["<context-1>", "<context-2>"]
  prd_sources: ["<path-1>", "<path-2>"]
  code_paths: ["<path-1>", "<path-2>"]
  excluded: ["<path-1>"]
  domain_source_available: true | false

drift_findings:
  - vocabulary_pair:
      domain_term: "<canonical PRD term>"
      code_identifier: "<actual code symbol>"
      file: "<path:line>"
      mismatch_type: synonym | typo | abbreviation | translation | stale_legacy
      severity: high | medium | low
      remediation:
        suggested_action: rename | update_prd | add_glossary | accept
        suggested_target: "<new identifier or PRD update text>"
        delegated_to: refactor-with-rename-trace | update-docs-with-code | manual

health_score:
  drift_rate: <0.0-1.0>
  consistency_score: <0-10>
  confidence: <0.0-1.0>
  low_confidence: true | false

glossary_updates:
  added: ["<term>"]
  modified: ["<term>"]
  deprecated: ["<term>"]
  target_file: "docs/glossary.md"
  action: create | update | skip

routing_conflicts_detected: false

next_steps:
  - "<follow-up skill or action>"
```

---

## 7. 자매 스킬

### 선행 (input 공급)

- `define-features` / `define-feature-spec` — PRD 어휘 source 작성 후 본 스킬 호출
- `identify-actors` + `compose-feature-from-use-cases` — actor / use case 어휘 baseline
- `concretize-idea` — 초기 PRD 의 *비즈니스 어휘* 설정
- `review-architecture` — 구조 검토 안 *어휘 일관성 dimension* cross-link

### 페어 (동시 동작)

- `freeze-edit-scope` — audit 도중 코드 *수정 차단* (audit 결과 오염 방지)
- `consult-codex` — 어휘 결정 모호 시 second opinion (의도된 동의어 vs drift 판정)

### 후속 (output 소비)

- `refactor-with-rename-trace` — drift high count > 0 시 rename 실행 (delegated_to: refactor-with-rename-trace)
- `update-docs-with-code` — Glossary 갱신 + PRD 어휘 update
- `write-adr` — 동의어 선택 / context 분리 결정 영속화
- `setup-quality-gates` — 주기적 audit 자동화 (pre-commit hook)

### 호출 흐름 예시

```
define-features (PRD 어휘 설정)
    → audit-ubiquitous-language (drift 검사)
        → refactor-with-rename-trace (high drift rename 실행)
        → update-docs-with-code (Glossary 갱신)
        → write-adr (어휘 결정 영속화, optional)
    → (주기적 trigger 시) setup-quality-gates (audit 자동화)
```

---

## 8. Anti-patterns

1. **코드 grep 만으로 audit 종료** — PRD / 도메인 source cross-reference 누락. 코드 일관성만 cover, *3 source* 미충족. 교정: §5 Step 3 의 PRD/도메인 추출 *반드시* 진입. 누락 시 status = partial.

2. **audit 종료 시점에 자동 rename 시도** — 분석 ↔ 실행 책임 경계 침범. 본 스킬 scope creep. 교정: 본 스킬은 *제안* 만 출력 (yaml `remediation.suggested_action`). 실행은 `refactor-with-rename-trace` 위임.

3. **test fixture / mock / 주석 skip** — production code 만 audit. test 의 oldName / 주석의 stale term 누락. 교정: §5 Step 2 의 audit scope 에 test / mock / 주석 *명시 포함*.

4. **단일 PRD / 단일 context 가정** — multi-PRD 또는 multi-bounded-context 코드베이스에서 *context 간 동의어를 drift 로 오판*. 교정: §3 입력에서 `bounded_contexts` 강제 확정. multi 시 *각 context 별 audit 분리*.

5. **drift_rate 0 시점에 종료** — *실제 0 mismatch* vs *측정 누락* 구분 없이 종료. 교정: §4 의 *Iron Law* — confidence < 0.8 시 *low_confidence true* + 사용자 확인 강제. 무조건 종료 금지.

6. **단일 언어 가정 (영어)** — i18n / 한국어 / 일본어 mixed 코드베이스에서 *번역 layer* 무시. `userId` ↔ `사용자_식별자` 같은 *번역 mismatch* 누락. 교정: §5 Step 1 의 *언어 식별* 명시. 단 *번역 누락 검사 자체* 는 `audit-i18n-coverage` 위임.

7. **AI 자율 합리화** — "이 mismatch 는 의도된 별칭 같음, skip" 같은 *self-judgment*. 교정: 모든 mismatch 를 *출력* 강제. *사용자 명시 무시* 만 인정. AI 가 자율 skip = 합리화 패턴.

8. **Style 검사 scope creep** — naming convention (camelCase / snake_case) 같은 *스타일* 까지 검사. 본 스킬은 *어휘 일관성* 만. 교정: 스타일은 `measure-code-health` / linter 위임. 본 스킬은 *어휘 의미 일관성* 한정.

9. **Glossary 갱신 누락** — audit 결과 출력만 하고 *Glossary 영속화* 누락 → 다음 audit 시 *동일 finding* 재발견. 교정: §5 Step 7 의 Glossary 생성 / 갱신 *반드시* 제안. 영속화는 `update-docs-with-code` 위임.

10. **Single-pass audit 종료 선언** — 한 번 실행 후 *완료* 선언. drift 는 시간 경과로 누적 — 주기 trigger 부재 시 *재발생* 보장. 교정: §10 종료 조건의 *주기 권장* (monthly / PR / release 직전). `setup-quality-gates` cross-link.

---

## 9. 체크리스트 (Step 별 자가 점검 — discipline-enforcing 필수)

각 Step 종료 시:

- [ ] Step 1 `audit_scope` 6 필드 모두 확정 (code_paths / prd_sources / bounded_contexts / language / excluded / domain source 가용성)
- [ ] Step 2 식별자 추출 시 production + test + fixture + mock + 주석 *모두* 포함
- [ ] Step 3 PRD source 가 *모든* 명시 source 포함 (`docs/prd.md` 외 ADR / design doc / feature-spec 등)
- [ ] Step 4 매칭 confidence < 0.8 시 *low_confidence true* + 사용자 확인 요청
- [ ] Step 5 모든 drift 가 `mismatch_type` + `severity` + `remediation.suggested_action` + `delegated_to` *4 필드 강제 분류*
- [ ] Step 5 의 `remediation.delegated_to` 가 *실 존재 skill* 또는 `manual` (자체 실행 금지)
- [ ] Step 6 `health_score` 의 `drift_rate` / `consistency_score` / `confidence` 모두 정량값
- [ ] Step 7 Glossary 갱신 / 생성 / skip 결정 (target_file 명시)
- [ ] Step 8 `next_steps` 에 *후속 skill* 또는 *manual action* 명시 (빈 리스트 금지 단 status=completed 시)
- [ ] §6 yaml output 의 모든 필수 필드 채움
- [ ] Iron Law — *fresh measurement evidence* 없이 `drift_findings: []` 출력 금지

11개 중 하나라도 No 이면 종료 선언 금지.

---

## 10. 종료 조건

다음을 *모두* 충족해야 호출 사이클 종료:

- §9 체크리스트 11개 모두 ✅
- `drift_findings` 가 빈 리스트 ([]) 인 경우 *low_confidence false* 이거나 *사용자 명시 확인* (Iron Law)
- `glossary_updates.action` 이 create / update / skip 중 하나 명시
- `next_steps` 가 빈 리스트 아님 (status = completed 시) 또는 status = partial / aborted (이유 명시)
- §6 yaml 호출자에게 반환

종료 후 호출자는 `next_steps` 의 후속 skill 자동 또는 수동 dispatch.

**주기적 호출 권장**: 본 스킬은 *single-pass* 가 아닌 *recurring*. drift 는 시간 경과로 누적되므로 monthly / quarterly / release 직전 호출 권장. 자동화는 `setup-quality-gates` 의 pre-commit hook 또는 CI job 으로.

---

## 11. References

- DDD ubiquitous language 원천: Eric Evans, *Domain-Driven Design* (2003), Chapter 2
- DDD 실용 적용: Vaughn Vernon, *Implementing Domain-Driven Design* (2013), Chapter 3
- 본 스킬은 *inspired-by* (ADR-003 분류) — DDD 이론 차용, buddy 도메인 자체 구성. mattpocock `ubiquitous-language` 본문 *0 read* (cross-machine 부재). verbatim 0건 유지.
- buddy cross-skill SSoT: [`router/references/verification-discipline.md`](../router/references/verification-discipline.md) — Iron Law (§4 원칙 10 적용)
- buddy 9-phase: [`router/references/skill-catalog.md`](../router/references/skill-catalog.md) §6 Quality
