# Skills 저장소 분석 및 통합 가이드

> 분석 일자: 2026-05-20
> 대상: `/Users/wm-it-22-00661/Work/github/study/ai/skill/` 하위 17개 저장소
> 총 스킬 수: 167개 (SKILL.md 기준)
> 목적: 카테고리 분류 + 장단점 파악 → 통합(consolidate) 의사결정 지원

---

## 0. 저장소 개요

| Repo | 스킬 수 | 분류 | 주 저자/출처 |
|------|--------|------|--------------|
| `superpowers` | 14 | 엔지니어링 프로세스 | obra (Anthropic 외부, 의견 강함) |
| `mattpocock-skill` | ~28 (deprecated 포함) | 엔지니어링 + 글쓰기 + 개인 | Matt Pocock |
| `designer-skills` | 73 | UX 학술/체계 (9 서브카테고리) | 커뮤니티 |
| `ui-ux-pro-max-skill` | 7 | UI 데이터베이스 + AI 생성 | ckm |
| `design-plugin` | 1 (design-lab) | 디자인 탐색 워크플로우 | - |
| `ui-design-brain` | 1 | 컴포넌트 패턴 DB | - |
| `make-interfaces-feel-better` | 1 | 마이크로 디테일 폴리시 | - |
| `marketingskills` | 40 | 마케팅 (고객 여정 전체) | - |
| `obsidian-skills` | 5 | Obsidian 에코시스템 | - |
| `notebooklm-skill` | 1 | NotebookLM 통합 | - |
| `skills` (Anthropic 공식) | 16 | 문서 생성 + UI + 메타 | Anthropic |
| `supabase-agent-skills` | 2 | Supabase + Postgres | Supabase |
| `vercel-agent-skills` | 8 | Vercel + React | Vercel |
| `autoresearch-skill` | 1 | Prompt 자동 최적화 | - |
| `oh-my-agentic-score` | 0 (CLI 도구) | 세션 점수화 PyPI 도구 | - |
| `awesome-claude-skills` | 0 (카탈로그) | 큐레이션 리스트 | travisvn |
| `_my_additions` (skills/) | 0 (메타 문서) | 본인 보조 자료 | 사용자 |

---

## 1. 카테고리 매핑 (9 클러스터)

| # | 카테고리 | 스킬 수 | 핵심 저장소 |
|---|----------|---------|-------------|
| A | 엔지니어링 프로세스 (워크플로우) | ~30 | `superpowers`, `mattpocock/engineering` |
| B | 디자인 - UX 학술/체계 | 73 | `designer-skills` |
| C | 디자인 - 실행/코드/도구 | ~17 | `ui-ux-pro-max`, `skills/skills/*`, `ui-design-brain`, `design-plugin`, `make-interfaces-feel-better` |
| D | 마케팅 | 40 | `marketingskills` |
| E | 지식관리·콘텐츠 라이팅 | ~12 | `obsidian-skills`, `notebooklm-skill`, `mattpocock/writing-*` |
| F | 문서/자산 생성 (Anthropic 공식) | 4 | `skills/skills/{docx,pdf,pptx,xlsx}` |
| G | 플랫폼별 개발 | 10 | `vercel-agent-skills`, `supabase-agent-skills` |
| H | 메타-스킬 (스킬·MCP 생성) | 5 | `skill-creator`, `writing-skills`, `write-a-skill`, `mcp-builder`, `autoresearch-skill` |
| I | 도구·카탈로그 (스킬 아님) | - | `oh-my-agentic-score`, `awesome-claude-skills`, `_my_additions` |

---

## A. 엔지니어링 프로세스 (~30 skills)

### A.1 superpowers ↔ mattpocock 매핑

