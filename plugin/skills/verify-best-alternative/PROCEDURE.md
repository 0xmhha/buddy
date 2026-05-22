# Verify Best Alternative — AI 편향 방지 강제 다관점 검토


AI 협업 개발에서 *모델이 첫 답에 commit하려는 경향*과 *학습 분포에 의한 편향*이 엔지니어링 결정 품질을 저하시키는 것을 차단하는 *강제 다관점 검토 절차*. "이 결정이 최선인가"를 기계적으로 검증 — 단일 방향이 그럴듯해 보이더라도 의도적으로 orthogonal한 N개 대안을 발산시키고, rubric으로 비교, **어느 관점에서 봐도 최선**인 설계·구현·알고리즘을 선택.

적용 영역 (**엔지니어링 한정**): 아키텍처 옵션 / 데이터 모델 / 알고리즘 선택 / API shape / 인증 모델 / 멀티-tenant 격리 전략 / 이벤트 스키마 / 시크릿 관리 전략 / 스택 선택 / 코드 네이밍 (함수·타입·변수·모듈) / prompt engineering (AI 기능 구현) — *AI 첫 답이 약한 기본*일 수 있는 모든 엔지니어링 결정. 요구사항과 환경 제약 하에서 *베스트 선택지로 작업이 진행되도록* 강제하는 것이 목적.

> **Scope 제한**: 본 스킬은 *엔지니어링 결정 한정*. 그래픽 디자인 (typography·color·layout), 브랜드/제품 네이밍, microcopy/UI text, 마케팅 카피, 사업 기획 옵션 검토는 *별도 스킬* (미래 작업, 본 스킬과 분리). "design"이라는 단어가 그래픽 디자인·엔지니어링 설계·브랜딩·사업 기획 모두에 쓰여 혼란을 야기하나, 본 스킬은 *엔지니어링 설계·구현 결정*에만 해당.

본 스킬은 *옵션 발산*이 목적이 아니라 *결정 품질 보장*. buddy의 §3 Technical Design 결정 스킬들(`design-system`, `define-tech-stack`, `design-api-contract`, `design-data-model`, `design-event-schema`, `design-auth-model`, `design-tenant-model`, `design-secret-management`)에서 *첫 답 commit 직전* 의무 호출되어야 한다 — 자동 dispatch 또는 명시 호출. 메커니즘(병렬 탐색 → 나란히 비교 → rubric 피드백 → 개선)은 *엔지니어링 영역 내에서* domain-agnostic.

> *이전 이름 `explore-design-variants`는 시각 디자인 한정으로 오해되어 2026-05-20 rename됨. 메커니즘과 본문은 동일.*

## 이 스킬을 사용하는 경우

- 엔지니어링 결정에서 *단일 방향이 premature commitment*일 때 — AI가 첫 답을 굳히려 함, 사용자도 다른 옵션 못 본 상태
- 요구사항·환경 제약 하에 *여러 viable한 후보가 존재*할 가능성이 높을 때 (예: 여러 DB·여러 framework·여러 API style·여러 알고리즘이 모두 통할 때)
- 잘못 고르면 *재작업 비용이 후보 생성 비용보다 큰* 결정 (스택 lock-in, 데이터 모델, 인증 구조 등)
- AI 또는 사용자가 *떠오른 첫 아이디어로 수렴*하는 자신을 catch
- §3 design 스킬 (design-system / define-tech-stack 등) 결정 Step 진입 직전 — 의무 호출

**이 스킬 쓰지 말 것**:
- 답이 *정확성*에 의해 결정되는 문제 (bug fix, 수학 문제, 알려진 spec 구현)
- 제약 하에 *하나의 viable 방향만* 존재 (예: legacy 호환 요구로 stack 고정)
- 사용자가 *이미 방향을 잠그고 실행만 원함*
- **그래픽 디자인 / 브랜드 / 마케팅 / 사업 기획 옵션 검토** — 별도 스킬 (미래)
- **막힘 상태에서 문제 분해** (옵션이 아예 안 보임) → `decompose-blocker` 사용
- **bug 증상은 명확한데 root cause 분석** → `diagnose-bug` 사용

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

Swap test의 엔지니어링 도메인별 적용:

- **아키텍처 옵션:** 다른 fundamental decomposition (예: monolith vs microservices vs serverless), "monolith with feature flag X" vs "monolith with feature flag Y"가 아니라.
- **데이터 모델:** 다른 normalization 전략 (예: 3NF vs denormalized read-model vs event-sourced), 또는 다른 primary key 전략 (UUID vs sequential vs composite). 같은 schema의 column 1~2개 차이 아님.
- **알고리즘 선택:** 다른 시간복잡도 trade-off (예: hash + O(1) lookup vs sorted + O(log n) + range query 가능). 같은 O() 안의 변형 아님.
- **API shape:** 다른 추상화 레벨 (예: low-level primitive vs declarative DSL vs object-oriented facade), REST의 세 flavor 아님.
- **인증 메커니즘:** 다른 trust model (예: session cookie vs JWT bearer vs OAuth2 federation), 같은 mechanism의 token TTL 차이 아님.
- **스택 선택:** 다른 *runtime category* (예: Go monolith vs Python + FastAPI vs Node + Express vs Rust + Axum), 같은 언어의 framework 변형 아님.
- **코드 네이밍:** 다른 *의미 framework* (예: `userToken` vs `sessionHandle` vs `authContext` — abstraction level이 다름), `userToken` vs `userTok` vs `usrToken` 같은 *철자 변형*은 한 옵션.

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

