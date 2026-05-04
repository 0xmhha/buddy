---
name: guard-destructive-commands
description: "[패턴 라이브러리] rm -rf, DROP TABLE, force push 등 destructive bash command 전 curated risk taxonomy + safe exceptions로 confirmation guard. 직접 invoke보다 orchestrator가 import해 사용. 트리거: 'rm -rf 가드' / 'DROP TABLE 막아' / 'force-push 차단' / 'be careful' / '위험 명령 가드'. 참조 위치: 4 핵심 작업 중, compose-safety-mode 컴포넌트, autoplan/critique-plan 고위험 단계."
type: skill
---

# Careful — 파괴적 명령 가드


파괴적 shell 명령을 위한 안전 모드. 모든 Bash 명령은 실행 전에 curated된 8가지 high-risk 패턴 taxonomy에 대해 체크된다. 매치되면 Claude가 멈추고 사용자에게 confirm 요청. 안전 컨텍스트의 작은 allowlist (disposable 빌드 artifact)가 prompt를 우회해 일반 cleanup이 마찰 없이 유지.

스킬은 **세션 스코프**: 프로덕션, 공유 인프라, 신경 쓰는 git 히스토리를 건드리는 세션에 대해 활성화. 세션 종료로 비활성화.

## 이 스킬을 사용하는 경우

- 프로덕션이나 스테이징 데이터베이스(`psql`, `mysql`, `mongosh`, …) 작업.
- 한 번의 잘못된 키 입력이 상태를 wipe하는 live 시스템 디버깅.
- 히스토리 재작성이 동료를 clobber할 공유 repo 작업.
- 잃을 수 없는 데이터가 있는 호스트에서 cleanup 스크립트 실행.
- Junior 에이전트와 페어링하거나 중요한 것 근처에서 자율 루프 실행.
- 사용자가 "be careful," "safety mode," "prod mode," 또는 "careful mode"라고 말함.

## 8가지 파괴적 패턴

아래 각 패턴은 가드가 체크하는 것, 왜 위험한지, 합법으로 흔히 나타나는 컨텍스트 shape. "common safe contexts" 노트는 진단용, 자동 allowlist 아님 — 다음 섹션의 7개 명시 예외만 자동으로 prompt 우회.

### 1. `rm -rf` (recursive force delete)

- **Pattern**: `rm\s+(-[a-zA-Z]*r|--recursive)` — `-r`, `-rf`, `-fr`, 또는 `--recursive` 있는 모든 `rm`, 플래그 bundle 포함.
- **위험 이유**: Recursive delete는 대부분 파일시스템에서 되돌릴 수 없다. 타겟 경로 오타 하나(`rm -rf /` vs `rm -rf ./`, unset `$VAR`가 empty로 확장)가 undo 없이 데이터 wipe.
- **흔한 safe context**: 빌드 출력 삭제 (`dist/`, `node_modules/`, `.next/`) — 아래 safe 예외로 커버.

### 2. `DROP TABLE` / `DROP DATABASE`

- **Pattern**: case-insensitive `drop\s+(table|database)`.
- **위험 이유**: 스키마와 모든 row를 영구 제거. 대부분 프로덕션 DB는 DDL 전 auto-snapshot 안 함. PITR 있어도 restore는 몇 시간 downtime.
- **흔한 safe context**: 테스트 fixture의 scratch DB, ephemeral CI 컨테이너 — 그러나 프로덕션에서 false negative 비용이 너무 높아 매번 prompt 마찰 가치.

### 3. `TRUNCATE`

- **Pattern**: case-insensitive `\btruncate\b`.
- **위험 이유**: 테이블의 모든 row를 per-row 트리거 발화 없이 삭제, 종종 WAL 엔트리 없이, 종종 애플리케이션 레이어 audit 로그로 catch 불가. 많은 엔진에서 `DELETE`보다 빠르고 *덜* 복구 가능.
- **흔한 safe context**: 격리된 DB에 대한 integration test run 사이 fixture 리셋.

### 4. `git push --force` / `git push -f`

- **Pattern**: `git\s+push\s+.*(-f\b|--force)`. Push 명령 어디든 short/long form 모두 catch (예: `--set-upstream` 뒤).
- **위험 이유**: 리모트 히스토리 재작성. 구 히스토리 pull하고 위에 commit한 동료는 그들의 작업이 브랜치에서 사라진 것을 본다. 공유 브랜치(`main`, `develop`, 릴리스 브랜치)에서 이는 outage.
- **흔한 safe context**: 소유한 solo feature 브랜치, rebase나 squash 후. 그때도 `--force-with-lease`가 더 안전, 자체 collision 체크를 갖고 있어 여기서 플래그 안 됨.

### 5. `git reset --hard`

- **Pattern**: `git\s+reset\s+--hard`.
- **위험 이유**: 워킹 트리의 모든 uncommitted 변경 폐기 AND 브랜치 포인터 이동. Commit이 이미 push됐거나 drop된 object가 `git reflog`에 있지 않은 한(기본 90일, `gc`로 쉽게 손실) undo 없음.
- **흔한 safe context**: 원하지 않는다고 확신하는 반쯤 완성된 실험 버리기 — 그러나 Claude는 확신과 성급함을 구별 못 하므로 질문.

