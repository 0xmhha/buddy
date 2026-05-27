# Pair Program Loop — driver / navigator + AI agent pair driving

## 1. 목적

전통 *driver / navigator* pair programming 패턴 + AI agent 페어 변형. **driver = 코드 작성, navigator = 검토 / 다음 step** 역할 명시 분리.

`build-with-tdd` 의 red-green-refactor 사이클과 cascade — pair 가 *각 step 의 검토 + 빠른 분기 결정* 강화.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 작업 대상 설명 | ✅ | knowledge | 사용자 발화 | "무엇을 함께 작업할까요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Working code (driver/navigator swap 기록) | artifact | source files + commit | `verify-quality` |

## 2. 사용 시점

- §5 build-feature 의 *복잡 / 높은 risk feature* 작성 시
- `dispatch-parallel-agents` 의 sub-stream 으로 — 1 stream 이 pair driving
- 신규 dev 온보딩 — pair learning
- AI agent 활용 시 — 사용자 driver / agent navigator 또는 역
- 디버깅 *재현 어려운 issue* — 두 시각으로 빠르게 좁힘

## 3. 입력

### 필수
- 작성 대상 task (acceptance criteria 있는 단일 task)
- pair 구성: human + human / human + AI / AI + AI
- `build-with-tdd` 진행 상태 (red / green / refactor)

### 선택
- 이전 pair session log (있으면 — pattern 학습)
- 도구 (VS Code Live Share / Tuple / 단순 screen share)

## 4. Stage 흐름

### Stage 1: 역할 정의

| 역할 | 책임 |
|------|------|
| **Driver** | 코드 작성 — keyboard 통제 |
| **Navigator** | 검토 + 다음 step + 더 큰 그림 — 코드 작성 X |

→ 5~15분 단위 *역할 swap*. AI agent pair 시 swap = "다음 step navigator 역할 전환" 명시.

### Stage 2: TDD red 단계

navigator 가 *failing test* spec 명시:
- "test 이름은 ..."
- "expected 값은 ..."
- "edge case 는 ..."

driver 가 작성. test fail 확인.

### Stage 3: TDD green 단계

navigator 가 *최소 구현* 명시:
- "이 test 만 통과하는 가장 간단한 코드"
- "premature optimization 차단"

driver 가 작성. test pass 확인.

### Stage 4: TDD refactor 단계

pair 가 함께:
- 중복 제거
- 명명 개선
- 함수 분해

→ test 통과 *유지* 하면서. refactor 중 test 깨지면 즉시 revert.

### Stage 5: AI agent pair 변형

AI agent 가 *navigator* 일 때:
- driver = 사용자
- agent 역할: *next step / edge case 제안 / refactor 제안*
- agent 의 *코드 직접 작성 X* (사용자 통제 유지)

AI agent 가 *driver* 일 때:
- navigator = 사용자
- agent 역할: *코드 작성*
- 사용자 = *spec 명시 + 검토 + 결정*

→ 두 모드 *명시 swap*. agent 가 *driver + navigator 동시* 면 사용자 통제 약화.

### Stage 6: Anti-pattern 회피

| anti-pattern | 회피 |
|------------|-----|
| Navigator 가 keyboard 잡음 | 역할 분리 강제 — swap 까지 대기 |
| Driver 가 *큰 그림* 결정 | navigator 영역 — driver 는 작성에 집중 |
| Swap 안 함 | 15분 timer 강제 |
| 둘 다 침묵 | navigator 가 *생각 audible* — driver 가 무엇 할지 명시 |
| Distraction (slack / notification) | DND 모드 + focus session |

## 5. 산출물 형식

```markdown
## Pair Session Log — {task}

### Pair 구성
- driver: {human / AI agent}
- navigator: {human / AI agent}
- swap 주기: 15min

### TDD cycle 진행
- red 시작 시간 / green 시작 / refactor 시작
- swap 횟수
- 작성 LOC

### Decision log
- 큰 결정 (architecture / naming / scope)

### Next session 입력
- 미완료 TODO
- 학습한 pattern
```

## 6. 검증

- [ ] Driver / navigator 역할 *명시 분리* (keyboard 통제 분명)?
- [ ] Swap 주기 15분 (또는 task 단위)?
- [ ] AI agent pair 시 *driver + navigator 동시 X* (사용자 통제 유지)?
- [ ] TDD red-green-refactor 사이클 따름?
- [ ] Anti-pattern (역할 흐림 / swap X / distraction) 회피?

## 7. 다음 phase

- `build-with-tdd` 의 후속 cycle
- `iterate-fix-verify` 의 fix → verify cascade
- `dispatch-parallel-agents` 의 multi-stream 안 1 stream

## 8. 참조

- Extreme Programming Explained (Kent Beck) — pair programming origin
- Pair Programming Illuminated (Williams + Kessler)
- AI agent pair pattern — emerging (no canonical reference yet)
- buddy `build-with-tdd` (구현됨) — TDD cycle 입력