### 엔지니어링 도메인별 제안 기준

- **아키텍처 옵션:** simplicity, 변경 tolerance, 운영 비용, 팀 친숙도, 실패 모드 blast radius
- **데이터 모델:** write 단순성, read 효율, migration 비용, schema evolution 가능성, integrity 보장 수준
- **알고리즘 선택:** 시간복잡도 (target input scale), 공간복잡도, 구현 복잡도, debuggability, library 가용성
- **API shape:** discoverable, misuse 어려움 (오용 방지), consistent, evolvable, debuggable
- **인증 모델:** 공격 표면, key/token lifecycle 복잡도, federation 비용, 사용자 경험, compliance 적합도
- **스택 선택:** 5년 lock-in 비용, 팀 hiring 풀, 생태계 성숙도, observability 도구, 학습 곡선
- **코드 네이밍:** 의도 명시도 (intent clarity), 일관성 (codebase convention 정합), grep-ability, callsite 가독성, 약어 회피
- **prompt engineering:** robustness (입력 변형에 강함), 길이 효율 (토큰), misuse 어려움, output 정합도, observable failure modes

3-5 기준 선택. 5 이상이면 사용자 disengage; 3 미만이면 signal이 action하기엔 너무 noisy.

### Score 해석

- **모든 variant가 전반적으로 4-5 점수** → brief가 틀림; 사용자가 공손. Push back: "이들 모두 잘 평가되지만 목표가 X라 하셨습니다 — 어떤 게 X를 실제 가장 잘 달성하고, 거기서 뭘 cut할까요?"
- **Wide spread** → 다양성 규칙 작동. 신호 있음.
- **대부분 axis 명확 승자, 한 axis 패자** → refinement 타겟 있음. 승자를 패자 axis focus로 반복.
- **승자 없음, 모든 variant가 다른 axis에서 강함** → 다음 라운드에 remix variant 고려 (A의 레이아웃 + B의 copy + C의 tone).

## 비교 Framework

나란히가 순차를 매번 이긴다. 사용자 눈과 판단은 절대 평가가 아니라 대조로 작동. 도메인 무관, axis별 비교가 싸도록 variant 제시.

엔지니어링 각 도메인 비교 framework:

- **아키텍처 옵션:** variant당 parallel 섹션, 동일 sub-header (Decomposition, Data flow, Failure modes, Cost, Migration path, Team familiarity) 의 단일 문서. Apples-to-apples 강제.
- **데이터 모델:** variant당 parallel 섹션, 동일 sub-header (Entities, R/W pattern, Normalization, Index, Migration, Risks) — design-data-model 출력 schema와 정합.
- **알고리즘 선택:** 표 — variant 행 × (시간복잡도 best/avg/worst, 공간복잡도, 구현 LOC 추정, library 가용성, debuggability) 열.
- **API shape:** variant당 parallel 섹션, 동일 sub-header (Style, Actor-Operation Map, Schema sample, Error taxonomy, Versioning, Contract test) — design-api-contract 출력과 정합.
- **인증 모델:** 표 — variant 행 × (mechanism, session TTL/rotation, federation 지원, MFA 통합 비용, recovery flow, compliance flag) 열.
- **스택 선택:** 표 — variant 행 × (언어/framework, DB, hosting, observability, CI, 5년 lock-in score, team-fit score) 열.
- **코드 네이밍:** 표 — variant 행 × (의도 명시도, codebase 일관성, grep-ability, callsite 예시 1줄, 약어 여부) 열.
- **prompt engineering:** variant당 동일 input 예시 처리 결과 나란히 + 표 (robustness, 토큰 길이, misuse 어려움, output 정합도).

Framework는 시각이 아니라 *구조*. 터미널에서도 동일 sub-header의 parallel 섹션이 대부분 일 수행.

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

## 엔지니어링 도메인 Adaptation

패턴은 동일. Artifact와 rubric 기준이 도메인별로 변경. *모두 엔지니어링 영역 — 그래픽 디자인·브랜드·마케팅·사업기획은 별도 스킬*.

### 아키텍처 옵션 (design-system / derive-system-topology 결정 시점)

