# Subagent Pressure Test — Step 2 RED · Step 9 REFACTOR 운용 상세

본 문서는 `write-a-skill`의 Step 2(RED — Adversarial Baseline)와 Step 9(REFACTOR — Subagent Pressure Test) 실행 메커니즘을 상세화한다. PROCEDURE.md 본문은 *왜·언제*를, 본 문서는 *어떻게*를 다룬다.

---

## 1. Subagent 호출 메커니즘

### 1.1 buddy 환경에서의 표준 dispatch

본 스킬은 Claude Code 환경에서 동작하므로 *Task tool* (또는 동등한 sub-agent dispatch 메커니즘)을 사용한다. 호출 형태:

```
Task tool:
  description: "RED test — adversarial baseline for <skill-name>"
  prompt: "<test scenario>"
  subagent_type: "general-purpose"
```

다음 조건을 만족시켜야 한다:
- subagent에는 *오직 description 1줄과 시나리오만* 전달 (Step 2의 경우). 본문 노출 금지 — 본문 없이도 작업 가능한지가 가치 검증의 핵심.
- Step 9의 REFACTOR test에서는 *완성된 PROCEDURE.md 본문도 함께* 전달.
- subagent의 응답은 호출자 컨텍스트에 *전체 텍스트로* 수집 (요약 금지 — pass/fail 판정 근거).

### 1.2 buddy 외부 환경 대응

OpenClaw / Codex / Cursor 등 다른 agent harness에서 실행될 경우, *동등 기능의 dispatch primitive*를 사용한다. 핵심 요구사항은 다음 3가지:

1. **컨텍스트 격리** — subagent가 호출자의 conversation history에 접근 불가
2. **선택적 컨텍스트 주입** — description 또는 본문만 명시적 전달
3. **응답 전문 수집** — subagent 응답을 truncate 없이 받음

이 3가지를 충족하지 못하는 환경에서는 *수동 시뮬레이션*으로 대체: 별도 세션을 열어 사람이 subagent 역할 수행.

### 1.3 호출 시점·횟수

- **Step 2 RED**: 최소 1회. 시나리오 3종 중 *최소 1개* (실제 사용자 시나리오 / 경계 시나리오 / anti-pattern 유발 시나리오) — 다중 시나리오를 한 호출에 묶지 말 것 (실패 모드 구분 어려움).
- **Step 9 REFACTOR**: 1~3회. Step 2 시나리오를 *완성된 본문과 함께* 재실행. 3회 상한 — 회귀 정책은 §3.3 참조.

---

## 2. Pass / Fail 판정 기준 (이진)

### 2.1 Step 2 RED — pass 조건

RED는 *실패해야 pass*다. 다음 셋 중 하나가 관찰되면 pass:

- subagent가 사용자 의도와 *명백히* 다른 결과를 냄
- subagent가 *anti-pattern을 자발적으로 수행* (예: 본문 verbatim 복사, free-form 산문 출력)
- subagent가 *결정 모호*로 stop / clarification 요청

RED가 *성공*하면 (subagent가 description만으로 의도대로 작업 완료) 본 스킬은 *불필요한 절차* — 작성 중단 검토 (PROCEDURE.md §5 Step 2).

### 2.2 Step 9 REFACTOR — pass 조건

다음 *모두* 충족 시 pass:

- subagent가 본문대로 절차 수행 (Step 1~10 순서 위배 0건)
- Step 2 RED 시나리오에 대해 *anti-pattern을 회피*하는 결정을 자발적으로 내림
- 본문에서 명시한 forcing question을 trigger되는 상황에서 *놓치지 않음*
- 출력 형식(yaml)을 schema대로 채움

판정 주체는 *write-a-skill 호출자*(상위 LLM 또는 사람). 정성적이지만 위 4항목을 binary로 분리해 각 ✓/✗ 표시 → 4개 모두 ✓일 때만 pass.

### 2.3 회색지대

- subagent가 *부분적으로 따르고 부분적으로 헤맴* → fail (REFACTOR 1회 더 시도). 절반의 성공은 성공 아님.
- subagent가 *본문에 없는 합리적 결정*을 추가로 내림 → 본문 보강 신호일 수도 있음 (§8 anti-patterns 또는 §4 핵심 원칙으로 흡수)

---

## 3. REFACTOR 회귀 시 산출물 unwind 절차

