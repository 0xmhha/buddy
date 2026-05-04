---
name: consult-design-system
description: "research → synthesize → output pipeline으로 complete design system 생성. aesthetic, typography, color, spacing, motion tokens + 컴포넌트 hierarchy. 트리거: '디자인 시스템 제안' / '토큰 정의' / '컴포넌트 시스템 만들자' / 'design tokens' / 'spacing/color 시스템' / 'design system consult' / '디자인 시스템 컨설트'. 입력: 제품 컨텍스트, 브랜드, 기존 컴포넌트. 출력: design tokens + 컴포넌트 시스템 + 사용 가이드. 흐름: review-design → consult-design-system → 구현."
type: skill
---

# Design-Consultation — Research-Synthesize-Output 파이프라인


이 스킬은 모호한 문제 ("디자인 시스템 필요")를 coherent, opinionated, 작성된 artifact ("DESIGN.md가 여기, 모든 섹션 정당화, 모든 선택 traceable")로 바꾸는 3-phase 파이프라인을 패키징.

파이프라인은 시각 디자인에 **특화 아님**. Generic shape:

1. **Research** — 인터뷰로 제품 DNA 추출, 그다음 지형 매핑.
2. **Synthesize** — 차원별로 근거와 함께 coherent 시스템 제안.
3. **Output** — 구조화 템플릿 사용해 단일 canonical 파일 작성.

같은 shape가 디자인 시스템, 기술 문서, 아키텍처 결정 기록 (ADR), 정책 문서에 작동. 끝의 *Domain Adaptation* 참조.

스킬은 form-wizard posture가 아니라 **consultant posture** 채택: 완전 시스템 제안, 작동 이유 설명, pushback 초대. Opinionated이지만 dogmatic 아님.

---

## 이 스킬을 사용하는 경우

- 기존 디자인 시스템, 문서 set, 컨벤션 없는 새 제품 시작, 공식화 필요.
- 분산된 컨벤션 존재 (일부 색상 선택 여기, 폰트 거기, wiki에 묻힌 ADR), 단일 source of truth로 통합 원함.
- 팀이 성장할 예정이고 newcomer가 재도출 아니라 결정 상속 필요.
- Taste와 care 신호하는 "첫인상" artifact (DESIGN.md, ARCHITECTURE.md, CONTRIBUTING.md) 원함.

다음 경우 이 스킬 skip:

- Canonical 문서 이미 존재하고 따르는 중. (리뷰 스킬로 audit 대신.)
- 제품이 진짜 memorable "thing" 가질 만큼 early — 먼저 사용자 리서치 할 것.

---

## 3-Phase 파이프라인

### Phase 1: Research

Research의 목표는 Phase 2에서 **opinionated 권리를 earn**. Research 없이 synthesis는 개인 taste. Research 있으면 defensible 추천.

Research는 두 반: 제품 이해, 그다음 지형 이해.

#### 1a. 제품 이해 인터뷰

소수의 high-leverage 질문. Codebase (README, package.json, 디렉토리 구조, 이전 artifact)에서 추론 가능한 것 pre-fill. 심문 금지. 인터뷰는 한두 질문, 10 아님.

**네 product-DNA 질문:**

1. **뭐고, 누구 대상, 어떤 space?**
   - 한 줄 제품 설명 confirm.
   - 타겟 사용자 명명 (직함, 숙련도, 멘탈 모델).
   - 카테고리 / 산업 / peer set 명명.
   - *README에서 pre-fill하고 빈 질문 대신 confirm.*

2. **어떤 타입 artifact?**
   - 웹 앱, 대시보드, 마케팅 site, editorial, 내부 tool, CLI 문서, ADR, 정책. Artifact 타입이 어떤 차원 중요한지 결정.
   - 문서용: reference vs 튜토리얼 vs explanation vs how-to.
   - ADR용: greenfield 결정 vs retrofit vs migration.

3. **Research 허락?**
   - "최상위 제품 / 유사 시스템이 뭐 하는지 research할까, 내 지식으로 작업할까?"
   - Research 비쌈. 일부 사용자는 속도와 당신 판단 신뢰 원함. 다른 사람은 defensibility 원하고 먼저 look around 원함.

4. **대화형 escape hatch.**
   - 명시 진술: "언제든 채팅으로 drop해서 뭐든 얘기 가능. 이건 엄격한 form이 아니라 대화."

