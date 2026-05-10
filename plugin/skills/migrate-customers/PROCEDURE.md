# Migrate Customers — 대규모 customer migration plan

## 1. 목적

product / API / pricing tier / data schema 의 *큰 변경* 시 기존 customer 의 **migration plan**. **batch 분할 + 자동 vs 수동 + rollback + communication** 통합.

`deprecate-feature` 의 *대규모 사용자 영향 케이스* 후속 — single feature 가 아닌 *제품 전체 영향* 시.

## 2. 사용 시점

- §9 manage-lifecycle 의 *대규모 변경* 결정 후
- product rebrand / merger / acquisition
- DB schema major migration (zero-downtime 불가능 영역)
- API v1 → v2 강제
- pricing tier 전면 개편
- 데이터 hosting region 변경 (compliance / cost)

## 3. 입력

### 필수
- migration 대상 (product / API / data / pricing)
- customer base 분포 (`analyze-user-cohort` 산출 — segment / tenure / tier)
- 변경 범위 (full / partial)

### 선택
- contractual obligation (enterprise SLA)
- compliance 강제 (예: data sovereignty 변경)

## 4. Stage 흐름

### Stage 1: Customer segmentation for migration

| segment | 특성 | migration 전략 |
|---------|-----|-----------|
| Tier-1 (enterprise) | 큰 ARR / dedicated CSM | 1:1 migration support |
| Tier-2 (mid) | 중간 사용량 | 자동 + opt-in 검증 |
| Tier-3 (SMB / free) | 자동 적용 | 자동 migration + email notice |
| Inactive | 90일+ 미사용 | 자동 + 재활성화 mail |

각 segment 의 *risk + effort* 다름 — strategy 별 분리.

### Stage 2: Migration 자동 vs 수동 결정

| 조건 | 권장 |
|------|-----|
| Schema mapping 명확 | 자동 |
| Customer 결정 필요 (예: tier 선택) | 수동 / opt-in |
| Data 손실 가능성 | 수동 + 사용자 confirm |
| Reversible | 자동 OK |
| Irreversible | 수동 + double-confirm |

### Stage 3: Batch schedule

전체 한 번에 X — *batch 분할*:

| batch | size | 시점 | 검증 |
|-------|-----|------|------|
| 1. Internal | 본인 / 팀 | week 1 | smoke test |
| 2. Friendly customer | 10~50 (opt-in) | week 2 | feedback 수집 |
| 3. Tier-3 자동 | 10% rollout | week 3 | metric 모니터링 |
| 4. Tier-3 자동 (확대) | 50% | week 4 | drift 검증 |
| 5. Tier-3 자동 (전체) | 100% | week 5 | sustained 검증 |
| 6. Tier-2 opt-in | 일정 기간 | week 6+ | individual confirm |
| 7. Tier-1 1:1 | individual | quarter | dedicated support |

→ batch 간 *abort condition* 명시 — fail 시 hold + rollback.

### Stage 4: Rollback 준비

각 batch:
- pre-migration snapshot (data backup)
- rollback procedure 문서화
- rollback 결정권자 + 절차
- 사용자 communication template (rollback 시 안내)

### Stage 5: Communication

| layer | timing |
|------|------|
| Pre-announcement | 30일 전 — email + in-app |
| Detailed notice | 14일 전 — migration date + 무엇 변하나 |
| Reminder | 7일 / 1일 전 |
| Day-of | migration 시점 — banner + email |
| Post-migration | 1주 후 — "어떻게 동작?" 후속 + feedback 요청 |

### Stage 6: Telemetry + 사후 분석

- migration success rate (segment 별)
- error rate spike
- support ticket spike (분류 별)
- churn 측정 (migration 직후 30일)
- NPS 변화

→ 사후 *blameless retro* (`conduct-postmortem` 구현됨).

## 5. 산출물 형식

```markdown
## Customer Migration Plan — {변경}

### Segment / strategy
| segment | 특성 | strategy |

### 자동 vs 수동
- 자동 영역: ...
- 수동 영역: ...

### Batch schedule
| batch | size | week | 검증 | abort 조건 |

### Rollback 준비
- snapshot / procedure / 결정권자

### Communication 5 layer

### Telemetry
- success rate / error / ticket / churn / NPS
```

## 6. 검증

- [ ] Customer segmentation (tier 별)?
- [ ] 자동 vs 수동 결정 명시?
- [ ] Batch schedule (internal → friendly → 자동 → opt-in → 1:1)?
- [ ] Rollback 준비 (snapshot + procedure)?
- [ ] Communication 5 layer?
- [ ] Telemetry 5 metric + 사후 retro?

## 7. 다음 phase

- `deprecate-feature` 와 cascade — *기능 sunset 의 대규모 케이스*
- `archive-product` — *전체 product* 시
- `conduct-postmortem` (구현됨) — 사후 학습
- `analyze-customer-feedback-corpus` (§8) — migration feedback 분석

## 8. 참조

- AWS Well-Architected: migration patterns
- Strangler Fig pattern (Martin Fowler) — 점진 migration
- Stripe API migration playbook (대규모 customer migration 사례)