Step 9 REFACTOR 3회 안에 pass 못 하면 *description 또는 scope 자체에 문제* — Step 1으로 회귀. 회귀 전 기존 산출물 처리:

### 3.1 산출물별 unwind 가이드

| 산출물 | 회귀 시 처리 |
|--------|------------|
| `plugin/skills/<name>/PROCEDURE.md` | *보존* (다음 cycle 입력으로 재사용) — 다만 *.draft* 마커 추가 |
| `plugin/skills/<name>/references/*.md` | 보존 |
| `plugin/skills/router/references/skill-catalog.md` 등재 | **롤백** — 추가한 1줄 제거. catalog는 *pass 후*에만 영속화 |
| `plugin/skills/router/references/routing-rules.md` 변경 | 롤백 |
| `plugin/commands/<name>.md` | 롤백 |
| `NOTICE` attribution block | *조건부 보존* — 외부 자산을 *실제로 차용*했다면 보존 (법적 의무), 차용 결정이 번복되면 롤백 |
| `README` Acknowledgments | NOTICE와 동일 |

### 3.2 unwind 순서

```
1. catalog·routing·command 롤백 (외부 가시 효과 제거)
2. NOTICE·README 검토 (조건부 보존/롤백)
3. PROCEDURE.md를 *.draft로 마커
4. Step 1 재진입 — description 또는 scope 재정의
```

### 3.3 회귀 사이클 상한

총 3회 회귀(= 4 cycle 시도)에도 pass 못 하면 *해당 스킬 작성 자체를 중단* 권장. consult-codex로 외부 의견 또는 critique-plan으로 scope 재검토.

---

## 4. RED 시나리오 강도 가이드

PROCEDURE.md §5 Step 2의 "시나리오 3종 중 하나" 표현은 *최소 강도*. 실제 권장 강도는 skill type에 따라 다르다.

| skill type | 최소 시나리오 수 | 권장 시나리오 수 |
|-----------|------------------|------------------|
| discipline-enforcing | 2 (실제 + anti-pattern 유발) | 3 (+ 경계 시나리오) |
| technique | 1 (실제) | 2 (+ 경계) |
| pattern | 1 (실제 사용 경계) | 2 |
| reference | 0~1 (스킬이 절차가 아닌 lookup이면 RED 면제 가능) | 1 |

discipline-enforcing은 *합리화 패턴*까지 RED로 검증해야 함 (§4 핵심 원칙 #10).

---

## 5. 호출 흐름 예시 (전체 cycle)

```
[Step 1 Scope]
  ↓ description 1줄 확정
[Step 2 RED]
  ↓ Task tool dispatch (시나리오 1개+)
  ↓ subagent 실패 모드 수집
[Step 3 GREEN]
  ↓ buddy 표준 §1~§8 골격 작성
[Step 4 Boundary] [Step 5 Anti-pattern] [Step 6 Output]
  ↓
[Step 7 Catalog] [Step 8 Attribution]
  ↓ 외부 가시 산출물 등록
[Step 9 REFACTOR]
  ↓ Task tool 재 dispatch (Step 2 시나리오 + 완성 본문)
  ↓ pass? → Step 10
       ↓ fail (cycle < 3)? → Step 5/8 보강 후 Step 9 재실행
       ↓ fail (cycle = 3)? → §3 unwind → Step 1 회귀
[Step 10 Verify]
  ↓ make test-routing pass
  ↓ catalog grep hit
  ↓ 종료 선언
```

---

## 6. Anti-rationalization: "RED test를 건너뛰어도 되는" 합리화

다음 합리화가 등장하면 *RED test 건너뛰기 거부*:

- "스킬이 명백히 단순해서 RED 불필요" → 단순함의 판정 주체는 *subagent*. 본인 판단 ≠ subagent 판단.
- "비슷한 스킬이 이미 있어서 패턴 검증됨" → 본 스킬의 description은 *새로움*. 기존 검증 재사용 안 됨.
- "Time/budget 부족" → RED 1회 dispatch는 ~30초·~수천 토큰. 건너뛰면 ship 후 발견 비용 100배.
- "내가 본문을 신중히 썼으니 작동할 것" → 자신 검증 = sample size 1. subagent는 sample size 2.

§4 핵심 원칙 #9 "Test before declare done"의 정신.
