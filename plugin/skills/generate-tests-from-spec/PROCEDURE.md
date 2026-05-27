# Generate Tests From Spec — spec → unit / integration / contract test skeleton

## 1. 목적

`define-feature-spec` 의 acceptance criteria + `define-acceptance-test-plan` 의 test 매트릭스를 입력으로 **test skeleton 자동 생성**. 사용자가 *test logic* 만 채움 — boilerplate 0.

`build-with-tdd` 의 *red 단계* 보조 — failing test 작성을 가속.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Acceptance criteria 또는 test plan | ✅ | artifact / knowledge | `define-acceptance-test-plan` 산출물 또는 feature spec | "테스트를 생성할 acceptance criteria를 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Test skeleton (unit/integration/contract/E2E) | artifact | test files + TODO markers | `build-with-tdd` |

## 2. 사용 시점

- §5 build-feature 안 — feature 구현 전 test skeleton 준비
- 새 endpoint / function 추가 시 — *대응 test* 자동 생성
- contract test (API / event schema) 자동 동기화
- `audit-test-coverage-meaningful` 가 *coverage gap* 보고 시 — 누락 test skeleton 자동 생성

## 3. 입력

### 필수
- `define-feature-spec` 의 acceptance criteria
- `define-acceptance-test-plan` 의 test 매트릭스 (per-actor unit / integration / E2E)
- 언어 / 프레임워크 (Jest / pytest / Go testing / RSpec 등)

### 선택
- 기존 test fixture (있으면 — re-use)
- mock library (msw / nock / responses 등)

## 4. Stage 흐름

### Stage 1: Test 분류

| 분류 | 영역 | 도구 후보 |
|------|------|--------|
| Unit | 단일 function / class | Jest / pytest / Go testing |
| Integration | 다중 component / DB / 외부 API mock | testcontainers / msw / supertest |
| Contract | API / event schema 정합 | Pact / Spectral / asyncapi-validator |
| E2E | full user journey | Playwright / Cypress / Maestro (mobile) |

### Stage 2: Skeleton 자동 생성

각 acceptance criterion → test skeleton:

```typescript
// Generated from feature-spec/auth.yaml#signup
describe("auth/signup", () => {
  // Acceptance: 유효한 email + password 입력 시 201 + JWT 반환
  it("returns 201 with JWT on valid signup", async () => {
    // arrange
    // TODO: test logic
    // assert
  });

  // Acceptance: 중복 email 시 409
  it("returns 409 on duplicate email", async () => {
    // TODO: test logic
  });
});
```

→ *생성 시점에 test logic 0* — `// TODO` 만. 사용자가 *one test at a time* 작성.

### Stage 3: Mock / fixture 자동 import

generator 가 spec 의 *외부 의존* 식별 후 mock 자동 import:

```typescript
import { setupServer } from 'msw/node';
import { handlers } from './generated/mocks/auth';

const server = setupServer(...handlers);
```

→ contract 변경 시 mock 도 자동 재생성 (`generate-from-api-contract` 와 정합).

### Stage 4: Contract test 자동화

API / event contract → contract test 자동:

| 영역 | 검증 |
|------|------|
| OpenAPI | request / response schema 정합 |
| GraphQL | query / mutation type 정합 |
| Event schema | producer / consumer payload 정합 |

→ contract test 는 *완전 자동* — 사용자 logic 작성 X. 단 fail 시 *contract update* + generator re-run 으로 fix.

### Stage 5: Coverage gap 자동 식별

생성된 skeleton vs 실제 작성된 test 비교:
- skeleton 있는데 logic 없음 → "TODO: test logic" grep
- 새 acceptance criterion 추가 → 새 skeleton 자동 생성

→ `audit-test-coverage-meaningful` (§6) 의 입력.

## 5. 산출물 형식

```markdown
## Test Skeleton Generation — {feature}

### Test 분류
| 분류 | 도구 | 위치 |

### Skeleton 위치 (auto-generated)
- unit: ...
- integration: ...
- contract: ...

### Mock / fixture 자동 import
- 위치: generated/mocks/

### TODO 식별
- skeleton 있는데 logic 없는 test: {grep 결과}
```

## 6. 검증

- [ ] Test 분류 (unit / integration / contract / E2E) 모두 cover?
- [ ] Skeleton 자동 생성 (acceptance criterion 1:1 매핑)?
- [ ] Mock / fixture 자동 import?
- [ ] Contract test 완전 자동 (사용자 logic 0)?
- [ ] Coverage gap 자동 식별?

## 7. 다음 phase

- `build-with-tdd` 의 red 단계 — skeleton 의 TODO 채움
- `audit-test-coverage-meaningful` (§6) — skeleton vs 작성 비교
- `monitor-regressions` — contract drift 감지

## 8. 참조

- Pact (Contract Testing) — pact.io
- buddy `define-acceptance-test-plan` (구현됨) — test 매트릭스 입력
- buddy `build-with-tdd` (구현됨) — red 단계 cascade
