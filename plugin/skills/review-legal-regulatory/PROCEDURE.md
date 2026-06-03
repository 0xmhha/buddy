# Review Legal Regulatory — region-agnostic 법률 / 규제 검토 frame

## 1. 목적

제품 빌딩 시 *법률 / 규제 영향* 의 일반 (region-agnostic) 검토 frame. **체크리스트 + evidence package** 산출 — *법률 자문 대체 X*, *전문가 검토 의뢰 입력 보강*.

`decide-target-market` 산출에 따라 *region cluster trigger* — 글로벌 default 면 본 skill 만, Korea / USA / EU 면 해당 cluster 의 region-specific skill 추가 활성화.

본 skill 의 *Layer 2* (region-agnostic core) 책임은 [`docs/two-tracks-charter.md`](../../../docs/two-tracks-charter.md) §2.4.2 의 region-cross-cutting framework 정합.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 제품 맥락 (problem statement + 데이터 흐름) | ✅ | knowledge | PRD 또는 사용자 설명 | "어떤 제품의 법률/규제 검토를 하나요?" |
| Target market 결정 | 선택 | decision | `decide-target-market` 산출물 | 없으면 region-agnostic 검토만 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Legal/regulatory review (7 sub-domain checklist + evidence package) | artifact | structured report | `prepare-launch-checklist` |

## 2. 사용 시점

- `decide-target-market` 산출 후 *법률 / 규제 영향 검토* 단계
- `prepare-launch-checklist` (§7) 직전 — 법률 readiness gate
- 신규 기능 / pivot 시 *규제 영향 재검토*
- 데이터 수집 정책 변경 시
- 결제 / 환불 정책 결정 시
- 다지역 진출 sequencing — region cluster 활성화 trigger 결정

## 3. 입력

### 필수
- PRD (problem statement + value hypothesis + 데이터 흐름)
- `decide-target-market` 산출 — region cluster 활성화 결과 (글로벌 / 단일 / 다지역)
- `map-customer-segments` 산출 — 데이터 주체 (data subject) 식별

### 선택
- 산업 명시 (medical / financial / education 등 *산업별 규제* 입력)
- 기존 약관 / 개인정보처리방침 draft (있으면)

## 4. Stage 흐름

### Stage 1: 7 sub-domain 검토 (region-agnostic)

| sub-domain | 검토 항목 | 기존 buddy skill |
|---------|--------|--------------|
| **Privacy / 개인정보** | PII 수집 / 저장 / 공유 / 삭제 | `review-privacy-data-risk` (구현됨) — sub-skill 호출 |
| **IP / 저작권 / 라이선스** | 본인 창작물 / 외부 라이브러리 / 사용자 콘텐츠 | `review-license-and-ip-risk` (구현됨) — sub-skill 호출 |
| **AI 책임** | AI 출력의 정확성 / 책임 범위 / 사용자 사전 고지 | `review-ai-safety-liability` (구현됨) — sub-skill 호출 |
| **약관 / 개인정보처리방침 readiness** | 약관 작성 + 공개 + 사용자 동의 흐름 | (본 skill stage 4) |
| **결제 / 환불 / 소비자보호** | 결제 PG 정합 + 환불 규정 + 분쟁 절차 | (본 skill stage 4) |
| **산업별 규제** | medical (HIPAA / 의료법) / financial (PCI / 자금세탁) / education (FERPA / 개인정보) | (본 skill stage 5) |
| **AI 서비스 책임 / 감사 의무** | AI 결정의 explainability + audit log | (본 skill stage 5) |

### Stage 2: Region cluster trigger 결정

`decide-target-market` 산출에 따라 분기:

| Region 결과 | trigger out |
|----------|-----------|
| 글로벌 default | 본 skill 만 (region-agnostic) |
| Korea | + `consult-korea-legal-context` (deferred — Korea cluster) |
| USA | + (deferred-deferred — template 만, trigger 시 신규 작성) |
| EU | + (deferred-deferred — GDPR 등 신규 작성) |

### Stage 3: Sub-skill 호출 (3 기존 buddy skill)

위 7 sub-domain 의 첫 3 (Privacy / IP / AI 책임) 은 *기존 buddy skill* 로 분기:

```
review-legal-regulatory
├── (cascade) review-privacy-data-risk (구현됨)
├── (cascade) review-license-and-ip-risk (구현됨)
└── (cascade) review-ai-safety-liability (구현됨)
```

세 sub-skill 산출 통합 → 본 skill 의 *Privacy / IP / AI 책임* 검토 결과.

