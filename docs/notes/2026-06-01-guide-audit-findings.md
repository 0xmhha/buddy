# Guide Audit Findings — 2026-06-01

> **목적**: `docs/plugin-skills-authoring-guide.md`와 `plugin/skills/router/references/engineering-phases.md` 두 문서를 Claude Code/Claude 모델 동작 적합성, lost-in-the-middle 위험, 논리적 결함, 누락 항목 관점에서 감사한 결과를 영속화.
>
> **배경**: 두 문서는 `evaluate-skill`(평가)과 `write-a-skill`(작성)의 기준 SSoT 역할. 이 두 문서가 흔들리면 모든 skill 평가·작성 작업의 신뢰성이 무너짐 → 선행 검증 필요.
>
> **상태**: 감사 완료. 🔴 High 3건 즉시 수정 진행 중. 🟡 Mid/🟢 Low는 후속.

---

## 0. 작업 컨텍스트

| 항목 | 값 |
|------|-----|
| 감사 대상 1 | `docs/plugin-skills-authoring-guide.md` (735줄, 최초 2026-05-29 작성) |
| 감사 대상 2 | `plugin/skills/router/references/engineering-phases.md` (711줄, 2026-05-26 작성 / 2026-05-29 최종 수정) |
| 감사 일자 | 2026-06-01 |
| 감사 기준 | (1) Claude Code/Claude 모델 호환성, (2) lost-in-the-middle 방지, (3) 논리적 일관성, (4) 누락 점검 |

---

## 1. 우선순위 분류

| Severity | 의미 | 건수 | 진행 |
|---------|------|------|------|
| 🔴 High | 잘못된 사실/심각한 모호성 — 즉시 수정 | 3 | ✅ 3/3 완료 |
| 🟡 Mid | 운영 영향 있는 결함 — 다음 사이클 | 20 | ✅ 20/20 완료 |
| 🟢 Low | 미세 개선 — 여유 시 | 6 | ✅ 6/6 완료 |
| **합계** | | **29** | ✅ **29/29 완료 (100%)** |

**정정 이력** (2026-06-01):
- 초기 보고서에서 Mid 16건 / Low 4건 / 합계 23건으로 카운트했으나, 실제 항목 재집계 결과 **Mid 20건 / Low 6건 / 합계 29건**으로 정정. High 3건은 변동 없음.
- 정정 근거: 초기 분류 시 A 섹션과 B 섹션의 ID 발급 카운트와 본 §1 표 카운트가 동기화되지 않음.

**Tier 1 일괄 처리 완료** (2026-06-01): A-L3, A-M2, A-M3, A-N3, A-N4, A-O1, A-O2, B-L10, B-M5, B-N8, B-O3, B-O4 (12건). 동반 처리: B-L6 부분, B-L11 부분.

**Tier 2 일괄 처리 완료** (2026-06-01): A-L4, A-C1, A-C2, A-C3, B-L5, B-L6, B-L7, B-N10 (8건). 의존성 사전 검증 후 동기 수정. 핵심: evaluate-skill Step 7과 authoring-guide §4.4 동기, concretize-idea를 매핑 SSoT로 명시, Anthropic model-config docs fetch 후 §1.4 신설.

**Tier 3 위험 회피 적용 완료** (2026-06-01): A-M1(3-Tier 분류), A-N1/N2(optional bonus), A-N7(신규만 적용), B-L9(Phase 7 정체성 보강), B-M4(TOC 추가). 결함 해소 + 부작용 회피 양립.

**정의 합의 3건 완료** (2026-06-01): B-L11(24시간 timeline 기준), B-N11(외부 사용자 영향 기반), A-N6(Senior PM 페르소나). 사용자 직접 결정.

---

## 2. A. `plugin-skills-authoring-guide.md` 결함

