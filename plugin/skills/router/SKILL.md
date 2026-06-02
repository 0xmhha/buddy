---
name: router
description: "Use when a buddy command requests dispatch to a target PROCEDURE. Reads `${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` and executes its instructions."
type: skill
---

# Buddy Router

Buddy plugin 내부 라우터. 모든 `/buddy:*` slash command가 이 skill을 호출하며, command가 전달한 `target PROCEDURE` 이름으로 해당 절차 파일을 읽어 그대로 수행한다.

## Dispatch contract

각 buddy command md 의 본문은 router 호출 시 다음을 명시한다:
- `mode:` `single` (기본) / `chain` / `parallel`
- `target PROCEDURE:` (single 모드) — 1 개 skill 이름
- `targets:` (chain·parallel 모드) — 콤마 분리 skill 이름 리스트
- 사용자 인자: `$ARGUMENTS` (command 가 받은 원본 인자 그대로)

router 는 이 4 개 필드를 command md 의 invocation block 에서 읽고, 그 외 frontmatter 는 참조하지 않는다.

## Path resolution

router 가 사용하는 `${CLAUDE_PLUGIN_ROOT}` 는 buddy plugin install root 의 placeholder 다. 해당 경로 `Read` 시 다음 순서로 resolve 한다:

1. `Read ${CLAUDE_PLUGIN_ROOT}/...` 를 그대로 시도. Claude Code runtime 이 placeholder 를 install path 로 substitute 하는 환경에서는 이 시도로 충분하다.
2. 1차 시도가 실패하거나 literal `${CLAUDE_PLUGIN_ROOT}` 가 그대로 노출되어 Read 가 안 풀리면 `Bash` 도구로 install root 를 발견한다:
    ```bash
    {
      find "$HOME/.claude/plugins" -maxdepth 4 -type d -name buddy 2>/dev/null \
        | xargs -I{} sh -c 'test -f "{}/skills/router/SKILL.md" && echo "{}"'
      git -C . rev-parse --show-toplevel 2>/dev/null \
        | xargs -I{} sh -c 'test -f "{}/plugin/skills/router/SKILL.md" && echo "{}/plugin"'
    } | head -1
    ```
    출력된 경로를 BUDDY_ROOT 로 잡고, 본문에 등장하는 `${CLAUDE_PLUGIN_ROOT}/...` 의 `${CLAUDE_PLUGIN_ROOT}` 부분을 BUDDY_ROOT 로 치환해 동일 상대 경로로 재시도한다.
3. 두 시도 모두 실패하면 사용자에게 buddy plugin 설치 경로 확인을 요청하고 중단한다 — 임의 추론·다른 skill 대체 금지.

이 절차는 router 본문에 등장하는 모든 `${CLAUDE_PLUGIN_ROOT}/...` 경로(예: `skills/<target>/PROCEDURE.md`, `skills/router/references/skill-catalog.md`, `skills/router/references/routing-rules.md`)에 동일하게 적용된다.

## How to dispatch (single)

입력으로 다음을 받는다:

- `target PROCEDURE: <name>` — 실행할 절차의 이름 (Buddy skill 디렉토리 이름과 동일)
- 사용자 인자 — command가 그대로 전달한 `$ARGUMENTS` 본문

수행 절차:

1. `Read` 도구로 `${CLAUDE_PLUGIN_ROOT}/skills/<name>/PROCEDURE.md` 파일을 로드한다. (`${CLAUDE_PLUGIN_ROOT}` 는 buddy plugin 설치 경로 — hardcode 금지)
2. 로드한 PROCEDURE 본문을 그 자체로 실행 지시문으로 취급한다 — 사용자 인자를 PROCEDURE의 입력으로 사용한다.
3. PROCEDURE가 요구하는 모든 단계를 누락 없이 수행한다.
4. PROCEDURE 파일이 존재하지 않으면 즉시 에러 메시지를 반환하고 임의로 추론하지 않는다.

## How to dispatch (chain mode)

Trigger: command md 가 `mode: chain` 과 `targets: name1, name2, name3` (또는 동등한 콤마 구분 형식)을 전달한 경우.

수행 절차:

1. `targets` 문자열을 콤마(`,`)로 분리해 순서 보존 리스트로 만든다. 각 이름의 앞뒤 공백을 trim 한다. 빈 토큰은 제거한다.
2. 리스트의 각 target 에 대해 **순서대로** 다음을 수행한다:
   1. `Read` 도구로 `${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` 를 로드한다.
   2. 로드한 PROCEDURE 본문을 실행한다. 입력 컨텍스트로는 (a) 사용자 인자 원본 + (b) 직전까지 실행된 모든 step 의 산출물을 함께 제공한다.
   3. 해당 step 이 만든 산출물(요약·결정·생성 파일 경로 등)을 현재 대화 컨텍스트에 보존해, 다음 target 이 이를 입력으로 참조 가능하게 한다.
      - 만약 PROCEDURE 가 명시 산출물을 정의하지 않으면(side-effecting skill: `auto-create-pr`, `ship-release` 등), 실행 결과 한 줄 요약(성공 / 실패 + 변경된 외부 상태 식별자: PR URL, git tag, 배포 ID 등)을 산출물로 간주한다.
3. 어느 target 의 PROCEDURE.md 가 존재하지 않으면 즉시 중단하고 어떤 target 이 실패했는지 명확히 보고한다 — 임의 추론·skip 금지.
4. 마지막 target 까지 완료되면, 각 step 이 무엇을 산출했는지 한 줄씩 요약한 최종 리포트를 출력한다 (step 순서 유지).

## How to dispatch (parallel mode)

Trigger: command md 가 `mode: parallel` 과 `targets: name1, name2, name3` 을 전달한 경우.

**왜 parent-reads pattern**: fresh subagent 는 부모의 runtime 권한 grant 를 상속하지 않는다 (Claude Code security 설계 — parent 가 subagent 에게 임의 권한 escalation 못 함). 따라서 `${CLAUDE_PLUGIN_ROOT}/skills/...` 같이 user 가 parent 에게만 승인한 경로는 subagent 가 직접 Read 할 수 없다. router (parent) 가 모든 PROCEDURE.md 를 읽고 본문을 prompt 에 embed 해 dispatch.

수행 절차:

1. `targets` 문자열을 콤마(`,`)로 분리해 리스트로 만든다. 각 이름의 앞뒤 공백을 trim 한다. 빈 토큰은 제거한다.
2. **router 가 직접 각 target 의 PROCEDURE 를 Read** 한다 (parent 권한 사용):
   - 각 target 에 대해 `Read ${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` 호출
   - 결과 본문을 메모리에 보관 (subagent prompt 에 embed 할 용도)
   - PROCEDURE 가 부재하면 그 target 만 "missing PROCEDURE" 로 표기, 나머지는 진행 — 전체 중단 금지
3. 각 target 마다 `Agent` 도구로 fresh subagent 를 하나씩 디스패치 (`subagent_type: general-purpose`). 각 subagent 에게 다음을 전달한다:
   - **PROCEDURE 본문 (parent 가 step 2 에서 읽은 것)** 을 prompt 안에 직접 embed
   - 지시: "다음은 `<target>` skill 의 PROCEDURE 본문이다. 이를 그대로 실행 지시문으로 취급해 모든 단계를 수행하라."
   - 공유 사용자 인자 (원본 `$ARGUMENTS`)
   - 제약: subagent 는 file Read 시도 금지 — 모든 절차는 prompt 안에 들어 있음. 다른 target 의 작업물·파일 수정 금지. 자기 결과만 보고로 반환.
4. 모든 subagent 가 완료될 때까지 대기. 각 subagent 는 자기 PROCEDURE 의 실행 리포트를 반환.
5. 결과 집계: target 별로 그룹핑해 결과 제시. 그 후 cross-target 관찰 (상호 모순, 공통 finding, 시너지) 이 있으면 별도 단락으로 합성.
6. step 2 에서 missing PROCEDURE 로 표기된 target 들은 결과 섹션 끝에 명시 — 전체 중단 금지.

**Token 비용 주의**: PROCEDURE 본문이 큰 경우 (예: review-engineering 700+ lines) subagent dispatch prompt 가 그만큼 커진다. parallel mode 는 원래 review/audit 처럼 동시 다발 분석에 적합한 패턴이라 이 비용은 의도된 trade-off 임. 만약 dispatch 비용이 일관되게 부담된다면 user 의 `~/.claude/settings.json` 에 `permissions.allow: ["Read(${CLAUDE_PLUGIN_ROOT}/**)"]` 추가로 subagent 가 직접 Read 가능하게 만들 수 있고, 이 경우 router 가 step 2 를 skip 해도 된다 (advanced setup, plugin install 만으로는 자동 안됨).

