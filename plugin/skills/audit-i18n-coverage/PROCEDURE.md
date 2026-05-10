# Audit i18n Coverage — locale 별 번역 누락 / fallback 누수 검출

## 1. 목적

`design-i18n-strategy` (§3) 에서 결정된 locale 매트릭스 + ICU MessageFormat 정합 검증. **번역 누락 + fallback 누수 + format 오류** 3 영역 audit.

`audit-accessibility` (§6, 구현됨) 와 같은 *production audit* 영역 — 사전 baseline 이 아니라 *현재 적용 결과* 검증.

## 2. 사용 시점

- §6 verify-quality 의 release gate 전
- 새 locale 추가 후 — 신규 locale 의 *완전성* 검증
- 번역 batch 수신 후 — 번역가 산출 정합
- 사용자 *번역 누락 / 어색한 표현* 보고 시
- 정기 (분기) audit — drift 누적 차단

## 3. 입력

### 필수
- `design-i18n-strategy` 산출 — locale list + fallback chain + 양식
- 번역 자산 위치 (예: `locales/en.json`, `locales/ko.json`, ICU MessageFormat 파일)
- 코드의 *번역 호출 site* — `t("key", args)` 패턴

### 선택
- 사용자 피드백 (어색한 번역 / 누락 보고)
- locale 별 사용자 수 (priority weight)

## 4. Stage 흐름

### Stage 1: 번역 key coverage

각 locale 의 *모든 source key* 채워졌는지:

```bash
# 예: en 이 source, ko / ja / ar 검증
diff <(jq -r 'keys[]' locales/en.json | sort) \
     <(jq -r 'keys[]' locales/ko.json | sort)
```

→ 누락 key list 산출.

### Stage 2: Fallback 누수 검출

source locale 외 *fallback 으로만* 도달하는 key 찾기:

| 상황 | 의미 |
|------|------|
| ko key 없음 + en 존재 → en 표시 | "fallback 누수" — 사용자가 한국어 모드인데 영어 출력 |
| ko key empty string | empty 표시 — production bug |
| ko key 있으나 *번역 안 됨* (영어 그대로) | "translation skip" — 번역가 누락 |

→ 각 locale 의 fallback rate 측정. > 5% 이면 *번역 미완성* 신호.

### Stage 3: ICU MessageFormat 정합

각 key 의 *plural / select / 변수* 정합:

```
en: "{count, plural, one {1 item} other {# items}}"
ko: "{count, plural, other {# 항목}}"  // ✓ 한국어는 단수/복수 X
ja: "{count} items"  // ✗ ICU 미사용 — fallback 안 됨
```

→ ICU 양식 자동 검증 (formatjs / @formatjs/cli) + missing variable / type mismatch 식별.

### Stage 4: format / locale-specific 검증

| 영역 | 검증 |
|------|------|
| 날짜 | `Intl.DateTimeFormat(locale)` 사용? hardcoded format X |
| 숫자 / 통화 | locale 별 separator (1,234 vs 1.234) |
| RTL | `dir="rtl"` 활성? layout flip 동작? |
| 단위 | km vs miles / kg vs lbs / 시간 24h vs 12h |

### Stage 5: Coverage matrix 산출

| locale | key total | filled | empty | fallback rate | ICU 오류 | priority |
|--------|----------|--------|-------|------------|--------|--------|
| en (source) | 1000 | 1000 | 0 | — | 0 | (source) |
| ko | 1000 | 950 | 30 | 2.0% | 5 | High |
| ja | 1000 | 800 | 0 | 20.0% | 12 | Medium |
| ar | 1000 | 1000 | 0 | 0% | 2 (RTL not applied) | Low |

→ priority 별 fix 우선순위 결정.

### Stage 6: CI 자동화

- pre-commit hook: 새 key 추가 시 *모든 locale empty 안 됨* 강제
- CI: locale matrix 검증 + threshold (fallback rate > 5% 시 fail)
- weekly report: drift 추적

## 5. 산출물 형식

```markdown
## i18n Coverage Audit — {제품}

### Coverage matrix
| locale | total | filled | empty | fallback% | ICU 오류 |

### 누락 key list (per locale)
- ko: {list}
- ja: {list}

### Fallback 누수 영역
- {key}: ko 비어 있고 en 사용

### ICU 오류
- {key} ko: ICU 양식 부정합

### Format / locale-specific gap
- 날짜 hardcoded: {file:line}
- RTL 미적용: {component}

### Priority fix 순서
1. ...
```

## 6. 검증

- [ ] 모든 locale 의 key coverage 측정?
- [ ] Fallback rate threshold (5%) 검증?
- [ ] ICU MessageFormat 자동 검증?
- [ ] 날짜 / 숫자 / 통화 / RTL format 점검?
- [ ] Priority 별 fix 순서 명시?
- [ ] CI 자동화 (pre-commit + matrix 검증)?

## 7. 다음 phase

- `design-i18n-strategy` 의 양식 보강 (drift 패턴 발견 시)
- 번역가 hand-off — 누락 key list
- `prepare-launch-checklist` 의 i18n gate 입력

## 8. 참조

- ICU MessageFormat specification
- formatjs (@formatjs/cli) — 자동 검증 도구
- buddy `design-i18n-strategy` (구현됨) — 입력
- buddy `audit-accessibility` (§6, 구현됨) — audit 양식 reference
