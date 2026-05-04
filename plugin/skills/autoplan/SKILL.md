---
name: autoplan
description: 자동 multi-stage 리뷰 파이프라인. 리뷰 스킬 (`review-scope` / `review-engineering` / `review-design` / `review-devex`)을 순차 chaining, 원칙 허용 시 auto-resolve, taste 결정만 사용자에게 노출. 단계별 handholding 없이 전체 리뷰 battery 실행에 사용.
type: skill
---

# Autoplan — Multi-Stage 리뷰 Orchestration

<overview>
One Command. Rough 계획을 입력하면 완전 리뷰된 계획이 출력된다.

`autoplan`이 설치된 리뷰 스킬을 디스크에서 읽고 full depth로 따름 — 각 스킬 수동 실행과 같은 엄격함, 같은 섹션, 같은 방법론. 유일한 차이: 중간 결정은 아래 6 원칙으로 auto-resolve. Taste 결정 (합리적 사람이 disagree 가능)은 단일 최종 승인 게이트에 노출.
</overview>

## 이 스킬을 사용하는 경우

- Multi-angle 리뷰 준비된 계획 파일 (디자인 문서, RFC, spec) 있음.
- scope/design/engineering/devex 리뷰 전반 15-30 중간 질문에 답해야 함.
- 한 명령으로 full 리뷰된 계획을 출력하고, 진짜 taste 결정만 인간 판단을 위해 노출하기 원함.
- 명시: "auto review", "automatic review pipeline", "run all reviews", "내 대신 결정해".

## Wire-Up

이 스킬은 **Claude Code의 리뷰 스킬을 순차 orchestrate**. 각 phase는 해당 스킬을 `Skill` tool로 invoke:

| 스킬 | 목적 |
|------|------|
| `review-scope` | 전략, scope, 전제 — "이게 올바른 문제인가" |
| `review-engineering` | 아키텍처, edge case, 테스트 coverage, performance |
| `review-design` | 정보 hierarchy, 상호작용 상태, 접근성 (UI scope만) |
| `review-devex` | API/CLI 인체공학, TTHW, 에러 메시지, 문서 (DX scope만) |

해당 스킬이 환경에 없으면 그 phase auto-skip (한 줄 노트).

선택: **외부 voice** (두 번째 의견 — 독립 컨텍스트의 서브 에이전트, 또는 외부 LLM CLI). 구성되면 각 phase가 **dual voice** 합의 pass 실행; 사용 불가면 single-voice 리뷰로 degrade하고 계속.

## 6 Auto-Decision 원칙

이 원칙들이 모든 중간 질문 auto-answer. 스킬의 IP — 설치된 copy에 전체 텍스트 추출.

### 1. 완성도 선택
- **적용:** 두 접근 모두 viable. 하나가 더 많은 edge case 커버.
- **Auto-action:** 더 완전한 옵션을 선택. AI는 완성도를 거의 공짜로 제공한다.
- **Escalate:** "완전" 옵션이 새 인프라, 새 외부 의존성 요구, 또는 context 없는 다른 시스템 교차.

### 2. 감당 가능한 범위 확장 (전면 재작성 금지)
- **적용:** Scope expansion 고려 중.
- **Auto-action:** Expansion이 **blast radius** (계획 수정 파일 + 직접 importer) 안이고 AND ~1일 미만 노력 (전형적 <5 파일, 새 인프라 없음)이면 승인. 감당 가능한 범위 안의 확장은 auto-approve.
- **Escalate:** Expansion이 blast radius 밖, 또는 multi-quarter migration·전체 rewrite처럼 큰 작업. 별도 계획으로 연기.

### 3. Pragmatic
- **적용:** 두 옵션이 같은 것을 동등하게 fix하는 경우.
- **Auto-action:** 더 깔끔/단순한 것을 선택. 5초 선택, 5분 re-litigate 아님.
- **Escalate:** 두 옵션이 *다른* downstream 효과 (하나가 public API 변경, 다른 건 아님 등).

### 4. DRY
- **적용:** 제안된 추가가 기존 기능 중복.
- **Auto-action:** 중복은 거부한다. 이미 있는 것을 재사용한다. 기존 utility/모듈을 이름으로 지목한다.
- **Escalate:** "기존" 것이 near-match이지만 load-bearing 의미적 차이 있음. 노출하여 사용자가 재사용 vs fork 결정.

