---
name: define-product-spec
description: "아이디어 검증과 사업성 검증 결과를 공식 PRD(Product Requirements Document)로 고정. Problem/Solution/User Stories/Success Criteria/Implementation Decisions/Out of Scope/feature 후보를 명문화. 트리거: 'PRD 작성' / '공식 명세 만들어줘' / 'PRD로 고정' / 'scope 확정해줘' / '유저스토리 작성' / '성공 기준 정의' / 'feature 후보 추출'. 입력: validate-idea/assess-business-viability/review-scope 통과 산출물. 출력: PRD Markdown + feature 후보 목록. 흐름: validate-idea → assess-business-viability → define-product-spec → split-work-into-features."
type: skill
---

# Define Product Spec — PRD 고정 인터뷰

## 1. 목적

`review-scope`가 scope를 형성하고 `critique-plan`이 plan을 비평한다면, **`define-product-spec`은 PRD로 고정**한다.

논의된 모든 결정을 **단일 권위 문서(Single Source of Truth)**로 변환해, agent와 human 개발자가 혼선 없이 작업할 수 있게 한다. 이후 `split-work-into-features`가 PRD를 받아 feature로 분해한다.

이 스킬을 통과한 PRD는 다음 두 보장을 가진다:
1. **측정 가능한 성공 기준** — 숫자 또는 예/아니오로 답 가능
2. **Vertical Slice 준비 완료** — 각 요구사항이 독립 feature로 분해 가능한 형태

## 2. 사용 시점 (When to invoke)

- `validate-idea` + `assess-business-viability` go verdict 후 첫 PRD 작성
- 기존 PRD에 신규 epic 추가
- pivot 결정 후 PRD 재작성
- 외부 stakeholder(투자자, 파트너)에게 공식 명세 제출 전
- 신규 팀원 onboarding을 위한 baseline 문서
- agent가 구현 시작 전 acceptance criteria 확보 필요

## 3. 입력 (Inputs)

### 필수
- `validate-idea` 통과 산출물 (problem, target user, value hypothesis)
- `assess-business-viability` go verdict (verdict reasoning + critical assumptions)
- `review-scope` 산출물 (in-scope / out-of-scope 결정)
- 1-2명의 stakeholder 의사결정 권한자

### 선택
- 경쟁사 분석 (positioning 결정 시)
- 디자인 wireframe / mockup
- 기존 PRD (수정 시)
- 사용자 인터뷰 transcript

### 입력 부족 시 forcing question
- "성공 기준이 측정 가능해? '빠르다'가 아니라 'LCP 2.5초 이내' 형태로."
- "이 user story의 acceptance criteria가 뭐야? 어떻게 done인지 판단해?"
- "out-of-scope를 명시했어? 명시 안 하면 scope crawl 시작."
- "non-functional requirement(보안/성능/가용성)는 빠뜨리지 않았어?"

## 4. 핵심 원칙 (Principles)

1. **기록의 권위** — "대화에서 그렇게 정했다" < PRD에 기록된 내용. 변경은 PRD 먼저 업데이트.
2. **측정 가능성** — 성공 기준은 숫자 또는 예/아니오 답. "괜찮다", "빠르다" 금지.
3. **Vertical slice 준비** — 각 요구사항이 schema → API → UI → test 관통 가능한 단위.
4. **Out-of-scope 명시** — scope crawl은 가장 흔한 실패. 명시적 제외 목록 필수.
5. **Acceptance criteria 모든 user story에** — "X일 때 Y한다" 형식. 자동 검증 가능 우선.
6. **Implementation decisions 보존** — validate 단계에서 확정된 trade-off는 반복 논의 방지.
7. **Stakeholder 1명 명시** — PRD 충돌 시 누가 결정하는가. ambiguity 방지.
8. **Living document, not contract** — 학습으로 업데이트. 단, 변경 시 변경 이유 기록.

## 5. 단계 (Phases)

### Phase 1. Context Capture
이전 단계 산출물에서 자동 추출:
- problem statement (validate-idea)
- target user / buyer (assess-business-viability)
- in-scope / out-of-scope (review-scope)
- pricing model (assess-business-viability)
- critical assumptions (validate-idea + assess-business-viability)

### Phase 2. Solution Articulation
1. proposed solution 한 단락 (기술적 + 비즈니스적)
2. 핵심 user flow 3-5개 (text 또는 diagram)
3. 핵심 차별점 (vs 경쟁사 / status quo)

### Phase 3. User Stories
"As a <persona>, I want <action>, so that <value>" 형식.
각 story에:
- acceptance criteria 2-5개
- priority (P0 / P1 / P2)
- dependencies (다른 story 또는 기술 prerequisite)

