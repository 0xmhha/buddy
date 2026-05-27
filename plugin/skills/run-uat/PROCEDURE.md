# run-uat — UAT scenario 실행 + go/no-go 판단

§7 Release & Beta phase 의 stage. §6 quality gate (automated) 와 GA release 사이에 designated stakeholder (PM / Tenant Admin / 1~3 user) 가 acceptance test plan 의 critical flow 를 직접 verify 하는 단계. **automated CI 만으로 검증 불가능한 도메인 정합성 / UX 적정성 / business rule 정확성** 을 잡아내는 마지막 인간 gate. 산출물은 GA decision (go / no-go / conditional go) + critical bug triage + acceptance evidence corpus.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| UAT 시나리오 | ✅ | artifact / knowledge | Phase 4 `define-acceptance-test-plan` 또는 사용자 정의 | "UAT 시나리오를 알려주세요." |
| 이해관계자 정보 | ✅ | knowledge | 사용자 도메인 지식 | "sign-off할 이해관계자는 누구인가요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| UAT sign-off (go/no-go + evidence) | artifact | structured report | `prepare-launch-checklist` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 산출물에 발견되면 §5 로 회귀해 보강:

- **"happy path 만"** — error / edge / interrupted-session / multi-tab / 권한 박탈 시나리오 누락 시 production 에서 깨짐.
- **Evidence 없이 "동작 확인"** — screenshot / video / log link 없이 "OK" 만 기록 → 후속 incident 시 reproduction 불가.
- **Critical bug 분류 모호** — blocker vs major vs minor 기준 없으면 go/no-go 가 정치적 결정으로 변질.
- **UAT 참여자 single-person** — 1 명의 confirmation bias 가 통과되면 production 에서 다양성 부족 노출. 최소 2-3 명 실행자.
- **Acceptance criteria 와 disconnect** — define-acceptance-test-plan 산출물 없이 ad-hoc UAT scenario → coverage gap.
- **Sign-off 없는 자동 통과** — 명시적 stakeholder sign-off (이름 + 시간 + 의견) 없이 GA 진행 → 책임 단절.

§5 의 모든 phase 누락 없이 수행하라. skip 시 production incident 위험.

## 1. 목적

automated test 만으로 잡히지 않는 도메인 정합성·UX 적정성·business rule 정확성을 designated stakeholder 가 verify. GA decision 의 인간 final gate.

## 2. 사용 시점 (When to invoke)

- §6 quality gate (automated CI / security audit / code health) pass 직후
- production deploy 전 stakeholder sign-off 단계
- breaking change major version / 신규 critical feature 출시 시
- acceptance criteria 변경 후 재실행
- post-incident 의 "회귀 방지 acceptance" 추가 검증 시

## 3. 입력 (Inputs)

### 필수
- §2 feature spec (per-actor acceptance criteria)
- §4 `define-acceptance-test-plan` 산출물 (cross-actor flow + critical scenario)
- UAT cohort 정의 — 최소 2 명 (PM 1 + 1 designated user/admin), 권장 3-5 명
- staging 환경 access (production-like data + masked PII)

### 선택
- 이전 release 의 UAT 결과 (회귀 baseline)
- 산업 규제 의무 (의료/금융 — UAT evidence 가 audit 산출물)
- 시간 제약 (UAT window: 권장 2-5 일, 긴급 patch 시 4-8 시간)

### 입력이 부족할 때 forcing question
- "UAT 실행자가 1 명뿐인가? confirmation bias 위험 — 최소 2 명 권장."
- "evidence 형식이 결정됐나? screenshot / video / Loom link / structured form 중 어느 것?"
- "GA blocker 기준이 명시됐나? '심각한 bug 0' 만으론 부족 — severity (blocker/major/minor) 정의."
- "staging 환경이 production-like 인가? 데이터 양 / 외부 SaaS 연동 / cron 까지 동작?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — go/no-go 는 단호한 판단. "거의 통과" 거부, "blocker 0 + major 0 + sign-off ≥ N → go" 형식.
- **사용자 입력을 challenge** — "UAT 했어요" 발화에 "몇 명? 어느 시나리오? evidence 어디?" push.
- **Specificity 강제** — "잘 됨" 거부, scenario id + 실행 시간 + step-by-step result + screenshot ID.
- **Evidence-first bias** — 모든 pass/fail 은 evidence 와 함께 기록. evidence 없는 통과는 "ad-hoc 인지 미실행" 분류.