## Skill index

Buddy 는 78 개 skill 을 **artifact 의존성 그래프(DAG)** 로 조직한다. 아래 §1~§9 는 그래프를 사람이 읽기 쉽게 묶은 **클러스터 라벨**이며 강제 실행 순서가 아니다 — 진입점은 "지금 존재하는 artifact 의 frontier" 로 결정된다. 각 클러스터는 진입점 orchestrator 와 그 안의 stage skill 집합을 가진다.

라이프사이클 정의·노드 명사·DoR/DoD 계약의 SSoT 는 [`references/se-lifecycle-naming.md`](./references/se-lifecycle-naming.md) (phase 정체성은 [`references/engineering-phases.md`](./references/engineering-phases.md)) 다.

아래 표의 마지막 컬럼은 **DoR(좌변 input) → DoD(우변 output)** 계약이다. 노드는 자신이 생산하는 산출물(output)로 명명되므로, 진입에 필요한 input 은 그래프 직전 노드의 output 과 같다.

| 클러스터 (DAG 노드) | Orchestrator (entry) | DoR (input) → DoD (output) |
|-------|----------------------|---------|
| §1 Discovery / Impact Analysis | `concretize-idea` / `assess-product-change` | idea/concept → PRD + 사업성 검증 (Mode A) · change request + codebase → 영향 평가 + scope (Mode B) |
| §2 Requirements Specification | `define-features` | PRD → actor / use case / system boundary → feature backlog (SRS) |
| §3 Software Design | `design-system` | feature backlog (SRS) → tech stack ADR + infra + API + data model (SDD) |
| §4 Iteration Planning | `plan-build` | software design (SDD) → actor 별 task graph + 병렬 실행 plan |
| §5 Construction | `build-feature` | iteration plan → working code + tests (TDD + parallel agents) |
| §6 Verification & Validation | `verify-quality` | code complete → QA + security + compliance sign-off (V&V evidence) |
| §7 Release & Deployment | `ship-release` | quality gate pass → tagged release + UAT + GA (beta 포함) |
| §8 Operation & Maintenance | `iterate-product` | production traffic → A/B + funnel + improvement backlog |
| §9 Retirement / Decommissioning | `manage-lifecycle` | usage data + 폐기 결정 → deprecation + migration + EOL plan |

Cross-phase 보조:

- `autoplan` — 어느 노드의 산출물(plan/PRD/ADR/task plan)에든 호출 가능한 4-mode review (review-scope → review-engineering → review-design → review-devex 순차).
- `consult-codex`, `save-context`, `restore-context` — 노드 종속 없는 공통 도구.
- `status` — 현재 존재하는 artifact 를 탐지해 진입 노드(frontier)를 추론.

라우팅 결정 워크플로우 (artifact frontier 기반):

1. **정확 매칭 (fast path)**: command name 또는 사용자 발화가 위 표의 entry-point skill 1개와 정확히 매칭되면 그 노드를 dispatch — 인라인 표만으로 충분.
2. **매칭 없으면 frontier 로 진입 노드 결정**: 현재 존재하는 artifact 를 보고(필요 시 `status`) 어느 노드까지 DoD 가 채워졌는지 판단해 그 다음 노드를 진입점으로 둔다.
3. **DoR 충족 검사 (prerequisite gate)**: 진입하려는 노드의 DoR(required input artifact)이 없으면, 그것을 생산하는 **upstream 노드로 자동 선행**한다 (예: Software Design 요청인데 SRS 부재 → 먼저 §2). 이는 backtrack·skip·scope-routing 을 아우르는 단일 규칙이다.
4. **stage 단독 명시 존중**: 사용자가 stage skill 명을 직접 지정하면 orchestrator 로 escalate 하지 않는다 (User Sovereignty).
5. **lazy-load 트리거**:
   - 위 인라인 표로 노드가 정해지면 추가 Read 불필요.
   - 노드 내 stage·domain·pattern skill 이 필요하거나 entry-point 가 아닌 target 이면 → `Read ${CLAUDE_PLUGIN_ROOT}/skills/router/references/skill-catalog.md` 로 전체 카탈로그 확인.
   - 2개 이상 skill 사이에서 모호하면 → `Read ${CLAUDE_PLUGIN_ROOT}/skills/router/references/routing-rules.md` §3 케이스별 결정 참조.
