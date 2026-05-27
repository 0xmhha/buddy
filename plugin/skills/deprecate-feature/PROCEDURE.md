# Deprecate Feature — sunset notice + telemetry + migration path

## 1. 목적

기존 기능을 *제거 / 변경* 시 사용자 영향 minimize. **deprecation timeline + sunset notice + telemetry + migration path** 4 영역 표준 절차.

deprecation 은 *trust 손상* 위험 영역 — *명시 + 충분 기간 + alternative 제공* 이 핵심.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feature + usage 데이터 | ✅ | artifact / knowledge | Phase 8 산출물 또는 사용자 설명 | "폐기할 feature와 현재 사용량을 알려주세요." |
| Sunset 결정 | ✅ | decision | 사용자 의사결정 | "폐기를 확정하나요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Deprecation plan (timeline + notice 5 layer + migration path) | artifact | structured document | `migrate-customers` |

## 2. 사용 시점

- §9 manage-lifecycle 의 *기능 노후화* 결정 후
- API breaking change 결정 시
- pricing tier 변경 (기존 tier 제거)
- 보안 취약점 발견된 기능 *긴급 sunset*
- 사용 0 인 *dead feature* 제거

## 3. 입력

### 필수
- 제거 대상 기능 + scope (전체 / 부분)
- 사용자 영향 측정 (`analyze-feature-adoption` 산출 — 사용자 수 / 빈도)
- 대체 path (있으면 — alternative feature / migration tool)

### 선택
- contractual obligation (enterprise 사용자 SLA 약속)
- 보안 / legal 강제 sunset 이유

## 4. Stage 흐름

### Stage 1: Deprecation 결정

deprecate vs 유지 판단:

| 조건 | deprecate 권장 |
|------|------------|
| 사용 0 (90일+) | yes — dead feature |
| 사용자 < 1% 인데 유지 비용 큼 | yes — opportunity cost |
| 보안 취약점 fix 어려움 | yes — risk |
| 새 alternative 가 우월 | yes |
| 사용자 많은데 strategic 변화 | careful — 충분 timeline |

### Stage 2: Timeline 결정

| 영향 | 권장 deprecation period |
|------|-------------------|
| Public API breaking | 12+ month |
| Internal API | 3~6 month |
| UI feature (사용자 적음) | 1~3 month |
| Pricing tier change | 6+ month (기존 사용자 grandfather) |
| Security 강제 | 즉시 disable + 30일 grace |

### Stage 3: Sunset notice 다층 발송

| layer | timing | content |
|------|------|---------|
| Email (in-app) | start of deprecation | 명시 + 이유 + alternative + timeline |
| In-app banner | 진행 중 | dismissible / persistent |
| Email reminder | 50% / 80% / 95% timeline | urgency 상승 |
| Final email | 1주 전 | "operating on borrowed time" |
| Sunset day | 즉시 | 사용 차단 + redirect to alternative |

### Stage 4: Telemetry — 영향 측정

deprecation period 동안:
- 사용량 추세 (감소 vs 정체)
- alternative 채택률
- 사용자 이탈 신호 (해당 feature only 사용자 churn)
- support ticket spike

→ 사용량 정체 시 *추가 outreach* + *migration help* 강화.

### Stage 5: Migration path 제공

기능 제거 시 *대체 경로* 명확:

| 영역 | 제공 |
|------|-----|
| Documentation | "{old} → {new}" migration guide |
| Tooling | 자동 conversion script (가능 시) |
| Compatibility shim | 일정 기간 backward-compat (deprecation alias) |
| Manual migration | 인터페이스 차이 + step-by-step |
| Support | 1:1 migration 지원 (enterprise) |

### Stage 6: Sunset 후 cleanup

deprecation 종료 후:
- 코드 제거 + test 정리
- documentation 의 *legacy* 표시
- learn → `persist-learning-jsonl` (구현됨)
- 같은 패턴 *재발 회피* 가이드

## 5. 산출물 형식

```markdown
## Deprecation Plan — {feature}

### 결정 근거
- 조건: ... (사용 / 비용 / risk / strategic)
- alternative: ...

### Timeline
- announcement: {date}
- end-of-life: {date}
- period: {N month}

### Sunset notice 5 layer
- email / banner / reminders / final / sunset day

### Migration path
- documentation / tooling / shim / manual / support

### Telemetry
- 사용량 추세 / alternative 채택 / churn / ticket

### Cleanup post-sunset
- code remove + docs + learning persist
```

## 6. 검증

- [ ] Deprecation 결정 조건 명확?
- [ ] Timeline period 영향 별 적절?
- [ ] Sunset notice 5 layer 발송?
- [ ] Telemetry 4 차원 (사용량 / 채택 / churn / ticket) 측정?
- [ ] Migration path 5 영역 제공?
- [ ] Sunset 후 cleanup + learning 영속화?

## 7. 다음 phase

- `migrate-customers` — 대규모 사용자 migration 시
- `archive-product` — feature 가 *전체 product* 일 때
- `persist-learning-jsonl` (구현됨) — 학습 영속화
- `analyze-actor-failure-rate` — sunset 후 actor 영향 측정

## 8. 참조

- API Deprecation Best Practices (Stripe API guide)
- "Don't Make Promises You Can't Keep" — API design (Joshua Bloch)
- buddy `refactor-with-rename-trace` (§5) — deprecation alias 기법