### 5. Clever보다 Explicit
- **적용:** 10-줄 명백한 fix vs 200-줄 추상화. Clever 한 줄 vs boring 직선 코드.
- **Auto-action:** Zero context로 30초 만에 신규 기여자가 읽고 이해할 수 있는 것을 선택.
- **Escalate:** "Clever" 옵션이 이 codebase의 확립된 패턴인 경우. 기존 패턴에 맞서지 말 것.

### 6. Action 편향
- **적용:** Finding이 우려되지만 blocker는 아님. 또는 두 reviewer가 borderline 이슈가 진짜 크리티컬인지 loop에 빠진 경우.
- **Auto-action:** 우려를 리포트에 flag, 차단은 안 함. Merge가 리뷰 사이클보다 낫고, 리뷰 사이클이 stale deliberation보다 낫다.
- **Escalate:** 우려가 데이터 손실, 보안, irreversible. 이들은 절대 그냥 merge 아님 — 느려져도 escalate.

### 충돌 해결 (context 의존 tiebreaker)
- **Scope phase:** P1 (완성도) + P2 (감당 가능한 범위 확장) 우선.
- **Engineering phase:** P5 (explicit) + P3 (pragmatic) 우선.
- **Design phase:** P5 (explicit) + P1 (완성도) 우선.
- **DevEx phase:** P5 (explicit) + P1 (감당 가능한 범위 확장) 우선.

### Tiebreaker 적용 예시 (Sonnet 명확화 강제)

**예시 1 — Scope phase, P1 vs P5 충돌**:
- P1 (완성도): "5 데이터 source 모두 처리"가 옳음
- P5 (explicit): "1 source부터 보여주는 직선 코드"가 30초 신규 기여자에게 명확
- **Tiebreaker 적용**: Scope phase는 P1 + P2 우선 → **P1 채택** (완성도 우선, 단 P2의 "감당 가능한 범위" 위반 시 ocean으로 escalate)

**예시 2 — Engineering phase, P3 vs P5 충돌**:
- P3 (pragmatic): "기존 mediator 패턴 재사용"이 codebase에 맞음
- P5 (explicit): "직접 호출이 더 명확하지만 codebase 패턴과 충돌"
- **Tiebreaker 적용**: Engineering phase는 P5 + P3 우선이지만, P5는 "기존 패턴에 맞서지 말 것" 명시 → **P3 채택** (codebase 패턴 우선)

**예시 3 — Design phase, P1 vs P5 충돌**:
- P1 (완성도): empty/error/loading state 모두 디자인
- P5 (explicit): error는 generic "Something went wrong"으로 단순화
- **Tiebreaker 적용**: Design phase는 P5 + P1 우선이지만 둘이 충돌 — P1이 "신규 기여자가 모든 state 명확히 이해" 강화 → **P1 채택** (state 모두 explicit 디자인)

**예시 4 — DevEx phase, P2 vs P5 충돌**:
- P2 (감당 가능한 범위): SDK 5 language 지원이 ocean — TypeScript만 우선
- P5 (explicit): TypeScript SDK가 explicit + 30초 신규 기여자에게 명확
- **Tiebreaker 적용**: DevEx phase는 P5 + P1 우선, P2와 일치 → **P5 + P2 채택** (TypeScript 1개부터)

**Tiebreaker fail → Escalate**:
4 phase 모두 P5 우선이거나 데이터 손실 / 보안 / irreversible 케이스 다수 발견 시 — 무한 deliberation 회피하려면 즉시 user에게 escalate. tiebreaker 적용 결과 모호 또는 적용 불가하면 dual voice 호출 (`consult-codex` 스킬).

## 결정 분류

모든 auto-decision이 세 bucket 중 하나로 분류:

### Mechanical
명확 한 답. 조용히 auto-decide. 예시:
- "테스트 suite 실행?" → 항상 yes.
- "완전한 계획에 scope 축소?" → 항상 no.
- "Second-opinion 모델 실행?" → 사용 가능하면 항상 yes.

### Taste
합리적 사람이 disagree 가능. 추천과 함께 auto-decide, 하지만 인간 리뷰 위해 **최종 게이트에 노출**. 세 자연 source:

1. **Close approach** — 상위 두 옵션이 다른 tradeoff로 모두 viable.
2. **Borderline scope** — blast radius 안이지만 3-5 파일, 또는 모호한 radius.
3. **Outside-voice 불일치** — 두 번째 모델이 다르게 추천하고 유효 point.

### User Challenge
두 모델 모두 **사용자가 선언한 방향**을 바꿔야 한다고 동의. Taste와 질적으로 다름. 절대 auto-decide 안 됨.

