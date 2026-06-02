# Plugin Skills — Classification Matrix

> **목적**: 전체 스킬을 Phase별로 분류하고, 영역(engineering/product/marketing)과 standalone 등급을 부여한다.
> Input/Output Contract 작성의 선행 작업.
>
> **기준 문서**: [`engineering-phases.md`](../plugin/skills/router/references/engineering-phases.md)
> **출처**: [`skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) 의 현재 phase 배치

---

## §0. 요약

| 구분 | 수량 | 비고 |
|------|------|------|
| 전체 스킬 | 152 | orchestrator 9 + sub-orchestrator 2 + stage 134 + cross-cutting 7 |
| Engineering | **115** | Input/Output Contract 작성 대상 |
| Product/Analytics | 21 | §1 business validation + §8 product analytics |
| Marketing | 4 | §8 marketing 전용 |
| Meta/Tooling | 12 | context 관리, 패턴 라이브러리, 도구 |

**영역 분류 기준:**
- **Engineering**: 코드, 아키텍처, 테스트, 배포, 운영 등 기술적 산출물을 직접 다루는 스킬
- **Product**: 사용자/시장/비즈니스 분석, 실험 설계 등 제품 의사결정 스킬
- **Marketing**: 마케팅 콘텐츠, 채널, SEO 등 마케팅 전용 스킬
- **Meta/Tooling**: 스킬 시스템 자체의 운영, 컨텍스트 관리, 패턴 라이브러리

---

## §1. Phase Orchestrators (9) + Sub-Orchestrators (2)

| Skill | Phase | 영역 | Standalone |
|-------|-------|------|-----------|
| `concretize-idea` | §1 (Mode A) | product | orchestrator |
| `assess-product-change` | §1 (Mode B) | **engineering** | orchestrator *(신규, 미구현)* |
| `define-features` | §2 | engineering | orchestrator |
| `design-system` | §3 | engineering | orchestrator |
| `plan-build` | §4 | engineering | orchestrator |
| `build-feature` | §5 | engineering | orchestrator |
| `verify-quality` | §6 | engineering | orchestrator |
| `ship-release` | §7 | engineering | orchestrator |
| `iterate-product` | §8 | product | orchestrator |
| `manage-lifecycle` | §9 | engineering | orchestrator |
| `autoplan` | cross-phase | engineering | sub-orchestrator |
| `finish-development-branch` | §5→§7 bridge | engineering | sub-orchestrator |

---

## §2. Phase 1 — Problem/Opportunity Validation

### Mode A (Greenfield) Stage Skills

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `validate-idea` | product | full-standalone | knowledge만 필요 (아이디어 설명) |
| `validate-advanced-edge-idea` | product | standalone-with-context | validate-idea 산출물 권장 |
| `assess-business-viability` | product | full-standalone | knowledge만 필요 |
| `analyze-market-size` | product | full-standalone | knowledge만 필요 |
| `map-customer-segments` | product | full-standalone | knowledge만 필요 |
| `map-jobs-to-be-done` | product | full-standalone | knowledge만 필요 |
| `conduct-customer-interview` | product | full-standalone | knowledge만 필요 |
| `analyze-competition-and-substitutes` | product | full-standalone | knowledge만 필요 |
| `decide-target-market` | product | standalone-with-context | assess-business-viability 산출물 권장 |
| `review-legal-regulatory` | engineering (compliance) | standalone-with-context | 제품 맥락 필요 |
| `review-pricing-and-gtm` | product | standalone-with-context | assess-business-viability 산출물 권장 |
| `define-product-spec` | product → engineering bridge | orchestrator-preferred | 선행 stage 산출물 종합 |

### Mode B (Existing Product) Stage Skills

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `assess-product-change` | **engineering** | full-standalone | change trigger + codebase → scope 판단. *(신규, 미구현)* |

**Phase 1 engineering 스킬**: 2개 (`review-legal-regulatory`, `assess-product-change`)
**Phase 1 product 스킬**: 11개

---

## §3. Phase 2 — Feature Definition

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `identify-actors` | engineering | standalone-with-context | PRD 권장, 없으면 사용자에게 질의 |
| `map-actor-use-cases` | engineering | standalone-with-context | actor list 필요, 없으면 질의 |
| `map-use-case-to-system-boundary` | engineering | orchestrator-preferred | actor + use case 필요 |
| `compose-feature-from-use-cases` | engineering | orchestrator-preferred | actor + use case + boundary 복수 artifact 의존 |
| `define-feature-spec` | engineering | standalone-with-context | feature 설명 있으면 단독 가능 |
| `score-feature-priority` | engineering | standalone-with-context | feature 목록 있으면 단독 가능 |
| `estimate-feature-effort` | engineering | standalone-with-context | feature spec 있으면 단독 가능 |
| `map-feature-dependencies` | engineering | orchestrator-preferred | feature backlog 전체 필요 |
| `split-work-into-features` | engineering | standalone-with-context | PRD 또는 큰 scope 설명 있으면 가능 |
| `query-feature-registry` | engineering | standalone-with-context | feature 후보 설명 있으면 가능 |
| `triage-work-items` | engineering | full-standalone | work item 설명만 있으면 가능 |

**Phase 2 engineering 스킬**: 11개 (전부)

---

## §4. Phase 3 — Technical Design

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `define-tech-stack` | engineering | standalone-with-context | 요구사항 질의로 대체 가능 |
| `design-data-model` | engineering | standalone-with-context | entity/패턴 질의로 대체 가능 |
| `design-api-contract` | engineering | standalone-with-context | actor/endpoint 질의로 대체 가능 |
| `design-event-schema` | engineering | standalone-with-context | 이벤트 목록 질의 가능 |
| `design-auth-model` | engineering | standalone-with-context | 인증 요구사항 질의 가능 |
| `design-tenant-model` | engineering | standalone-with-context | 멀티테넌트 요구사항 질의 가능 |
| `design-observability` | engineering | standalone-with-context | 모니터링 요구사항 질의 가능 |
| `design-secret-management` | engineering | standalone-with-context | 시크릿 범위 질의 가능 |
| `design-i18n-strategy` | engineering | standalone-with-context | i18n 요구사항 질의 가능 |
| `design-accessibility-baseline` | engineering | standalone-with-context | 접근성 요구사항 질의 가능 |
| `design-deploy-strategy` | engineering | standalone-with-context | 배포 환경 질의 가능 |
| `design-artifact-storage` | engineering | standalone-with-context | artifact 유형 질의 가능 |
| `design-billing-system` | engineering | standalone-with-context | 결제 요구사항 질의 가능 |
| `design-embedding-search` | engineering | standalone-with-context | 검색 요구사항 질의 가능 |
| `design-mcp-server` | engineering | standalone-with-context | MCP 요구사항 질의 가능 |
| `design-claude-hooks` | engineering | standalone-with-context | hook 요구사항 질의 가능 |
| `design-interaction-pattern` | engineering | standalone-with-context | UX 요구사항 질의 가능 |
| `map-use-cases-to-infra` | engineering | orchestrator-preferred | use case map + system boundary 복수 의존 |
| `derive-system-topology` | engineering | orchestrator-preferred | use case + boundary + tech stack 복수 의존 |
| `decide-form-factor-app-vs-web` | engineering | standalone-with-context | 제품 요구사항 질의 가능 |
| `apply-design-system` | engineering | standalone-with-context | 디자인 시스템 맥락 필요 |
| `consult-design-system` | engineering | full-standalone | 설명만 있으면 디자인 시스템 생성 |
| `audit-ui-quality` | engineering | full-standalone | 기존 UI 있으면 바로 감사 |
| `prototype-from-spec` | engineering | standalone-with-context | spec 또는 설명 필요 |
| `write-adr` | engineering | full-standalone | 결정 내용만 있으면 작성 |
| `consult-codex` | engineering | full-standalone | 질문만 있으면 호출 |
| `verify-best-alternative` | engineering | standalone-with-context | 결정 대상 맥락 필요 |
| `critique-plan` | engineering | standalone-with-context | plan 문서 또는 설명 필요 |
| `review-architecture` | engineering | standalone-with-context | 아키텍처 맥락 필요 |
| `review-engineering` | engineering | standalone-with-context | plan 맥락 필요 |
| `review-scope` | engineering | standalone-with-context | scope 맥락 필요 |
| `review-design` | engineering | standalone-with-context | design 맥락 필요 |
| `review-devex` | engineering | standalone-with-context | DX 맥락 필요 |

**Phase 3 engineering 스킬**: 33개 (전부)

---

## §5. Phase 4 — Implementation Planning

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `decompose-feature-to-actor-tracks` | engineering | standalone-with-context | feature 설명 질의 가능 |
| `decompose-track-to-tasks` | engineering | standalone-with-context | track 설명 질의 가능 |
| `map-task-dependencies` | engineering | orchestrator-preferred | task 전체 목록 필요 |
| `plan-parallel-execution` | engineering | orchestrator-preferred | task DAG 필요 |
| `define-acceptance-test-plan` | engineering | standalone-with-context | feature/요구사항 질의 가능 |
| `estimate-build-timeline` | engineering | orchestrator-preferred | task DAG + effort 필요 |
| `publish-to-tracker` | engineering | orchestrator-preferred | task plan artifact 필요 |

**Phase 4 engineering 스킬**: 7개 (전부)

---

## §6. Phase 5 — Development

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `build-with-tdd` | engineering | **full-standalone** | "무엇을 구현할지"만 있으면 즉시 시작 |
| `diagnose-bug` | engineering | **full-standalone** | "무엇이 문제인지"만 있으면 즉시 시작 |
| `iterate-fix-verify` | engineering | **full-standalone** | finding 목록만 있으면 즉시 시작 |
| `pair-program-loop` | engineering | **full-standalone** | 작업 대상만 있으면 시작 |
| `refactor-with-rename-trace` | engineering | **full-standalone** | 대상 식별자만 있으면 시작 |
| `dispatch-parallel-agents` | engineering | standalone-with-context | task 목록 필요, 질의 가능 |
| `generate-from-api-contract` | engineering | standalone-with-context | API contract 또는 spec 필요 |
| `generate-tests-from-spec` | engineering | standalone-with-context | acceptance criteria 필요 |
| `freeze-edit-scope` | engineering (meta) | **full-standalone** | 디렉토리 지정만 필요 |
| `update-docs-with-code` | engineering | **full-standalone** | 코드 변경 diff에서 자동 감지 |

**Phase 5 engineering 스킬**: 10개 (전부). **6개가 full-standalone** — Phase 중 가장 독립적.

---

## §7. Phase 6 — Verification & Quality

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `audit-security` | engineering | **full-standalone** | 기존 코드 대상 즉시 감사 |
| `audit-test-coverage-meaningful` | engineering | **full-standalone** | 기존 테스트 대상 즉시 감사 |
| `audit-ubiquitous-language` | engineering | **full-standalone** | 기존 코드 어휘 즉시 감사 |
| `measure-code-health` | engineering | **full-standalone** | 기존 코드 즉시 측정 |
| `audit-accessibility` | engineering | standalone-with-context | UI 존재 필요 |
| `audit-cost-efficiency` | engineering | standalone-with-context | 인프라 정보 필요 |
| `audit-i18n-coverage` | engineering | standalone-with-context | i18n 설정 존재 필요 |
| `audit-live-devex` | engineering | standalone-with-context | 배포된 제품 필요 |
| `chaos-test` | engineering | standalone-with-context | 대상 시스템 정보 필요 |
| `run-load-test` | engineering | standalone-with-context | 대상 endpoint 정보 필요 |
| `run-browser-qa` | engineering | standalone-with-context | UI URL 필요 |
| `test-per-actor-use-case` | engineering | orchestrator-preferred | actor + use case 맵 필요 |
| `test-cross-actor-flow` | engineering | orchestrator-preferred | cross-actor flow 정의 필요 |
| `classify-qa-tiers` | engineering (meta) | **full-standalone** | QA 대상 분류만 |
| `classify-review-risks` | engineering (meta) | **full-standalone** | diff 대상 분류만 |
| `review-ai-safety-liability` | engineering (compliance) | standalone-with-context | AI 기능 맥락 필요 |
| `review-privacy-data-risk` | engineering (compliance) | standalone-with-context | 데이터 처리 맥락 필요 |
| `review-license-and-ip-risk` | engineering (compliance) | standalone-with-context | 의존성 목록 필요 |
| `review-terms-policy-readiness` | engineering (compliance) | standalone-with-context | 제품/서비스 맥락 필요 |

**Phase 6 engineering 스킬**: 19개 (전부). **6개 full-standalone**.

---

## §8. Phase 7 — Release

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `setup-quality-gates` | engineering | standalone-with-context | 프로젝트 환경 필요 |
| `auto-create-pr` | engineering | **full-standalone** | 현재 branch에서 즉시 실행 |
| `automate-release-tagging` | engineering | standalone-with-context | PR 정보 필요 |
| `sync-release-docs` | engineering | standalone-with-context | diff 정보 필요 |
| `write-changelog` | engineering | standalone-with-context | 버전 정보 필요 |
| `guard-destructive-commands` | engineering (meta) | **full-standalone** | ambient 적용 |
| `compose-safety-mode` | engineering (meta) | **full-standalone** | ambient 적용 |
| `run-uat` | engineering | orchestrator-preferred | UAT 시나리오 + 이해관계자 필요 |
| `run-beta-program` | engineering | orchestrator-preferred | beta 코호트 + 피드백 구조 필요 |
| `setup-canary-deploy` | engineering | standalone-with-context | 배포 환경 질의 가능 |
| `setup-feature-flags` | engineering | standalone-with-context | feature 목록 질의 가능 |
| `setup-rollback-runbook` | engineering | standalone-with-context | 배포 환경 질의 가능 |
| `prepare-launch-checklist` | engineering | orchestrator-preferred | 다수 선행 artifact (QA, 보안, 문서 등) |
| `setup-incident-paging` | engineering | standalone-with-context | 팀/인프라 구조 질의 가능 |

**Phase 7 engineering 스킬**: 14개 (전부).

---

## §9. Phase 8 — Operations & Iteration

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `handle-incident` | **engineering (SRE)** | **full-standalone** | 인시던트 발생 시 즉시 시작 |
| `conduct-postmortem` | **engineering (SRE)** | standalone-with-context | 인시던트 정보 필요 |
| `monitor-regressions` | **engineering (SRE)** | **full-standalone** | ambient 모니터링 |
| `analyze-actor-failure-rate` | **engineering (SRE)** | standalone-with-context | 운영 데이터 필요 |
| `analyze-cost-anomaly` | **engineering (SRE)** | standalone-with-context | 비용 데이터 필요 |
| `audit-error-budget` | **engineering (SRE)** | standalone-with-context | SLO 정의 + 운영 데이터 필요 |
| `summarize-retro` | **engineering** | **full-standalone** | git history에서 자동 생성 |
| `design-ab-experiment` | product | standalone-with-context | 가설 + 제품 맥락 필요 |
| `analyze-ab-experiment` | product | standalone-with-context | 실험 데이터 필요 |
| `analyze-user-funnel` | product | standalone-with-context | funnel 정의 + 데이터 필요 |
| `generate-improvement-tasks` | product → engineering bridge | standalone-with-context | 분석 결과 필요 |
| `analyze-feature-adoption` | product | standalone-with-context | usage 데이터 필요 |
| `analyze-user-cohort` | product | standalone-with-context | cohort 데이터 필요 |
| `triage-customer-support-ticket` | product | standalone-with-context | 티켓 내용 필요 |
| `analyze-customer-feedback-corpus` | product | standalone-with-context | 피드백 corpus 필요 |
| `optimize-conversion-funnel` | product | standalone-with-context | funnel 데이터 필요 |
| `plan-growth-experiment` | product | standalone-with-context | growth 맥락 필요 |
| `draft-marketing-copy` | marketing | standalone-with-context | 제품/타겟 맥락 필요 |
| `plan-marketing-channel` | marketing | standalone-with-context | 제품/타겟 맥락 필요 |
| `audit-seo-aso` | marketing | standalone-with-context | 사이트/앱 URL 필요 |
| `automate-marketing-content` | marketing | standalone-with-context | 콘텐츠 전략 필요 |
**Phase 8 engineering 스킬**: 7개. **Product**: 10개. **Marketing**: 4개.

*`save-context`, `restore-context`, `persist-learning-jsonl`은 Cross-cutting으로 재분류 (§12 참조).*

---

## §10. Phase 9 — Lifecycle Management

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `deprecate-feature` | engineering | standalone-with-context | feature + usage 정보 필요 |
| `migrate-customers` | engineering | standalone-with-context | migration 대상 정보 필요 |
| `archive-product` | engineering | standalone-with-context | product 정보 필요 |
| `spin-off-feature` | engineering | standalone-with-context | feature + 분리 근거 필요 |

**Phase 9 engineering 스킬**: 4개 (전부).

---

## §11. Cross-cutting Utilities

| Skill | 영역 | Standalone | 비고 |
|-------|------|-----------|------|
| `decompose-blocker` | engineering | **full-standalone** | stuck 상태에서 자동 trigger |
| `status` | engineering (meta) | **full-standalone** | artifact 탐지로 즉시 실행 |
| `write-a-skill` | engineering (meta) | **full-standalone** | 스킬 설명만 있으면 작성 |
| `apply-builder-ethos` | meta (philosophy) | **full-standalone** | ambient 적용 |
| `benchmark-llm-models` | engineering (tooling) | standalone-with-context | 비교 대상 모델 필요 |
| `detect-install-type` | engineering (tooling) | **full-standalone** | 자동 감지 |
| `guide-setup-wizard` | engineering (tooling) | standalone-with-context | 설정 대상 필요 |
| `save-context` | meta | **full-standalone** | 세션 상태 저장. §8에서 재분류 |
| `restore-context` | meta | **full-standalone** | 체크포인트 로드. §8에서 재분류 |
| `persist-learning-jsonl` | meta (pattern library) | **full-standalone** | JSONL 학습 저장 패턴. §8에서 재분류 |
| `review-legal-regulatory` | engineering (compliance) | standalone-with-context | §1/§7 등 복수 phase 호출. §1에서 재분류 |

**Cross-cutting 스킬**: 11개 (기존 7 + 재분류 4).

---

## §12. 재분류 결과 (확정 2026-05-26)

| Skill | 변경 전 | 변경 후 | 근거 (PROCEDURE.md 확인) |
|-------|--------|--------|------------------------|
| `save-context` | §8 Operate | **→ Cross-cutting** | "다른 작업으로 전환, /clear 전, 세션 종료" — phase 무관 세션 관리 |
| `restore-context` | §8 Operate | **→ Cross-cutting** | save-context 자매. cross-branch 명시 |
| `persist-learning-jsonl` | §8 Operate | **→ Cross-cutting** | "AI 어시스턴트용 프로젝트 메모리 레이어" — phase 무관 패턴 라이브러리 |
| `review-legal-regulatory` | §1 Idea | **→ Cross-cutting (compliance)** | §1 + §7 + 신규 기능/pivot 시 — 복수 phase 호출 |
| `summarize-retro` | §8 Operate | **§8 유지** | 회고는 운영/개선 사이클 관행. 주 용도가 §8 |
| `monitor-regressions` | §8 Operate | **§8 유지** | 이전 세션 G2 결정 유지. production 모니터링이 primary |
| `generate-improvement-tasks` | §8 Operate | **§8 유지** | "8단계 stage skill" 명시 + §8→§2 bridge 역할 |

---

## §13. Standalone 등급 요약

| 등급 | 수량 | 비율 |
|------|------|------|
| **full-standalone** | 35 | 23% |
| **standalone-with-context** | 82 | 54% |
| **orchestrator-preferred** | 15 | 10% |
| **orchestrator** (phase/sub) | 11 | 7% |
| **meta (ambient)** | 9 | 6% |
| **합계** | 152 | 100% |