도메인 원칙:

1. **Acceptance criteria → UAT scenario 1:1 매핑** — define-acceptance-test-plan 의 cross-actor flow + per-actor critical 모두 cover. coverage gap = UAT 미완료.
2. **Critical bug severity 정량 분류** — blocker (사용자 critical path 차단) / major (workaround 가능하지만 release-blocking) / minor (cosmetic / non-critical). go/no-go 는 blocker + major 기준.
3. **Stakeholder sign-off 명시** — 각 실행자의 이름 + 시간 + 의견 (free text). sign-off 없는 row 는 "미완료".
4. **Evidence 표준화** — screenshot (full-page), video (Loom URL or local recording), structured form (Google Form / Notion DB). free-text "OK" 거부.
5. **Conditional go 허용 + condition 명시** — "post-launch hotfix 1 within 48h" 같은 condition 명시 시 go 가능, 그러나 condition 책임자 명시 + ETA + monitoring 필수.

## 5. 단계 (Phases)

### Phase 1. UAT scenario 정의

§4 define-acceptance-test-plan 의 cross-actor flow + per-actor critical 을 UAT scenario 로 변환:

| Scenario ID | Source (test plan) | Actor | Persona | Steps (high-level) | Expected result |
|-------------|-------------------|-------|---------|---------------------|------------------|
| UAT-001 | cross-actor flow #1 | User | new signup | 1) /signup form → 2) email verify → 3) login → 4) /me 조회 | 모두 success, /me 가 verified=true |
| UAT-002 | cross-actor flow #2 | User | forgot password | 1) /reset/request → 2) email link → 3) /reset/confirm → 4) login with new pw | 모두 success, 기존 sessions 모두 revoked |
| ... | ... | ... | ... | ... | ... |

scenario 수: cross-actor flow 8 + per-actor critical 5 = ~13 baseline. 추가 edge case (flaky network, multi-tab, 권한 박탈, expired token) 5+ 권장.

### Phase 2. 실행자 배정 + 일정

| Scenario ID | Owner (UAT 실행자) | Slot | Backup |
|-------------|---------------------|------|--------|
| UAT-001 | PM (홍길동) | day 1 morning | designated user |
| UAT-002 | designated user (김철수) | day 1 afternoon | PM |
| UAT-003 (admin RLS) | Tenant Admin (이영희) | day 2 morning | PM |
| ... | ... | ... | ... |

원칙:
- 실행자 ≥ 2 명 (confirmation bias 방지)
- 시나리오 ≥ 13 → window 2-5 일 권장
- 각 실행자 capacity ≤ 5 scenario / day (focus 보장)
- backup 명시 — primary 결근 시 즉시 picking up 가능

### Phase 3. Evidence 수집 protocol

각 scenario 실행 시 다음 evidence 수집 의무:

| Evidence type | Required | Tool |
|---------------|----------|------|
| Screenshot (success step) | yes | OS native, attach to scenario form |
| Screenshot (failure step) | yes (실패 시) | OS native, with timestamp + URL bar |
| Video (full flow) | conditional (cross-actor flow 만) | Loom / OBS / QuickTime |
| Browser console log | conditional (실패 시) | DevTools export |
| Network log (HAR) | conditional (실패 시 + API issue 의심) | DevTools export |
| Step timing | yes | manual record (start_at / end_at) |
| 실행자 의견 (free text) | yes | structured form 의 "comment" 칸 |

evidence storage: 권장 — Notion DB / Google Drive folder / S3 prefix (per release tag). retention: GA 후 1 년 (compliance audit 필요 시).

### Phase 4. Critical bug triage

UAT 중 발견 bug 분류 + GA blocker 판단:

| Severity | 정의 | GA 영향 | Fix SLA |
|----------|------|---------|---------|
| **blocker** | 사용자 critical path 차단 (login 불가, signup 불가, data loss, 전체 outage) | GA block, fix 후 UAT 재실행 | 4-24h |
| **major** | workaround 있으나 release-blocking (UX 심각 저하, SLO 위반, security degradation) | GA block 또는 conditional go (hotfix ETA 명시) | 24-72h |
| **minor** | cosmetic / non-critical (typo, 디자인 inconsistency, edge-case error message 누락) | GA 통과, post-launch backlog | 차기 release |

