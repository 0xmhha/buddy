# ship-release — 7단계 Release & Beta Orchestrator

7단계 라이프사이클 단계의 진입점. Quality gate pass → Pre-launch safety nets + UAT/Beta + Tagged GA release.

**진입 조건**: 6단계 quality gate 통과 (QA report + security sign-off 완료).
**산출물**: Pre-launch safety nets (canary plan / feature flags / rollback runbook / incident paging / launch checklist), UAT + beta decision, Tagged release + release notes, GA deployment.
**다음 phase**: Production traffic 발생 → `iterate-product` (8단계).

---

## Stage 흐름 (Phase 3 v1.0.4 기준 14 stage)

```
ship-release (7단계 phase orchestrator)
├── 7-1 단계 Release Preparation (4)
│   ├── stage 1: setup-quality-gates       [Done] husky / lint-staged / typecheck / test / secret scan
│   ├── stage 2: write-changelog           [Done] semver + CHANGELOG release-summary
│   ├── stage 3: sync-release-docs         [Done] doc drift audit + auto-update
│   └── stage 4: auto-create-pr            [Done] commit → branch push → PR 생성
├── 7-2 단계 Pre-Launch Safety Nets (5 — Phase 3)
│   ├── stage 5: setup-feature-flags       [Done] flag system + kill switch + targeting + cleanup
│   ├── stage 6: setup-canary-deploy       [Done] canary 단계 + metric gate + auto-promote/rollback
│   ├── stage 7: setup-rollback-runbook    [Done] decision tree + step-by-step + verification
│   ├── stage 8: setup-incident-paging     [Done] on-call rotation + escalation + alert wiring + runbook index
│   └── stage 9: prepare-launch-checklist  [Done] 17+ 항목 6-axis cross-functional readiness gate
├── 7-3 단계 Beta / UAT (2)
│   ├── stage 10: run-uat                  [Done] designated stakeholder UAT + go/no-go
│   └── stage 11: run-beta-program         [Done] 클로즈드 5-20 cohort + structured 피드백 + GA gating
└── 7-4 단계 GA Release (3)
    ├── stage 12: automate-release-tagging [Done] semver + git tag + release note
    ├── stage 13: guard-destructive-commands [Done] 배포 전 위험 명령 가드
    └── stage 14: compose-safety-mode      [Done] max safety mode 합성 (guard + freeze-edit-scope)
```

> Phase 3 (v1.0.4) 에서 stage 5-9 + 10-11 모두 구현 완료. 이전 v1.0.3 의 9 stage 에서 14 stage 로 확장. ordering rationale: safety net 인프라 (5-9) 가 UAT/beta (10-11) 전에 ready 되어야 — UAT 시점에 rollback / paging 가 작동 검증되어야 GA 안전.

## 권장 호출 패턴 (Phase 3 신규 chain)

§7 의 7-1 + 7-2 + 7-3 + 7-4 일괄 cover (큰 release 의 경우):

```bash
# 큰 release — 모든 layer
/buddy:chain setup-quality-gates,write-changelog,sync-release-docs,auto-create-pr,setup-feature-flags,setup-canary-deploy,setup-rollback-runbook,setup-incident-paging,prepare-launch-checklist,run-uat,run-beta-program,automate-release-tagging -- "<feature name> v<version>"
```

또는 7-2 단계 (Pre-Launch Safety Nets) 만 일괄:

```bash
# 처음 production 진입 — safety net 5 skill 구축
/buddy:chain setup-feature-flags,setup-canary-deploy,setup-rollback-runbook,setup-incident-paging,prepare-launch-checklist -- "<project name>"
```

또는 release 마다 (safety net 이미 구축된 후):

```bash
# routine release — UAT/beta + tagging
/buddy:chain run-uat,run-beta-program,prepare-launch-checklist,automate-release-tagging -- "<feature name> v<version>"
```

각 step 의 산출물이 다음 step 의 입력으로 cascade:
- **Safety net 인프라 (5-9)** → UAT 의 rollback / paging 검증 trigger
- **UAT decision (10)** → run-beta-program 의 입력 (UAT 통과한 feature 만 beta)
- **Beta decision (11)** → prepare-launch-checklist 의 row 입력
- **Launch checklist go (9)** → automate-release-tagging trigger
- **Release tag (12)** → guard-destructive-commands 활성 + compose-safety-mode max

---

## 실행 절차

### 7-1 단계 Release Preparation

