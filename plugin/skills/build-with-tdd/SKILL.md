---
name: build-with-tdd
description: "신규 기능을 red-green-refactor TDD 루프(test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터)로 구현. tracer bullet 우선, observable behavior 중심. 트리거: 'TDD로 가자' / '테스트 먼저 작성' / 'red-green-refactor' / 'tracer bullet' / 'failing test 먼저' / 'TDD 사이클' / 'behavior test'. 입력: feature spec 또는 acceptance criteria. 출력: passing test suite + 최소 구현 코드. 흐름: split-work-into-features/define-feature-spec → build-with-tdd → measure-code-health/run-browser-qa."
type: skill
---

# Build With TDD — Red-Green-Refactor 강제 루프

## 1. 목적

신규 기능을 **테스트 먼저 작성 → 실패 확인 → 최소 구현 → 통과 → 리팩터** 루프로 구현한다. iterate-fix-verify가 일반 fix 루프라면, build-with-tdd는 **strict TDD discipline**을 강제한다 — tracer bullet, 한 번에 하나의 behavior test, observable behavior 중심.

이 스킬은 Stage 6 (Development) 단계에 위치하며, split-work-into-features로 슬라이스 단위가 정의된 직후 또는 define-product-spec의 acceptance criteria가 확정된 직후에 호출된다. 결과물은 단순한 "코드"가 아니라 **passing test suite + 최소 구현 + TDD 사이클 로그**이다.

핵심 차별점:
- iterate-fix-verify는 "동작하지 않는 코드를 동작하게 만든다"이고
- build-with-tdd는 "처음부터 검증 가능한 형태로만 코드를 작성한다"이다

따라서 두 스킬은 페어로 동작하지만, build-with-tdd가 먼저 적용되면 fix 루프 진입 빈도 자체가 줄어든다.

## 2. 사용 시점

다음 상황에서 호출하라:

- 신규 feature 구현 시작 (split-work-into-features 통과 후)
- bug 재발 방지 (regression test 먼저 작성)
- 복잡한 비즈니스 로직 (가격 계산, 권한 검증, state machine, 할인 룰)
- API endpoint 신규 추가 (request/response contract 고정)
- legacy 코드 리팩터 전 안전망 구축 (characterization test 작성)
- contract test 필요 시 (consumer-driven, provider-driven)
- 외부 의존성 추상화 (DB, HTTP client, message queue) 시 mock 경계 정의
- 동시성/race condition 영역 (lock, transaction, retry 로직)

다음 상황에서는 **호출하지 마라**:

- 1회용 스크립트 또는 throwaway prototype (오버엔지니어링)
- UI tweak 수준의 시각적 변경 (visual regression test가 적합)
- 데이터 마이그레이션 1회성 작업 (dry-run + checksum이 적합)
- spike (학습 목적의 탐색 — 끝나면 버린다)

## 3. 입력

### 필수 입력

- feature spec 또는 acceptance criteria (split-work-into-features 출력)
- target observable behavior (입력 → 출력 매핑이 명시된 형태)
- 기존 test infrastructure 정보 (jest/vitest/pytest/go test/rspec 등)
- runtime/언어 정보 (Node.js LTS, Python 3.x, Go 1.x 등)

### 선택 입력

- 기존 코드 (refactor 대상 — characterization test 필요 여부 판단)
- bug repro (regression test 작성 시 — 최소 재현 시나리오)
- coverage 목표 (line/branch/condition/mutation 별)
- 외부 의존성 목록 (mock 경계 결정용)

### 입력 부족 시 forcing question

다음 중 하나라도 답할 수 없으면 호출자에게 되묻는다:

- "이 기능의 observable behavior가 뭐야? 입력 → 기대 출력으로 표현해줘."
- "acceptance criteria가 'X일 때 Y한다' 형식이야? 아니면 'A를 만들어야 한다' 같은 모호한 형태야?"
- "test framework 뭐 써? jest/vitest/pytest/go test/rspec 중에?"
- "이걸 100% 자동으로 검증할 수 있는 시나리오야 manual 단계가 필요해?"
- "외부 시스템(DB, HTTP API, 메시지 큐) 접점이 있어? 있다면 어디까지 mock하고 어디부터 integration으로 갈 거야?"
- "기존 코드를 수정하는 거야 신규 모듈이야? 기존이면 characterization test 먼저 필요해."

