---
name: setup-quality-gates
description: "개발 환경에 husky + lint-staged + Prettier + typecheck + unit test + secret scan + commitlint를 pre-commit / pre-push / CI 게이트로 설치. 8단계 quality gate cycle 자동화. 트리거: 'quality gate 설정' / 'husky 설치' / 'pre-commit 추가' / 'lint-staged 자동화' / '타입체크 강제' / 'CI 게이트' / 'secret scan 추가'. 입력: 프로젝트 stack(node/python/go/etc), 패키지 매니저, CI 시스템(GH Actions/CircleCI/etc). 출력: husky config + lint-staged config + pre-commit hooks + CI workflow. 흐름: define-product-spec → setup-quality-gates → build-with-tdd."
type: skill
---

# Setup Quality Gates — 8단계 검증 루프 자동화

## 1. 목적

개발 환경에 **자동화된 quality gate**를 설치해 매번 commit / push / CI에서 검증되도록 한다. 사람이 잊어버려도 자동으로 차단.

8단계 quality gate cycle (Claude Code SuperClaude framework 기준):
1. **syntax** — language parser
2. **type** — TypeScript / mypy / Pyright / etc.
3. **lint** — ESLint / Ruff / golangci-lint
4. **security** — secret scan + dependency vulnerability + SAST
5. **test** — unit test (≥80% coverage) + integration test
6. **performance** — bundle size, lighthouse (UI) / bench (backend)
7. **documentation** — README / changelog / inline doc 검증
8. **integration** — E2E test + deployment validation

각 단계에 자동화 도구 + pre-commit / pre-push / CI 어디서 실행할지 정함.

## 2. 사용 시점 (When to invoke)

- 신규 프로젝트 setup (define-product-spec 완료 직후)
- 기존 프로젝트에 quality gate 추가 (legacy 정비)
- CI 실행 시간 단축 (pre-commit 강화로 CI 줄이기)
- secret commit 사고 후 (재발 방지)
- coverage 80% 미달로 매주 깨짐 (강제 게이트로 차단)
- 신규 팀원 합류 (consistent dev env)
- 회사 정책 (모든 PR이 게이트 통과)

## 3. 입력 (Inputs)

### 필수
- project root path
- package manager (npm / pnpm / yarn / bun / poetry / cargo / go mod)
- 주요 language (TypeScript / Python / Go / Rust)
- CI 시스템 (GitHub Actions / GitLab CI / CircleCI / Jenkins)
- repository host (GitHub / GitLab / Bitbucket)

### 선택
- 기존 husky / pre-commit 설정
- 회사 정책 (commit message format, branch naming)
- coverage threshold

### 입력 부족 시 forcing question
- "package manager 뭐 써? lock file 위치 다름."
- "monorepo야 single repo야? lint-staged 설정 다름."
- "secret scan 도구 정해진 거 있어? gitleaks / detect-secrets / trufflehog?"
- "CI에서 reusable action 사용 가능해 self-hosted야?"

## 4. 핵심 원칙 (Principles)

1. **Pre-commit은 빠르게** — <10초. lint-staged로 변경된 파일만.
2. **Pre-push는 중간** — typecheck + unit test (변경 영향 영역). <60초 목표.
3. **CI는 전체** — 모든 file lint + 전체 typecheck + 전체 test + coverage + integration.
4. **Skip 옵션은 비상시만** — `--no-verify` 사용 시 audit log. 일상화 금지.
5. **Auto-fix 우선** — Prettier, ESLint --fix는 자동 수정. 사람 시간 아낌.
6. **Secret scan은 pre-commit** — 늦으면 git history에 남음. push 전 차단.
7. **Coverage 80% 강제** — 80% 미만 PR merge 차단.
8. **Commitlint은 conventional commits** — feat / fix / refactor / docs / test / chore / perf / ci.

## 5. 단계 (Phases)

### Phase 1. Inventory & Detection
- project structure
- package manager + version
- existing tools (있으면 통합)
- CI system

### Phase 2. Tool Selection (stack별)

