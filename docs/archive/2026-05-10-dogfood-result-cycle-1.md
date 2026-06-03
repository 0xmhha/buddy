# Dogfood Validation Result — Cycle 1 (2026-05-10)

> **Predecessor**: `2026-05-10-dogfood-validation-scenarios.md`
> **Trigger**: `/buddy:concretize-idea "<test>"` invocation in this session
> **Status**: Stage 1 (validate-idea) entry verified, escape-hatch terminated for efficiency
> **Cycle scope**: §3 discovery, §4 single-skill 부분, §5 cascade 진입, §6 attribution 도달성

본 보고서는 dogfood validation scenarios `§7 양식` 에 따라 작성. 실제 다이얼로그가 아닌 *PROCEDURE 구조 + 라우팅 + 게이트 동작* 검증에 초점 — canned idea 의 mock 응답으로 forcing-question 일대일 그릴링 진행 시 anti-sycophancy 진단의 valid 가치가 zero 라는 판단(§3 의 *mock 한계* 항목 참조).

---

## §1. 검증 목적 매핑 (scenarios doc §1 → 본 cycle 결과)

| 영역 | 측정 | Cycle 1 결과 |
|------|------|------|
| Discovery | `/buddy:<name>` autocomplete | ✅ `/buddy:concretize-idea` 자동완성 / dispatch 성공 |
| Description | command frontmatter description baseline 제외 | ✅ system-reminder available-skills 에 `buddy:router` 1개만 등장 (44 신규 description 미등장) — ADR-001 정합 |
| Routing | router 가 자연어 → 신규 skill dispatch | ⚠️ 자연어 dispatch 미검증 (직접 슬래시 명령 invocation 만 검증) |
| Cascade | A → B 호출 흐름 정합 | 🟡 concretize-idea(entry) → validate-idea(Stage 1) 진입까지만 — Stage 2~8 미진입 |
| PROCEDURE 정합 | 12-section 양식 일관 | ✅ concretize-idea / validate-idea 본문 구조 정합 (sample 2/148) |
| 외부 자산 attribution | NOTICE / README / ADR 도달성 | ✅ NOTICE 8 entry / ADR Index 3 ADR / README Acknowledgments 표 도달 가능 |

---

## §2. Discovery

- **자동완성 등장 신규 commands**: 본 cycle 에선 `/buddy:concretize-idea` 만 직접 호출, 99 commands 대량 검증은 미수행.
- **누락**: 없음 (검증 범위 내).
- **system-reminder available-skills**: `buddy:router` 1개만 노출 — ADR-001 의 `disable-model-invocation: true` 가 의도대로 작동.

---

## §3. Routing

- **router → 신규 skill dispatch 정확도**: 1/1 (concretize-idea PROCEDURE.md 로드 성공, 첫 시도에서 placeholder path `${CLAUDE_PLUGIN_ROOT}/skills/<name>/PROCEDURE.md` 가 실 install path 로 resolve — Bash fallback 미발동).
- **자연어 dispatch 부정확 case**: 미검증 — Cycle 2 에서 `/buddy:run "<자연어>"` 또는 router 의 lazy-load 카탈로그 부분 검증 필요.
- **mock 한계**: forcing-question 일대일 그릴링은 *진짜 founder* 의 polish-된-답 → 실제-답 transition 진단을 위한 도구. canned idea 의 mock 응답에 대해 "demand 가 zero" 같은 진단을 자동 출력하는 건 anti-sycophancy 의미 없음 — Cycle 1 은 *PROCEDURE 본문이 올바르게 load 되고 진입점 question 이 정확히 출력* 까지만 검증.

---

## §4. PROCEDURE 정합

샘플 2개 (concretize-idea / validate-idea) 본문 측정:

| Skill | 본문 | 12-section 항목 | 비고 |
|-------|------|----------------|------|
| concretize-idea | 133 line | ✅ Stage 흐름 / 실행 절차 / 산출물 형식 / User Gate / 다음 phase / 참조 — 8 stage 모두 ID 명시 | bracket `[name]` notation L26 *현 시점 stale* (B3 — §6) |
| validate-idea | 517 line | ✅ Two-mode (Startup / Builder) / Anti-Sycophancy 규칙 / 특정성 강제 기법 / 6 forcing 질문 / 디자인 문서 템플릿 / 다음 단계 핸드오프 / 중요 규칙 | batch-금지 (L163, L515) 정합 — Q1 단독 출력 후 응답 대기 동작 |

- **self-check (§6 이라 표기) 동작**: validate-idea 는 *self-check* 섹션 부재. Q1~Q6 후 *전제 체크 + 2-3 대안 + escape hatch* 패턴은 self-check 의 substitute 로 동작 — 양식 명칭 통일 미흡 (B6 — §6).