**Forcing 질문 (항상 질문):**

> "누군가 이걸 처음 만난 뒤 기억했으면 하는 한 가지?"

한 문장 답. 느낌 ("이건 진지한 작업용 진지한 소프트웨어"), 시각 ("거의 검은색인 파랑"), 주장 ("다른 것보다 빠름"), posture ("관리자 아니라 빌더용"), 또는 문서용 ("이걸 읽고 5분 만에 뭔가 ship 가능") 될 수 있음.

기록. Phase 2의 후속 모든 결정이 이 memorable thing 봉사해야. 모든 것에 memorable하려는 시스템은 아무것도 memorable 안 함.

#### 1b. 지형 Research

사용자가 질문 3에 yes하면만 이 phase 실행.

**3-layer synthesis** (같은 구조가 디자인 레퍼런스, 문서 peer, 아키텍처 패턴에 작동):

- **Layer 1 — Tried and true:** 이 카테고리의 모든 제품 / 문서 / 시스템이 공유하는 패턴? Table stakes — 사용자 기대. Layer 1 skip하면 incompetent 보임. Layer 1 따르는 건 바닥, 천장 아님.

- **Layer 2 — New and popular:** 현재 담론이 뭐 말하나? 어떤 새 패턴 emerging? 회의적으로. 인기가 정확성 아님.

- **Layer 3 — First principles:** 이 제품의 사용자와 포지셔닝에 대해 아는 것 주어지면, 관습 접근이 *여기* 틀린 이유 있나? 어디서 의도적으로 카테고리 규범 깨야?

**Eureka 규칙:** Layer 3 reasoning이 관습 지혜 모순할 때 명시 명명:

> "모든 [카테고리] 제품이 [가정] 가정해서 X 함. 하지만 이 제품 사용자는 [증거] — 그래서 대신 Y 해야."

Eureka 순간이 Phase 1의 가장 가치 있는 output. Phase 2의 리스크 선택 정당화.

**우선순위 순 research 방법:**

1. **직접 검사.** 브라우저 tool 있으면 상위 3-5 레퍼런스 site 방문, 스크린샷 (느낌)과 구조 snapshot (뼈대) 둘 다 캡처. 문서엔 소스가 아니라 실제 렌더링 페이지 읽기. ADR엔 결정과 결과 코드 둘 다 읽기.

2. **검색.** "[카테고리] best [artifact-type] 2025", "[카테고리] design patterns", "[카테고리] documentation examples" 웹 검색.

3. **내장 지식.** 도구와 검색 사용 불가 때 당신 훈련 데이터가 상당한 Layer 1과 Layer 2 coverage. 의존. 모르는 것에 정직.

**Research output:** 리포트가 아니라 짧은 대화형 요약. 본 패턴 명명, 수렴 ("모두 X 사용"), 차이 ("소수 outlier가 Y 함"), gap ("아무도 Z 안 함, 흥미로운 이유..."). 이 요약이 Phase 2에 직접 공급.

---

### Phase 2: Synthesize

Synthesis phase는 research 발견을 사용자가 전체로 반응할 수 있는 **완전하고 coherent 제안**으로 변환. 메뉴나 객관식 그리드 제시 금지. 한 시스템 제안. 왜 모든 piece가 모든 다른 piece 강화하는지 설명. 그다음 사용자에게 특정 차원 조정 초대.

아래 차원들은 디자인 시스템에 도메인 특정. **구조** — 각각 추천과 근거 있는 이름 붙은 차원의 작은 set — 는 어떤 도메인에든 일반화 (*Domain Adaptation* 참조).

#### 2a. 미적 방향

미적 방향이 load-bearing 결정. 다른 모든 차원이 강화하거나 맞섬.

선택할 인식 가능한 미적 방향의 짧은 메뉴:

