---
description: form factor 별 design system 채택 (shadcn/MUI/HIG/Material/RN Paper). 4 차원 평가 (component coverage / customization / a11y / 활성도) + token 5 종 + pattern library + adoption tracking 80%+.
argument-hint: "<제품 이름 또는 form factor>"
disable-model-invocation: true
---

# /buddy:apply-design-system

decide-form-factor 결과 + brand → design system 채택. token (color/spacing/typography/radius/shadow) 정합 + pattern library + adoption 80%+ 추적.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `apply-design-system`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