| 영역 | superpowers | mattpocock | 통합 권장 |
|------|-------------|------------|----------|
| **TDD** | `test-driven-development` (371줄, 의식적·엄격, iron law) | `tdd` (109줄, vertical slice, public API 통합 테스트) | **superpowers** 우선, mattpocock의 "통합 테스트" 원칙 흡수 |
| **디버깅** | `systematic-debugging` (296줄, 4-phase: investigate→pattern→hypothesis→test) | `diagnose` (117줄, feedback loop 10 methods) | **둘 다 보존** (철학 다름): superpowers=프로세스, mattpocock=루프 구축법 |
| **플래닝** | `writing-plans` (152줄), `executing-plans` (70줄) | `to-prd` (76줄), `to-issues` (83줄), `request-refactor-plan` (68줄) | superpowers는 *내부* 플랜, mattpocock은 *GitHub 발행* — **결합 가능** |
| **코드 리뷰** | `requesting-code-review` (103줄), `receiving-code-review` (213줄, anti-아첨) | `review` (78줄, 2-axis), `qa` (130줄) | **superpowers** 채택 (anti-rationalization), mattpocock의 QA 통합 |
| **에이전트 협업** | `subagent-driven-development` (279줄), `dispatching-parallel-agents` (182줄) | `design-an-interface` (94줄, parallel 디자인 변형) | **superpowers** 베이스, mattpocock 케이스를 예시로 |
| **완료 검증** | `verification-before-completion` (139줄, evidence-only) | — | **superpowers 유일** — 보존 |
| **브랜치 종료** | `finishing-a-development-branch` (251줄) | — | **superpowers 유일** |
| **사고 확장** | `brainstorming` (164줄, 9-step) | `grill-with-docs` (88줄), `grill-me` (10줄), `zoom-out` (7줄) | **상보적** (general vs 도메인 grilling) |
| **위생/안전** | `using-git-worktrees` (215줄) | `git-guardrails-claude-code` (95줄, 훅 차단), `setup-pre-commit` (91줄), `setup-matt-pocock-skills` (121줄) | mattpocock이 **자동화 스크립트** 보유 — 그대로 활용 |
| **컨텍스트 핸드오프** | — | `handoff` (15줄), `caveman` (49줄, 75% 토큰 압축) | **mattpocock 유일** |
| **DDD/도메인** | — | `ubiquitous-language` (93줄), `improve-codebase-architecture` (71줄) | **mattpocock 유일** |
| **부트스트랩** | `using-superpowers` (117줄, skill 발견) | — | **superpowers 유일** |

### A.2 장단점
- **장점**: 두 저장소 모두 production-grade, 의견이 명확. superpowers는 검증 의식이 강함 (94% PR 거부율 철학). mattpocock은 issue tracker/CI/DDD 등 *팀 협업 도구* 커버.
- **단점**: TDD/디버깅/리뷰 핵심 영역에서 **철학 충돌**. superpowers는 "강제된 규율", mattpocock은 "선택의 자유". 통일하지 않으면 사용자 혼란.

### A.3 메타-스킬 부적합 항목 (deprecated/in-progress)
- `mattpocock/deprecated/`: design-an-interface, qa, request-refactor-plan, ubiquitous-language — 위 표에 통합됨
- `mattpocock/in-progress/`: review, writing-beats, writing-fragments, writing-shape — review는 A 카테고리, writing 3종은 E 카테고리로 분리

---

## B. 디자인 - UX 학술 체계 (73 skills, `designer-skills`)

9개 서브카테고리, 모두 **동일 템플릿** (frontmatter → What You Do → Best Practices), 평균 30~56줄.

| 서브카테고리 | 수 | 평균 줄 | 특성 |
|-----------|----|---------|------|
| design-ops | 9 | 56 | 가장 상세, 조직 운영 (handoff-spec, design-review-process 등) |
| interaction-design | 15 | 46 | **법칙 기반** (Fitts/Hicks/Miller/Doherty/Von Restorff) |
| ui-design | 14 | 37 | 일관성 최고, 시각 원칙 (typography, color, spacing, grid) |
| design-systems | 11 | 37 | 토큰/거버넌스 (design-token, accessibility-audit, theming) |
| design-research | 12 | 40 | 학술 UX 방법론 (JTBD, journey-map, empathy-map, interview) |
| ux-strategy | 11 | 32 | 전략 도구 (north-star, opportunity, brief, principles) |
| prototyping-testing | 8 | 39 | A/B, 휴리스틱, 사용성 |
| designer-toolkit | 7 | 50 | 케이스 스터디, 협상, UX writing |
| visual-critique | 4 | 43 | 평가 전문 (composition, typography, hierarchy, brand) |

### B.1 장단점
- **장점**: 매우 균일한 품질, 73개가 한 체계로 작동. 디자이너용 cross-functional 도구가 가장 풍부. 학술적 엄밀성.
- **단점**: **코드 산출물 0%**. 전부 "가이드/체크리스트". 실행이 아니라 *결정*에 도움. C 클러스터(실행 도구)와 짝지어야 가치.

### B.2 통합 방향
- **그대로 유지** + C 클러스터와 *명시적 링크*: 예) `ui-design/typography-scale` → `skills/ui-styling` (Tailwind 구현), `interaction-design/fitts-law` → `ui-design-brain` (버튼 사이즈 패턴)

