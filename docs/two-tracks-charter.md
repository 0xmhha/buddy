# Buddy — Two Tracks Charter

> **목적**: `buddy` 한 repo 가 두 개의 독립 자산을 담고 있다는 사실 + 각 자산의 정체성 / 책임 경계 / 진화 방향을 영구화한다. 이후 모든 작업 (skill 추가, code 작성, release, dogfood feedback 분류) 은 *어느 트랙에 속하는지* 를 먼저 판단하고 진행한다.
>
> **작성 시점**: 2026-05-10 / **상태**: 초안 (사용자 확정 후 lock-in)

---

## 1. 한 줄 요약

| 트랙 | 한 줄 정의 |
|------|----------|
| **plugin buddy** | Claude Code plugin — skill / MCP / agent / hook 통합 카탈로그. 사용자 전체 product lifecycle 작업을 *창의적 + 효율적 + 안정적* 으로 지원 |
| **cli buddy** | TUI 화면 기반 자동화 agent 관리 툴 — plugin buddy 를 *내재화* 해서 agent 가 작업을 자동 실행하도록 지원 |

---

## 2. plugin buddy — 사용자 발화 기반 정의

### 2.1 정체성 (what is it)

- Claude Code 의 **plugin** 으로 설치되어 동작 (`claude plugin install buddy@buddy`)
- 별도 binary CLI 가 아니라 Claude Code 안에서 *사용자 작업을 보조하는 자산 묶음*
- 설치 시 4 가지 자산이 함께 사용자 환경에 들어감:
  - **skill** — 작업 절차 (어떻게 진행하는지 의 표준 흐름)
  - **MCP** (Model Context Protocol) — Claude 가 외부 도구를 호출하는 표준 인터페이스
  - **agent** — 특정 책임을 가진 sub-agent (예: code reviewer, designer)
  - **hook** — Claude Code 이벤트 (PreToolUse / PostToolUse / Stop 등) 시점에 실행되는 자동 절차

### 2.2 핵심 가치 (why does it exist)

> [사용자 발화 그대로 인용]
>
> "plugin 설치로 인하여, mcp, agent, hook 도 추가로 설치되어, 'buddy' 프로젝트의 스킬들을 작업에 활용할수 있도록 하면서, 그 작업시에 mcp 와 agent, hook 을 활용하여 매우 창의적이고 효율적이며 안정적으로 작업이 진행될수 있도록 하는것이 첫번째 목표"

세 키워드: **창의적 + 효율적 + 안정적**.

### 2.3 Scope (전체 product lifecycle)

> [사용자 발화 그대로 인용]
>
> "Scope 은 idea 를 디벨롭 하고, 사업성을 분석하여, 구현을 위해 앱 혹은 web 을 결정하여 디자인을 적용하고, 설계등을 진행하여 서비스를 빌딩 하고, 자동화 테스트와 서비스 배포, 사용자 사용성 분석을 위한 A/B 테스트, 그로스 해킹, 마케팅 지원, 유지보수 지원을 효율적으로 처리하도록 지원"

| Stage | 영역 |
|-------|------|
| 1 | idea 발굴 + 디벨롭 |
| 2 | 사업성 분석 |
| 3 | 앱 vs web 구현 형태 결정 |
| 4 | 디자인 적용 |
| 5 | 설계 |
| 6 | 서비스 빌딩 |
| 7 | 자동화 테스트 |
| 8 | 서비스 배포 |
| 9 | A/B 테스트 (사용성 분석) |
| 10 | 그로스 해킹 |
| 11 | 마케팅 지원 |
| 12 | 유지보수 지원 |

> 위 12 stage 는 기존 `docs/archive/skill-map.md` 의 11-stage 모델 및 `plugin/skills/router/references/skill-catalog.md` 의 9-phase orchestrator 와 *유사* 하지만 1:1 매핑은 별도 작업. 본 charter 는 *Scope 의 외연* 을 정의.

### 2.4 구성 요소 — 현재 보유

| 구성 요소 | 위치 | 현재 수량 |
|----------|------|---------|
| skill (PROCEDURE.md) | `plugin/skills/<name>/PROCEDURE.md` | 105 |
| slash command | `plugin/commands/<name>.md` | 57 |
| MCP server | `plugin/mcp/` | 미정 (확인 필요) |
| agent | `plugin/agents/` | 미정 (확인 필요) |
| hook | `plugin/hooks/` | 미정 (확인 필요) |

> "skill 105 / command 57" 은 `docs/HANDOFF.md` 의 Last Updated v1.0.8 시점 트랙 상태 표 인용.