**TypeScript / JavaScript**:
- husky (git hook manager)
- lint-staged (변경 파일만 lint)
- Prettier (format)
- ESLint (lint)
- typescript-eslint
- vitest / jest (test + coverage)
- commitlint + @commitlint/config-conventional
- gitleaks (secret scan)

**Python**:
- pre-commit (framework)
- ruff (lint + format)
- mypy / pyright (type)
- pytest + pytest-cov
- bandit (security)
- detect-secrets

**Go**:
- pre-commit
- golangci-lint
- gofumpt
- go test -cover
- gosec
- gitleaks

### Phase 3. Pre-commit Hooks
변경 파일 대상:
- format auto-fix (Prettier / ruff format / gofumpt)
- lint (ESLint / ruff / golangci-lint)
- secret scan (gitleaks / detect-secrets)
- commitlint (commit-msg hook)

### Phase 4. Pre-push Hooks
변경 영향 영역 대상:
- typecheck
- unit test
- coverage check (changed files)

### Phase 5. CI Workflow
모든 파일 대상:
- lint (전체)
- typecheck (전체)
- unit test + coverage (전체, 80% threshold)
- integration test
- bundle size (UI)
- security scan (dependency + SAST)
- documentation lint (markdown)

### Phase 6. Auto-fix Workflow
- pre-commit에서 auto-fix → re-stage
- Prettier, ESLint --fix, ruff --fix
- 사람 개입 없이 처리 가능 항목

### Phase 7. Skip / Override Policy
- `--no-verify` 사용 시 commit message에 자동 tag
- audit log (누가 언제 skip)
- main branch는 skip 불가 (server-side enforcement)

### Phase 8. Documentation
- README에 setup 방법 + skip 정책
- CONTRIBUTING.md에 commit format
- onboarding doc

## 6. 출력 템플릿 (Output Format — 예시 TypeScript / Next.js)

```yaml
package_manager: pnpm
language: typescript

devDependencies_added:
  - husky
  - lint-staged
  - prettier
  - eslint
  - "@commitlint/cli"
  - "@commitlint/config-conventional"
  - vitest
  - "@vitest/coverage-v8"

scripts_added:
  prepare: "husky"
  lint: "eslint ."
  format: "prettier --write ."
  typecheck: "tsc --noEmit"
  test: "vitest run"
  coverage: "vitest run --coverage"

husky_hooks:
  pre-commit: |
    pnpm lint-staged
    gitleaks protect --staged
  commit-msg: |
    pnpm commitlint --edit "$1"
  pre-push: |
    pnpm typecheck && pnpm test

lint_staged_config:
  "*.{ts,tsx,js,jsx}":
    - "prettier --write"
    - "eslint --fix"
  "*.{md,json,yml,yaml}":
    - "prettier --write"
  "*.py":
    - "ruff format"
    - "ruff check --fix"

commitlint_config:
  extends: ["@commitlint/config-conventional"]
  rules:
    type-enum:
      - 2
      - always
      - ["feat", "fix", "refactor", "docs", "test", "chore", "perf", "ci"]

ci_workflow:
  file: ".github/workflows/ci.yml"
  jobs:
    lint:
      runs_on: ubuntu-latest
      steps:
        - checkout
        - setup_node
        - install (pnpm install --frozen-lockfile)
        - lint (pnpm lint)
        - format_check (pnpm prettier --check .)
    typecheck:
      runs_on: ubuntu-latest
      steps:
        - typecheck (pnpm typecheck)
    test:
      runs_on: ubuntu-latest
      steps:
        - test (pnpm coverage)
        - coverage_threshold: 80
        - upload_codecov
    security:
      runs_on: ubuntu-latest
      steps:
        - dependency_audit (pnpm audit --prod)
        - secret_scan (gitleaks detect)
        - sast (CodeQL)
    e2e:
      runs_on: ubuntu-latest
      steps:
        - playwright_test
        - upload_artifacts (screenshots, videos)

bundle_size:
  tool: bundlesize | size-limit
  threshold:
    initial: "500KB"
    total: "2MB"

skip_policy:
  no_verify_allowed: emergency
  audit_log: ".husky/skip-log"
  branch_protection:
    main: required_status_checks
    main: no_force_push
    main: required_review_count: 1

verification:
  install_test:
    - "pnpm install"
    - "git commit --allow-empty -m 'test: setup verification'"  # commit msg 통과 확인
    - "git commit --allow-empty -m 'invalid'"  # 실패 확인
  hook_test:
    - "echo 'AKIA1234567890ABCDEF' > test.txt && git add test.txt && git commit -m 'test: secret'"  # gitleaks 차단 확인
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `define-product-spec` — `Skill` tool로 invoke (stack 결정 후)
- 페어: `audit-security` — secret / dependency / SAST 룰
- 페어: `measure-code-health` — coverage / complexity / 품질 점수
- 다음 단계: `build-with-tdd` — gate 통과 보장된 환경에서 TDD 시작
- 다음 단계: `write-changelog` — release gate 활용

## 9. Claude Code Hook 설치 (buddy 통합)

§5 build-feature 진입 전, 또는 buddy plugin install 직후 Claude Code hook을 활성화한다.
husky pre-commit gate와 독립적으로 동작 — Claude Code hook은 *AI 도구 호출* 레이어를 감싼다.

### 자동 설치 (권장)

```bash
# user-global 설치 (~/.claude/settings.json)
bash plugin/skills/setup-quality-gates/scripts/apply-claude-hooks.sh --scope user

