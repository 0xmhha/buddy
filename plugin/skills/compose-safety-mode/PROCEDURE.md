---
name: compose-safety-mode
description: "[패턴 라이브러리 / META] guard-destructive-commands + freeze-edit-scope 같은 multiple safety hooks를 하나의 max safety mode로 compose. layered guardrail system 설계 패턴. 직접 invoke보다 orchestrator가 import해 사용. 트리거: 'safe mode 켜자' / 'careful + freeze' / '위험 작업 보호' / 'safety mode'. 참조 위치: autoplan/critique-plan 고위험 단계 권장, production 작업, billing/auth 코드 수정."
type: skill
---

# Snippet: Safety Mode Composition


> Composition meta-pattern. 개별 훅(careful, freeze)은 별도 스킬.

## 이 snippet을 사용하는 경우
- 여러 PreToolUse 훅 결합 (destructive-command guard + path-scope freeze + rate limit 등)
- "단일 명령으로 모든 가드 ON" UX
- 독립적인 상태 파일들에 걸친 조율된 tear-down

## Composition Strategy

각 안전 훅(safety hook)은 **독립적으로** 동작하도록 설계한다.
"max safety" 모드는 *훅을 합치는* 게 아니라 *훅들을 동시에 활성화*하는 얇은 래퍼다.

핵심 규칙:
- 각 훅은 자신만의 `matcher`(Bash / Edit / Write 등)에 바인딩한다.
- 한 도구가 여러 훅의 `matcher`에 매칭되면 모두 실행된다(순서대로).
- **결과 결합은 OR**: 한 훅이라도 deny하면 그 도구 호출은 차단된다.
- 훅 본체는 사이블링 스킬의 스크립트를 그대로 재사용 → 중복 구현 금지(SSoT).

```yaml
# 예시: 두 개의 독립 훅을 한 스킬에서 한꺼번에 등록
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "bash <path/to/careful/check-careful.sh>"
    - matcher: "Edit"
      hooks:
        - type: command
          command: "bash <path/to/freeze/check-freeze.sh>"
    - matcher: "Write"
      hooks:
        - type: command
          command: "bash <path/to/freeze/check-freeze.sh>"
```

*이유: 훅 간 결합도가 0이어야 한쪽 버그가 다른 쪽 검증을 무력화하지 못한다.
또한 단독 스킬(`/careful`, `/freeze`)도 그대로 살아있어야 사용자가 부분 활성화할 수 있다.*

## State File Layout

여러 훅의 상태 파일은 **이름공간이 충돌하지 않도록** 분리한다.

```
${HOOK_STATE_DIR}/         # 예: $HOME/.<your-tool>/, $CLAUDE_PLUGIN_DATA
├── careful-overrides.txt  # /careful 훅 전용 (override 카운트/타임스탬프)
├── freeze-dir.txt         # /freeze 훅 전용 (허용 디렉토리 절대경로)
└── guard-active.txt       # composition wrapper 전용 (활성 플래그)
```

원칙:
- 훅 하나당 파일 하나. 절대 한 JSON에 모든 상태를 합치지 않는다.
- 파일 이름에 훅 이름을 prefix → grep·디버깅·정리가 쉽다.
- 디렉토리는 `mkdir -p`로 멱등 보장. 부재 시 훅은 "비활성"으로 fail-safe 동작.

*이유: 단일 상태 파일은 한 훅의 schema 변경이 다른 훅을 깨뜨린다(드리프트).*

## Activation UX

"한 번 묻고, 모두 켠다." 사용자는 모드 진입에 필요한 최소 입력만 제공.

흐름:
1. 한 가지 필수 파라미터만 묻는다 (예: freeze 대상 디렉토리).
   - 다른 훅(예: careful)이 파라미터를 요구하지 않으면 묻지 않는다.
2. 입력을 정규화 후 각 훅의 상태 파일에 동시 기록.
3. 활성화 결과를 **항목별로** 보고한다 ("1) ... 2) ..."). 무엇이 켜졌는지 사용자가 즉시 검증 가능.
4. 부분 비활성화 경로도 안내 (예: "freeze만 끄려면 `/unfreeze`").

```
✓ Guard mode active. Two protections are now running:
  1. Destructive command warnings — rm -rf, DROP TABLE, ... will warn (overridable)
  2. Edit boundary — file edits restricted to <path>/. Outside edits blocked.
To remove the edit boundary: /unfreeze. To deactivate everything: end the session.
```

## Deactivation UX

세션 종료가 기본 deactivation. 명시 명령은 **부분 해제만** 제공한다.

- `/unfreeze` → freeze 상태 파일만 삭제. careful은 유지.
- 전체 해제는 별도 명령을 두지 않거나, 두더라도 각 상태 파일을 *원자적으로* 삭제.
- 삭제 실패 시 "어떤 훅이 여전히 활성"인지 명시. 조용히 부분 성공으로 끝내지 않는다.

원자적 tear-down 패턴:
```bash
# 모든 상태 파일을 한 번의 트랜잭션처럼: temp dir로 옮긴 뒤 정리
TMP=$(mktemp -d)
mv "${HOOK_STATE_DIR}"/{careful-overrides,freeze-dir,guard-active}.txt "$TMP"/ 2>/dev/null || true
rm -rf "$TMP"
```

*이유: 일부 훅만 꺼지고 나머지가 살아있으면 사용자는 "보호됨"으로 오인한다(false safety).*

