# Buddy — Backlog (단일 SSoT)

> **이 문서가 cross-track 잔여 작업의 단일 SSoT.** 완료된 항목은 **본 문서에 등장하지 않는다** (코드 / commit / ADR 이 증거). 새 잔여 항목 발견 시 본 문서로 통합.
>
> **Baseline**: v0.7.5 (2026-05-19) + ADR-006/007/008. Dogfood cycle-2 진행 중.
>
> **Identity / direction 문서 (변경 X, gap 검토만 — §3)**:
> [`two-tracks-charter.md`](./two-tracks-charter.md) · [`v0.1-spec.md`](./v0.1-spec.md) · [`cli-buddy-spec.md`](./cli-buddy-spec.md) · [`decision-1-schema-fields.md`](./decision-1-schema-fields.md) · [`skill-map.md`](./skill-map.md) · [`dogfood-guide.md`](./dogfood-guide.md) · [`superpowers/decisions/README.md`](./superpowers/decisions/README.md) · [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md)
>
> **이 SSoT 의 redirect (slimmed)**: [`HANDOFF.md`](./HANDOFF.md) §1 · [`roadmap.md`](./roadmap.md) · [`tasks.md`](./tasks.md)

---

## §1. 한 줄 — 전체 진행률 약 **90%**

핵심 deliverable (plugin orchestrator + cli buddy agent runtime + hook reliability monitor + MCP tools + analytics adapter) 은 **모두 ship 완료**. 남은 것은 (a) **B-2 production dogfood** — v1.0.0 publish 의 유일한 unclosed entry condition, (b) **polish** (i18n / release SSoT / TUI follow-on-of-follow-on), (c) **trigger-bound** (외부 신호 대기).

---

## §2. 진행률 표

| 영역 | 진행률 | 비고 |
|------|--------|------|
| Plugin orchestrator — 9-phase × 148 skill | **100%** | Skill Completion Cycle 100% (Batch 1~7) |
| Plugin commands — 99개 | **100%** | N-1 closure (ADR-001) + B6 lint `--strict` (ADR-006) |
| MCP tools — 12개 (analytics 7 + feature 5) | **100%** | W4-2.1~2.6 (v0.3.0) + ADR-008 lock-in |
| cli buddy W3 cascade (W3-1~W3-6) | **100%** | TUI 7 modes + agent runtime + scheduler + cascade chain + 자산 재배치 + reference agent |
| Hook reliability monitor (M1~M6 + v0.1.0 release) | **100%** | release binary 4 platforms + tag-triggered workflow |
| Dogfood validation — cycle-1 (B1-B5 fix) + cycle-2 (structural 5/5) | **100%** | |
| **Plugin v1.0.0 entry conditions (B-1~B-4)** | **75%** | B-1/B-3/B-4 closed, B-2 잔여 |
| Dogfood **cycle-2 production** (B-2 trigger) | ~5% | pre-flight Done, 3 paths user-paced — Wave 1 |
| i18n sweep (M5 deferred — 4 sub-task) | 25% | W2-3 ✅ (en/ko 57/57 parity), 3 잔여 — Wave 2 |
| Release polish (M6 deferred — 4 sub-task) | 25% | ci.yml ✅ (v0.6.2), 3 잔여 — Wave 3 |
| TUI / runtime UX follow-on (5 sub-task) | 0% | dogfood signal 대기 — Wave 4 |
| Trigger-bound (Korea / USA-EU / PG-MySQL / notarize / cycle-2 live) | N/A | 외부 신호 대기 — Wave 5 |
| Indefinite defer (Non-goal 또는 trigger 부재) | N/A | 잡지 말 것 — Wave 6 |

**가중치 합산**:
- Core deliverable (plugin / cli buddy / hook / MCP): ~98%
- v1.0.0 entry conditions: 75%
- Polish + DX: ~20%
- **종합 약 90%** (subjective).

---

## §3. 코드 ↔ identity 문서 gap 검토

> identity 문서는 *유지* (사용자 명시). 그 안에 *코드와 어긋난 부분이 있는지* 만 검토.