답이 모호하면 build-with-tdd를 시작하지 말고 define-product-spec 또는 split-work-into-features로 회귀시킨다.

## 4. 핵심 원칙 (Principles)

1. **Test 먼저, 항상** — implementation 한 줄도 쓰기 전에 failing test부터. 사이클 끝마다 "test가 fail하는 걸 직접 봤어?" 확인.
2. **한 번에 하나의 behavior** — 한 test에 한 assertion 또는 한 시나리오. 여러 behavior 묶지 말 것. 여러 behavior가 한 test에 들어 있으면 fail 메시지로 어느 behavior가 깨졌는지 즉시 알 수 없다.
3. **Observable behavior 중심** — return value, side effect (DB write, HTTP call), state change. private method/internal data structure 테스트 금지. 내부를 테스트하면 리팩터링 시 test도 함께 깨져 안전망이 무력화된다.
4. **Tracer bullet 먼저** — 가장 간단한 end-to-end happy path 1개를 먼저 동작시킨 다음, 그 위에 edge case 추가. depth 우선이 아니라 breadth 우선. 이래야 통합 시점이 마지막에 몰리지 않는다.
5. **Red 단계 명시 확인** — test가 실패하는 이유를 봐야 test가 valid함을 안다. green-only로 시작 금지. "어차피 fail할 거니까"는 절대 금기.
6. **Green은 최소 구현** — test 통과시키는 만큼만. "더 좋게" 만들기 시도 금지 (refactor 단계로). hard-coded value도 일시적으로 OK.
7. **Refactor는 test 통과 후만** — green 상태에서만 코드 개선. red 상태에서 refactor 금지. red + refactor 동시 진행은 디버깅 지옥의 입구다.
8. **Implementation detail 금지** — `expect(cache.size).toBe(3)` 대신 `expect(getUser('id')).toBe(...)` — 외부 행동만. 내부 자료구조는 언제든 바뀔 수 있다.
9. **Test도 코드다** — DRY/명명/구조 규칙 동일하게 적용. test 코드 품질이 떨어지면 production 코드 품질도 따라 떨어진다.
10. **Mock은 경계에서만** — DB driver, HTTP client, system clock 같은 IO 경계에서만 mock. 비즈니스 로직 mock은 안티패턴.

## 5. 단계 (Phases)

### Phase 1. Behavior Mapping

acceptance criteria를 behavior 목록으로 변환한다. 각 behavior는 (입력, 기대 출력) 쌍으로 표현해야 하며, 모호한 자연어 요구사항은 이 단계에서 검증 가능한 형태로 변환된다.

예시 (로그인 기능):

```
behavior_1: 올바른 email/password로 로그인 → 200 + session token
behavior_2: 잘못된 password → 401 + generic error message
behavior_3: 존재하지 않는 email → 401 + generic error message (계정 존재 노출 금지)
behavior_4: rate limit 초과 → 429 + retry-after header
behavior_5: 잠긴 계정으로 로그인 시도 → 423 Locked + admin contact info
behavior_6: 정상 로그인 시 last_login_at 컬럼 갱신
```

각 behavior는 독립적으로 테스트 가능해야 한다. 의존성이 있다면 behavior 분할이 잘못된 것 — 더 작게 쪼개라.

### Phase 2. Tracer Bullet 선정

모든 behavior 중 가장 간단한 happy path 1개를 선택한다. 이 하나로 end-to-end pipeline (입력 → 처리 → 출력)이 동작하는지 확인한다. 위 예시에서 tracer = behavior_1.

tracer bullet 선정 기준:
- 외부 의존성 최소
- 가장 짧은 코드 경로
- 시스템의 "주축"을 관통하는 흐름 (HTTP → 비즈니스 로직 → DB → response)

이 단계가 끝나면 backbone이 살아 있고, 이후 모든 behavior는 이 backbone에 살을 붙이는 형태가 된다.

### Phase 3. Red — Failing Test 작성

