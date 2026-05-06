---
name: diagnose-bug
description: "버그를 증상 반응이 아닌 재현 가능한 원인 분석으로 해결. minimize repro → multiple hypothesis → targeted instrumentation → fix → regression test → original path 재검증. 트리거: '버그 디버깅' / '재현해줘' / 'minimal repro' / '왜 이런 현상?' / 'root cause 분석' / 'regression test 추가' / 'hypothesis 세워줘'. 입력: bug report, error log, 재현 시나리오. 출력: minimized repro + root cause + fix + regression test. 흐름: monitor-regressions/triage-work-items → diagnose-bug → iterate-fix-verify."
type: skill
---

# Diagnose Bug — 재현 기반 원인 분석 루프

## 1. 목적

증상에 반응해 임의 수정하지 않고, **재현 가능한 원인 분석**을 통해 버그를 해결한다. iterate-fix-verify가 일반 수정 루프라면, diagnose-bug는 **buggy state를 재현 → 가설 → 계측 → 수정 → 회귀 방지** 전용이다.

이 스킬의 본질은 "임의로 코드 만지지 마라, 먼저 재현하고 원인을 모르는 한 손대지 마라"이다. 재현 없는 fix는 가설일 뿐이며, 원인 없는 fix는 다른 곳에서 다시 터질 부채다.

핵심 차별점:
- iterate-fix-verify: test가 fail이면 통과시키는 일반 fix-verify 루프. 원인 분석은 부산물.
- diagnose-bug: 원인 분석이 **목적**. fix는 분석의 결과물. minimize → hypothesize → instrument를 강제.

## 2. 사용 시점

- production 버그 리포트 수신
- monitor-regressions에서 console error / page fail 감지
- intermittent failure (가끔 발생, flaky test)
- "내 환경에서는 동작해" 시나리오 (환경 의존 버그)
- 성능 회귀 (느려짐, 메모리 누수)
- security 인시던트의 root cause 분석
- post-mortem 작성
- legacy 코드의 origin 미상 버그
- third-party dependency 업그레이드 후 발생한 회귀

사용 부적합:
- 명확한 syntax error → 그냥 고치기
- 명세 변경에 따른 의도된 변경 → implement 스킬 사용
- 신규 기능 개발 → build-with-tdd 사용

## 3. 입력

필수:
- bug 증상 설명 (한 문장)
- 발생 환경 (browser, OS, version, locale, timezone)
- error log / stack trace (있으면)
- 재현 시도 결과 (재현 안 됨 / 가끔 재현 / 항상 재현)

선택:
- video / screenshot
- network HAR
- DB state dump (PII 마스킹)
- user account ID (privacy 주의)
- 최근 deploy / dependency 변경 로그
- feature flag state

입력 부족 시 forcing question:
- "재현 단계가 정확히 뭐야? 1, 2, 3 형식으로 적어줘."
- "100번 중 몇 번 발생해? 항상? 가끔? 한 번 봤어?"
- "예전엔 동작했어? 어느 commit / 어느 release부터 깨진 거 같아?"
- "다른 사용자도 발생해, 너만 발생해?"
- "browser dev tools console에 뭔가 있어? Network 탭 4xx/5xx 있어?"
- "feature flag / A/B test 관련 있을 수 있어?"
- "배포 직후부터 발생했어, 아니면 점진적으로?"

input 5개 이상 비어 있으면 분석 시작 거부 — "재현 정보 부족, Phase 1.5 우선 실행."

## 4. 핵심 원칙 (Principles)

1. **재현 없이 fix 없다** — 재현 안 되는 버그를 "고쳐줬다"는 가설일 뿐이다. 재현 → fix → fix 후 재현 안 됨 확인. 이 cycle 빠지면 fix가 아닌 추측.

2. **Minimize repro 강제** — 100줄 component에서 5줄로 줄이면 원인이 드러난다. 줄이지 못하면 원인을 모르는 것이다. minimal repro = 분석 도구.

3. **여러 hypothesis 동시 (3+)** — "이거다"는 confirmation bias. 3개 이상 가설 세우고 instrument로 구분. 1개 가설로 직진하면 잘못된 fix가 production에 들어간다.

4. **Targeted instrumentation** — println 도배 금지. 가설을 구분할 수 있는 specific log/breakpoint. 가설 A를 ruled-in/ruled-out 할 수 있는 신호여야 의미 있다.

5. **Fix는 root cause에** — 증상 가까이 fix 금지 (예: error catch + ignore, default value 우회). 원인 layer에서 수정. data layer 버그를 UI에서 가리지 마라.

6. **Regression test는 fix 전에** — fix 후엔 자연스럽게 통과해서 test가 valid한지 모른다. fix 전 작성 → fail 확인 → fix → pass. RED-GREEN cycle 강제.

