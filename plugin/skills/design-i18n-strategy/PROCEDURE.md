# Design i18n Strategy — locale 분기 + ICU MessageFormat + RTL + 번역 워크플로우

## 1. 목적

다국어 (internationalization, i18n) 전략 설계. **locale 식별 / message format / RTL 지원 / 번역 워크플로우 / fallback** 5 영역 통합.

`decide-target-market` 산출이 *다지역* 일 때 본 skill 우선. 글로벌 default (영어 only) 면 *최소 i18n* — 단 *향후 확장 친화* 양식 채택 권장.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Target market / 지원 언어 | ✅ | knowledge | `decide-target-market` 산출물 또는 사용자 설명 | "지원할 언어와 지역은?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| i18n 전략 (locale 구조 + fallback + RTL + format) | artifact | structured YAML | `audit-i18n-coverage`, `build-feature` |

## 2. 사용 시점

- §3 design-system 안에서 — `decide-target-market` 산출 *Korea / 다지역* 시
- `define-tech-stack` 결정 후 — 프레임워크 의 i18n 라이브러리 결정
- 새 locale 추가 시 (예: 한국어 → +일본어) — 기존 양식 유지 검증
- `audit-i18n-coverage` (§6) 가 *번역 누락* 보고 시 — 양식 재검토

## 3. 입력

### 필수
- `decide-target-market` 산출 — primary / secondary locale list
- `define-tech-stack` — frontend / backend 프레임워크
- `map-customer-segments` — locale 별 segment

### 선택
- 기존 번역 자산 (있으면 — migration)
- RTL 시장 진출 가설 (아랍어 / 히브리어)

## 4. Stage 흐름

### Stage 1: Locale 식별 + fallback chain

| locale | priority | fallback |
|--------|---------|---------|
| en (default) | 1 | (없음 — 모든 key cover) |
| ko | 2 | en |
| ja | 3 | en |
| ar (RTL) | 4 | en + RTL flag |

→ fallback chain *명시* — silent fallback 금지.

### Stage 2: Message format — ICU MessageFormat

단순 key-value 가 아닌 ICU MessageFormat 사용 (복수형 / 성별 / 날짜 / 숫자 처리):

```
{count, plural, one {1 item} other {# items}}
{gender, select, male {He} female {She} other {They}}
```

각 message:
- key: namespaced (`auth.signup.success`)
- args: typed (string / number / date)
- 컨텍스트 주석 (번역가용)

### Stage 3: RTL (Right-to-Left) 지원

RTL locale (ar / he / fa / ur) 진출 시:
- CSS `dir="rtl"` 자동 적용
- layout flip (margin-left → margin-inline-start)
- icon mirror (← → 화살표 등)
- 단위 / 시간 format locale 적용

### Stage 4: 번역 워크플로우

| 단계 | 도구 후보 |
|------|--------|
| Source extract | i18next-parser / Lingui / babel-plugin-i18n |
| Translation memory | Crowdin / Phrase / Lokalise / 자체 git |
| Review | bilingual reviewer + 도메인 전문가 |
| Deploy | CI / CD pipeline 안에서 build artifact 포함 |

→ AI 번역 + human review 조합 권장 (cost / speed / quality 균형).

### Stage 5: Locale 별 변형 (translation 너머)

번역만으로 부족한 영역:

| 영역 | locale 변형 |
|------|---------|
| 날짜 / 숫자 / 통화 | locale formatter (Intl.DateTimeFormat) |
| 결제 통화 | 지역 PG + 환율 |
| legal 문구 | region cluster (review-legal-regulatory + extension) |
| 이미지 / 색 | 문화 적합성 (e.g. 색의 의미 다름) |
| onboarding | 지역 사용자 *기대 흐름* |

### Stage 6: 정합 검증 — `audit-i18n-coverage` (§6) 입력

본 skill 산출은 *전략*. 실제 *번역 누락 검증* 은 `audit-i18n-coverage` skill 가 담당. 양 skill 의 *형식 정합* 보장.

## 5. 산출물 형식

```markdown
## i18n Strategy — {제품}

### Locale + fallback
| locale | priority | fallback |

### Message format
- 양식: ICU MessageFormat
- key namespace 컨벤션: ...
- 도구: ...

### RTL 지원
- 활성 locale: ...
- layout flip 도구: ...

### 번역 워크플로우
- extract → memory → review → deploy

### Locale 변형 (번역 외)
| 영역 | 변형 |
```

## 6. 검증

- [ ] Locale list + fallback chain 명시?
- [ ] ICU MessageFormat 채택 (key-value 만 X)?
- [ ] RTL locale 진출 가설 시 RTL 지원 명시?
- [ ] 번역 워크플로우 4 단계 (extract / memory / review / deploy)?
- [ ] Locale 변형 5 영역 (날짜 / 통화 / legal / 문화 / onboarding) 명시?

## 7. 다음 phase

- `audit-i18n-coverage` (§6) 의 입력
- region cluster (Korea / USA / EU 등) 의 locale 변형 통합
- buddy 자체 `internal/persona/` (en / ko 카탈로그) 와 정합 검증

## 8. 참조

- ICU MessageFormat specification — Unicode CLDR
- W3C i18n best practices
- buddy `internal/persona/` (en/ko 카탈로그) — 자체 dogfood reference