## 동반 스킬
- `guard-destructive-commands` 스킬 — destructive command 패턴 훅 (이전 이름: careful)
- `freeze-edit-scope` 스킬 — directory-scope edit lock 훅 (이전 이름: freeze)

## buddy Plugin Hook 표준 (확장)

이 스킬은 buddy plugin scope의 **PreToolUse hook 표준**을 정의한다. `design-claude-hooks` 스킬이 hook 일반 설계라면, 본 스킬은 **자동 코드 구현 시스템에 특화된 safety composition**을 처방한다.

### 4-Tier Hook Allowlist 정책

buddy plugin은 hook이 자동 trigger하는 스킬을 4 tier로 분류:

| Tier | 정책 | 대상 스킬 |
|---|---|---|
| **GREEN** | 자동 invoke OK | autoplan, validate-idea, review-scope, critique-plan, audit-security, measure-code-health, sync-release-docs, save/restore-context, define-product-spec, split-work-into-features, triage-work-items, validate-advanced-edge-idea, build-with-tdd, diagnose-bug, dispatch-parallel-agents, auto-create-pr, automate-release-tagging, design-deploy-strategy, design-claude-hooks |
| **YELLOW** | 사용자 명시 invoke만 | review-design, review-devex, review-engineering, review-architecture, apply-builder-ethos, consult-codex, consult-design-system, explore-design-variants, audit-live-devex, summarize-retro, query-feature-registry, setup-quality-gates, assess-business-viability, review-pricing-and-gtm, review-license-and-ip-risk, review-privacy-data-risk, review-ai-safety-liability, review-terms-policy-readiness, design-mcp-server, design-billing-system, design-embedding-search, design-artifact-storage |
| **RED (참조 전용)** | hook 차단 | persist-learning-jsonl, iterate-fix-verify, classify-qa-tiers, run-browser-qa, classify-review-risks, write-changelog, monitor-regressions, benchmark-llm-models, guard-destructive-commands, freeze-edit-scope, compose-safety-mode, detect-install-type, guide-setup-wizard |
| **ARCHIVE** | hook 절대 금지 | route-intent, route-multi-platform, route-spec-to-code |

### Layered Safety Composition (자동화 시스템 안전 강제)

자동 코드 구현 시스템에서 다음 6 layer가 **반드시** 활성:

1. **Layer 1: Destructive command guard** (`guard-destructive-commands`)
   - rm -rf, DROP TABLE, force push, kubectl delete 등 8 패턴 + 7 safe exception
   - matcher: Bash
   - decision: ask (user confirm 강제)

2. **Layer 2: Edit-scope freeze** (`freeze-edit-scope`)
   - 활성 worker가 자기 worktree 외부 Edit/Write 차단
   - matcher: Edit | Write | NotebookEdit
   - decision: deny (worker scope 외)

3. **Layer 3: Secret commit block** (gitleaks 통합)
   - API key / private key / token commit 차단
   - matcher: Bash (git commit / git push)
   - decision: deny if secret detected

4. **Layer 4: Force push protection**
   - main / release/*  branch 대상 `git push --force` 차단
   - matcher: Bash (git push)
   - decision: deny + suggest --force-with-lease

5. **Layer 5: Production deploy gate**
   - production deploy 명령은 명시 confirm + reviewer 2명+ 승인
   - matcher: Bash (deploy / kubectl apply / terraform apply / npm publish)
   - decision: ask (multi-step approval)

6. **Layer 6: Skill allowlist enforcement**
   - ARCHIVE / RED tier skill을 hook이 자동 invoke 시도 시 차단
   - matcher: Skill tool
   - decision: deny + redirect to GREEN tier

### Composition 실행 순서

`PreToolUse` matcher 매칭 시 **OR-deny** semantics로 layer 순차 실행:
- Layer 1 → 2 → 3 → 4 → 5 → 6 순서 (먼저 fail-fast 가능 layer 우선)
- 1개라도 deny → 전체 deny
- 1개라도 ask → ask (user confirm)
- 모두 allow → allow

각 layer는 독립 hook script (single source of truth 원칙). 본 스킬은 **합성 manifest**만 제공.

### settings.json 권장 구성 (buddy plugin)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/guard-destructive.sh"
          },
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/secret-commit-block.sh"
          },
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/force-push-protect.sh"
          },
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/deploy-gate.sh"
          }
        ]
      },
      {
        "matcher": "Edit|Write|NotebookEdit",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/freeze-edit-scope.sh"
          }
        ]
      },
      {
        "matcher": "Skill",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/skill-allowlist.sh"
          }
        ]
      }
    ]
  }
}
```

### Activation UX

buddy plugin 설치 시 default `safe` mode 활성. user가 명시적으로 비활성하지 않는 한 6 layer 모두 ON.

```
$ buddy install
✓ buddy plugin installed
✓ safe mode active by default
  Layer 1: destructive command guard       [ON]
  Layer 2: edit-scope freeze (workspace)   [STANDBY — activated per-worker]
  Layer 3: secret commit block             [ON]
  Layer 4: force push protection (main)    [ON]
  Layer 5: production deploy gate          [ON]
  Layer 6: skill allowlist                 [ON]

To deactivate: buddy safe-mode off (not recommended)
To customize per-layer: buddy safe-mode config
```

### Audit Log

모든 layer가 audit JSONL에 기록 (`~/.claude/buddy/safe-mode-audit.jsonl`):
- timestamp / layer / tool / decision / reason / latency

False positive ratio 모니터링 (목표 <1%). >5%면 layer 정책 재검토.
