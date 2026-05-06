---
name: iterate-fix-verify
description: "[패턴 라이브러리] finding 하나씩 fix → atomic commit → re-verify 반복 repair loop. before/after evidence + rollback safety 유지. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '반복 수정' / 'iterative fix' / '수정 루프' / 'find-fix-verify'. 참조 위치: critique-plan 후 발견 이슈 수정, diagnose-bug fix 단계, build-with-tdd refactor."
type: skill
---

# Snippet: Iterative Fix-and-Verify Loop


> Fix 루프 패턴 — visual design을 넘어 모든 iterative repair에 일반화. Plan-time variant는 `review-design` 스킬에서.

## 이 snippet을 사용하는 경우

- "이슈 발견, 하나씩 수정, 각각 검증, 반복" 형태의 모든 워크플로우.
- 디자인 polish, 코드 정리, 문서 일관성, 의존성 업그레이드, 접근성 pass, lint 정리.
- Atomic commit + before/after 증거가 bisectability와 revert를 저렴하게 만드는 모든 곳.

이 루프는 의도적으로 **fix당 느리지만 복구는 빠르다**: 하나의 변경, 하나의 commit, 하나의 검증. 추가 commit 비용은 번들된 회귀를 디버깅하는 비용에 비하면 무시할 수 있다.

---

## Triage 먼저 (루프 진입 전 게이트)

발견된 finding을 임팩트 순으로 정렬한 뒤, 어떤 것을 고칠지 결정:

- **High impact** — 먼저 수정. 신뢰, 정확성, 첫인상에 영향.
- **Medium impact** — 다음 수정. Polish 저하, 무의식적으로 느껴짐.
- **Low / polish** — 예산이 남으면만 수정.

제어 불가능한 source(서드파티 위젯, 다른 팀이 소유한 콘텐츠, 건드릴 수 없는 인프라)에서 수정 불가능한 항목은 임팩트와 무관하게 **deferred**로 표시. Deferred 항목을 루프에 끌고 오지 말 것.

---

## 루프 (finding당, 임팩트 순)

### 1. Source 찾기

책임 있는 파일(들)을 검색 — class name, symbol, glob, grep. 무언가를 건드리기 전에 정확한 파일(들)을 가리킬 수 있는지 확인.

- finding과 **직접 관련된** 파일만 수정.
- 가장 작은 blast-radius 레이어 선호 (styling > config > component > structural).

### 2. Fix (과학적 디버깅 & TDD)

*   **Diagnose (증명 루틴):** 감으로 고치지 마라. 다음 단계를 엄격히 따른다.
    1. **Reproduce:** 버그를 재현하는 최소 단위의 실패하는 테스트(Failing Test)를 먼저 작성한다.
    2. **Hypothesize:** 왜 실패하는지에 대한 가설을 세운다.
    3. **Instrument:** 로그나 디버거를 통해 가설을 검증할 증거를 수집한다.
*   **TDD (Behavior-First):** 
    1. **Observable Behavior:** 내부 구현이 아니라 외부에서 관찰 가능한 동작을 테스트한다.
    2. **Red-Green-Refactor:** 테스트가 실패하는 것을 확인(Red)하고, 통과할 만큼만 최소한으로 구현(Green)한 뒤, 구조를 개선(Refactor)한다.
*   가장 작은 변경으로 최선의 효과를 내되, 주변 코드 오염을 최소화한다.

### 3. Atomic commit (fix 하나당 commit 하나)

```
git add <only-the-files-you-changed>
git commit -m "<scope>: FINDING-NNN — short description"
```

- **Fix 하나당 commit 하나. 번들 절대 금지.** 번들된 fix는 bisectability를 파괴하고 all-or-nothing revert를 강제한다.
- 명시적 파일 경로만 stage — `git add .`나 `-A` 금지 (관련 없는 dirty state 딸려감 방지).
- 메시지 포맷은 finding ID를 명명해 report ↔ commit ↔ revert가 모두 연결되도록.

### 4. before/after 증거로 re-verify

영향받은 surface를 re-exercise하고 매 fix마다 **before/after 쌍**을 캡처한다. 형태는 도메인에 따라 다르다:

| Domain | before/after artifact |
|--------|----------------------|
| Visual / UI | 스크린샷 쌍 (`finding-NNN-before.png` / `-after.png`) |
| Code behavior | failing-test-output / passing-test-output |
| Performance | 메트릭 snapshot 전후 (load time, bundle size, query plan) |
| Docs | 렌더링된 diff (old paragraph / new paragraph) |
| Lint / type | 에러 개수 전후, 또는 특정 에러 제거 확인 |

쌍이 영수증이다. 이게 없으면 "fixed"는 증거가 아니라 주장이다.

### 5. 결과 분류

- **verified** — re-test가 fix를 확인; 새 에러 없음.
- **best-effort** — fix 적용됐으나 완전 확인 불가 (특정 state, env, 또는 사람 눈 필요).
- **reverted** — 회귀 감지 → `git revert HEAD` 즉시 → finding **deferred** 표시.

빠른 revert는 기능이다. Atomic commit이 이를 무료로 만든다.

### 6. 기록 + 다음으로

실행 중인 리포트에 추가: finding ID, 상태, commit SHA, 건드린 파일, before/after artifact 경로. 그다음 다음 finding으로 step 1 복귀.

---

## 자기 조절 (매 N fix마다 리스크 평가)

약 5 fix마다 또는 **모든 revert 후**, 러프 리스크 점수를 계산하고 계속할지 결정:

```
0%에서 시작
매 revert:                            +15%
매 safe-layer 변경 (CSS/config):        +0%
매 structural 변경 (component/code):    +5% per file
Fix #10 이후:                         +1% per additional fix
finding과 관련 없는 파일 건드리기:      +20%
```

**리스크 > 20%:** STOP. 지금까지 한 일을 사용자에게 보여줌. 계속할지 질문.

**하드 상한: 세션당 30 fix**, 남은 backlog와 무관하게.

---

## 중단 조건

다음 중 **하나라도** 참이면 루프 중단:

1. **고칠 수 있는 이슈 없음** — backlog가 비었거나 deferred만 남음.
2. **Diminishing returns** — 마지막 3개 fix가 모두 "polish" tier이고 사용자 가시 임팩트 없음, OR 하드 상한(30) 초과.
3. **스코프 크리프 리스크** — fix가 finding과 관련 없는 파일 건드림, 인접 모듈 refactor, 원 증상에서 멀리 떨어진 구조적 변경 필요. 중단하고 별도 작업으로 surface.
4. **리스크 임계치 초과** — 위 자기 조절 참조 (>20%).
5. **revert만으로 복구 안 되는 회귀 감지** — 패치 계속 말고 escalate.

---

## 최종 pass (루프 종료 후)

1. 마지막 것만이 아니라 **영향받은 모든 surface**에서 원래 audit / scan / test suite 재실행.
2. 최종 score / metric을 baseline과 비교.
3. **최종 상태가 baseline보다 나쁘면:** 두드러지게 WARN — 격리된 상태에서 fix하는 동안 무언가 조용히 회귀했다.
4. 리포트 작성: 총 findings, 적용된 fix (verified / best-effort / reverted), deferred 항목, baseline → 최종 delta, 그리고 한 줄짜리 PR 요약.
