# Attribution Classification — 차용 4분류 경계 결정 가이드

본 문서는 ADR-003 §2.4의 4분류 — `verbatim` / `adopt-with-edits` / `reference-only` / `inspired-by` — 사이 경계를 *케이스 예시*로 구체화한다. PROCEDURE.md §4 핵심 원칙 #5와 Step 8 표가 *기준*을, 본 문서가 *경계 판단*을 다룬다.

---

## 1. 결정 트리

```
외부 자산을 봤거나 만졌는가?
├─ NO → classification: none
└─ YES
   └─ 본문/코드를 *글자 그대로* 1줄 이상 옮겼는가?
      ├─ YES → verbatim (금지 — ADR-003 §2.4 정책 위반)
      └─ NO
         └─ 외부 자산의 *구조·골격·순서*를 따랐는가?
            ├─ YES → adopt-with-edits
            └─ NO
               └─ 외부 자산을 *이름·개념으로 언급*만 했는가?
                  ├─ YES → reference-only
                  └─ NO (디자인 결정에 영향만) → inspired-by
```

### 1.1 결정 기준 핵심

- **글자 그대로 옮김 = verbatim**: 표현·구두점·예시 텍스트 동일. 5단어 이상 일치하면 verbatim 의심.
- **구조 차용 = adopt-with-edits**: 섹션 구성, Step 순서, 표의 분류축, 의사결정 트리 같은 *형식적 골격*을 따름. 본문 표현은 본인이 다시 씀.
- **이름·개념 언급 = reference-only**: 외부 자산을 "X 패턴" 식으로 *호명*하지만 본문은 완전 독립.
- **디자인 영향 = inspired-by**: 외부 자산의 *방향성·철학*이 결정에 영향. 호명도 안 함.

---

## 2. 경계 케이스 예시

### 2.1 verbatim ↔ adopt-with-edits 경계

**상황**: superpowers `writing-skills`의 "Bulletproofing Skills Against Rationalization" 섹션 4 항목(Close Every Loophole / Address Spirit vs Letter / Rationalization Table / Red Flags List)을 본인 스킬의 §8 anti-patterns 보조 구조로 채택.

| 행위 | 분류 |
|------|------|
| 4 항목 이름·설명을 그대로 옮김 | **verbatim** (금지) |
| 4 항목 이름은 유지, 설명·예시는 본인 voice로 재작성 | **adopt-with-edits** |
| 4 항목 *개수와 의도*는 유지, 이름도 본인 작명 (예: "1. 회피 빌미 차단 / 2. 정신 vs 문구 / 3. 합리화 표 / 4. 적신호 목록") | **adopt-with-edits** (구조 차용이므로) |
| 본문 §8에 "superpowers writing-skills의 anti-rationalization 패턴 참조"만 언급, 본인 항목은 독립 | **reference-only** |

**판정 핵심**: *4 항목 구조* 자체가 외부 통찰. 이걸 따른 것 만으로 adopt-with-edits. 본문 표현 차이가 아닌 *형식적 골격*이 기준.

### 2.2 adopt-with-edits ↔ reference-only 경계

**상황**: mattpocock `write-a-skill`의 description 작성 규칙("Max 1024 chars / 3인칭 / 'Use when' 절")을 본인 PROCEDURE.md §4에 채택.

| 행위 | 분류 |
|------|------|
| 규칙 3개를 본인 §4에 명문화 | **adopt-with-edits** |
| 본문에 "mattpocock의 description 작성 규칙 적용"만 언급 | **reference-only** |
| 규칙은 사용하지만 출처 명시 없음 | (attribution 누락 — ADR-003 §2.3 위반) |

**판정 핵심**: 외부 *규칙·기준·결정 트리* 자체가 가치 있는 통찰이면 adopt-with-edits. 단순 인용은 reference-only.

### 2.3 reference-only ↔ inspired-by 경계

**상황**: Anthropic skill-creator의 "Progressive Disclosure" 개념을 본인 스킬에 적용.

