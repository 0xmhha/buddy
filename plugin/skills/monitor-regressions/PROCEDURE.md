---
name: monitor-regressions
description: "[패턴 라이브러리] delta-based threshold + transient tolerance + per-page isolation으로 monitoring과 regression detection 설계 (console err, perf, page fail). 직접 invoke보다 orchestrator가 import해 사용. 트리거: '회귀 모니터' / 'console err 추적' / 'perf 회귀' / 'page fail 감시'. 참조 위치: write-changelog 후 배포 모니터, triage-work-items 입력 소스, summarize-retro 데이터."
type: skill
---

# Snippet: Monitoring Patterns (Delta + Transient + Isolation)


> 모니터링/threshold 패턴 — 브라우저 자동화 및 리포팅 인프라는 도구별로 분리.

## 이 snippet을 사용하는 경우
- CI 성능 회귀 detector 구축
- 배포 후 canary 모니터링 루프 구축
- "새 측정치를 baseline과 비교, 회귀 시 alert"하는 모든 워크플로우

## 패턴 1: Delta 기반 Threshold

### 왜 delta, absolute가 아닌가?
Absolute threshold는 환경과 시간에 따라 drift. 2000ms 로드 타임은 복잡한 대시보드엔 괜찮고 랜딩 페이지엔 끔찍. 업계 표준이 아니라 **자기** baseline과 비교. Delta 기반 threshold는 느리지만 안정적인 시스템에서 false alert 없이 *변화*를 잡는다.

### 공식

**어떤** 조건이라도 만족 시 alert:

```
relative_increase >= 0.50    # baseline보다 50% 느림
OR
absolute_increase >= 500ms   # 0.5초 wall-clock 증가
```

`OR` (AND 아님)이 중요: 10ms → 100ms인 빠른 엔드포인트는 +900%지만 +90ms — relative 규칙으로 잡힘. 5s → 5.6s인 느린 엔드포인트는 +12%지만 +600ms — absolute 규칙으로 잡힘.

### 메트릭별 튜닝 가능

| 메트릭 | Relative threshold | Absolute threshold |
|--------|--------------------|--------------------|
| TTFB | 50% | 200ms |
| FCP | 50% | 500ms |
| LCP | 30% | 800ms |
| CLS | 100% | 0.05 |
| Bundle size | 25% | 50KB |
| Page load (canary) | 100% (2x) | 1000ms |

프로젝트별 조정 — 시작점. Bundle size는 byte-count가 결정적이고 더 작은 delta가 의미 있어 25%; 타이밍 메트릭은 네트워크 variance를 흡수하기 위해 더 넓은 band 필요.

## 패턴 2: Transient Tolerance

### 왜?
단일 샘플 alert는 네트워크 jitter, GC pause, 콜드 캐시, 일회성 transient 에러(rate limit, 드롭된 패킷)로 false positive 생성. 거짓 경보를 남발하지 말 것.

### 규칙
Alert 전 threshold 초과 **연속 ≥2 관측** 요구. 단일 transient 네트워크 blip은 alert가 아니다.

### 구현

```
state = { metric: [...recent_observations] }   # 메트릭별, 엔드포인트별

on each new observation:
  state.metric.push(observation)
  if state.metric.length > 3: state.metric.shift()  # 최근 3개 유지

  recent_failures = state.metric.filter(o => o > threshold).length
  if recent_failures >= 2:
    ALERT
  else:
    NO ALERT  (transient일 수 있음 — 다음 체크 대기)
```

### 트레이드오프
- 빠른 감지 = 낮은 threshold count (1-2)
- 적은 false positive = 높은 threshold count (3-5)

기본값: **3개 중 2개** (최근 3 관측 중 2 실패). 경계 시 사용자 대상 메시지는 즉시 알림 발화 대신 "transient일 수 있음 — 다음 체크 대기"가 맞다.

## 패턴 3: 페이지별 (엔드포인트별) 격리

### 왜?
`/checkout`의 회귀가 `/home`의 signal을 덮으면 안 된다. 모니터링 시 페이지/엔드포인트 간 집계 금지. Baseline에 콘솔 에러 3개인 페이지는 여전히 3개면 괜찮다 — 다른 페이지의 새 에러 1개가 alert.