7. **Original repro path 재검증** — fix가 다른 부분 깨뜨리지 않았는지 확인. 원래 사용자 시나리오 다시 실행. 인접 기능 sanity check.

8. **Workaround ≠ fix** — try-catch로 숨기거나 if-else로 우회는 fix 아니다. 원인 명시 + 추후 처리 ticket. mitigation으로 명시 (root_cause_fix: false).

9. **같은 원인의 다른 사례 검색** — 1곳 fix 후 grep으로 동일 패턴 검색. 같은 데이터 가정 / timing 가정이 다른 곳에 있을 가능성 높다.

10. **Postmortem 강제 (production)** — production 버그는 timeline + impact + prevention 작성. persist-learning-jsonl로 lesson 저장. 같은 실수 반복 방지.

## 5. 단계 (Phases)

### Phase 1. Reproduce

1. 사용자 보고 시나리오 그대로 재현 시도
2. 환경 일치 (browser version, OS, data state, locale, timezone)
3. 재현 빈도 측정 (10회 시도 중 몇 회 발생)
4. 재현되면 → Phase 2
5. 재현 안 되면 → Phase 1.5

재현 환경은 production과 가깝게:
- staging DB snapshot
- 동일 browser version (browserstack 등)
- locale / timezone 일치
- feature flag / A-B test bucket 일치

### Phase 1.5. Re-reproduce (재현 실패 시)

원인 후보:
- 환경 차이 (locale, timezone, network speed, data state)
- production data shape (특정 record 형태에서만 발생)
- 동시성 (load 있을 때만 발생)
- timing (사용자가 빠르게 클릭하는 패턴)
- cache (stale cache, CDN cache, browser cache)

조치:
- production data sanitized snapshot 확보
- log/trace에서 actual sequence 추출 (timestamp 정렬)
- bisect (어느 commit부터 깨졌는지 git bisect)
- canary 사용자 grant 후 reproduce 시도
- chaos 환경 (slow network 시뮬레이션, throttle CPU)

여전히 재현 실패 시:
- 추가 instrumentation production 배포 (sampled logging)
- 증상 발생 사용자에게 trace ID 요청
- 가설 기반 mitigation만 우선 배포 (root cause 미상 명시)

### Phase 2. Minimize Repro

목표: 분석 가능한 최소 단위로 축소. minimal repro = 가설 검증 도구.

전략:
1. 시나리오 단계 제거 → 여전히 발생?
2. 데이터 단순화 → 여전히 발생?
3. component 격리 (다른 부분 mock) → 여전히 발생?
4. 외부 의존성 stub → 여전히 발생?
5. 50줄+ → 5-10줄 정도로 줄이기 목표

축소가 안 되는 부분이 원인 영역. 더 줄일 수 없는 지점이 핵심 단서.

축소 결과물:
- 격리된 reproduction case (script / test / gist)
- 사라지는 변경 (positive control): 한 줄만 바꿔도 사라지면 그 줄이 원인 후보
- 살아남는 변경 (negative control): 다른 부분 바꿔도 그대로 발생하면 무관

### Phase 3. Hypothesize (3+ 가설)

가설 카테고리:
- **race condition** — async timing, promise 순서, debounce 누락
- **data state** — 특정 record / null / empty array / extreme value
- **environment** — browser version, locale, timezone, OS
- **third-party dependency** — cdn outage, api 변경, library minor version
- **recent change** — 최근 commit, dependency update, config 변경
- **caching** — stale cache, browser cache, CDN cache, memoize key
- **boundary condition** — 0, max int, empty string, unicode, leap year
- **permission/state** — 특정 user role, feature flag, A/B bucket

각 가설을 distinguishable 만들 instrumentation 설계:
- 가설 A가 맞으면 어떤 신호가 보이는가?
- 가설 B가 맞으면 어떤 신호가 보이는가?
- 신호가 겹치면 가설 분리 부족 → 더 specific 하게

3개 미만이면 confirmation bias 위험. 4-5개가 healthy.

### Phase 4. Instrument (가설 구분)

가설 A를 구분할 specific log/breakpoint, 가설 B를 구분할 specific log/breakpoint, ... 실행 → 어느 가설이 fit하는지 식별.

instrumentation 예시:

```javascript
// timing 가설
console.log('[diagnose] before-fetch user:', userId, 'ts:', Date.now());
console.log('[diagnose] after-fetch user:', userId, 'ts:', Date.now(), 'elapsed:', Date.now() - startTime);

// data state 가설
console.log('[diagnose] cache state:', { hasUser: cache.has(userId), size: cache.size });

// race condition 가설
console.log('[diagnose] concurrent calls:', activeCalls, 'lock state:', lock.locked);
```

