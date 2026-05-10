# Analyze Market Size — TAM / SAM / SOM 산출 + cross-check

## 1. 목적

PRD 또는 사업성 가설을 입력으로 **TAM / SAM / SOM** (Total / Serviceable / Serviceable-Obtainable Addressable Market) 을 산출. *bottom-up* 과 *top-down* 두 방법을 동시 적용해 cross-check 로 정확도를 확보한다.

`assess-business-viability` 의 *시장 규모 차원* 입력으로 활용. 단독 호출도 가능 (예: 투자 자료 deck 작성 전).

## 2. 사용 시점

- `validate-idea` 통과 후 사업성 7차원 평가 (`assess-business-viability`) 직전 또는 안에서 호출
- 투자 deck (Series A/B) 작성 전 TAM 검증
- 신규 시장 확장 전 SAM / SOM 재산정
- 경쟁사 등장으로 *내가 가져갈 share* (SOM) 재계산
- pivot 결정 시 *작은 SAM 으로 좁히기* 검토

## 3. 입력

### 필수
- 제품 / 서비스 정의 (PRD 의 problem statement + value hypothesis)
- 가격 모델 가설 (subscription / usage / freemium)
- 후보 시장 (글로벌 / 특정 지역 — `decide-target-market` 산출 활용 가능)

### 선택
- 경쟁사 매출 / 사용자 수 (top-down 입력)
- 사용자 직접 측정 가능한 customer count (bottom-up 입력)

## 4. Stage 흐름

### Stage 1: Top-down

공식 통계 / industry report 에서 TAM 추정.

| 항목 | source 예시 |
|------|----------|
| TAM | Statista / Gartner / IDC / 산업 협회 보고서 |
| 검색 도구 | tavily-mcp (외부 검색) / arxiv-mcp-server (학술) |

산식:
- TAM = total customers × average revenue per customer (annual)

### Stage 2: Bottom-up

내 제품의 *실제 reachable* 고객으로부터 SAM / SOM 추정.

| 항목 | 산식 |
|------|------|
| SAM | TAM × 지역 / 산업 / segment 필터 |
| SOM | SAM × 진입 첫 1~3 년 *현실적 share* (보통 1~5%) |

### Stage 3: Cross-check

Top-down (Stage 1) 과 Bottom-up (Stage 2) 의 *불일치 검토*.

| 불일치 패턴 | 진단 |
|-----------|------|
| Top-down ≫ Bottom-up | 시장 정의 너무 광범위, 또는 reachable 작음 — SAM 다시 좁힘 |
| Top-down ≪ Bottom-up | 시장 정의 너무 좁음, 또는 SOM share 가정 과대 — 가정 재검토 |
| 비슷 (±20% 이내) | 추정 정합 — 통과 |

### Stage 4: Sensitivity analysis

핵심 가정의 ±50% 변동 시 SOM 변화 측정. 가정 fragile 식별.

## 5. 산출물 형식

```markdown
## Market Size Analysis — {제품 이름}

### Top-down (Stage 1)
- TAM: ${X}M (source: ...)
- 산식: customers × ARPC

### Bottom-up (Stage 2)
- SAM: ${Y}M (지역 / segment 필터)
- SOM (year 1): ${Z}M (share %)

### Cross-check (Stage 3)
- Top-down vs Bottom-up: {match / mismatch}
- 불일치 시 조정: ...

### Sensitivity (Stage 4)
- ARPC ±50%: SOM ${A}~${B}M
- Share ±50%: SOM ${C}~${D}M

### Conclusion
- Market 가 *충분히 큰가?* {yes / no / borderline}
- 다음 phase 입력: assess-business-viability §5 시장 차원
```

## 6. 검증

- [ ] Top-down + Bottom-up 두 방법 모두 적용?
- [ ] Cross-check 결과 명시?
- [ ] Sensitivity ±50% 검토?
- [ ] source 명시 (검증 가능한 통계)?
- [ ] 결론이 *go / pivot / no-go* 입력으로 사용 가능한가?

## 7. 다음 phase

- `assess-business-viability` 의 *시장 규모 차원* 입력
- 또는 `define-product-spec` 의 *target user* 입력 (segment 가 narrow 해진 경우)

## 8. 참조

- HBR market sizing frameworks
- tavily-mcp / arxiv-mcp-server (검색 도구)
- [`assess-business-viability`](../assess-business-viability/PROCEDURE.md) — cascade out
