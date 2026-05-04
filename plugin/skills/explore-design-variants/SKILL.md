---
name: explore-design-variants
description: "N variants를 parallel 생성하고 structured feedback으로 iterate. divergent 탐색 패턴 — 디자인뿐 아니라 copy/architecture option/naming brainstorm. 트리거: '디자인 옵션 보여줘' / '변형 brainstorm' / '다른 디자인 안' / 'shotgun 디자인' / 'design variants' / '여러 안 보여줘' / 'alternative designs'. 입력: 한 가지 시안 + 평가 차원. 출력: 3-5개 변형 + 각 trade-off. 흐름: review-design → explore-design-variants → consult-design-system."
type: skill
---

# Explore Design Variants — 병렬 Variant 탐색


수렴 전 **divergent 탐색**을 위한 범용 패턴. 한 답을 만들고 방어하는 대신, 의도적으로 다른 N개 답을 만들고, 구조화된 피드백을 모으고, 반복. 취향, 판단, 제약 트레이드오프가 연루된 모든 곳에 작동 — 시각 디자인, microcopy, 아키텍처 옵션, 네이밍, prompt 초안, API shape, 에러 메시지, "떠오른 첫 아이디어"가 약한 기본인 모든 것.

이름이 "design-shotgun"인 이유는 시각 디자인에서 출발했기 때문, 그러나 메커니즘 — 병렬 탐색 → 나란히 비교 → rubric 피드백 → 개선 — 은 domain-agnostic.

## 이 스킬을 사용하는 경우

- 단일 방향이 premature commitment일 때 (아직 뭐가 좋은지 모름).
- 여러 옵션 생성 비용이 잘못 고르는 비용에 비해 작을 때.
- 사용자가 "옵션 보여줘", "이건 안 좋아", "탐색", "brainstorm"이라 말했거나, 답이 정답이 아니라 취향인 질문을 했을 때.
- Stakeholder가 방향에 불일치하고 반응할 구체적 artifact 필요.
- 떠오른 첫 아이디어로 수렴하는 자신을 catch.

**이 스킬 쓰지 말 것:** 답이 정확성에 의해 결정 (버그 fix, 수학 문제), 제약 주어지면 하나의 viable 방향만, 또는 사용자가 이미 방향을 잠그고 실행만 원할 때.

## 패턴

1. **N개 variant 병렬 생성** (보통 3-5).
2. **다양성 강제** — variant는 다른 방향을 탐색해야, 단일 insight 주변에 cluster 금지.
3. **일관된 비교 framework로 나란히 제시**.
4. **Rubric으로 구조화 피드백 수집** (자유 형식 의견 아님).
5. **반복** — 승자 refine, 패자 kill, gap regenerate, 반복.

Step 1-4는 한 라운드. Step 5는 수렴 전 여러 라운드 실행 가능. 사용자가 확신과 함께 "이거"라고 말하거나 rubric이 최상위 variant가 quality bar 넘었다 할 때 중단.

## Variant 생성

### Step 0 — Concept stub 먼저

Compute 소비 전 (또는 domain 관련 시 코드/copy 쓰기 전), 각 variant에 대해 **한 줄 텍스트 concept** 작성. 각 concept은 구별되는 창의적 방향, 작은 variation 아님. Lettered list로 제시:

```
N 방향 탐색:

A) "이름" — 이 방향의 한 줄 요약
B) "이름" — 이 방향의 한 줄 요약
C) "이름" — 이 방향의 한 줄 요약
```

그다음 confirm: "이들이 내가 생성할 N 방향입니다. 맞나요? swap하고 싶은 게 있나요?" 이게 생성 budget 소비 전 방향 레벨의 불일치를 catch.

Concept 개정 최대 2 라운드, 그다음 commit하고 생성.

### 다양성 규칙 (hard 요구사항)

각 variant는 같은 axis의 다른 강도가 아니라 **design space의 다른 axis를** 탐색해야. 다양성 테스트는 구체적이고 unforgiving.

