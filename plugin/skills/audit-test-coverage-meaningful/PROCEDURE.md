# Audit Test Coverage Meaningful — line coverage 가 아니라 mutation / behavior 검증

## 1. 목적

`go test -cover` / `pytest --cov` 같은 *line coverage* 만으로는 *test 의 의미* 검증 못 함 — *line 실행* ≠ *behavior 검증*. 본 skill 은 **mutation testing + behavior assertion + edge case coverage** 3 차원 audit.

`agent-evaluation` (외부 reference, MIT) 의 *input vs output trust scoring* 패턴 차용 — *line coverage = input metric*, *mutation kill rate = output metric*.

## 2. 사용 시점

- §6 verify-quality 의 *test quality 검증* 영역
- 분기 audit — line coverage 80%+ 도달 후 *진짜 검증* 평가
- `generate-tests-from-spec` (§5) 가 skeleton 자동 생성 후 *logic 채워졌나* 검증
- production bug 사후 — *test 가 잡았어야* 하는 영역 식별
- code review 시 *추가 test* 권장 강도 결정

## 3. 입력

### 필수
- 기존 test suite (unit / integration / E2E)
- 언어 / mutation 도구 (Stryker / mutmut / go-mutesting)
- coverage report (line / branch baseline)

### 선택
- 사용자 incident log (test 가 *못 잡은* 영역)
- 도메인 전문가 review (edge case 감수)

## 4. Stage 흐름

### Stage 1: Line coverage = input metric

기존 line coverage 측정. 단 *결과만 baseline* — *quality 평가 X*.

| 언어 | 도구 |
|------|------|
| TS / JS | Vitest / Jest --coverage / c8 |
| Python | pytest --cov / coverage.py |
| Go | go test -cover |
| Rust | cargo tarpaulin |

→ 80%+ 가 *최소 baseline*. 100% 가 *quality 보장 X* (line 실행했다고 behavior 검증한 것 아님).

### Stage 2: Mutation testing = output metric

mutation testing 은 *코드를 의도적으로 변형* 후 *test 가 잡는지* 검증:

| mutation 종류 | 예 |
|---------|---|
| Conditional boundary | `>` → `>=`, `<` → `<=` |
| Arithmetic | `+` → `-`, `*` → `/` |
| Logical | `&&` → `\|\|`, `!x` → `x` |
| Return value | `return x` → `return null` |
| Constant | `0` → `1`, `true` → `false` |

각 mutant 에 test suite 실행:
- Test fail → mutant *killed* (정상)
- Test pass → mutant *survived* (test 의 *gap*)

**Mutation Score** = killed / total mutants. 80%+ 권장.

도구:
- TS: Stryker
- Python: mutmut
- Go: go-mutesting / Pitest (Java reference)

### Stage 3: Behavior assertion 차원

각 test 의 *assertion* 이 *behavior 검증* 인지 *implementation detail* 인지:

| 패턴 | 분류 |
|------|------|
| `expect(result).toEqual(...)` | behavior (output) ✓ |
| `expect(mockFn).toHaveBeenCalledWith(...)` | implementation detail (call pattern) — fragile |
| `expect(state.internalCounter).toBe(3)` | implementation detail — refactor 시 깨짐 |
| `expect(response.status).toBe(200)` | behavior (contract) ✓ |

→ implementation detail 비율 30%+ 면 *brittle test suite* 신호. behavior assertion 으로 refactor.

### Stage 4: Edge case coverage

각 function / endpoint 의 *edge case* 검증:

| edge case | 검증 |
|---------|------|
| Empty input | "" / [] / {} / null |
| Boundary | min / max / 0 / -1 / 큰 값 |
| Concurrent | race condition (특히 stateful) |
| Failure | 의존 fail (network / DB / timeout) |
| Locale | i18n / timezone / charset |

→ edge case test 명시 비율 측정.

### Stage 5: Trust score (agent-evaluation 패턴 차용)

3 차원 종합 score:

```
trust_score = (
  line_coverage * 0.2 +     // input metric (낮은 weight)
  mutation_score * 0.4 +    // output metric (mutation kill)
  behavior_ratio * 0.2 +    // behavior assertion 비율
  edge_case_ratio * 0.2     // edge case coverage
)
```

→ trust_score 80%+ = *test suite 가 진짜 검증*. < 50% = test 가 *false confidence* 만 줌.

### Stage 6: Improvement plan

낮은 영역에 *targeted action*:

| 영역 | action |
|------|------|
| Line coverage 낮음 | 누락 file / function 에 test 추가 |
| Mutation score 낮음 | survived mutant 분석 → 더 정확한 assertion |
| Behavior ratio 낮음 | implementation detail assertion 을 behavior 로 refactor |
| Edge case 낮음 | edge case checklist 적용 |

## 5. 산출물 형식

```markdown
## Test Coverage Audit — {제품 / 분기}

### 4 차원 score
| 차원 | 값 | weight | 가중 |
| line coverage | 85% | 0.2 | ... |
| mutation score | 60% | 0.4 | ... |
| behavior ratio | 70% | 0.2 | ... |
| edge case ratio | 50% | 0.2 | ... |
| **trust_score** | | | XX% |

### Survived mutants 분석
| mutant | 위치 | test suite gap |

### Brittle test (implementation detail)
| test | 패턴 | refactor 권장 |

### Improvement plan
1. mutation score → 80% 목표
2. ...
```

## 6. 검증

- [ ] Line coverage *baseline* 으로만 사용 (quality 결론 X)?
- [ ] Mutation testing 도구 적용 + score 80%+ 목표?
- [ ] Behavior assertion 비율 70%+ ?
- [ ] Edge case checklist 5+ 영역 적용?
- [ ] Trust score 종합 산출?
- [ ] Improvement plan 영역별 action?

## 7. 다음 phase

- `generate-tests-from-spec` (§5) — survived mutant 의 *new test skeleton* 자동 생성
- `pair-program-loop` — brittle test refactor pair session
- `audit-error-budget` (§8) — production incident 와 test gap 정합

## 8. 참조

- Mutation Testing (Stryker / Pitest) — output metric
- agent-evaluation (외부 reference, MIT — Kevin + Claude 작성, OMAS v2 trust-centric scoring) — input vs output 패턴
- Working Effectively with Unit Tests (Jay Fields) — behavior assertion
- buddy `generate-tests-from-spec` (§5, 구현됨) — gap 자동 채움 cascade
