# §3 Software Design (기술 설계) — Stage Skills

- **Orchestrator (entry)**: `design-system`
- **DoR → DoD**: requirements spec (SRS) → SDD = tech stack ADR + API contract + data model + infra
- 정체성 원본: `engineering-phases.md` §2 Phase 3 (카테고리 3-A~3-I 정의) | 명사 원본: `se-lifecycle-naming.md` §1
- 스킬이 많아 engineering-phases.md의 카테고리로 sub-group. 각 design-* 스킬은 첫 답 commit 직전 `verify-best-alternative` 의무 호출(편향 방지).

## 3-A. Architecture Foundation
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `define-tech-stack` | command + dispatch | language/framework/DB/runtime/hosting 8+ 차원 alternatives 비교 + 5년 lock-in 평가 | 앱 vs 웹 폼팩터 선결 → `decide-form-factor-app-vs-web` 먼저 |
| `decide-form-factor-app-vs-web` | command + dispatch | app/web/hybrid/desktop 결정 (7차원 + ADR). §3 stage 0 | 스택 세부 → `define-tech-stack` |
| `derive-system-topology` | command + dispatch | actor 그래프 + infra → service/data flow/trust boundary diagram | use case→infra 매트릭스 → `map-use-cases-to-infra` |
| `map-use-cases-to-infra` | command + dispatch | actor × use case → infra component 양방향 매트릭스 (§2→§3 bridge) | 토폴로지 도식 → `derive-system-topology` |
| `write-adr` | command + dispatch | Architecture Decision Record 작성 (7 섹션 + supersede 체인) | 결정 자체 미수립 → 해당 design-* 스킬 먼저 |

## 3-B. Data & Contract Design
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-data-model` | command + dispatch | entity 매핑 + read/write 패턴 + normalization + index + zero-downtime migration | API 인터페이스 → `design-api-contract` |
| `design-api-contract` | command + dispatch | REST/GraphQL/RPC + operation 매핑 + schema + error taxonomy + versioning (sync) | async 이벤트 → `design-event-schema` |
| `design-event-schema` | command + dispatch | async event schema-first — producer/consumer contract + DLQ + idempotency | 동기 요청/응답 → `design-api-contract` |
| `design-embedding-search` | command + dispatch | BM25 + vector + metadata filter + reranking hybrid search | 일반 관계형 조회 → `design-data-model` |

## 3-C. Security & Tenancy
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-auth-model` | command + dispatch | OAuth2/JWT/SAML/SSO/RBAC 5축 통합 설계 | tenant 격리 → `design-tenant-model` / 시크릿 → `design-secret-management` |
| `design-tenant-model` | command + dispatch | multi-tenant 격리 (shared-RLS vs schema-per vs DB-per) + 3-layer defense | 인증/인가 → `design-auth-model` |
| `design-secret-management` | command + dispatch | secret store + rotation + access audit + leak detection | 인증 흐름 → `design-auth-model` |

## 3-D. Operations Strategy
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-observability` | command + dispatch | logs/metrics/traces 3 pillar + retention + SLO/SLI + alert | 배포 방식 → `design-deploy-strategy` / 운영 중 SLO 감사(§6) → `audit-error-budget` |
| `design-deploy-strategy` | command + dispatch | canary/blue-green/rolling/recreate 선택 + env 분리 + IaC | 실제 canary 셋업(§7) → `setup-canary-deploy` |
| `design-artifact-storage` | command + dispatch | patch/git_bundle/template/package immutable artifact 저장·검증·배포 | release 배포(§7) → `automate-release-tagging` |

## 3-E. Quality Baseline
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-i18n-strategy` | command + dispatch | locale + fallback + ICU MessageFormat + RTL + 번역 워크플로우 | 번역 커버리지 감사(§6) → `audit-i18n-coverage` |
| `design-accessibility-baseline` | command + dispatch | WCAG 2.2 AA + a11y annotation + component baseline + axe-core CI | a11y 감사(§6) → `audit-accessibility` |

## 3-F. UX/UI Design
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `apply-design-system` | command + dispatch | form factor별 design system 채택 (shadcn/MUI/HIG 등) + token 5종 | 신규 design system 생성 → `consult-design-system` |
| `consult-design-system` | dispatch | research→synthesize→output로 complete design system 생성 | 기존 system 채택만 → `apply-design-system` |
| `design-interaction-pattern` | command + dispatch | gesture/motion/feedback/state transition + reduced-motion | 정적 시각 품질 감사 → `audit-ui-quality` |
| `prototype-from-spec` | command + dispatch | low-fi → state diagram → high-fi → user testing → dev handoff | 구현 단계(§5) → `build-with-tdd` |

## 3-G. Domain-Specific Design
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-billing-system` | command + dispatch | SaaS 결제 (Stripe/Toss + subscription + metering + dunning) | — (도메인 특화 진입점) |
| `design-mcp-server` | command + dispatch | MCP(Model Context Protocol) server 설계 | Claude hook 설계 → `design-claude-hooks` |
| `design-claude-hooks` | command + dispatch | Claude Code PreToolUse/PostToolUse/Stop/SessionStart hook 설계 | MCP server → `design-mcp-server` |

## 3-H. Design Reviews
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `review-architecture` | dispatch | 구조 무결성·결합도·deep module·interface depth 상위 검토 | implementation plan 리뷰 → `review-engineering` |
| `review-engineering` | dispatch | Engineering manager 페르소나로 plan의 아키텍처·data flow·edge·coverage·perf 리뷰 | scope 형성 단계 → `review-scope` |
| `review-scope` | dispatch | Creator 페르소나로 plan scope 형성·결정 early-stage 리뷰 | 구현 가능성 심층 → `review-engineering` |
| `review-design` | dispatch | Designer-mode plan을 dimension별 0-10 score + reverse-path | 코드/구조 → `review-architecture` |
| `review-devex` | dispatch | developer-facing DX plan 리뷰 (EXPANSION/POLISH/TRIAGE) | live 제품 실측 DX(§6) → `audit-live-devex` |
| `audit-ui-quality` | command + dispatch | visual/interaction polish/micro-detail/a11y 시각 audit + severity | 인터랙션 *설계* → `design-interaction-pattern` |

## 3-I. Decision Support
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `consult-codex` | command + dispatch | 외부 LLM CLI로 review/challenge/consult second opinion | 내부 다관점 대안 비교 → `verify-best-alternative` |
| `verify-best-alternative` | command + dispatch | [편향 방지] 엔지니어링 결정 commit 직전 orthogonal 대안 발산 + rubric 비교 (design-* sub-step 의무) | 전략·사업 plan critique → `critique-plan` |
| `critique-plan` | dispatch | Implementation plan strategic critique (CEO/founder 페르소나) | 엔지니어링 결정 검토 → `verify-best-alternative` |

## Disambiguation (노드 내)
- `design-api-contract`(sync 요청/응답) vs `design-event-schema`(async 이벤트): 통신 패러다임.
- `apply-design-system`(채택) vs `consult-design-system`(생성): 기존 유무.
- `review-architecture`(구조) vs `review-engineering`(실행계획) vs `review-scope`(범위) vs `review-design`(디자인 차원): 리뷰 대상 축.
- 설계(design-*) vs 감사(audit-*): §3는 "어떻게 만들지 결정", §6는 "만든 것이 기준 충족하나".

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3 케이스 C. DoR(SRS/feature backlog) 부재 → §2 `define-features` 선행.
