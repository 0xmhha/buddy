# Decide Form Factor — App vs Web vs Hybrid 결정

## 1. 목적

§3 design-system 의 *stage 0* — `define-tech-stack` 직전 발화. **앱 (native iOS/Android, hybrid) vs Web (SPA, MPA, PWA) vs Desktop (Electron / Tauri)** 결정.

이 결정은 *다년 락인* — 후속 design / build / deploy / maintenance 모두 영향. 잘못 정하면 *재구현 비용 1~2 자릿수* 증가.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 제품 요구사항 및 target 사용자 | ✅ | knowledge | 사용자 도메인 지식 | "제품의 주요 사용 환경은? (모바일/데스크톱/오프라인 등)" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Form factor 결정 (app/web/hybrid/desktop + 근거) | decision | ADR 또는 inline record | `define-tech-stack` |

## 2. 사용 시점

- §3 design-system 진입 직전 — `assess-business-viability` 통과 후
- pivot 결정 시 — 기존 form factor 가 *user 의 사용 패턴* 과 mismatch 시
- 신규 distribution channel 결정 시 (예: 모바일 시장 진출)
- 경쟁사가 다른 form factor 채택 후 *비교 검토*

## 3. 입력

### 필수
- PRD + value hypothesis
- `map-customer-segments` — early adopter 의 *device 패턴* (mobile-first / desktop / 둘 다)
- `map-jobs-to-be-done` — job 이 *어디서 / 언제* 발생하는지 (이동 중 / 데스크 / 가정)
- `decide-target-market` — 지역 별 device 친숙도 (예: 한국 모바일 dominant)

### 선택
- 경쟁사 form factor 분포 (`analyze-competition-and-substitutes` 산출)
- 제약 (offline 사용 / push 알림 / camera / GPS 등 native API)

## 4. Stage 흐름

### Stage 1: Form factor 후보 매트릭스

| form factor | distribution | 개발 비용 | 지속 비용 | platform 의존 |
|-----------|----------|---------|---------|------------|
| **Web (SPA)** | URL — 즉시 | LOW | LOW | 없음 (브라우저 호환) |
| **Web (PWA)** | URL + install | LOW-MED | LOW | 일부 (push 제한) |
| **Native iOS** | App Store | HIGH | HIGH (review process) | Apple |
| **Native Android** | Play Store / sideload | HIGH | MED | Google |
| **Hybrid (React Native / Flutter)** | Stores | MED-HIGH | MED | 양쪽 |
| **Desktop (Electron)** | binary download | MED | MED | OS 별 build |
| **Desktop (Tauri)** | binary download | MED | LOW | OS 별 build |

### Stage 2: 7 차원 평가

각 후보를 7 차원으로 점수 (1~5):

| 차원 | 측정 |
|------|-----|
| User device fit | early adopter 의 사용 device 정합 |
| Job context fit | job 이 발생하는 *물리 환경* (모바일 / 데스크) |
| Native API 필요 | camera / GPS / push / biometric / file 시스템 |
| Distribution speed | URL 즉시 vs Store review (1~2주) |
| 개발 비용 | 인력 / 시간 |
| 유지 비용 | 분기 release / OS update 대응 |
| 시장 정합 | 지역 / 산업의 form factor 기대 |

### Stage 3: 결정 — 단일 / 다중

| 결정 | 조건 |
|------|------|
| **단일 form factor** | 7 차원 명확히 dominant 한 후보 |
| **다중 (web + native)** | 사용 시점 다름 (예: web 으로 전환, mobile 로 일상) |
| **순차 (web → native)** | 초기 web 으로 빠른 검증, 이후 native 확장 |

### Stage 4: ADR 작성

`write-adr` skill invoke. context (Stage 1 매트릭스 + Stage 2 7 차원 점수) + decision + consequences (개발 cost / 유지 / distribution) + alternatives.

## 5. 산출물 형식

```markdown
## Form Factor Decision — {제품}

### 후보 매트릭스
| form factor | distribution | cost | platform |

### 7 차원 평가
| 차원 | Web | Native iOS | Hybrid | Desktop | weight |

### 결정
**Primary**: {form factor}
**Secondary** (다중 / 순차): {if any}
**Rationale**: {한 단락}

### ADR
[ADR-{N}](../superpowers/decisions/{date}-form-factor.md)
```

## 6. 검증

- [ ] 후보 매트릭스 ≥3 form factor?
- [ ] 7 차원 모두 점수?
- [ ] 결정이 *단일 / 다중 / 순차* 명시?
- [ ] ADR 작성?
- [ ] *재구현 비용* 인지 (1~2 자릿수 증가) 명시?

## 7. 다음 phase

- `define-tech-stack` — form factor 결정 결과로 tech stack 후보 narrow (예: web 이면 React/Vue, native 이면 Swift/Kotlin/RN)
- `apply-design-system` — form factor 별 design system (iOS HIG / Material / web component library)
- `decide-target-market` 의 region 결과와 cross-check (모바일 dominant 지역에 web only 면 mismatch)

## 8. 참조

- Mobile vs Web decision matrix (Apple HIG / Material Design)
- React Native + Flutter 비교 (도메인 reference)
- buddy 자체 charter scope 3 ("앱 vs web 결정") — Layer 1 region 결정과 같은 *foundation* layer