**Swap test:** 두 variant의 표면 세부를 swap해도 인상이 실질적으로 바뀌지 않으면 둘은 너무 비슷. 약한 쪽을 의도적으로 다른 premise로 regenerate.

Swap test의 도메인별 적용:

- **시각 디자인:** 다른 font family, 색상 팔레트, AND 레이아웃 접근. A와 B 둘 다 light 배경에 centered hero의 sans-serif면 실패.
- **Copy / microcopy:** 다른 voice, 다른 수사적 move, 다른 길이. A와 B 둘 다 "공손 + 기능적 + 12 단어"면 하나 실패.
- **아키텍처 옵션:** 다른 fundamental decomposition (예: monolith vs services vs serverless), "monolith with feature flag X" vs "monolith with feature flag Y"가 아니라.
- **네이밍:** 다른 metaphor family, 세 synonym 아님 (`Pulse`, `Beat`, `Tempo`는 세 옵션 아님 — 하나 옵션).
- **API shape:** 다른 추상화 레벨 (예: low-level primitive vs declarative DSL vs object-oriented facade), REST의 세 flavor 아님.

수렴을 catch하면 빠진 것 명명: "세 개 모두 [공유 axis] 탐색. variant C를 [orthogonal axis]로 regenerate."

### 병렬 Dispatch 패턴

전체 wall time이 N variant가 아니라 한 variant 분이 되도록 **단일 dispatch로 N개 독립 탐색 에이전트** spawn. 각 에이전트:

- 전체 brief + variant별 premise 수신.
- 다른 variant의 지식 없이 작동 (cross-contamination 없음, 합의 압력 없음).
- 자급자족 artifact + 간단한 근거 반환 ("이 variant는 X에 최적화, Y 트레이드오프").
- 명확한 상태 리포트: `VARIANT_<letter>_DONE`, `VARIANT_<letter>_FAILED: <reason>`, 또는 `VARIANT_<letter>_RATE_LIMITED`.

에이전트 prompt 템플릿:

```
You are generating one variant of an exploration. Your premise: {variant-specific direction}
Full brief: {shared brief}
Constraints all variants share: {shared constraints}
Your differentiator: {what makes this variant distinct from the others — describe the axis}

Produce: {artifact spec — one PNG, one paragraph, one architecture sketch, etc.}
Report: VARIANT_<letter>_DONE: <one-line summary of what you produced and the
  trade-off it represents>, OR VARIANT_<letter>_FAILED: <reason>.

Do not hedge. Pick a direction and commit to it. Other agents are exploring
other directions in parallel — your job is to make YOUR direction land hard.
```

Runtime이 병렬 dispatch 못 하면 순차 생성하되 "각 에이전트는 다른 것을 못 봄" 제약 유지. 생성 중 cross-pollination은 variant cluster 붕괴의 #1 원인.

**실패 처리:** 에이전트 실패하면 조용히 slot drop 금지. variant regenerate(count 보존) 또는 명시 리포트: "4개 중 3 variant 생성; D는 [reason]으로 실패. 3개로 진행할까요, D 재시도할까요?"

## 구조화 피드백 Rubric

자유 형식 "어떤 게 제일 좋아?"는 노이즈, post-hoc 합리화, 반복 어려운 답 생성. Rubric 사용. Rubric은 사용자가 달리 blend할 axes를 분리 강제하고, 다음 라운드에 실제 action 가능한 피드백 신호 제공.

기준을 도메인에 adapt. Shape는 동일:

| Variant | Criterion 1 (1-5) | Criterion 2 (1-5) | Criterion 3 (1-5) | 노트 (variant별) | Verdict |
|---------|-------------------|-------------------|-------------------|---------------------|---------|
| A       |                   |                   |                   |                     |         |
| B       |                   |                   |                   |                     |         |
| C       |                   |                   |                   |                     |         |

Variant별 테이블 뒤, 하나의 **전체 방향 문장** 요청 — 여기서 사용자가 다음 라운드에 원하는 synthesis 표현 (예: "A로 가되, B의 더 큰 CTA").

### 도메인별 제안 기준