### A-1. 논리적 결함

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **A-L1** | §1.2.1 | ~~`description` 최대 1,536자라고 명시. Anthropic 공식 spec은 통상 1024자로 알려져 있음~~ → **공식 docs 확인 결과 1,536자가 맞음**. 단, 가이드는 (1) `description` + `when_to_use` **합산** cap임을 누락, (2) `maxSkillDescriptionChars` 설정 가능 누락, (3) "skill listing truncation"이지 작성 한계 아님 — 보강 완료 (2026-06-01) | 🔴 High | ✅ 완료 |
| **A-L2** | §1.2.2~§1.2.4 vs §0.1 | §0에서 "PROCEDURE.md는 frontmatter 없음" 명시. 그러나 §1.2.2 "buddy 적용 점검 대상"이 `auto-create-pr`, `ship-release` 등 **PROCEDURE.md 스킬**로 적힘 → frontmatter 없는 파일에 `disable-model-invocation` 적용 불가, 실제로는 `plugin/commands/*.md`에 적용해야 함. 적용 대상 혼동 + §1.2.3 `user-invocable`은 PROCEDURE.md에 적용 자체가 **구조적으로 불가능** — buddy 등가 매핑으로 정정 완료 (2026-06-01) | 🔴 High | ✅ 완료 |
| **A-L3** | §4 B3, B4 | 평가 체크리스트 항목에 "이미 적용됨" 문구 포함 → LLM 평가 시 "패스로 가정"하여 실제 검증 skip 위험. 문구 제거 + F5 `user-invocable` 항목을 buddy 등가 매핑(command 부재 확인)으로 정정 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-L4** | §4.4 | 가중치 합 계산식의 "전체 항목 가중치 합" 분모 정의 모호. Persona 비권장 스킬에서 P1-P4 항목을 분모에 포함하는지 미정의 → 점수 비교 불가. **4가지 케이스(C1-C4)** 분모 명시 (C1=35, C2=29, C3=30, C4=24) + 예시 계산 + evaluate-skill Step 7 동기 수정 (2026-06-01) | 🟡 Mid | ✅ 완료 |