triage 의사결정자: PM (또는 release manager). engineer + UAT 실행자 input 받아 final 결정. triage 시간: scenario 발견 후 24h 이내.

### Phase 5. Go/no-go decision + sign-off

GA decision criteria:

| Decision | Pass criteria |
|----------|---------------|
| **go** | blocker 0 + major 0 (또는 conditional go condition 명시) + sign-off ≥ 80% (실행자 / 시나리오 모두) |
| **conditional go** | blocker 0 + major ≤ 1 with hotfix ETA ≤ 48h + monitoring plan + accountable owner |
| **no-go** | blocker > 0 또는 major ≥ 2 또는 sign-off < 80% |

각 실행자 sign-off:
- 이름 + 시간 + assigned scenario 결과 (pass/fail) + free-text 의견
- "I attest that scenario UAT-XXX behaved per expected result with attached evidence" 식의 명시적 attestation
- electronic record (form submission, signed PDF) — verbal sign-off 거부 (audit trail 단절)

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지. side-effecting 산출물 (evidence storage / sign-off form submission) 은 외부 도구로 위임, 본 skill 은 plan + 결과 표 산출.

```markdown
## run-uat Output — <feature name> v<version>

### Summary
<3 줄: total scenario / executed by N person / blocker count / major count / decision>

### UAT Scenario Table
| Scenario ID | Source | Actor | Owner | Slot | Status | Evidence link |
|-------------|--------|-------|-------|------|--------|---------------|
| UAT-001 | cross-actor flow #1 | User | PM | day 1 AM | pass | https://notion.so/... |
| UAT-002 | cross-actor flow #2 | User | designated user | day 1 PM | pass | https://loom.com/... |
| UAT-003 | admin RLS | Tenant Admin | Tenant Admin | day 2 AM | fail (UAT-BUG-001) | https://drive.google.com/... |
| ... | ... | ... | ... | ... | ... | ... |

### Critical Bug Triage
| Bug ID | Severity | Scenario | Description | Owner | Fix ETA | GA impact |
|--------|----------|----------|-------------|-------|---------|-----------|
| UAT-BUG-001 | blocker | UAT-003 | tenant A 의 admin 이 tenant B audit_log 1 row 노출 (RLS bypass) | backend lead | 8h | GA block |
| ... | ... | ... | ... | ... | ... | ... |

### Stakeholder Sign-off
| Executor | Role | Scenarios | Sign-off time | Verdict | Comment |
|----------|------|-----------|----------------|---------|---------|
| 홍길동 | PM | UAT-001, 004, 008 | 2026-05-NN 14:30 | pass | flow 매끄러움, signup→verify→login 자연스러움 |
| 김철수 | designated user | UAT-002, 005 | 2026-05-NN 16:00 | pass | reset 프로세스 직관적 |
| 이영희 | Tenant Admin | UAT-003, 010 | 2026-05-NN 11:00 | fail (UAT-BUG-001) | 다른 tenant 데이터 보임 — blocker |
| ... | ... | ... | ... | ... | ... |

### GA Decision
| Field | Value |
|-------|-------|
| Decision | **no-go** (또는 go / conditional go) |
| Rationale | UAT-BUG-001 (tenant isolation breach) blocker 1 발견 |
| Conditions (if conditional) | n/a |
| Owner of next action | backend lead — RLS policy fix + 재실행 UAT-003 + sign-off |
| Re-UAT trigger | UAT-BUG-001 fix merged → UAT-003 재실행 → 동일 sign-off 절차 |
| Target re-UAT date | 2026-05-NN+1 |

### Evidence Index
- Notion DB: https://notion.so/uat-saas-auth-v1.0.0
- Loom folder: https://loom.com/folder/uat-saas-auth-v1.0.0
- Bug tracker filter: https://linear.app/tag/UAT-v1.0.0

### Cascade
- **§7 run-beta-program**: UAT pass (또는 conditional go) 후 closed beta 진입 — UAT scenario 가 beta 의 baseline
- **§7 prepare-launch-checklist**: UAT 결과 (decision + evidence link) 가 launch checklist 의 row 입력
- **§8 conduct-postmortem** (해당 시): blocker 발견 후 GA 지연 시 post-action review

### Next Step
<구체 action — 1줄: 예 "UAT-BUG-001 fix → 재실행 UAT-003 → conditional go 재평가">
```