### 2.5 현재 상태

- **ACTIVE 트랙** — 본 cycle 의 main 작업 영역
- 105 skill 중 *상당수 구현 완료*, *일부 미구현* (`docs/archive/tasks.md` §A-2 의 잔여 29 skill — 단 사용자 발화에서 "@docs/ 하위 문서들 추가 검토하면 도움 됨" 이라고 명시했으므로 본 숫자는 *baseline*, 추가 누락 발견 가능)
- 미구현 skill 보완 시 외부 reference 활용 가능:
  - `/Users/kevin/work/github/aidax-dag/ai-cli/skill/<projects>/`
  - `/Users/kevin/work/github/aidax-dag/ai-cli/agent/<projects>/`
  - `/Users/kevin/work/github/aidax-dag/ai-cli/harness/<projects>/`
  - `/Users/kevin/work/github/aidax-dag/ai-cli/mcp/<projects>/`
- 단 외부에서 *무조건 가져오는 것* 은 금지 — 호환성 / 궁합 점검 필요. 신규 작성도 가능

---

## 3. cli buddy — 사용자 발화 기반 정의

### 3.1 정체성

- Golang 기반 binary CLI 도구 (`bin/buddy`)
- TUI (Terminal User Interface) 화면 지원
- **plugin buddy 를 내재화** 해서 자동화 agent 안에서 활용

### 3.2 핵심 가치

> [사용자 발화 그대로 인용]
>
> "cli buddy 는 어떤 용도냐면, 'plugin buddy' 를 활용하여 자동화 agent 로 동작할수 있도록 지원하는 툴이다. 자동화된 agent 를 관리하고, 실행 및 종료 시키고, 설정을 변경하는등을 지원하는 툴이다. cli buddy 는 tui 로 화면을 지원하면서, 여러 자동화된 agent 를 설정하고 관리하는 툴"

### 3.3 Scope

| 책임 | 내용 |
|------|------|
| agent 생성 | 특정 작업 (예: 웹툰 그리기) 을 자동 수행할 agent 정의 + 등록 |
| agent 실행 | 등록된 agent 를 백그라운드에서 자동 실행 |
| agent 종료 | 실행 중인 agent 정지 / 폐기 |
| 설정 변경 | agent 별 파라미터 / 스케줄 / 의존성 설정 수정 |
| 다중 관리 | 여러 agent 를 TUI 한 화면에서 동시 모니터링 / 제어 |

### 3.4 구성 요소 — 현재 보유

| 구성 요소 | 위치 | 현재 상태 |
|----------|------|---------|
| CLI entry | `cmd/buddy/main.go` | v0.1.0 — hook reliability monitor 일부 기능만 구현 |
| daemon | `internal/daemon/`, `internal/aggregator/` | 동작 |
| DB / outbox | `internal/db/`, `internal/schema/` | 동작 |
| 설치 / 진단 / 통계 | `internal/install/`, `internal/diagnose/`, `internal/queries/` | 동작 |
| **TUI** | (미구현) | ❌ — 진짜 목적 (자동화 agent 관리) 의 핵심 부재 |
| **agent runtime** | (미구현) | ❌ — agent 생성 / 실행 / 종료 로직 부재 |
| **plugin buddy 내재화 layer** | (미구현) | ❌ |

### 3.5 사용 예시 — 사용자 발화 그대로

> "웹툰 그리기 agent 를 생성한다고 할때, 'plugin buddy' 기능을 활용하여, buddy cli 가 agent 를 생성하고, agent 를 자동으로 실행해두고 관리한다. agent 는 웹툰의 세계관등의 작업을 정리하고, 매일 연재할 스토리를 정리하고, 그림을 생성하여 webtoon-by-ai 서비스에 특정 시간에 배포하여 작품을 노출하고, 유저들은 이것을 감상할수 있도록 할 수 있다."

도식:

```
[사용자] —> [cli buddy TUI] —> "웹툰 그리기 agent 생성"
                                       │
                                       ▼
              [agent runtime 등록 + 스케줄 설정]
                                       │
                                       ▼ (특정 시간 trigger)
   [agent 실행]
       │
       │ 내부에서 plugin buddy 의 skill / MCP / agent / hook 호출:
       │   - 세계관 정리 skill
       │   - 매일 스토리 정리 skill
       │   - 그림 생성 skill (외부 image gen MCP 호출)
       │   - 배포 skill (webtoon-by-ai service 에 publish)
       │
       ▼
   [webtoon-by-ai 서비스에 작품 노출]
       │
       ▼
   [유저 감상]
```