주 reviewer + 외부 voice가 사용자가 명시 지정한 것을 merge, 분할, 추가, 제거 추천할 때 — 그게 User Challenge. 사용자는 항상 모델이 결여한 context 보유.

User Challenge는 taste 결정보다 richer framing으로 최종 게이트 이동:
- **사용자가 말한 것:** 원래 방향.
- **두 모델이 추천하는 것:** 제안된 변경.
- **왜:** 모델 reasoning.
- **놓친 context:** blind spot 명시 인정.
- **틀리면 비용:** 사용자 원래가 맞았을 때 뭐 일어남.

사용자가 선언한 방향이 **기본값**. 모델이 변경 case 만들어야지, 반대 아님.

**예외:** 두 모델 모두 변경을 보안 또는 feasibility 리스크 (선호 아님)로 flag하면 framing이 명시: "두 모델 모두 이를 선호가 아니라 보안/feasibility 리스크로 믿음." 사용자가 여전히 결정한다.

## 결정 게이트: Taste vs Auto-Resolvable

Load-bearing 테스트. 만나는 모든 결정에:

```
답이 6 원칙 중 하나로 ambiguity 없이 결정?
├── YES → MECHANICAL. 조용히 auto-decide. Audit trail에 log.
└── NO → 합리적 시니어 엔지니어가 다르게 pick 가능?
        ├── NO → MECHANICAL (한 옵션 명확 지배). Auto-decide. Log.
        └── YES → 사용자가 명시 지정한 것 변경?
                 ├── YES (두 voice 동의해 변경) → USER CHALLENGE.
                 │   최종 게이트 위해 hold. 사용자 결정.
                 └── NO → TASTE. 원칙 사용해 auto-decide, 하지만
                         최종 게이트에 노출하여 사용자 override 가능.
```

**사용자에게 도달하는 질문 type은 단 두 가지다:**
1. 전제 확인 (Phase 1) — 무엇을 해결 중인지.
2. User Challenge — 두 모델이 사용자가 명시한 것을 변경하라고 말할 때.

나머지 모두 auto-decide. 사용자는 끝에서 모든 taste 결정이 list된 단일 배치 승인 게이트를 획득해 cheap override 가능.

## 순차 vs 병렬 실행

**Phase는 엄격한 순서로 실행 필수:** CEO → Design → Eng → DX.

각 phase는 다음 phase 시작 전 완전히 완료해야 한다. 절대 병렬 phase 실행 금지 — 각각이 이전 위에 빌드. CEO의 전제 framing이 Eng가 "in scope"로 평가하는 것에 영향. Design의 hierarchy 결정이 Eng의 컴포넌트 경계에 영향. Eng의 아키텍처가 DX의 API surface에 영향.

각 phase 사이에 phase-전환 요약을 emit하고, 다음 시작 전 이전 phase의 모든 필수 output이 쓰여졌는지 검증한다.

**Phase 안에서 dual voice는 병렬 실행 가능** (주 스킬 + 외부 voice가 독립). 추천 패턴: 외부 voice 먼저 kick off (느리고 종종 네트워크 바운드), 그게 반환되는 동안 주 reviewer 실행. 합의 테이블 빌드 전 둘 완료 필수.

## 파이프라인 워크플로우

### Phase 0: 접수 + Restore Point

**Step 1 — Restore point 캡처.** 계획 파일의 현재 내용을 변이 전 외부 파일에 저장. 계획 파일에 restore 경로를 가리키는 한 줄 주석을 앞에 덧붙임. 뭔가 잘못되면 사용자가 원본을 copy back 가능.

**Step 2 — Context 읽기.** `CLAUDE.md` (프로젝트 컨벤션), `TODOS.md` (이미 연기된 것), 최근 git log, 베이스 브랜치 대비 diff 읽기.

**Step 3 — Scope 감지.** 조건부 phase를 trigger하는 용어가 있는지 계획에서 grep:
- **UI scope** (design phase trigger): component, screen, form, button, modal, layout, dashboard, sidebar, nav, dialog. 2회 이상 일치 필요.
- **DX scope** (DevEx phase trigger): API, endpoint, REST, GraphQL, webhook, CLI, command, flag, SDK, library, package, npm, pip, import, agent, MCP, developer docs, getting started, integration, debug, error message. 2회 이상 일치 필요. 제품 자체가 developer tool이거나 AI agent가 주 사용자면 또한 trigger.