- **Brutally Minimal** — 타입과 여백만. 장식 없음. Modernist.
- **Maximalist Chaos** — 조밀, 레이어, 패턴 heavy. Y2K 현대 만남.
- **Retro-Futuristic** — 빈티지 tech 향수. CRT glow, 픽셀 그리드, 따뜻한 monospace.
- **Luxury / Refined** — Serif, 고 contrast, 관대한 여백, 귀금속.
- **Playful / Toy-like** — Rounded, bouncy, bold primary. Approachable, fun.
- **Editorial / Magazine** — 강한 타이포 hierarchy, 비대칭 그리드, pull quote.
- **Brutalist / Raw** — 노출 구조, 시스템 폰트, 가시 그리드, no polish.
- **Art Deco** — 기하 정밀, 금속 액센트, 대칭, 장식 border.
- **Organic / Natural** — 흙 톤, rounded form, 손그림 텍스처, grain.
- **Industrial / Utilitarian** — Function-first, data-dense, monospace 액센트, muted 팔레트.

**장식 레벨:** minimal (타이포가 모든 일) / intentional (미묘 텍스처, grain, 배경 처리) / expressive (전체 창의 방향, 레이어 깊이, 패턴).

#### 2b. 타이포그래피

**역할당 한 폰트** pick, 옵션 list 아님. 사용자는 항상 override 가능; decisive 단일 추천 default.

채울 역할:
- **Display / Hero** — 프론트페이지 personality.
- **Body** — reader가 대부분 시간 쓰는 텍스트.
- **UI / Labels** — 버튼, form field, navigation. 종종 body와 동일.
- **Data / Tables** — 컬럼 정렬 위해 tabular numbers (`tabular-nums`) 지원 필수.
- **Code** — 기술 콘텐츠용 monospace.

**역할별 폰트 shortlist** (여기서 시작, 의도적 deviate):
- Display / Hero: Satoshi, General Sans, Instrument Serif, Fraunces, Clash Grotesk, Cabinet Grotesk
- Body: Instrument Sans, DM Sans, Source Sans 3, Geist, Plus Jakarta Sans, Outfit
- Data / Tables: Geist (tabular-nums), DM Sans (tabular-nums), JetBrains Mono, IBM Plex Mono
- Code: JetBrains Mono, Fira Code, Berkeley Mono, Geist Mono

**절대 추천 금지** (폰트 블랙리스트): Papyrus, Comic Sans, Lobster, Impact, Jokerman, Bleeding Cowboys, Permanent Marker, Bradley Hand, Brush Script, Hobo, Trajan, Raleway, Clash Display, Courier New (body용).

**Primary로 절대 추천 금지** (수렴 함정 — 모든 AI 디자인 tool이 여기 gravitate, 사용은 "taste 발휘 안 함" 신호): Inter, Roboto, Arial, Helvetica, Open Sans, Lato, Montserrat, Poppins, Space Grotesk. 사용자가 이름으로 요청하면만 사용.

**타입 scale:** 레벨당 구체 px나 rem 값 있는 modular scale 명시 (예: 12 / 14 / 16 / 20 / 24 / 32 / 48 / 64). 단계 list 없이 "modular scale 사용"만 말하지 말 것.

#### 2c. 색상

**접근:** restrained (1 액센트 + neutral, 색상이 rare하고 의미) / balanced (primary + secondary + semantic) / expressive (색상이 주 디자인 tool, bold 팔레트).

**팔레트 구성** — 구체 hex 값 제안:
- **Primary** — 브랜드 색상. 대표하는 것과 나타나는 곳 진술.
- **Secondary** — 지원 색상. 사용 규칙 진술.
- **Neutral** — 따뜻 또는 차가운 회색, 가장 밝음에서 가장 어둠까지 full range (5-9 단계).
- **Semantic** — success / warning / error / info hex 값. Skip 금지.
- **Dark 모드 전략** — surface 재디자인 (invert만 말고), saturation 10-20% 감소, WCAG AA contrast 확보.

**거부할 안티패턴:**
- 기본 액센트로 보라 / violet 그라디언트 (AI slop 신호).
- Primary CTA로 그라디언트 버튼.
- 제안된 배경에 WCAG AA 실패하는 색상 선택.

#### 2d. 레이아웃 & 간격

**레이아웃 접근:** grid-disciplined (엄격한 컬럼, 예측 정렬) / creative-editorial (비대칭, 겹침, 그리드 깨기) / hybrid (앱엔 그리드, 마케팅엔 creative).

**그리드:** breakpoint별 컬럼, max 콘텐츠 너비.

**간격 scale** — 기본 단위 (4px 또는 8px)와 밀도 (compact / comfortable / spacious). Scale 구체 명시:

