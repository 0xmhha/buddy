# §8 Operation & Maintenance (운영·개선) — Stage Skills

- **Orchestrator (entry)**: `iterate-product`
- **DoR → DoD**: deployed artifact + production traffic → ops metrics + improvement backlog (→ §2 재진입) / 폐기 결정 (→ §9)
- 정체성 원본: `engineering-phases.md` §2 Phase 8 | 명사 원본: `se-lifecycle-naming.md` §1
- Lean Build-Measure-Learn 루프 포함. 자연 종료 없음 — §2 재진입 또는 §9 진입 시 cycle 종료.

## Engineering Ops
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `handle-incident` | dispatch | 프로덕션 인시던트 대응 런북 — 심각도 분류 → 완화 → 근본 원인 → fix 배포 | 종료 후 회고 → `conduct-postmortem` / paging 구조 설계(§7) → `setup-incident-paging` |
| `conduct-postmortem` | dispatch | 인시던트 종료 후 비난 없는 포스트모템 (타임라인 + 5 Whys + action) | 대응 진행 중 → `handle-incident` |
| `monitor-regressions` | dispatch [패턴 라이브러리] | delta-based threshold + transient tolerance로 regression 감지 | 코드 건강 종합(§6) → `measure-code-health` |
| `analyze-actor-failure-rate` | command + dispatch | actor failure matrix + trust score + 6 recovery 패턴 (Nygard) | 비용 spike → `analyze-cost-anomaly` |
| `analyze-cost-anomaly` | command + dispatch | cloud/SaaS spike anomaly detection + root cause 6분류 + recovery | 설계 단계 비용 추정(§6) → `audit-cost-efficiency` |
| `audit-error-budget` | command + dispatch | SLO burn rate multi-window/multi-burn (Google SRE) + release gate | observability *설계*(§3) → `design-observability` |
| `summarize-retro` | command + dispatch | git history → evidence-based weekly retrospective | 개선 task 변환 → `generate-improvement-tasks` |
| `generate-improvement-tasks` | dispatch | 실험·funnel·postmortem·피드백 분석 → RICE improvement task (§2 재진입 bridge) | 분석 자체 → 아래 analyze-* 스킬 |

## Product Analytics & Experiments
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `design-ab-experiment` | dispatch | A/B 실험 설계 — 가설/표본/대조군/지표/기간 통계 설계 | 완료 실험 분석 → `analyze-ab-experiment` |
| `analyze-ab-experiment` | dispatch | 완료 A/B 결과 — 통계+실용 유의성으로 Ship/Revert/Continue | 실험 설계 전 → `design-ab-experiment` |
| `analyze-user-funnel` | dispatch | actor별 funnel 전환/이탈 분석 (§2 use case 기반 drop-off) | CRO 최적화 실행 → `optimize-conversion-funnel` |
| `analyze-feature-adoption` | command + dispatch | awareness→trial→habit funnel + power user cohort | 획득 cohort retention → `analyze-user-cohort` |
| `analyze-user-cohort` | command + dispatch | acquisition cohort retention(D1/D7/D30/D90) + LTV/CAC + churn 분류 | feature 채택 funnel → `analyze-feature-adoption` |
| `optimize-conversion-funnel` | command + dispatch | AARRR funnel + 5 CRO sub-domain + biggest-drop bottleneck | 단순 이탈 분석 → `analyze-user-funnel` |
| `plan-growth-experiment` | command + dispatch | Hacking Growth ICE/RICE + sprint cadence + win/loss 영속화 | 단일 A/B 설계 → `design-ab-experiment` |
| `triage-customer-support-ticket` | command + dispatch | ticket 분류 + severity + recurring pattern + product feedback loop | 텍스트 corpus 토픽 모델링 → `analyze-customer-feedback-corpus` |
| `analyze-customer-feedback-corpus` | command + dispatch | CS/NPS/review/interview corpus → 토픽 모델링 + sentiment + NPS 분리 | 개별 ticket 분류 → `triage-customer-support-ticket` |

## Marketing
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `draft-marketing-copy` | command + dispatch | 5 카피 유형(landing/email/ad/social/vs page) + headline variant + framework | 채널 전략 → `plan-marketing-channel` |
| `plan-marketing-channel` | command + dispatch | 6 channel channel-fit + LTV/CAC + brand 단계 mix | 카피 작성 → `draft-marketing-copy` |
| `audit-seo-aso` | command + dispatch | SEO 5영역 + ASO 6영역 + keyword research + content gap (E-E-A-T) | 콘텐츠 자동화 → `automate-marketing-content` |
| `automate-marketing-content` | command + dispatch | email sequence + cold email cadence + content calendar + automation | 일회성 카피 → `draft-marketing-copy` |

## Disambiguation (노드 내)
- `handle-incident`(대응 중) vs `conduct-postmortem`(종료 후) vs `setup-incident-paging`(§7 사전 구조 설계).
- `analyze-user-funnel`(진단) vs `optimize-conversion-funnel`(CRO 실행).
- `design-ab-experiment`(설계) vs `analyze-ab-experiment`(결과 분석).
- 운영 중 분석(§8 analyze-*) vs 설계 단계 감사(§6 audit-cost-efficiency / §3 design-observability).

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3 케이스 E. 개선 → §2 재진입(`generate-improvement-tasks`). 폐기 결정 → §9 `manage-lifecycle`.