| 행위 | 분류 |
|------|------|
| 본문에 "Anthropic skill-creator의 progressive disclosure 원칙 적용" 명시 | **reference-only** |
| 본문에 출처 언급 없이 개념(*본문 < 400줄, 상세는 references/*)만 적용 | **inspired-by** |
| 본문에 출처 + 개념 *정의*까지 옮김 | **adopt-with-edits** (개념 정의 자체가 외부 통찰) |

**판정 핵심**: *호명 여부*가 reference-only / inspired-by 경계. 단 호명 없이 영향 받은 게 입증 가능하면 inspired-by 명시 (정직성).

### 2.4 inspired-by ↔ none 경계

**상황**: 작업 중 superpowers TDD 글을 읽었지만 본인 스킬은 *완전히 다른 도메인*.

| 행위 | 분류 |
|------|------|
| 본인 스킬의 디자인 결정에 *식별 가능한 영향* (예: RED-GREEN-REFACTOR 사이클 적용) | **inspired-by** |
| 외부 자산이 *컨텍스트로 존재했지만* 본인 결정에 직접 영향 없음 | **none** |

**판정 핵심**: 외부 자산의 *식별 가능한 영향*이 본인 결정에 있는가. 단순히 *알고 있는 상태*는 none.

---

## 3. 흔한 오판

### 3.1 "구조도 안 따르고 본문도 다시 썼으니 inspired-by 아닌가?"

→ 종종 구조 차용을 본인이 인식 못 함. 예: "사이클·Step·체크리스트"가 외부 자산 패턴이면 그것 자체가 구조 차용 = adopt-with-edits.

**진단 질문**: 외부 자산을 안 봤어도 본인이 같은 구조를 떠올렸을까? Yes면 inspired-by 가능, No면 adopt-with-edits.

### 3.2 "여러 외부 자산을 섞었으니 verbatim 아니다"

→ 출처별 verbatim 여부를 *개별로* 판정. 한 출처에서 verbatim 1줄이라도 있으면 verbatim 분류. 섞음은 verbatim 회피 못 함.

### 3.3 "공통 지식이라 attribution 면제"

→ "TDD"·"Red-Green-Refactor" 같은 *업계 표준 용어*는 공통 지식. 하지만 *특정 인용 가능한 출처의 가공된 표현*(예: superpowers의 "Iron Law" 작명)은 attribution 의무 있음.

---

## 4. classification 별 NOTICE block 템플릿

### 4.1 adopt-with-edits

```
<upstream-project>
Copyright (c) YYYY <author>
<License — full text>

Pattern adoption per ADR-003 §2.4 (adopt-with-edits):
- plugin/skills/<name>/PROCEDURE.md adopts <specific pattern>
  from upstream's <specific skill/section>. Structure/principle
  borrowed; body 100% independently authored in buddy voice.

No verbatim code or text was adopted.
```

### 4.2 reference-only

```
<upstream-project> (reference-only per ADR-003 §2.4)
Source: <URL or path>
License: <SPDX or terms>

Referenced by plugin/skills/<name>/PROCEDURE.md for the
<specific concept>. No code, prompt text, or asset adopted;
buddy implementation is independently authored.
```

### 4.3 inspired-by

```
<upstream-project> (inspired-by per ADR-003 §2.4)
Source: <URL or path>

Influence on plugin/skills/<name>/PROCEDURE.md's design decisions
(e.g., <specific decision>). No direct adoption or citation.
```

### 4.4 verbatim — 금지

verbatim block은 정책상 작성 불가. verbatim adoption이 발견되면 *즉시 본문 재작성* 후 분류 재결정.

---

## 5. 판정자

분류 결정은 *write-a-skill 호출자 본인*이 한다. 호출자가 외부 자산 접촉 사실을 가장 잘 안다. 모호한 경우 `consult-codex`로 second opinion. NOTICE 작성 후에는 PR 리뷰어가 1차 검증, 라이센스 위반 의심 시 외부 자문.