# project 설치 (.claude/settings.json)
bash plugin/skills/setup-quality-gates/scripts/apply-claude-hooks.sh --scope project

# 미리보기 (dry-run)
bash plugin/skills/setup-quality-gates/scripts/apply-claude-hooks.sh --dry-run
```

### 등록되는 hook 요약

| Hook 종류 | Matcher | buddy 서브커맨드 | 역할 |
|-----------|---------|-----------------|------|
| PreToolUse | Edit\|Write\|MultiEdit | `buddy hook-wrap freeze-scope` | 허가된 경로 외 파일 편집 차단 |
| PreToolUse | Bash | `buddy hook-wrap guard-destructive` | rm -rf / DROP / force-push 등 확인 요청 |
| PostToolUse | Edit\|Write\|MultiEdit | `buddy hook-wrap run-tests` | 파일 저장 후 관련 테스트 자동 실행 |
| Stop | — | `buddy hook-wrap save-context` | 세션 종료 시 컨텍스트 저장 |
| SessionStart | — | `buddy hook-wrap restore-context` | 세션 시작 시 컨텍스트 복원 |

### 수동 설치 (템플릿 직접 적용)

```bash
# 템플릿 위치: plugin/skills/setup-quality-gates/references/claude-hooks-template.json
# 기존 ~/.claude/settings.json에 "hooks" 키를 병합한다.
jq -s '.[0] * .[1]' ~/.claude/settings.json \
  plugin/skills/setup-quality-gates/references/claude-hooks-template.json \
  > /tmp/settings-merged.json && mv /tmp/settings-merged.json ~/.claude/settings.json
```

### 참조

- 훅 설계 상세: `design-claude-hooks` skill
- 훅 이벤트 스키마: `plugin/skills/design-claude-hooks/references/hook-events.md`
- 패턴 예시: `plugin/skills/design-claude-hooks/references/hook-patterns.md`

## 8. Anti-patterns

1. **Pre-commit이 너무 느림 (>10초)** — 매번 짜증. lint-staged로 변경 파일만.
2. **Husky만 설치하고 hook 비활성** — install되었지만 .husky/* 빠짐. verification 강제.
3. **Auto-fix 없이 lint만** — 매번 사람이 수정. Prettier / ESLint --fix 자동.
4. **Secret scan 누락** — git history에 남으면 rotation 비용 큼. pre-commit 필수.
5. **CI에서만 검증** — push 후 발견. pre-push로 typecheck / test 강제.
6. **Coverage threshold 없음 또는 너무 낮음 (50%)** — 사실상 게이트 무력화. 80% 강제.
7. **Commitlint 없이 자유 commit** — release note 자동화 어려움. conventional commits.
8. **`--no-verify` 일상화** — audit log + main branch server-side enforcement.
