# Analyze Cost Anomaly — 비용 spike 탐지 + root cause + budget alert

## 1. 목적

cloud / SaaS 비용의 *비정상 증가* 자동 탐지 + root cause 분석 + 예방 조치. **anomaly detection + drill-down + alert 정책** 통합.

`audit-cost-efficiency` (§6, 구현됨) 와 책임 분리 — efficiency 는 *최적화 일반*, anomaly 는 *spike 사건 대응*.

## 2. 사용 시점

- 월말 invoice 확인 시 *예상 초과* 발견
- 분기 cost review — trend 이상 식별
- cloud bill alert 발화 시 (real-time)
- 신규 feature release 후 — 비용 영향 측정
- chaos-test 결과 — failure injection 시 비용 spike

## 3. 입력

### 필수
- cloud / SaaS billing data (월 / 일 단위)
- service / component 별 cost attribution
- 예산 baseline (월 / 분기)

### 선택
- `derive-system-topology` — component 매핑
- traffic / usage 데이터 (cost vs usage 정합)

## 4. Stage 흐름

### Stage 1: Anomaly detection 알고리즘

| 방법 | 적용 |
|------|------|
| Threshold | "월 $X 초과 시 alert" — 단순 / false positive 많음 |
| Moving average | 7일 / 30일 평균 대비 ±2σ — 정확도 높음 |
| Day-of-week pattern | 요일 별 baseline (예: 평일 vs 주말) |
| Seasonality | 분기 / 연말 패턴 |
| ML-based | Prophet / AWS Cost Anomaly Detection |

→ 단순 threshold + moving average 조합 권장 (start simple).

### Stage 2: Drill-down 차원

| 차원 | 측정 |
|------|------|
| Service | EC2 / RDS / S3 / Lambda / etc |
| Component | service × instance type |
| Region | 지역 별 |
| Account / Project | multi-account 시 |
| Tag | environment / team / feature |

→ tag 가 *정확한 attribution* 의 핵심. 사전 tagging 정책 필수.

### Stage 3: Root cause 분류

| 종류 | 예 |
|------|---|
| Traffic spike | 마케팅 캠페인 / 외부 referral |
| Configuration drift | autoscaling 설정 잘못 |
| Bug (cost-related) | infinite loop / log volume spike |
| Forgotten resource | 분기 종료 후 안 끈 dev env |
| 3rd-party 변경 | API 단가 변동 / SaaS pricing tier change |
| Security incident | bitcoin mining 등 abuse |

### Stage 4: Alert 정책

| level | trigger | 대응 |
|-------|--------|-----|
| Info | 일 비용 +20% | dashboard 표시 |
| Warning | 일 비용 +50% or 월 50% 도달 | Slack 알림 |
| Critical | 월 100% 초과 or 보안 의심 | on-call page + 즉시 조사 |

→ alert fatigue 회피 — Critical 만 page.

### Stage 5: Recovery action

발견 후 *5 분 안* 절차:
1. anomaly 확인 (false positive 검증)
2. drill-down (어느 service / component)
3. root cause 분류
4. 즉시 mitigation (e.g. resource 끄기 / autoscaling 제한)
5. 사후 postmortem trigger (`conduct-postmortem` 구현됨)

## 5. 산출물 형식

```markdown
## Cost Anomaly Report — {date}

### Detection
- 알고리즘: {moving average + threshold}
- trigger: {일 비용 +X% / 월 budget Y%}

### Drill-down
| service | component | 일반 cost | spike cost | 차이 |

### Root cause
- 분류: {traffic / config / bug / forgotten / 3rd-party / security}
- 상세: ...

### Mitigation 5-step log
1. ...

### Prevention
- alert 정책 갱신: ...
- runbook 추가: ...
```

## 6. 검증

- [ ] Anomaly 알고리즘 (threshold + moving average) ?
- [ ] Drill-down 5 차원 (service / component / region / account / tag) ?
- [ ] Tagging 정책 활성?
- [ ] Root cause 6 분류 적용?
- [ ] Alert level 3 단계 (info / warning / critical) ?
- [ ] Recovery 5-step 명시?

## 7. 다음 phase

- `audit-cost-efficiency` (§6) 와 cross-reference — anomaly fix 가 *상시 efficiency* 영향
- `conduct-postmortem` (구현됨) — incident 학습
- `audit-error-budget` (§8) — 비용 incident 가 SLO budget 도 burn

## 8. 참조

- AWS Cost Anomaly Detection / GCP Recommender — anomaly 도구
- FinOps Foundation — cost engineering practices
- buddy `audit-cost-efficiency` (§6, 구현됨) — efficiency 일반 영역과 분리