**Step 4 — 스킬 invoke 준비.** 각 리뷰 스킬을 `Skill` tool로 호출 (스킬 본문을 메모리에 로드):
- `review-scope`
- `review-design` (UI scope 감지 시만)
- `review-engineering`
- `review-devex` (DX scope 감지 시만)

**섹션 skip list — invoke된 스킬을 따를 때 이 섹션은 skip** (autoplan이 이미 처리, 또는 batch 모드에 적용 안 됨):
- 스킬 자체 preamble / setup
- 질문 포맷 directive (autoplan이 자체 공급)
- 외부 voice setup (autoplan이 전역 한 번)
- "Prerequisite 스킬 제공" 섹션

리뷰 특정 방법론, 섹션, 필수 output**만** 따름.

Output: "무엇을 작업하는지: [계획 요약]. UI scope: [yes/no]. DX scope: [yes/no]. 디스크에서 리뷰 스킬 로드. Auto-decision으로 full 리뷰 파이프라인 시작."

### Phase 1: Scope 리뷰 (전략 & scope)

`review-scope`을 end-to-end 따름. 모든 결정 질문 override → 6 원칙 사용해 auto-decide. 유일한 예외:

**전제 게이트 (Phase-1에서 사용자에게 도달하는 유일한 질문).** 전제는 "어떤 문제를 해결 중이며 어떤 가정이 뒷받침하는가". 인간 판단 필요 — 전제 auto-decide는 무엇을 빌드할지 auto-decide이고, 이 스킬 scope 밖이다.

나머지 모두 override 규칙:
- **모드 선택:** SELECTIVE EXPANSION (batch 리뷰 기본).
- **대안:** 최고 완성도를 선택 (P1). 동점이면 가장 단순한 것을 선택 (P5). 상위 2 close면 → TASTE DECISION 표시.
- **Scope expansion:** blast radius 안 + <1일 노력 → 승인 (P2). 밖 → TODOS.md로 연기. 중복 → 거부 (P4). Borderline (3-5 파일) → TASTE DECISION.
- **전략 선택:** 외부 voice가 유효 이유로 disagree → TASTE DECISION. 두 모델이 사용자가 선언한 구조 변경에 동의 → USER CHALLENGE.

**Scope phase 필수 output:**
- 전제마다 이름이 붙고 평가가 끝난 전제 challenge.
- "이미 존재하는 것" 맵 (하위 문제 → 기존 코드).
- 연기 항목과 근거 있는 "NOT in scope" 섹션.
- 실패 모드 / 에러 registry.
- Dual voice 합의 테이블 (외부 voice 구성 시).
- 완료 요약.

Phase-전환 요약 emit. Phase 2 전에 모든 output이 디스크에 있는지 검증.

### Phase 2: Design 리뷰 (조건부 — UI scope 없으면 skip)

`review-design`을 end-to-end 따름. 모든 결정 질문 override → 6 원칙 사용해 auto-decide.

Override 규칙:
- **구조 이슈** (누락 상태, 깨진 hierarchy): auto-fix (P5).
- **미적 / taste 이슈:** TASTE DECISION 표시.
- **디자인 시스템 정렬:** 디자인 시스템 문서 존재하고 fix 명백하면 auto-fix.

필수 output:
- 모든 디자인 차원이 점수와 함께 평가됨.
- 식별되고 auto-decide된 이슈.
- Dual voice 합의 테이블 (외부 voice 구성 시).

Phase-전환 요약 emit.

### Phase 3: Engineering 리뷰 (아키텍처, 테스트, performance)

`review-engineering`을 end-to-end 따름. 모든 결정 질문 override → 6 원칙 사용해 auto-decide.

Override 규칙:
- **Scope challenge:** 절대 축소 금지 (P2).
- **아키텍처 선택:** clever보다 explicit (P5). 외부 voice가 유효 이유로 disagree → TASTE DECISION.
- **테스트 coverage:** 항상 모든 관련 suite 포함 (P1). 각 gap에 대해 add-test vs defer-with-rationale 결정 → log.
- **연기된 scope:** TODOS.md에 auto-write (모든 phase에서 수집).

**크리티컬: 섹션 3 (테스트 리뷰)은 절대 skip이나 compress 금지.** 실제 코드를 읽고 테스트 다이어그램을 빌드: 모든 새 UX flow, data flow, codepath, branch list. 각각을 어떤 type의 테스트가 커버하는가? 그 테스트가 존재하는가? 테스트 gap auto-decide 의미: gap 식별 → add 또는 defer 결정 → log. 분석을 skip한다는 의미가 아님.

