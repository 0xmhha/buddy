# Verification Discipline — 완료 발화 직전 evidence 게이트

> Cross-skill shared reference. 매 commit / PR / task / agent delegation / TDD GREEN / fix verified 발화 *직전* 적용. 모든 buddy 스킬에서 lazy-load 가능한 SSoT.
>
> 출처: superpowers `verification-before-completion` 의 핵심 원칙 (Iron Law / Gate Function / Rationalization Prevention) 을 buddy 도메인 어휘로 *inspired-by* 이식. ADR-003 §2.4 verbatim 0건 유지.

---

## 1. The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

이번 메시지 안에서 verification 명령을 *실행하지 않았다면*, "통과" / "완료" / "fix됨" 을 발화할 수 없다.

이 규칙의 *letter* 를 위반하는 것은 *spirit* 을 위반하는 것이다. 다른 단어로 우회 (e.g., "잘 동작하는 듯", "문제 없어 보임") 도 동일 위반.

---

## 2. Gate Function (5-step)

**모든 status / 만족 / 완료 발화 직전:**

```
1. IDENTIFY: 이 발화를 증명하는 명령은 무엇인가?
2. RUN: 그 명령을 *fresh* 하게 *완전* 실행
3. READ: 출력 전체 + exit code + 실패 개수 확인
4. VERIFY: 출력이 발화를 *확증* 하는가?
   - NO 면: 실제 상태를 evidence 와 함께 진술
   - YES 면: 발화 + evidence 함께 진술
5. ONLY THEN: 발화
```

step 1 개라도 건너뛰면 *verification 아님 — 거짓말*.

---

## 3. 도메인별 evidence 매핑 (buddy 시나리오)

| 발화 | 충분한 evidence | 불충분 (rejected) |
|------|----------------|------------------|
| "tests pass" | test 명령 출력: 0 failures, exit 0 | "직전 run", "should pass", linter 통과만 |
| "build succeeds" | build 명령: exit 0 | "linter 통과", "log 정상 보임" |
| "bug fixed" | 원래 증상 reproduction → 현재 통과 | "코드 바꿨음 → 고쳐졌을 것" |
| "regression test 작동" | red-green cycle: fix revert → red → fix restore → green | "test 1회 통과" |
| "agent delegation 성공" | VCS diff 확인 + 실제 변경 검증 | agent 의 "success" 보고만 |
| "requirements 충족" | line-by-line checklist 매핑 | "test 통과 = phase 완료" |
| "fix verified (iterate-fix-verify)" | before/after artifact 쌍 + re-exercise | "best effort", 시각 확인만 |
| "GREEN (TDD)" | RED 확인 → 구현 → GREEN 실행 결과 | RED 단계 skip, 또는 구현만 |
| "make test-routing 통과" | 명령 실행 + "10/10 checks passed" 라인 | "이전에 통과한 적 있음" |
| "Quality Gate 통과 (6단계)" | dimension 별 actual 결과 + classify-qa-tiers 산출물 | dimension 별 "no issues found" 빈 통과 |

---

## 4. Red Flags — STOP signal

다음 문구·상태가 나타나면 *발화 직전 STOP*:

- "should", "probably", "seems to", "looks like", "잘 될 것"
- 만족 표현 ("Great!", "Perfect!", "Done!", "완료!", "OK!") *verification 명령 실행 전*
- commit / push / PR 직전 + verification 미실행
- agent 의 success 보고를 *VCS diff 검증 없이* 채택
- 부분 verification (linter 만, dry-run 만, unit 만) 으로 전체 통과 주장
- "이번만", "지쳤음", "급함" 등 합리화 단서

---

## 5. Rationalization Prevention 표

| 합리화 | 현실 |
|--------|------|
| "이번 한 번만 skip" | 예외 0건 — letter/spirit 둘 다 |
| "확신 있음" | 확신 ≠ evidence |
| "linter 통과 = build OK" | linter 는 compiler 아님 |
| "agent 가 success 라고 함" | independent verify 필요 |
| "지쳤음" | 피로 ≠ 면제 |
| "partial check 충분" | partial 은 증명 0 |
| "다른 단어 쓰면 규칙 안 걸림" | spirit > letter |
| "지난번에 통과했었음" | fresh 가 아니면 무효 |

