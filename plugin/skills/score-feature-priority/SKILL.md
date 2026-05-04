---
name: score-feature-priority
description: This skill should be used when the user wants to "prioritize features", "score features", "rank the backlog", "decide what to build first", "apply RICE scoring", or needs to prioritize a list of features using a structured framework.
---

# score-feature-priority

RICE / ICE / MoSCoW 프레임워크로 feature 우선순위를 결정한다. §2 `define-features`의 일곱 번째 stage.

---

## 프레임워크 선택

| 프레임워크 | 적합한 상황 |
|-----------|-----------|
| **RICE** | 데이터가 있을 때. 정량적 판단. |
| **ICE** | 빠른 rough scoring. 스타트업 초기. |
| **MoSCoW** | 출시 범위 결정. 이해관계자 정렬. |

모두 적용하고 결과를 비교하는 것을 권장.

---

## RICE Scoring

```
RICE Score = (Reach × Impact × Confidence) / Effort

Reach: 분기 내 영향받는 사용자 수 (실측 또는 추정)
Impact: 목표 달성 기여도 (1=minimal, 2=low, 4=medium, 8=high, 3=massive)
Confidence: 추정 신뢰도 (100%=high, 80%=medium, 50%=low)
Effort: 구현 공수 (person-months, 소수점 허용)
```

예시:
```yaml
features:
  - feature_id: signup-email-password
    rice:
      reach: 1000      # 월 1,000명 신규 가입 예상
      impact: 8        # high — 핵심 진입점
      confidence: 80   # 데이터 없음, 유사 제품 참조
      effort: 0.5      # 2주 (0.5 person-month)
      score: 12800     # (1000 × 8 × 0.80) / 0.5
      
  - feature_id: oauth-google-login
    rice:
      reach: 800       # 기존 가입자 중 Google 선호 추정
      impact: 4        # medium — 대안 수단
      confidence: 60   # 낮은 신뢰도
      effort: 0.75     # 3주
      score: 2560      # (800 × 4 × 0.60) / 0.75
```

## ICE Scoring (빠른 판단)

```
ICE Score = Impact × Confidence × Ease

Impact: 얼마나 큰 효과인가? (1-10)
Confidence: 얼마나 확신하는가? (1-10)
Ease: 얼마나 쉽게 구현할 수 있는가? (1-10)
```

## MoSCoW 분류

```
Must:    출시에 없으면 제품이 성립하지 않는 것
Should:  중요하지만 대안이 있는 것
Could:   있으면 좋지만 생략 가능한 것
Won't:   이번 릴리즈 범위 밖 (나중에 재검토)
```

---

## 우선순위 결정 절차

### 1. Feature 목록 입력

`define-feature-spec` 출력의 feature 목록을 입력으로 받는다.

### 2. RICE Score 계산

```yaml
feature_priorities:
  - feature_id: {id}
    rice_score: {N}
    ice_score: {N}
    moscow: {Must/Should/Could/Won't}
    rationale: {결정 근거 1줄}
```

### 3. 최종 우선순위 결정

```
우선순위 = RICE score가 주, MoSCoW Must가 override
Must + RICE 상위 = Sprint 1
Should + RICE 중위 = Sprint 2
Could + RICE 하위 = Sprint 3+
Won't = backlog (비활성)
```

### 4. 리소스 기반 조정

implementation track (actor별) 기준으로 병렬 가능한 feature를 묶는다:
- 같은 frontend track 의존 feature들은 순차 처리
- 서로 다른 actor track 의존 feature들은 병렬 처리 가능

---

## 출력 형식

```markdown
## Feature Priority Board

| 순위 | Feature | RICE | MoSCoW | Sprint | 이유 |
|------|---------|------|--------|--------|------|
| 1 | signup-email-password | 12,800 | Must | 1 | 핵심 진입점 |
| 2 | password-reset | 5,200 | Must | 1 | 보안 필수 |
| 3 | oauth-google-login | 2,560 | Should | 2 | 전환율 향상 |
| 4 | two-factor-auth | 1,800 | Could | 3 | 보안 강화 |

### Sprint 1 (Must): {feature list}
### Sprint 2 (Should): {feature list}
### Sprint 3+ (Could): {feature list}
### Won't (이번 릴리즈 제외): {feature list}
```

---

## 다음 단계

→ `map-feature-dependencies` — feature 간 선후 그래프
