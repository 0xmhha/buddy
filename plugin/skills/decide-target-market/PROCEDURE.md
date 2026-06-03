# Decide Target Market — 글로벌 / 단일 지역 / 다지역 결정

## 1. 목적

`assess-business-viability` 가 *사업이 되나?* 를 검증한 후, 이 skill 은 **"어느 시장에 먼저 진출하나?"** 를 결정한다. 글로벌 default 인지, 단일 지역 (Korea / USA / EU 등) 인지, 다지역 순차 진출인지 명시한다.

이 결정이 이후 모든 region-specific 작업의 trigger 다 — 예를 들어 `consult-korea-legal-context` (Korea cluster) 는 본 skill 의 산출이 *Korea 포함* 일 때만 활성화.

## 2. 사용 시점 (When to invoke)

- `assess-business-viability` 통과 후 *어느 시장 우선* 결정 직전
- 진출 전략 (GTM channel) 설계 전 — 시장 별 채널 다름
- 법률 / 규제 검토 (`review-legal-regulatory`) 진입 직전 — region-agnostic frame 인지 region cluster 활성화 인지 결정
- 다지역 SaaS 의 launch sequencing 결정 시
- 신규 시장 확장 (geo expansion) 검토 시점

## 3. 입력 (Inputs)

### 필수
- `assess-business-viability` 산출물 (TAM / SAM / SOM 추정, 고객 segment 가설)
- 제품 지역 친화도 (예: 한국어 콘텐츠 / 다국어 / 영어 only)

### 선택
- 사용자 명시 *우선 시장* (있으면 그대로 채택)
- 경쟁사 진출 패턴 (auto-import 가능 시)

## 4. Stage 흐름

### Stage 1: market matrix 작성

후보 시장을 4 차원으로 평가:

| 차원 | 측정 |
|------|-----|
| 시장 규모 | SAM / SOM (region 별 분할) |
| 진입 비용 | 현지화 / 법률 / 규제 / 번역 / 결제 시스템 |
| 규제 복잡도 | privacy law (GDPR / PIPA / CCPA), 산업별 규제 |
| 경쟁 강도 | direct / indirect / substitute count |

후보 시장 = 최소 3 (글로벌 / Korea / USA / EU / 기타) 평가.

### Stage 2: 우선순위 결정

3 옵션 중 하나 채택:

| 옵션 | 조건 | trigger out |
|------|-----|----------|
| 글로벌 default | 다국어 / 영어 dominant + 현지화 비용 낮음 | region-agnostic core 만 활성화 |
| 단일 지역 우선 | 특정 지역 시장 규모 dominant + 현지화 효과 큼 | 해당 region cluster 활성화 |
| 다지역 순차 | 2~3 지역 sequencing — primary → secondary | 우선순위 region cluster 먼저 활성화 |

### Stage 3: region cluster trigger 결정

채택된 시장 별로 *cluster trigger* 명시:

| 시장 | trigger 결과 |
|------|-----------|
| 글로벌 | (no region cluster) — `review-legal-regulatory` 만 |
| Korea | Korea cluster 3 활성화 (`consult-korea-legal-context` / `draft-korea-patent-application` / `audit-korea-cii-vulnerability`) |
| USA | (template 만 — 외부 자산 부재, 신규 작성 필요) |
| EU | (template 만 — GDPR 등) |

### Stage 4: ADR 작성

`write-adr` skill invoke — 본 결정을 ADR 로 영속화. context (Stage 1 matrix) + decision (Stage 2 옵션) + consequences (cluster activation) + alternatives (rejected 옵션들).

## 5. 산출물 형식

```markdown
## Target Market Decision — {제품 이름}

### Market Matrix
| 시장 | SAM | 진입 비용 | 규제 | 경쟁 |
|------|-----|---------|-----|-----|

### Decision
**Primary**: {글로벌 / Korea / USA / EU / 기타}
**Secondary** (다지역): {ordered list}
**Rationale**: {한 단락}

### Region Cluster Trigger
{활성화될 cluster 명시 — 예: Korea cluster 3 skill activate / 글로벌 default}

### ADR Reference
[ADR-{N}](../superpowers/decisions/{date}-target-market.md)
```

## 6. 검증 (self-check)

- [ ] Stage 1 matrix 가 ≥3 시장 평가됐는가?
- [ ] Stage 2 결정이 3 옵션 중 하나로 명시됐는가?
- [ ] Stage 3 cluster trigger 가 *명시적* 인가? (silent activation 금지)
- [ ] 글로벌 default 채택 시 region-agnostic core (review-legal-regulatory) 만 활성화 되는가?
- [ ] 단일 / 다지역 채택 시 해당 cluster trigger 가 cascade out 되는가?
- [ ] ADR 작성됐는가?

## 7. 다음 phase

본 skill 산출이 *region cluster trigger* 결과에 따라 다음 분기:

- 글로벌 → `review-legal-regulatory` (region-agnostic frame)
- Korea → `consult-korea-legal-context` + `review-legal-regulatory`
- USA / EU / 기타 → (template 작성 필요 — trigger 발생 시 즉시 신규 skill)

## 8. 참조

- [`docs/two-tracks-charter.md`](../../../docs/two-tracks-charter.md) §3 — plugin buddy scope 12 stage
- 2026-05-10 D-G 정정 lock-in: region-cross-cutting 3 layer framework — Layer 1 (region 결정, 본 skill) / Layer 2 (region-agnostic core = `review-legal-regulatory`) / Layer 3 (region-specific extension = Korea cluster 3건 deferred / USA·EU 0건). 원본 inventory note 는 commit `9ab21dd` 로 제거, 결정 자체는 유지.
- HBR market entry frameworks / GTM strategy (도메인 source — Step 4 plan §5 reference)
