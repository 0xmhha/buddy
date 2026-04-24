# Buddy

> 무에서 유를 창조하는 작업, 옆에서 같이 도와주는 친구.
> Claude Code 위에서 *세션 운영의 신뢰성·관제·오케스트레이션*을 한 자리에서.

```
                  ╭──────────────────────────╮
   you ──── ▶  │      buddy (control)     │ ── ▶  Claude Code sessions
                  │  hooks · state · tasks   │       (1, 2, 3, ... N)
                  ╰──────────────────────────╯
```

---

## 왜 만드는가

Harness 생태계 33개 프로젝트(claude-code, codex, opencode, SuperClaude, oh-my-*,
arc-reactor, recon, spec-kit, serena, cli-wrapper, token-monitor 등)를 분석한
결과, 모든 도구가 같은 4가지 갭을 공유한다는 사실이 드러났습니다:

1. **Hook reliability** — hook이 silent failure 나도 사용자는 모름
2. **State schema** — JSON 상태가 검증 없이 흩어져 충돌·손상 발생
3. **Task dependency / retry loop** — wave 병렬 후 실패 시 재실행 로직 없음
4. **통합 observability** — 토큰·세션·비용·hook 상태가 별개 도구로 분산

Buddy는 위 4가지를 *Claude Code 한정*으로 통합 해결하는 control plane입니다.
Cross-harness parity 야심은 의도적으로 포기 — 한 harness에서 깊이 있게 작동하는
도구가, 모든 harness에서 얕게 작동하는 도구보다 가치 있습니다.

> 분석 보고서: [`../harness-engineering-analysis.md`](../harness-engineering-analysis.md)

---

## 정체성: "친구"

Buddy는 *비서*도 *오케스트레이터*도 아닙니다. **친구**입니다.

| 친구가 하는 일 | Buddy가 하는 일 |
|---------------|----------------|
| 옆에 조용히 있다가 필요할 때 알려준다 | hook 실패·state 충돌·토큰 burn rate 알림 |
| 내가 잊은 걸 기억해 준다 | session/task 상태를 schema로 검증하며 보존 |
| 같이 막힌 문제를 풀어준다 | 실패한 task 재시도, 의존 그래프로 다음 단계 제시 |
| 잔소리하지 않는다 | 침묵이 default. fail loud, success quiet |

**페르소나 결정은 사용자 입력 필요** — `docs/v0.1-spec.md` §6 참조.

---

## 로드맵

장기 비전은 A+B+C 통합 control plane:
- **A) Control Plane** — 멀티-세션 통합 dashboard (token/cost/status)
- **B) Orchestration** — task DAG executor (wave + retry + dependency)
- **C) Reliability** — hook health monitor + state schema validation + WAL replay

v0.1은 **C)부터** wedge로 진입. 이유: 갭이 가장 비어 있고, 다른 도구의 의존
대상(=control plane이 다른 도구를 신뢰성 있게 관리하려면 신뢰 인프라가 먼저).

| 버전 | wedge | 주요 기능 |
|------|-------|----------|
| **v0.1** | Reliability | hook health monitor, state schema (Zod), WAL replay |
| v0.2 | + Control Plane | multi-session dashboard, token/cost unified view |
| v0.3 | + Orchestration | task DAG, wave executor, auto-retry |
| v1.0 | 통합 | AGENTS.md auto-sync, plugin model, MCP server |

---

## 1차 타겟

- **Claude Code only** (cross-harness parity는 의도적으로 보류)
- macOS / Linux (Windows는 v1.0+)
- **Go 1.22+** static binary, 런타임 의존성 0
- M3부터 `0xmhha/cli-wrapper`를 daemon supervision에 통합

---

## 현재 상태

✅ M1 + M2 (Go) 완료 — schema/SQLite/outbox/hook wrapper 모두 구현, 모든 테스트 통과.
🚧 M3 (daemon + cli-wrapper) 다음 차례.

```bash
make build         # bin/buddy 생성 (~9.5MB static binary)
make test          # 전체 테스트 (race detector 포함)
./bin/buddy --help # CLI 진입점
```

---

## 역사적 메모

v0.1 PoC는 처음 TypeScript/Node로 시작했으나, hook wrapper의 self-overhead
(매 hook 호출당 50-100ms vs Go 5-10ms)와 cli-wrapper와의 native 통합 가능성
때문에 Go로 pivot했습니다. TS 자산은 `archive/ts-poc/` 에 보존.

---

## 라이선스

Apache 2.0