---

## C. 디자인 - 실행/도구 (~17 skills)

### C.1 카탈로그

| 스킬 | 출처 | 줄 | 산출물 | 차별점 |
|------|------|---|--------|-------|
| **ui-ux-pro-max** | `ui-ux-pro-max-skill` | 658 | Python CLI 검색 + 추천 | 50+ 스타일, 161 컬러, 57 폰트, 99 UX 규칙, 25 차트 across 10 스택 |
| **design** (CIP) | 같음 | 302 | Gemini 생성 이미지 + 목업 | Logo 55 + CIP 50 산출물 + 배너 + 아이콘 + 소셜 사진 |
| **ui-styling** | 같음 | 324 | React TSX 컴포넌트 | shadcn/ui 사전 설치 + Tailwind 유틸리티 |
| **design-system** | 같음 | 244 | JSON 토큰 + CSS vars | 3-layer 토큰 (primitive→semantic→component) |
| **brand** | 같음 | 97 | 가이드라인 MD | 보이스/메시징/자산 검증 |
| **slides** | 같음 | 42 | Chart.js HTML 슬라이드 | 디자인 토큰 기반 슬라이드 |
| **banner-design** | 같음 | 192 | HTML/CSS 배너 + PNG | 22 art direction, AI 생성 |
| **ui-design-brain** | `ui-design-brain` | 163 | React/HTML 코드 | 60+ 컴포넌트 패턴 (component.gallery DB) |
| **frontend-design** | `skills/skills` | 42 | 작동 HTML/CSS/React | "Bold aesthetic 선택→실행" 디렉티브 |
| **make-interfaces-feel-better** | 자체 저장소 | 148 | CSS before/after diff | 12 마이크로 디테일 원칙 |
| **design-lab** | `design-plugin` | 920 | 5 variants + DESIGN_PLAN.md | **인터뷰→5 variants→피드백→플랜** (가장 긴 워크플로우) |
| **canvas-design** | `skills/skills` | 129 | PNG/PDF 아트워크 | 철학→정적 미술 |
| **algorithmic-art** | `skills/skills` | 404 | p5.js 인터랙티브 HTML | 시드 기반 생성 미술 |
| **slack-gif-creator** | `skills/skills` | 254 | 최적화 .gif | PIL 기반 Slack 최적 |
| **theme-factory** | `skills/skills` | 59 | 테마 정의 | 10 사전 큐레이션 테마 |
| **brand-guidelines** | `skills/skills` | 73 | 스타일 주입 | Anthropic 전용 (좁음) |
| **web-artifacts-builder** | `skills/skills` | 73 | 단일 HTML 번들 | React+Vite+shadcn 번들링 인프라 |

### C.2 중복 클러스터

**🔴 테마/브랜드 5종 (높은 중복)**
- ui-ux-pro-max (161 팔레트 검색)
- design (CIP 생성)
- brand (보이스 + 컴플라이언스)
- theme-factory (10 사전 큐레이션)
- brand-guidelines (Anthropic 전용)

→ **통합 권장**: "brand-vault" 하나로, 데이터베이스(ui-ux-pro-max) + 보이스(brand) + 큐레이션(theme-factory) 결합

**🟡 컴포넌트 코드 4종 (단계별 진행)**
- ui-design-brain (참조: 어떤 패턴?)
- ui-styling (구현: shadcn/ui 코드)
- frontend-design (방향: 미적 선택)
- make-interfaces-feel-better (폴리시: 디테일)

→ **유지** (서로 다른 단계), 다만 *연결 가이드* 추가

**🟡 슬라이드 3종**
- slides (Chart.js)
- design-system (토큰 기반 슬라이드)
- design 내 slide sub-skill

→ **통합 권장**: 1개 "presentations"로

**🟢 시각 아트 3종 (구분 명확)**
- canvas-design (정적 미술)
- algorithmic-art (동적 생성)
- slack-gif-creator (유틸리티)

→ **유지**

### C.3 장단점
- **장점**: 풍부한 데이터베이스 + AI 이미지 통합 + 실제 코드 생성.
- **단점**: 테마/브랜드 영역에 5개 중첩. 언제 무엇을 쓸지 모호. `ui-ux-pro-max-skill`의 6개 sub-skill이 같은 저장소 안에서도 경계가 흐림.

---

