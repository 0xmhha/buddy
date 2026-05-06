---
name: measure-code-health
description: project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite health dashboard로 산출한다. refactor 전후나 weekly trend tracking에 사용한다.
type: skill
---

# Health — Composite 코드 품질 대시보드


당신은 **CI 대시보드 소유 Staff Engineer**. 코드 품질이 단일 metric이 아니라 타입 안전성, lint cleanliness, 테스트 coverage, dead code, 스크립트 hygiene의 composite임을 안다. 당신의 일은 사용 가능한 모든 도구 실행, 결과 점수, 명확한 대시보드 제시, 품질이 개선/하락하는지 팀이 알도록 트렌드 추적.

**HARD GATE:** 이슈 fix 금지. 대시보드와 추천만 생산. 무엇에 action할지는 사용자 결정.

## 이 스킬을 사용하는 경우
- 사용자가 "헬스 체크", "코드 품질 점수", 또는 "코드베이스 얼마나 건강해?" 요청
- 주간 cadence — week-over-week 트렌드 추적.
- 주요 refactor 전후 — 실제 delta 측정.
- 새 repo 온보딩 — 코드 품질의 한 화면 overview.
- 큰 PR merge 후 — 회귀 없는지 확인.

## Composite 점수 공식

```
Health Score = (typecheck × 0.25)
             + (lint      × 0.20)
             + (test      × 0.30)
             + (deadcode  × 0.15)
             + (shell     × 0.10)
```

카테고리가 **skip**되면 (도구 미설치, 미구성), 남은 카테고리 간 가중치를 비례 재분배. 누락된 도구에 대해 점수 페널티 금지.

### 왜 이 가중치?

- **Typecheck (25%)** — 타입 에러는 잡기 가장 싼 결함이고 종종 실제 런타임 버그 신호 (undefined 접근, 잘못된 API shape). 높은 signal-to-noise. 타입 체크 통과가 명백한 품질 신호 중 몇 안 되는 것이라 무겁게 가중.
- **Lint (20%)** — 스타일과 패턴 강제. 중요하지만 부분 주관적; 일부 경고는 취향, 다른 것은 실제 버그 (예: `noExplicitAny`). Typecheck보다 낮은 이유는 false-positive rate이 높음.
- **Test (30%) — 가장 높은 가중치** — 테스트는 코드가 *동작*이 정확함(컴파일만이 아니라)의 유일한 신호. Type과 lint 통과는 테스트 suite 빨갛으면 의미 없음. 100% type-clean 코드베이스 + 실패 테스트는 깨짐; 오타 가득하지만 테스트 통과하는 코드베이스는 보통 작동.
- **Deadcode (15%)** — Dead export, 미사용 파일, 고아 의존성은 조용한 rot. Bundle inflate, reader 혼란, 실제 의도 숨김. 제거가 기계적이고 동적/메타프로그래밍 많은 코드에서 신호 degrade해서 낮은 가중치.
- **Shell (10%)** — Shell 스크립트는 CI/CD와 dev tooling 어디나 있지만 애플리케이션 코드보다 작은 surface. Shellcheck는 실제 portability와 quoting 버그 catch. 모든 프로젝트에 스크립트 있지 않고 실패 모드가 사용자 대상이 아니라 운영적이라 가장 낮은 가중치.

## 컴포넌트별 Scoring

각 카테고리는 같은 shape로 0-10 점수: clean exit = 10, 그다음 error/warning count로 degrade.

### Typecheck (0-10)
- **10** — Clean (exit 0, zero errors)
- **7-9** — `<10` errors
- **4-6** — `<50` errors
- **0-3** — `>=50` errors

### Lint (0-10)
- **10** — Clean (exit 0, zero warnings)
- **7-9** — `<5` warnings
- **4-6** — `<20` warnings
- **0-3** — `>=20` warnings

Linter가 구별하면 에러를 경고 대비 2x 가중치로 취급.

### Test (0-10) — pass rate로 가중
- **10** — 모두 통과 (exit 0)
- **7-9** — `>95%` pass rate
- **4-6** — `>80%` pass rate
- **0-3** — `<=80%` pass rate

Runner가 coverage 노출하면 blend: 100% pass rate이지만 30% coverage인 suite가 90% coverage의 100% pass rate보다 낮은 점수여야. 제안 공식:

```
test_score = pass_rate_score * 0.7 + coverage_score * 0.3
```

Coverage 사용 불가면 pass_rate 단독.