### Phase 4. Success Criteria (KPI)
4 카테고리로:
- **사용자 행동**: activation rate, retention, engagement
- **비즈니스 지표**: ARR, conversion, churn
- **품질 기준**: error rate, performance(LCP/FID/CLS), test coverage
- **운영 기준**: uptime, support response time, incident rate

각 KPI는 baseline + target + 측정 방법 명시.

### Phase 5. Functional & Non-Functional Requirements
- functional: 기능 동작 명세
- non-functional: 성능 / 보안 / 가용성 / 접근성 / 국제화

### Phase 6. Implementation Decisions
validate / scope 단계에서 확정된 결정 보존:
- 기술 스택 선택 + 이유
- 외부 의존성 (DB, 결제, 인증 provider)
- 데이터 모델 핵심 결정
- 알려진 trade-off

### Phase 7. Out-of-Scope
명시적 제외 목록. "이번 phase에서 안 할 것".

### Phase 8. Feature Candidates
PRD를 vertical slice 가능한 feature 후보로 변환 (다음 스킬 입력).
각 후보:
- 임시 ID (PRD-CAND-N)
- 한 줄 설명
- 관련 user story 매핑
- 재사용 가능 여부 가설 (`split-work-into-features`에서 확정)

## 6. 출력 템플릿 (Output Format)

```markdown
# PRD: <Product / Epic Name>

> **Status**: draft | review | locked | superseded
> **Owner**: <name>
> **Last Updated**: <YYYY-MM-DD>
> **Version**: 1.0

## 1. Problem Statement
<문제의 본질, 대상 사용자, pain point>

## 2. Proposed Solution
<기술적 + 비즈니스적 해결 방안>

## 3. Target User & Buyer
- **User**: <persona>
- **Buyer**: <persona, user와 다르면>
- **Use Cases**: 3-5개

## 4. Success Criteria
| Category | KPI | Baseline | Target | Measurement |
|---|---|---|---|---|
| User Behavior | activation rate | X% | Y% | event tracking |
| Business | conversion | X% | Y% | funnel analytics |
| Quality | LCP | 4.2s | <2.5s | Lighthouse CI |
| Operations | uptime | n/a | 99.9% | health check |

## 5. User Stories
### US-1: <story title>
- **Story**: As a <persona>, I want <action>, so that <value>
- **Acceptance Criteria**:
  - [ ] <criteria 1>
  - [ ] <criteria 2>
- **Priority**: P0
- **Dependencies**: <none / US-X>

## 6. Functional Requirements
- FR-1: <requirement>
- FR-2: ...

## 7. Non-Functional Requirements
- Performance: LCP <2.5s, FID <100ms
- Security: OWASP Top 10 통과, PII 암호화
- Availability: 99.9% uptime
- Accessibility: WCAG 2.1 AA
- I18n: ko, en

## 8. Implementation Decisions
- Stack: <Next.js, Postgres, ...>
- Auth Provider: <Supabase / Auth0 / 자체>
- Payment: <Stripe / Toss / point wallet>
- Trade-offs:
  - <decision>: chose <option> over <alternative> because <reason>

## 9. Out of Scope
- <명시 제외 1>
- <명시 제외 2>

## 10. Feature Candidates (다음 스킬 입력)
| ID | Description | Related US | Reuse Likely |
|---|---|---|---|
| PRD-CAND-1 | email/password login | US-1, US-2 | yes |
| PRD-CAND-2 | feature catalog search | US-3 | maybe |

## 11. Open Questions
- <검증 필요 가설>
- <stakeholder 결정 대기 항목>
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `validate-idea`, `assess-business-viability`, `review-scope` — `Skill` tool로 invoke
- 페어: `critique-plan` — PRD 작성 후 비평 받기
- 다음 단계: `split-work-into-features` — feature candidate를 vertical slice feature로 분해
- 후속 검증: `review-license-and-ip-risk`, `audit-security` — 상용 출시 전

## 8. Anti-patterns

1. **"빠른 로딩", "사용자 친화적"** — 측정 불가. LCP, FID, CLS 같은 숫자로.
2. **Out-of-scope 누락** — scope crawl 시작점. 명시적 제외 목록 필수.
3. **Acceptance criteria 없는 user story** — done 판단 불가. 2-5개 criteria 강제.
4. **Implementation decision 미기록** — 같은 trade-off 반복 논의. 결정 + 이유 보존.
5. **PRD를 lock 후 변경** — 변경 자체는 OK. 변경 이력 + 이유 미기록이 문제.
6. **Stakeholder 다수 명시 (책임 분산)** — 충돌 시 누가 결정? 1명 owner.
7. **Feature candidate 생략** — split-work-into-features가 다시 처음부터. 후보 목록은 핸드오프 필수.
8. **PRD를 design doc과 혼동** — PRD = what / why. design doc = how. 분리 유지.
