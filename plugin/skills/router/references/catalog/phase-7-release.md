# §7 Release & Deployment (출시·배포) — Stage Skills

- **Orchestrator (entry)**: `ship-release`
- **DoR → DoD**: V&V passed code → release package + deployed artifact + launch checklist pass
- 정체성 원본: `engineering-phases.md` §2 Phase 7 (UAT 위치 + §6 경계) | 명사 원본: `se-lifecycle-naming.md` §1
- 경계: §6은 "코드 자체"의 품질, §7은 "사용자가 만났을 때"의 품질(UAT·beta·canary).

## PR & Quality Gates
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `setup-quality-gates` | command + dispatch | husky + lint-staged + typecheck + test + secret scan을 pre-commit/pre-push에 설치 | 일회성 품질 평가(§6) → `measure-code-health` |
| `auto-create-pr` | command + dispatch | commit → branch push → PR 생성 자동화 | 5-stage sub-orchestration 필요 → `finish-development-branch` (catalog 인덱스) |

## Release Tagging & Docs
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `automate-release-tagging` | dispatch | merged PR set → semver auto-decision + git tag + release note | 문서 동기화 → `sync-release-docs` |
| `sync-release-docs` | dispatch | code diff 기준 affected docs audit + auto-update/ask | 코드 변경 시점 문서(§5) → `update-docs-with-code` |
| `write-changelog` | dispatch [패턴 라이브러리] | version bump + CHANGELOG release-summary + voice rules | 전체 release 자동화 → `automate-release-tagging` |

## Safety Hooks (ambient)
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `guard-destructive-commands` | dispatch [패턴 라이브러리] | rm -rf/DROP TABLE/force push 등 destructive 명령 전 risk 차단 | 단일 디렉토리 edit lock → `freeze-edit-scope`(§5) |
| `compose-safety-mode` | dispatch [패턴 라이브러리/META] | guard + freeze 등 복수 safety hook을 max safety로 합성 | 단일 hook이면 개별 스킬 직접 |

## Launch Readiness & Rollout
| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `run-uat` | dispatch (via `ship-release` 또는 `/buddy:run run-uat`) | designated stakeholder가 critical flow verify + go/no-go sign-off | 폭넓은 early adopter cohort → `run-beta-program` |
| `run-beta-program` | dispatch (via `ship-release` 또는 `/buddy:run run-beta-program`) | closed beta cohort(5-20) 운영 + 피드백 + GA gating | 단일 sign-off → `run-uat` |
| `setup-canary-deploy` | command + dispatch | canary 단계 비율 + metric gate + auto-promote/rollback | 배포 *전략 설계*(§3) → `design-deploy-strategy` |
| `setup-feature-flags` | command + dispatch | feature flag system + kill switch + targeting + lifecycle | 단계적 트래픽 비율 → `setup-canary-deploy` |
| `setup-rollback-runbook` | command + dispatch | rollback decision tree(언제 rollback/forward fix) + 실행 절차 | 인시던트 대응 런북(§8) → `handle-incident` |
| `prepare-launch-checklist` | command + dispatch | 17+ 항목 cross-functional GA gate (eng/security/ops/product/legal/cost) | 약관/정책 준비만(§6) → `review-terms-policy-readiness` |
| `setup-incident-paging` | command + dispatch | on-call rotation + escalation + alert wiring + runbook 인덱스 | 실제 인시던트 발생 시(§8) → `handle-incident` |

## Disambiguation (노드 내)
- `run-uat`(소수 stakeholder go/no-go) vs `run-beta-program`(다수 cohort 피드백).
- `setup-canary-deploy`(트래픽 단계 비율) vs `setup-feature-flags`(기능 on/off targeting).
- 설계(§3 `design-deploy-strategy`) vs 셋업(§7 `setup-canary-deploy`): 전략 vs 실제 구성.

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. UAT 실패 → §5/§6 backtrack. git 자동화 안전 → `git-safety-rules.md`.