tracer behavior에 대한 test 코드를 작성하고 실행한다. 결과는 반드시 **실패**여야 한다. 실패 메시지가 expected 한지 확인한다 (예: "function not implemented" 또는 "module not found" 또는 "expected 200 received undefined").

red 단계 체크리스트:
- test 이름이 behavior를 그대로 묘사하는가? (`it('returns 200 with token when credentials are valid')`)
- assertion이 외부 행동만 검증하는가?
- fail 메시지를 보고 무엇을 구현해야 하는지 명확한가?
- test가 다른 test와 독립적인가? (실행 순서에 무관)

### Phase 4. Green — 최소 구현

test를 통과시키는 **최소 코드**를 작성한다. hard-coded return value도 OK. 다음 test가 강제로 진짜 로직을 끌어낼 것이다. "더 좋게" 만들기 절대 금지.

예시 (behavior_1만 통과시키는 최소 구현):

```javascript
// 이 한 줄로 첫 test가 통과한다면 OK
function login(email, password) {
  return { status: 200, token: 'fake-token' };
}
```

이게 절대 production-ready가 아닌 건 명백하다. 하지만 그건 다음 behavior가 강제할 일이다. 지금은 green 상태를 만드는 것이 목적.

### Phase 5. Test 추가 (다음 behavior)

behavior_2 test를 추가한다. 기존 hard-coded 구현으론 통과할 수 없으므로 다시 red. 진짜 로직을 추가하여 green으로 전환. 이 과정을 통해 가짜 구현이 자연스럽게 진짜 로직으로 evolve한다.

예:

```javascript
// behavior_1 + behavior_2 둘 다 통과시키려면 hard-coded는 깨진다
function login(email, password) {
  const user = db.findUserByEmail(email);
  if (!user || !verifyPassword(password, user.passwordHash)) {
    return { status: 401, error: 'Invalid credentials' };
  }
  return { status: 200, token: createToken(user) };
}
```

### Phase 6. Refactor

green 상태에서 코드 개선:
- duplication 제거 (test 코드도 포함)
- 변수/함수 명명 개선
- abstraction 추출 (extract function/class)
- 복잡도 분리 (early return, guard clause)

각 refactor 후 test 다시 실행 → green 유지 확인. red가 한 번이라도 뜨면 즉시 직전 상태로 rollback.

refactor 금지 사항:
- 새 behavior 추가 (그건 다음 red 단계)
- test 파일 삭제 (test가 잘못됐다면 별도 사이클로 수정)
- mock 경계 변경 (별도 사이클로)

### Phase 7. Edge Case 확장

behavior_3 ~ N을 하나씩 추가한다. 각 behavior마다 red → green → refactor 한 사이클. 절대로 한 번에 여러 behavior를 묶지 않는다.

순서 권장:
1. happy path variants (다른 valid input)
2. expected error path (잘못된 입력에 대한 명시적 에러)
3. boundary condition (off-by-one, empty input, max input size)
4. concurrency/race (해당 시)
5. external failure (DB down, timeout, network partition — 해당 시)

### Phase 8. Coverage 검증

다음을 모두 확인한다:

- 모든 acceptance criteria가 test로 커버되었는가?
- branch coverage 80%+ (line coverage만으론 부족)
- mutation testing 결과 (선택, 깊이 검증) — Stryker, mutmut 등
- contract test (외부 API 의존 시) 통과
- integration test 별도 1개 이상 (DB 포함 end-to-end)

coverage가 목표 미달이면 누락된 behavior를 식별하고 Phase 5로 회귀.

## 6. 출력 템플릿

다음 yaml 구조로 결과를 호출자에게 반환한다:

```yaml
feature_id: "<id>"
behaviors:
  - id: behavior_1
    description: "<입력 → 기대 출력>"
    test_file: "tests/auth-login.spec.ts:42"
    status: red | green | refactored
    notes: "<...>"
  - id: behavior_2
    description: "<...>"
    test_file: "tests/auth-login.spec.ts:67"
    status: refactored
    notes: "<...>"

tracer_bullet: behavior_1

tdd_log:
  - timestamp: "<ISO8601>"
    phase: red | green | refactor
    behavior: behavior_1
    action: "<무엇을 했는가>"
    test_result: pass | fail | skip
    coverage_delta: +X%
  - timestamp: "<ISO8601>"
    phase: green
    behavior: behavior_1
    action: "최소 구현 추가"
    test_result: pass
    coverage_delta: +12%

implementation_files:
  - path: "src/auth/login.ts"
    lines_added: 47
    lines_removed: 0
  - path: "src/auth/token.ts"
    lines_added: 18
    lines_removed: 0

test_files:
  - path: "tests/auth-login.spec.ts"
    test_count: 6
    lines: 124
  - path: "tests/auth-login.integration.spec.ts"
    test_count: 1
    lines: 42

coverage:
  lines: 94%
  branches: 87%
  conditions: 82%
  acceptance_criteria_coverage: 6/6
  mutation_score: 76% (optional)

anti_pattern_violations:
  - violation: "behavior_4 test에서 internal cache.size 검증"
    remediation: "외부 행동(getUser return value)으로 변경"
    fixed: true

next_steps:
  - "measure-code-health 호출하여 cyclomatic complexity 검증"
  - "run-browser-qa 호출하여 E2E 시나리오 검증 (해당 시)"
```

이 출력은 반드시 호출자(상위 오케스트레이터 또는 다음 스킬)가 그대로 소비할 수 있는 형태여야 한다. 자유 서술이 아닌 structured data.

## 7. 자매 스킬

build-with-tdd는 다음 스킬들과 명시적으로 연결된다.

### 앞 단계 (선행 스킬)

- `split-work-into-features` — feature를 슬라이스 단위로 쪼갠 결과를 입력으로 받음. `Skill` tool로 invoke.
- `define-product-spec` — acceptance criteria가 정의된 spec을 입력으로 받음. spec이 모호하면 build-with-tdd 시작 전에 회귀.
- `route-spec-to-code` — spec을 어떤 코드 위치에 매핑할지 결정. 큰 모놀리식에서는 이 스킬이 먼저 호출되어 파일/모듈 경계를 정함.

### 페어 (동시 동작)

- `iterate-fix-verify` — TDD 사이클 중 green이 되지 않을 때 (예상치 못한 fail) fix 루프로 위임. build-with-tdd가 strict discipline을 강제하지만, 디버깅 자체는 iterate-fix-verify의 책임.
- `freeze-edit-scope` — refactor 단계에서 scope creep 방지. "이거 김에 저것도 고치자"를 차단.

### 후속 단계 (다음 스킬)

- `measure-code-health` — 구현 후 cyclomatic complexity, duplication, naming 등 정량 품질 검증.
- `run-browser-qa` — UI 포함 시 E2E 시나리오 검증. 단위/통합 테스트 외에 사용자 관점 검증.
- `audit-security` — 인증/권한/입력 검증 영역이면 보안 리뷰 필수.
- `review-engineering` — 구현이 architecture 원칙을 따르는지 시니어 리뷰.

### 호출 흐름 예시

```
split-work-into-features
    → (slice 1)
    → build-with-tdd
        → (red-green-refactor cycle x N)
        → (필요 시) iterate-fix-verify
        → measure-code-health
        → (UI면) run-browser-qa
        → (보안 영역) audit-security
    → (slice 2)
    → build-with-tdd ...
```

## 8. Anti-patterns

다음은 build-with-tdd 적용 중 자주 나타나는 안티패턴과 교정 방법이다.

1. **Test를 implementation 후에 작성** — "코드 짜고 테스트 추가"는 TDD가 아니라 단순 retrofit testing. test가 먼저 작성되어 production 코드의 형태를 이끌어야 한다. 교정: 다음 behavior부터는 test 먼저 commit.

2. **Red 단계 skip** — "어차피 fail할 거니까 바로 green으로". fail 메시지가 expected한지 확인하지 않으면, test가 잘못 작성되어 영원히 green인 상태(false positive)를 잡을 수 없다. 교정: 매 사이클마다 의도적으로 한 번은 fail 출력을 본다.

3. **Green에서 "더 잘 만들기"** — green 상태에서 추가 abstraction, 추가 함수 분리. refactor 단계로 미뤄야 한다. green = 통과만. refactor = 개선. 두 단계를 섞으면 사이클이 뭉개진다. 교정: green 단계 종료 시 commit하고, refactor 단계는 별도 commit.