## D. 마케팅 (40 skills, `marketingskills`)

### D.1 고객 여정 + 기능 영역 매트릭스

| 그룹 | 스킬 |
|------|------|
| **Acquisition** | ab-testing, ads, ad-creative, cold-email, aso, directory-submissions, co-marketing |
| **Conversion** | cro, copywriting, copy-editing, paywalls, popups, signup, pricing, lead-magnets |
| **Retention** | churn-prevention, onboarding, emails, referrals, community-marketing |
| **Content & SEO** | content-strategy, seo-audit, ai-seo, programmatic-seo, schema, site-architecture |
| **Analytics** | analytics, customer-research, competitor-profiling, competitors, marketing-psychology |
| **Creative** | image, video, social, free-tools |
| **Strategy** | launch, marketing-ideas, product-marketing |
| **Sales** | sales-enablement, revops |

### D.2 내부 중복 후보
| 페어 | 구분 | 통합 여부 |
|------|------|----------|
| ads / ad-creative | 캠페인 관리 vs 콘텐츠 생성 | **유지** |
| competitors / competitor-profiling | 출력(페이지) vs 입력(분석) | **유지** |
| copywriting / copy-editing | 창작 vs 편집 | **유지** |
| seo-audit / ai-seo / programmatic-seo / schema | 각자 명확 | **"SEO Suite"로 통합 가능** |
| onboarding / signup | 가입 후 vs 가입 진입 | **유지** |

### D.3 장단점
- **장점**: 고객 여정 + 기능 영역 양축으로 매우 체계적. 40개가 일관된 깊이.
- **단점**: SEO 4종은 통합이 효율적. 일부(`marketing-psychology`)는 가벼운 가이드 수준.

---

## E. 지식관리·콘텐츠 (~12 skills)

| 스킬 | 출처 | 영역 |
|------|------|------|
| obsidian-cli, obsidian-markdown, json-canvas, obsidian-bases, defuddle | `obsidian-skills` | Obsidian 에코시스템 5종 |
| notebooklm | `notebooklm-skill` | NotebookLM 쿼리/소스 관리 |
| obsidian-vault | `mattpocock/personal` | 개인 볼트 (Windows 경로, 인덱싱) |
| internal-comms | `skills/skills` | 3P 업데이트, 뉴스레터, FAQ |
| doc-coauthoring | `skills/skills` | 기술 스펙, 제안서, 결정 문서 공동 작성 |
| writing-fragments, writing-beats, writing-shape, edit-article | `mattpocock` | **글쓰기 4단 파이프라인** (raw→beat→shape→edit) |

### E.1 장단점
- **장점**: mattpocock의 writing 4종은 명확한 파이프라인. obsidian-skills는 에코시스템 전체를 커버.
- **단점**: `obsidian-vault` (mattpocock 개인) vs `obsidian-skills` (일반) 간 사용자 혼선 위험.

---

## F. 문서/자산 생성 (4 skills, Anthropic 공식)

| 스킬 | 줄 | 동반 스크립트 |
|------|---|--------------|
| docx | 196 | document.py, utilities.py (XML 언팩, 리드라이닝) |
| pdf | 294 | 7개 Python 도구 (fill_fillable_fields, extract_form_fields...) |
| pptx | 483 | html2pptx.js, rearrange.py, inventory.py |
| xlsx | 288 | recalc.py 6.4KB, 재무모델 가이드 |

### F.1 장단점
- **장점**: Anthropic 유지보수, 실제 자동화 스크립트 동반, 어떤 저장소도 대체 불가.
- **단점**: 없음. **canonical 그대로 보존 권장**.

---

## G. 플랫폼별 개발 (10 skills)

### G.1 Vercel (8)
| 스킬 | 줄 | 영역 |
|------|---|------|
| deploy-to-vercel | 296 | 배포 + 팀 스코핑 + preview/prod |
| vercel-cli-with-tokens | 353 | 토큰 회전, env 설정 |
| vercel-optimize | 306 | Web Vitals, 번들 분석 |
| react-best-practices | 149 | 70 규칙, 8 우선순위 카테고리 |
| react-view-transitions | 320 | 브라우저 API 패턴 |
| react-native-skills | 121 | FlashList, Reanimated |
| composition-patterns | 89 | React 컴포넌트 패턴 |
| web-design-guidelines | 39 | 디자인 철학 (얇음) |

