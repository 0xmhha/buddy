# Claude Code settings.json Hook 스키마

`~/.claude/settings.json` (user-global) 또는 `.claude/settings.json` (project)의
hooks 섹션 완전 명세.

## 전체 구조

```json
{
  "hooks": {
    "PreToolUse":   [ <HookGroup>, ... ],
    "PostToolUse":  [ <HookGroup>, ... ],
    "Stop":         [ <HookGroup>, ... ],
    "SessionStart": [ <HookGroup>, ... ]
  }
}
```

## HookGroup

```json
{
  "matcher": "<regex>",
  "hooks": [ <HookEntry>, ... ]
}
```

| 필드 | 필수 | 설명 |
|------|------|------|
| `matcher` | 선택 | tool_name에 매칭할 정규식. 없으면 모든 tool에 적용. |
| `hooks` | 필수 | HookEntry 배열. 순서대로 실행. |

## HookEntry

```json
{
  "type": "command",
  "command": "/path/to/script.sh [args]"
}
```

| 필드 | 값 | 설명 |
|------|-----|------|
| `type` | `"command"` | 현재 유일한 타입 |
| `command` | string | 실행할 shell 명령 (PATH + 환경변수 사용 가능) |

## 전체 예시 (buddy 통합)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          { "type": "command", "command": "buddy hook-wrap freeze-scope" }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          { "type": "command", "command": "buddy hook-wrap guard-destructive" },
          { "type": "command", "command": "buddy hook-wrap secret-scan" }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          { "type": "command", "command": "buddy hook-wrap run-tests" }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "buddy hook-wrap save-context" }
        ]
      }
    ],
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "buddy hook-wrap restore-context" }
        ]
      }
    ]
  }
}
```

## 우선순위 — user vs project

| 파일 위치 | 적용 범위 | 우선순위 |
|----------|----------|----------|
| `~/.claude/settings.json` | 모든 세션 | 낮음 |
| `.claude/settings.json` (프로젝트) | 해당 프로젝트 세션만 | 높음 (덮어씀) |

두 파일 모두 존재할 때 project 설정이 user-global을 **덮어쓴다** (merge가 아님).
buddy `apply-claude-hooks.sh --scope project`는 프로젝트 파일에만 기록.

## 환경변수

hook command 안에서 사용 가능:

| 변수 | 설명 |
|------|------|
| `CLAUDE_SESSION_ID` | 현재 세션 ID |
| `CLAUDE_TRANSCRIPT_PATH` | 세션 트랜스크립트 절대 경로 |
| `HOME` | 사용자 홈 디렉토리 |
| `PATH` | 기본 PATH (claude가 상속) |

buddy hook-wrap 전용:

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `BUDDY_FREEZE_STATE` | `~/.claude/freeze-scope.txt` | freeze-scope 상태 파일 |
| `BUDDY_CONTEXT_DIR` | `~/.claude/contexts/` | save-context 저장 디렉토리 |
| `BUDDY_AUDIT_LOG` | `~/.claude/hooks/audit.jsonl` | 감사 로그 경로 |
| `BUDDY_DB` | `~/.buddy/buddy.db` | buddy SQLite DB |

## Composition 규칙

같은 matcher에 hook이 여러 개일 때:

```
HookGroup.hooks = [A, B, C]
실행 순서: A → B → C (순차)

결과 결합 (PreToolUse):
- 모두 allow → allow
- 1개 이상 ask → ask (가장 먼저 ask한 메시지 표시)
- 1개 이상 deny → deny (가장 먼저 deny한 메시지 표시)
```

## Fail-safe 정책

hook script가 crash (non-zero exit + no valid JSON)하면:

| 설정 없을 때 기본값 | Claude Code 동작 |
|-----------------|----------------|
| PreToolUse crash | allow (통과) |
| PostToolUse crash | 무시 |
| Stop crash | 무시 |
| SessionStart crash | 무시 |

보안 강화가 필요하면 PreToolUse hook 최상단에 fail-safe를 deny로 설정:

```bash
# crash 시 deny로 fail-safe
trap 'jq -n "{permissionDecision:\"deny\",message:\"[hook error] guard crashed\"}" ; exit 0' ERR
```