원칙:
- prefix 일관 ([diagnose])로 cleanup 쉽게
- 가설 구분 안 되는 일반 log 금지
- 1-3개의 specific log
- production 배포 시 sampled / feature-flagged
- fix 후 모두 제거 (또는 영구 telemetry로 승격 결정)

### Phase 5. Verify Root Cause

가설 1개 fit → 진짜 원인인지 추가 검증:

- minimal repro에서 그 변수만 바꿔도 사라지는가? (positive)
- 그 변수를 다시 원래대로 되돌리면 다시 발생하는가? (negative)
- 다른 시나리오에서도 같은 패턴 발생? (generalization)
- 원인 → 증상 인과 chain 명확? (mechanism)

이 4가지 답이 모두 yes일 때만 root cause 확정. 하나라도 모호하면 가설 추가 / instrument 재설계.

흔한 실수: 첫 fit 가설을 root cause로 선언하고 직진. 사실은 correlation이고 진짜 원인은 다른 layer.

### Phase 6. Write Regression Test (fix 전)

test가 minimal repro를 자동화. 실행 → fail 확인. 이 fail이 진짜 버그를 잡았는가?

순서가 중요:
1. test 작성
2. test 실행 → **반드시 fail**
3. fail message가 버그 증상을 정확히 묘사하는가?
4. 그렇지 않으면 test 수정
5. fail 확인 후 → Phase 7

fix 후 작성하면 자연스럽게 통과해서 test가 진짜 원인을 잡았는지 모른다.

test layer:
- unit: 가능하면 우선
- integration: 가능하면 다음
- E2E: 마지막 수단 (느리고 flaky)

### Phase 7. Fix at Root Cause

원칙:
- 증상 가까이 fix 금지
- 원인 layer에서 수정
- 다른 곳에 동일 패턴 있는지 grep (예: 같은 가정의 다른 호출지점)
- fix 단위는 작게 (revertable)

fix 종류 명시:
- root_cause_fix: 진짜 원인 제거
- mitigation: 증상 완화 (부분 처리)
- workaround: 증상 회피 (원인은 후속 ticket)

mitigation/workaround면 follow-up ticket 필수.

### Phase 8. Verify Both Paths

1. regression test 실행 → pass
2. original user 시나리오 다시 실행 → 정상 동작
3. 관련 시나리오 (인접 기능) 실행 → 깨지지 않음
4. minimal repro 다시 실행 → 사라짐 확인
5. CI 전체 → green

3번이 가장 자주 누락. fix가 다른 부분 깼는지 항상 확인.

### Phase 9. Postmortem (production 버그면)

작성 항목:
- timeline (도입 → 발견 → 수정)
- impact (사용자 수, 데이터 영향, downtime)
- root cause (Phase 5 verified)
- detection (어떻게 발견했나, 더 빨리 발견할 수 있었나)
- prevention (코드 / 프로세스 / 모니터링 변경)
- action items (owner, due date)

persist-learning-jsonl로 lesson 저장:
- "X 패턴은 Y 가정 깨지면 Z로 실패한다"
- 같은 카테고리 버그가 향후 빠르게 식별되도록

blameless 원칙 — 사람이 아닌 시스템 / 프로세스에 초점.

## 6. 출력 템플릿

```yaml
bug_id: "<id>"
report:
  summary: "<한 문장>"
  reporter: "<...>"
  reported_at: "<iso>"
  environment:
    browser: "<...>"
    os: "<...>"
    version: "<...>"
    locale: "<...>"
    timezone: "<...>"
  reproduced_locally: yes | no | intermittent
  frequency: "X/10 attempts"

repro:
  steps:
    - "<step 1>"
    - "<step 2>"
    - "<step 3>"
  preconditions: ["<...>"]
  bisect_commit: "<sha or unknown>"
  recreation_cost_min: <number>

minimal_repro:
  before_lines: 120
  after_lines: 8
  artifact: "<file path or gist link>"
  positive_control: "<한 줄 변경하면 사라지는 변경>"
  negative_control: "<바꿔도 그대로인 변경>"

hypotheses:
  - id: H1
    category: race | data | env | dep | cache | boundary | permission
    description: "<...>"
    distinguishing_signal: "<instrument로 보일 신호>"
    verdict: confirmed | ruled-out | inconclusive
  - id: H2
    category: "<...>"
    description: "<...>"
    distinguishing_signal: "<...>"
    verdict: "<...>"
  - id: H3
    category: "<...>"
    description: "<...>"
    distinguishing_signal: "<...>"
    verdict: "<...>"

instrumentation:
  added_logs: ["<file:line>"]
  added_breakpoints: ["<file:line>"]
  removed_after_fix: yes | no
  promoted_to_telemetry: ["<...>"]

root_cause:
  description: "<원인 한 단락>"
  layer: data | timing | env | dep | cache | logic | ui | infra
  affected_code: ["<file:line>"]
  mechanism: "<원인 → 증상 인과 chain>"
  verified_by:
    positive_control: yes | no
    negative_control: yes | no
    generalization: yes | no
    mechanism_clear: yes | no

regression_test:
  test_file: "<path>:<line>"
  test_layer: unit | integration | e2e
  status: pre-fix-fail | post-fix-pass
  fail_message_matches_symptom: yes | no

fix:
  files_changed: ["<file>"]
  fix_type: root_cause | mitigation | workaround
  follow_up_ticket: "<id or n/a>"
  side_effect_check:
    adjacent_features_tested: ["<...>"]
    original_repro_resolved: yes | no
    minimal_repro_resolved: yes | no
    ci_green: yes | no
  similar_pattern_search:
    grep_query: "<...>"
    other_occurrences: ["<file:line>"]
    fixed_in_same_pr: yes | no

postmortem:
  required: yes | no
  timeline:
    introduced: "<commit / date>"
    detected: "<date>"
    fixed: "<date>"
  impact:
    users_affected: "<...>"
    data_affected: "<...>"
    downtime: "<...>"
  detection:
    how: "<...>"
    could_be_earlier: yes | no
    earlier_signal: "<...>"
  prevention:
    code_changes: ["<...>"]
    process_changes: ["<...>"]
    monitoring_changes: ["<...>"]
  action_items:
    - description: "<...>"
      owner: "<...>"
      due: "<...>"
  learning_logged: yes | no
  learning_id: "<id from persist-learning-jsonl>"
```