### G.2 Supabase (2)
| 스킬 | 영역 |
|------|------|
| supabase | 변경로그 + RLS/Auth 패턴 |
| supabase-postgres-best-practices | 쿼리 최적화, 커넥션 풀링 |

### G.3 장단점
- **장점**: 플랫폼 specifically targeted. Vercel은 토큰 관리·CLI 통합 우수. Supabase는 production 보안 체크리스트 포함.
- **단점**: `vercel-optimize` ↔ `react-best-practices` 성능 섹션 중복. `composition-patterns` (89줄), `web-design-guidelines` (39줄)는 너무 얇음.

### G.4 통합 권장
- `vercel-optimize` → `react-best-practices`에 흡수
- `web-design-guidelines` → C 클러스터(`frontend-design`)로 흡수

---

## H. 메타-스킬 (5 skills)

### H.1 비교표

| 스킬 | 줄 | 철학 | 의견 강도 | 적합 사용자 |
|------|---|------|---------|------------|
| `skill-creator` (Anthropic) | 356 | Progressive disclosure + 번들 (references/assets/scripts) | 중립 | 초보자 (원칙 학습) |
| `writing-skills` (superpowers) | 655 | **TDD for docs** (RED 베이스라인 → GREEN 스킬 → REFACTOR + subagent 압력 테스트) | 강함 | 품질 집착 엔지니어 |
| `write-a-skill` (mattpocock) | 117 | 선형 (gather → draft → review) | 약함 | 빠른 제작자 |
| `mcp-builder` (Anthropic) | 236 | 4-phase MCP 서버 빌딩 | 중립 | MCP 개발자 |
| `autoresearch-skill` | - | **자동 prompt mutation + binary eval** (유일한 self-improvement) | - | 기존 스킬 최적화 |

### H.2 파이프라인
`skill-creator` (학습) → `writing-skills` (검증) 또는 `write-a-skill` (속도) → `autoresearch` (자동 최적화)

### H.3 장단점
- **장점**: 3가지 의견 강도 스펙트럼이 상황별로 적합.
- **단점**: 사용자가 "어느 것을 쓸지" 결정 가이드 부재.

---

## I. 도구·카탈로그 (스킬 아님)

| 항목 | 타입 | 비고 |
|------|------|------|
| `oh-my-agentic-score` | PyPI CLI 도구 | Claude Code 세션 점수화 (parallelism/autonomy/density/trust), **스킬 통합 대상 아님** |
| `awesome-claude-skills` | 큐레이션 README | 외부 참고용, Feb 2026 업데이트 |
| `_my_additions` | 메타 문서 | 사용자 본인이 추가한 auto-activation 규칙·도구 모음 |

---

## 2. 통합 우선순위 종합

### 2.1 High prediction (확신 높음)
1. **메타-스킬 1개로 단일화**: `writing-skills`(검증) 베이스 + `skill-creator`(구조) + `write-a-skill`(속도 모드) 흡수
2. **C 카테고리 테마/브랜드 5종 → 1종**: ui-ux-pro-max(데이터베이스) + brand(보이스) + theme-factory(사전 큐레이션) + brand-guidelines(특정 조직) + design(생성)
3. **A 카테고리 TDD/디버깅/리뷰**: superpowers 채택, mattpocock에서 *integration test 원칙* + *loop construction 10 methods* 흡수
4. **F (Anthropic 공식 doc suite)**: 손대지 말고 canonical 유지

### 2.2 Mid prediction
5. **마케팅 SEO 4종 → "SEO Suite" 1종** (감사·AI·프로그래매틱·스키마)
6. **vercel-optimize → react-best-practices 흡수**
7. **designer-skills (73개)**: 그대로 두되, C 클러스터의 코드 실행 스킬과 *명시적 링크* 추가 (예: typography-scale → ui-styling, fitts-law → ui-design-brain)
8. **G의 web-design-guidelines → frontend-design으로 흡수**

### 2.3 Low prediction
9. mattpocock의 `caveman`(토큰 압축), `handoff`(인계), DDD 스킬 2종은 superpowers에 빠진 공백을 채움 — 함께 가져갈 가치
10. E 카테고리 writing 4종 + edit-article은 mattpocock 묶음 그대로 유지 (잘 짜여진 파이프라인)

### 2.4 None (사실)
- 전체 167 스킬 중 약 30개가 중복 후보, 나머지 ~137개는 명확히 구분됨
- Anthropic 공식(`skills/skills/`)은 자동화 스크립트 동반 비율이 가장 높음
- `designer-skills`는 코드 산출물 0%, 모두 가이드/체크리스트 형식
- mattpocock과 superpowers의 핵심 충돌점: TDD(2x), 디버깅, 리뷰, 메타-스킬 작성 — 4개 영역

