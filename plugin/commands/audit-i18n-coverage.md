---
description: locale 별 번역 누락 + fallback 누수 (rate > 5% 시 fail) + ICU MessageFormat 정합 + format / locale-specific (날짜 / 통화 / RTL) 검증. coverage matrix + priority fix 순서.
argument-hint: "<제품 이름 또는 locale 자산 경로>"
disable-model-invocation: true
---

# /buddy:audit-i18n-coverage

design-i18n-strategy 적용 결과 검증. key coverage + fallback rate + ICU 양식 + format 4 영역 audit. CI 자동화 (pre-commit + matrix). prepare-launch-checklist 의 i18n gate 입력.