필수 output:
- 아키텍처 다이어그램 (ASCII 또는 동등물).
- Codepath를 coverage에 매핑하는 테스트 다이어그램.
- 디스크에 쓰여진 테스트 plan artifact.
- Critical-gap flag가 있는 실패 모드 registry.
- Dual voice 합의 테이블 (외부 voice 구성 시).

Phase-전환 요약 emit.

### Phase 3.5: DevEx 리뷰 (조건부 — DX scope 없으면 skip)

`review-devex`을 end-to-end 따름. 모든 결정 질문 override → 6 원칙 사용해 auto-decide.

Override 규칙:
- **페르소나:** README/docs에서 추론하여, 이 제품이 타겟팅하는 가장 흔한 개발자 타입을 선택 (P6).
- **Magical moment (사용자가 경탄하는 순간):** 경쟁 tier에 도달하는 가장 적은 노력의 전달 수단 (P5).
- **Getting-started 마찰:** 항상 더 적은 단계를 향해 최적화 (P5).
- **에러 메시지 품질:** 항상 문제 + 원인 + fix 요구 (P1).
- **API/CLI 네이밍:** cleverness보다 consistency (P5).
- **Opinionated 기본값 vs 유연성:** TASTE DECISION 표시.

필수 output:
- 개발자 journey map.
- 타겟 있는 Time-to-Hello-World 평가.
- 모든 차원 점수 있는 DX 스코어카드.
- Dual voice 합의 테이블 (외부 voice 구성 시).

Phase-전환 요약 emit.

### 결정 Audit Trail

각 auto-decision 후 계획 파일에 row를 append:

```markdown
<!-- AUTONOMOUS DECISION LOG -->
## Decision Audit Trail

| # | Phase | Decision | Classification | Principle | Rationale | Rejected |
|---|-------|----------|----------------|-----------|-----------|----------|
```

결정당 한 row, 진행하며 점진 작성. Audit은 대화 context가 아니라 디스크에 존재.

### 게이트 전 검증

최종 게이트 제시 전 필수 output이 실제 생산됐는지 검증한다. 실행된 각 phase에 대해 체크리스트 walk:
- 모든 필수 섹션이 생산됨 (한 줄 요약이 아니라 실제 분석 포함).
- 필요한 곳에 모든 artifact가 디스크에 쓰여짐.
- Dual voice 실행 (또는 unavailable 표시).
- 합의 테이블 생산.
- 결정 audit trail이 auto-decision당 최소 한 row를 보유.

무엇이든 누락이면 돌아가 생산. 최대 2 재시도 — 그래도 누락이면 미완 상태임을 표시하는 warning을 붙여 게이트로 진행. 무한 loop 금지.

### Phase 4: 최종 승인 게이트

**STOP하고 사용자에게 최종 상태 제시.**

```
## Review Complete

### Plan Summary
[1-3 문장 요약]

### Decisions Made: [N] total ([M] auto-decided, [K] taste 선택, [J] user challenge)

### User Challenges (두 모델이 선언 방향에 disagree)
[각각에:]
**Challenge [N]: [제목]** (from [phase])
You said: [사용자 원래 방향]
두 모델 추천: [변경]
왜: [reasoning]
놓쳤을 수 있는 것: [blind spot]
틀리면 비용: [변경의 downside]
[보안/feasibility면: "두 모델 모두 이를 선호가 아닌 보안/feasibility 리스크로 flag."]
당신의 결정 — 명시적으로 변경하지 않으면 원래 방향을 유지한다.

### Your Choices (taste 결정)
[각각에:]
**Choice [N]: [제목]** (from [phase])
[X] 추천 — [principle]. 하지만 [Y]도 viable:
  [Y 선택 시 downstream 임팩트 한 문장]

### Auto-Decided: [M] 결정 [계획 파일의 Decision Audit Trail 참조]

### Review Scores
- CEO: [요약, dual voice 실행 시 consensus N/N 포함]
- Design: [요약 또는 "UI scope 없음으로 skip"]
- Eng: [요약, dual voice 실행 시 consensus N/N 포함]
- DX: [요약 또는 "DX scope 없음으로 skip"]

### Cross-Phase Theme
[2+ phase의 dual voice에서 독립적으로 나타난 우려:]
**Theme: [topic]** — [Phase 1, Phase 3]에서 flag. 고신뢰도 신호.
[Theme 없으면:] "Cross-phase theme 없음 — 각 phase의 우려가 구별됨."

### Deferred to TODOS.md
[이유와 함께 auto-defer된 항목]
```