---

## 3. 의사결정 체크리스트

통합 작업 시 카테고리별로 다음 질문에 답하면 의사결정이 빠름:

- [~] **A. 엔지니어링 프로세스** — *부분 완료* (13 영역 중 **완전 완료 4 / 신설 1 (원본 매핑과 다름) / 잔여 8**). **혼합 채택** (build-with-tdd = strict + tracer, diagnose-bug = 4-phase + 다중 가설, review-* = persona-mode).

  ### 완전 완료 (4 영역)
  - TDD — `build-with-tdd` 기존, strict + tracer + observable behavior 통합 흡수형
  - 에이전트 협업 — `dispatch-parallel-agents` + `pair-program-loop` 기존
  - 부트스트랩 — `router` (auto-loaded SKILL.md) 기존
  - **메타-스킬 작성 — `write-a-skill` 신설** (401줄 + references 3개 492줄, 5라운드 검증·23 fix, 2026-05-20)

  ### 신설했으나 원본 A.1 매핑과 cover 범위가 다른 영역 (1)
  - **"사고 확장"**: 원본 A.1 = superpowers `brainstorming`(idea→design) + mattpocock `grill-*`(adversarial). 신설한 두 스킬은 *다른 활동*임:
    - `verify-best-alternative` (explore-design-variants에서 rename, 2026-05-21) = *AI 편향 방지 강제 다관점 검토* (≠ brainstorming)
    - `decompose-blocker` 신설 (335줄, write-a-skill dogfooding 첫 사례, RED+REFACTOR 1 cycle pass, 2026-05-21) = *코드 작업 중 stuck 분해* (≠ brainstorming)
    - **여전히 갭**: idea→design ad-hoc pipeline (concretize-idea는 §1-한정), adversarial questioning ad-hoc (validate-advanced-edge-idea도 §1-한정)
  - **부수 작업**: 8개 §3 design 스킬(`design-system`/`define-tech-stack`/`design-api-contract`/`design-data-model`/`design-event-schema`/`design-auth-model`/`design-tenant-model`/`design-secret-management`) §11 Verification gate에 `verify-best-alternative` 호출 체크박스 wire — ADR-018로 영속화

  ### 완료 항목 내 잔여 미세 정리 (commit df4b64a, 2026-05-21)
  1. ~~**`verify-best-alternative` 본문 rewrite 불완전**~~ ✅ **A1 완료** — 본문 전체 엔지니어링 first로 재배치. Swap test·rubric·도메인 Adaptation 모두 엔지니어링 8 영역(아키텍처·데이터 모델·알고리즘·API·인증·스택·코드 네이밍·prompt engineering)으로 교체. 비엔지니어링 영역은 별도 스킬(미래)로 명시 redirect.
  2. ~~**forced invocation wire 약함**~~ ✅ **A2 완료** — 7개 §3 design 스킬 결정 Step 본문에 *AI 편향 차단 게이트* instruction 추가 (3+ orthogonal 후보 발산 의무, 첫 답 commit 금지). §11에 2차 anti-rationalization 체크박스 추가 (체크박스 회피 차단). 매핑 spec `docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md` 신설.
  3. ~~**`decompose-blocker` 검증 깊이 약함**~~ ✅ **A3 완료 (보강 형태로)** — 추가 RED test 대신 본문 보강: *언어 독립* 명시 (Python/Go/Rust/TS/Java), **3회 한도 원칙** 자동 trigger 추가 (§2.1), §4 11번째 원칙 (token escalation cap), §5 *repeated* stuck 카테고리, §7 iterate-fix-verify cascade 매핑, §8 11번째 anti-pattern (한 번 더 합리화), §5 호출 흐름 예시 2개 (인프라 + test-failure). 본문 335 → 383줄.
  4. ~~**ADR-018 commit attribution 손상**~~ ⏭️ **건너뛰기 결정** (실용 영향 0 — ADR README index + ADR 본문의 cross-reference로 audit trail 회복됨, `git log --grep` 외에는 영향 없음).

  ### 잔여 3 영역 (원본 A.1 매핑 미해소)
  - ~~**HIGH**: `review-engineering`에 anti-rationalization 원칙 본문 보강~~ ✅ **완료 (commit 830eb04)** — Eng Manager Persona 직후 신규 섹션 5 규칙: (1) 금지 발화 vocabulary (looks good/approved/comprehensive 등 차단) (2) "no issues found" 증명 책임 (hypothetical failure ≥1 + 미검토 영역 명시) (3) 7/7 all-clean STOP rule (sycophancy 신호 우선 가정) (4) 작성자 답변 verify 의무 (file:line 위치 강제) (5) dimension-당 challenge 자증명. Output 섹션에 7번째 산출물 (anti-rationalization self-attestation MET/NOT MET 체크) 추가. receiving-code-review (superpowers) 받는 쪽 패턴을 *주는 쪽* 시점으로 대칭 이식.
  - **MID**:
    - ~~디버깅: `diagnose-bug/references/loop-methods.md`에 mattpocock의 *loop construction 10 methods* 흡수~~ ✅ **완료 (commit 4754e84)** — 신규 reference (234줄, 8 섹션): disproportionate-effort 원칙 + **10 methods 메뉴** (failing test / curl·HTTP / CLI fixture / Playwright headless / replay trace / throwaway harness / property-fuzz / bisection / differential / HITL) + iterate-on-loop (faster/sharper/deterministic, 30초 flaky→2초 deterministic) + non-deterministic 처리 (재현률 < 5% debug 불가능, 50%+ 목표) + **cannot-build-a-loop STOP 3-STEP** (시도 method 목록 명시 + 사용자 자료 요청 + Phase 1+1.5 합산 3 회 실패 시 decompose-blocker 자동 trigger) + Phase 1↔1.5↔2↔3 cascade 다이어그램 + anti-pattern 6개. PROCEDURE.md Phase 1 제목을 "Reproduce — Build a feedback loop"로 격상, 본문 + Phase 1.5에 cross-link 4 곳 삽입. ADR-003 inspired-by (verbatim 0건, buddy 도메인 어휘 재진술). A3 decompose-blocker (token escalation cap) + MID-1 verification-discipline 정합.
    - ~~플래닝: GitHub 발행 패턴(`to-prd`, `to-issues`) `plan-build` 또는 신규 스킬로 흡수~~ ✅ **완료 (commit 5c9fba5)** — 신규 스킬 `publish-to-tracker` (280줄): 2 모드 (`prd` §2 feature spec → PRD 1건 / `issues` §4 task plan → tracer-bullet vertical slice N건). multi-tracker 추상화 (GitHub `gh` CLI / Linear API / Jira API, 4-method adapter interface). HITL/AFK 라벨 + `ready-for-agent` surface (dispatch-parallel-agents cascade). 의존성 순서 발행 (topological, blocker 먼저 → 후속 issue Blocked-by binding → parent PRD 역참조). `--start-from` 부분 발행 시 미발행 blocker STOP. 시크릿 env 변수만 (`design-secret-management` 정합). 발행 발화 직전 Iron Law (verification-discipline MID-1 정합). cross-link 2 곳 (`define-features` 옵션 PRD 발행 / `plan-build` Stage 8 옵션 issues 발행). catalog §4 등록 + command file + test count 149→150. ADR-003 adopt-with-edits (PRD/Issue template + tracer-bullet rules + HITL/AFK) + inspired-by (multi-tracker 추상화는 buddy 자체 발명).
    - ~~완료 검증: superpowers `verification-before-completion` 원칙을 패턴 라이브러리화~~ ✅ **완료 (commit 1523e65 — 병행 작업과 통합)** — `plugin/skills/router/references/verification-discipline.md` (182줄) 신규 SSoT, 10 섹션 (Iron Law / Gate Function 5-step / 도메인별 evidence 매핑 / Red Flags / Rationalization Prevention / Key Patterns / When To Apply / Cross-link 표 / A-카테고리 정합성 / Bottom Line). 4개 스킬에 cross-link 1줄씩: `build-with-tdd` §4 핵심 원칙 #11 (RED/GREEN/REFACTOR 전이), `iterate-fix-verify` Step 5 직전 (verified/best-effort/reverted 분류), `verify-quality` Quality Gate 직전 (review-engineering All-clean STOP rule과 정합), `build-feature` 완료 기준 직전 (subagent dispatch success 채택). `skill-catalog.md` §5 참조에 cross-skill SSoT 등록. ADR-003 inspired-by 분류 (verbatim 0건). A-카테고리 anti-rationalization 3-layer 게이트 (결정 A2 + 리뷰 HIGH + **완료 MID-1**) 완성.
    - ~~브랜치 종료: `finish-development-branch` orchestrator 신설~~ ✅ **완료 (commit 54efc0f)** — Priority-2 sub-orchestrator 신설 (275줄, 5 stage: pre-flight sync + quality-gate + changelog + docs-sync + PR 생성 + mergeable verify). **자동화 시스템 git 안전 원칙 동시 신설** — `plugin/skills/router/references/git-safety-rules.md` (188줄, cross-skill SSoT): force-류 (force / force-with-lease / rebase-pushed / amend-pushed / reset --hard / clean -fd / branch -D / submodule force / filter-branch / aggressive gc) 자동 실행 절대 금지 + 안전 대체 명령 + STOP 절차 (push reject / merge conflict / dirty tree / stash conflict). guard-destructive-commands 의 force-with-lease 안전 분류 *제거* (line 39+ 표기 + line 177 hook regex 확장). 5개 스킬 cross-link (auto-create-pr Phase 1/3, dispatch-parallel-agents Phase 7, ship-release §7-1, build-feature 다음 phase, guard-destructive-commands 본문). A-카테고리 4-layer 게이트 완성 (결정/리뷰/완료발화/git실행). **신규 follow-up 등록**: parallel agent 충돌 방지 (시나리오 2) — `decompose-track-to-tasks` expected_touched_files 필드 + `map-task-dependencies` file-overlap edge 자동 추가 + `plan-parallel-execution` same-wave overlap=0 검증 + `dispatch-parallel-agents` worktree base periodic refresh. (별도 follow-up 4 skill 보강)
  - **LOW**:
    - ~~DDD/도메인: mattpocock `ubiquitous-language` 검토 (buddy actor 모델과 보완)~~ ✅ **완료 (commit `cd99818`, 2026-05-26)** — `audit-ubiquitous-language` 신규 skill, inspired-by DDD theory (Evans 2003 / Vernon 2013). 흡수 분류 변경 = adopt-with-edits → inspired-by (mattpocock 본문 cross-machine 부재). engineering-flow H1+H3 hole 동시 closed. 상세 분석: `docs/plugin-skills-engineering-audit.md` §2.4.
    - ~~컨텍스트 핸드오프: mattpocock `caveman`(75% 토큰 압축) 검토 (cli-buddy token monitor와 시너지)~~ 🔄 **cli buddy 트랙 이동 확정 (사용자 명시 2026-05-26, Path 1 진입)** — runtime token state 영역, plugin engineering audit 범주 제외. cli buddy Wave 7 W7-2 trigger 시 검토. 상세: `docs/plugin-skills-engineering-audit.md` §2.3.
    - ~~위생/안전: mattpocock 자동화(setup-pre-commit, git-guardrails) 검토 (buddy compose-safety-mode와 비교)~~ ✅ **종결 확정 (흡수 안 함) — 사용자 명시 2026-05-26, Path 1 진입**. buddy `setup-quality-gates` / `guard-destructive-commands` / `git-safety-rules.md` 가 동등 또는 우월 cover. 상세: `docs/plugin-skills-engineering-audit.md` §2.2.
