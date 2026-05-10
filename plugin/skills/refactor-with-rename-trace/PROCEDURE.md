# Refactor With Rename Trace — LSP rename + 호출 그래프 추적 + safe refactor

## 1. 목적

이름 변경 (rename) 같은 *광범위 영향 refactor* 의 안전 절차. **LSP rename + 호출 그래프 추적 + 단계별 검증** 통합.

random search-replace 의 *놓친 callsite* / *잘못 매칭* 차단. test-driven refactor 가 baseline.

## 2. 사용 시점

- §5 build-feature 안 — function / class / variable 이름 변경
- 도메인 용어 정정 (예: `customer` → `user`)
- API 변경 시 callsite 모두 식별 필요
- dead code 제거 — 호출 0 검증
- 복잡도 큰 module 분해 — 의존 그래프 시각화

## 3. 입력

### 필수
- refactor 대상 (file / function / class / pattern)
- LSP 활성 IDE (VS Code / IntelliJ / Cursor / Zed)
- test suite (refactor 후 검증 baseline)

### 선택
- 호출 그래프 도구 (madge / dependency-cruiser / Go callgraph)
- 기존 deprecation history (이전 rename log)

## 4. Stage 흐름

### Stage 1: Pre-refactor — 호출 그래프 + test baseline

| 항목 | 도구 |
|------|------|
| 호출 그래프 | madge (JS/TS) / dependency-cruiser / go callgraph / pyan (Python) |
| 호출 위치 검색 | LSP "Find All References" / `grep -rn` cross-check |
| Test baseline | full test suite green 확인 — refactor 전 *deterministic* |

→ refactor 전 *test 1 개라도 fail* 하면 fix 먼저. broken baseline 위 refactor = 위험.

### Stage 2: LSP rename

LSP 의 `Rename Symbol` (F2 in VS Code) 사용:
- 단일 명령으로 *모든 callsite* 변경
- 동일 이름 *다른 의미* 자동 분리 (LSP 가 scope 인식)
- string literal / comment 안 동명 *건드리지 않음*

→ search-replace 보다 *정확*. 단 LSP 인식 못 하는 영역 (config / template / 외부 file) 은 별도.

### Stage 3: Cross-check — grep + LSP

LSP rename 후:
- `grep -rn "old-name"` 으로 누락 영역 찾기
- 발견 시 *왜 LSP 가 못 찾았나* 분석:
  - dynamic dispatch (reflection / string lookup)
  - 외부 file (config / template / CI script / docs)
  - LSP 인식 안 되는 언어 영역

→ 누락 발견 시 *manual fix*.

### Stage 4: Test 검증

refactor 후:
1. unit test 전체 run — green 확인
2. integration test — green 확인
3. lint / typecheck — clean 확인
4. (큰 refactor 시) E2E test sample run

→ 1 test 라도 fail 시 *즉시 revert* — refactor 단위 작아야 revert 비용 낮음.

### Stage 5: Commit 단위 — 단일 refactor = 단일 commit

| 패턴 | commit message |
|------|------------|
| Pure rename (의미 변화 X) | `refactor: rename {old} to {new}` |
| Rename + 책임 변경 | 별도 commit 분리 (rename 먼저, 책임 변경 다음) |
| Pure rename + 다수 file | 위 1 commit OK |

→ rename 과 *기능 변경 동시 commit* 금지. PR review 부담 + revert 어려움.

### Stage 6: Deprecation chain (optional)

public API rename 시 *backward compat* 유지:

```typescript
// New API
export function getUser(...) { ... }

// Deprecated alias — remove in v2.0
/** @deprecated Use getUser instead */
export const getCustomer = getUser;
```

→ deprecation period (6~12개월) 후 alias 제거. semver major 변경.

## 5. 산출물 형식

```markdown
## Refactor Log — {old → new}

### Pre-refactor
- 호출 그래프 도구: ...
- callsite count (grep): {N}
- test baseline: green / red

### LSP rename
- IDE: ...
- LSP-found references: {N}
- LSP 인식 못한 영역: ...

### Cross-check
- grep "old-name" residual: {N}
- manual fix: ...

### Test 검증
- unit / integration / lint / typecheck: pass / fail

### Commit
- {SHA} `refactor: rename old to new`

### Deprecation (if public API)
- alias period: 6m
- semver: major
```

## 6. 검증

- [ ] Pre-refactor *test baseline green*?
- [ ] LSP rename + grep cross-check?
- [ ] LSP 인식 못한 영역 (config / template / 외부) 모두 식별?
- [ ] Refactor 후 test / lint / typecheck *모두 green*?
- [ ] Commit 단위 *pure rename* (기능 변경 동시 X)?
- [ ] Public API 시 deprecation alias?

## 7. 다음 phase

- `update-docs-with-code` — rename 후 docs 동기화
- `build-with-tdd` — refactor 후 새 cycle 진입
- `monitor-regressions` — production 영향 추적

## 8. 참조

- Refactoring (Martin Fowler) — refactor 카탈로그
- LSP (Language Server Protocol) specification
- madge / dependency-cruiser — 호출 그래프 도구