> *권장 옵션*: 본 §7-1 단계 4 stage (Quality Gate / Changelog / Docs Sync / PR 생성) 는 [`finish-development-branch`](../finish-development-branch/PROCEDURE.md) sub-orchestrator 로 일괄 호출 가능 — Pre-flight remote sync (Stage 0) + mergeable 검증 (Stage 5, Iron Law) 추가됨. 안전 git 정책 ([`router/references/git-safety-rules.md`](../router/references/git-safety-rules.md)) 자동 적용. 호출 형태: `/buddy:finish-development-branch "<변경 요약>"`. 개별 stage 호출도 유지.

**Stage 1: Quality Gate 확인**

`setup-quality-gates` skill 호출 — 5 단계에서 이미 설정했으면 통과 여부만 확인:
- [ ] typecheck 통과
- [ ] lint 통과
- [ ] all tests 통과
- [ ] security scan 통과
- [ ] no secret leak

**Stage 2: Changelog**

`write-changelog` 호출 — semver 결정 + CHANGELOG release 섹션 + user-facing change summary.

**Stage 3: Doc Sync**

`sync-release-docs` 호출 — code change 대비 docs drift 감지 + auto-update.

**Stage 4: PR 생성**

`auto-create-pr` 호출 — commit → branch push → PR 생성 자동화.

### 7-2 단계 Pre-Launch Safety Nets (Phase 3 신규)

**Stage 5: Feature Flag System**

`setup-feature-flags` 호출 — flag system 결정 + 4 taxonomy + kill switch inventory + targeting + cleanup SLA + governance. 산출물: flag system + ops kill switch list + cleanup automation.

**Stage 6: Canary Deploy Plan**

`setup-canary-deploy` 호출 — stage 비율 + dwell time + metric gate + auto-promote/rollback policy + platform 별 implementation hint. 산출물: canary plan + metric gate threshold + manual override protocol.

**Stage 7: Rollback Runbook**

`setup-rollback-runbook` 호출 — decision tree + platform 별 step-by-step + schema migration safety matrix + verification + post-mortem trigger. 산출물: docs/runbooks/rollback-<project>.md.

**Stage 8: Incident Paging**

`setup-incident-paging` 호출 — on-call rotation + severity 4 분류 + escalation policy + alert routing matrix + runbook index + drill cadence + fairness audit. 산출물: paging tool config + rotation schedule + alert→runbook 매트릭스.

**Stage 9: Launch Checklist**

`prepare-launch-checklist` 호출 — 6 axis (Engineering / Security / Ops / Product / Legal / Cost) × 평균 3 row = 17+ 항목 cross-functional readiness gate. 산출물: launch checklist + decision recommendation + final approver requirement.

### 7-3 단계 Beta / UAT

**Stage 10: UAT**

`run-uat` 호출 — designated stakeholder (PM + 1-3 user) 가 critical scenario 실행 + evidence + sign-off + go/no-go. 산출물: UAT decision + critical bug triage + evidence corpus link.

**Stage 11: 베타 프로그램**

`run-beta-program` 호출 (해당 시 — small release skip 가능) — 5-20 cohort × 1-4 주 × structured feedback × GA gating decision + post-beta cleanup. 산출물: beta decision + feedback corpus summary + cohort transition plan.

### 7-4 단계 GA Release

**Stage 12: Release Tagging**

`automate-release-tagging` 호출 — merged PR set → semver auto-decision + git tag (v{MAJOR}.{MINOR}.{PATCH}) + release note 발행. prepare-launch-checklist 의 go 권고 + leadership sign-off 후에만 실행.

**Stage 13-14: 배포 안전망**

`guard-destructive-commands` + `compose-safety-mode` — 배포 전 위험 명령 가드, max safety mode (guard + freeze-edit-scope 동시 활성).

---

## Go-Live Readiness Checklist

Phase 3 부터 inline checklist 는 `prepare-launch-checklist` skill 로 redirect (17+ 항목 6 axis 의 cross-functional 통합):

```bash
/buddy:prepare-launch-checklist "<feature name> v<version>"
```

본 orchestrator inline (legacy quick check, 간이용):
- [ ] 6단계 quality gate 전체 통과
- [ ] CHANGELOG 작성 완료
- [ ] Docs sync 완료
- [ ] PR 승인됨
- [ ] Pre-launch safety net 5 skill 모두 산출물 ready
- [ ] UAT 통과 (critical issue 0 또는 conditional go)
- [ ] 베타 피드백 수집 완료 (해당하면)
- [ ] Release tag 생성됨
- [ ] 모니터링 alert 설정됨 (setup-incident-paging 산출)

---

## 다음 phase

- `/buddy:iterate-product` — 8단계 Operate & Iterate (production traffic 발생 후)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` §4 §7, §§8 Q7=(b)
- Phase 3 plan: `docs/superpowers/plans/2026-05-08-phase3-release-safety-nets-plan.md`
