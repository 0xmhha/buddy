#!/usr/bin/env python3
"""Add Input Requirements and Output Contract sections to PROCEDURE.md files."""

import re
from pathlib import Path

SKILLS_DIR = Path("/Users/wm-it-22-00661/Work/github/study/ai/buddy/plugin/skills")

# Phase -> skill -> (inputs, outputs)
# inputs: list of (name, required, type, source, fallback_question)
# outputs: list of (name, type, format, consumers)
CONTRACTS = {
    # §3 Technical Design
    "define-tech-stack": {
        "inputs": [
            ("Feature 요구사항 또는 제품 설명", "✅", "knowledge", "사용자 도메인 지식 또는 Phase 2 산출물", '"어떤 제품/시스템의 기술 스택을 결정하나요? 핵심 요구사항을 설명해 주세요."'),
        ],
        "outputs": [
            ("Tech stack decision (8차원 비교 + lock-in 평가)", "artifact", "structured YAML + ADR draft", "`write-adr`, `design-data-model`, `design-api-contract`"),
        ],
    },
    "design-data-model": {
        "inputs": [
            ("Entity 목록 또는 도메인 설명", "✅", "artifact / knowledge", "Phase 2 `define-feature-spec` 또는 사용자 설명", '"시스템의 핵심 entity(데이터 객체)는 무엇인가요?"'),
            ("Read/write 패턴", "✅", "knowledge", "사용자 도메인 지식", '"주요 데이터 접근 패턴은? (OLTP 위주 / 분석 위주 / 혼합)"'),
            ("Tech stack 결정", "선택", "decision", "`define-tech-stack` 산출물", "없으면 DB 종류 무관 범용 설계"),
        ],
        "outputs": [
            ("Data model (ERD + schema DDL + index 전략 + migration plan)", "artifact", "structured YAML + DDL", "`generate-from-api-contract`, `build-with-tdd`"),
        ],
    },
    "design-api-contract": {
        "inputs": [
            ("Actor-use case map 또는 API 대상 설명", "✅", "artifact / knowledge", "`map-actor-use-cases` 산출물 또는 사용자 설명", '"API를 사용하는 actor와 use case를 알려주세요."'),
            ("Tech stack 결정", "선택", "decision", "`define-tech-stack` 산출물", "없으면 REST 기본 가정"),
        ],
        "outputs": [
            ("API contract (style + resource + schema + error taxonomy + versioning)", "artifact", "OpenAPI spec 또는 structured YAML", "`generate-from-api-contract`, `test-per-actor-use-case`, `design-event-schema`"),
        ],
    },
    "design-event-schema": {
        "inputs": [
            ("시스템 토폴로지 또는 서비스 간 통신 맥락", "✅", "artifact / knowledge", "`derive-system-topology` 산출물 또는 사용자 설명", '"어떤 서비스 간 비동기 통신이 필요한가요?"'),
        ],
        "outputs": [
            ("Event schema (event 목록 + payload + routing + DLQ 정책)", "artifact", "structured YAML", "`build-feature`"),
        ],
    },
    "design-auth-model": {
        "inputs": [
            ("Actor list 또는 사용자 유형 설명", "✅", "artifact / knowledge", "`identify-actors` 산출물 또는 사용자 설명", '"시스템의 사용자 유형과 권한 요구사항을 알려주세요."'),
        ],
        "outputs": [
            ("Auth/authz 모델 (인증 방식 + 권한 체계 + 세션 전략)", "artifact", "structured YAML + ADR", "`build-feature`, `audit-security`"),
        ],
    },
    "design-tenant-model": {
        "inputs": [
            ("멀티테넌트 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"멀티테넌트가 필요한가요? 테넌트 간 격리 수준은?"'),
        ],
        "outputs": [
            ("Tenant 모델 (RLS / schema-per / DB-per 결정 + 격리 전략)", "artifact", "structured YAML + ADR", "`design-data-model`, `build-feature`"),
        ],
    },
    "design-observability": {
        "inputs": [
            ("시스템 토폴로지", "선택", "artifact", "`derive-system-topology` 산출물", "없으면 단일 서비스 가정"),
            ("SLO 목표", "✅", "knowledge", "사용자 도메인 지식", '"핵심 SLO 지표는? (가용성 99.9%, 응답시간 p99 < 500ms 등)"'),
        ],
        "outputs": [
            ("Observability 전략 (logs/metrics/traces/SLO 구성)", "artifact", "structured YAML", "`setup-incident-paging`, `audit-error-budget`"),
        ],
    },
    "design-secret-management": {
        "inputs": [
            ("시크릿 인벤토리", "✅", "knowledge", "사용자 도메인 지식", '"관리해야 할 시크릿 종류는? (API key, DB 비밀번호, 인증서 등)"'),
        ],
        "outputs": [
            ("Secret management 전략 (저장소 + rotation + audit + leak detection)", "artifact", "structured YAML", "`audit-security`, `build-feature`"),
        ],
    },
    "design-i18n-strategy": {
        "inputs": [
            ("Target market / 지원 언어", "✅", "knowledge", "`decide-target-market` 산출물 또는 사용자 설명", '"지원할 언어와 지역은?"'),
        ],
        "outputs": [
            ("i18n 전략 (locale 구조 + fallback + RTL + format)", "artifact", "structured YAML", "`audit-i18n-coverage`, `build-feature`"),
        ],
    },
    "design-accessibility-baseline": {
        "inputs": [
            ("Target 사용자 및 접근성 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"접근성 목표 수준은? (WCAG 2.1 AA 등) 주요 사용자 특성은?"'),
        ],
        "outputs": [
            ("Accessibility baseline (WCAG 레벨 + 검증 도구 + annotation 가이드)", "artifact", "structured YAML", "`audit-accessibility`, `build-feature`"),
        ],
    },
    "design-deploy-strategy": {
        "inputs": [
            ("배포 환경 및 리스크 허용도", "✅", "knowledge", "사용자 도메인 지식", '"배포 환경은? (클라우드/온프레미스) 허용 가능한 downtime은?"'),
        ],
        "outputs": [
            ("Deploy 전략 (canary/blue-green/rolling + env 분리 + IaC)", "artifact", "structured YAML + ADR", "`setup-canary-deploy`, `setup-rollback-runbook`"),
        ],
    },
    "design-artifact-storage": {
        "inputs": [
            ("Artifact 유형 및 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"저장/배포할 artifact 종류는? (binary, container, template 등)"'),
        ],
        "outputs": [
            ("Artifact storage 설계 (저장소 + 버전관리 + 검증 + 배포)", "artifact", "structured YAML", "`ship-release`"),
        ],
    },
    "design-billing-system": {
        "inputs": [
            ("Pricing model 및 결제 요구사항", "✅", "knowledge", "`review-pricing-and-gtm` 산출물 또는 사용자 설명", '"결제 모델은? (구독/종량/일회성) 결제 수단은?"'),
        ],
        "outputs": [
            ("Billing system 설계 (PSP 연동 + subscription + metering + invoice)", "artifact", "structured YAML", "`build-feature`"),
        ],
    },
    "design-embedding-search": {
        "inputs": [
            ("검색 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"어떤 데이터를 검색하나요? 예상 규모와 정확도 요구사항은?"'),
        ],
        "outputs": [
            ("Hybrid search 설계 (BM25 + vector + rerank + metadata filter)", "artifact", "structured YAML", "`build-feature`"),
        ],
    },
    "design-mcp-server": {
        "inputs": [
            ("MCP 서버 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"어떤 도구/리소스를 MCP로 노출하나요?"'),
        ],
        "outputs": [
            ("MCP server 설계 (tool surface + transport + auth)", "artifact", "structured YAML", "`build-feature`"),
        ],
    },
    "design-claude-hooks": {
        "inputs": [
            ("Hook 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"어떤 Claude Code 이벤트에 hook을 걸고 싶나요? (PreToolUse, PostToolUse 등)"'),
        ],
        "outputs": [
            ("Hook 설계 (event + matcher + command + scope)", "artifact", "structured YAML", "`build-feature`"),
        ],
    },
    "design-interaction-pattern": {
        "inputs": [
            ("UX 요구사항", "✅", "knowledge", "사용자 도메인 지식", '"어떤 사용자 인터랙션을 설계하나요? (gesture, motion, feedback 등)"'),
        ],
        "outputs": [
            ("Interaction pattern 설계 (gesture + motion + feedback 규칙)", "artifact", "structured YAML", "`build-feature`, `apply-design-system`"),
        ],
    },
    "map-use-cases-to-infra": {
        "inputs": [
            ("Use case map", "✅", "artifact", "`map-actor-use-cases` 산출물", "먼저 `/buddy:map-actor-use-cases` 를 실행하세요"),
            ("System boundary map", "✅", "artifact", "`map-use-case-to-system-boundary` 산출물", "먼저 `/buddy:map-use-case-to-system-boundary` 를 실행하세요"),
        ],
        "outputs": [
            ("Infra mapping (use case별 실행 인프라 매핑)", "artifact", "structured YAML", "`derive-system-topology`"),
        ],
    },
    "derive-system-topology": {
        "inputs": [
            ("Use case map + system boundary", "✅", "artifact", "`map-use-cases-to-infra` 또는 Phase 2 산출물", '"시스템의 서비스 구성과 데이터 흐름을 알려주세요."'),
            ("Tech stack 결정", "선택", "decision", "`define-tech-stack` 산출물", "없으면 범용 topology 도출"),
        ],
        "outputs": [
            ("System topology (service graph + data flow + trust boundary)", "artifact", "structured YAML + 다이어그램", "`design-deploy-strategy`, `design-observability`, `design-event-schema`"),
        ],
    },
    "decide-form-factor-app-vs-web": {
        "inputs": [
            ("제품 요구사항 및 target 사용자", "✅", "knowledge", "사용자 도메인 지식", '"제품의 주요 사용 환경은? (모바일/데스크톱/오프라인 등)"'),
        ],
        "outputs": [
            ("Form factor 결정 (app/web/hybrid/desktop + 근거)", "decision", "ADR 또는 inline record", "`define-tech-stack`"),
        ],
    },
    "apply-design-system": {
        "inputs": [
            ("Design system 토큰/컴포넌트", "✅", "artifact / knowledge", "`consult-design-system` 산출물 또는 기존 디자인 시스템", '"적용할 디자인 시스템이 있나요? 없으면 생성합니다."'),
        ],
        "outputs": [
            ("적용된 디자인 (토큰 + 컴포넌트 매핑)", "artifact", "structured YAML + 코드", "`build-feature`"),
        ],
    },
    "consult-design-system": {
        "inputs": [
            ("제품/브랜드 설명", "✅", "knowledge", "사용자 도메인 지식", '"어떤 제품의 디자인 시스템을 만들까요? 브랜드 톤을 설명해 주세요."'),
        ],
        "outputs": [
            ("Design system 문서 (토큰 + 컴포넌트 + 패턴)", "artifact", "structured document", "`apply-design-system`"),
        ],
    },
    "audit-ui-quality": {
        "inputs": [
            ("기존 UI", "✅", "artifact", "배포된 UI 또는 로컬 dev 서버", '"감사할 UI의 URL 또는 경로를 알려주세요."'),
        ],
        "outputs": [
            ("UI quality audit report", "artifact", "structured report", "`iterate-fix-verify`"),
        ],
    },
    "prototype-from-spec": {
        "inputs": [
            ("Feature spec 또는 설명", "✅", "artifact / knowledge", "`define-feature-spec` 산출물 또는 사용자 설명", '"어떤 기능의 프로토타입을 만드나요?"'),
        ],
        "outputs": [
            ("프로토타입 (탐색용 코드 또는 디자인)", "artifact", "코드 / HTML / 디자인 파일", "`review-design`"),
        ],
    },
    "write-adr": {
        "inputs": [
            ("결정 내용 (무엇을, 왜, 어떤 대안을 고려했는지)", "✅", "knowledge", "사용자 또는 선행 design-* 스킬 산출물", '"어떤 결정을 기록하나요? 배경과 선택지를 설명해 주세요."'),
        ],
        "outputs": [
            ("ADR 문서 (Status/Context/Decision/Consequences/Alternatives)", "artifact", "`docs/decisions/` markdown", "전 phase 참조"),
        ],
    },
    "consult-codex": {
        "inputs": [
            ("질문 또는 검토 대상", "✅", "knowledge", "사용자 발화", '"무엇에 대해 second opinion을 구하나요?"'),
        ],
        "outputs": [
            ("외부 LLM second opinion", "artifact", "structured review / answer", "의사결정 지원"),
        ],
    },
    "verify-best-alternative": {
        "inputs": [
            ("결정 대상 + 현재 선택안", "✅", "knowledge", "사용자 또는 선행 design-* 스킬", '"어떤 결정을 검토하나요? 현재 선택한 안은?"'),
        ],
        "outputs": [
            ("Multi-perspective review (N개 대안 + rubric 비교)", "artifact", "structured comparison", "`write-adr`"),
        ],
    },
    "critique-plan": {
        "inputs": [
            ("Plan 문서 또는 설명", "✅", "artifact / knowledge", "사용자 또는 선행 plan 스킬", '"비판적 검토할 plan을 공유해 주세요."'),
        ],
        "outputs": [
            ("Strategic critique (CEO/founder 관점)", "artifact", "structured findings", "plan 수정"),
        ],
    },
    "review-architecture": {
        "inputs": [
            ("아키텍처 맥락 (코드 또는 설계 문서)", "✅", "artifact / knowledge", "기존 코드베이스 또는 Phase 3 산출물", '"검토할 아키텍처의 범위를 알려주세요."'),
        ],
        "outputs": [
            ("Architecture review findings", "artifact", "structured findings", "`iterate-fix-verify`"),
        ],
    },
    "review-engineering": {
        "inputs": [
            ("Plan 또는 구현 맥락", "✅", "artifact / knowledge", "Phase 4 산출물 또는 PR diff", '"검토할 plan이나 구현 범위를 알려주세요."'),
        ],
        "outputs": [
            ("Engineering review findings (anti-rationalization 5 규칙 적용)", "artifact", "structured findings", "`iterate-fix-verify`"),
        ],
    },
    "review-scope": {
        "inputs": [
            ("Scope 맥락 (PRD 또는 plan)", "✅", "artifact / knowledge", "Phase 1-4 산출물 또는 사용자 설명", '"범위 검토할 대상을 알려주세요."'),
        ],
        "outputs": [
            ("Scope review findings", "artifact", "structured findings", "plan 수정"),
        ],
    },
    "review-design": {
        "inputs": [
            ("Design 맥락 (UI/UX 설계 또는 plan)", "✅", "artifact / knowledge", "Phase 3 산출물 또는 사용자 설명", '"디자인 검토할 대상을 알려주세요."'),
        ],
        "outputs": [
            ("Design review (차원별 0-10 score + 개선 path)", "artifact", "structured scores + findings", "plan 수정"),
        ],
    },
    "review-devex": {
        "inputs": [
            ("DX 맥락 (developer-facing 제품 또는 API)", "✅", "artifact / knowledge", "Phase 3 산출물 또는 사용자 설명", '"DX 검토할 developer-facing 제품을 알려주세요."'),
        ],
        "outputs": [
            ("DX review findings (persona/competitor/friction map)", "artifact", "structured findings", "plan 수정"),
        ],
    },
    # §4 Implementation Planning
    "decompose-feature-to-actor-tracks": {
        "inputs": [
            ("Feature specs", "✅", "artifact / knowledge", "`define-feature-spec` 산출물 또는 사용자 설명", '"분해할 feature를 설명해 주세요."'),
        ],
        "outputs": [
            ("Actor-track decomposition (actor별 implementation track)", "artifact", "structured YAML", "`decompose-track-to-tasks`, `dispatch-parallel-agents`"),
        ],
    },
    "decompose-track-to-tasks": {
        "inputs": [
            ("Actor track", "✅", "artifact / knowledge", "`decompose-feature-to-actor-tracks` 산출물 또는 사용자 설명", '"분해할 actor track을 알려주세요."'),
        ],
        "outputs": [
            ("Ordered task list (atomic, 1 PR scope, verifiable)", "artifact", "structured YAML", "`map-task-dependencies`"),
        ],
    },
    "map-task-dependencies": {
        "inputs": [
            ("Task list", "✅", "artifact", "`decompose-track-to-tasks` 산출물", "먼저 `/buddy:decompose-track-to-tasks` 를 실행하세요"),
        ],
        "outputs": [
            ("Task DAG (internal + cross-actor edges, critical path)", "artifact", "structured YAML", "`plan-parallel-execution`, `estimate-build-timeline`"),
        ],
    },
    "plan-parallel-execution": {
        "inputs": [
            ("Task DAG", "✅", "artifact", "`map-task-dependencies` 산출물", "먼저 `/buddy:map-task-dependencies` 를 실행하세요"),
        ],
        "outputs": [
            ("Parallel execution plan (worker batch + sync points)", "artifact", "structured YAML", "`dispatch-parallel-agents`, `build-feature`"),
        ],
    },
    "define-acceptance-test-plan": {
        "inputs": [
            ("Feature specs + acceptance criteria", "✅", "artifact / knowledge", "`define-feature-spec` 산출물 또는 사용자 설명", '"테스트 계획을 세울 feature와 수용 기준을 알려주세요."'),
        ],
        "outputs": [
            ("Acceptance test plan (per-actor + cross-actor)", "artifact", "structured YAML", "`generate-tests-from-spec`, `verify-quality`"),
        ],
    },
    "estimate-build-timeline": {
        "inputs": [
            ("Task DAG + effort estimates", "✅", "artifact", "`map-task-dependencies` + `estimate-feature-effort` 산출물", "먼저 task 분해와 effort 추정을 실행하세요"),
        ],
        "outputs": [
            ("Calendar timeline (critical path + risk buffer)", "artifact", "structured YAML", "프로젝트 관리"),
        ],
    },
    "publish-to-tracker": {
        "inputs": [
            ("Task plan 또는 PRD", "✅", "artifact", "Phase 4 산출물 또는 Phase 1 PRD", "먼저 구현 계획을 수립하세요"),
        ],
        "outputs": [
            ("외부 tracker issues (GitHub/Linear/Jira)", "artifact", "issue links", "프로젝트 관리, `dispatch-parallel-agents`"),
        ],
    },
    # §5 Development
    "build-with-tdd": {
        "inputs": [
            ("구현할 기능 설명", "✅", "knowledge", "사용자 발화 또는 Phase 4 task", '"어떤 기능을 TDD로 구현하나요?"'),
        ],
        "outputs": [
            ("Working code + passing test suite + TDD cycle log", "artifact", "source files + test files", "`verify-quality`"),
        ],
    },
    "diagnose-bug": {
        "inputs": [
            ("버그 설명 (증상, 에러 메시지, 재현 조건)", "✅", "knowledge", "사용자 발화 또는 버그 리포트", '"어떤 버그인가요? 증상과 재현 방법을 알려주세요."'),
        ],
        "outputs": [
            ("Root cause analysis + fix", "artifact", "코드 수정 + commit", "`iterate-fix-verify`"),
        ],
    },
    "iterate-fix-verify": {
        "inputs": [
            ("Finding 목록 (수정할 항목들)", "✅", "knowledge / artifact", "리뷰 결과 또는 QA report", '"수정할 finding 목록을 알려주세요."'),
        ],
        "outputs": [
            ("Fixed code (finding별 atomic commit)", "artifact", "commit history", "`verify-quality`"),
        ],
    },
    "pair-program-loop": {
        "inputs": [
            ("작업 대상 설명", "✅", "knowledge", "사용자 발화", '"무엇을 함께 작업할까요?"'),
        ],
        "outputs": [
            ("Working code (driver/navigator swap 기록)", "artifact", "source files + commit", "`verify-quality`"),
        ],
    },
    "refactor-with-rename-trace": {
        "inputs": [
            ("Rename 대상 식별자", "✅", "knowledge", "사용자 발화", '"어떤 이름을 변경하나요? (변경 전 → 변경 후)"'),
        ],
        "outputs": [
            ("Renamed code (LSP + grep 검증 완료)", "artifact", "단일 commit", "`verify-quality`"),
        ],
    },
    "dispatch-parallel-agents": {
        "inputs": [
            ("Task list (병렬 실행 가능한 독립 작업들)", "✅", "artifact / knowledge", "`plan-parallel-execution` 산출물 또는 사용자 지정", '"병렬로 처리할 task 목록을 알려주세요."'),
        ],
        "outputs": [
            ("Aggregated working code (worktree별 결과 통합)", "artifact", "source files + commits", "`verify-quality`"),
        ],
    },
    "generate-from-api-contract": {
        "inputs": [
            ("API contract (OpenAPI / GraphQL / gRPC)", "✅", "artifact", "`design-api-contract` 산출물", "먼저 `/buddy:design-api-contract` 를 실행하세요"),
        ],
        "outputs": [
            ("Generated SDK / stub code (AUTO-GENERATED 헤더)", "artifact", "source files", "`build-with-tdd`"),
        ],
    },
    "generate-tests-from-spec": {
        "inputs": [
            ("Acceptance criteria 또는 test plan", "✅", "artifact / knowledge", "`define-acceptance-test-plan` 산출물 또는 feature spec", '"테스트를 생성할 acceptance criteria를 알려주세요."'),
        ],
        "outputs": [
            ("Test skeleton (unit/integration/contract/E2E)", "artifact", "test files + TODO markers", "`build-with-tdd`"),
        ],
    },
    "freeze-edit-scope": {
        "inputs": [
            ("Lock 대상 디렉토리", "✅", "knowledge", "사용자 지정", '"edit을 제한할 디렉토리 경로를 알려주세요."'),
        ],
        "outputs": [
            ("Scope lock (세션 동안 유지)", "artifact", "session state", "(ambient)"),
        ],
    },
    "update-docs-with-code": {
        "inputs": [
            ("코드 변경 diff", "✅", "artifact", "git diff 또는 최근 commit", "(자동 감지 — 현재 코드 변경에서 추출)"),
        ],
        "outputs": [
            ("Updated docs (README/ADR/CHANGELOG/HANDOFF/skill-catalog)", "artifact", "docs/ 파일 갱신", "`sync-release-docs`"),
        ],
    },
    # §6 Quality
    "classify-qa-tiers": {
        "inputs": [
            ("변경 설명 또는 diff", "✅", "knowledge / artifact", "사용자 발화 또는 git diff", '"어떤 변경에 대한 QA tier를 분류하나요?"'),
        ],
        "outputs": [
            ("QA tier 분류 (Quick/Standard/Exhaustive)", "decision", "inline record", "test 스킬 선택 근거"),
        ],
    },
    "test-per-actor-use-case": {
        "inputs": [
            ("Actor-use case map", "✅", "artifact", "`map-actor-use-cases` 산출물", '"테스트할 actor와 use case를 알려주세요."'),
            ("Working code", "✅", "artifact", "Phase 5 산출물", "(현재 코드베이스에서 자동 감지)"),
        ],
        "outputs": [
            ("Per-actor test report (coverage gap 0 유지)", "artifact", "structured report", "`verify-quality`"),
        ],
    },
    "test-cross-actor-flow": {
        "inputs": [
            ("Cross-actor flow 정의", "✅", "artifact / knowledge", "Phase 2 산출물 또는 사용자 설명", '"테스트할 cross-actor flow를 설명해 주세요. (예: signup→email→verify→login)"'),
        ],
        "outputs": [
            ("E2E test report (multi-actor chain 검증)", "artifact", "structured report", "`verify-quality`"),
        ],
    },
    "run-load-test": {
        "inputs": [
            ("Target endpoint + SLA 목표", "✅", "knowledge", "사용자 도메인 지식", '"부하 테스트 대상 endpoint와 SLA 목표(RPS, 응답시간)를 알려주세요."'),
        ],
        "outputs": [
            ("Load test report (sustained/soak/spike/stress 결과)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "chaos-test": {
        "inputs": [
            ("대상 시스템 + 실패 가설", "✅", "knowledge", "사용자 도메인 지식", '"어떤 실패 시나리오를 테스트하나요? 가설을 알려주세요."'),
        ],
        "outputs": [
            ("Chaos test report (가설 검증 결과 + blast radius)", "artifact", "structured report", "`conduct-postmortem`"),
        ],
    },
    "run-browser-qa": {
        "inputs": [
            ("UI URL 또는 로컬 서버 경로", "✅", "knowledge", "사용자 지정", '"QA할 UI의 URL을 알려주세요."'),
        ],
        "outputs": [
            ("Browser QA report (snapshot diff/form/responsive)", "artifact", "structured report + screenshots", "`iterate-fix-verify`"),
        ],
    },
    "audit-test-coverage-meaningful": {
        "inputs": [
            ("기존 테스트 코드", "✅", "artifact", "현재 코드베이스", "(자동 감지 — 테스트 파일에서 추출)"),
        ],
        "outputs": [
            ("Trust score (line 0.2 + mutation 0.4 + behavior 0.2 + edge 0.2)", "artifact", "structured report", "`iterate-fix-verify`"),
        ],
    },
    "measure-code-health": {
        "inputs": [
            ("기존 코드베이스", "✅", "artifact", "현재 코드베이스", "(자동 감지)"),
        ],
        "outputs": [
            ("Code health score (0-10 composite dashboard)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "audit-security": {
        "inputs": [
            ("기존 코드베이스", "✅", "artifact", "현재 코드베이스", "(자동 감지)"),
        ],
        "outputs": [
            ("Security audit report (OWASP + CSO-mode findings)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "audit-accessibility": {
        "inputs": [
            ("UI (배포 또는 로컬)", "✅", "artifact", "배포된 UI 또는 로컬 dev 서버", '"감사할 UI의 URL을 알려주세요."'),
        ],
        "outputs": [
            ("Accessibility audit report (WCAG 2.1 AA + axe)", "artifact", "structured report", "`iterate-fix-verify`"),
        ],
    },
    "audit-i18n-coverage": {
        "inputs": [
            ("i18n 설정 + locale 파일", "✅", "artifact", "현재 코드베이스", "(자동 감지 — i18n 설정 파일에서 추출)"),
        ],
        "outputs": [
            ("i18n coverage report (locale별 번역 누락 + fallback rate)", "artifact", "structured report", "`iterate-fix-verify`"),
        ],
    },
    "audit-cost-efficiency": {
        "inputs": [
            ("인프라 구성 정보", "✅", "artifact / knowledge", "IaC 파일 또는 사용자 설명", '"비용 분석할 인프라 구성을 알려주세요."'),
        ],
        "outputs": [
            ("Cost efficiency report (per-component + $/MAU + waste detection)", "artifact", "structured report", "`analyze-cost-anomaly`"),
        ],
    },
    "audit-live-devex": {
        "inputs": [
            ("배포된 developer product", "✅", "artifact", "배포된 제품 URL 또는 문서", '"DX audit할 제품의 접근 경로를 알려주세요."'),
        ],
        "outputs": [
            ("DX audit report (TTHW timing + evidence)", "artifact", "structured report", "`iterate-fix-verify`"),
        ],
    },
    "audit-ubiquitous-language": {
        "inputs": [
            ("코드베이스 + PRD/도메인 어휘", "✅", "artifact", "현재 코드베이스 + 도메인 문서", "(자동 감지)"),
        ],
        "outputs": [
            ("Vocabulary drift report (mismatch pairs + severity + remediation)", "artifact", "structured YAML", "`refactor-with-rename-trace`"),
        ],
    },
    "classify-review-risks": {
        "inputs": [
            ("코드 diff", "✅", "artifact", "git diff 또는 PR diff", "(자동 감지)"),
        ],
        "outputs": [
            ("Risk classification (11 category)", "artifact", "structured report", "`review-engineering`"),
        ],
    },
    "review-ai-safety-liability": {
        "inputs": [
            ("AI 기능 맥락", "✅", "knowledge", "사용자 도메인 지식", '"검토할 AI 기능의 역할과 자율성 수준을 알려주세요."'),
        ],
        "outputs": [
            ("AI safety review (hallucination/autonomy/liability)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "review-privacy-data-risk": {
        "inputs": [
            ("데이터 처리 맥락", "✅", "knowledge", "사용자 도메인 지식", '"어떤 개인정보/민감 데이터를 처리하나요? 데이터 흐름을 설명해 주세요."'),
        ],
        "outputs": [
            ("Privacy review (GDPR/PIPA 등 규제 frame)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "review-license-and-ip-risk": {
        "inputs": [
            ("의존성 목록", "✅", "artifact", "package.json / go.mod / requirements.txt 등", "(자동 감지 — 의존성 파일에서 추출)"),
        ],
        "outputs": [
            ("License review (호환성 + IP 리스크 + remediation)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "review-terms-policy-readiness": {
        "inputs": [
            ("제품/서비스 맥락", "✅", "knowledge", "사용자 도메인 지식", '"어떤 제품/서비스의 약관 준비도를 검토하나요?"'),
        ],
        "outputs": [
            ("Policy readiness review (ToS/Privacy/AUP/Refund/DPA)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    # §7 Release
    "setup-quality-gates": {
        "inputs": [
            ("프로젝트 환경", "✅", "artifact", "현재 코드베이스", "(자동 감지 — 프로젝트 구조에서 추출)"),
        ],
        "outputs": [
            ("Quality gate 구성 (pre-commit/pre-push hooks)", "artifact", "설정 파일", "`build-with-tdd`"),
        ],
    },
    "auto-create-pr": {
        "inputs": [
            ("현재 branch + commits", "✅", "artifact", "git 상태", "(자동 감지)"),
        ],
        "outputs": [
            ("Pull Request", "artifact", "PR URL", "`review-engineering`"),
        ],
    },
    "automate-release-tagging": {
        "inputs": [
            ("Merged PR set", "✅", "artifact", "git history", "(자동 감지 — merged PR에서 추출)"),
        ],
        "outputs": [
            ("Git tag + release notes (semver auto-decision)", "artifact", "git tag", "`sync-release-docs`"),
        ],
    },
    "sync-release-docs": {
        "inputs": [
            ("코드 변경 diff", "✅", "artifact", "git diff", "(자동 감지)"),
        ],
        "outputs": [
            ("Updated docs (affected 문서 auto-update)", "artifact", "docs/ 파일 갱신", "`prepare-launch-checklist`"),
        ],
    },
    "write-changelog": {
        "inputs": [
            ("버전 정보 + 변경 내역", "✅", "knowledge / artifact", "git log 또는 사용자 설명", '"어떤 버전의 changelog를 작성하나요?"'),
        ],
        "outputs": [
            ("CHANGELOG entry (release-summary format)", "artifact", "CHANGELOG.md 갱신", "`sync-release-docs`"),
        ],
    },
    "guard-destructive-commands": {
        "inputs": [
            ("(ambient — 자동 감지)", "✅", "artifact", "실행 예정 bash command", "(자동 감지)"),
        ],
        "outputs": [
            ("Risk warning + safe exception 판단", "decision", "inline warning", "(ambient)"),
        ],
    },
    "compose-safety-mode": {
        "inputs": [
            ("(ambient — 자동 감지)", "✅", "artifact", "현재 세션 상태", "(자동 감지)"),
        ],
        "outputs": [
            ("Combined safety hooks (max safety mode)", "artifact", "session state", "(ambient)"),
        ],
    },
    "run-uat": {
        "inputs": [
            ("UAT 시나리오", "✅", "artifact / knowledge", "Phase 4 `define-acceptance-test-plan` 또는 사용자 정의", '"UAT 시나리오를 알려주세요."'),
            ("이해관계자 정보", "✅", "knowledge", "사용자 도메인 지식", '"sign-off할 이해관계자는 누구인가요?"'),
        ],
        "outputs": [
            ("UAT sign-off (go/no-go + evidence)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
    "run-beta-program": {
        "inputs": [
            ("Beta 코호트 정의", "✅", "knowledge", "사용자 도메인 지식", '"beta 참여자는 몇 명이고 어떻게 선정하나요?"'),
        ],
        "outputs": [
            ("Beta feedback report (structured 피드백 + GA gating 판단)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "setup-canary-deploy": {
        "inputs": [
            ("배포 환경 정보", "✅", "knowledge", "사용자 도메인 지식", '"canary 배포 환경과 단계별 비율을 알려주세요."'),
        ],
        "outputs": [
            ("Canary deploy 구성 (단계 비율 + metric gate + auto-promote/rollback)", "artifact", "설정 파일", "`ship-release`"),
        ],
    },
    "setup-feature-flags": {
        "inputs": [
            ("Feature 목록", "✅", "knowledge", "사용자 도메인 지식", '"feature flag로 관리할 기능 목록을 알려주세요."'),
        ],
        "outputs": [
            ("Feature flag 구성 (kill switch + targeting + lifecycle)", "artifact", "설정 파일", "`ship-release`"),
        ],
    },
    "setup-rollback-runbook": {
        "inputs": [
            ("배포 환경 맥락", "✅", "knowledge", "사용자 도메인 지식", '"rollback 대상 환경과 서비스를 알려주세요."'),
        ],
        "outputs": [
            ("Rollback runbook (decision tree + 실행 절차 + verification)", "artifact", "structured document", "`handle-incident`"),
        ],
    },
    "prepare-launch-checklist": {
        "inputs": [
            ("QA report", "✅", "artifact", "`verify-quality` 산출물", "먼저 `/buddy:verify-quality` 를 실행하세요"),
            ("Security audit", "✅", "artifact", "`audit-security` 산출물", "먼저 `/buddy:audit-security` 를 실행하세요"),
            ("Release docs", "선택", "artifact", "`sync-release-docs` 산출물", "없으면 문서 항목 skip"),
        ],
        "outputs": [
            ("Launch checklist pass (17+ 항목 cross-functional gate)", "artifact", "structured checklist", "`ship-release`"),
        ],
    },
    "setup-incident-paging": {
        "inputs": [
            ("팀 구조 + 인프라 정보", "✅", "knowledge", "사용자 도메인 지식", '"on-call 팀 구성과 alert 대상 서비스를 알려주세요."'),
        ],
        "outputs": [
            ("Incident paging 구성 (rotation + escalation + alert + runbook index)", "artifact", "structured document", "`handle-incident`"),
        ],
    },
    # §8 Operations
    "handle-incident": {
        "inputs": [
            ("인시던트 보고 (증상, 영향 범위, 시작 시점)", "✅", "knowledge", "사용자 발화 또는 alert", '"어떤 인시던트인가요? 증상과 영향 범위를 알려주세요."'),
        ],
        "outputs": [
            ("Incident response (severity 분류 → 완화 → fix → 통신)", "artifact", "structured timeline + fix", "`conduct-postmortem`"),
        ],
    },
    "conduct-postmortem": {
        "inputs": [
            ("인시던트 데이터 (timeline, root cause, 영향)", "✅", "artifact / knowledge", "`handle-incident` 산출물 또는 사용자 설명", '"포스트모템 대상 인시던트를 알려주세요."'),
        ],
        "outputs": [
            ("Postmortem (blameless, 5 Whys, action items)", "artifact", "structured document", "`generate-improvement-tasks`"),
        ],
    },
    "monitor-regressions": {
        "inputs": [
            ("Baseline metrics", "✅", "artifact", "이전 측정 데이터 또는 최초 측정", "(자동 감지 — baseline에서 delta 비교)"),
        ],
        "outputs": [
            ("Regression alerts (delta-based threshold 초과 항목)", "artifact", "structured alerts", "`handle-incident`"),
        ],
    },
    "analyze-actor-failure-rate": {
        "inputs": [
            ("운영 데이터 (actor별 실패율)", "✅", "artifact", "모니터링 시스템 데이터", '"분석할 actor와 기간을 알려주세요."'),
        ],
        "outputs": [
            ("Failure analysis (trust score + 6 recovery 패턴)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "analyze-cost-anomaly": {
        "inputs": [
            ("비용 데이터 (cloud/SaaS)", "✅", "artifact", "billing dashboard 또는 API", '"비용 이상 탐지 대상 서비스와 기간을 알려주세요."'),
        ],
        "outputs": [
            ("Cost anomaly report (spike detection + root cause + recovery)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "audit-error-budget": {
        "inputs": [
            ("SLO 정의", "✅", "artifact / knowledge", "`design-observability` 산출물 또는 사용자 정의", '"SLO 지표와 목표를 알려주세요."'),
            ("운영 데이터", "✅", "artifact", "모니터링 시스템", "(자동 수집)"),
        ],
        "outputs": [
            ("Error budget burn rate report (multi-window + release gate)", "artifact", "structured report", "`ship-release` (gate)"),
        ],
    },
    "summarize-retro": {
        "inputs": [
            ("Git history (특정 기간)", "✅", "artifact", "git log", "(자동 감지 — 최근 1주 기본)"),
        ],
        "outputs": [
            ("Evidence-based retrospective (work types, hotspots, focus score)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "generate-improvement-tasks": {
        "inputs": [
            ("분석 결과 (A/B, funnel, postmortem, feedback 등)", "✅", "artifact / knowledge", "§8 분석 스킬 산출물 또는 사용자 설명", '"어떤 분석 결과를 task로 변환하나요?"'),
        ],
        "outputs": [
            ("Improvement task list (RICE 기반, §2 재진입 입력)", "artifact", "structured YAML", "`define-features` (§2 재진입)"),
        ],
    },
    "design-ab-experiment": {
        "inputs": [
            ("가설", "✅", "knowledge", "사용자 도메인 지식", '"테스트할 가설을 알려주세요. (변수, 측정 지표, 기대 변화)"'),
        ],
        "outputs": [
            ("실험 설계 (표본 크기 + 대조군 + 측정 지표 + 기간)", "artifact", "structured YAML", "`analyze-ab-experiment`"),
        ],
    },
    "analyze-ab-experiment": {
        "inputs": [
            ("실험 데이터 (control/treatment 결과)", "✅", "artifact", "실험 플랫폼 데이터", '"실험 결과 데이터를 제공해 주세요."'),
        ],
        "outputs": [
            ("실험 분석 결과 (Ship/Revert/Continue 결정)", "decision", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "analyze-user-funnel": {
        "inputs": [
            ("Funnel 정의 + 분석 데이터", "✅", "knowledge / artifact", "사용자 정의 또는 analytics 데이터", '"분석할 funnel 단계와 데이터를 알려주세요."'),
        ],
        "outputs": [
            ("Funnel drop-off 분석 (단계별 전환율 + 이탈 원인)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "analyze-feature-adoption": {
        "inputs": [
            ("Usage 데이터", "✅", "artifact", "analytics 플랫폼 데이터", '"분석할 feature와 usage 데이터를 알려주세요."'),
        ],
        "outputs": [
            ("Adoption report (awareness→trial→habit funnel + abandonment 가설)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "analyze-user-cohort": {
        "inputs": [
            ("Cohort 데이터", "✅", "artifact", "analytics 플랫폼 데이터", '"분석할 cohort 기간과 데이터를 알려주세요."'),
        ],
        "outputs": [
            ("Cohort analysis (D1/D7/D30/D90 retention + LTV/CAC)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "triage-customer-support-ticket": {
        "inputs": [
            ("티켓 내용", "✅", "knowledge", "사용자 또는 CS 시스템", '"분류할 티켓 내용을 알려주세요."'),
        ],
        "outputs": [
            ("Triaged ticket (분류 + severity + routing)", "artifact", "structured record", "`diagnose-bug` 또는 `generate-improvement-tasks`"),
        ],
    },
    "analyze-customer-feedback-corpus": {
        "inputs": [
            ("Feedback corpus (CS/NPS/review/interview 텍스트)", "✅", "artifact", "CS 시스템 또는 사용자 제공", '"분석할 피드백 데이터를 제공해 주세요."'),
        ],
        "outputs": [
            ("Topic analysis (토픽 모델링 + sentiment + verbatim quotes)", "artifact", "structured report", "`generate-improvement-tasks`"),
        ],
    },
    "optimize-conversion-funnel": {
        "inputs": [
            ("Funnel 데이터", "✅", "artifact", "analytics 플랫폼", '"최적화할 funnel과 데이터를 알려주세요."'),
        ],
        "outputs": [
            ("CRO 추천 (biggest-drop bottleneck + A/B pipeline)", "artifact", "structured report", "`design-ab-experiment`"),
        ],
    },
    "plan-growth-experiment": {
        "inputs": [
            ("Growth 맥락 (현재 지표 + 목표)", "✅", "knowledge", "사용자 도메인 지식", '"현재 성장 지표와 목표를 알려주세요."'),
        ],
        "outputs": [
            ("Growth experiment sprint plan (ICE/RICE 우선순위)", "artifact", "structured YAML", "`design-ab-experiment`"),
        ],
    },
    "draft-marketing-copy": {
        "inputs": [
            ("제품/타겟 맥락", "✅", "knowledge", "사용자 도메인 지식", '"어떤 제품의 마케팅 카피를 작성하나요? target audience는?"'),
        ],
        "outputs": [
            ("Marketing copy variants (headline + body + CTA)", "artifact", "structured text", "(마케팅 사용)"),
        ],
    },
    "plan-marketing-channel": {
        "inputs": [
            ("제품/시장 맥락", "✅", "knowledge", "사용자 도메인 지식", '"어떤 제품의 마케팅 채널을 계획하나요?"'),
        ],
        "outputs": [
            ("Channel strategy (6 channel fit + LTV/CAC + mix)", "artifact", "structured YAML", "(마케팅 사용)"),
        ],
    },
    "audit-seo-aso": {
        "inputs": [
            ("사이트/앱 URL", "✅", "knowledge", "사용자 지정", '"SEO/ASO 감사할 URL을 알려주세요."'),
        ],
        "outputs": [
            ("SEO/ASO audit report (5+6 영역 + keyword + content gap)", "artifact", "structured report", "(마케팅 사용)"),
        ],
    },
    "automate-marketing-content": {
        "inputs": [
            ("콘텐츠 전략", "✅", "knowledge", "사용자 도메인 지식", '"자동화할 마케팅 콘텐츠 유형과 채널을 알려주세요."'),
        ],
        "outputs": [
            ("Automated content pipeline (email sequence + calendar + metrics)", "artifact", "structured YAML", "(마케팅 사용)"),
        ],
    },
    # §9 Lifecycle Management
    "deprecate-feature": {
        "inputs": [
            ("Feature + usage 데이터", "✅", "artifact / knowledge", "Phase 8 산출물 또는 사용자 설명", '"폐기할 feature와 현재 사용량을 알려주세요."'),
            ("Sunset 결정", "✅", "decision", "사용자 의사결정", '"폐기를 확정하나요?"'),
        ],
        "outputs": [
            ("Deprecation plan (timeline + notice 5 layer + migration path)", "artifact", "structured document", "`migrate-customers`"),
        ],
    },
    "migrate-customers": {
        "inputs": [
            ("Migration 범위", "✅", "artifact / knowledge", "`deprecate-feature` 산출물 또는 사용자 설명", '"마이그레이션 대상과 목적지를 알려주세요."'),
        ],
        "outputs": [
            ("Migration report (tier별 진행률 + rollback + communication)", "artifact", "structured report", "`archive-product`"),
        ],
    },
    "archive-product": {
        "inputs": [
            ("EOL 결정", "✅", "decision", "사용자 의사결정", '"제품 EOL을 확정하나요?"'),
            ("제품 정보", "✅", "knowledge", "사용자 도메인 지식", '"아카이브할 제품명과 범위를 알려주세요."'),
        ],
        "outputs": [
            ("EOL documentation (data export + tombstone + legal + knowledge)", "artifact", "structured document", "(아카이브)"),
        ],
    },
    "spin-off-feature": {
        "inputs": [
            ("분리 대상 feature + 근거", "✅", "knowledge", "사용자 도메인 지식", '"어떤 기능을 분리하나요? 분리 근거는?"'),
        ],
        "outputs": [
            ("Separated product/repo (코드 분리 + brand/운영 분리 계획)", "artifact", "structured plan + 코드", "(새 lifecycle)"),
        ],
    },
    # Cross-cutting
    "decompose-blocker": {
        "inputs": [
            ("Stuck 상태 설명", "✅", "knowledge", "사용자 발화 또는 자동 trigger (3회 시도 후 미해결)", '"무엇에 막혀있나요? 시도한 것과 결과를 알려주세요."'),
        ],
        "outputs": [
            ("Decomposed problem + 행동 후보 1-3개", "artifact", "structured analysis", "`diagnose-bug`, `build-with-tdd`"),
        ],
    },
    "status": {
        "inputs": [
            ("(없음 — 자동 감지)", "선택", "artifact", "현재 디렉토리의 artifact 존재 여부", "(자동 감지)"),
        ],
        "outputs": [
            ("현재 phase + 다음 권장 command", "artifact", "formatted output", "(사용자 안내)"),
        ],
    },
    "write-a-skill": {
        "inputs": [
            ("스킬 개념 설명", "✅", "knowledge", "사용자 발화", '"어떤 스킬을 만드나요? 스킬의 목적을 설명해 주세요."'),
        ],
        "outputs": [
            ("PROCEDURE.md + skill-catalog 등재", "artifact", "markdown files", "(메타)"),
        ],
    },
    "apply-builder-ethos": {
        "inputs": [
            ("(ambient — 자동 적용)", "선택", "knowledge", "프로젝트 맥락", "(자동 적용)"),
        ],
        "outputs": [
            ("3 원칙 주입 (Boil the Lake / Search Before Building / User Sovereignty)", "decision", "ambient", "(ambient)"),
        ],
    },
    "benchmark-llm-models": {
        "inputs": [
            ("비교 대상 모델 + 평가 기준", "✅", "knowledge", "사용자 지정", '"어떤 모델을 비교하나요? 평가 기준은?"'),
        ],
        "outputs": [
            ("Benchmark results (모델별 성능 비교)", "artifact", "structured report", "(의사결정 지원)"),
        ],
    },
    "detect-install-type": {
        "inputs": [
            ("(자동 감지)", "선택", "artifact", "파일 시스템 탐색", "(자동 감지)"),
        ],
        "outputs": [
            ("Install type 분류 (global-git/local-git/vendored/package-manager/dev-symlink)", "decision", "inline", "`guide-setup-wizard`"),
        ],
    },
    "guide-setup-wizard": {
        "inputs": [
            ("설정 대상", "✅", "knowledge", "사용자 지정", '"어떤 credential/config를 설정하나요?"'),
        ],
        "outputs": [
            ("구성된 환경 (auto-detect → picker → verify)", "artifact", "설정 파일", "(설정 완료)"),
        ],
    },
    "save-context": {
        "inputs": [
            ("현재 세션 상태", "✅", "artifact", "현재 대화 맥락 + git 상태", "(자동 수집)"),
        ],
        "outputs": [
            ("Checkpoint file (decisions + remaining work + git status)", "artifact", "markdown file", "`restore-context`"),
        ],
    },
    "restore-context": {
        "inputs": [
            ("Checkpoint file", "✅", "artifact", "`save-context` 산출물", "(자동 탐색 — 최신 checkpoint 로드)"),
        ],
        "outputs": [
            ("Restored context summary (작업 항목 + 다음 단계)", "artifact", "formatted output", "(사용자 안내)"),
        ],
    },
    "persist-learning-jsonl": {
        "inputs": [
            ("Learning entry", "✅", "knowledge", "스킬 실행 중 발견된 패턴/함정", "(스킬이 자동 호출)"),
        ],
        "outputs": [
            ("JSONL record (append-only)", "artifact", "JSONL line", "(cross-session 참조)"),
        ],
    },
    "review-legal-regulatory": {
        "inputs": [
            ("제품 맥락 (problem statement + 데이터 흐름)", "✅", "knowledge", "PRD 또는 사용자 설명", '"어떤 제품의 법률/규제 검토를 하나요?"'),
            ("Target market 결정", "선택", "decision", "`decide-target-market` 산출물", "없으면 region-agnostic 검토만"),
        ],
        "outputs": [
            ("Legal/regulatory review (7 sub-domain checklist + evidence package)", "artifact", "structured report", "`prepare-launch-checklist`"),
        ],
    },
}


def find_insertion_point(content: str) -> int:
    """Find the line index to insert I/O Contract sections."""
    lines = content.split('\n')
    for i, line in enumerate(lines):
        # Insert before first --- separator (after intro)
        if i > 2 and line.strip() == '---':
            return i
        # Or before ## 0. STOP / ## 1. / ## 2. sections
        if i > 2 and (line.startswith('## 0.') or line.startswith('## 1.') or line.startswith('## 2.')):
            return i
    # Fallback: insert after line 5
    return min(5, len(lines))


def format_io_section(skill_name: str, contract: dict) -> str:
    """Format the I/O Contract markdown section."""
    lines = []
    lines.append("")
    lines.append("## Input Requirements")
    lines.append("")
    lines.append("| Input | Required | Type | Source | 미제공 시 |")
    lines.append("|-------|----------|------|--------|----------|")
    for inp in contract["inputs"]:
        name, req, typ, source, fallback = inp
        lines.append(f"| {name} | {req} | {typ} | {source} | {fallback} |")
    lines.append("")
    lines.append("## Output Contract")
    lines.append("")
    lines.append("| Output | Type | Format | Consumers |")
    lines.append("|--------|------|--------|-----------|")
    for out in contract["outputs"]:
        name, typ, fmt, consumers = out
        lines.append(f"| {name} | {typ} | {fmt} | {consumers} |")
    lines.append("")
    return '\n'.join(lines)


def main():
    updated = 0
    skipped = 0
    missing = 0

    for skill_name, contract in CONTRACTS.items():
        proc_path = SKILLS_DIR / skill_name / "PROCEDURE.md"
        if not proc_path.exists():
            print(f"MISSING: {skill_name}")
            missing += 1
            continue

        content = proc_path.read_text()

        if "## Input Requirements" in content:
            print(f"SKIP (already has I/O): {skill_name}")
            skipped += 1
            continue

        insert_idx = find_insertion_point(content)
        lines = content.split('\n')

        io_section = format_io_section(skill_name, contract)
        lines.insert(insert_idx, io_section)

        proc_path.write_text('\n'.join(lines))
        updated += 1
        print(f"UPDATED: {skill_name}")

    print(f"\nDone: {updated} updated, {skipped} skipped, {missing} missing")


if __name__ == "__main__":
    main()
