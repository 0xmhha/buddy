# Skill Completion Plan — plugin buddy 42 신규 skill + 3 통합

> **목적**: [`docs/notes/2026-05-10-missing-skills-inventory.md`](../../notes/2026-05-10-missing-skills-inventory.md) 의 42 신규 skill + 3 기존 PROCEDURE 갱신 + 3 deferred-extension (Korea cluster) 의 *작성 진입 plan*. Step 5 (실제 skill 작성) 의 직접 입력.
>
> **Step 1~3 산출 입력**:
> - [`2026-05-10-missing-skills-inventory.md`](../../notes/2026-05-10-missing-skills-inventory.md) — 42 신규 + 3 deferred + 3 통합 + 1 deferred MCP 매트릭스
> - [`2026-05-10-external-skills-inventory.md`](../../notes/2026-05-10-external-skills-inventory.md) — 92 외부 프로젝트 인벤토리
> - [`2026-05-10-skill-matrix-3a-groups-2-4.md`](../../notes/2026-05-10-skill-matrix-3a-groups-2-4.md) — 그룹 2 + 그룹 4 매트릭스
> - [`2026-05-10-skill-matrix-3b-group-1.md`](../../notes/2026-05-10-skill-matrix-3b-group-1.md) — 그룹 1 매트릭스

---

## 1. 작성 원칙

### 1.1 PROCEDURE.md 양식 (기존 구현된 skill 들의 12 section)

| § | section | 책임 |
|---|--------|------|
| §0 | 헤더 (목적 / 진입 조건 / 산출물 / 다음 phase) | 1줄 정체성 |
| §1~§3 | Stage 흐름 / 분리 근거 / Stage 모델 | 작업 절차 |
| §4 | 실행 절차 — Stage 별 detail | 본체 |
| §5 | 산출물 형식 | template |
| §6 | User Gate (필요 시) | 사용자 confirm 시점 |
| §7 | 다음 phase | cascade out |
| §8 | 참조 | 외부 docs |
| §9 | 권장 호출 패턴 | chain 예시 |
| §10 | Open Questions (있으면) | 미결정 항목 |
| §11 | 검증 self-check (verification gate) | acceptance 자동 검증 |

### 1.2 명명 컨벤션

- `kebab-case` + `동사+명사`
- region 명시 시 `{verb}-{region}-{topic}` (예: `consult-korea-legal-context`)
- 약어 (CRO / SEO / ASO) 본문 첫 사용 시 풀이

### 1.3 dispatch 정합

- 신규 skill 추가 시 `plugin/skills/<name>/PROCEDURE.md` + `plugin/commands/<name>.md` (frontmatter 의 `disable-model-invocation: true` 필수, ADR-001 컨벤션)
- `plugin/skills/router/references/skill-catalog.md` 등재
- `plugin/.claude-plugin/plugin.json` 의 `commands` 배열 추가

### 1.4 호환성 점검 (외부 차용 시)

- license / attribution 확인 (외부 자산 license 와 buddy Apache 2.0 호환성)
- 변환 비용 (LOW: 양식만 / MEDIUM: 본문 일부 재구성 / HIGH: 전면 재작성)
- 명명 컨벤션 정정

---

## 2. Batch 분할 — 7 batch (cascade 의존 순서)

### 2.1 의존 cascade

```
Batch 1 (foundation)
  └── decide-target-market (Layer 1 region 결정 — 하위 batch 의 trigger)
  └── §1 customer/market 4 (analyze-market-size / map-customer-segments / map-jobs-to-be-done / conduct-customer-interview)
        │
        ▼
Batch 2 (§1 wrap-up + §2 + 그룹 2 신규)
  └── analyze-competition-and-substitutes (§1)
  └── estimate-feature-effort (§2)
  └── review-legal-regulatory (그룹 2 신규)
        │
        ▼
Batch 3 (§3 design 부가 + 그룹 4 stage 3 / 4)
  └── §3 부가 4 (observability / secret / i18n / a11y)
  └── 그룹 4 stage 3 (decide-form-factor-app-vs-web)
  └── 그룹 4 stage 4 4 (apply / audit / prototype / interaction)
        │
        ▼
Batch 4 (§5 build 부가)
  └── 5 build skill (§5 D)
        │
        ▼
Batch 5 (§6 verify 부가)
  └── 3 audit / chaos skill (§6 A)
        │
        ▼
Batch 6 (§8 data + 그룹 4 stage 10+11 통합)
  └── §8 7 + 그룹 4 6 = 13 skill
        │
        ▼
Batch 7 (§9 lifecycle + 그룹 2 통합 3)
  └── §9 4 + define-product-spec PROCEDURE 갱신 + review-engineering PROCEDURE 갱신
```