### 3.6 현재 상태

- **부분 구현 트랙** — v0.1.0 의 hook reliability monitor 가 *cli buddy 의 한 기능* (= 사용자 작업 안정성 보조). 단 *진짜 목표 (자동화 agent 관리)* 의 핵심 (TUI / agent runtime / plugin buddy 내재화) 은 미구현
- 향후 작업 시 *기존 v0.1.0 코드* 와 *agent 관리 layer* 의 책임 경계 / 통합 방식 결정 필요 (별도 spec)

---

## 4. 두 트랙의 관계

### 4.1 의존 방향

```
  ┌──────────────────────────┐
  │       cli buddy          │   (자동화 agent 관리)
  │                          │
  │   ┌──────────────────┐   │
  │   │  plugin buddy    │   │   (skill / MCP / agent / hook 카탈로그)
  │   │  (내재화)        │   │
  │   └──────────────────┘   │
  │                          │
  └──────────────────────────┘
              ▲
              │ 설치 / 호출
              │
        [Claude Code]
              ▲
              │
            [사용자]
```

- cli buddy 는 plugin buddy 를 **내재화**. 즉 cli buddy 가 동작할 때 내부적으로 plugin buddy 의 skill / MCP / agent / hook 을 활용
- plugin buddy 는 *cli buddy 없이도 독립 동작* (Claude Code 사용자가 직접 plugin install 후 사용)
- cli buddy 는 *plugin buddy 없이는 자동화 agent 의 작업 능력 부족* (의존)

### 4.2 책임 경계

| 책임 | plugin buddy | cli buddy |
|------|--------------|-----------|
| 사용자가 직접 작업 진행 | ✅ Claude Code 안에서 skill 호출 | ❌ |
| 자동화 (사용자 개입 없이 실행) | ❌ 사용자 trigger 필요 | ✅ agent 자동 실행 / 스케줄 |
| TUI 화면 | ❌ Claude Code 의 chat UI | ✅ 자체 TUI |
| 다중 instance 관리 | ❌ | ✅ 여러 agent 동시 |
| skill / MCP / agent / hook 카탈로그 | ✅ 본체 | ❌ plugin buddy 의존 |

### 4.3 한 repo 에 공존하는 이유

- 두 트랙 모두 사용자 한 사람의 도구 (개인 도구)
- cli buddy 가 plugin buddy 를 *내재화* 라 같은 source 안에서 개발 / release 하는 게 자연
- 단 *코드 / 문서 / release / dogfood feedback 모두 트랙 별로 분류*. charter 의 가장 큰 의미는 *분류 기준 영구화*

---

## 5. 파일 / 디렉토리 매핑

| 디렉토리 / 파일 | 소속 트랙 |
|----------------|---------|
| `plugin/` | plugin buddy |
| `plugin/skills/` | plugin buddy (skill 카탈로그) |
| `plugin/commands/` | plugin buddy (slash command) |
| `plugin/mcp/` | plugin buddy (MCP server) |
| `plugin/agents/` | plugin buddy (agent 정의) |
| `plugin/hooks/` | plugin buddy (hook 정의) |
| `plugin/.claude-plugin/` | plugin buddy (plugin manifest) |
| `cmd/buddy/` | cli buddy (CLI entry + subcommand) |
| `cmd/buddy-mcp/` | 두 트랙 공유 (현재 — 추후 plugin buddy 로 흡수 검토) |
| `internal/db/`, `internal/schema/` | cli buddy (state store) |
| `internal/daemon/`, `internal/aggregator/` | cli buddy (background runtime) |
| `internal/install/`, `internal/diagnose/` | cli buddy (CLI subcommand 구현) |
| `internal/sessions/`, `internal/pricing/` | cli buddy (v0.1 시점에 작성된 이번 cycle 산출, 진짜 cli buddy 목적과의 관계는 추후 평가) |
| `internal/persona/` | cli buddy (CLI 사용자 메시지 카탈로그) |
| `archive/ts-poc/` | cli buddy (이전 TS PoC 자산, 보존용) |
| `docs/HANDOFF.md`, `docs/archive/v0.1-spec.md`, `docs/archive/roadmap.md` | cli buddy (Go CLI 트랙 SSoT) |
| `docs/archive/skill-map.md`, `docs/archive/tasks.md` §A-2 | plugin buddy (skill 카탈로그 SSoT) |
| `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` | plugin buddy (9-phase 아키텍처 SSoT) |
| `docs/superpowers/decisions/` | plugin buddy (ADR) |
| 본 charter (`docs/two-tracks-charter.md`) | 두 트랙 공통 — repo 전체 governance |
| `docs/response-format-guide.md` | repo 외 — AI 응답 형식 reference (트랙 무관) |