## 7. Cross-phase cascade

본 skill 이 후속 phase 에 미치는 영향:

- **§7 prepare-launch-checklist**: UAT decision + sign-off 가 launch checklist 의 "UAT pass" row 입력
- **§7 run-beta-program**: UAT pass 후 beta 진입 — UAT scenario 가 beta 의 baseline coverage
- **§7 setup-rollback-runbook**: UAT 중 발견 critical 시나리오가 rollback trigger 의 input
- **§8 handle-incident**: UAT 에서 못 잡은 issue 가 production incident 시 root cause 분석의 baseline (UAT coverage gap 식별)
- **§8 conduct-postmortem**: GA 지연 / no-go 사례의 retrospective input

## 8. 다음 skill (next in stage flow)

- `run-beta-program` — UAT pass 후 closed beta 진입
- `prepare-launch-checklist` — UAT 결과를 readiness gate 의 row 로 통합
- (re-execute 시) `run-uat` 자체 재호출 — fix merge 후 동일 scenario 재실행

권장 chain (Phase 3 §7 cascade):
```
/buddy:chain run-uat,run-beta-program,prepare-launch-checklist -- "<feature name> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `run-browser-qa`** (§6) — run-browser-qa 는 automated browser test (Playwright). 본 skill 은 designated stakeholder 가 manual 로 실행하는 acceptance. automated 가 cover 못 하는 도메인 / business rule / UX 정합성 영역.
- **vs `run-beta-program`** (§7) — run-beta-program 은 real users (5-20 early adopter), 본 skill 은 designated stakeholder (PM / admin / 1~3 user). UAT 가 먼저, beta 가 후속.
- **vs `define-acceptance-test-plan`** (§4) — 그것은 plan 작성, 본 skill 은 plan 의 critical flow 를 actually 실행. plan 의 산출물이 본 skill 의 입력.
- **vs `setup-quality-gates`** (§7) — setup-quality-gates 는 automated CI gate, 본 skill 은 human acceptance gate. 두 gate 모두 release 의 layer.
- **vs `handle-incident`** (§8) — handle-incident 는 production 에서 발생한 incident, 본 skill 은 pre-launch 의 acceptance. UAT 미통과 = pre-launch 에서 차단되어 incident 화 안 됨.

## 10. 중요 규칙

- **Read-only on production** — UAT 는 staging 에서 실행. production data 수정 금지.
- **Evidence 의무** — 모든 pass/fail 에 attached evidence (screenshot / video / form). evidence 없는 row = 미완료.
- **Stakeholder ≥ 2** — single-person UAT 는 confirmation bias risk 로 거부.
- **Sign-off electronic record** — verbal sign-off 거부, audit trail 단절.
- **Severity 정량** — "심각한 bug" 같은 vague 분류 거부, blocker/major/minor 정의 강제.
- **Conditional go 의 condition + ETA + owner 명시 의무** — vague condition 거부.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (scenario 정의 → 실행자 배정 → evidence protocol → bug triage → go/no-go)
- [ ] §6 의 6 출력 섹션 (Summary / Scenario Table / Bug Triage / Sign-off / GA Decision / Evidence Index / Cascade / Next Step) 모두 채워짐
- [ ] UAT Scenario Table 이 §4 acceptance test plan 의 모든 cross-actor flow + per-actor critical scenario 매핑
- [ ] 실행자 ≥ 2 명, 모든 scenario 에 owner + slot + backup 명시
- [ ] Evidence link 모든 row 에 존재 (status pass / fail 무관, "미수집" 도 명시)
- [ ] Critical Bug Triage 의 모든 bug 가 severity + owner + fix ETA + GA impact 명시
- [ ] Stakeholder Sign-off 의 모든 row 에 이름 + 시간 + verdict + comment, electronic record
- [ ] GA Decision 의 rationale 명시, conditional go 시 condition + ETA + owner 모두 명시
- [ ] §4 posture 적용 — go/no-go 단호, "거의 통과" 류 hedge 없음
- [ ] §0 anti-pattern 부재 — happy path 외 edge / evidence / multi-person / criteria 매핑 / sign-off 모두 충족

하나라도 no 면 해당 phase 로 회귀 후 재검증.
