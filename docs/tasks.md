# Buddy — Task Inventory (deprecated, see `BACKLOG.md`)

> **이 문서는 *deprecated***. cross-track 잔여 작업의 단일 SSoT 는 [`BACKLOG.md`](./BACKLOG.md).
>
> v0.4.1 시점 (2026-05-11) 의 §A (Plugin) / §B (Go CLI) / §C (Housekeeping) / §D (Open Questions) 구조는 본 cycle (2026-05-19) 의 BACKLOG.md 통합으로 *대부분 완료* 되어 슬림화. 역사적 인벤토리는 `git log -- docs/tasks.md` 로 추적.

---

## 본 문서의 역사적 역할 (post-deprecation)

| 영역 | post-cleanup 위치 |
|------|----------------|
| §A Plugin (A-1 SSoT drift / A-2 stage skill 잔여 29 / A-3 MCP / A-4 dogfood) | **대부분 완료** — Skill Completion Cycle 100% / 12 MCP tools / B6 lint --strict (ADR-006) / cycle-1+cycle-2-structural Done. 잔여 = cycle-2 production dogfood → `BACKLOG.md` Wave 1 |
| §B Go CLI (B-1 dogfood feedback / B-2 i18n sweep / B-3 release polish / B-4 v0.2 Control Plane / B-5 v0.3 Orchestration / B-6 v1.0 통합) | **대부분 supersede 됨** — ADR-002 / ADR-005 에 의해 cli buddy spec W3-1~W3-6 cascade 로 통합. 잔여 polish (i18n 3건, release polish 3건) → `BACKLOG.md` Wave 2/3 |
| §C Housekeeping (C-1 main.go split / C-2 module path / C-3 gofmt drift) | **모두 완료 또는 clean** (v0.6.3 split + 4ce3ccb module rename + gofmt clean) |
| §D Open Questions (D-1~D-7) | trigger-bound 또는 supersede — `BACKLOG.md` Wave 5/6 |

---

## A-5 (Always-deferred, 본 cycle 에서 지속)

| ID | 작업 | trigger |
|----|------|---------|
| A-5.1 | Quick Win C — `plugin/commands/*.md` body slim (~600 → ~200 byte/file) | 멀티-invoke trace 측정 가능 시 |
| A-5.2 | CONTRIBUTING.md + lint enforcement (`disable-model-invocation: true`) | 첫 contributor PR 직전 |

(위 2건은 `BACKLOG.md` Wave 6 의 W6-7 / W6-8 와 동일.)

---

## 참조

- [`BACKLOG.md`](./BACKLOG.md) — **단일 SSoT** (잔여 작업 + 진행률 + 코드 gap)
- [`HANDOFF.md`](./HANDOFF.md) — 세션 인계 가이드
- [`roadmap.md`](./roadmap.md) — M1~M6 historical milestone log
- [`cli-buddy-spec.md`](./cli-buddy-spec.md) — cli buddy spec (Accepted, ADR-005)
- `git log --since='2026-05-01' -- docs/tasks.md` — 역사적 인벤토리 추적