```
2xs(2) xs(4) sm(8) md(16) lg(24) xl(32) 2xl(48) 3xl(64)
```

**Border radius:** 계층 scale (예: sm:4px, md:8px, lg:12px, full:9999px). 모든 element에 uniform bubbly radius 피하기 — AI slop 시그니처.

#### 2e. 모션

**접근:** minimal-functional (이해 돕는 전환만) / intentional (미묘 진입 애니메이션, 의미 상태 전환) / expressive (전체 choreography, 스크롤 기반, playful).

**Easing:** enter `ease-out`, exit `ease-in`, move `ease-in-out`.

**Duration 버킷:** micro (50-100ms), short (150-250ms), medium (250-400ms), long (400-700ms). 700ms 이상은 broken 느낌.

#### Synthesis output: SAFE / RISK breakdown

크리티컬 움직임. Coherence 혼자는 table stakes — 모든 제품이 coherent일 수 있고 여전히 peer와 동일 보임. 흥미로운 질문은 **어디서 리스크 취하나?**

제안을 두 bucket으로 split 제시:

```
SAFE CHOICES (카테고리 baseline — 사용자 기대):
  - [카테고리 컨벤션 매치하는 2-3 결정, 각각 근거 포함]

RISKS (제품이 자기 얼굴 얻는 곳):
  - [의도적 컨벤션 이탈 2-3]
  - 각 리스크에: 뭔지, 왜 작동, 뭘 얻나, 뭘 비용]
```

Safe 선택이 카테고리에 literate 유지. 리스크가 제품 memorable 만듦. 항상 최소 두 리스크 제안, 각각 Phase 1의 *memorable thing*에 묶임.

#### 일관성 검증

사용자가 한 차원 override할 때 나머지가 여전히 coherent인지 체크. Mismatch를 gentle nudge로 flag — 절대 block 금지:

- Brutalist 미학 + expressive 모션 → "Brutalist는 보통 minimal 모션과 페어. 당신 combo는 unusual, intentional이면 괜찮음."
- Expressive 색상 + restrained 장식 → "Bold 팔레트와 minimal 장식 작동 가능, 하지만 색상이 많은 무게 운반."
- Creative-editorial 레이아웃 + data-heavy 제품 → "Editorial 레이아웃은 data 밀도와 맞설 수 있음. Hybrid가 둘 다 유지."

항상 사용자 최종 선택 수락. Nudge, block 아님.

---

### Phase 3: Output

Output은 구조화 템플릿 사용해 작성된 단일 canonical 파일. 템플릿의 일은 모든 결정을 **discoverable, justifiable, revisable**하게. 모든 섹션이 선택, 근거, (유용한 곳) 고려한 대안 명명.

#### Output 템플릿 (DESIGN.md 또는 동등물)

```markdown
# Design System — [Project Name]

## Product Context
- **What this is:** [1-2 sentence description]
- **Who it's for:** [target users — job, skill level, mental model]
- **Space / industry:** [category, peer set]
- **Project type:** [web app / dashboard / marketing site / editorial / internal tool]
- **Memorable thing:** [the one-sentence answer from the Phase 1 forcing question]

## Aesthetic Direction
- **Direction:** [name from the menu]
- **Decoration level:** [minimal / intentional / expressive]
- **Mood:** [1-2 sentence description of how the product should feel]
- **Reference sites:** [URLs from Phase 1 research, if any]

## Typography
- **Display / Hero:** [font name] — [rationale]
- **Body:** [font name] — [rationale]
- **UI / Labels:** [font name or "same as body"]
- **Data / Tables:** [font name] — [rationale, must support tabular-nums]
- **Code:** [font name]
- **Loading strategy:** [CDN URL, self-hosted, or system fallback]
- **Scale:** [modular scale with concrete px / rem values per level]

## Color
- **Approach:** [restrained / balanced / expressive]
- **Primary:** [hex] — [what it represents, where it appears]
- **Secondary:** [hex] — [usage rule]
- **Neutrals:** [warm or cool grays, hex range from lightest to darkest]
- **Semantic:** success [hex], warning [hex], error [hex], info [hex]
- **Dark mode:** [strategy — redesign surfaces, reduce saturation 10-20%, contrast notes]

## Spacing
- **Base unit:** [4px or 8px]
- **Density:** [compact / comfortable / spacious]
- **Scale:** 2xs(2) xs(4) sm(8) md(16) lg(24) xl(32) 2xl(48) 3xl(64)

## Layout
- **Approach:** [grid-disciplined / creative-editorial / hybrid]
- **Grid:** [columns per breakpoint]
- **Max content width:** [value]
- **Border radius:** [hierarchical scale — sm:4px, md:8px, lg:12px, full:9999px]

## Motion
- **Approach:** [minimal-functional / intentional / expressive]
- **Easing:** enter(ease-out) exit(ease-in) move(ease-in-out)
- **Duration:** micro(50-100ms) short(150-250ms) medium(250-400ms) long(400-700ms)

## Risks Taken
For each deliberate departure from category convention:
- **[Risk name]:** [what it is]
- **Why it works:** [rationale tied to the memorable thing]
- **What you gain:** [the upside]
- **What it costs:** [the downside, honestly named]

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| [today] | Initial design system created | [Phase 1 product context, Phase 1b research findings or "agent built-in knowledge"] |
```

