---
name: route-multi-platform
description: "[ARCHIVE / 참조 전용] LLM이 직접 invoke 금지. command-routing 패턴이 흡수했으므로 historical reference로만 유지. 원본 의도: multiple AI agents 또는 developer tools와 통합할 때 per-host path/credential map과 scoped tokens로 platform-specific routing을 수행하는 패턴."
type: skill
---

# Snippet: Multi-Platform Routing Pattern

> **⚠️ ARCHIVE NOTICE**
>
> 이 스킬은 **참조 전용 라우팅 패턴 reference**입니다. LLM이 직접 invoke하지 마세요.
> 라우팅 의도는 `command-routing` 패턴(SKILLS_CLEANUP_PLAN §6.2 D2)에 흡수되었으며,
> 현재 buddy plugin은 4 핵심(autoplan/validate-idea/review-scope/critique-plan)이 라우팅 책임을 가집니다.
>
> 본 본문은 향후 라우팅 정책을 재설계할 때의 historical reference로만 유지됩니다.

> 호스트별 매핑 + scoped 토큰 라우팅 패턴 — pairing 흐름 등 도구별 plumbing 제외.

## 이 snippet을 사용하는 경우

- 여러 AI 에이전트와 통합하는 도구 (OpenClaw, Codex, Cursor 등)
- 각 호스트가 자체 config 디렉토리를 갖는 cross-tool credential resolution
- 곳곳에 하드코딩하지 않고 호스트별 경로가 필요한 플러그인 시스템
- 호스트별 instruction 블록(curl, exec, shell)을 emit하는 모든 CLI

## 패턴 1: 플랫폼 감지

중간에 추측하지 말고 **target host**를 first-class 변수로 해결한다. 신호를 조합 — 모호할 땐 사용자에게 선택을 맡긴다:

1. **명시적 사용자 선택 (권장).** 한 번 묻고, 답을 저장한다.
   매핑 예시 (single source of truth):
   - `A → openclaw`
   - `B → codex`
   - `C → cursor`
   - `D → claude`
   - `E → generic` (호스트별 config 없음; raw HTTP 출력)
2. **Env var 오버라이드.** `TARGET_HOST=openclaw`는 prompt를 건너뛴다.
3. **파일시스템 신호.** `~/.openclaw/`, `~/.codex/`, `~/.cursor/` 존재 여부로 어떤 호스트가 로컬에 설치됐는지 추정.
4. **부모 프로세스 / invocation context.** 최후의 수단; 취약함.

`TARGET_HOST`가 설정되면 이후 모든 단계가 이를 통해 라우팅된다. 중간에 재감지 금지 — workflow-scoped 상수다.

## 패턴 2: 경로 매핑 테이블

하나의 테이블 = 하나의 source of truth. 스킬 prose, 코드, 테스트 모두 같은 매핑을 참조한다.

| Platform   | Config dir       | 호스트별 자격증명 경로                              |
|------------|------------------|----------------------------------------------------|
| openclaw   | `~/.openclaw/`   | `~/.openclaw/skills/<tool>/<artifact>.json`        |
| codex      | `~/.codex/`      | `~/.codex/skills/<tool>/<artifact>.json`           |
| cursor     | `~/.cursor/`     | `~/.cursor/skills/<tool>/<artifact>.json`          |
| claude     | `~/.claude/`     | `~/.claude/skills/<tool>/<artifact>.json`          |
| generic    | 없음             | copy-paste용 stdout 출력                           |

규칙:
- **호스트 간 같은 형태** — `<host-root>/skills/<tool>/<artifact>.json`. 예측 가능한 형태여야 에이전트가 하나의 규칙으로 자기 자격증명을 찾을 수 있다.
- **Generic은 write가 아니라 print.** 알 수 없는 호스트? 경로를 추측하지 말고 portable instruction 블록(curl + JSON)을 emit한다.
- **workflow 깊숙이에 `~/.openclaw/...`를 하드코딩 금지.** 테이블로 한 번 resolve하고, 경로를 매개변수로 전달한다.

## 패턴 3: Scoped 토큰 생성

장기 유효 공유 비밀은 실패 모드다. 페어링마다 **단기·단일 목적** 토큰을 발급한다.

Two-token 모델:
1. **Setup key** — 일회성, TTL ~5분, 단일 사용. 다른 에이전트가 redeem할 만큼만 copy-paste를 견딘다. 누구도 leak을 악용하기 전에 만료된다.
2. **Session token** — setup key와 교환해 발급. TTL ~24h. 특정 에이전트 identity, capability 집합(read+write vs admin), 선택적 도메인 allowlist로 scope 지정.

강제해야 할 속성:
- **에이전트별 격리.** 토큰에 에이전트 identity가 담긴다; 한 에이전트의 토큰이 다른 에이전트의 탭/리소스를 사칭할 수 없다.
- **Capability scoping.** 기본 = 최소 권한(read+write). Elevated capability(JS 실행, cookie/storage 접근)는 mint 시점에 명시적 `--admin` 플래그 필요, 런타임 escalation은 절대 불가.
- **토큰별 revoke + 전역 rotate.** `revoke <agent>`는 하나의 토큰을 무효화. `rotate`는 부모 비밀을 무효화하고, 그 아래 발급된 모든 scoped 토큰에 cascade.
- **토큰당 rate limit.** ~10 req/s + `Retry-After` 헤더.
- **URL/로그에 토큰 금지.** 헤더 전용(`Authorization: Bearer ...`).

Capability matrix (예시):

| Capability                         | default (read+write) | admin |
|------------------------------------|----------------------|-------|
| Navigate / click / fill / snapshot | yes                  | yes   |
| New tab (자기 탭만)                | yes                  | yes   |
| 다른 에이전트 탭 읽기              | no                   | no    |
| 임의 JS 실행                       | no                   | yes   |
| Cookie / storage 읽기              | no                   | yes   |

"자기 탭만" 제약은 클라이언트를 믿는 게 아니라, 토큰 안의 에이전트 identity를 기준으로 서버 측에서 강제한다.