---

## 6. Key Patterns (호출 시점별)

### A. TDD GREEN 발화 (`build-with-tdd`)

```
✅ RED 확인 → 구현 → 테스트 fresh run → 출력 확인 → "GREEN. test X/X pass."
❌ "구현 완료, GREEN 일 것" — RED 미확인 또는 실행 0회
```

### B. iterate-fix-verify Step 5 결과 분류

```
✅ before/after artifact 쌍 + re-exercise 출력 → "verified"
✅ 부분 검증 명시 → "best-effort (특정 env/state 한정)"
❌ "fix 적용했으니 verified" — before/after 0건
```

### C. Agent delegation 결과 채택

```
✅ agent "완료" 보고 → git diff / VCS 변경 확인 → 변경 실재 확인 → 채택
❌ agent 가 "성공" 이라 했으니 채택
```

### D. verify-quality Quality Gate

```
✅ dimension 별 실제 명령 출력 + numeric metric → Gate 판단
❌ dimension 별 "no issues found" 7회 → Gate 통과 (review-engineering Anti-Rationalization 의 "All-clean STOP rule" 과 동일 위반)
```

### E. Requirements 충족 발화

```
✅ plan 재독 → line-by-line checklist → 각 항목 verify → gap 보고 또는 완료
❌ "test 통과 = requirements 충족"
```

---

## 7. When To Apply

**ALWAYS before:**

- 어떤 형태든 success / 완료 발화
- 어떤 형태든 만족 / 긍정 진술
- commit / PR 생성 / task 완료 마킹
- 다음 task / 다음 phase 이동
- agent 위임 결과 채택

**Rule 적용 범위:**

- 정확한 phrase
- paraphrase + synonym
- success 의 implication
- 완료 / 정확함을 *시사하는* 모든 communication

---

## 8. Cross-link — 어느 스킬이 이 reference 를 참조해야 하나

| 스킬 | 적용 시점 |
|------|----------|
| `build-with-tdd` | RED → GREEN 전이 발화 / REFACTOR 안전성 발화 |
| `iterate-fix-verify` | Step 5 결과 분류 / 최종 pass (전체 audit re-run) |
| `verify-quality` | Quality Gate 판단 표 직전 / 각 stage 종료 발화 |
| `build-feature` | "완료 기준" checklist 항목 발화 / subagent dispatch 결과 채택 직전 |
| `review-engineering` | "no issues found" 발화 시 (자체 Anti-Rationalization 의 All-clean STOP rule 과 정합) |
| `diagnose-bug` | bug 재현 ↔ fix 적용 ↔ 재실행 cycle 의 *fix 적용 후* |
| `iterate-product` (8단계) | A/B 실험 결과 채택 직전 (statistical significance 검증) |
| 모든 agent dispatch 후 | agent 의 success 보고 채택 직전 (VCS diff 독립 verify 의무) |

각 스킬 본문에는 1줄 cross-reference 만 추가:

```markdown
> 완료 발화 전 [`router/references/verification-discipline.md`](../router/references/verification-discipline.md) 의 Iron Law + Gate Function 5-step 적용.
```

---

## 9. A-카테고리 anti-rationalization 시스템 정합성

| 시점 | 게이트 | 위치 |
|------|--------|------|
| **결정** (design) | 3+ orthogonal 후보 강제 (AI bias gate) | 7 design 스킬 본문 + `engineering-decision-gate-mapping.md` (A2) |
| **리뷰** (plan) | 5 anti-rationalization 규칙 + self-attestation | `review-engineering` PROCEDURE.md (HIGH 완료) |
| **완료** (claim) | Iron Law + Gate Function 5-step | *본 reference* (MID 1 — 현재) |

세 layer 모두 박혀야 sycophancy / 합리화 진입점 차단 완성. 하나라도 누락 시 우회 경로 발생.

---

## 10. Bottom Line

**verification 에 shortcut 없음.**

명령 실행 → 출력 확인 → *그 후* 발화.

이건 협상 대상 아님.