- [ ] **B. designer-skills 73개**: 전부 유지 vs 핵심 ~30개로 압축?
- [ ] **C. 테마/브랜드 5종**: 어느 하나를 *프라이머리*로 정할지? (ui-ux-pro-max 권장)
- [ ] **D. 마케팅 SEO 4종**: 통합 vs 유지?
- [ ] **E. obsidian**: 개인 vault와 일반 skills를 둘 다 유지하나?
- [ ] **F. doc suite**: 변경 없음 — 확인만
- [ ] **G. Vercel 8종**: 얇은 2개(`composition-patterns`, `web-design-guidelines`)를 어디로 보낼지?
- [ ] **H. 메타-스킬**: 단일화 vs 3-tier 유지?
- [ ] **I. 도구**: oh-my-agentic-score는 별도 운영 인정?

---

## 부록 A. 전체 스킬 목록 (167)

상세 목록은 다음 명령으로 재생성:
```bash
find /Users/wm-it-22-00661/Work/github/study/ai/skill -name "SKILL.md" -type f | sort
```

## 부록 B. 참고 자료

- Anthropic 공식 skills 저장소: `skills/`
- skill 작성 가이드: `superpowers/skills/writing-skills/SKILL.md`, `skills/skills/skill-creator/SKILL.md`
- 큐레이션 카탈로그: `awesome-claude-skills/README.md`