---

## §5. Cascade

- **concretize-idea(entry) → validate-idea(Stage 1)**: ✅ 진입 정합. Stage 1 gate (idea 명확성) 가 placeholder `<test>` 입력에 대해 정확히 발동 → user 가 mode 결정 → validate-idea PROCEDURE 로드 → smart routing (pre-product → Q1+Q2+Q3) → Q1 단독 출력.
- **validate-idea → validate-advanced-edge-idea**: ❌ 미검증 (Cycle 1 에선 Q1 응답 부재로 escape hatch 발동, Stage 2 미진입).
- **Stage 3~8 (assess-business-viability / analyze-competition-and-substitutes / review-pricing-and-gtm / map-customer-segments / define-product-spec / autoplan)**: ❌ 미검증.

---

## §6. Attribution 도달성

| 영역 | 결과 |
|------|------|
| NOTICE entries | ✅ 8/8 (mattpocock, gstack, marketingskills, designer-skills, gpt-researcher, agent-evaluation, humanizer, superpowers) — ADR-003 정합 |
| README Acknowledgments | ✅ 표 형식, 10 항목 (W2-4 commit `ea9954a`) |
| ADR Index | ✅ `docs/superpowers/decisions/README.md` 3 ADR 등록 (ADR-001/002/003) — W2-5 commit `ea9954a` |

---

## §7. 발견 issue

| ID | severity | 위치 | 내용 | fix 후보 |
|----|----------|------|------|---------|
| B1 | minor | `docs/notes/2026-05-10-dogfood-validation-scenarios.md` §3-§4 | §3 = autocomplete discovery 검증, §4 = single-skill 호출 — *canned business scenario* 가 doc 어디에도 정의 안 됨. 본 응답에서 "구독형 한국어 학습 SaaS" 를 즉석 채택 | §3.X canned scenarios 표 추가, scenario 별 idea statement / target user / 1 문장 problem 선언 |
| B2 | minor | `plugin/skills/validate-idea/PROCEDURE.md` L171 | Q1 정확 phrasing "내일 사라지면 진짜로 화날" 어순/문법 어색 | "내일 사라지면 진짜로 화낼" 또는 "내일 없어지면 진짜로 화내는" |
| B3 | major | `plugin/skills/concretize-idea/PROCEDURE.md` L18, L20, L26 | Stage 4 / 6 의 `[analyze-competition-and-substitutes]` / `[map-customer-segments]` 가 bracket notation 으로 표기 — Skill Completion Cycle 종료 후 두 skill 모두 *작성 완료* (Batch 1 / 2). bracket 제거 + L26 의 "신규 작성 필요" 문구 outdated | bracket 제거 + L26 stale 문장 삭제, Stage 본문에 직접 invoke 명시 |
| B4 | minor | `plugin/skills/concretize-idea/PROCEDURE.md` L49 | "(`analyze-competition-and-substitutes` skill 미존재 시 orchestrator가 수행)" 가 stale — skill 존재함 | fallback 문구 삭제 |
| B5 | minor | `plugin/skills/concretize-idea/PROCEDURE.md` L59 | 동일 — `map-customer-segments` 도 존재 | fallback 문구 삭제 |
| B6 | minor | PROCEDURE 양식 통일 | concretize-idea 는 `## 산출물 형식` / validate-idea 는 `## Output: 디자인 문서` — section 명칭 inconsistency. 12-section 양식에 *self-check* 명시는 일부 skill 에만 (assess-business-viability 등) | PROCEDURE skeleton template 도입 후 일괄 통일 (Cycle 2 후보) |
| B7 | conceptual | router smart-skip | validate-idea Q4 의 smart-skip 규칙 (L161 *이전 답이 이미 나중 질문 커버했으면 skip*) 이 router 단에선 *대화 메모리 의존* — `/buddy:run validate-idea` 재호출 시 이전 답 유실 | session state 메커니즘 도입 (별도 design 필요) — Cycle 3+ 후보 |

---

## §8. Follow-up — 다음 cycle input

### 즉시 fix (Cycle 1 피드백 반영, 단일 PR)

- B3 / B4 / B5 일괄 (concretize-idea PROCEDURE 의 bracket / fallback 문구 stale 제거) — *최소 변경 30% 미만* 원칙.
- B1 (scenarios doc 의 §3.X canned scenarios 표 추가) — *문서 보강*.
- B2 (validate-idea L171 phrasing) — *prose-only*.

### Cycle 2 (validation 확장)

- §4 의 5 single-skill 테스트 (decide-target-market / review-legal-regulatory / audit-test-coverage-meaningful / optimize-conversion-funnel / archive-product) 실 dispatch 검증.
- §5 cascade 4 stage (concretize-idea → decide-target-market → review-legal-regulatory → design-system) 실 dispatch 검증.
- 자연어 dispatch (router 의 lazy-load skill catalog reference 트리거) 검증.