### 6. `git checkout .` / `git restore .`

- **Pattern**: `git\s+(checkout|restore)\s+\.`
- **위험 이유**: 워킹 트리의 모든 uncommitted 변경 폐기. 브랜치 포인터 이동 없는 `reset --hard`와 같은 작업 손실 리스크. 편집이 필요했다는 것을 깨닫기 직전 "cleanup하려고" 종종 실행.
- **흔한 safe context**: Source가 checked-in이고 regeneration이 싼 경우 실패한 codegen run 후 generated-file mess 되돌리기.

### 7. `kubectl delete`

- **Pattern**: `kubectl\s+delete`.
- **위험 이유**: Kubernetes 리소스 제거 — pod, deployment, namespace, PVC. Namespace delete는 안의 모든 workload와 PersistentVolumeClaim에 cascade. 프로덕션에서는 outage; PVC reclaim policy `Delete`면 데이터 손실도.
- **흔한 safe context**: 로컬 kind/minikube 클러스터 tear down, 명확히 명명된 ephemeral 테스트 namespace 삭제. 잘못된 context(`kubectl config current-context`)의 비용이 너무 높아 prompt skip 불가.

### 8. `docker rm -f` / `docker system prune`

- **Pattern**: `docker\s+(rm\s+-f|system\s+prune)`.
- **위험 이유**: `docker rm -f`는 실행 중 포함 컨테이너를 force-stop하고 제거 — persist 안 된 상태 보유 가능. `docker system prune`(특히 `-a`나 `--volumes` 포함)은 이미지, 빌드 캐시, 네트워크 삭제, 선택적 volume 삭제, volume 플래그가 "그냥 로컬에서 돌고 있던" DB를 wipe하는 것.
- **흔한 safe context**: 디스크가 다 찬 dev 머신 cleanup — 그러나 사라질 volume에 중요한 게 없는지 사용자가 confirm해야.

## 7가지 Safe 예외

이들이 prompt 없이 auto-allow되는 유일한 경로. 모두 well-known disposable 빌드 디렉토리에 대한 `rm -rf` 형태. 파괴적 패턴에 매치되는 다른 모든 것은 user-confirm prompt 통과.

예외는 `rm`의 **모든** 경로 인자가 이 타겟 중 하나일 때만 발화. Safe 타겟과 unsafe 타겟 혼합 (`rm -rf node_modules /etc/foo`)도 여전히 prompt 트리거.

### 1. `node_modules`

JavaScript/TypeScript 패키지 디렉토리. `npm install` / `pnpm install` / `bun install`로 재생성. 일상 cleanup 타겟.

### 2. `.next`

Next.js 빌드 캐시와 출력. 다음 `next dev` / `next build`에 재생성.

### 3. `dist`

대부분 bundler(Vite, Rollup, tsc 등)가 쓰는 generic 빌드 출력 디렉토리.

### 4. `__pycache__`

Python bytecode 캐시. 다음 모듈 import에 재생성.

### 5. `.cache`

많은 도구(Babel, ESLint, Parcel, Yarn 등)가 쓰는 generic 캐시 디렉토리.

### 6. `build`

Generic 빌드 출력 (CMake, Gradle, Make, esbuild, …). 다음 빌드 invocation으로 재생성.

### 7. `.turbo`와 `coverage`

Turborepo task 캐시(`.turbo`)와 테스트 coverage 출력 (Jest, Vitest, c8, nyc 등의 `coverage/`). 둘 다 다음 run에 재생성되고 source 포함 안 함.

## Override 메커니즘

파괴적 패턴이 발화하면 훅이 `permissionDecision: "ask"` 응답과 리스크를 명명하는 짧은 `[careful]` 경고를 반환. Claude Code가 명령 실행 전 사용자에게 prompt 표시. 거기서 두 경로:

1. **Confirm** — 명령이 작성된 대로 실행. 경고 인정됨.
2. **Cancel** — Claude가 명령 거부됨을 듣고, 계획 조정 (더 안전한 대안 선택, 명확화 질문, 또는 중단).

"permanently allow" 없음 — 세션 내 모든 파괴적 명령이 매번 prompt. 안전 모드는 load-bearing이어야 하고, 일회 unlock 아님.

가드 전체 비활성화하려면 세션 종료 또는 새로 시작. 훅 등록은 세션 스코프.

## Generic Hook 통합

이 taxonomy를 Claude Code에 wire up하려면 `Bash` 도구에 매치하는 `PreToolUse` 훅 등록. 훅 스크립트는 stdin에서 도구 입력을 JSON으로 읽고, `command` 필드 추출, 패턴 체크 실행, stdout에 JSON 응답 출력.

Claude Code 훅 계약:

- **허용**: `{}` (또는 `{"permissionDecision":"allow"}`) 출력 후 exit 0.
- **사용자에게 ask**: `{"permissionDecision":"ask","message":"..."}` 출력 후 exit 0. 메시지가 confirmation prompt에 표시.
- **완전 거부** (이 스킬엔 드물게 필요): `{"permissionDecision":"deny","message":"..."}`.