| 문서 | 역할 | 코드와 gap | 처리 |
|------|------|----------|------|
| `two-tracks-charter.md` | 두 트랙 책임 경계 SSoT | NO GAP | 변경 X |
| `v0.1-spec.md` | M1~M5 LOCKED spec (역사적) | NO GAP | 변경 X (역사적 보존) |
| `cli-buddy-spec.md` | cli buddy spec (Accepted, ADR-005) | minor — §9 W3-2 의 "deferred follow-on-of-follow-on (scheduler live indicator, log-tail scrollback)" 와 §9 W3-4 의 "branch-aware selection / self-check fail 의미" 가 *코드 미반영* (= Wave 4 항목). 이건 spec 자체가 *명시적으로 deferred 로 표기* 했으므로 진짜 drift 아님 | 변경 X (spec 이 정직) |
| `decision-1-schema-fields.md` | 옵션 A schema 결정 근거 | NO GAP | 변경 X |
| `skill-map.md` | 11-stage → 9-phase 매핑 reference | NO GAP (현행 SSoT 는 superpowers/specs/.../architecture.md) | 변경 X (역사적 보존) |
| `dogfood-guide.md` | v0.7.x dogfood 실행 절차 | NO GAP (사용자 페이스 진행 중) | 변경 X |
| `dogfood-feedback-template.md` | finding 양식 | NO GAP | 변경 X |
| `response-format-guide.md` | response 양식 | NO GAP (코드와 직접 매핑 X) | 변경 X |
| `superpowers/decisions/*` | ADR 8 건 모두 Accepted | NO GAP | 변경 X |
| `superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` | plugin 9-phase 아키텍처 SSoT | NO GAP (100% cover) | 변경 X |
| `superpowers/specs/2026-05-10-analytics-mcp-spec.md` | analytics-mcp spec | NO GAP (W4-2.1~2.6 Done) | 변경 X |
| **`HANDOFF.md`** | 세션 인계 가이드 (operational) | §1 표는 v0.7.2 baseline (현재 v0.7.5) — 한 줄 갱신 필요 | **§1 patch** |
| **`roadmap.md`** | M1~M6 milestone log (역사) + v0.2/v0.3/v1.0 outline (cur. superseded) | §4 v0.2 / §5 v0.3 / §6 v1.0 outline 이 ADR-002/005 로 *부분 supersede* + 본 문서로 위임 | **§4~§6 redirect 로 슬림화** |
| **`tasks.md`** | cross-track 작업 인벤토리 (v0.4.1 frozen, 대부분 완료) | 거의 전체가 완료된 항목 — 본 문서가 SSoT | **본 문서 redirect stub 로 슬림화** |

---

## §4. 잔여 작업 — Wave 1~6 (우선순위 정렬)

### Wave 1 — 사용자 페이스 진행 중 (모든 후속 trigger)

| ID | 작업 | 상태 | 비용 |
|----|------|------|------|
| **W1-1** | **B-2 production dogfood cycle-2 완주** — 외부 SaaS 1 건 이상 end-to-end + 9-phase ≥3 phase 사용 + agent ≥1 schedule 완주 + hook stats 누락 0 | ⚠ in progress | HIGH (1 week+) |
| W1-2 | Cycle-2 finding triage — bug-fix / UX-fix / ADR / deferred 분기 | conditional | LOW per finding |
| W1-3 | Cycle-2 close + 다음 cycle trigger 명시 | conditional | LOW |

→ v1.0.0 publish 의 unique blocker. 모든 다른 Wave 보다 명백히 우선.

### Wave 2 — i18n sweep (M5 deferred, 작은 비용 + 완성도 ↑)

| ID | 작업 | 위치 | 비고 |
|----|------|------|------|
| W2-1 | `config.ValidationError.Reason` persona catalog wiring + `translateConfigError` bullet 렌더링 | `internal/persona/` (`KeyConfigReason*` 8 keys 이미 declared, *코드 line 108 명시* "Not yet wired") | < 1 day |
| W2-2 | `queries.ErrInvalidLimit` / `ErrInvalidWindow` 카탈로그 이전 — 현재 한국어 sentinel (`"--window 은 5m, 1h, 24h 중 하나야."`) | `internal/queries/stats.go` + `events.go` | < 1 day |
| W2-3 | ~~English locale 카탈로그 채우기~~ | ✅ Done (`internal/persona/en.go` 57 keys, ko 와 parity) | — |
| W2-4 | Subcommand `--config` 인지 locale 해석 — 현재 root `PersistentPreRunE` 가 `config.DefaultPath()` 만 읽음 (`cmd/buddy/main.go:118` 주석 명시) | `cmd/buddy/main.go` | < 1 day |