### Cycle 3+ (deferred)

- B6 PROCEDURE 양식 일괄 통일 — skeleton template + lint script.
- B7 session state 메커니즘 — 별도 design (cli buddy spec §6.2 feature-management-mcp 에 통합 후보).

---

## §9. quality gate 결과

scenarios doc §8 의 *Skill Completion Cycle quality gate* 기준:

| 기준 | 결과 |
|------|------|
| plugin 재install / 재인식 | ✅ `/plugin` 출력 = "buddy 1.1.0 latest", `/reload-plugins` 출력 = "22 plugins · 121 skills · 56 agents · 17 hooks" — buddy 가 reload 대상에 포함됨 |
| 신규 commands discovery | ✅ `/buddy:concretize-idea` 직접 호출 성공 (1/99 sample) |
| router → PROCEDURE 라우팅 | ✅ 1/1 sample |
| PROCEDURE 본문 정합 | ✅ 2/148 sample (concretize-idea / validate-idea) |
| 단일 skill 호출 (§4) | ⚠️ 1/5 — *partial* |
| cascade (§5) | ⚠️ stage 1 entry only — *partial* |
| attribution 도달성 (§6) | ✅ 3/3 (NOTICE / README / ADR Index) |

**판정**: Cycle 1 은 *Skill Completion Cycle quality gate* 의 *minimum viable subset* 통과. *full pass* 는 Cycle 2 에서 §4-§5 확장 시 가능.

---

## §10. Addendum — Cycle 2 structural verification

Live dispatch 는 token cost 6-8x 로 별 세션 권장 — 본 cycle 에서는 *PROCEDURE.md 본문 정합성 + cascade chain 명시* 만 spot-check.

### §10.1 §4 single-skill 5 정합성

| skill | scenarios doc §4 expected | 실 PROCEDURE.md 정합 |
|-------|---------------------------|---------------------|
| `decide-target-market` (109줄) | Stage 1 (matrix) ~ Stage 4 (ADR), §5 산출 형식, §6 self-check 6 항목 | ✅ 4 stage + §5 + §6 모두 명시 |
| `review-legal-regulatory` (173줄) | 7 sub-domain + 3 sub-skill cascade (privacy / IP / AI) | ✅ 7 영역 표 + 3 sub-skill `(구현됨)` 라벨 명시 |
| `audit-test-coverage-meaningful` (163줄) | 4 차원 trust score + agent-evaluation OMAS v2 ref | ✅ line / mutation / behavior / edge case + `agent-evaluation (외부 reference, MIT)` 본문 인용 |
| `optimize-conversion-funnel` (124줄) | AARRR 5 단계 + 5 CRO sub-domain + marketingskills attribution | ✅ Pirate Metrics (AARRR) 5 단계 + onboarding / form / page / paywall-upgrade / popup 5 sub-domain + marketingskills 본문 인용 |
| `archive-product` (169줄) | 6~12 month timeline + migrate-customers cascade + GDPR data portability | ✅ Stage 2 timeline 표 (-3 ~ +6 month) + `migrate-customers (§9) 호출` + GDPR / 개인정보보호법 *data portability* 명시 |

### §10.2 §5 cascade chain 정합성

| skill | `## 다음 phase` 명시 |
|-------|--------------------|
| `concretize-idea` | ✅ → `define-features` (2단계) |
| `decide-target-market` | ✅ §7. 다음 phase 섹션 존재 |
| `review-legal-regulatory` | ✅ 분기 명시 (글로벌 → 다음 phase / Korea → cluster trigger 등) |
| `design-system` | ✅ → `plan-build` (4단계) |

### §10.3 판정

- **Cycle 2 structural verification**: ✅ pass (5/5 single-skill 본문 정합 + 4/4 cascade chain 명시).
- **Cycle 2 live dispatch**: ❌ 미수행 — 별 세션 trigger 시 진행. session token cost 6-8x 예상.

---

## §11. References

- `docs/notes/2026-05-10-dogfood-validation-scenarios.md` — 본 결과의 검증 scenario doc
- `docs/superpowers/decisions/README.md` — ADR Index (3 ADR)
- `docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md` — ADR-001 (description baseline 제외 정책)
- `docs/superpowers/decisions/2026-05-10-superpowers-attribution.md` — ADR-003 (NOTICE attribution 정책)
- `plugin/skills/concretize-idea/PROCEDURE.md` — Stage 1 gate / 8-stage flow
- `plugin/skills/validate-idea/PROCEDURE.md` — anti-sycophancy / forcing question
- `NOTICE` — 8 외부 자산 attribution