**동반 파일** — 프로젝트 주 instruction 파일(CLAUDE.md, AGENTS.md, README.md, 프로젝트 컨벤션에 의존)에 라우팅 규칙 append:

```markdown
## Design System
Always read DESIGN.md before making any visual or UI decisions.
All font choices, colors, spacing, and aesthetic direction are defined there.
Do not deviate without explicit user approval.
```

**최종 confirm 단계:** 쓰기 전 모든 결정 source와 함께 list. 명시 사용자 confirm 없이 agent default 사용한 결정 flag — 사용자가 정확히 뭐 shipping하는지 알아야. 세 옵션 제공: ship, 특정 것 변경, 처음부터.

---

## Domain Adaptation

Research → synthesize → output 파이프라인은 목표가 분산되거나 부재 컨벤션에서 coherent canonical 문서 생산인 어떤 도메인에도 일반화. Phase는 동일 유지; 변경되는 건 Phase 2의 **차원**과 Phase 3의 **템플릿**.

### 기술 문서

**Phase 1 — Research**
- 제품 DNA: 뭐 프로젝트, 누가 문서 읽음, 어떤 숙련도, 어떤 task 가져왔나.
- Forcing 질문: "새 reader가 문서 오픈 후 10분 안에 뭐 할 수 있어야?"
- 지형: 3-5 peer 프로젝트 문서 읽기. 구조 (Diátaxis, README-only, 전체 reference site), 톤 (formal / friendly / terse), 진입점 디자인 주목.
- Layer 3 eureka: peer 문서가 이 audience에 어디서 실패?

**Phase 2 — Synthesize (문서 차원)**
- **정보 아키텍처** — Diátaxis (tutorial / how-to / reference / explanation), README-only, monorepo navigation, search-first.
- **톤** — formal / friendly / terse / instructional. 하나 pick하고 유지.
- **코드 예시 밀도** — sparse (prose-led) / balanced / example-led.
- **길이 컨벤션** — 페이지당 엄격 단어 상한, 또는 unbounded.
- **버전 전략** — 현재 버전 단일 source vs multi-version site.
- **발견 surface** — 사이드바 nav, 검색 bar, top-level 인덱스 페이지.
- SAFE / RISK split: peer 컨벤션 작동하는 곳 copy; audience가 정당화하는 곳 이탈 (예: power user용 terse reference, novice용 더 많은 prose).

**Phase 3 — Output 템플릿 (DOCS.md 또는 기여자 가이드)**
```markdown
# Documentation Conventions — [Project Name]
## Audience
## Information Architecture
## Tone & Voice
## Code Example Density
## Page Length & Structure
## Versioning Strategy
## Discovery Surface
## Entry Points
## Risks Taken
## Decisions Log
```

### 아키텍처 결정 기록 (ADR)

**Phase 1 — Research**
- 제품 DNA: 어떤 시스템, 어떤 scale (사용자, 요청 볼륨, 데이터 크기), 어떤 팀 크기, 어떤 제약 (latency 예산, 규제, 레거시).
- Forcing 질문: "다음 12개월에 절대 깨지면 안 되는 아키텍처 속성?"
- 지형: 유사 scale의 peer 시스템이 어떻게 해결했는지 research. 뭐 후회? 공개 post-mortem 금.
- Layer 3 eureka: 관습 지혜가 이 구체 scale이나 제약에 어디서 실패?