### Stage 4: Inline 검토 (약관 / 결제)

위 sub-skill 외 *약관 / 결제* 영역은 본 skill 안에서 inline 검토:

#### 4.1 약관 / 개인정보처리방침 readiness

체크리스트:
- [ ] 약관 / 개인정보처리방침 *작성됐는가*?
- [ ] 사용자 동의 흐름 (signup) 에 *명시적 체크박스* 있는가? (silent opt-in 금지)
- [ ] 변경 시 *공지 절차* (이메일 / 앱 푸시 / 14일 전) 준비됐는가?
- [ ] 약관 *언어* 가 사용자 locale 정합한가?
- [ ] 약관에 *분쟁 해결 / 관할 법원* 명시?

#### 4.2 결제 / 환불 / 소비자보호

체크리스트:
- [ ] 결제 PG 가 *지역 법령 정합* 한가? (e.g. PCI-DSS for card payment)
- [ ] 환불 규정이 *명시* 됐는가? (며칠 / 부분 환불 / 디지털 재화 특성)
- [ ] 분쟁 해결 *연락처 / 절차* 명시?
- [ ] 미성년자 결제 / 보호자 동의 절차?

### Stage 5: 산업별 규제 + AI 책임 / 감사

산업 명시 시:

| 산업 | 규제 후보 (글로벌 / region 무관) |
|------|----------------------|
| Medical | HIPAA (USA) / 의료법 / 의료기기법 |
| Financial | PCI-DSS / KYC / AML |
| Education | FERPA / 학생 데이터 보호 |
| Children data | COPPA (USA) / 13세 미만 보호 |

AI 책임:
- [ ] AI 출력의 *정확성 보장 X* 명시 (사용자 사전 고지)
- [ ] AI 결정의 *audit log* 보관?
- [ ] *explainability* 요구 영역 (의료 / 금융) 시 추가 검토

### Stage 6: Evidence package 작성

전문가 검토 의뢰 시 입력 자료:

```markdown
## Evidence Package — {제품}
- 데이터 흐름 도표
- 사용자 데이터 주체 분류 (`map-customer-segments`)
- 7 sub-domain 체크리스트 결과
- region cluster 활성화 결과
- 산업 규제 식별
- AI 책임 검토 결과
- Open question (전문가 검토 필요 항목)
```

## 5. 산출물 형식

```markdown
## Legal/Regulatory Review — {제품}

### 7 sub-domain 결과
| sub-domain | 상태 | sub-skill 결과 |
|---------|-----|--------------|

### Region cluster trigger
- 결과: {global / Korea / USA / EU}
- 활성화 cluster: {none / Korea-3 / template-only}

### 약관 / 결제 inline 검토
- 약관: {pass / gap}
- 결제: {pass / gap}

### 산업별 규제 (해당 시)
- 산업: {medical / financial / ...}
- 규제: {list}

### AI 책임
- audit log: {yes / no}
- explainability: {required / not}

### Open questions (전문가 의뢰)
- ...
```

## 6. 검증

- [ ] 7 sub-domain 모두 검토?
- [ ] Region cluster trigger 명시 (silent activation 금지)?
- [ ] 3 기존 buddy sub-skill (privacy / IP / AI safety) cascade?
- [ ] 약관 / 결제 *체크리스트* 모두 통과 또는 gap 명시?
- [ ] 산업 명시 시 산업별 규제 검토?
- [ ] Evidence package 작성 (전문가 의뢰 입력)?

## 7. 다음 phase

- `prepare-launch-checklist` (§7) 의 *법률 readiness* 항목 입력
- Region cluster trigger 결과:
  - 글로벌 → 다음 phase 진행
  - Korea → `consult-korea-legal-context` (deferred — Korea 진출 시 신규 작성)
  - 다른 region → template 만 보존, trigger 시 신규 skill 작성

## 8. 참조

- 2026-05-10 D-G 정정 lock-in: region-cross-cutting 3 layer framework — Layer 1 (region 결정 = `decide-target-market`) / Layer 2 (region-agnostic core, 본 skill) / Layer 3 (region-specific extension = Korea cluster 3건 deferred / USA·EU 0건). 원본 inventory note 는 commit `9ab21dd` 로 제거, 결정 자체는 유지.
- 기존 buddy `review-privacy-data-risk` / `review-license-and-ip-risk` / `review-ai-safety-liability` (sub-skill)
- (deferred) `consult-korea-legal-context` (Korea cluster, trigger 시 활성화)
- 본 skill 은 *법률 자문 대체 X*. 항상 *전문가 검토 의뢰* 가 최종 step.