**인지 부하 관리:**
- 0 user challenge: User Challenges 섹션 skip.
- 0 taste 결정: Your Choices 섹션 skip.
- 1-7 taste 결정: flat list.
- 8+: phase별 그룹. Warning 추가: "이 계획은 ambiguity가 비정상적으로 높습니다 (taste 결정 [N]개). 신중히 리뷰하세요."

사용자 옵션:
- A) As-is 승인 (모든 추천 수락).
- B) Override와 승인 (어떤 taste 결정 변경할지 지정).
- B2) User-challenge 응답과 승인 (각각 수락 또는 거부).
- C) Interrogate (특정 결정에 대해 질문).
- D) Modify (계획 자체 변경 필요 — 영향받은 phase 재실행).
- E) Reject (처음부터 시작).

**옵션 처리:**
- A: APPROVED 표시, log 쓰기, 다음 워크플로우 단계 제안.
- B: 어떤 override인지 질문, 적용, 게이트 재제시.
- C: 자유 형식 답, 게이트 재제시.
- D: 변경, 영향받은 phase만 재실행. 최대 3 cycle.
- E: 처음부터 시작.

## 결과 Aggregation

최종 게이트가 aggregation. 모든 phase에 대해:

1. **모든 결정**이 분류 (mechanical / taste / user-challenge)되고 디스크 audit trail에 log.
2. **Taste 결정**이 게이트의 한 "Your Choices" list에 배치.
3. **User Challenge**는 richer framing을 획득하고 taste 결정 위에 존재 (더 많은 사고 필요).
4. **Mechanical 결정**은 게이트 UI에서 조용하지만 post-hoc 검사를 위해 audit trail에 보인다.
5. **Cross-phase theme**은 별도로 노출 — 같은 우려가 2+ phase에 독립적으로 나타나면, 단일 phase가 크리티컬로 flag하지 않았어도 highlight 가치 있는 고신뢰 신호다.
6. **점수**는 각 phase에서 inline 요약 (per-phase output에 묻히지 않음) — 사용자는 단일 화면 overview를 획득한다.
7. **연기 항목**은 TODOS.md로 흐름 — 아무것도 잃지 않도록.

사용자는 한 화면을 읽고, 한 결정을 내리고, ship 또는 iterate.

## 중요 규칙

- **절대 abort 금지.** 사용자가 autoplan을 선택했다. 그 선택을 존중한다. Taste 결정을 게이트에 노출; 파이프라인 중간에 대화형 리뷰로 redirect 절대 금지.
- **오직 두 개의 게이트만.** (1) Phase 1의 전제 확인. (2) User Challenge. 나머지 모두 auto-decide.
- **모든 결정 log.** 조용한 auto-decision 없음. 모든 선택이 audit trail row를 획득한다.
- **Full depth는 글자 그대로 full depth를 의미한다.** 로드된 스킬 파일의 섹션을 compress하거나 skip하지 말 것 (Phase 0의 skip list 제외). Full depth 의미: 섹션이 요청하는 코드 읽기, 섹션이 요구하는 output 생산, 모든 이슈 식별, 각각 결정. 섹션의 한 문장 요약은 full depth가 아님 — skip. 리뷰 섹션에 3 문장 미만을 쓰는 자신을 catch하면 compress 중이라는 신호다.
- **Artifact가 deliverable.** 테스트 plan artifact, 실패 모드 registry, 에러/rescue 테이블, 아키텍처 다이어그램 — 리뷰 완료 시 디스크나 계획 파일에 반드시 존재. 존재하지 않으면 리뷰는 불완전하다.
- **외부 voice는 best-effort.** 두 번째 모델이 unavailable (auth 누락, binary 누락, timeout)이면 phase에 `[single-voice]` 태그를 붙이고 계속. 외부 voice 가용성으로 파이프라인을 차단하지 않는다.

## Output

다음을 포함한 리뷰된 계획 파일:
- 각 phase의 finding을 인라인 annotation.
- 모든 auto-decision이 log된 하단 `## Decision Audit Trail` 테이블.
- 존재하는 곳에 cross-phase theme을 call out.
- 사용자가 최종 게이트를 통과하면 명확한 `APPROVED` marker.

추가로, 모든 taste 결정과 user challenge를 한 화면에 배치하는 단일 사용자 대상 승인 메시지.