- **시각 디자인:** clarity, hierarchy, taste, on-brand-ness, accessibility 힌트.
- **Copy:** clarity, voice 매치, 길이 적절, action 유도, no jargon.
- **아키텍처:** simplicity, 변경 tolerance, 운영 비용, 팀 친숙도, 실패 모드 blast radius.
- **네이밍:** memorable, pronounceable, not-already-taken, signals-the-thing, ages-well.
- **API shape:** discoverable, misuse 어려움, consistent, evolvable, debuggable.

3-5 기준 선택. 5 이상이면 사용자 disengage; 3 미만이면 signal이 action하기엔 너무 noisy.

### Score 해석

- **모든 variant가 전반적으로 4-5 점수** → brief가 틀림; 사용자가 공손. Push back: "이들 모두 잘 평가되지만 목표가 X라 하셨습니다 — 어떤 게 X를 실제 가장 잘 달성하고, 거기서 뭘 cut할까요?"
- **Wide spread** → 다양성 규칙 작동. 신호 있음.
- **대부분 axis 명확 승자, 한 axis 패자** → refinement 타겟 있음. 승자를 패자 axis focus로 반복.
- **승자 없음, 모든 variant가 다른 axis에서 강함** → 다음 라운드에 remix variant 고려 (A의 레이아웃 + B의 copy + C의 tone).

## 비교 Framework

나란히가 순차를 매번 이긴다. 사용자 눈과 판단은 절대 평가가 아니라 대조로 작동. 도메인 무관, axis별 비교가 싸도록 variant 제시.

각 도메인:

- **시각:** literal 나란히 렌더링 (이미지 그리드, 비교 HTML 페이지, 인쇄 시트). 같은 스케일, 같은 crop, 같은 주변 chrome.
- **Copy:** variant당 한 컬럼, 사용 context(버튼 레이블, 에러 메시지, empty state 등)당 한 row의 markdown 테이블. 사용자가 각 voice가 context에 어떻게 재생되는지 봄.
- **아키텍처:** variant당 parallel 섹션, 동일 sub-header(Decomposition, Data flow, Failure modes, Cost, Migration path)의 단일 문서. Apples-to-apples 강제.
- **네이밍:** 왼쪽에 이름, 각 axis(memorability, domain fit, availability 등) 컬럼 + 사용자가 채우는 "first reaction" 컬럼의 테이블.

Framework는 시각이 아니라 구조. 터미널에서도 동일 sub-header의 parallel 섹션이 대부분 일 수행.

## 반복 루프

피드백 후 세 상황 중 하나:

1. **명확 승자, 작은 tweak** → 특정 피드백으로 승자에 단일 refinement pass 실행. 결과 표시. Confirm. 완료.
2. **방향은 명확 승자, 실행 약함** → 더 타이트한 brief로 variant 생성 재실행: "A 방향으로 갑니다. A의 premise 사용하지만 [사용자가 불평한 axis]에서 다양한 3개 refinement 생성."
3. **승자 없음, variant 간 부분 선호** → "remix" 라운드 실행. 사용자 spec에 따라 element 결합하는 2-3 새 variant 생성 (예: "A의 레이아웃 + B의 색상 + C의 copy tone").

**중단 조건:**

- 사용자가 확신("I guess" 또는 "maybe" 없이)으로 "이거"라고 함.
- 최상위 variant가 설정 bar에서 모든 rubric axis 통과.
- 수렴 없이 라운드 4 도달 — escalate: "4 라운드 반복 중. Blocker가 [관찰된 패턴]으로 보입니다. Re-scope하려 돌아갈까요, 현재 최선에 commit하고 진행할까요?"

**안티패턴: 움직이는 타겟 쫓기.** 사용자 피드백이 이전 라운드 피드백과 모순되면 명명: "지난 라운드엔 더 높은 밀도 원하셨고, 이번엔 더 많은 여백 요청. 어느 쪽으로 push하길 원하세요?" 꼬투리 잡기가 아니라 — 진동(oscillation)을 막아주는 것.

## 도메인 Adaptation