**Phase 2 — Synthesize (ADR 차원)**
- **결정** — 선택 자체, 한 문장에 명명.
- **상태** — proposed / accepted / deprecated / superseded.
- **Context** — 어떤 힘 작용 (기술, 조직, 규제).
- **고려된 대안** — 최소 둘, 각각 한 줄 dismissal 근거.
- **결과** — positive와 negative, 둘 다 정직 명명.
- **Compliance** — 이걸 따르는지 어떻게 알까? 어떤 테스트나 리뷰가 위반 catch?
- SAFE / RISK split: 이 결정의 어느 부분이 산업 default 따름, 어느 부분이 의도적 break.

**Phase 3 — Output 템플릿 (ADR-NNN-title.md)**
```markdown
# ADR-NNN: [Title]
## Status
## Context
## Decision
## Alternatives Considered
- [Alternative A] — rejected because [reason]
- [Alternative B] — rejected because [reason]
## Consequences
### Positive
### Negative
## Compliance
## Risks Taken
## Decisions Log (revisions)
```

### 정책 문서

**Phase 1 — Research**
- 제품 DNA: 어떤 조직, 어떤 관할, 누가 정책에 bound, 뭐가 필요 trigger.
- Forcing 질문: "이 정책이 바꿔야 할 한 행동, 6개월 뒤 성공이 어떻게 보이나?"
- 지형: 같은 도메인 peer 정책 읽기. Enforcement 메커니즘, 예외 process, 리뷰 cadence 주목.
- Layer 3 eureka: peer 정책이 실제 enforcement에서 어디 실패?

**Phase 2 — Synthesize (정책 차원)**
- **Scope** — 누구와 뭐 커버, 어디 적용, 언제 효력.
- **필수 행동** — 사람이 반드시 해야 할 것.
- **금지 행동** — 사람이 반드시 하지 말아야 할 것.
- **허용 예외** — 누가 grant, 어떤 process, 어떻게 기록.
- **Enforcement 메커니즘** — 자동 체크, 수동 리뷰, audit cadence.
- **위반 결과** — 비례적, 명명, traceable.
- **리뷰 cadence** — 이 정책 언제 revisit.
- SAFE / RISK split: 표준 규제 언어 vs 조직별 입장.

**Phase 3 — Output 템플릿 (POLICY-name.md)**
```markdown
# Policy: [Name]
## Scope
## Effective Date
## Required Behaviors
## Prohibited Behaviors
## Permitted Exceptions
## Enforcement Mechanism
## Consequences for Violation
## Review Cadence
## Risks Taken (deliberate non-standard positions)
## Decisions Log (revisions)
```

---

## 중요 규칙 (모든 도메인 걸쳐 운반)

1. **제안, 메뉴 제시 말 것.** Consultant지 form 아님. Research 기반 차원당 한 opinionated 추천, 그다음 사용자가 특정 piece 조정.
2. **모든 추천에 근거 필요.** "Y 때문에" 없이 "X 추천" 절대 금지. 근거가 사용자 편집 cycle 살아남는 것.
3. **개별 최적 선택보다 coherence.** 모든 piece가 모든 다른 piece 강화하는 시스템이 개별 "best"지만 mismatch 선택 시스템 이김.
4. **Memorable thing이 load-bearing.** 모든 Phase 2 결정이 Phase 1 forcing-질문 답으로 trace. 결정이 거기 묶일 수 없으면 방향이 아니라 장식.
5. **SAFE / RISK split은 non-optional.** Coherence만으로는 구별 안 되는 peer 생산. 리스크가 artifact memorable 만듦. 항상 최소 두 리스크 제안, 각각 정직 비용.
6. **대화 톤.** 이건 워크플로우 아니라 파트너십. 사용자가 결정 얘기 원하면 사려 깊은 파트너로 engage.
7. **사용자 최종 선택 수락.** Coherence 이슈에 nudge, 하지만 disagree해 artifact 쓰기 block이나 refuse 금지.
8. **자기 output에 slop 없음.** 추천, 미리보기, 최종 문서 — 모두 사용자에게 채택 요청하는 taste 시연해야.
