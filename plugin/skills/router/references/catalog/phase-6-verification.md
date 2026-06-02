# §6 Verification & Validation (검증·확인) — Stage Skills

- **Orchestrator (entry)**: `verify-quality`
- **DoR → DoD**: working code + developer tests → V&V evidence = QA + security + a11y + compliance + code health sign-off
- 정체성 원본: `engineering-phases.md` §2 Phase 6 (Phase 5와 책임 경계) | 명사 원본: `se-lifecycle-naming.md` §1
- 경계: §5는 "내 테스트가 green"이면 종료, §6는 "상용 출시 가능한 품질 bar"를 별도 평가.

## Testing
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `classify-qa-tiers` | dispatch [패턴 라이브러리] | QA intensity를 Quick/Standard/Exhaustive 3 tier로 분류 | 특정 테스트 실행 → 아래 test-* 스킬 |
| `test-per-actor-use-case` | command + dispatch | actor의 use case 단위 통합 테스트 (frontend E2E/backend integration/3rd-party contract) | 다중 actor 연쇄 흐름 → `test-cross-actor-flow` |
| `test-cross-actor-flow` | command + dispatch | cross-actor E2E (signup→email→verify→login 등) full-stack 검증 | 단일 actor 범위 → `test-per-actor-use-case` |
| `run-load-test` | dispatch (via `verify-quality` 또는 `/buddy:run run-load-test`) | sustained/soak/spike/stress 4 시나리오 + breaking point | 장애 주입 실험 → `chaos-test` |
| `chaos-test` | command + dispatch | failure injection(network/pod/CPU/DB/time) + hypothesis-driven + game day | 정상 부하 한계 → `run-load-test` |
| `run-browser-qa` | dispatch [패턴 라이브러리] | browser 자동화 QA — snapshot diff/form/responsive/dialog | 시각 디자인 품질(§3) → `audit-ui-quality` |

## Coverage & Code Health
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `audit-test-coverage-meaningful` | command + dispatch | line + mutation + behavior + edge → trust score (단순 line% 아님) | test skeleton 생성(§5) → `generate-tests-from-spec` |
| `measure-code-health` | command + dispatch | typecheck/lint/test/deadcode/shell → 0-10 weighted health dashboard | 리뷰 risk 분류 → `classify-review-risks` |
| `classify-review-risks` | dispatch [패턴 라이브러리] | code review에서 놓치는 risk 11 category 분류 (SQL safety/LLM trust 등) | 종합 점수 대시보드 → `measure-code-health` |

## Quality Audits
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `audit-security` | command + dispatch | CSO-mode 보안 감사 (OWASP/secrets/JWT 등) | AI 기능 책임/안전 → `review-ai-safety-liability` |
| `audit-accessibility` | command + dispatch | WCAG 2.1 AA + axe + Lighthouse + screen reader 감사 | a11y *기준선 설계*(§3) → `design-accessibility-baseline` |
| `audit-i18n-coverage` | command + dispatch | locale별 번역 누락 + fallback rate + ICU 정합 + RTL | i18n *전략 설계*(§3) → `design-i18n-strategy` |
| `audit-cost-efficiency` | command + dispatch | Infracost + per-component + unit economics($/MAU) + waste | 운영 중 비용 spike(§8) → `analyze-cost-anomaly` |
| `audit-live-devex` | dispatch [패턴 라이브러리] | 빌드/배포된 live 제품을 실제 따라하며 TTHW timing DX audit | plan 단계 DX(§3) → `review-devex` |
| `audit-ubiquitous-language` | command + dispatch | 코드 식별자/PRD/도메인 어휘 3자 일관성 audit (DDD) | rename 실행(§5) → `refactor-with-rename-trace` |

## Compliance Reviews
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `review-ai-safety-liability` | dispatch | AI 기능 책임·할루시네이션·자동 의사결정·content provenance 검토 | 일반 보안 → `audit-security` |
| `review-privacy-data-risk` | dispatch | PII/민감정보 lifecycle을 GDPR/PIPA/HIPAA frame으로 검토 | 라이선스/IP → `review-license-and-ip-risk` |
| `review-license-and-ip-risk` | dispatch | 의존성/asset/AI 생성 코드 라이선스 호환성·IP 출처 검토 | 개인정보 → `review-privacy-data-risk` |
| `review-terms-policy-readiness` | dispatch | 출시 전 ToS/Privacy Policy/AUP/Refund/Cookie/DPA 준비도 | 출시 체크리스트 전반(§7) → `prepare-launch-checklist` |

## Disambiguation (노드 내)
- 설계(§3 design-*/audit-ui-quality) vs 검증(§6 audit-*): "어떻게 만들지" vs "만든 것이 기준 충족하나".
- `run-load-test`(정상 부하 한계) vs `chaos-test`(장애 주입 복원력).
- `audit-test-coverage-meaningful`(테스트 의미) vs `measure-code-health`(코드 종합 건강).

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. quality gate 실패 → §5 backtrack (fix and re-verify). 완료 발화 전 → `verification-discipline.md` 게이트.