### 규칙
- 각 페이지/엔드포인트는 자기 baseline 보유
- 각 페이지/엔드포인트는 자기 threshold 카운터 보유 (패턴 2의 transient state)
- Alert는 "site 전체"가 아니라 어떤 페이지가 회귀했는지 식별

### 저장 레이아웃

```
<state-dir>/baselines/
  baseline.json         # manifest: 페이지 리스트 + 참조 값
  home.json             # { url, metrics: {ttfb, fcp, ...}, samples: N }
  checkout.json
  product-detail.json
  ...
  screenshots/          # 선택: 페이지별 시각 baseline
    home.png
    checkout.png
```

각 baseline 파일: N 샘플 평균 (안정성을 위해 전형적으로 N >= 10).

## Baseline 캡처 + 비교 흐름

### Phase 1: Baseline 캡처 (일회성 또는 의도된 변경 후)

```
for each page/endpoint:
  for sample in 1..N:        # N >= 10
    measure metrics
    record sample
  baseline = compute_average(samples)
  save baseline to <state-dir>/baselines/<page>.json
```

Baseline 재캡처 시점:
- 의도된 성능 작업 후
- Bundle size에 영향 주는 의존성 업그레이드 후
- 호스팅/infra 변경 후
- 건강한 canary로 성공한 배포 (새 normal)
- 시간 감쇠: 무관하게 30-90일마다

Baseline이 없으면 지금 배포 전 snapshot을 참조점으로. Baseline 없으면 absolute 숫자는 보고할 수 있지만 회귀는 감지 불가 — 첫 단계로 baseline 캡처 권장.

### Phase 2: 새 Run 비교

```
for each page/endpoint:                          # 패턴 3: 격리
  measurement = measure now
  baseline = load baseline for this page
  delta_relative = (measurement - baseline) / baseline
  delta_absolute = measurement - baseline

  # 패턴 1: delta 기반 threshold (OR, AND 아님)
  if delta_relative >= rel_threshold OR delta_absolute >= abs_threshold:
    update transient_state[page][metric]         # 패턴 2: tolerance
    if transient_state says ALERT (>=2 of last 3):
      report regression for THIS page
```

### Phase 3: 리포트

```
Performance Regression Report
=============================

Pages with regressions (>=2 of last 3 observations):
  /checkout
    LCP: 2.1s -> 3.4s (+62%, +1300ms)   ALERT
    Bundle: 240KB -> 290KB (+21%, +50KB) ALERT

  /product-detail
    FCP: 800ms -> 950ms (+19%, +150ms)  no alert (below threshold)

Pages within tolerance: 8 of 10

Action recommended:
- Investigate /checkout LCP regression (largest delta)
- Bisect commits since last baseline
- Continue monitoring /product-detail (single observation, may be transient)
```

### Phase 4: Baseline 업데이트 (건강한 run 후)

Canary/벤치마크가 깔끔히 통과하면 baseline 갱신 제안:
- A) 현재 측정치로 baseline 업데이트 (배포가 새 normal)
- B) 이전 baseline 유지 (추가 확인 대기)

Run이 건강했으면 A 권장 — baseline drift는 *업데이트 안 할 때의* 비용.

## 안티패턴

- 단일 샘플 alerting (transient 노이즈에 alert)
- Site-wide 집계 (페이지별 signal 상실 — `/checkout` 회귀가 `/home` 개선에 숨음)
- Absolute-only threshold (시간에 따라 drift, 환경 간 misfire)
- 매 run 후 re-baseline (baseline 유지 목적 무력화)
- 빠른 메트릭과 느린 메트릭에 같은 threshold (10ms 엔드포인트의 90ms 증가는 중요; 5s 엔드포인트의 90ms 증가는 아님 — 메트릭별 튜닝 사용)
- Relative + absolute threshold의 AND 조합 (양 극단에 블라인드 스팟 생성 — OR 사용)
- Delta가 아니라 absolute count에 alert ("콘솔 에러 3개"는 baseline이 3이면 괜찮; "새 에러 1개"가 alert)