### 2.2 batch 크기

| batch | skill 수 | 평균 비용 | 예상 응답 회수 |
|-------|--------|--------|----------|
| 1 | 5 | LOW-MEDIUM | 1 |
| 2 | 3 | LOW | 1 |
| 3 | 9 | MEDIUM | 2 |
| 4 | 5 | LOW | 1 |
| 5 | 3 | MEDIUM | 1 |
| 6 | 13 | MEDIUM | 2 |
| 7 | 6 (4 + 2 통합) | MEDIUM | 1 |
| **합계** | **44** (42 신규 + 2 통합) | | **9 응답** |

→ 9 응답에 모든 skill 작성. 단 사용자 페이스 / 분량 조절로 변동 가능.

> deferred 3 (Korea cluster) + 1 통합 (review-code-architecture → review-engineering) 은 본 plan 외. deferred 는 trigger 발생 시, 통합 1 은 batch 7 의 review-engineering 갱신에 흡수.

---

## 3. Batch 별 detail

### 3.1 Batch 1 — foundation (5 skill)

**Cascade 책임**: 이후 모든 batch 의 *region 결정 + market validation* 입력 제공.

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `decide-target-market` | 그룹 4 stage 2 | (d) 신규 | — (Layer 1 신규) | MEDIUM |
| `analyze-market-size` | 그룹 1 §1 | (d) 신규 | tavily-mcp + arxiv-mcp-server (도구) | MEDIUM |
| `map-customer-segments` | 그룹 1 §1 | (d) 신규 | marketingskills/customer-research (참고) | MEDIUM |
| `map-jobs-to-be-done` | 그룹 1 §1 | (d) 신규 | — (JTBD 프레임워크 일반) | LOW |
| `conduct-customer-interview` | 그룹 1 §1 | (b) 수정 차용 | marketingskills/customer-research + designer-skills/design-research + gpt-researcher | MEDIUM |