### A-2. lost-in-the-middle / 모호성

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **A-M1** | §2.1 | 권장 섹션 12개 → 본문 비대화. **3-Tier(필수 4 + 권장 4 + 선택 4)** 재분류 + 본문 길이별 적용 가이드 + 트랜지션 정책(B5-B8은 명확한 사유 시 pass, B9-B12는 보너스로만)으로 기존 PROCEDURE.md 점수 하락 방지 (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |
| **A-M2** | §1.2.5 | 9개 키 단순 나열, buddy 적용 가이드 없음 → LLM이 임의 적용 가능. 표에 "buddy 적용" 컬럼 추가하여 키별 정책 명시 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-M3** | §2.2.4 | "5개 이상 분리 / 20줄 초과 분리" 경계, 중간 영역(예: 3예시×15줄) 모호. 단일 임계(총 라인 합 ≤30/30-60/>60)로 통일 + 경계 판단 추가 기준 (2026-06-01) | 🟢 Low | ✅ 완료 |
| **A-O1** | §1.1 | "Claude는 본문을 보지 않고 frontmatter만 본다" — 부정확. progressive disclosure 3단계로 보강 (A-L2 수정 시 동반 완료, 재점검에서 확인 2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-O2** | §3.1 | "Anthropic의 공식 입장"으로 인용된 문장의 정확한 출처 URL 없음. paraphrase 명시 + 출처 2개 URL 추가 (2026-06-01) | 🟢 Low | ✅ 완료 |

### A-3. Claude 모델 호환성 누락

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **A-C1** | §6.1 G1 | Haiku 4.5 누락 + 모델별 행동 차이 가이드 없음. §1.4.3 모델별 행동 차이 표(Opus/Sonnet/Haiku) + §6.1 G1 갱신 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-C2** | §1.2.5 | `effort` 권장 값 가이드 없음. §1.4.2 effort level 모델별 지원 표(Opus 4.7/4.8: 5 levels, Opus 4.6/Sonnet 4.6: 4 levels, Haiku 미지원) + 사용 시점 가이드 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-C3** | §1.2.5 | `model` 키 — model ID 형식 가이드 없음. §1.4.1 alias/full name/`[1m]` suffix 형식 + buddy 사용 권장 정책 (2026-06-01) | 🟡 Mid | ✅ 완료 |

### A-4. 핵심 누락 (Anthropic prompt engineering 원칙)

| ID | 누락 항목 | 비고 | Severity | 상태 |
|----|---------|------|---------|------|
| **A-N1** | XML 태그 사용 가이드 | §7에 URL link만. **§2.4 신설** — "optional, bonus" 명시 + buddy 정책(마크다운 부족 시에만 사용) + 4가지 XML 태그 사용 표 (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |
| **A-N2** | Chain-of-Thought 패턴 | §7 link만. **§2.5 신설** — "optional, bonus" + 좁은 사용 케이스 명시 (여러 가능성 비교·self-verification만) + 남발 경고 (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |
| **A-N3** | 시스템 vs 사용자 프롬프트 구분 | PROCEDURE.md가 어느 쪽 역할인지 명시 없음. §2.0 신설 — PROCEDURE.md의 프롬프트 역할 + 사용자 메시지와의 차이 표 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-N4** | long-context 처리 전략 | 본문 길어지면 보조 파일 외 다른 전략 가이드 없음. §2.2.1에 5가지 전략 표(보조 파일 lazy-load / 본문 압축 / 위·아래 배치 / 단원 분리 / forking subagent) + lost-in-the-middle 명시 (2026-06-01) | 🟢 Low | ✅ 완료 |
| **A-N6** | §3.5.1 페르소나 권장 표 — Phase 9 누락 | **결정**: Senior Product Manager (sunset 전문) 페르소나 권장. 사용자 영향·소통·timeline 중심. deprecate-feature/archive-product/migrate-customers/spin-off-feature 일관 적용 (사용자 합의 2026-06-01) | 🟡 Mid | ✅ 완료 |
| **A-N7** | 한국어/영어 혼용 정책 | **§2.6 신설** — 권장 언어 매트릭스 + "신규 작성 skill에만 적용 (기존 동결)" 명시 + 신규 작성 체크리스트 + evaluate-skill 평가 항목 추가 X (회귀 방지) (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |

---

## 3. B. `engineering-phases.md` 결함

### B-1. 논리적 결함

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **B-L5** | Phase 1 Mode A | "Customer segment map: Phase 2 `identify-actors` 미사용 (Phase 1로 이관)" — 결정 근거 링크 없음. ADR 부재 확인 후 concretize-idea PROCEDURE.md "중요 변경 (2026-05-29)" 노트 참조 + 이관 근거(Stage 4-6이 customer 입력 요구) inline 명시 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-L6** | Phase 1 Mode B | Scope 분류 4종 vs Output 표 3종 불일치. §3 Routing 표 (B-L10 동반) + **Phase 1 Mode B Output 표를 scope(3종)와 routing(4종 옵션)으로 분리** + assess-product-change Step 4 기준 명시 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-L7** | Phase 1 Mode A | 12개 스킬 나열, orchestrator `concretize-idea`의 stage와 매핑 표 없음. Mode A 9개 스킬을 Stage 1-9에 직접 매핑한 표 + supporting skill 4개(market-size/JTBD/interview/target-market) 별도 분류 + concretize-idea PROCEDURE.md를 매핑 SSoT로 명시 (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-L8** | Phase 5 vs Phase 6 | Phase 5 종료 조건 "테스트 통과" vs Phase 6 정체성 "검증" — 책임 경계 모호. **Phase 5 = developer-authored test green / Phase 6 = 상용 quality bar(coverage·security·a11y·compliance·code health) 별도 평가**로 분리 + 두 phase에 책임 경계 박스 추가 (2026-06-01) | 🔴 High | ✅ 완료 |
| **B-L9** | Phase 7 | `run-uat`가 Phase 7 소속인데 "UAT 실패 → §5/§6 backtrack". UAT는 Phase 6 정체성에도 부합. **이관 대신 Phase 7 정체성 보강** + Phase 6/7 책임 경계 박스(코드 자체 품질 vs 실사용자 acceptance) 추가 (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |
| **B-L10** | §3 Skip | "Mode B Small: §1 → §5"를 "skip"이라 표현. 실제로는 §1 내부에서 라우팅이므로 **skip 아닌 routing** — 용어 부정확. §3을 "Skip vs Routing" 두 카테고리로 명확 분리, Mode B routing 표에 Defer/Reject(scope 4종) 포함 + hotfix 차이 한 줄 명시 (B-L6, B-L11 일부 동반 완료) (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-L11** | §3 Skip | Hotfix 발동 조건 모호. **결정**: time-pressure 기반(24시간 내 fix 필요). severity 무관. Mode B small은 normal cycle, hotfix는 §1 skip + 사후 24시간 내 incident report 보상 (사용자 합의 2026-06-01) | 🟡 Mid | ✅ 완료 |

### B-2. lost-in-the-middle 위험

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **B-M4** | 전체 | 711줄 부분 참조 시 LLM context 누락 가능. **분할 대신 TOC 추가** — 문서 상단에 §1~§5 + Phase 1~9 상세 anchor + 부분 참조 가이드. 참조 링크 깨짐 위험 회피 (2026-06-01) | 🟡 Mid | ✅ 완료 (위험 회피 적용) |
| **B-M5** | Phase 3 | 소속 스킬 35개 (최다) — 특정 스킬 찾기 어려움, 다른 phase와 분량 불균형. 9개 카테고리(3-A Architecture Foundation / 3-B Data & Contract / 3-C Security & Tenancy / 3-D Operations Strategy / 3-E Quality Baseline / 3-F UX/UI / 3-G Domain-Specific / 3-H Reviews / 3-I Decision Support)로 그룹화 (2026-06-01) | 🟡 Mid | ✅ 완료 |

### B-3. 누락 / 표현 결함

| ID | 위치 | 문제 설명 | Severity | 상태 |
|----|------|----------|---------|------|
| **B-N8** | Phase 7 Output | semver 결정 정책 한 줄 누락. "breaking → major, feature → minor, fix → patch (semver 2.0.0, `automate-release-tagging`이 변경 셋 분석 후 자동 제안)" 명시 (2026-06-01) | 🟢 Low | ✅ 완료 |
| **B-N10** | §3 routing 표 | Mode B medium-large 경계 케이스 처리 정책 미정의. §3 Routing 표에 assess-product-change Step 4 분류 기준 컬럼 추가 + "Medium-Large 경계 케이스" 박스(보수적 판단 기준 3가지 + 단순 파일 수만으로 분류 X 명시) (2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-N11** | Phase 9 | Phase 9(deprecation) vs Mode B(변경) 중첩 영역. **결정**: 외부 사용자 영향 기반 — 외부 사용자(API consumer/end user) 영향 있으면 Phase 9, 내부만이면 Mode B. Phase 9 정의에 분류 정책 표 + 판단 예시 4건 추가 (사용자 합의 2026-06-01) | 🟡 Mid | ✅ 완료 |
| **B-O3** | Phase 5, 7 | "만든다"/"내보낸다" 1단어 핵심 질문 — phase 정체성 표현 부족. Phase 5는 B-L8 수정 시 보강, Phase 7은 "사용자에게 안전하게 전달할 준비가 되었고, 전달했는가?"로 정정 + 정체성에 패키징/태깅/배포/공지/launch readiness 명시 (2026-06-01) | 🟢 Low | ✅ 완료 |
| **B-O4** | Phase 3 | "Phase 3은 가장 많은 stage skill을 보유한 phase" — 메타 정보가 정체성 정의에 섞임. "호출 패턴" 설명으로 변경 (orchestrator 경유의 가치를 명시 - tech stack과 data model 결정의 일관성 보장 예시) (2026-06-01) | 🟢 Low | ✅ 완료 |

---

## 4. 🔴 High 3건 처리 계획

### A-L1 — `description` cap 수치 검증

1. Anthropic 공식 docs (https://code.claude.com/docs/en/skills.md, https://docs.anthropic.com/en/docs/build-with-claude/skills/) 에서 정확한 character cap 확인
2. 확인된 값으로 §1.2.1 수정
3. 본 문서의 A-L1 상태를 "완료"로 갱신

### A-L2 — Frontmatter 적용 대상 명확화

1. §1.2.2 `disable-model-invocation` "buddy 적용 점검 대상" 항목을 **command 이름으로 명확화** — `auto-create-pr.md` → `plugin/commands/auto-create-pr.md`
2. §1.2.3 `user-invocable` 항목도 동일 처리
3. §0.2 적용 범위 표에 "각 키별 적용 대상" 컬럼 추가하여 모호성 제거
4. §1.2.0 신설 또는 §1.1에 "이 섹션의 모든 키는 `router/SKILL.md` 또는 `plugin/commands/*.md`에만 적용. PROCEDURE.md는 frontmatter 없음." 명시

### B-L8 — Phase 5/6 책임 경계

원칙: **Phase 5는 "코드+테스트가 생성된 상태", Phase 6은 "quality gate 통과 상태"** 로 분리.

1. Phase 5 종료 조건: "task 완료 + 자체 테스트(unit/integration) 작성·통과 + commit 완료" — "quality gate 통과"는 제외
2. Phase 6 정체성: "Phase 5 산출물을 **상용 품질 기준(coverage, security, compliance, code health)** 으로 검증" 명확화
3. Phase 5 Output Artifacts에서 "Tests"는 유지 (작성/실행 책임은 Phase 5)
4. Phase 6 Input Artifacts에서 "Working code + tests"는 유지하되, "Phase 5 산출물의 quality 평가 대상" 으로 의미 명확화

---

## 5. 🟡 Mid / 🟢 Low 처리 방침 — 우선순위 + 위험 분석

각 수정안에 대해 **(a) 결함 해소 효과**와 **(b) 수정이 유발할 수 있는 부정적 영향**을 함께 평가. 위험 회피형으로 정렬한다.

### 5.1 Tier 1 — 안전 우선 (단순 문구·메타 정보, 즉시 처리)

수정 시 다른 파일/스킬에 미치는 영향 없음. 정보 추가/정리만으로 끝남.

| # | ID | 수정 요지 | 위험 평가 |
|---|----|---------|---------|
| 1 | A-L3 | §4 체크리스트의 "(이미 적용됨)" 문구 제거 | ✅ 없음 |
| 2 | A-O1 | §1.1의 progressive disclosure 표현 재점검 (이미 일부 보강됨) | ✅ 없음 |
| 3 | A-N3 | PROCEDURE.md가 시스템 프롬프트 역할임을 §2 도입부에 메타 정보로 추가 | ✅ 없음 |
| 4 | A-M2 | §1.2.5 9개 키별 buddy 적용 가이드 한 줄씩 추가 | ✅ 없음 (정보 추가만) |
| 5 | B-L10 | §3 Skip의 "skip" → "routing" 용어 정정 | ✅ 없음 |
| 6 | B-M5 | Phase 3 35개 스킬 표를 카테고리(design/review/audit/...)로 그룹화 | ✅ 없음 (구조만 변경) |
| 7~12 | A-M3, A-O2, A-N4, B-N8, B-O3, B-O4 | Low 6건 — 출처 추가, 표현 보강, 한 줄 정책 명시 | ✅ 없음 |

**소요 추정**: 2-3시간 일괄 처리 가능.

### 5.2 Tier 2 — 의존성 확인 후 처리 (단일 스킬·문서 의존)

수정 전 관련 스킬/문서의 현재 상태 확인 필요. 일치하지 않으면 양쪽 동기 수정.

| # | ID | 수정 요지 | 의존 대상 | 위험 |
|---|----|---------|---------|------|
| 13 | A-L4 | §4.4 가중치 분모 계산 명시 (Persona 비권장 스킬 처리) | `evaluate-skill/PROCEDURE.md`의 점수 계산 로직 | ⚠️ 두 문서 동기 수정 필요 |
| 14 | B-L5 | Mode A "Customer segment map Phase 2 미사용" 결정 근거 ADR 링크 | 해당 ADR 실제 존재 확인. 없으면 새 ADR 작성 또는 결정 근거 한 줄 inline | ⚠️ ADR 미존재 시 추가 작업 |
| 15 | B-L6 | scope 분류 3종(small/medium/large) vs 4종(+defer-reject) 불일치 정렬 | `assess-product-change/PROCEDURE.md`의 실제 출력 | ⚠️ 스킬과 일치시켜야 함 |
| 16 | B-L7 | Phase 1 Mode A 12 스킬과 `concretize-idea` 9 stage 매핑 표 추가 | `concretize-idea/PROCEDURE.md`의 stage 정의 | ⚠️ 매핑 SSoT 결정 (어느 문서가 원본인가) |
| 17 | B-N10 | Mode B medium-large 경계 케이스 정책 추가 | `assess-product-change/PROCEDURE.md` | ⚠️ 스킬과 일치시켜야 함 |
| 18 | A-C1, A-C2, A-C3 | Haiku 4.5 추가, effort/model 키 권장 값 가이드 | Anthropic 공식 docs 재확인 | ⚠️ 잘못된 정보 추가 시 역효과 |

**소요 추정**: 4-6시간 (의존 파일 검증 포함).

### 5.3 Tier 3 — 구조적 영향 큼 (별도 사이클 권장, 신중 처리)

수정 시 기존 156개 PROCEDURE.md / 30개 command / 평가 시스템 전반에 영향. **결함 해소 < 수정 부작용** 가능성 존재 → 의사결정 필요.

#### 5.3.1 🔴 가장 위험한 수정 5건 — 부정적 영향 평가

| ID | 수정 시도 | 잠재 부작용 | 권장 대응 |
|----|---------|---------|---------|
| **B-M4** | engineering-phases.md (711줄)를 phases-1-4.md / phases-5-9.md로 분할 | `authoring-guide`, `evaluate-skill`, `write-a-skill`, `skill-catalog`, 156 PROCEDURE.md 일부에서 본 문서 참조 — **모든 참조 링크 깨짐** | 분할 대신 **목차(TOC) 추가 + 섹션별 anchor 정리**로 LLM 부분 참조 효율 개선. 분할은 권장 X |
| **A-N7** | 한국어/영어 혼용 정책 명시 | 156 PROCEDURE.md 전부 영향. 정책 결정 후 일제 정렬 필요 → 대규모 일괄 수정 | "현 상태 동결 + 신규 작성 시만 적용" 전략으로 압력 완화 |
| **A-M1** | §2.1 권장 섹션 12개 → 6-8개 압축 | 기존 PROCEDURE.md 다수가 12 섹션 기준으로 작성됨 → 갑자기 "초과 섹션 보유"로 판정 → **evaluate-skill 점수 하락**·재작성 압력 | "필수 4 + 권장 4 + 선택 4"로 재분류. 기존 스킬 점수 보정 트랜지션 계획 동반 |
| **A-N1** | XML 태그 사용 가이드 추가 (`<context>`, `<example>`, `<output>`) | 기존 PROCEDURE.md는 마크다운 기반, XML 태그 거의 사용 안 함 → 추가 시 **기존 스킬 일제히 "권장 미달"** → evaluate-skill 점수 충격 | "권장(optional, B+ 항목)"으로 추가. 평가 가중치 0 또는 보너스 점수로만 |
| **A-N2** | Chain-of-Thought 패턴 (`<thinking>` 등) 추가 | 동일. `<thinking>`은 Claude의 응답 형식인데 PROCEDURE.md에 명시 강제 시 부자연스러움 | "절차에서 명시적 추론 단계가 필요한 경우만" 좁은 범위로 한정 |
| **B-L9** | `run-uat`을 Phase 7 → Phase 6으로 이관 (또는 Phase 7 정체성에 UAT 포함 명시) | `skill-catalog`, `routing-rules`, `run-uat/PROCEDURE.md`, `verify-quality`, `ship-release` 모두 동기 수정 필요 → catalog 일관성 깨질 위험 | 이관 대신 **Phase 7 정체성에 "UAT는 launch readiness 일환"으로 명시**해서 위치 정당화 |

#### 5.3.2 정의 합의 필요 3건 (단순 수정 불가)

| ID | 결함 | 왜 단순 수정 불가 |
|----|------|---------|
| **B-L11** | Hotfix 발동 조건 모호 | "긴급" 정의 자체가 주관적. Mode B small과의 경계가 운영 정책에 가까움 — 단순 문서 수정으로 결정 불가, 사용자 결정 필요 |
| **B-N11** | Phase 9 (deprecation) vs Mode B (변경) 중첩 영역 | deprecate-feature는 Phase 9인데, 작은 deprecation은 Mode B small scope로 처리 가능 — 분류 정책 결정 필요 |
| **A-N6** | Phase 9 페르소나 권장 표 누락 | Phase 9 스킬(`deprecate-feature`, `archive-product`, `migrate-customers`)에 페르소나 미적용 상태 — 권장 추가 시 적용 압력. 어느 페르소나가 적합한지 결정 필요 |

### 5.4 권장 처리 순서

```
1. Tier 1 (안전 12건) 일괄 처리 — 즉시
   └→ evaluate-skill/write-a-skill 검증 (Task #3, #4)

2. Tier 2 (의존성 6건) — 의존 파일 검증 후 동기 수정
   └→ 각 수정 전 의존 파일 grep으로 일치 확인

3. Tier 3 (위험 5건) — 별도 의사결정 사이클
   ├→ B-M4: 분할 권장 안 함, TOC 보강으로 대체
   ├→ A-N7: 동결 + 신규만 적용
   ├→ A-M1, A-N1, A-N2: 평가 가중치 보정 동반
   └→ B-L9: 정체성 문구 보강으로 대체

4. 정의 합의 필요 3건 (B-L11, B-N11, A-N6) — 사용자 결정 후 진행
```

### 5.5 "수정이 오히려 이슈를 발생시키는" 항목 요약

가장 주의가 필요한 4건:

| ID | 핵심 위험 | 회피 전략 |
|----|---------|---------|
| **B-M4** (711줄 분할) | 광범위한 참조 링크 깨짐 | 분할 대신 TOC 추가 |
| **A-N7** (한·영 혼용 정책) | 156 PROCEDURE.md 일제 영향 | 신규만 적용 |
| **A-M1** (12 섹션 압축) | 기존 스킬 점수 충격 | 평가 가중치 보정 동반 |
| **A-N1/N2** (XML/CoT 가이드) | 기존 스킬 "권장 미달" 판정 | "optional" 또는 "bonus"로 추가 |

이들은 **결함 해소보다 부작용이 클 수 있으므로** 별도 의사결정 후 처리 권장.

---

## 6. 후속 작업 (이 감사가 완료된 후)

- [ ] **Task 3**: `evaluate-skill` PROCEDURE.md가 본 두 가이드 문서를 평가 기준에 충분히 반영하는지 검증
- [ ] **Task 4**: `write-a-skill` PROCEDURE.md가 본 두 가이드 문서를 작성 절차에 충분히 반영하는지 검증
- [ ] 두 메타-스킬에 누락된 가이드 항목을 반영
- [ ] (선택) 본 결함 트래킹 문서를 ADR로 승격 검토

---

## 7. 참조

- `docs/plugin-skills-authoring-guide.md` — 감사 대상 1
- `plugin/skills/router/references/engineering-phases.md` — 감사 대상 2
- `plugin/skills/evaluate-skill/PROCEDURE.md` — 본 가이드를 평가 기준으로 사용하는 메타-스킬
- `plugin/skills/write-a-skill/PROCEDURE.md` — 본 가이드를 작성 기준으로 사용하는 메타-스킬

---

**문서 끝**.
