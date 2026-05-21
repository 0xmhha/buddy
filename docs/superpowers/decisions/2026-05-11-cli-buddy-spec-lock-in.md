# ADR-005 — cli buddy spec scope lock-in (Draft → Accepted)

> **Status**: Accepted
> **Date**: 2026-05-11
> **Deciders**: buddy maintainer
> **Tags**: cli-buddy, agent-runtime, embedding, charter-alignment, spec-lock-in
> **Related**:
> - [`docs/archive/cli-buddy-spec.md`](../../archive/cli-buddy-spec.md) — Draft → Accepted by this ADR
> - [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §3 cli buddy 정체성
> - [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002)
> - [`docs/superpowers/decisions/2026-05-11-plugin-version-reset.md`](./2026-05-11-plugin-version-reset.md) (ADR-004)

---

## 1. Context

### 1.1 cli buddy 현재 상태

`docs/cli-buddy-spec.md` (371 lines) 가 charter §6.2 의 *cli buddy 진화 1 순위 작업* 의 산출물로 작성됨 (2026-05-10). 본문 §11 의 다음 액션 1번:

> "본 spec lock-in (사용자 confirm 후 Status = Accepted)"

이 lock-in 이 발생하지 않으면 W3-2 (TUI) / W3-3 (agent runtime) / W3-4 (embedding layer) / W3-5 (v0.1.0 재배치) / W3-6 (reference agent) cascade 가 모두 trigger 대기. plugin buddy 의 v1.0.0 entry condition (ADR-004 §2.2 #1) 도 이 lock-in 에 의존.

### 1.2 lock-in 시점 정당화

plugin buddy v0.3.0 ship (analytics-mcp W4-2.1~W4-2.6) 후 다음 의미 있는 trigger 가 cli buddy 트랙. spec 본문 §4.1 의 4 옵션 중 *권장 (a) Claude Code subprocess* 가 명시 — 본 ADR 은 그 권장을 *공식 결정으로 영속화*.

### 1.3 spec 본문에 이미 권장 lock-in 된 항목

| 영역 | 권장 |
|------|------|
| §1.3 scope | **agent 생성 / 실행 / 종료 / 설정 변경** 4 책임 (사용자 발화 인용) |
| §3.1 4 sub-component | TUI + agent runtime + plugin buddy embedding layer + state store (기존 v0.1.0 SQLite 재사용) |
| §4.1 embedding 패턴 | **(a) Claude Code subprocess** — 단순 + claude code 사용자 가정 + 기존 인증 활용 |
| §5.1 v0.1.0 자산 재배치 | hook reliability monitor 를 cli buddy 의 *sub-feature* 로 재배치 |
| §6 CLI subcommand surface | 기존 v0.1.0 7 subcommand + 신규 7 (tui / agent CRUD / schedule list) |
| §7 v0.2/v0.3/v1.0 outline rewrite | ADR-002 의 actual rewrite — roadmap.md 의 항목들이 *cli buddy 의 sub-feature 로 흡수* |
| §9 Phase 분할 | W3-1 (spec) ~ W3-6 (reference agent) 6 phase, 3-6 month estimate |
| §10 open questions | Q-1 ~ Q-6 (TUI vs web / AI 모델 / log retention / multi-machine / agent 간 통신 / sandbox security) — deferred |

본 ADR 은 위 권장을 *모두 채택* 으로 lock-in.

---

## 2. Decision

### 2.1 cli-buddy-spec.md Status

`docs/cli-buddy-spec.md` 의 Status 를 **`Draft` → `Accepted`** 로 갱신. 본 ADR-005 가 lock-in 근거.

### 2.2 Lock-in 항목 (spec 본문 권장 그대로)

| 항목 | 결정 |
|------|------|
| **§1.3 4-책임 scope** | agent 생성 / 실행 / 종료 / 설정 변경. Non-goals: skill 자산 추가 (plugin buddy 책임), 외부 SaaS 직접 통합, 사용자 데이터 영속 저장, 다중 사용자 협업 |
| **§4.1 embedding pattern** | **(a) Claude Code subprocess** — `claude` CLI 호출 + stdin/stdout 통신 |
| **§5.1 v0.1.0 자산 재배치** | hook reliability monitor (`internal/daemon/`, `internal/aggregator/`) 를 cli buddy 의 *별도 mode* 로 진입. 기존 7 subcommand 호환 유지 (§6.1) |
| **§7 ADR-002 actual rewrite** | roadmap.md v0.2 / v0.3 / v1.0 outline 의 항목들이 *cli buddy 안에서 흡수* — 본 ADR 채택으로 trigger |
| **§9 Phase 분할** | W3-1 (spec, Done) → W3-2 (TUI) → W3-3 (agent runtime) → W3-4 (plugin buddy embedding) → W3-5 (v0.1.0 재배치) → W3-6 (reference agent). 6 phase × 평균 1~3 week = **3~6 month** single-dev cadence |
| **§10 Open questions Q-1~Q-6** | Deferred. trigger 시점 (`§11 다음 액션`) 에 별도 ADR 또는 spec revision 으로 처리 |

### 2.3 v1.0.0 milestone 정의 (ADR-004 §2.2 와 정합)

`plugin track v1.0.0 = cli buddy W3-2~W3-6 cascade 완료 + production-proven`. 본 ADR 의 lock-in 은 ADR-004 §2.2 #1 의 *trigger 해제* — 실제 cascade 작업 시작은 별도.

### 2.4 cli buddy versioning namespace

`cli` 도 plugin 과 동일 `vX.Y.Z` namespace 공유 (ADR-004 §2.4 revised). 기존 `v0.1.0` (cli binary, 2026-04-26) 가 cli track 의 시작점. cli buddy v0.2.0 release 시점은 *W3-2~W3-6 모두 완료 + reference agent (웹툰) 동작* 후 (spec §8 acceptance criteria).

---

## 3. Alternatives considered (rejected)

### 3.1 §4.1 option (b) MCP server stdio (rejected as primary, retained as future v2 path)

**Approach**: plugin buddy 의 MCP server (analytics-mcp + 후속) 를 cli buddy 가 직접 호출.

**Rejected because**:
- 현재 plugin buddy 의 MCP server (buddy-mcp) 는 doctor / stats / feature_* / analytics_query_* 만 노출 — *skill 호출* 은 MCP tool 로 노출 안 됨 (skill 은 router 를 통해 Claude session 안에서 동작)
- cli buddy 의 agent 가 skill chain (concretize-idea → ... → ship-release) 을 실행하려면 *Claude session 자체가 필요* → 결국 option (a) 의 subprocess 패턴 회피 불가
- option (b) 는 *skill MCP exposure* (별도 spec) 가 lock-in 된 후에 v2 path 로 재검토

### 3.2 §4.1 option (c) 자체 LLM 호출 (rejected)

**Approach**: Anthropic API 직접 호출, PROCEDURE.md 를 cli buddy 가 parse.

**Rejected because**:
- claude code 의 hook / agent / MCP runtime 환경 *외부* — buddy 의 4 자산 (skill / MCP / agent / hook) 활용 못 함
- LLM 호출 비용 자체 부담 — 사용자 의 Claude Code subscription 활용 못 함
- charter §2.1 의 "Claude Code 의 plugin 으로 설치되어 동작" 정신 위반

### 3.3 §4.1 option (d) hybrid (rejected as primary, retained as future optimization)

**Approach**: basic = (a), advanced = (b).

**Rejected because**:
- 두 path 유지 부담 — 동일 agent 가 단순한 task 와 advanced task 사이 분기 시 *상태 동기화* 복잡
- 첫 ship 은 단순함 우선 (KISS 원칙 — `~/.claude/rules/coding-style.md`)
- (b) 의 trigger 가 발생 (skill MCP exposure spec lock-in) 시 hybrid 가 자연 진화 — 즉 *현재 (a)* 가 *미래 hybrid 의 baseline*

### 3.4 Spec 의 §1.3 4-책임 scope 확장 (rejected)

**Approach**: agent 간 *통신*, *분산 실행*, *외부 SaaS 직접 통합* 을 v0.2.0 부터 포함.

**Rejected because**:
- charter §3.3 의 4-책임 (agent 생성 / 실행 / 종료 / 설정) 이 사용자 발화 lock-in — 확장은 *별도 사용자 결정 + ADR* 후
- 추가 책임은 spec acceptance gate (§8) 와 phase 분할 (§9) 의 cost 추정을 무의미하게 만듦
- Q-4 / Q-5 / Q-6 (multi-machine / agent 간 통신 / sandbox) 가 v2.0+ 로 deferred — 본 ADR 은 그 deferred 정책도 lock-in

---

## 4. Consequences

### 4.1 Positive

- **W3-2 ~ W3-6 cascade 의 trigger 해제** — TUI 설계, agent runtime, embedding layer, v0.1.0 재배치, reference agent (웹툰) 모두 *별 cycle 진입 가능*
- **ADR-004 §2.2 condition #1 의 진척** — plugin track v1.0.0 entry condition 4 중 #1 (cli buddy W3-2~6 완료) 의 *시작점* 확보
- **roadmap.md §4 / §5 / §6 의 actual rewrite trigger** (ADR-002) — v0.2 dashboard / v0.3 task DAG / v1.0 통합 outline 이 *cli buddy 의 sub-feature* 로 흡수 가능
- **spec 본문 §11 의 다음 액션 1번 완료** — Status flip 으로 *다음 액션 2번 (본 ADR 작성) 도 본 commit 으로 동시 완료*

### 4.2 Negative

- **3-6 month 의 single-dev cadence commit** — spec §9 의 추정. plugin buddy v0.x 의 minor bump 보다 훨씬 느린 cadence. *사용자 페이스* 변동 시 deferred 시점 늘어남
- **Open question 6 (Q-1~Q-6) deferred** — 향후 *각 phase 진입 시* 별 ADR 필요. 누적 ADR 수 증가
- **option (b) MCP / option (d) hybrid 미선택** — *추후 skill MCP exposure* 가 나오면 본 ADR 의 §4.1 결정 재검토 필요 (ADR-005 supersede 또는 ADR-005-extension)

### 4.3 Neutral

- buddy 기존 자산 *코드 변경 0* — 본 ADR + cli-buddy-spec Status flip + ADR Index 갱신만. *plan-only* commit
- cli buddy v0.1.0 binary (2026-04-26 release) 는 *그대로 유효* — W3-5 의 *재배치* 가 코드 위치 / 진입점 변경만, semantics 유지

---

## 5. Verification

### 5.1 Static (적용 시점)

- `docs/cli-buddy-spec.md` 의 Status field 가 `Draft` 가 아닌 `Accepted` 명시 (본 ADR reference 포함)
- `docs/superpowers/decisions/README.md` 의 ADR Index 표에 ADR-005 row 추가
- `docs/superpowers/decisions/README.md` 의 *향후 ADR 후보* 표에서 *cli buddy spec scope decision* 항목 제거 (본 ADR 로 충족)
- `docs/HANDOFF.md` 의 cli buddy 트랙 상태 — W3-1 spec lock-in Done 반영

### 5.2 Runtime (trigger 시)

- W3-2 (TUI) 진입 시 — 본 ADR 의 §2.2 decisions 가 모두 *참조 가능* 한 상태 유지
- v1.0.0 release prep — ADR-004 §2.2 condition #1 의 검증 시 본 ADR 이 *trigger 해제 evidence* 로 활용

---

## 6. Trigger to revisit

- **skill MCP exposure spec** lock-in (별 spec) — §4.1 option (b) MCP / (d) hybrid 가 재검토 trigger
- **cli buddy W3-2 진행 중 design 결정** 이 spec 본문과 silent conflict — 본 ADR supersede 또는 ADR-005-extension
- **Q-1 (TUI vs web dashboard)** 의 결정 시점 — 본 ADR 은 *TUI 채택* (charter §3.1) 가정. web dashboard 채택 결정 시 *cli buddy 의 form factor* 변경
- **charter §3 의 cli buddy 정의 자체 변경** — 사용자 명시 결정 시 본 ADR 의 §2.2 lock-in 전체 재검토

---

## 7. References

- [`docs/archive/cli-buddy-spec.md`](../../archive/cli-buddy-spec.md) — 본 ADR 의 lock-in 대상 spec (§1.3 / §4.1 / §5.1 / §7 / §9 / §10)
- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §3 cli buddy / §6.2 진화 방향
- [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002) — roadmap × charter gap, *cli buddy 안으로 흡수* 정책
- [`docs/superpowers/decisions/2026-05-11-plugin-version-reset.md`](./2026-05-11-plugin-version-reset.md) (ADR-004) — plugin track version reset + v1.0.0 entry condition #1 (본 ADR 이 trigger 해제)
- [`docs/archive/roadmap.md`](../../archive/roadmap.md) §4 / §5 / §6 (재평가 마킹) — 본 ADR 채택 후 *actual rewrite* trigger
- [`docs/HANDOFF.md`](../../HANDOFF.md) — 트랙 상태 (cli buddy W3-1 Done 반영)