4. **여러 behavior 한 test에 묶기** — `it('handles login', () => { ... 5개 시나리오 ... })`. fail 메시지로 어느 behavior가 깨졌는지 즉시 알 수 없다. 교정: 한 `it` 블록 = 한 behavior.

5. **Implementation detail 테스트** — `expect(component.state.cache.size).toBe(3)` 같이 내부 상태 노출. 리팩터 시 test도 함께 깨져 안전망 무력화. 교정: 외부에서 관찰 가능한 결과(return value, side effect, rendered output)로 assertion.

6. **Mock everything** — DB, HTTP, FS 전부 mock하면 통합 시점에 schema migration breakage, contract drift, serialization issue가 한꺼번에 터진다. 교정: unit test는 mock으로, 별도 integration test 1개 이상에서 실제 DB/HTTP 사용.

7. **100% line coverage 목표** — line coverage 100%여도 branch coverage 50%, mutation score 30%면 사실상 약한 안전망. 교정: branch + condition + mutation까지 보고, 임계값을 mutation score 70%+로 잡는다.

8. **Tracer bullet 없이 depth 우선** — auth backend를 끝까지 다 만들고 frontend 통합 → 통합 단계에서 schema 불일치, CORS 누락, 인증 헤더 형식 차이가 한꺼번에 발견된다. 교정: 가장 얇은 end-to-end 경로 1개를 먼저 통과시키고, 그 위에 두께를 더한다.

9. **Test 코드 품질 방치** — production 코드는 깔끔한데 test는 copy-paste 지옥. 시간이 지나면 test 유지보수 비용이 폭발한다. 교정: test도 동일한 코드 품질 기준 적용 (DRY, 명명, 구조).

10. **Refactor 단계에서 새 behavior 추가** — "리팩터링 김에 이것도 고치자". refactor의 정의는 "외부 행동 불변, 내부 구조 개선"이다. 새 behavior는 새 red-green 사이클로. 교정: refactor 중에는 test 추가 금지, behavior 변경 금지.

11. **System under test 외부 시스템에 의존** — test가 인터넷, 외부 API, 시스템 시간(`Date.now()`), 랜덤 값에 의존하면 flakiness 폭발. 교정: clock injection, 시드 고정, fake server 또는 contract test로 격리.

12. **Test 이름이 implementation 묘사** — `it('calls dbService.query with userId')`처럼 내부 호출을 묘사. behavior 묘사로 변경: `it('returns user profile when user exists')`.

## 9. 체크리스트 (사이클별 자가 점검)

각 red-green-refactor 사이클 종료 시 다음을 점검한다:

- [ ] 이 사이클에서 한 번이라도 fail 메시지를 직접 보았는가?
- [ ] test 이름이 behavior(외부 행동)을 묘사하는가?
- [ ] assertion이 internal state가 아닌 observable result를 검증하는가?
- [ ] green 단계 commit과 refactor 단계 commit이 분리되어 있는가?
- [ ] 다른 test와 독립적으로 실행 가능한가? (순서 무관, 격리됨)
- [ ] 외부 의존성이 적절한 경계에서 mock되었는가? (비즈니스 로직 mock 금지)
- [ ] 이번 사이클로 coverage가 의미 있게 증가했는가? (단순 line이 아닌 branch)
- [ ] tracer bullet이 여전히 동작하는가? (regression 없음)

8개 항목 중 하나라도 No이면 다음 사이클로 진행하지 말고 직전 상태를 교정한다.

## 10. 종료 조건

build-with-tdd 호출이 종료되는 조건:

- 모든 acceptance criteria가 passing test로 커버됨
- branch coverage 목표 (기본 80%, 보안 영역 90%) 달성
- mutation score 목표 달성 (기본 70%)
- anti-pattern 위반 0건 (또는 모두 remediation 완료)
- tracer bullet이 backbone으로서 여전히 동작
- 출력 yaml이 호출자에게 반환됨

종료 후 호출자는 measure-code-health, run-browser-qa, audit-security 등 후속 스킬로 자동 진행할 수 있다.