### Deadcode (0-10)
- **10** — Clean (exit 0, zero unused exports)
- **7-9** — `<5` unused exports
- **4-6** — `<20` unused
- **0-3** — `>=20` unused

### Shell 스크립트 품질 (0-10)
- **10** — Clean (zero shellcheck findings)
- **7-9** — `<5` issues
- **4-6** — `>=5` issues
- **0-3** — 많은 critical findings (SC2086 unquoted vars, SC2046 word-split 등)

프로젝트에 `*.sh` 파일 없으면 전체 skip (10% 가중치 재분배).

## 도구 Adapter 패턴

이 스킬은 **도구 무관**. 각 카테고리는 스택의 adapter로 채우는 slot. Adapter는 세 책임:

1. **Run** — 명령 실행, stdout/stderr/exit code/duration 캡처.
2. **Parse** — count (errors, warnings, failures, unused) 추출.
3. **Report** — 위 rubric으로 0-10 점수 반환.

### TypeScript / Node.js

| Slot | Adapter | Parse target |
|------|---------|--------------|
| Typecheck | `tsc --noEmit` | `error TS\d+:` 매치 라인 |
| Lint | `eslint . --format json` (또는 `biome check .`) | JSON `errorCount` + `warningCount` |
| Test | `vitest run --coverage` (또는 `jest --coverage`) | `Tests: X passed, Y failed`; `coverage-summary.json`에서 coverage |
| Deadcode | `knip` (또는 `ts-prune`) | `Unused exports`, `Unused files` 하 라인 |
| Shell | `shellcheck **/*.sh` | 구별되는 `In <file> line N:` 블록 |

### Python

| Slot | Adapter | Parse target |
|------|---------|--------------|
| Typecheck | `mypy .` (또는 `pyright`) | `Found N errors` 요약 라인 |
| Lint | `ruff check .` (또는 `flake8`) | `Found N errors` / `:NNN:` 라인 count |
| Test | `pytest --cov` | `N passed, M failed`; `.coverage` 또는 `--cov-report=json`에서 coverage |
| Deadcode | `vulture .` (또는 `unimport`) | `unused (function|variable|import)` 매치 라인 |
| Shell | `shellcheck **/*.sh` | 위와 동일 |

### Go

| Slot | Adapter | Parse target |
|------|---------|--------------|
| Typecheck | `go vet ./...` (컴파일러가 타입 강제; vet가 의미 체크 추가) | 각 non-empty stderr 라인 = 1 issue |
| Lint | `golangci-lint run` | `N issues` 요약 |
| Test | `go test -cover ./...` | `--- FAIL:` count; `-coverprofile`에서 coverage |
| Deadcode | `deadcode ./...` (golang.org/x/tools/cmd/deadcode) | dead 함수당 한 라인 |
| Shell | `shellcheck **/*.sh` | 위와 동일 |

### Rust

| Slot | Adapter | Parse target |
|------|---------|--------------|
| Typecheck | `cargo check --all-targets` | `error[E\d+]:` count |
| Lint | `cargo clippy --all-targets -- -D warnings` | `warning:`와 `error:` count |
| Test | `cargo test` | `test result: ok. N passed; M failed` |
| Deadcode | `cargo +nightly udeps` (미사용 의존성) | `unused dependencies:` 블록 |
| Shell | `shellcheck **/*.sh` | 위와 동일 |

## Auto-Detect 모드

사용자가 health 스택 구성 안 했으면 repo에서 알려진 manifest 파일 스캔해 스택 제안.

```bash
# Type checker
[ -f tsconfig.json ] && echo "TYPECHECK: tsc --noEmit"
[ -f mypy.ini ] || grep -q '\[tool.mypy\]' pyproject.toml 2>/dev/null && echo "TYPECHECK: mypy ."
[ -f go.mod ] && echo "TYPECHECK: go vet ./..."
[ -f Cargo.toml ] && echo "TYPECHECK: cargo check --all-targets"

# Linter
[ -f biome.json ] || [ -f biome.jsonc ] && echo "LINT: biome check ."
ls eslint.config.* .eslintrc.* .eslintrc 2>/dev/null | head -1 && echo "LINT: eslint ."
grep -q 'ruff\|flake8' pyproject.toml requirements*.txt 2>/dev/null && echo "LINT: ruff check ."
[ -f .golangci.yml ] || [ -f .golangci.yaml ] && echo "LINT: golangci-lint run"
[ -f Cargo.toml ] && echo "LINT: cargo clippy --all-targets"

# Test runner
[ -f package.json ] && grep -q '"test"' package.json && echo "TEST: npm test"
grep -q 'pytest' pyproject.toml requirements*.txt 2>/dev/null && echo "TEST: pytest"
[ -f go.mod ] && echo "TEST: go test ./..."
[ -f Cargo.toml ] && echo "TEST: cargo test"

# Dead code
command -v knip >/dev/null && echo "DEADCODE: knip"
command -v ts-prune >/dev/null && echo "DEADCODE: ts-prune"
command -v vulture >/dev/null && echo "DEADCODE: vulture ."
command -v deadcode >/dev/null && echo "DEADCODE: deadcode ./..."

# Shell linting
command -v shellcheck >/dev/null && find . -name '*.sh' -not -path '*/node_modules/*' | head -1 && echo "SHELL: shellcheck"
```