**Acceptance**:
- 5 skill PROCEDURE.md + commands/*.md + skill-catalog 등재 + plugin.json 갱신
- §11 verification gate 모두 작성
- `decide-target-market` 의 §7 (다음 phase) 가 §1 customer/market 의 4 skill cascade out

### 3.2 Batch 2 — §1 wrap-up + §2 + 그룹 2 신규 (3 skill)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `analyze-competition-and-substitutes` | 그룹 1 §1 | (b) 수정 차용 | marketingskills/competitor-alternatives + competitor-profiling | LOW |
| `estimate-feature-effort` | 그룹 1 §2 | (d) 신규 | — (T-shirt sizing 표준) | LOW |
| `review-legal-regulatory` | 그룹 2 신규 | (d) 신규 | — (region-agnostic frame, Layer 2) | MEDIUM |

**Acceptance**:
- 3 skill 작성 + 등재
- `review-legal-regulatory` 의 §7 가 *region 결정 결과 글로벌* 일 때 stop, *Korea* 일 때 Korea cluster 3 invoke

### 3.3 Batch 3 — §3 design 부가 + 그룹 4 stage 3 / 4 (9 skill)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `design-observability` | 그룹 1 §3 B | (d) 신규 | — (SLO/SLI 일반) | MEDIUM-HIGH |
| `design-secret-management` | 그룹 1 §3 B | (d) 신규 | — (secret rotation 일반) | MEDIUM |
| `design-i18n-strategy` | 그룹 1 §3 B | (d) 신규 | — (ICU MessageFormat / RTL) | MEDIUM |
| `design-accessibility-baseline` | 그룹 1 §3 B | (b) 수정 차용 | designer-skills/interaction-design (a11y 부분) | MEDIUM |
| `decide-form-factor-app-vs-web` | 그룹 4 stage 3 | (d) 신규 | — | MEDIUM |
| `apply-design-system` | 그룹 4 stage 4 | (b) 수정 차용 | designer-skills/design-systems plugin | MEDIUM |
| `audit-ui-quality` | 그룹 4 stage 4 | (b) 수정 차용 | make-interfaces-feel-better + designer-skills/designer-toolkit | MEDIUM |
| `prototype-from-spec` | 그룹 4 stage 4 | (b) 수정 차용 | designer-skills/prototyping-testing plugin | MEDIUM |
| `design-interaction-pattern` | 그룹 4 stage 4 | (b) 수정 차용 | designer-skills/interaction-design plugin | MEDIUM |

**Acceptance**:
- 9 skill 작성 + 등재
- `decide-form-factor-app-vs-web` 가 §3 의 *stage 0* 으로 진입 (define-tech-stack 선행)
- `apply-design-system` 가 *기존 design system 채택* 모드, `prototype-from-spec` 이 *신규 design 작성* 모드 책임 분리

### 3.4 Batch 4 — §5 build 부가 (5 skill)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `generate-from-api-contract` | 그룹 1 §5 D | (d) 신규 | autonomous-coding-agents (참고) | LOW |
| `generate-tests-from-spec` | 그룹 1 §5 D | (d) 신규 | autonomous-coding-agents/QA agent (참고) | LOW |
| `pair-program-loop` | 그룹 1 §5 D | (d) 신규 | — (TDD red-green-refactor) | LOW |
| `refactor-with-rename-trace` | 그룹 1 §5 D | (d) 신규 | memory-bank (참고) | LOW |
| `update-docs-with-code` | 그룹 1 §5 D | (d) 신규 | — | LOW |

**Acceptance**:
- 5 skill 작성 + 등재
- 모두 *IDE 대체 가능* 영역 — 본 batch 는 *문서화* 우선, 깊이는 trigger 발생 시 보강

### 3.5 Batch 5 — §6 verify 부가 (3 skill)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `audit-i18n-coverage` | 그룹 1 §6 A | (d) 신규 | — | MEDIUM |
| `chaos-test` | 그룹 1 §6 A | (d) 신규 | — (chaos engineering 일반) | MEDIUM |
| `audit-test-coverage-meaningful` | 그룹 1 §6 A | (b) 수정 차용 | agent-evaluation | MEDIUM |

**Acceptance**:
- 3 skill 작성 + 등재
- `audit-i18n-coverage` 가 `design-i18n-strategy` (Batch 3) 의 산출 입력

### 3.6 Batch 6 — §8 data + 그룹 4 stage 10+11 (13 skill)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `analyze-feature-adoption` | 그룹 1 §8 F | (b) 수정 차용 | marketingskills/analytics-tracking | LOW |
| `analyze-user-cohort` | 그룹 1 §8 F | (b) 수정 차용 | marketingskills/churn-prevention + analytics-tracking | LOW |
| `analyze-actor-failure-rate` | 그룹 1 §8 F | (b) 수정 차용 | agent-evaluation | MEDIUM |
| `analyze-cost-anomaly` | 그룹 1 §8 F | (d) 신규 | — | MEDIUM-HIGH |
| `triage-customer-support-ticket` | 그룹 1 §8 F | (d) 신규 | — | MEDIUM |
| `analyze-customer-feedback-corpus` | 그룹 1 §8 F | (b) 수정 차용 | marketingskills/customer-research + humanizer | MEDIUM |
| `audit-error-budget` | 그룹 1 §8 F | (d) 신규 | — (SLO burn rate) | MEDIUM-HIGH |
| `optimize-conversion-funnel` | 그룹 4 stage 10+11 | (b) 수정 차용 | marketingskills 5 CRO 통합 | MEDIUM |
| `plan-growth-experiment` | 그룹 4 stage 10+11 | (d) 신규 | — | MEDIUM |
| `draft-marketing-copy` | 그룹 4 stage 10+11 | (b) 수정 차용 | marketingskills copywriting + copy-editing + ad-creative | LOW |
| `plan-marketing-channel` | 그룹 4 stage 10+11 | (b) 수정 차용 | marketingskills 6 channel 통합 | MEDIUM |
| `audit-seo-aso` | 그룹 4 stage 10+11 | (b) 수정 차용 | marketingskills ai-seo + aso-audit | LOW |
| `automate-marketing-content` | 그룹 4 stage 10+11 | (b) 수정 차용 | marketingskills email-sequence + cold-email + content-strategy | MEDIUM |

**Acceptance**:
- 13 skill 작성 + 등재
- §8 의 7 + 그룹 4 의 6 — *production traffic 의존* 영역이라 PROCEDURE.md 작성만, 실 동작 검증은 deferred

### 3.7 Batch 7 — §9 lifecycle + 통합 (4 + 2 PROCEDURE 갱신)

| skill | cluster | 결정 | 외부 입력 | 비용 |
|-------|--------|------|---------|------|
| `deprecate-feature` | 그룹 1 §9 G | (d) 신규 | — | MEDIUM |
| `migrate-customers` | 그룹 1 §9 G | (d) 신규 | — | MEDIUM |
| `archive-product` | 그룹 1 §9 G | (d) 신규 | — | MEDIUM |
| `spin-off-feature` | 그룹 1 §9 G | (d) 신규 | — | MEDIUM |
| **통합** | **결정** | **외부 입력** | **비용** |
| `define-product-spec` PROCEDURE 갱신 | 그룹 2 통합 | 통합 (define-product-context + write-prd 흡수) | — | LOW |
| `review-engineering` PROCEDURE 갱신 | 그룹 2 통합 | 통합 (review-code-architecture 흡수) | — | LOW |

**Acceptance**:
- 4 신규 skill 작성 + 등재
- 2 PROCEDURE 갱신 — 흡수된 책임이 명확히 추가됨 (sub-section 형태)

---

## 4. 외부 자산 차용 절차 ((b) 17 건)

### 4.1 license / attribution 확인 단계

각 (b) 차용 시 다음 5 단계:

1. 외부 자산 license 확인 (LICENSE 파일 read)
2. license 와 buddy Apache 2.0 호환성 점검
   - **MIT / BSD / Apache 2.0** → 호환 (attribution 필요)
   - **GPL / AGPL** → 비호환 (license 충돌)
   - **proprietary** → 사용 금지
3. attribution 추가:
   - `NOTICE` 파일에 외부 자산 명시 (현재 mattpocock/skills + gstack 패턴 따름)
   - `README.md` 의 Acknowledgments 섹션 갱신
4. 본문 변환 (PROCEDURE.md 12 section 양식)
5. 명명 정정 (buddy 컨벤션 — 동사+명사 + region-explicit)

### 4.2 외부 자산 license 매트릭스 (사전 확인 필요)

| 외부 자산 | license (확인 필요) | 호환성 |
|---------|----------------|------|
| marketingskills | (Step 5 시점 확인) | 추정 MIT/Apache (agentskills.io 호환) |
| designer-skills | (Step 5 시점 확인) | 추정 MIT/Apache (Claude Code skill) |
| make-interfaces-feel-better | (Step 5 시점 확인) | 추정 호환 |
| agent-evaluation | (Step 5 시점 확인) | 확인 필요 |
| autonomous-coding-agents | (Step 5 시점 확인) | 확인 필요 |
| memory-bank | (Step 5 시점 확인) | 확인 필요 |
| gpt-researcher | (Step 5 시점 확인) | 확인 필요 |
| tavily-mcp / arxiv-mcp-server | (Step 5 시점 확인) | 추정 MIT (MCP server 표준) |

→ Step 5 의 *각 skill 작성 직전* 에 license 확인. 비호환 발견 시 (c) 참고만 으로 강등.

---

## 5. 신규 (d) 25 건의 도메인 source 가이드

(d) 25 = 그룹 1 의 21 + 그룹 2 신규 1 + 그룹 4 의 2 (form-factor + plan-growth-experiment) + Layer 1 의 1 (decide-target-market).

| 도메인 | source 권장 |
|------|----------|
| §1 customer/market (3) | Lean Startup / Customer Development (Steve Blank) / Strategyzer Value Proposition Canvas |
| §2 effort (1) | Agile Estimating and Planning / T-shirt sizing best practices |
| §3 observability (1) | Google SRE Book + OpenTelemetry spec |
| §3 secret (1) | OWASP Cryptographic Storage Cheat Sheet + 12-factor app |
| §3 i18n (1) | ICU MessageFormat + Unicode CLDR |
| §5 build (5) | TDD by Example (Kent Beck) + Modern Software Engineering (Dave Farley) |
| §6 i18n-coverage / chaos (2) | Chaos Engineering (Casey Rosenthal) |
| §8 data (3 — cost / ticket / error budget) | SRE Book Ch.4 SLO + Cost-aware engineering |
| §9 lifecycle (4) | API Deprecation guide (Stripe) + EOL best practices |
| 그룹 4 form-factor (1) | Mobile vs Web decision matrix (Apple HIG / Material Design) |
| 그룹 4 plan-growth-experiment (1) | Hacking Growth (Sean Ellis) + ICE/RICE prioritization |
| Layer 1 decide-target-market (1) | Market entry strategy (HBR) + GTM frameworks |

→ 본 source 들은 *PROCEDURE.md 작성 시 참조*. 실제 차용은 *원칙 / 프레임워크 추출* 만, 본문은 buddy 컨벤션으로 재작성.

---

## 6. Layer 3 deferred 3 (Korea cluster) — template 보존

| skill | trigger | 외부 입력 | 작성 시점 |
|------|--------|---------|--------|
| `consult-korea-legal-context` | target market = Korea | korean-legal-guide_skill | trigger 시 |
| `draft-korea-patent-application` | Korea + 특허 | patent-application-drafting_skill | trigger 시 |
| `audit-korea-cii-vulnerability` | Korea + CII 보호법 | KESE-KIT | trigger 시 |

→ 본 plan 에서는 *template* 만 보존. 실 작성은 사용자가 *target market = Korea* 결정 시 (= `decide-target-market` Batch 1 산출이 Korea 일 때).

---

## 7. 진행 cadence + Step 5 진입 조건

### 7.1 cadence

| 단계 | 단위 | 사용자 confirm |
|------|-----|------------|
| Step 5a | Batch 1 (5 skill) 작성 — 1 응답 | confirm 후 다음 batch |
| Step 5b ~ 5g | Batch 2~7 작성 — 8 응답 | 각 batch 후 confirm |

→ 9 응답 cycle. 사용자 페이스에 따라 *동시 batch 묶음* 가능.

### 7.2 Step 5 진입 조건

- 본 Step 4 plan commit 완료
- 사용자 confirm — Batch 1 진입
- (선택) HANDOFF.md 의 plugin buddy 트랙 상태 갱신 — *skill 보완 진행 중 (Batch N/7)* 표시

### 7.3 batch 별 commit 정책

- 각 batch 종료 = 1 commit (skill 묶음 신규 + plugin.json + skill-catalog 갱신)
- commit message: `feat(skills): batch N/7 - <cluster summary>`

---

## 8. 검증 / acceptance gate

### 8.1 batch 별 acceptance

| 항목 | gate |
|------|-----|
| PROCEDURE.md 작성 | 12 section 모두 채움 |
| commands/*.md 작성 | frontmatter `disable-model-invocation: true` |
| skill-catalog 등재 | trigger / phase / cluster 명시 |
| plugin.json `commands` 배열 갱신 | jq 검증 통과 |
| 외부 차용 시 attribution | NOTICE / README 갱신 |
| §11 self-check | 모든 항목 자가 검증 |

### 8.2 전체 acceptance (모든 batch 종료 후)

| 항목 | gate |
|------|-----|
| plugin/skills/ count | 106 → 148 (42 신규 + 0 통합) |
| plugin/commands/ count | 57 → 99 (42 신규) |
| 외부 자산 license 확인 | 17 (b) 모두 확인 완료 |
| Korea cluster template | 3 deferred 명시됨 |
| HANDOFF.md / tasks.md 갱신 | A-2 잔여 29 → 0, 그룹 2 / 4 / 4-extension 결정 lock-in |

---

## 9. Open Questions (Step 5 진입 시)

| 질문 | 결정 시점 |
|------|--------|
| Batch 동시 진행 vs 순차? | 사용자 페이스에 따라 |
| 외부 자산 license 비호환 발견 시 처리? | 발견 시점에 (c) 강등 |
| (d) 25 건의 도메인 source 가 사용자 의도와 충돌 시? | 발견 시점에 사용자 confirm |
| Korea cluster trigger 발화 시점? | 사용자가 *target market = Korea* 결정 시 |

---

## 10. 다음 액션

1. 본 plan commit
2. 사용자 confirm — Batch 1 진입
3. Batch 1 작성 (5 skill: decide-target-market + §1 customer 4)
4. Batch 1 commit
5. 다음 batch 진입 — 사용자 confirm 후
