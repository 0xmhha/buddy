---
description: locale + fallback chain + ICU MessageFormat + RTL 지원 + 번역 워크플로우 (extract / memory / review / deploy) + locale 변형 5 영역 (날짜/통화/legal/문화/onboarding).
argument-hint: "<제품 이름 또는 target locale list>"
disable-model-invocation: true
---

# /buddy:design-i18n-strategy

decide-target-market 결과 다지역 시 i18n 전략. ICU MessageFormat + RTL + 번역 워크플로우 + 번역 외 5 변형. audit-i18n-coverage (§6) 입력.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-i18n-strategy`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