Auto-detect 후 **항상 AskUserQuestion으로 사용자에게 confirm**:

> 이 프로젝트에 대해 이 health-check 도구들 감지:
>
> - Type check: `tsc --noEmit`
> - Lint: `biome check .`
> - Tests: `npm test`
> - Dead code: `knip`
> - Shell lint: `shellcheck *.sh`
>
> A) 맞음 — 이 스택 persist하고 계속
> B) 일부 도구 조정 필요
> C) Persist skip — 한 번만 실행

사용자가 A 또는 B (조정 후) 선택하면 다음 invocation에 재읽기 가능하도록 스킬이 스택을 config 파일에 persist. 제안 위치: `~/.claude/health/<project-slug>/stack.yaml` 또는 프로젝트 주 instruction 파일의 `## Health Stack` 섹션.

```yaml
typecheck: tsc --noEmit
lint: biome check .
test: npm test
deadcode: knip
shell: shellcheck *.sh scripts/*.sh
```

이후 run에서 이 config 먼저 읽고 auto-detect skip.

## 도구 실행

감지된 각 도구 순차 실행 (일부는 lockfile이나 compiler cache 공유). 각 도구에 대해 캡처:

1. 시작 시간
2. Stdout + stderr (결합)
3. Exit code
4. 종료 시간
5. 리포트용 출력 마지막 50 라인

```bash
START=$(date +%s)
OUTPUT=$(tsc --noEmit 2>&1 | tail -50)
EXIT=$?
END=$(date +%s)
echo "TOOL:typecheck EXIT:$EXIT DURATION:$((END-START))s"
```

도구가 설치 안 됐거나 "command not found" 반환하면 이유와 함께 **SKIPPED** 기록 — 0으로 점수 금지.

## 출력 포맷

카테고리, 도구, 점수, 상태, duration, 한 줄 디테일의 대시보드 테이블 렌더링.

```
CODE HEALTH DASHBOARD
=====================

Project: <name>
Branch:  <current branch>
Date:    <ISO date>

Category      Tool              Score   Status      Duration   Details
----------    ----------------  -----   ---------   --------   -------
Type check    tsc --noEmit      10/10   CLEAN       3s         0 errors
Lint          biome check .      8/10   WARNING     2s         3 warnings
Tests         npm test          10/10   CLEAN       12s        47/47 passed
Dead code     knip               7/10   WARNING     5s         4 unused exports
Shell lint    shellcheck        10/10   CLEAN       1s         0 issues

COMPOSITE SCORE: 9.1 / 10

Duration: 23s total
```

점수별 상태 라벨:
- **10** → `CLEAN`
- **7-9** → `WARNING`
- **4-6** → `NEEDS WORK`
- **0-3** → `CRITICAL`

7 미만 카테고리는 재실행 없이 action 가능하도록 상위 5-10 raw finding append:

```
DETAILS: Lint (3 warnings)
  src/utils.ts:42 — lint/complexity/noForEach: Prefer for...of
  src/api.ts:18  — lint/style/useConst: Use const instead of let
  src/api.ts:55  — lint/suspicious/noExplicitAny: Unexpected any
```

## 트렌드 추적

각 run을 프로젝트별 히스토리 파일의 JSONL 한 줄로 저장:

```
~/.claude/health/<project-slug>/runs.jsonl
```

각 라인:

```json
{
  "ts": "2026-04-23T14:30:00Z",
  "branch": "main",
  "score": 9.1,
  "components": {
    "typecheck": 10,
    "lint": 8,
    "test": 10,
    "deadcode": 7,
    "shell": 10
  },
  "duration_s": 23
}
```

Skipped 카테고리는 `null` 사용 — 트렌드 reader가 "skipped"를 "zero 점수"와 구별 가능하도록.

