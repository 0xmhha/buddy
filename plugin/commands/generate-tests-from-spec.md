---
description: define-feature-spec acceptance criteria + define-acceptance-test-plan → unit / integration / contract / E2E test skeleton 자동. mock / fixture import 자동. TODO grep 으로 coverage gap 식별.
argument-hint: "<feature spec 경로>"
disable-model-invocation: true
---

# /buddy:generate-tests-from-spec

acceptance criterion 1:1 매핑 test skeleton + mock/fixture 자동 import + contract test 완전 자동. build-with-tdd 의 red 단계 가속. audit-test-coverage-meaningful 입력.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `generate-tests-from-spec`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