### 훅 스켈레톤 예시

예: `~/.claude/skills/careful/bin/check-careful.sh`로 저장 후 `chmod +x`. 이는 generic 템플릿 — logging, 출력 포맷, 예외 리스트를 환경에 맞게 adapt.

```bash
#!/usr/bin/env bash
# PreToolUse hook — warn before destructive Bash commands.
set -euo pipefail

INPUT=$(cat)

# "command" 필드 추출. grep이 99% 케이스 처리; python은 escaped quote 포함
# 명령의 fallback.
CMD=$(printf '%s' "$INPUT" \
  | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' \
  | head -1 | sed 's/.*:[[:space:]]*"//;s/"$//' || true)
if [ -z "$CMD" ]; then
  CMD=$(printf '%s' "$INPUT" \
    | python3 -c 'import sys,json; print(json.loads(sys.stdin.read()).get("tool_input",{}).get("command",""))' \
    2>/dev/null || true)
fi
[ -z "$CMD" ] && { echo '{}'; exit 0; }

CMD_LOWER=$(printf '%s' "$CMD" | tr '[:upper:]' '[:lower:]')

# --- Safe 예외: 알려진 빌드 artifact에 대한 rm -rf만 ---
SAFE_TARGETS='node_modules .next dist __pycache__ .cache build .turbo coverage'
if printf '%s' "$CMD" | grep -qE 'rm\s+(-[a-zA-Z]*r[a-zA-Z]*\s+|--recursive\s+)'; then
  SAFE_ONLY=true
  RM_ARGS=$(printf '%s' "$CMD" | sed -E 's/.*rm\s+(-[a-zA-Z]+\s+)*//;s/--recursive\s*//')
  for target in $RM_ARGS; do
    case "$target" in
      -*) ;;  # flag, skip
      *)
        match=false
        for safe in $SAFE_TARGETS; do
          case "$target" in */"$safe"|"$safe") match=true; break;; esac
        done
        [ "$match" = false ] && { SAFE_ONLY=false; break; }
        ;;
    esac
  done
  [ "$SAFE_ONLY" = true ] && { echo '{}'; exit 0; }
fi

# --- 파괴적 패턴 체크 ---
WARN=""
check() { [ -z "$WARN" ] && printf '%s' "$2" | grep -qE "$1" && WARN="$3"; }

check 'rm\s+(-[a-zA-Z]*r|--recursive)'        "$CMD"       'recursive delete (rm -r) — permanent removal.'
check 'drop\s+(table|database)'                "$CMD_LOWER" 'SQL DROP — permanently deletes database objects.'
check '\btruncate\b'                           "$CMD_LOWER" 'SQL TRUNCATE — deletes all rows from a table.'
check 'git\s+push\s+.*(-f\b|--force)'          "$CMD"       'git force-push — rewrites remote history.'
check 'git\s+reset\s+--hard'                   "$CMD"       'git reset --hard — discards uncommitted work.'
check 'git\s+(checkout|restore)\s+\.'          "$CMD"       'discards all uncommitted changes in working tree.'
check 'kubectl\s+delete'                       "$CMD"       'kubectl delete — removes Kubernetes resources.'
check 'docker\s+(rm\s+-f|system\s+prune)'      "$CMD"       'Docker force-remove / prune — may delete running containers or volumes.'

if [ -n "$WARN" ]; then
  WARN_ESCAPED=$(printf '%s' "$WARN" | sed 's/"/\\"/g')
  printf '{"permissionDecision":"ask","message":"[careful] %s"}\n' "$WARN_ESCAPED"
else
  echo '{}'
fi
```

### Settings.json wire-up

훅을 Claude Code 설정에 등록 (`~/.claude/settings.json` user-scope, 또는 프로젝트의 `.claude/settings.json`). 스크립트를 저장한 경로로 조정.

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/skills/careful/bin/check-careful.sh",
            "statusMessage": "Checking for destructive commands..."
          }
        ]
      }
    ]
  }
}
```

### 훅 검증

등록 후 known-safe와 known-destructive 명령으로 테스트:

```bash
# Prompt 없이 실행돼야
echo "hello"

# [careful] confirmation 트리거돼야
git reset --hard HEAD
```

Prompt가 발화 안 하면 체크: 스크립트 실행 가능 (`chmod +x`), `settings.json`의 경로가 absolute 또는 올바로 expand, 훅이 exit 0 (non-zero exit는 도구 호출을 사용자 prompt 없이 전체 차단).

### Taxonomy 확장

패턴 추가하려면 `check '<regex>' "$CMD" '<warning>'` 라인 더 append. Regex를 `\s+`와 word boundary (`\b`)로 anchor 유지해 substring의 false positive 회피 (예: 관련 없는 식별자 안의 `truncate`).

Safe 예외 추가하려면 디렉토리 이름을 `SAFE_TARGETS`에 append. 보편적으로 재생성 가능하고 절대 source 포함 안 하는 경로에만 이렇게.