저장 후 마지막 5-10 엔트리 읽어 트렌드 테이블 렌더링:

```
HEALTH TREND (last 5 runs)
==========================
Date          Branch         Score   TC   Lint  Test  Dead  Shell
----------    -----------    -----   --   ----  ----  ----  -----
2026-04-19    main           9.4     10   9     10    8     10
2026-04-20    feat/auth      8.8     10   7     10    7     10
2026-04-21    feat/auth      8.2     10   6     9     7     10
2026-04-22    feat/auth      9.1     10   8     10    7     10
2026-04-23    main           9.1     10   8     10    7     10

Trend: STABLE (no significant change vs last run)
```

트렌드 라벨 규칙:
- **IMPROVING** — composite 점수가 이전 run 대비 `>=0.3` 상승
- **STABLE** — 이전 run의 `±0.3` 내
- **REGRESSING** — composite 점수가 이전 run 대비 `>0.3` 하락

### 회귀 Alert

한 run에서 composite 점수가 **`>0.5`** 하락, OR 단일 카테고리가 **`>=2` point** 하락하면 `REGRESSIONS DETECTED` 블록 렌더링:

```
REGRESSIONS DETECTED
====================
  Lint: 9 → 6 (-3) — 12 new lint warnings introduced
    Most common: lint/complexity/noForEach (7 instances)
  Tests: 10 → 9 (-1) — 2 test failures
    FAIL src/auth.test.ts > should validate token expiry
    FAIL src/auth.test.ts > should reject malformed JWT
```

회귀를 특정 파일/카테고리에 attribute해 사용자가 어디 볼지 알도록. Auto-fix 금지.

## 추천

대시보드와 트렌드 후 우선순위 추천 리스트 렌더링. **impact = weight × (10 − score)** 내림차순으로 랭크. 10 미만 카테고리만 표시.

```
RECOMMENDATIONS (by impact)
============================
1. [HIGH] Fix 2 failing tests       (Tests: 9/10, weight 30%, impact 0.3)
   Run: npm test -- --reporter=verbose
2. [MED]  Address 12 lint warnings  (Lint: 6/10, weight 20%, impact 0.8)
   Run: biome check . --write    (auto-fixable)
3. [LOW]  Remove 4 unused exports   (Dead: 7/10, weight 15%, impact 0.45)
   Run: knip --fix               (auto-removable)
```

각 추천에 포함:
- 심각도 (impact 기반 HIGH / MED / LOW)
- 카테고리와 현재 점수
- 조사 또는 fix할 정확한 명령

## 중요 규칙

1. **Wrap, replace 금지.** 프로젝트 자체 도구 실행. 구성된 도구가 리포트하는 것 대신 자체 정적 분석 대체 금지. 사용자는 자기 도구 신뢰; 당신은 aggregate, second-guess 아님.
2. **Read-only.** 이슈 fix 금지. 대시보드 제시. 사용자 결정.
3. **구성된 스택 존중.** 스택 persist돼 있으면 정확한 명령 사용. 저장된 config 위에 auto-detect 금지.
4. **Skipped는 failed 아님.** 도구 사용 불가면 skip하고 가중치 재분배. Shell 스크립트 없는 프로젝트가 점수의 10% 잃으면 안 됨.
5. **실패에 raw output 표시.** 도구가 에러 리포트하면 재실행 없이 action 가능하도록 실제 출력 (마지막 ~10-50 라인) 포함.
6. **첫 run = 트렌드 없음.** "첫 health 체크 — 아직 트렌드 데이터 없음. 변경 후 /health 재실행해 진행 추적." 말하기.
7. **점수에 정직.** 100 타입 에러 + 모든 테스트 통과 코드베이스는 건강하지 않음. Composite가 현실 반영 필수. Inflate 금지.
8. **요청 없으면 세션당 한 run.** Health 체크가 느릴 수 있음 (전체 테스트 suite, 큰 monorepo의 타입 체크). Unprompted 재실행 금지.

## 3. 결과 조치 및 연계

*   **Triage 연계:** 대시보드에서 발견된 7점 미만의 카테고리나 구체적인 위반 사항(`DETAILS`)은 반드시 `triage-work-items` 스킬을 사용하여 작업 아이템으로 등록하십시오. 
*   **상태 자동화:** 등록 시 `needs-triage` 상태로 시작하며, 심각도에 따라 우선순위를 부여합니다.
*   **Follow-up:** 다음 헬스 체크 시, 이전의 위반 사항들이 해결되었는지 트렌드 데이터를 통해 확인합니다.


