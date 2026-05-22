# Loop Methods — feedback loop 구성 메뉴 (Phase 1 본질)

> diagnose-bug Phase 1 의 *feedback loop 구성* 단계는 본 skill 의 본질이며 나머지 Phase 는 mechanical. 빠른·결정적·날카로운 loop 가 있으면 bisection / hypothesis-testing / instrumentation 은 *그 loop 를 소비*할 뿐이다. loop 가 없으면 코드를 아무리 들여다봐도 해결되지 않는다.
>
> 출처: mattpocock-skill `engineering/diagnose` Phase 1 의 핵심 통찰 (10 methods + iterate-on-loop + non-deterministic + cannot-build-a-loop) 을 *inspired-by* (ADR-003 §2.4) 로 흡수. verbatim 0건. buddy 도메인 어휘 (Playwright / LocalStack / testcontainers / Pact / git bisect) 로 재진술.

---

## 1. Disproportionate Effort 원칙

Phase 1 에 *불균형적 노력* 을 쏟아라. **공격적·창의적·포기 금지.**

"feedback loop 가 잘 만들어졌으면 버그의 90% 는 이미 해결" 이라는 mental model 로 접근. failing test 한 가지에 anchoring 하지 말고 *아래 10 가지 메뉴 중 가장 빠른 loop* 를 선택.

빠른 결정적 loop 1개 ≫ 느린 flaky loop 5개.

---

## 2. Loop 구성 10 가지 메뉴 (대략 이 순서로 시도)

### 1. Failing test — 가장 작은 seam 에서

unit / integration / e2e 중 *버그 코드 경로에 가장 가까운* layer 에서 failing test 작성. buddy 스택에서:

- **frontend** — Playwright (E2E) 또는 Vitest (component)
- **backend** — Vitest + testcontainers (integration) 또는 Go test
- **3rd-party 경계** — Pact contract test

→ buddy 의 `test-per-actor-use-case` skill 산출물 매트릭스를 *재사용* 가능.

### 2. Curl / HTTP script

dev server 가 떠 있다면 `curl` 또는 `httpie` 로 endpoint 호출 + `jq` 로 응답 diff. 5초 loop.

```bash
curl -sf https://localhost:3000/api/users/123 | jq '.role' | diff - expected.txt
```

### 3. CLI invocation + fixture input

CLI tool 의 버그면 fixture stdin → stdout 을 알려진 snapshot 과 diff. `cmp` / `diff -u`.

```bash
./buddy advise < fixtures/session-abc.json | diff - golden/session-abc.expected
```

### 4. Headless browser script (Playwright / Puppeteer)

UI 버그면 headless 가 *human 클릭보다 빠르고 결정적*. DOM / console / network 다 assert 가능.

```javascript
await page.click('[data-test=submit]');
await expect(page.locator('[data-test=error]')).toContainText('과부족');
```

### 5. Replay captured trace

real network request / payload / event log 를 *file 로 저장* → 격리된 code path 에 replay. production 의존성 없음.

- HAR 파일 → mock server 로 replay
- DB row dump → migration test fixture
- WebSocket frame log → mock socket 으로 replay
- analytics event JSONL → 단일 함수 호출로 replay

### 6. Throwaway harness

system 의 *최소 subset* 만 띄움. 한 service + mocked deps + 단일 함수 호출로 버그 path 자극.

```typescript
// 100 line throwaway main.ts — 본 service 가 의존하는 redis/db 만 LocalStack
const result = await processOrder(fixtureOrder, { db: testDb, cache: miniRedis });
```

production code 의 wiring 을 *복사하지 말고 import*. wiring 버그까지 cover.

### 7. Property / fuzz loop

"가끔 잘못된 출력" 류 버그면 random input 1000 회 → 실패 mode 패턴 탐색. fast-check (TS) / Hypothesis (Python) / proptest (Rust).

```typescript
fc.assert(fc.property(fc.integer(), fc.string(), (i, s) => {
  return computeChecksum(i, s) === referenceImpl(i, s);
}), { numRuns: 1000 });
```

### 8. Bisection harness

알려진 두 상태 (commit / dataset / dependency version) 사이에서 발생한 회귀면 자동화:

```bash
git bisect start <bad-sha> <good-sha>
git bisect run scripts/check-bug.sh   # 0 = good, 1 = bad
```

`check-bug.sh` 가 30 초 이내에 답하면 N=100 commit 도 ~10 분.

### 9. Differential loop

같은 input 을 (a) 이전 버전 vs 현재 버전 또는 (b) config A vs config B 에 통과 → 출력 diff.

```bash
./tool-v1.2.3 < input.json > out-old.json
./tool-v1.2.4 < input.json > out-new.json
diff out-old.json out-new.json
```

### 10. HITL bash script — 최후 수단

사람이 클릭해야만 하는 경우, 사람을 *script 가 drive*. `scripts/hitl-loop.template.sh` 같은 형태로:

```bash
while true; do
  read -p "버그 봤어요? (y/n/q): " seen
  case "$seen" in y) log "POSITIVE" ;; n) log "NEGATIVE" ;; q) break ;; esac
done
```

캡처된 output 은 다음 Phase 의 입력. *완전 수동* 보다 100× 낫다.

---

## 3. Iterate-on-loop — loop 를 *product* 로 취급

일단 *어떤* loop 라도 확보했으면, 다음 질문 반복:

| 차원 | 개선 |
|------|------|
| **faster** | setup cache, 무관 init 생략, test 범위 좁힘. 30 초 → 2 초 목표 |
| **sharper** | "안 깨졌음" 보다 *특정 symptom* assert. error message 패턴, 정확한 numeric output |
| **deterministic** | 시간 pin, RNG seed 고정, filesystem 격리, network freeze, parallel race 제거 |

