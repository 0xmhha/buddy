# Spin Off Feature — 기능 분리 → 별도 product / repo 전환

## 1. 목적

기존 product 의 *한 기능* 이 충분히 *독립 가치* 발견 시 **별도 product / repo 로 분리**. 코드 + customer + brand + 운영 분리.

`archive-product` 의 *부분 케이스* 또는 *strategic 확장*. 종종 acquisition / 별도 funding 의 trigger.

## 2. 사용 시점

- §9 manage-lifecycle 의 *기능 가치 재평가* 결과 spin-off 적합
- internal feature 가 *외부 사용 요구* 강 (다른 product 도 사용 원함)
- 별도 funding / acquisition 가능성
- compliance 영역 분리 (예: 의료 영역만 별도)
- team 운영 효율 (큰 product 의 *작은 팀이 큰 영향*)

## 3. 입력

### 필수
- spin-off 대상 feature
- 현재 사용자 분포 (해당 feature only / multi-feature 사용)
- 코드 의존 그래프 (`derive-system-topology` 산출)

### 선택
- 시장 분석 (`analyze-market-size` 산출 — 독립 시장 규모)
- 경쟁 (`analyze-competition-and-substitutes` 산출 — 독립 시장의 경쟁)

## 4. Stage 흐름

### Stage 1: Spin-off 적합성 평가

| 차원 | 측정 |
|------|------|
| Independence | 기존 product 없이 *standalone* 동작 가능? |
| Market | 독립 TAM 충분? (`analyze-market-size`) |
| Customer overlap | 기존 customer 와 *독립 customer base* 분리? |
| Tech complexity | 분리 비용 (code / data / infra) |
| Brand fit | 분리 후 *별도 brand* 가 적합? |

→ 5 차원 모두 강 → spin-off 적합.

### Stage 2: 코드 분리 plan

| 패턴 | 적용 |
|------|------|
| Strangler Fig | 기존 코드 점진 이동 |
| Library extract | shared logic 을 library 로 |
| Database split | shared DB → 각 product own DB |
| API boundary | 분리 후 API 로 통신 |
| CI/CD 분리 | repo 분리 → 별도 release |

→ 분리 timeline (보통 3~6 month). zero-downtime 유지.

### Stage 3: Customer 결정 — opt-in / 자동 / migration

| segment | 처리 |
|---------|-----|
| Spin-off 만 사용 | 자동 이동 + 안내 |
| 둘 다 사용 | 양쪽 가입 유지 (별도 billing) |
| 기존 product 만 사용 | 영향 없음 (해당 feature 만 분리) |

→ billing / authentication 분리 결정 (single-sign-on 유지 vs 완전 분리).

### Stage 4: Brand 분리

| 영역 | 결정 |
|------|-----|
| 이름 | 별도 brand vs sub-brand (e.g. "ProductB by ParentCo") |
| 도메인 | 별도 domain |
| 로고 / visual | 별도 design vs parent 정합 |
| 마케팅 | 별도 channel 또는 cross-promotion |
| Support | 별도 team 또는 shared |

### Stage 5: 운영 분리

| 영역 | 결정 |
|------|-----|
| Engineering | 별도 team vs shared |
| Product mgmt | 별도 PM |
| GTM | 별도 sales / marketing |
| Finance | revenue / cost 별도 추적 |
| Legal | 별도 entity (LLC / corp) |

→ 단계 별 분리 — *완전 분리 (separate company)* vs *internal product* 결정.

### Stage 6: Funding / acquisition 옵션

spin-off 의 *추가 가치 unlock* :
- **Internal**: 별도 product line 으로 운영 (cross-subsidize 또는 별도 P&L)
- **VC funding**: 별도 entity 로 funding 유치
- **Acquisition target**: 다른 회사가 acquire (return)
- **Open source**: 코드 OSS 화 (community + 차후 commercial)
- **Sunset**: 분리 시도 실패 시 archive

### Stage 7: Telemetry + 사후 분석

분리 후 6 month / 1 year :
- 두 product 의 metric 비교
- customer overlap 변화
- engineering 효율 (LOC / velocity / quality)
- financial impact

→ 사후 *blameless retro* + lessons → `persist-learning-jsonl` (구현됨).

## 5. 산출물 형식

```markdown
## Spin-Off Plan — {feature → new product}

### 적합성 평가
| 차원 | score (1~5) |

### 코드 분리 plan
- pattern: Strangler / library / DB split
- timeline: ...

### Customer 결정
| segment | 처리 |

### Brand 분리
| 영역 | 결정 |

### 운영 분리
| 영역 | 분리 정도 |

### Funding / acquisition 옵션
- ...

### Telemetry 6/12 month
- ...
```

## 6. 검증

- [ ] 5 차원 적합성 평가?
- [ ] 코드 분리 패턴 명시 (Strangler / library / DB / API / CI)?
- [ ] Customer segment 별 처리?
- [ ] Brand 분리 5 영역 결정?
- [ ] 운영 분리 5 영역 결정?
- [ ] Funding / acquisition 옵션 명시?
- [ ] 6/12 month telemetry 계획?

## 7. 다음 phase

- `archive-product` cascade — 분리 시도 실패 시
- `define-product-spec` — 분리 후 신규 product 의 *fresh* PRD
- `decide-target-market` — 신규 product 의 region 결정
- `persist-learning-jsonl` (구현됨) — 학습 영속화

## 8. 참조

- Strangler Fig pattern (Martin Fowler)
- Eric Ries — *Pivot* vs *Spin-off* 차이
- Google Spin-off 사례 (Niantic / Verily / Loon)
- buddy `archive-product` (§9) — 책임 분리 (분리 vs 종료)