### Wave 3 — Release polish (M6 deferred)

| ID | 작업 | 위치 | 비고 |
|----|------|------|------|
| W3-1 | `VERSION` 파일 / build-time embed — 현재 `0.7.5` 5 곳 분산 (main.go / Makefile / CHANGELOG / README / release.yml). `make verify-versions` 보호 중. release cadence 잦아짐 (1 day 안 v0.7.0~v0.7.4) | top-level + Makefile/release.yml inject | DRY 가치 측정 가능 시 |
| W3-2 | Third-party Actions SHA pinning + Dependabot — 현재 `actions/checkout@v6`, `setup-go@v6`, `softprops/action-gh-release@v3` 모두 mutable major tag | `.github/workflows/*.yml` | 보안 강화, < 1 day |
| ~~W3-3~~ | ~~별도 `ci.yml`~~ | ✅ Done (v0.6.2 — `.github/workflows/ci.yml`) | — |
| W3-4 | macOS notarization — Apple Developer ID 필요, 현재 `xattr -d com.apple.quarantine` 안내 | release.yml + sign step | Wave 5 trigger-bound 로 분류 가능 |

### Wave 4 — TUI / agent runtime follow-on (UX polish, dogfood signal 가능)

각 항목 < 1~2 day. *cycle-2 dogfood 결과* 가 우선순위 결정.

| ID | 작업 | 위치 |
|----|------|------|
| W4-1 | W3-4 branch-aware skill selection — `NextPhase.Branches` cascade 시 *어느 branch 채택* policy (env-var / CLI flag / interactive 중 design pending) | `internal/agent/runtime.go` |
| W4-2 | W3-4 self-check fail → step retry / abort 의미 정립 — 현재 parser 만 verdict 추출, runtime 동작 동일 | `internal/agent/runtime.go` |
| W4-3 | `buddy agent edit <id>` CLI subcommand — 현재 TUI `e` 만 존재, CLI 없음 | `cmd/buddy/agent_cmd.go` |
| W4-4 | Scheduler pane live "currently running" indicator (W3-2 deferred follow-on-of-follow-on) | `internal/tui/model.go` |
| W4-5 | Log-tail scrollback + auto-stop on run-end (W3-2 deferred follow-on-of-follow-on) | `internal/tui/model.go` |

### Wave 5 — Trigger-bound (외부 신호 대기)

| ID | 작업 | Trigger |
|----|------|---------|
| W5-1 | Korea cluster 3 skill (`consult-korea-legal-context` / `draft-korea-patent-application` / `audit-korea-cii-vulnerability`) | target market = Korea 결정 |
| W5-2 | USA / EU cluster | target market = USA 또는 EU 결정 |
| W5-3 | analytics-mcp PostgreSQL / MySQL adapter — 같은 `Adapter` interface, driver 만 교체 | production scale 발생 |
| W5-4 | macOS notarization 진행 | Apple Dev ID 확보 + widely distributed |
| W5-5 | Cycle-2 *live dispatch* (5 single-skill + 4 stage cascade) | 별 세션 (token 6-8x) |

### Wave 6 — Indefinite defer (Non-goal 또는 trigger 부재)