## 7. 자매 스킬

- 앞 단계: `monitor-regressions` (CI/Sentry signal) → diagnose-bug. `Skill` tool로 invoke.
- 앞 단계: `triage-work-items` (incoming bug 분류) → diagnose-bug.
- 페어: `iterate-fix-verify` — fix 단계가 동일 form. fix가 작으면 iterate-fix-verify로 위임 가능.
- 후속: `persist-learning-jsonl` — postmortem learning 저장.
- 후속: `build-with-tdd` — regression test 작성 패턴 동일 (RED → GREEN).
- 후속: `code-review` — fix가 root cause인지 reviewer 검증.

호출 흐름:
```
triage-work-items (bug 식별)
  ↓
diagnose-bug (이 스킬)
  ↓
iterate-fix-verify (fix 큰 경우 위임) | 직접 fix
  ↓
persist-learning-jsonl (lesson 저장)
```

## 8. Anti-patterns

1. **재현 없이 fix** — "이게 원인일 거야" 추정 fix. 재현 안 되면 fix 검증 불가. 무조건 Phase 1 통과해야 Phase 7 진입.

2. **첫 가설 confirmation bias** — "이거다" 1개로 직진. 3개 이상 hypothesis 강제. 첫 가설은 90% 틀린다.

3. **printf debugging 도배** — 가설 구분 안 되는 일반 log 100줄. specific log 1-3개로 가설 ruled-in/out 가능해야 함.

4. **Symptom fix** — try-catch + ignore, if-then-default-value, swallow 후 retry. root cause 회피. data layer 버그를 UI에서 가리는 패턴 최악.

5. **Regression test fix 후 작성** — 자연스럽게 통과해서 test가 진짜 원인을 잡았는지 모름. RED 단계 강제.

6. **Original path 재검증 skip** — fix가 다른 곳 깼는지 확인 안 함. fix가 새 버그 introduce하는 흔한 패턴.

7. **Bug 1개 fix 후 동일 패턴 검색 skip** — 같은 원인 (예: 같은 데이터 가정) 다른 곳에 있을 가능성. grep 강제. 한 번에 같이 fix.

8. **Postmortem 없이 닫기** — production 버그는 lesson learned 저장. persist-learning-jsonl 없으면 같은 실수 반복.

9. **재현 환경 production과 다름** — locale/timezone/data shape 무시. production 재현 안 되는 staging 재현은 다른 버그.

10. **mitigation을 root cause fix로 위장** — try-catch로 가렸는데 "fixed" 라벨. fix_type 필드에 명시 강제.

11. **하나의 fit 가설을 verify 없이 root cause 선언** — Phase 5의 4가지 검증 통과 안 한 가설은 후보일 뿐.

12. **instrument 코드 production 잔존** — Phase 4에서 추가한 [diagnose] log가 그대로 남음. fix 직후 cleanup 또는 영구 telemetry 승격 결정.

13. **bisect 회피** — "어느 commit부터 깨졌는지 모르겠다"로 종결. git bisect 강제 시도. bisect 가능하면 root cause 후보 영역 급격히 축소.

14. **flaky test를 buggy로 분류 후 retry로 종결** — 가끔 fail = 진짜 race condition일 가능성 높음. retry로 가리면 production에서 터진다.