- N = 2~4 (4 이상은 분석 depth 희석).
- 다양성 axis: fundamental decomposition (mono/services/serverless/hybrid), data ownership, sync vs async boundaries, build vs buy, state location (DB/cache/edge).
- 비교: 동일 sub-header (Decomposition / Data flow / Failure modes / Cost / Migration path / Team familiarity) 의 parallel 섹션.
- Rubric: simplicity, change-tolerance, ops cost, team-fit, blast radius.
- 반복: 깊은 엣지 케이스 분석으로 승자 refine, 또는 두 옵션을 hybrid로 merge.

### 데이터 모델 (design-data-model 결정 시점)

- N = 2~4.
- 다양성 axis: normalization (3NF / denormalized / event-sourced / document), PK 전략 (UUID / sequential / composite), index 전략 (covering / partial / generated column), partitioning (single / sharded / temporal).
- 비교: 동일 sub-header (Entities / R-W Pattern / Normalization / Index / Migration / Risks) 의 parallel 섹션 — design-data-model §6 출력 schema와 정합.
- Rubric: write 단순성, read 효율, migration 비용, schema evolution, integrity 수준.
- 반복: 승자 refine 또는 read/write 분리된 hybrid 검토.

### 알고리즘 선택 (build-with-tdd 또는 refactor 결정 시점)

- N = 2~4.
- 다양성 axis: 시간복잡도 트레이드오프 (hash O(1) vs sorted O(log n) vs naive O(n)), space-time trade-off, online vs batch, deterministic vs probabilistic (예: bloom filter).
- 비교: 표 — variant 행 × (시간복잡도 best/avg/worst / 공간복잡도 / 구현 LOC / library 가용성 / debuggability) 열.
- Rubric: target input scale 적합도, 구현 복잡도, debuggability, library 가용성, future-proof.
- 반복: target scale 변경 가정으로 재평가, 또는 hybrid (작은 input은 naive / 큰 input은 indexed).

### API shape (design-api-contract 결정 시점)

- N = 2~4.
- 다양성 axis: 추상화 레벨 (low-level primitive / declarative DSL / OO facade), protocol (REST / GraphQL / RPC / streaming), error semantic (status code / typed errors / Result type).
- 비교: 동일 sub-header (Style / Actor-Operation Map / Schema sample / Error taxonomy / Versioning / Contract test) 의 parallel 섹션 — design-api-contract §6과 정합.
- Rubric: discoverable, misuse 어려움, consistent, evolvable, debuggable.
- 반복: 실제 client 코드 sample 작성해 풍부도 검증, 또는 두 protocol의 hybrid (예: REST + GraphQL gateway).

### 인증 모델 (design-auth-model 결정 시점)

- N = 2~4.
- 다양성 axis: trust model (session / JWT bearer / OAuth2 federation / SAML), MFA 통합 (TOTP / WebAuthn / SMS), recovery flow.
- 비교: 표 — variant 행 × (mechanism / session TTL·rotation / federation / MFA 통합 비용 / recovery / compliance flag) 열.
- Rubric: 공격 표면, lifecycle 복잡도, federation 비용, UX, compliance 적합도.
- 반복: threat model 갱신 후 재평가, 또는 두 mechanism의 hybrid (예: session + JWT for service-to-service).

### 스택 선택 (define-tech-stack 결정 시점)

- N = 2~4 (좁히기).
- 다양성 axis: language family (Go / Rust / Python / TS / JVM), framework category (minimalist / batteries-included / opinionated), DB family (relational / document / graph / time-series), hosting (serverless / k8s / VM / managed PaaS).
- 비교: 표 — variant 행 × (언어/framework / DB / hosting / observability / CI / 5년 lock-in score / team-fit score) 열 — define-tech-stack §6 dimension table과 정합.
- Rubric: lock-in 비용, hiring 풀, 생태계 성숙도, observability, 학습 곡선.
- 반복: 5년 prediction 가정 변경하여 재평가, 또는 mixed-runtime (예: Go backend + Python ML).

### 코드 네이밍 (refactor 또는 새 모듈 추가 시점)

- N = 3~5 (네이밍은 cheap exploration).
- 다양성 axis: 의미 framework (data-shape / process / role), abstraction level (concrete / interface / metaphor), 단어 갯수 (1 / 2-3 / 4+).
- 비교: 표 — variant 행 × (의도 명시도 / codebase 일관성 / grep-ability / callsite 예시 1줄 / 약어 여부) 열.
- Rubric: intent clarity, codebase convention 정합, grep-ability, callsite 가독성, 약어 회피.
- 반복: callsite 5개에서 가독성 sample, 또는 두 후보의 axis 결합 (예: A의 framework + B의 abstraction level).

### Prompt Engineering (AI 기능 구현 시점)

- N = 3~4.
- 다양성 axis: 구조 (numbered step / prose / role-based / example-driven / chain-of-thought), 제약 레벨, 가정 reader 전문성 (zero-shot / few-shot).
- 비교: 각 variant가 처리한 동일 input 예시 나란히 + 표 (robustness / 토큰 길이 / misuse 어려움 / output 정합도).
- Rubric: variation robustness, 토큰 효율, misuse 어려움, output 정합, observable failure modes.
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