> **30 초 flaky loop ≈ no loop. 2 초 deterministic loop = debugging superpower.**

iterate-on-loop 자체에 *1 시간* 투자해도 OK — 그 뒤 bisection 이 *수 시간 단축*.

---

## 4. Non-deterministic bugs — reproduction rate raise

목표는 *clean repro* 가 아니라 *재현률 상승*. 1% flake → 50% flake 까지만 올려도 debug 가능.

### 전술

1. **Loop 100×, parallel** — `for i in {1..100}; do test & done; wait`
2. **Stress 환경** — CPU throttle, memory limit, IO 지연 주입
3. **Timing window 좁히기** — `setTimeout` 줄임, debounce 0
4. **Sleep 주입** — race condition 의심 지점에 100ms sleep
5. **Concurrency 강화** — worker N=1 → N=10
6. **Data 다양화** — fixture 1개 → fixture 100개 (다양한 edge value)

재현률 < 5% 면 *debug 불가능*. 50%+ 까지 끌어올리는 것이 Phase 1 의 진짜 목표.

---

## 5. Cannot-build-a-loop — 명시 STOP

위 10 가지 메뉴 + iterate + non-deterministic 모두 시도했는데 loop 가 안 만들어지면:

### STEP 1: STOP. 다음을 *명시*하라.

- 시도한 method 목록 (10 중 어느 것)
- 각 method 가 *왜* 실패했는지 (환경 의존 / 데이터 부재 / 권한 부족 / 재현률 너무 낮음)
- 현재 loop 없이 진단 시도 *금지* — hypothesize 진행 불가

### STEP 2: 사용자에게 다음 중 하나 요청

| 요청 항목 | 무엇을 받을 것인지 |
|-----------|------------------|
| 환경 접근 | staging / canary 사용자 / 임시 elevated permission |
| 캡처 artifact | HAR file, log dump, core dump, screen recording with timestamps, DB state dump (PII 마스킹) |
| Production 임시 instrumentation 권한 | sampled logging, distributed trace, custom metric — *기한 명시* (e.g. 24h) |
| 재현 환경 setup 도움 | feature flag state, A/B bucket, browser version, locale, timezone |

### STEP 3: decompose-blocker 자동 trigger

Phase 1 + Phase 1.5 의 loop 구성 시도가 *합산 3 회 연속 실패* 시 [`decompose-blocker`](../../decompose-blocker/PROCEDURE.md) 자동 호출 (A3 정합).

- token escalation 차단 — 4 회째 시도 진입 금지
- fact / 추측 / 모름 3 분류 → 분해 축 결정 → 사용자 ≤3 질문 → 가설 압축 → 다음 행동
- decompose-blocker 도 *fix 수행 X* — 다음 스킬 dispatch 준비까지

---

## 6. Phase 1 ↔ Phase 1.5 / Phase 2 / Phase 3 cascade

```
[Phase 1: Build a feedback loop]
  ├─ 10 methods 메뉴 시도 (이 reference)
  ├─ iterate-on-loop (faster/sharper/deterministic)
  ├─ non-deterministic? → reproduction rate raise
  └─ cannot-build-a-loop? → STOP + 사용자 요청 + decompose-blocker
       (3 회 실패 시 자동 trigger)

  └─ loop 확보 됨 → Phase 2 (Minimize Repro) 진행

[Phase 1.5: Re-reproduce (재현 실패)]
  └─ buddy 의 PROCEDURE.md Phase 1.5 본문 참조
     ↑ 본 reference 의 §5 (cannot-build-a-loop) 와 의미 일부 중첩 — Phase 1.5 는
       *재현 실패 후 대응 옵션* (production data / bisect / canary), §5 는
       *loop 자체를 못 만들 때의 STOP 절차*. 둘 다 적용 가능.

[Phase 2: Minimize Repro]
  └─ Phase 1 의 loop 를 *분석 가능 최소 단위* 로 축소
     (script / test / gist 형태)

[Phase 3: Hypothesize]
  └─ Phase 1 loop 가 *없으면 진입 금지*
     (mattpocock 명시 STOP)
```

---

## 7. Anti-patterns

본 reference 적용 시 회피:

1. **Failing test 한 방향 anchoring** — 10 methods 중 1번에 집착, 2-10 시도 회피
2. **Flaky loop 안주** — 재현률 30% 인데 그대로 사용. 50%+ 까지 raise 필수
3. **Loop 없이 Phase 3 진입** — mattpocock 명시 STOP 위반
4. **사용자 요청 회피** — 환경 접근 / HAR / instrumentation 권한 요청 대신 자체 추측 진행
5. **decompose-blocker escalation 지연** — 3 회 실패 후에도 4-5 회 자기 시도 (token escalation)
6. **iterate-on-loop 생략** — 30 초 flaky loop 로 bisection 진입 (수 시간 낭비)

---

## 8. 정합성

| 영역 | 정합 대상 |
|------|----------|
| Phase 1 본문 | diagnose-bug/PROCEDURE.md §5 Phase 1 |
| Phase 1.5 본문 | diagnose-bug/PROCEDURE.md §5 Phase 1.5 |
| token escalation 차단 | decompose-blocker 3 회 한도 원칙 (A3) |
| 완료 발화 evidence | router/references/verification-discipline.md Iron Law (MID 1) |
| QA layer 선택 | verify-quality 의 actor-별 test layer 매핑 |