패턴은 동일. Artifact와 rubric 기준이 변경.

### 시각 디자인

- N = 3 typical, 중요 스크린엔 5-8.
- 다양성 axis: typeface family, 색상 팔레트, 레이아웃 그리드, 밀도.
- 비교: 나란히 이미지 그리드, 같은 crop과 스케일.
- Rubric: clarity, hierarchy, on-brand, accessibility, "고객에게 보여줄까".
- 반복: 승자 variant refine, 또는 element remix.

### Copy / Microcopy

- N = 3-5.
- 다양성 axis: voice(formal/casual), 길이, 수사적 move (instructive/inviting/warning), reading level.
- 비교: variant당 한 컬럼, context(button, empty state, error, success)당 한 row의 테이블.
- Rubric: clarity, voice fit, action-driving, jargon-free, length-appropriate.
- 반복: 승자 refine OR remix (이 voice + 저 길이).

### 아키텍처 옵션

- N = 2-4 (4 이상은 분석 depth 희석).
- 다양성 axis: fundamental decomposition (mono/services/serverless/hybrid), data ownership, sync vs async boundaries, build vs buy.
- 비교: 동일 sub-header(Decomposition, Data flow, Failure modes, Cost, Migration path, Team familiarity)의 parallel 섹션.
- Rubric: simplicity, change-tolerance, ops cost, team-fit, blast radius.
- 반복: 더 깊은 엣지 케이스 분석으로 승자 refine, 또는 두 옵션을 hybrid로 merge.

### 네이밍

- N = 5-10 (네이밍은 cheap exploration value 많음).
- 다양성 axis: metaphor family, 음절 수, 어원, phonetics.
- 비교: axes(memorable, pronounceable, available, signals-the-thing, ages-well, "first reaction") 테이블.
- Rubric: signal, availability, 구별성, ages-well, "회의에서 소리 내 말하기" 테스트.
- 반복: landing한 metaphor 가져와 같은 family에서 5개 더 생성. 또는 runner-up의 가장 강한 속성 pick하고 그것 유지하는 variant 요청.

### Prompt / Spec 초안

- N = 3-4.
- 다양성 axis: 구조 (numbered step vs prose vs role-based vs example-driven), 제약 레벨, 가정 reader 전문성.
- 비교: 각 variant가 처리한 동일 input 예시로 나란히.
- Rubric: clarity, variation robustness, 길이, misuse 어려움, outputs-match-intent.
- 반복: adversarial input으로 승자 refine, 또는 다른 variant의 구조 + 제약 레벨 remix.

## 출력

수렴 후 deliver:

1. **선택된 artifact** (승인 variant, 라운드 간 refine 가능).
2. **짧은 근거** — 왜 이게 이겼는지, 트레이드오프, close runners-up이 포기한 것.
3. **구조화된 피드백 기록** (rubric 점수 + 코멘트) — audit용, 나중 결정 설명용, 다음 탐색 intuition 훈련용.
4. **선택: "거부한 것과 이유"** 노트 — 나중에 relitigate 유혹 kill, 이 계승자를 위한 옵션 공간 문서화.

다른 워크플로우(예: 구현 계획 작성 전 디자인 옵션 원하는 planning 스킬)에서 호출되면 구조화된 피드백 반환해 호출 스킬이 다시 묻지 않고 선택 방향 소비 가능.

## 주시할 실패 모드

- **Concept stub의 premature 수렴.** Lettered list가 한 아이디어의 세 flavor처럼 보이면 병렬 생성 budget 소비 전 regenerate.
- **공손한 채점** (모두 4-5). Push back; 사용자가 받아들일 트레이드오프 요청.
- **라운드 간 모순.** 복합화 말고 명명.
- **Rubric drift.** 탐색 중 기준 변경 금지; 비교 가능성 상실.
- **사용자가 생성 안 한 variant pick.** 훌륭한 신호 — 다양성이 작동했고 사용자가 이제 뭘 원하는지 알았다는 뜻. 그들의 새 방향 주변으로 한 라운드 더 실행.