---

## 6. 진화 방향

### 6.1 plugin buddy

| 우선순위 | 작업 | 트리거 |
|---------|------|--------|
| 1 (최우선) | 미구현 skill 보완 — `docs/archive/tasks.md` §A-2 의 잔여 + `@docs/` 추가 검토 발견 항목 | 이번 cycle |
| 2 | 외부 reference (`aidax-dag/ai-cli/{skill,agent,harness,mcp}/`) 호환성 점검 → 적합한 것만 통합 | 1번 진행 중 |
| 3 | MCP / agent / hook 자산 확장 (현재 plugin/mcp, plugin/agents, plugin/hooks 보유 자산 확인 후) | 1, 2 종료 후 |
| 4 | skill 간 cascade / dispatch 일관성 (단어 / PROCEDURE 양식 / 의존 그래프) | 지속 |

### 6.2 cli buddy

| 우선순위 | 작업 | 트리거 |
|---------|------|--------|
| 1 | 진짜 목적 (자동화 agent 관리) 의 spec 작성 — `docs/archive/cli-buddy-spec.md` (가칭) | plugin buddy 1번 종료 후 |
| 2 | TUI / agent runtime / plugin buddy 내재화 layer 설계 | 1번 종료 후 |
| 3 | 기존 v0.1.0 (hook reliability monitor) 을 cli buddy 의 sub-feature 로 재배치 또는 별개 유지 결정 | 1번 진행 중 |
| 4 | 웹툰 agent 같은 사용 예시를 reference implementation 으로 작성 | 2번 종료 후 |

### 6.3 호환성 우선 원칙

> [사용자 발화 그대로 인용]
>
> "이 수많은 skill 이 겹치는 부분이 너무 많고, 조금씩 아쉬운 부분이 존재"
>
> "여기에 존재하는 스킬들중 우리 프로젝트로 가져와서 사용함으로써, 스킬능력이 오히려 떨어질수 있는 부분도 존재하니 무조건 가져올것이 아니라, 우리의 'buddy' 프로젝트 스킬들과 호환 및 궁합이 좋은지 점검이 필요하다"

→ **모든 skill 추가 / 변경 / 통합 결정은 *호환성 점검* 을 선행한다**. 단순 양적 확장이 아니라 *겹침 제거 + 호환 보장* 이 우선.

---

## 7. Charter 변경 정책

본 charter 는 *프로젝트 정체성 lock-in* 문서. 변경 조건:

| 변경 종류 | 절차 |
|----------|------|
| 단어 / 정의 미세 수정 | PR 만 — 본 charter 직접 수정 |
| 트랙 추가 / 폐지 | 사용자 명시 결정 → ADR (`docs/superpowers/decisions/<date>-track-change.md`) → charter 갱신 |
| 한 트랙의 핵심 책임 변경 (예: cli buddy 가 *agent 관리 외* 책임 추가) | 사용자 명시 결정 → ADR → charter 갱신 |

> 본 charter 와 다른 문서 (`README.md`, `HANDOFF.md`, `archive/roadmap.md`, `archive/tasks.md`) 가 충돌할 때 — **본 charter 가 SSoT**. 다른 문서를 charter 에 맞춰 갱신.

---

## 8. 즉시 발생하는 후속 작업

본 charter lock-in 시점에 *동시에 진행 필요* 한 항목:

1. `README.md` — 현재 "9-phase lifecycle orchestrator, 57 commands, 105 skills behind one router" 문구가 plugin buddy 만 반영. cli buddy 의 진짜 목적 (자동화 agent 관리) 추가
2. `docs/HANDOFF.md` §0 트랙 상태 표 — 현재 "Plugin / Go CLI" 분류가 본 charter 의 "plugin buddy / cli buddy" 와 단어 정합. 단어 통일 갱신
3. `docs/archive/roadmap.md` — Go CLI 트랙 v0.2 / v0.3 / v1.0 outline 이 *cli buddy 의 진짜 목적* (TUI agent 관리) 과 정합 여부 재평가. v0.2 = "Control Plane multi-session dashboard" 가 agent 관리의 일부인지 / 별도 기능인지 결정
4. `docs/archive/tasks.md` C-2 (Module path drift) — 본 charter 직전 commit `4ce3ccb` 으로 완료, 항목 제거

> 위 4 항목은 charter 와 *별도 commit* 으로 진행. 이 commit 은 charter 자체만.