| ID | 작업 | 사유 |
|----|------|------|
| W6-1 | v1.0 plugin model — 외부 binary plugin + IPC | roadmap §6 outline, trigger 부재. cli-buddy-spec §7 mapping 따라 deferred |
| W6-2 | v1.0 cross-harness — Codex / OpenCode 지원 | harness-analysis §3 갭 C "환상", v1.0+ 검토만 |
| W6-3 | v1.0 AGENTS.md auto-sync | **cli-buddy-spec §7 의 명시적 Non-goal** — *AGENTS.md = capability metadata* 와 *cli buddy 의 agent (실행체)* 가 다른 개념 |
| W6-4 | D-2 v0.2 멀티-머신 통합 | single-machine 가정 (v0.1-spec §1 Non-goal) |
| W6-5 | D-4 v0.3 DAG 시각화 형식 (graphviz / mermaid / TUI) | trigger 미발생 |
| W6-6 | D-5 v1.0 plugin 권한 경계 (DB 직접 vs IPC) | trigger 미발생 |
| W6-7 | A-5.1 Quick Win C — commands/*.md body slim (~600 → ~200 byte/file) | 멀티-invoke trace 측정 가능 시 |
| W6-8 | A-5.2 CONTRIBUTING.md + lint enforcement (`disable-model-invocation: true`) | 첫 contributor PR 직전까지 |

---

## §5. 우선순위 한 줄

```
Wave 1  >>  Wave 2  ≥  Wave 3  >  Wave 4  >>  Wave 5  >>>  Wave 6
```

**Tiebreaker** (autoplan scope phase = P1 완성도 + P2 감당 가능 우선):
- Wave 2 가 Wave 3 보다 약간 우선 — i18n 은 M5 deferred 의 *명시적 약속 회수*
- Wave 4 안에서: W4-1 (branch-aware) > W4-2 (self-check fail) > W4-3 (edit CLI) > W4-4 / W4-5 (TUI cosmetic) — cascade *runtime semantic* 영향 큰 쪽 우선
- W3-4 (notarization) 는 Wave 3 vs Wave 5 양쪽 가능 — 본 문서는 **Wave 3 에 표기, Wave 5 cross-ref** (Apple Dev ID 가 외부 trigger 이므로 Wave 5 이동도 합리)

---

## §6. 다음 액션 추천

1. **본 문서 review** — Wave 분류 + 진행률 표 사용자 confirm.
2. **Wave 1 진행 — B-2 production dogfood** — [`docs/notes/2026-05-19-dogfood-result-cycle-2.md`](./notes/2026-05-19-dogfood-result-cycle-2.md) 의 §B 3 paths 직접 실행 + finding 즉시 §C 기록.
3. **Wave 2 (i18n) 는 Wave 1 finding 가 i18n 영역일 때 우선** 진입. 아니면 Wave 1 close 후 자연 진입.

---

## §7. SSoT 관계 — 다른 문서들과의 분담

| 문서 | 책임 (post-cleanup) |
|------|-------------------|
| **본 문서 `BACKLOG.md`** | cross-track *잔여 작업* 의 단일 SSoT + 진행률 |
| `HANDOFF.md` | 세션 인계 가이드 (operational) — §1 "현재 위치" 한 줄 + 본 문서 redirect |
| `roadmap.md` | M1~M6 *historical milestone log* (역사적 보존). v0.2/v0.3/v1.0 outline 부분은 본 문서로 위임 |
| `tasks.md` | 본 문서 redirect stub (v0.4.1 인벤토리는 git history 참조) |
| `two-tracks-charter.md` / `cli-buddy-spec.md` / `v0.1-spec.md` / 기타 identity 문서 | **변경 X** — 정체성 + 방향성 SSoT |
| `dogfood-guide.md` / `dogfood-feedback-template.md` | dogfood 실행 절차 + 양식 — *변경 X* |
| `superpowers/decisions/*` / `superpowers/specs/*` | ADR + spec — *변경 X* (역사적 결정 + 아키텍처 SSoT) |
| `notes/{date}-*.md` | dated handoff / dogfood result — *역사적 보존* (date-stamped) |

본 문서가 stale 해지면 **본 문서부터 갱신**. tasks.md / roadmap.md 의 backlog 부분은 *본 문서로 통합 완료*.

---

## §8. References

- [`docs/HANDOFF.md`](./HANDOFF.md) — 세션 인계 (§1 patched to v0.7.5)
- [`docs/cli-buddy-spec.md`](./cli-buddy-spec.md) §9 phase table — W3 cascade 100%
- [`docs/notes/2026-05-11-cycle-handoff.md`](./notes/2026-05-11-cycle-handoff.md) — 이전 cycle close baseline
- [`docs/notes/2026-05-19-dogfood-result-cycle-2.md`](./notes/2026-05-19-dogfood-result-cycle-2.md) — 진행 중 cycle (Wave 1 의 실행 매뉴얼)
- [`docs/superpowers/decisions/README.md`](./superpowers/decisions/README.md) — ADR Index (8건 모두 Accepted)
- [`docs/dogfood-guide.md`](./dogfood-guide.md) — v0.7.x dogfood 실행 절차
